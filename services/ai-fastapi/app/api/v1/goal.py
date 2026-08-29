from typing import Any, Dict

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

from app.core.graph_engine import GraphEngine
from app.core.llm_client import LLMClient
from app.core.goal_intelligence import GoalAnalyzer


class GoalRequest(BaseModel):
    user_text: str


router = APIRouter()


# Initialize globals once
engine = GraphEngine()
llm = LLMClient()
analyzer = GoalAnalyzer(llm_client=llm, graph_engine=engine)


@router.post("/analyze")
def analyze_goal(request: GoalRequest) -> Dict[str, Any]:
    """Analyze user goal text and map to a domain plus prior skills."""
    try:
        result = analyzer.analyze_user_intent(request.user_text)
    except Exception as exc:
        raise HTTPException(status_code=500, detail=str(exc))

    return result
