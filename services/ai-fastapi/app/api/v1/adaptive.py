from typing import Any

from app.core.adaptive_engine import AdaptiveEngine
from fastapi import APIRouter
from pydantic import BaseModel, Field


class NextActionRequest(BaseModel):
    node_id: str
    score: float
    failed_attempts: int = 0
    p_mastery: float | None = Field(
        default=None,
        description=(
            "BKT P(mastery) value from the mastery estimator. When provided, "
            "this takes precedence over the raw score for the advance decision."
        ),
    )


router = APIRouter()

# single engine instance
engine = AdaptiveEngine()


@router.post("/next-action")
def next_action(request: NextActionRequest) -> dict[str, Any]:
    """Determine the next adaptive action for a learner.

    Prioritises BKT ``p_mastery`` (>= 0.95 to advance) when provided.
    Falls back to the raw ``score`` (>= 0.80) for backward compatibility.
    """
    result = engine.determine_next_action(
        request.node_id,
        request.score,
        request.failed_attempts,
        p_mastery=request.p_mastery,
    )
    return {"node_id": request.node_id, **result}
