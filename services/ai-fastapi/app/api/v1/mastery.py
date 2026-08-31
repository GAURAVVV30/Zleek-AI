"""Mastery estimation API — exposes the BKT model as a standalone endpoint.

This router lets the Go orchestration service (or any client) query the BKT
model directly with a node ID and a full attempt history without going through
the full evaluation pipeline.  It is also the endpoint the Go service calls
when it wants to recalculate mastery from a stored history of binary signals.
"""

from typing import Any

from app.core.bkt_estimator import get_bkt_estimator
from fastapi import APIRouter
from pydantic import BaseModel, Field, field_validator

router = APIRouter()


# ---------------------------------------------------------------------------
# Request / Response models
# ---------------------------------------------------------------------------

class MasteryUpdateRequest(BaseModel):
    """Request payload for a full-history BKT mastery estimate."""

    domain_id: str = Field(..., description="The domain containing the target node.")
    node_id: str = Field(..., description="The skill/node to estimate mastery for.")
    attempt_history: list[int] = Field(
        ...,
        description="Ordered list of binary responses: 1 = correct, 0 = incorrect.",
        min_length=1,
    )
    custom_params: dict[str, float] | None = Field(
        None,
        description=(
            "Optional BKT parameter overrides: p_init, p_learn, p_guess, p_slip. "
            "When omitted, the node's registered params (or defaults) are used."
        ),
    )

    @field_validator("attempt_history")
    @classmethod
    def validate_binary(cls, v: list[int]) -> list[int]:
        for val in v:
            if val not in (0, 1):
                raise ValueError(f"attempt_history must contain only 0 or 1, got {val}")
        return v


class IncrementalMasteryRequest(BaseModel):
    """Request payload for incrementally updating an existing P(mastery)."""

    node_id: str = Field(..., description="The skill/node being updated.")
    current_p_mastery: float = Field(
        ..., ge=0.0, le=1.0, description="The stored P(mastery) from the previous call."
    )
    new_response: int = Field(
        ..., description="Latest binary response: 1 = correct, 0 = incorrect."
    )

    @field_validator("new_response")
    @classmethod
    def validate_binary_response(cls, v: int) -> int:
        if v not in (0, 1):
            raise ValueError(f"new_response must be 0 or 1, got {v}")
        return v


# ---------------------------------------------------------------------------
# Routes
# ---------------------------------------------------------------------------

@router.post("/update", summary="Compute BKT mastery from full attempt history")
def mastery_update(request: MasteryUpdateRequest) -> dict[str, Any]:
    """Compute P(mastery) using Bayesian Knowledge Tracing.

    Accepts the full attempt history for a node and returns the updated
    mastery probability along with the parameters used.

    The model declares mastery when ``P(mastery) >= 0.95``, which is
    mathematically derived from the sequence of correct/incorrect answers
    rather than an arbitrary score cut-off.

    Returns:
        - ``p_mastery``  — Updated probability of mastery (0.0 – 1.0)
        - ``mastered``   — True if p_mastery >= 0.95
        - ``threshold``  — The mastery threshold (0.95)
        - ``attempts``   — Number of attempts processed
        - ``params``     — BKT parameters used for this estimate
        - ``history``    — The attempt history that was processed
        - ``node_id``    — Echo of the requested node ID
        - ``domain_id``  — Echo of the requested domain ID
    """
    estimator = get_bkt_estimator()
    result = estimator.estimate(
        node_id=request.node_id,
        attempt_history=request.attempt_history,
        custom_params=request.custom_params,
    )
    return {"domain_id": request.domain_id, "node_id": request.node_id, **result}


@router.post("/update-incremental", summary="Incrementally update P(mastery) with one new response")
def mastery_update_incremental(request: IncrementalMasteryRequest) -> dict[str, Any]:
    """Update an existing P(mastery) with a single new binary response.

    This is the efficient path for the Go service: instead of replaying the
    entire history, pass the stored ``current_p_mastery`` and just the latest
    binary response.

    Returns:
        Same shape as ``/update`` but with ``attempts=1``.
    """
    estimator = get_bkt_estimator()
    result = estimator.estimate_incremental(
        node_id=request.node_id,
        current_p_mastery=request.current_p_mastery,
        new_response=request.new_response,
    )
    return {"node_id": request.node_id, **result}


@router.get("/params/{node_id}", summary="Get BKT parameters for a node")
def get_mastery_params(node_id: str) -> dict[str, Any]:
    """Return the BKT parameters registered for a given node.

    Falls back to default parameters if the node has no custom registration.
    """
    estimator = get_bkt_estimator()
    params = estimator.get_params(node_id)
    return {"node_id": node_id, "params": params, "source": "registered" if node_id in estimator._skill_params else "default"}
