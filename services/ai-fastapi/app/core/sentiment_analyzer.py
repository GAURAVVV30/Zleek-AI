"""Sentiment and engagement analysis for adaptive remediation.

Uses the ``j-hartmann/emotion-english-distilroberta-base`` HuggingFace model
to classify the emotional state of a learner's free-text answer into one of
six Ekman emotions: anger, disgust, fear, joy, sadness, surprise, neutral.

The model is lazy-loaded on first use so FastAPI startup remains fast.
If ``transformers`` / ``torch`` are not installed, the module falls back to a
keyword-based heuristic that approximates the same emotion labels — ensuring
the system degrades gracefully in CPU-only or minimal-dependency environments.

Integration point
-----------------
``AssessmentEvaluator.evaluate_submission()`` calls ``analyze()`` after every
failed mastery check and uses the returned ``tone_override`` to modify the
LLM remediation prompt accordingly.
"""

from __future__ import annotations

import logging
import re
from collections.abc import Callable
from typing import Any, cast

logger = logging.getLogger(__name__)

# ---------------------------------------------------------------------------
# Emotion → Tone mapping
# Frustration-adjacent emotions trigger the encouraging remediation tone.
# ---------------------------------------------------------------------------
FRUSTRATION_EMOTIONS = {"anger", "disgust", "fear"}
POSITIVE_EMOTIONS = {"joy", "surprise"}

# Simple keyword heuristics used when transformers is unavailable.
_KEYWORD_PATTERNS: dict[str, list[str]] = {
    "anger": [
        r"\bstupid\b", r"\bdumb\b", r"\bconfusing\b", r"\bconfused\b",
        r"\bfrustrat\w*\b", r"\bawful\b", r"\bwrong\b", r"\bhate\b",
        r"\bcan'?t\b.{0,20}\bunderstand\b", r"\bmakes no sense\b",
    ],
    "sadness": [
        r"\bgive up\b", r"\bcan'?t do\b", r"\bimpossible\b", r"\bgave up\b",
        r"\bdisappoin\w*\b", r"\btoo hard\b",
    ],
    "fear": [
        r"\bscared\b", r"\bnervous\b", r"\banxious\b", r"\bworried\b",
        r"\bnot sure\b", r"\bnot confident\b",
    ],
    "joy": [
        r"\bgot it\b", r"\bunderstand\b", r"\bclear\b", r"\beasy\b",
        r"\bmakes sense\b", r"\bgreat\b", r"\bexcit\w*\b",
    ],
    "neutral": [],
}


def _keyword_analyze(text: str) -> dict[str, object]:
    """Lightweight keyword-based fallback emotion detector."""
    text_lower = text.lower()
    for emotion, patterns in _KEYWORD_PATTERNS.items():
        if emotion == "neutral":
            continue
        for pattern in patterns:
            if re.search(pattern, text_lower):
                return {
                    "emotion": emotion,
                    "confidence": 0.70,  # fixed moderate confidence for heuristic
                    "method": "keyword_heuristic",
                }
    return {"emotion": "neutral", "confidence": 0.90, "method": "keyword_heuristic"}


# ---------------------------------------------------------------------------
# Main analyzer
# ---------------------------------------------------------------------------

class SentimentAnalyzer:
    """Emotion classification pipeline with graceful degradation.

    Usage::

        analyzer = SentimentAnalyzer()
        result = analyzer.analyze("I have no idea what's happening, this is so confusing!")
        # {
        #   "emotion": "anger",
        #   "confidence": 0.88,
        #   "tone_override": "encouraging",
        #   "method": "transformer",
        #   "raw_label": "anger"
        # }
    """

    def __init__(self) -> None:
        self._pipeline: Callable[..., Any] | None = None
        self._pipeline_loaded: bool = False
        self._use_heuristic: bool = False

    def _load_pipeline(self) -> None:
        """Lazy-load the HuggingFace transformer pipeline on first call."""
        if self._pipeline_loaded:
            return

        try:
            from transformers import pipeline  # type: ignore

            loaded = pipeline(
                "text-classification",
                model="j-hartmann/emotion-english-distilroberta-base",
                top_k=1,
                truncation=True,
                max_length=512,
            )
            # Cast to a generic callable for static typing friendliness
            self._pipeline = cast(Callable[..., Any], loaded)
            self._use_heuristic = False
        except (ImportError, OSError):
            # transformers/torch not available or model files not present — fall back to heuristics
            self._pipeline = None
            self._use_heuristic = True
            logger.exception("Transformer pipeline unavailable, using keyword heuristic")

        self._pipeline_loaded = True

    def analyze(self, text: str) -> dict[str, object]:
        """Classify the dominant emotion in ``text``.

        Args:
            text: The learner's free-text input (answer, comment, or spoken
                  transcription). Ideally 1-3 sentences.

        Returns:
            A dict with:
                - ``emotion``      — dominant emotion label (str, lowercase)
                - ``confidence``   — model confidence 0–1 (float)
                - ``tone_override``— ``"encouraging"`` if frustration detected,
                                     ``"standard"`` otherwise
                - ``method``       — ``"transformer"`` or ``"keyword_heuristic"``
                - ``raw_label``    — original label from model/heuristic
        """
        if not text or not text.strip():
            return {
                "emotion": "neutral",
                "confidence": 1.0,
                "tone_override": "standard",
                "method": "empty_input",
                "raw_label": "neutral",
            }

        self._load_pipeline()

        if self._use_heuristic or self._pipeline is None:
            result = _keyword_analyze(text)
        else:
            try:
                assert self._pipeline is not None
                raw = self._pipeline(text)
                # pipeline returns [[{label, score}]] with top_k=1
                best = raw[0][0] if isinstance(raw[0], list) else raw[0]
                result = {
                    "emotion": best["label"].lower(),
                    "confidence": round(float(best["score"]), 4),
                    "method": "transformer",
                }
            except (RuntimeError, ValueError, TypeError, IndexError, KeyError, OSError):
                logger.exception("Transformer pipeline failed during analysis")
                # If the transformer pipeline fails at runtime, fall back to heuristics
                result = _keyword_analyze(text)

        emotion = result.get("emotion", "neutral")
        tone = "encouraging" if emotion in FRUSTRATION_EMOTIONS else "standard"

        return {
            "emotion": emotion,
            "confidence": result.get("confidence", 0.0),
            "tone_override": tone,
            "method": result.get("method", "unknown"),
            "raw_label": emotion,
        }


# ---------------------------------------------------------------------------
# Prompt injection helper
# ---------------------------------------------------------------------------

def build_tone_system_prompt(base_system_prompt: str, tone_override: str) -> str:
    """Prepend a tone instruction to a system prompt based on detected emotion.

    Args:
        base_system_prompt: The original system prompt for the LLM evaluator.
        tone_override:      ``"encouraging"`` or ``"standard"``.

    Returns:
        Modified system prompt with tone guidance prepended.
    """
    if tone_override != "encouraging":
        return base_system_prompt

    tone_prefix = (
        "IMPORTANT: The learner appears frustrated or confused. "
        "Your response must be exceptionally patient, warm, and encouraging. "
        "Break down the concept into the smallest possible steps. "
        "Celebrate any partial understanding. "
        "Avoid technical jargon — use relatable analogies instead. "
        "Start your feedback with a genuine, uplifting acknowledgement of their effort.\n\n"
    )
    return tone_prefix + base_system_prompt


# ---------------------------------------------------------------------------
# Module-level singleton
# ---------------------------------------------------------------------------
_analyzer_singleton: SentimentAnalyzer | None = None


def get_sentiment_analyzer() -> SentimentAnalyzer:
    """Return the shared ``SentimentAnalyzer`` singleton."""
    global _analyzer_singleton
    if _analyzer_singleton is None:
        _analyzer_singleton = SentimentAnalyzer()
    return _analyzer_singleton
