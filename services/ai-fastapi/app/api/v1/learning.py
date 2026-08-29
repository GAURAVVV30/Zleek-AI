from typing import Any, Dict, List

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

from app.core.graph_engine import GraphEngine
from app.core.rag_engine import ResourceRetriever
from app.core.llm_client import LLMClient
from app.core.evaluator import AssessmentEvaluator


class GenerateLessonRequest(BaseModel):
    domain_id: str
    node_id: str


class EvaluateAnswerRequest(BaseModel):
    domain_id: str
    node_id: str
    student_answer: str


router = APIRouter()


# Global singletons to avoid reloading graphs/clients on each request
engine = GraphEngine()
retriever = ResourceRetriever()
llm = LLMClient()
evaluator = AssessmentEvaluator(llm_client=llm, graph_engine=engine)


@router.post("/lesson")
def generate_lesson(request: GenerateLessonRequest) -> Dict[str, Any]:
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

    # Retrieve supporting resources
    resources = retriever.get_resources_for_node(node_id, n_results=5)
    resource_urls: List[str] = [r.get("url") for r in resources if r.get("url")]

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

    lesson = llm.generate_structured_json(system_prompt=system_prompt, user_prompt=user_prompt, response_schema=response_schema)

    return {"node_id": node_id, "resources": resources, "lesson": lesson}


@router.post("/evaluate")
def evaluate_answer(request: EvaluateAnswerRequest) -> Dict[str, Any]:
    """Evaluate a student's answer and return score, pass/fail, and feedback."""
    try:
        result = evaluator.evaluate_submission(request.domain_id, request.node_id, request.student_answer)
    except ValueError as exc:
        raise HTTPException(status_code=404, detail=str(exc))

    return result
