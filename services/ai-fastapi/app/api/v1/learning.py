"""Learning endpoints — lesson generation and enriched answer evaluation.

The ``/evaluate`` endpoint now returns a full enriched response including:
  - LLM score and Socratic feedback
  - BKT mastery probability (D1)
  - Sentiment / emotional state (D2)
"""
import logging
from typing import Any

from app.core.evaluator import AssessmentEvaluator
from app.core.graph_engine import GraphEngine
from app.core.guardrails_engine import get_guardrails_engine
from app.core.llm_client import LLMClient
from app.core.rag_engine import ResourceRetriever
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

logger = logging.getLogger(__name__)


class GenerateLessonRequest(BaseModel):
    domain_id: str
    node_id: str


class EvaluateAnswerRequest(BaseModel):
    domain_id: str
    node_id: str
    student_answer: str
    attempt_history: list[int] | None = Field(
        default=None,
        description=(
            "Ordered list of previous binary responses for this node "
            "(1=correct, 0=incorrect). Used by BKT to compute mastery across "
            "the full attempt history. When omitted, only the current attempt "
            "is used."
        ),
    )


router = APIRouter()


# Global singletons — ResourceRetriever is lazy to avoid chromadb import crash
engine = GraphEngine()
_retriever = None
llm = LLMClient()
evaluator = AssessmentEvaluator(llm_client=llm, graph_engine=engine)


def _get_retriever():
    """Lazy-load ResourceRetriever to avoid chromadb import crash at startup."""
    global _retriever
    if _retriever is None:
        try:
            _retriever = ResourceRetriever()
        except (ImportError, RuntimeError, TypeError):
            logger.exception("Failed to initialize ResourceRetriever")
            _retriever = None
    return _retriever


@router.post("/lesson")
def generate_lesson(request: GenerateLessonRequest) -> dict[str, Any]:
    """Generate a short structured lesson for a node using RAG + LLM.

    Returns lesson content and the resource URLs used as context.
    """
    domain_id = request.domain_id
    node_id = request.node_id

    # Validate domain/node
    try:
        node = engine._node_lookup[domain_id][node_id]
    except KeyError:
        raise HTTPException(status_code=404, detail=f"Node '{node_id}' not found in domain '{domain_id}'")

    # Retrieve supporting resources (best-effort)
    ret = _get_retriever()
    resources: list[dict[str, Any]] = []
    if ret is not None:
        try:
            resources = ret.get_resources_for_node(node_id, n_results=5) or []
        except (AttributeError, RuntimeError):
            logger.exception("Failed to retrieve resources for node %s", node_id)
    resource_urls: list[str] = [
        url for r in resources if (url := r.get("url")) and isinstance(url, str)
    ]

    # Build prompt for structured lesson
    key_concepts = node.get("key_concepts") or node.get("concepts") or []
    concepts_text = "\n".join([f"- {c}" for c in key_concepts]) if key_concepts else ""

    system_prompt = (
        "You are an expert instructor. Create a concise, structured lesson that teaches the concept."
        " Include explicit 'Key Concepts', a step-by-step explanation, and any LaTeX math expressions if relevant."
    )

    user_prompt = (
        f"Node ID: {node_id}\n"
        f"Key concepts:\n{concepts_text}\n\n"
        f"Use these supporting resources:\n{chr(10).join(resource_urls)}\n\n"
        "Return only valid JSON with: title, content_markdown, latex_expressions (array of LaTeX strings)."
    )

    response_schema = {
        "title": {"type": "string"},
        "content_markdown": {"type": "string"},
        "latex_expressions": {"type": "array"},
    }

    try:
        lesson = llm.generate_structured_json(
            system_prompt=system_prompt, user_prompt=user_prompt, response_schema=response_schema
        )
    except (RuntimeError, ValueError, TypeError):
        logger.exception("LLM failed to generate lesson for node %s", node_id)
        raise HTTPException(status_code=500, detail="LLM generation failed")

    return {"node_id": node_id, "resources": resources, "lesson": lesson}


@router.post("/evaluate")
def evaluate_answer(request: EvaluateAnswerRequest) -> dict[str, Any]:
    """Evaluate a student's answer with NeMo Guardrails + BKT + Sentiment.

    Pipeline:
    1. Guardrails check — blocks direct answer / code requests (NeMo / keyword)
    2. Sentiment analysis — detects frustration, adjusts LLM tone
    3. LLM grading — Socratic feedback via NVIDIA NIM
    4. BKT mastery update — probabilistic mastery tracking

    Response includes:
    - ``blocked``       — True if guardrails fired (LLM never called)
    - ``score``         — normalized 0.0-1.0 from LLM
    - ``passed``        — True if score >= 0.7
    - ``feedback``      — tone-adapted Socratic feedback
    - ``bkt``           — BKT mastery estimate
    - ``sentiment``     — detected emotion
    """
    # ---- Step 1: Guardrails check (BEFORE LLM) ----
    guardrails = get_guardrails_engine()
    guard_result = guardrails.check(request.student_answer)
    if guard_result["blocked"]:
        return {
            "blocked": True,
            "reason": guard_result["reason"],
            "refusal_message": guard_result["refusal_message"],
            "guardrails_method": guard_result["method"],
            "score": None,
            "passed": False,
        }

    # ---- Step 2-4: Full evaluation pipeline ----
    try:
        result = evaluator.evaluate_submission(
            domain_id=request.domain_id,
            node_id=request.node_id,
            student_answer=request.student_answer,
            attempt_history=request.attempt_history,
        )
    except ValueError as exc:
        raise HTTPException(status_code=404, detail=str(exc))

    return {"blocked": False, "guardrails_method": guard_result["method"], **result}
