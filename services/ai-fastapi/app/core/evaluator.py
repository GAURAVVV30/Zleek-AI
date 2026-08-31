"""Assessment evaluator integrating LLM grading, BKT mastery, and sentiment analysis.

Pipeline for each submission
-----------------------------
1. Fetch the node's ``assessment_rubric`` from ``GraphEngine``.
2. Run sentiment analysis on the student's answer (D2).
3. Build LLM prompts — tone-adjusted if frustration is detected (D2).
4. Grade the answer with the LLM (existing flow).
5. Convert the LLM score to a binary signal and update BKT mastery (D1).
6. Return a rich response that includes score, BKT mastery probability,
   sentiment, and Socratic feedback.

Architecture boundary
---------------------
This service *proposes* mastery (``mastered: bool``, ``p_mastery: float``).
The Go orchestration service remains the final authority for writing the
official Competency Record and unlocking graph nodes.
"""

from typing import Any

from app.core.bkt_estimator import BKTEstimator, get_bkt_estimator
from app.core.graph_engine import GraphEngine
from app.core.llm_client import LLMClient
from app.core.sentiment_analyzer import (
    SentimentAnalyzer,
    build_tone_system_prompt,
    get_sentiment_analyzer,
)

# Score threshold for converting a graded score into a BKT binary signal.
# A score >= this value is treated as a "correct" response (1), else "incorrect" (0).
SCORE_CORRECT_THRESHOLD: float = 0.7


class AssessmentEvaluator:
    """Evaluate learner submissions using LLM + BKT + sentiment analysis.

    The evaluator fetches the node's ``assessment_rubric`` from ``GraphEngine``,
    prompts the LLM (with sentiment-adapted tone) to score the answer, converts
    the score into a binary BKT signal, and returns a normalized JSON result
    enriched with mastery probability and emotional context.
    """

    def __init__(
        self,
        llm_client: LLMClient,
        graph_engine: GraphEngine,
        bkt_estimator: BKTEstimator | None = None,
        sentiment_analyzer: SentimentAnalyzer | None = None,
    ) -> None:
        self.llm = llm_client
        self.graph = graph_engine
        self.bkt = bkt_estimator or get_bkt_estimator()
        self.sentiment = sentiment_analyzer or get_sentiment_analyzer()

    def evaluate_submission(
        self,
        domain_id: str,
        node_id: str,
        student_answer: str,
        attempt_history: list[int] | None = None,
    ) -> dict[str, Any]:
        """Evaluate a student's answer for a specific domain node.

        Args:
            domain_id:       The domain identifier containing the node.
            node_id:         The node identifier that carries the assessment rubric.
            student_answer:  The learner's submitted answer (text).
            attempt_history: Optional list of previous binary responses for this
                             node (1=correct, 0=incorrect).  When provided, BKT
                             mastery is computed across the full history + this
                             attempt.  When omitted, only this attempt is used.

        Returns:
            A dict containing:
                - ``score``           — normalized float 0.0–1.0
                - ``passed``          — bool (score >= SCORE_CORRECT_THRESHOLD)
                - ``feedback``        — Socratic feedback string
                - ``remediation_hint``— short actionable hint (only if failed)
                - ``bkt``             — BKT mastery result dict (p_mastery, mastered, …)
                - ``sentiment``       — emotion analysis dict
                - ``binary_signal``   — the 0/1 signal fed into BKT
        """

        # ------------------------------------------------------------------ #
        # 1. Retrieve node and rubric
        # ------------------------------------------------------------------ #
        try:
            node = self.graph._node_lookup[domain_id][node_id]
        except KeyError as exc:
            raise ValueError(f"Node '{node_id}' not found in domain '{domain_id}'.") from exc

        rubric = node.get("assessment_rubric") or {}
        diagnostic_question = rubric.get("diagnostic_question", "Please evaluate the following response.")
        key_criteria = rubric.get("key_evaluation_criteria", [])
        remediation_hint_default = rubric.get("remediation_hint", "Consider reviewing the core concepts and examples.")

        # Pull any per-node BKT params from the graph and register them
        bkt_params = node.get("bkt_params")
        if bkt_params and isinstance(bkt_params, dict):
            self.bkt.register_skill(
                node_id,
                p_init=bkt_params.get("p_init", 0.30),
                p_learn=bkt_params.get("p_learn", 0.20),
                p_guess=bkt_params.get("p_guess", 0.25),
                p_slip=bkt_params.get("p_slip", 0.10),
            )

        # ------------------------------------------------------------------ #
        # 2. Sentiment analysis (D2) — run BEFORE prompting the LLM
        # ------------------------------------------------------------------ #
        sentiment_result = self.sentiment.analyze(student_answer)
        tone_override = sentiment_result.get("tone_override", "standard")

        # ------------------------------------------------------------------ #
        # 3. Build LLM prompts — tone-adjusted if frustration is detected
        # ------------------------------------------------------------------ #
        base_system_prompt = (
            "You are a strict pedagogical evaluator. Grade answers against the provided criteria."
            " Provide a numeric score between 0.0 and 1.0, concise constructive feedback, and a remediation hint if the"
            " answer does not meet the threshold. Be objective and base the score on the criteria only."
        )
        system_prompt = build_tone_system_prompt(base_system_prompt, tone_override)

        criteria_text = (
            "\n".join([f"{i+1}. {c}" for i, c in enumerate(key_criteria)])
            if key_criteria
            else "(no explicit criteria provided)"
        )

        user_prompt = (
            f"Diagnostic question: {diagnostic_question}\n"
            f"Key evaluation criteria:\n{criteria_text}\n\n"
            f"Student answer:\n{student_answer}\n\n"
            "Return ONLY valid JSON with the following fields: \n"
            "- score: a number between 0.0 and 1.0\n"
            f"- passed: boolean (true if score >= {SCORE_CORRECT_THRESHOLD})\n"
            "- feedback: constructive, Socratic feedback guiding the student\n"
            "- remediation_hint: (optional) short actionable hint if failed\n"
        )

        response_schema = {
            "score": {"type": "number", "min": 0.0, "max": 1.0},
            "passed": {"type": "boolean"},
            "feedback": {"type": "string"},
            "remediation_hint": {"type": "string"},
        }

        # ------------------------------------------------------------------ #
        # 4. LLM grading
        # ------------------------------------------------------------------ #
        result = self.llm.generate_structured_json(
            system_prompt=system_prompt,
            user_prompt=user_prompt,
            response_schema=response_schema,
        )

        if isinstance(result, dict) and result.get("error"):
            return {"error": result.get("error"), "raw": result.get("raw")}

        # Normalize score
        try:
            score_val = float(result.get("score", 0.0))
        except (TypeError, ValueError):
            return {"error": "Invalid or missing 'score' in LLM response.", "raw": result}

        score = max(0.0, min(1.0, score_val))
        passed = score >= SCORE_CORRECT_THRESHOLD

        feedback = result.get("feedback") or "No feedback provided."
        remediation = result.get("remediation_hint") or (remediation_hint_default if not passed else None)

        # ------------------------------------------------------------------ #
        # 5. BKT mastery update (D1)
        # ------------------------------------------------------------------ #
        binary_signal = 1 if passed else 0

        # Build full history: previous attempts + this one
        history = list(attempt_history) if attempt_history else []
        history.append(binary_signal)

        bkt_result = self.bkt.estimate(node_id=node_id, attempt_history=history)

        # ------------------------------------------------------------------ #
        # 6. Assemble and return the enriched response
        # ------------------------------------------------------------------ #
        response: dict[str, Any] = {
            "score": round(score, 3),
            "passed": passed,
            "feedback": feedback,
            "binary_signal": binary_signal,
            "bkt": bkt_result,
            "sentiment": sentiment_result,
        }

        if not passed and remediation:
            response["remediation_hint"] = remediation

        return response
