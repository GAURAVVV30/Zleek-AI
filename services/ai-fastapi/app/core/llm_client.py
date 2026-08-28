import json
import os
import re
from typing import Any, Dict, Optional

from dotenv import load_dotenv

load_dotenv()


class LLMClient:
    """Lightweight client wrapper for Groq or Gemini providers.

    The client will prefer `GROQ_API_KEY` if present, falling back to
    `GEMINI_API_KEY`. Imports are performed lazily so the module can be
    imported in environments without the provider SDKs installed.
    """

    def __init__(self, model: Optional[str] = None):
        groq_key = os.getenv("GROQ_API_KEY")
        gemini_key = os.getenv("GEMINI_API_KEY")

        self.provider: Optional[str] = None
        self.api_key: Optional[str] = None
        self.client: Optional[Any] = None

        if groq_key:
            self.provider = "groq"
            self.api_key = groq_key
        elif gemini_key:
            self.provider = "gemini"
            self.api_key = gemini_key

        # Choose sensible default model per provider
        if self.provider == "groq":
            self.model = model or "llama-3.3-70b-versatile"
        else:
            self.model = model or "gemini-2.5-flash"

        if self.api_key:
            try:
                if self.provider == "groq":
                    import groq

                    # Example client init; actual SDK may differ.
                    self.client = groq.Client(api_key=self.api_key)
                else:
                    # google-genai
                    import google_genai as genai

                    # If package is named `google_genai` or `google.genai`, adapt accordingly.
                    try:
                        self.client = genai.Client(api_key=self.api_key)
                    except Exception:
                        # try alternative import name
                        import google.genai as genai2

                        self.client = genai2.Client(api_key=self.api_key)
            except Exception:
                # Leave client as None; methods will raise informative errors.
                self.client = None

    def _ensure_client(self) -> None:
        if not self.api_key:
            raise RuntimeError("No LLM API key found. Set GROQ_API_KEY or GEMINI_API_KEY in the environment.")
        if not self.client:
            raise RuntimeError("LLM SDK not available or client failed to initialize. Install provider SDK.")

    def generate_text(self, system_prompt: str, user_prompt: str, temperature: float = 0.3) -> str:
        """Generate text from the selected LLM.

        Args:
            system_prompt: System-level instruction guiding the model.
            user_prompt: User prompt or question.
            temperature: Sampling temperature.

        Returns:
            Generated text output from the model.
        """

        try:
            self._ensure_client()
        except RuntimeError as exc:
            return f"LLM unavailable: {exc}"

        # Prefer provider-specific generate APIs with graceful fallbacks
        try:
            if self.provider == "groq":
                # groq.Client usage may differ; this is a conservative pattern
                response = self.client.generate(
                    model=self.model,
                    prompt=f"System: {system_prompt}\nUser: {user_prompt}",
                    temperature=temperature,
                )
                # Attempt to extract text from common response shapes
                return getattr(response, "text", str(response))

            else:
                # google genai style
                # attempt different client method names defensively
                if hasattr(self.client, "generate_text"):
                    resp = self.client.generate_text(model=self.model, prompt=f"{system_prompt}\n{user_prompt}")
                    return getattr(resp, "text", str(resp))
                elif hasattr(self.client, "generate"):
                    resp = self.client.generate(model=self.model, prompt=f"{system_prompt}\n{user_prompt}")
                    return getattr(resp, "text", str(resp))

            return ""
        except Exception as exc:
            return f"LLM call failed: {exc}"

    def generate_structured_json(self, system_prompt: str, user_prompt: str, response_schema: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Generate a JSON-structured response and parse it into a Python dict.

        The method instructs the model to reply with strict JSON matching
        `response_schema` when provided. It tries to parse the model's
        reply and will attempt to extract embedded JSON if necessary.

        Args:
            system_prompt: System-level instruction.
            user_prompt: User-level prompt.
            response_schema: Optional JSON schema-like dict used only for prompting.

        Returns:
            A Python dict parsed from the model's JSON output. On failure,
            returns a dict with an `error` field describing the issue.
        """

        schema_instruction = ""
        if response_schema:
            schema_instruction = f"Respond only with JSON matching this schema: {json.dumps(response_schema)}\n"

        full_prompt = f"{schema_instruction}System: {system_prompt}\nUser: {user_prompt}"

        raw = self.generate_text(system_prompt, f"{schema_instruction}{user_prompt}")

        # Try direct JSON parse
        try:
            return json.loads(raw)
        except Exception:
            # Attempt to extract JSON substring
            m = re.search(r"\{(?:.|\n)*\}", raw)
            if m:
                try:
                    return json.loads(m.group(0))
                except Exception as exc:
                    return {"error": f"Failed to parse JSON: {exc}", "raw": raw}

        return {"error": "Model did not return valid JSON.", "raw": raw}


def get_llm_client() -> Dict[str, Any]:
    """Compatibility helper returning a short status dict for older code.

    Returns whether an API key is present and the default model name.
    """
    groq_key = os.getenv("GROQ_API_KEY")
    gemini_key = os.getenv("GEMINI_API_KEY")
    provider = "groq" if groq_key else ("gemini" if gemini_key else "none")
    model = "llama-3.3-70b-versatile" if provider == "groq" else ("gemini-2.5-flash" if provider == "gemini" else "")
    return {"provider": provider, "api_key_present": bool(groq_key or gemini_key), "model": model}
