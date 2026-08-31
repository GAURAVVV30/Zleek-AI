"""NeMo Guardrails API — Socratic pedagogical enforcement endpoints.

Endpoints
---------
POST /guardrails/check    — Check if a student message violates Socratic rules
GET  /guardrails/status   — Return guardrails engine status (NeMo vs keyword)
"""

from typing import Any

from app.core.guardrails_engine import get_guardrails_engine
from fastapi import APIRouter
from pydantic import BaseModel, Field

router = APIRouter()


class GuardrailsCheckRequest(BaseModel):
    student_text: str = Field(
        ...,
        description="The student's raw input text to check against Socratic guardrails.",
        min_length=1,
        max_length=4000,
    )


@router.post("/check", summary="Check student input against Socratic guardrails")
def check_guardrails(request: GuardrailsCheckRequest) -> dict[str, Any]:
    """Intercept student messages that violate Socratic tutoring principles.

    This endpoint sits BEFORE the LLM evaluator in the learning pipeline.
    When a student attempts to get a direct answer or code solution, NeMo
    Guardrails intercepts the request and returns a pedagogically appropriate
    refusal — the LLM is **never called**.

    Example blocked inputs:
    - "Just write the Python code for me"
    - "Tell me the correct answer"
    - "Complete my assignment"

    Returns:
        - ``blocked``         — True if the input was blocked
        - ``reason``          — Why it was blocked (pattern name)
        - ``refusal_message`` — Socratic response to send to the student
        - ``method``          — ``"nemo_guardrails"`` or ``"keyword_guardrail"``
        - ``original_text``   — Echo of the student's input
    """
    engine = get_guardrails_engine()
    return engine.check(request.student_text)


@router.get("/status", summary="Guardrails engine status")
def guardrails_status() -> dict[str, Any]:
    """Return the active guardrails engine configuration.

    Reports whether NVIDIA NeMo Guardrails is active (requires NVIDIA_API_KEY)
    or the keyword-based fallback is in use.
    """
    engine = get_guardrails_engine()
    return engine.status()
