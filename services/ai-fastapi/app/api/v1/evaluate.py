from fastapi import APIRouter
from pydantic import BaseModel

router = APIRouter()


class EvaluationRequest(BaseModel):
    user_id: str
    domain: str = "software_architect"
    score: float
    feedback: str | None = None


@router.post("/evaluate")
def evaluate_user(payload: EvaluationRequest):
    return {
        "user_id": payload.user_id,
        "domain": payload.domain,
        "score": payload.score,
        "grade": "Pass" if payload.score >= 70 else "Needs Improvement",
        "feedback": payload.feedback or "Keep practicing the core architecture concepts.",
    }
