import logging
from typing import Any

from app.core.goal_intelligence import GoalAnalyzer
from app.core.graph_engine import GraphEngine
from app.core.llm_client import LLMClient
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

logger = logging.getLogger(__name__)


class GoalRequest(BaseModel):
    user_text: str


router = APIRouter()


# Initialize globals once
engine = GraphEngine()
llm = LLMClient()
analyzer = GoalAnalyzer(llm_client=llm, graph_engine=engine)


@router.post("/analyze")
def analyze_goal(request: GoalRequest) -> dict[str, Any]:
    """Analyze user goal text and map to a domain plus prior skills."""
    try:
        result = analyzer.analyze_user_intent(request.user_text)
    except (ValueError, TypeError, RuntimeError) as exc:
        logger.exception("Error analyzing goal for input: %s", request.user_text)
        raise HTTPException(status_code=500, detail=str(exc))
    except Exception:
        logger.exception("Unexpected error analyzing goal")
        raise HTTPException(status_code=500, detail="Internal server error")

    return result
