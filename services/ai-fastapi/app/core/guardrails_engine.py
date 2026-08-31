"""NeMo Guardrails engine — Socratic pedagogical enforcer.

This module wraps NVIDIA NeMo Guardrails to intercept student messages BEFORE
they reach the LLM and enforce strict Socratic tutoring boundaries.

When ``nemoguardrails`` is not installed or ``NVIDIA_API_KEY`` is absent, the
engine falls back to a fast, deterministic keyword-based guardrail that covers
the most common jailbreak patterns with zero latency.

Architecture boundary
---------------------
The guardrails check is applied at the **API layer** (in ``learning.py`` and
``voice.py``) BEFORE calling the LLM evaluator.  This ensures pedagogical
integrity regardless of how the student submits their answer.

Blocked patterns (via Colang + keyword fallback)
-------------------------------------------------
- Direct answer requests: "just give me the answer", "tell me the solution"
- Code generation: "write the code for me", "give me the code"
- Homework completion: "do my assignment", "complete this for me"
"""

from __future__ import annotations

import logging
import os
import re
from pathlib import Path

# ---------------------------------------------------------------------------
# Keyword-based fallback patterns (used when nemoguardrails is unavailable)
# ---------------------------------------------------------------------------

_BLOCK_PATTERNS: list[tuple[str, list[str], str]] = [
    (
        "direct_answer_request",
        [
            r"\bjust\s+(give|tell|show)\s+me\s+the\s+(answer|solution|result)\b",
            r"\bwhat\s+(is|are)\s+the\s+(correct\s+)?answer\b",
            r"\btell\s+me\s+the\s+solution\b",
            r"\bwhat.{0,10}answer\b",
        ],
        (
            "I'm your Socratic tutor — my role is to guide you to the answer, not give it to you. "
            "Let me ask: what do you already know about this concept? Start there and we'll build up together."
        ),
    ),
    (
        "code_generation_request",
        [
            r"\b(write|generate|give me|create|make|produce)\s+.{0,20}(code|script|function|program|solution)\s*(for me|please|now)?\b",
            r"\bjust\s+write\s+(it|the code|the solution)\b",
            r"\bwrite\s+(a\s+)?(python|java|javascript|sql|c\+\+|go|rust)\s+(code|script|function|program)\b",
            r"\bcomplete\s+(the\s+)?(code|function|program)\s*(for me)?\b",
            r"\bfinish\s+the\s+(code|function)\b",
            r"\bgive\s+me\s+(the\s+)?(code|solution|answer)\b",
            r"\bcode\s+(for|to)\s+me\b",
        ],
        (
            "I'm here to help you become a developer, not to be your code generator. "
            "Let's break this down: what's the first logical step you'd take? "
            "Think about the inputs and outputs first — what should the function receive and return?"
        ),
    ),
    (
        "homework_completion_request",
        [
            r"\bdo\s+(my\s+)?(assignment|homework|task|project)\b",
            r"\bcomplete\s+(this\s+)?(for me|assignment|task)\b",
            r"\bsolve\s+this\s+for me\b",
            r"\bwrite\s+my\s+(essay|report|assignment)\b",
        ],
        (
            "Completing assignments for you would rob you of the learning experience. "
            "Instead, let's tackle this together. "
            "What part of the problem statement is unclear? Let's start with the simplest piece."
        ),
    ),
]


def _keyword_check(text: str) -> dict | None:
    """Fast keyword-based guardrail — runs in microseconds, zero API cost."""
    text_lower = text.lower().strip()
    for pattern_name, patterns, refusal_message in _BLOCK_PATTERNS:
        for pattern in patterns:
            if re.search(pattern, text_lower):
                return {
                    "blocked": True,
                    "reason": pattern_name,
                    "refusal_message": refusal_message,
                    "method": "keyword_guardrail",
                }
    return None


# ---------------------------------------------------------------------------
# NeMo Guardrails wrapper
# ---------------------------------------------------------------------------

GUARDRAILS_CONFIG_DIR = Path(__file__).parent / "guardrails"

# Module logger
logger = logging.getLogger(__name__)


class GuardrailsEngine:
    """Socratic pedagogical enforcer powered by NVIDIA NeMo Guardrails.

    Usage::

        engine = GuardrailsEngine()
        result = engine.check("Just write the Python code for me")
        if result["blocked"]:
            return result["refusal_message"]   # send this back to the student
        # else proceed to LLM evaluation

    The engine degrades gracefully:
        - With NVIDIA_API_KEY + nemoguardrails → Full NeMo semantic rail
        - Without either → Fast keyword-based heuristic guardrail
    """

    def __init__(self) -> None:
        self._nemo_rails = None
        self._nemo_loaded = False
        self._nemo_available = False

    def _load_nemo(self) -> None:
        """Lazy-load NeMo Guardrails on first call."""
        if self._nemo_loaded:
            return

        nvidia_key = os.getenv("NVIDIA_API_KEY", "").strip()

        if not nvidia_key:
            self._nemo_available = False
            self._nemo_loaded = True
            return

        try:
            from nemoguardrails import LLMRails, RailsConfig  # type: ignore

            # Inject NVIDIA API key into environment for NeMo's openai client
            os.environ["OPENAI_API_KEY"] = nvidia_key
            os.environ["OPENAI_BASE_URL"] = "https://integrate.api.nvidia.com/v1"

            config = RailsConfig.from_path(str(GUARDRAILS_CONFIG_DIR))
            self._nemo_rails = LLMRails(config)
            self._nemo_available = True

        except (ImportError, OSError, RuntimeError) as exc:
            self._nemo_available = False
            self._nemo_load_error = str(exc)
            logger.exception("Failed to load NeMo Guardrails")

        self._nemo_loaded = True

    def check(self, student_text: str) -> dict:
        """Check a student message against guardrails.

        Args:
            student_text: The student's raw input text.

        Returns:
            A dict with:
                - ``blocked``         — True if the message was blocked
                - ``reason``          — Reason code if blocked (str)
                - ``refusal_message`` — What to send back to the student (str)
                - ``method``          — ``"nemo_guardrails"`` or ``"keyword_guardrail"``
                - ``original_text``   — The input (echoed for logging)
        """
        if not student_text or not student_text.strip():
            return {
                "blocked": False,
                "reason": None,
                "refusal_message": None,
                "method": "passthrough",
                "original_text": student_text,
            }

        # ---- Always run fast keyword check first (zero-latency pre-filter) ----
        keyword_result = _keyword_check(student_text)
        if keyword_result:
            return {**keyword_result, "original_text": student_text}

        # ---- Try NeMo semantic rails (catches paraphrased jailbreaks) ----
        self._load_nemo()

        if self._nemo_available and self._nemo_rails is not None:
            try:
                import asyncio

                async def _async_check():
                    messages = [{"role": "user", "content": student_text}]
                    response = await self._nemo_rails.generate_async(messages=messages)
                    return response

                # Run in a new event loop if not already in one
                try:
                    loop = asyncio.get_event_loop()
                    if loop.is_running():
                        # We're inside FastAPI — schedule as a coroutine
                        import concurrent.futures
                        with concurrent.futures.ThreadPoolExecutor() as pool:
                            future = pool.submit(asyncio.run, _async_check())
                            nemo_response = future.result(timeout=10)
                    else:
                        nemo_response = loop.run_until_complete(_async_check())
                except RuntimeError:
                    nemo_response = asyncio.run(_async_check())

                # NeMo returns the bot message — check if it's a refusal
                response_text = nemo_response if isinstance(nemo_response, str) else str(nemo_response)

                # If NeMo modified the response it means rails triggered
                if response_text and response_text.strip() != student_text.strip():
                    return {
                        "blocked": True,
                        "reason": "nemo_rail_triggered",
                        "refusal_message": response_text,
                        "method": "nemo_guardrails",
                        "original_text": student_text,
                    }

            except (RuntimeError, OSError, asyncio.TimeoutError):
                # NeMo failed — keyword check already passed, allow through
                logger.exception("NeMo guardrails check failed during async run")

        return {
            "blocked": False,
            "reason": None,
            "refusal_message": None,
            "method": "keyword_guardrail" if not self._nemo_available else "nemo_guardrails",
            "original_text": student_text,
        }

    def status(self) -> dict:
        """Return guardrails engine status."""
        self._load_nemo()
        return {
            "nemo_available": self._nemo_available,
            "keyword_fallback": True,
            "config_dir": str(GUARDRAILS_CONFIG_DIR),
            "nvidia_key_present": bool(os.getenv("NVIDIA_API_KEY", "").strip()),
            "method": "nemo_guardrails" if self._nemo_available else "keyword_guardrail",
        }


# ---------------------------------------------------------------------------
# Module-level singleton
# ---------------------------------------------------------------------------
_engine_singleton: GuardrailsEngine | None = None


def get_guardrails_engine() -> GuardrailsEngine:
    """Return the shared ``GuardrailsEngine`` singleton."""
    global _engine_singleton
    if _engine_singleton is None:
        _engine_singleton = GuardrailsEngine()
    return _engine_singleton
