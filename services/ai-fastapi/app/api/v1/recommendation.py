from __future__ import annotations

from typing import Any

from app.core.graph_engine import GraphEngine
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

"""Recommendation API endpoints for domain-specific personalized learning paths."""

router = APIRouter(tags=["recommendations"])
engine = GraphEngine()


class PersonalizedRoadmapRequest(BaseModel):
    """Request payload for generating a personalized learning roadmap."""

    domain_id: str = Field(..., description="The domain identifier to build a learning path for.")
    completed_nodes: list[str] = Field(
        default_factory=list,
        description="Node IDs already completed or mastered by the learner.",
    )


@router.post("/personalize-roadmap", summary="Build a personalized learning roadmap")
async def personalize_roadmap(request: PersonalizedRoadmapRequest) -> dict[str, Any]:
    """Generate the remaining learning path for a learner in a domain.

    The endpoint loads the selected domain graph from the shared JSON knowledge base,
    computes a topological order, removes completed nodes, and returns the remaining
    ordered learning sequence in a clean JSON response.
    """
    try:
        path = engine.get_personalized_path(request.domain_id, request.completed_nodes)
    except ValueError as exc:
        raise HTTPException(status_code=404, detail=str(exc)) from exc

    return {
        "domain_id": request.domain_id,
        "completed_nodes": request.completed_nodes,
        "path": path,
    }
