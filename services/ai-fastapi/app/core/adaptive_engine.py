from typing import Dict


class AdaptiveEngine:
    """Deterministic adaptive learning decision engine.

    This engine decides the next action for a learner based on their score
    and the number of failed attempts. No LLMs involved; logic is rule-based
    and fully deterministic.
    """

    def determine_next_action(self, node_id: str, score: float, failed_attempts: int) -> Dict[str, str]:
        """Return the next action for a learner.

        Args:
            node_id: The current learning node identifier (unused in rules but
                included for potential future personalization hooks).
            score: Normalized score between 0.0 and 1.0.
            failed_attempts: Number of times the learner has attempted and failed.

        Returns:
            A dict containing `action` and `message` describing the next step.
        """

        try:
            score_val = float(score)
        except Exception:
            score_val = 0.0

        if score_val >= 0.8:
            return {"action": "advance", "message": "Competency proven. Ready for the next node."}

        if failed_attempts <= 0:
            # treat as first failed attempt
            failed_attempts = 1

        if score_val < 0.8 and failed_attempts == 1:
            return {"action": "remediate", "message": "Let's review the foundational concepts again."}

        # score < 0.8 and failed_attempts >= 2
        return {"action": "human_intervention", "message": "You seem stuck. Let's try a completely different approach or project."}
