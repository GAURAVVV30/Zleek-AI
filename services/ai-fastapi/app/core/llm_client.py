"""LLM client — NVIDIA NIM primary, Groq fallback.

Provider priority
-----------------
1. **NVIDIA NIM** (``NVIDIA_API_KEY``) — TensorRT-LLM optimized, enterprise-grade.
   Uses ``meta/llama-3.1-70b-instruct`` via the standard OpenAI SDK pointed at
   ``https://integrate.api.nvidia.com/v1``. Free API credits at build.nvidia.com.

2. **Groq** (``GROQ_API_KEY``) — ultra-fast inference fallback.
   Uses ``llama-3.3-70b-versatile`` via the Groq Python SDK.

3. **Gemini** (``GEMINI_API_KEY``) — final fallback.

The public interface (``generate_text``, ``generate_structured_json``) is
unchanged — all callers (AssessmentEvaluator, GoalAnalyzer, learning.py, etc.)
work without modification.
"""

from __future__ import annotations

import json
import logging
import os
import re
from typing import Any, ClassVar

# pyrefly: ignore [missing-import]
from dotenv import load_dotenv

load_dotenv()

# Module logger
logger = logging.getLogger(__name__)


# ---------------------------------------------------------------------------
# Provider detection
# ---------------------------------------------------------------------------

def _detect_provider() -> tuple[str, str]:
    """Return (provider_name, api_key) for the highest-priority available key."""
    nvidia_key = os.getenv("NVIDIA_API_KEY", "").strip()
    groq_key = os.getenv("GROQ_API_KEY", "").strip()
    gemini_key = os.getenv("GEMINI_API_KEY", "").strip()

    if nvidia_key:
        return "nvidia_nim", nvidia_key
    if groq_key:
        return "groq", groq_key
    if gemini_key:
        return "gemini", gemini_key
    return "none", ""


# ---------------------------------------------------------------------------
# Provider-specific chat helpers
# ---------------------------------------------------------------------------

def _call_nvidia_nim(
    api_key: str,
    model: str,
    system_prompt: str,
    user_prompt: str,
    temperature: float = 0.3,
    max_tokens: int = 1024,
) -> str:
    """Call NVIDIA NIM via the OpenAI-compatible chat completions API."""
    from openai import OpenAI  # type: ignore

    client = OpenAI(
        base_url="https://integrate.api.nvidia.com/v1",
        api_key=api_key,
    )
    response = client.chat.completions.create(
        model=model,
        messages=[
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_prompt},
        ],
        temperature=temperature,
        max_tokens=max_tokens,
    )
    return response.choices[0].message.content or ""


def _call_groq(
    api_key: str,
    model: str,
    system_prompt: str,
    user_prompt: str,
    temperature: float = 0.3,
    max_tokens: int = 1024,
) -> str:
    """Call Groq via its chat completions API."""
    from groq import Groq  # type: ignore

    client = Groq(api_key=api_key)
    response = client.chat.completions.create(
        model=model,
        messages=[
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_prompt},
        ],
        temperature=temperature,
        max_tokens=max_tokens,
    )
    return response.choices[0].message.content or ""


# ---------------------------------------------------------------------------
# Main LLMClient class
# ---------------------------------------------------------------------------

class LLMClient:
    """Unified LLM client with NVIDIA NIM → Groq → Gemini provider cascade.

    Usage::

        llm = LLMClient()
        text = llm.generate_text("You are a tutor.", "Explain recursion.")
        data = llm.generate_structured_json("Be precise.", "List 3 concepts.", {...})
        print(llm.provider)   # "nvidia_nim" | "groq" | "gemini" | "none"
    """

    # Default models per provider
    _DEFAULT_MODELS: ClassVar[dict[str, str]] = {
        "nvidia_nim": "meta/llama-3.2-11b-vision-instruct",
        "groq": "llama-3.3-70b-versatile",
        "gemini": "gemini-2.5-flash",
        "none": "",
    }

    def __init__(self, model: str | None = None) -> None:
        self.provider, self.api_key = _detect_provider()
        self.model = model or self._DEFAULT_MODELS[self.provider]

    # ------------------------------------------------------------------
    # Core text generation
    # ------------------------------------------------------------------

    def generate_text(
        self,
        system_prompt: str,
        user_prompt: str,
        temperature: float = 0.3,
        max_tokens: int = 1024,
    ) -> str:
        """Generate text from the configured LLM provider.

        Args:
            system_prompt: System-level instruction.
            user_prompt:   User prompt or question.
            temperature:   Sampling temperature (0.0 = deterministic).
            max_tokens:    Maximum tokens in the response.

        Returns:
            Generated text string. Returns an error description on failure.
        """
        if self.provider == "none":
            return "LLM unavailable: no API key configured (NVIDIA_API_KEY or GROQ_API_KEY)."

        try:
            if self.provider == "nvidia_nim":
                return _call_nvidia_nim(
                    self.api_key, self.model, system_prompt, user_prompt,
                    temperature, max_tokens,
                )
            if self.provider == "groq":
                return _call_groq(
                    self.api_key, self.model, system_prompt, user_prompt,
                    temperature, max_tokens,
                )
            # Gemini fallback
            return self._call_gemini(system_prompt, user_prompt)
        except (RuntimeError, OSError, ValueError, TypeError) as exc:
            logger.exception("LLM provider call failed")
            # On NVIDIA NIM failure, try Groq as an emergency fallback
            if self.provider == "nvidia_nim":
                groq_key = os.getenv("GROQ_API_KEY", "").strip()
                if groq_key:
                    try:
                        return _call_groq(
                            groq_key,
                            "llama-3.3-70b-versatile",
                            system_prompt,
                            user_prompt,
                            temperature,
                            max_tokens,
                        )
                    except (RuntimeError, OSError, ValueError, TypeError):
                        logger.exception("Groq fallback failed")
            return f"LLM call failed: {exc}"

    def _call_gemini(self, system_prompt: str, user_prompt: str) -> str:
        """Gemini fallback using google-genai SDK."""
        try:
            try:
                # pyrefly: ignore [missing-import]
                from google import genai
                client = genai.Client(api_key=self.api_key)
            except ImportError:
                import google_genai as genai  # type: ignore
                client = genai.Client(api_key=self.api_key)

            if hasattr(client, "generate_content"):
                resp = client.generate_content(
                    model=self.model,
                    contents=f"{system_prompt}\n{user_prompt}",
                )
                return getattr(resp, "text", str(resp))
        except (RuntimeError, OSError, ValueError, TypeError) as exc:
            logger.exception("Gemini provider failed")
            return f"Gemini call failed: {exc}"
        return ""

    # ------------------------------------------------------------------
    # Structured JSON generation
    # ------------------------------------------------------------------

    def generate_structured_json(
        self,
        system_prompt: str,
        user_prompt: str,
        response_schema: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Generate a JSON-structured response and parse it into a Python dict.

        Args:
            system_prompt:   System-level instruction.
            user_prompt:     User-level prompt.
            response_schema: Optional schema hint included in the prompt.

        Returns:
            Parsed Python dict. On failure, returns ``{"error": ..., "raw": ...}``.
        """
        schema_instruction = ""
        if response_schema:
            schema_instruction = (
                f"Respond ONLY with valid JSON matching this schema: "
                f"{json.dumps(response_schema)}\n"
            )

        full_user_prompt = f"{schema_instruction}{user_prompt}"
        raw = self.generate_text(system_prompt, full_user_prompt)

        # Direct parse
        try:
            return json.loads(raw)
        except json.JSONDecodeError:
            # fall through to substring extraction
            pass

        # Extract JSON substring (handles markdown code fences)
        m = re.search(r"\{(?:.|\n)*\}", raw, re.DOTALL)
        if m:
            try:
                return json.loads(m.group(0))
            except json.JSONDecodeError as exc:
                logger.exception("Failed to parse JSON substring from model output")
                return {"error": f"Failed to parse JSON: {exc}", "raw": raw}

        return {"error": "Model did not return valid JSON.", "raw": raw}

    # ------------------------------------------------------------------
    # Status / introspection
    # ------------------------------------------------------------------

    def get_status(self) -> dict[str, Any]:
        """Return current provider info for health checks."""
        return {
            "provider": self.provider,
            "model": self.model,
            "api_key_present": bool(self.api_key),
        }


# ---------------------------------------------------------------------------
# Legacy compatibility helper (used by older code)
# ---------------------------------------------------------------------------

def get_llm_client() -> dict[str, Any]:
    """Compatibility helper — returns status dict for older callers."""
    client = LLMClient()
    return client.get_status()
