"""Deterministic adaptive learning decision engine.

Updated to use the BKT mastery threshold (p_mastery >= 0.95) as the primary
signal for advancement, with the raw LLM score as a fallback.  The engine
remains rule-based and fully deterministic — no LLMs are involved.
"""


from typing import Any

# BKT mastery threshold — must match bkt_estimator.MASTERY_THRESHOLD
BKT_MASTERY_THRESHOLD: float = 0.95

# Legacy score threshold (used when BKT data is unavailable)
LEGACY_SCORE_THRESHOLD: float = 0.80


class AdaptiveEngine:
    """Deterministic adaptive learning decision engine.

    The primary decision signal is the BKT ``p_mastery`` probability.  When
    ``p_mastery`` is not provided (e.g. for backward-compatibility), the engine
    falls back to the raw LLM score compared against ``LEGACY_SCORE_THRESHOLD``.

    Decision rules
    --------------
    - p_mastery >= 0.95 (or score >= 0.80 in legacy mode) → **advance**
    - First failure (failed_attempts == 1) → **remediate**
    - Repeated failure (failed_attempts >= 2) → **human_intervention**
    """

    def determine_next_action(
        self,
        node_id: str,
        score: float,
        failed_attempts: int,
        p_mastery: float | None = None,
    ) -> dict[str, Any]:
        """Return the next action for a learner.

        Args:
            node_id:        The current learning node identifier.
            score:          Normalized LLM score between 0.0 and 1.0.
            failed_attempts: Number of consecutive failed attempts.
            p_mastery:      Optional BKT P(mastery) value.  When provided,
                            this takes precedence over the raw ``score``.

        Returns:
            A dict with ``action`` and ``message`` describing the next step,
            plus ``decision_basis`` indicating which signal was used.
        """

        # ---- Determine mastery using BKT if available ----
        if p_mastery is not None:
            try:
                p_mastery_val = float(p_mastery)
            except (TypeError, ValueError):
                p_mastery_val = 0.0

            if p_mastery_val >= BKT_MASTERY_THRESHOLD:
                return {
                    "action": "advance",
                    "message": (
                        f"Mastery confirmed — P(mastery) = {round(p_mastery_val * 100)}%. "
                        "You're ready for the next concept."
                    ),
                    "decision_basis": "bkt_mastery",
                    "p_mastery": round(p_mastery_val, 4),
                }
        else:
            # ---- Legacy: fall back to raw score ----
            try:
                score_val = float(score)
            except (TypeError, ValueError):
                score_val = 0.0

            if score_val >= LEGACY_SCORE_THRESHOLD:
                return {
                    "action": "advance",
                    "message": "Competency proven. Ready for the next node.",
                    "decision_basis": "legacy_score",
                    "p_mastery": None,
                }

        # ---- Learner has not yet reached mastery ----
        if failed_attempts <= 0:
            failed_attempts = 1

        if failed_attempts == 1:
            return {
                "action": "remediate",
                "message": "Let's review the foundational concepts again.",
                "decision_basis": "bkt_mastery" if p_mastery is not None else "legacy_score",
                "p_mastery": round(float(p_mastery), 4) if p_mastery is not None else None,
            }

        # failed_attempts >= 2
        return {
            "action": "human_intervention",
            "message": "You seem stuck. Let's try a completely different approach or project.",
            "decision_basis": "bkt_mastery" if p_mastery is not None else "legacy_score",
            "p_mastery": round(float(p_mastery), 4) if p_mastery is not None else None,
        }
