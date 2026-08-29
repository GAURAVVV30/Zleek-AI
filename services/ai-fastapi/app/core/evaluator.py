from typing import Any, Dict, Optional

from app.core.llm_client import LLMClient
from app.core.graph_engine import GraphEngine


class AssessmentEvaluator:
    """Evaluate learner submissions using a deterministic rubric and an LLM.

    The evaluator fetches the node's `assessment_rubric` from `GraphEngine`,
    prompts the LLM to score the `student_answer` against the rubric's
    `key_evaluation_criteria`, and returns a normalized JSON result.
    """

    def __init__(self, llm_client: LLMClient, graph_engine: GraphEngine) -> None:
        self.llm = llm_client
        self.graph = graph_engine

    def evaluate_submission(self, domain_id: str, node_id: str, student_answer: str) -> Dict[str, Any]:
        """Evaluate a student's answer for a specific domain node.

        Args:
            domain_id: The domain identifier containing the node.
            node_id: The node identifier that carries the assessment rubric.
            student_answer: The learner's submitted answer (text).

        Returns:
            A dict containing: `score` (float 0.0-1.0), `passed` (bool),
            `feedback` (string), and `remediation_hint` (string) only if failed.
        """

        # Retrieve node and rubric
        try:
            node = self.graph._node_lookup[domain_id][node_id]
        except KeyError as exc:
            raise ValueError(f"Node '{node_id}' not found in domain '{domain_id}'.") from exc

        rubric = node.get("assessment_rubric") or {}
        diagnostic_question = rubric.get("diagnostic_question", "Please evaluate the following response.")
        key_criteria = rubric.get("key_evaluation_criteria", [])
        remediation_hint = rubric.get("remediation_hint", "Consider reviewing the core concepts and examples.")

        # Build prompts
        system_prompt = (
            "You are a strict pedagogical evaluator. Grade answers against the provided criteria."
            " Provide a numeric score between 0.0 and 1.0, concise constructive feedback, and a remediation hint if the"
            " answer does not meet the threshold. Be objective and base the score on the criteria only."
        )

        criteria_text = "\n".join([f"{i+1}. {c}" for i, c in enumerate(key_criteria)]) if key_criteria else "(no explicit criteria provided)"

        user_prompt = (
            f"Diagnostic question: {diagnostic_question}\n"
            f"Key evaluation criteria:\n{criteria_text}\n\n"
            f"Student answer:\n{student_answer}\n\n"
            "Return ONLY valid JSON with the following fields: \n"
            "- score: a number between 0.0 and 1.0\n"
            "- passed: boolean (true if score >= 0.8)\n"
            "- feedback: constructive, Socratic feedback guiding the student\n"
            "- remediation_hint: (optional) short actionable hint if failed\n"
        )

        # Small guidance schema for the LLM (used only in prompt)
        response_schema = {
            "score": {"type": "number", "min": 0.0, "max": 1.0},
            "passed": {"type": "boolean"},
            "feedback": {"type": "string"},
            "remediation_hint": {"type": "string"},
        }

        result = self.llm.generate_structured_json(system_prompt=system_prompt, user_prompt=user_prompt, response_schema=response_schema)

        # If the LLM client returned an error object, propagate it in a consistent shape
        if isinstance(result, dict) and result.get("error"):
            return {"error": result.get("error"), "raw": result.get("raw")}

        # Normalize and validate response
        try:
            score_val = float(result.get("score", 0.0))
        except Exception:
            return {"error": "Invalid or missing 'score' in LLM response.", "raw": result}

        # Clamp score
        score = max(0.0, min(1.0, score_val))
        passed = bool(result.get("passed", score >= 0.8))

        # Ensure passed is consistent with score threshold
        if score >= 0.8:
            passed = True
        else:
            passed = False

        feedback = result.get("feedback") or "No feedback provided."
        remediation = result.get("remediation_hint") or (remediation_hint if not passed else None)

        response: Dict[str, Any] = {"score": round(score, 3), "passed": passed, "feedback": feedback}
        if not passed and remediation:
            response["remediation_hint"] = remediation

        return response
