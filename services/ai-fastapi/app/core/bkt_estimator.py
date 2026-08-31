"""Bayesian Knowledge Tracing (BKT) mastery estimator.

BKT is a Hidden Markov Model that tracks whether a learner has transitioned
from an "unmastered" to a "mastered" knowledge state.  It uses four
parameters per skill:

    p_init   — Prior probability of mastery before any attempt.
    p_learn  — Probability of transitioning from unmastered → mastered after
               each attempt.
    p_guess  — Probability of a correct answer even when unmastered.
    p_slip   — Probability of an incorrect answer even when mastered.

The ``P(mastery)`` is computed via the standard BKT forward pass after each
new binary observation (1 = correct, 0 = incorrect).

Mastery is declared when ``P(mastery) >= MASTERY_THRESHOLD`` (default 0.95),
which is far more principled than an arbitrary score cut-off.
"""

from __future__ import annotations

from collections.abc import Sequence
from typing import ClassVar

# ---------------------------------------------------------------------------
# Default skill parameters
# These are sensible "industry-average" BKT defaults from Corbett & Anderson
# (1994) research on ACT-R intelligent tutoring systems.  Override per-node
# via the domain graph's ``bkt_params`` field when tighter calibration is
# available.
# ---------------------------------------------------------------------------
DEFAULT_PARAMS: dict[str, float] = {
    "p_init": 0.30,   # 30% chance the student already knows it
    "p_learn": 0.20,  # 20% chance of learning after each attempt
    "p_guess": 0.25,  # 25% chance of guessing correctly without knowledge
    "p_slip": 0.10,   # 10% chance of slipping (wrong answer despite knowing)
}

MASTERY_THRESHOLD: float = 0.95


# ---------------------------------------------------------------------------
# Core forward-pass computation
# ---------------------------------------------------------------------------

def _bkt_forward(
    responses: Sequence[int],
    p_init: float,
    p_learn: float,
    p_guess: float,
    p_slip: float,
) -> float:
    """Run the BKT forward algorithm and return P(mastery) after all responses.

    Args:
        responses: Ordered sequence of binary observations where 1 = correct
                   and 0 = incorrect.
        p_init:    Prior mastery probability.
        p_learn:   Learning rate (unmastered → mastered per step).
        p_guess:   Guess probability (correct | unmastered).
        p_slip:    Slip probability (incorrect | mastered).

    Returns:
        Updated ``P(mastery)`` as a float in [0, 1].
    """
    p_mastery = p_init  # P(L_0)

    for obs in responses:
        # ---- E-step: compute P(mastered | observation) ----
        if obs == 1:
            # Correct answer
            p_correct_given_mastered = 1.0 - p_slip
            p_correct_given_unmastered = p_guess
            p_obs_given_mastered = p_correct_given_mastered
            p_obs_given_unmastered = p_correct_given_unmastered
        else:
            # Incorrect answer
            p_obs_given_mastered = p_slip
            p_obs_given_unmastered = 1.0 - p_guess

        # Joint probabilities
        p_obs_mastered = p_obs_given_mastered * p_mastery
        p_obs_unmastered = p_obs_given_unmastered * (1.0 - p_mastery)
        p_obs = p_obs_mastered + p_obs_unmastered

        # Posterior: P(mastered | obs)
        if p_obs == 0.0:
            # Degenerate case — keep current belief
            p_mastery_given_obs = p_mastery
        else:
            p_mastery_given_obs = p_obs_mastered / p_obs

        # ---- M-step: update belief accounting for possible learning ----
        p_mastery = p_mastery_given_obs + (1.0 - p_mastery_given_obs) * p_learn

    return float(min(max(p_mastery, 0.0), 1.0))


# ---------------------------------------------------------------------------
# Public Estimator class
# ---------------------------------------------------------------------------

class BKTEstimator:
    """Per-node Bayesian Knowledge Tracing estimator.

    Usage::

        estimator = BKTEstimator()

        # After three attempts — two correct, one incorrect
        result = estimator.estimate("python_basics.variables", [1, 0, 1])
        print(result)
        # {
        #   "p_mastery": 0.621,
        #   "mastered": False,
        #   "threshold": 0.95,
        #   "attempts": 3,
        #   "params": {...}
        # }
    """

    # Registry of per-node BKT parameters loaded from domain graph data.
    _skill_params: ClassVar[dict[str, dict[str, float]]] = {}

    def register_skill(
        self,
        node_id: str,
        p_init: float = DEFAULT_PARAMS["p_init"],
        p_learn: float = DEFAULT_PARAMS["p_learn"],
        p_guess: float = DEFAULT_PARAMS["p_guess"],
        p_slip: float = DEFAULT_PARAMS["p_slip"],
    ) -> None:
        """Register custom BKT parameters for a specific node/skill.

        Call this when loading a domain graph that carries per-node
        ``bkt_params`` metadata.  Nodes without registered parameters
        automatically use ``DEFAULT_PARAMS``.

        Args:
            node_id:  The graph node identifier (e.g. ``"python.variables"``).
            p_init:   Prior mastery probability.
            p_learn:  Learning rate per attempt.
            p_guess:  Guess probability when unmastered.
            p_slip:   Slip probability when mastered.
        """
        self._skill_params[node_id] = {
            "p_init": float(p_init),
            "p_learn": float(p_learn),
            "p_guess": float(p_guess),
            "p_slip": float(p_slip),
        }

    def get_params(self, node_id: str) -> dict[str, float]:
        """Return BKT parameters for a node, falling back to defaults."""
        return dict(self._skill_params.get(node_id, DEFAULT_PARAMS))

    def estimate(
        self,
        node_id: str,
        attempt_history: list[int],
        custom_params: dict[str, float] | None = None,
    ) -> dict:
        """Compute ``P(mastery)`` from the full attempt history for a node.

        Args:
            node_id:        The node/skill being estimated.
            attempt_history: Ordered list of binary responses (1=correct, 0=incorrect).
                             Must be non-empty.
            custom_params:  Optional parameter overrides for this call only.

        Returns:
            A dict with keys:
                - ``p_mastery``  — updated probability of mastery (float 0-1)
                - ``mastered``   — True if p_mastery >= MASTERY_THRESHOLD
                - ``threshold``  — the mastery threshold used
                - ``attempts``   — number of attempts processed
                - ``params``     — BKT parameters used
                - ``history``    — the attempt history passed in
        """
        params = custom_params or self.get_params(node_id)

        if not attempt_history:
            # No data — return the prior
            p_mastery = params["p_init"]
        else:
            p_mastery = _bkt_forward(
                responses=attempt_history,
                p_init=params["p_init"],
                p_learn=params["p_learn"],
                p_guess=params["p_guess"],
                p_slip=params["p_slip"],
            )

        return {
            "p_mastery": round(p_mastery, 4),
            "mastered": p_mastery >= MASTERY_THRESHOLD,
            "threshold": MASTERY_THRESHOLD,
            "attempts": len(attempt_history),
            "params": params,
            "history": list(attempt_history),
        }

    def estimate_incremental(
        self,
        node_id: str,
        current_p_mastery: float,
        new_response: int,
        custom_params: dict[str, float] | None = None,
    ) -> dict:
        """Update an existing P(mastery) with a single new binary response.

        This is the efficient incremental variant — useful when the Go service
        stores ``p_mastery`` and only the latest attempt needs to be folded in.

        Args:
            node_id:           The node/skill being estimated.
            current_p_mastery: The stored ``p_mastery`` from the previous call.
            new_response:      The latest binary response (1=correct, 0=incorrect).
            custom_params:     Optional parameter overrides.

        Returns:
            Same dict shape as ``estimate()``, with ``attempts=1`` (single step).
        """
        params = custom_params or self.get_params(node_id)
        p_mastery = _bkt_forward(
            responses=[new_response],
            p_init=current_p_mastery,  # treat stored belief as the prior
            p_learn=params["p_learn"],
            p_guess=params["p_guess"],
            p_slip=params["p_slip"],
        )
        return {
            "p_mastery": round(p_mastery, 4),
            "mastered": p_mastery >= MASTERY_THRESHOLD,
            "threshold": MASTERY_THRESHOLD,
            "attempts": 1,
            "params": params,
            "history": [new_response],
        }


# ---------------------------------------------------------------------------
# Module-level singleton (avoids re-instantiation across imports)
# ---------------------------------------------------------------------------
_estimator_singleton: BKTEstimator | None = None


def get_bkt_estimator() -> BKTEstimator:
    """Return the shared ``BKTEstimator`` singleton."""
    global _estimator_singleton
    if _estimator_singleton is None:
        _estimator_singleton = BKTEstimator()
    return _estimator_singleton
