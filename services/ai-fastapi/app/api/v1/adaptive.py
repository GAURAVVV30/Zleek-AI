from typing import Any, Dict

from fastapi import APIRouter
from pydantic import BaseModel

from app.core.adaptive_engine import AdaptiveEngine


class NextActionRequest(BaseModel):
    node_id: str
    score: float
    failed_attempts: int = 0


router = APIRouter()

# single engine instance
engine = AdaptiveEngine()


@router.post("/next-action")
def next_action(request: NextActionRequest) -> Dict[str, Any]:
    """Determine the next adaptive action for a learner based on score and attempts."""
    result = engine.determine_next_action(request.node_id, request.score, request.failed_attempts)
    return {"node_id": request.node_id, **result}
