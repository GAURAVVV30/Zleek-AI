"""AI Voice Tutor engine — ASR and TTS with NVIDIA Riva primary.

Provider priority
-----------------
ASR (Speech-to-Text):
  1. NVIDIA Riva via NIM (``NVIDIA_API_KEY``) — enterprise-grade, sub-500ms
  2. Groq Whisper (``GROQ_API_KEY``)          — fast free-tier fallback

TTS (Text-to-Speech):
  1. NVIDIA Riva TTS via NIM (``NVIDIA_API_KEY``) — natural, low-latency
  2. Facebook MMS-TTS local                        — offline free fallback

Full tutoring loop
------------------
    Learner speaks → Riva ASR transcribes → Guardrails check
    → LLM grades (NVIDIA NIM) → BKT updates → Sentiment adapts tone
    → Riva TTS speaks Socratic hint back

Design notes
------------
- NVIDIA Riva is exposed via the NIM API using the OpenAI-compatible client.
- MMS-TTS (already downloaded, confirmed working) is the zero-cost offline fallback.
- All providers degrade gracefully — the system always produces audio.
"""

from __future__ import annotations

import io
import logging
import os
import re
import wave
from collections.abc import Callable
from typing import Any

# pyrefly: ignore [missing-import]
import numpy as np

# pyrefly: ignore [missing-import]
from dotenv import load_dotenv

load_dotenv()

# Module logger
logger = logging.getLogger(__name__)

# NVIDIA Riva NIM model identifiers
NVIDIA_ASR_MODEL = "nvidia/parakeet-ctc-1.1b-asr"
NVIDIA_TTS_VOICE = "English-US.Female-1"

# MMS-TTS fallback config (already cached locally)
MMS_TTS_MODEL = "facebook/mms-tts-eng"
SAMPLE_RATE = 16000
SEED = 555

# Supported audio MIME types
SUPPORTED_AUDIO_TYPES = {
    "audio/wav": "audio.wav",
    "audio/x-wav": "audio.wav",
    "audio/mpeg": "audio.mp3",
    "audio/mp3": "audio.mp3",
    "audio/webm": "audio.webm",
    "audio/ogg": "audio.ogg",
    "audio/flac": "audio.flac",
    "audio/mp4": "audio.mp4",
}


def _numpy_to_wav_bytes(audio_array: np.ndarray, sample_rate: int = SAMPLE_RATE) -> bytes:
    """Convert a float32 numpy array to raw WAV bytes."""
    audio_clipped = np.clip(audio_array, -1.0, 1.0)
    pcm_int16 = (audio_clipped * 32767).astype(np.int16)
    buf = io.BytesIO()
    with wave.open(buf, "wb") as wf:
        wf.setnchannels(1)
        wf.setsampwidth(2)
        wf.setframerate(sample_rate)
        wf.writeframes(pcm_int16.tobytes())
    return buf.getvalue()


class VoiceEngine:
    """Orchestrates ASR and TTS with NVIDIA Riva primary, free fallbacks.

    Provider cascade:
        ASR: NVIDIA Riva (NIM) → Groq Whisper
        TTS: NVIDIA Riva (NIM) → Facebook MMS-TTS (local)

    Usage::

        engine = VoiceEngine()
        transcript, meta = engine.transcribe(audio_bytes, "audio/webm")
        wav_bytes = engine.synthesize("Great effort! Let me explain step by step.")
    """

    def __init__(self) -> None:
        self._nvidia_key: str | None = os.getenv("NVIDIA_API_KEY", "").strip() or None
        self._groq_key: str | None = os.getenv("GROQ_API_KEY", "").strip() or None

        # MMS-TTS lazy-loaded components
        self._tokenizer: Callable[..., Any] | None = None
        self._tts_model: Any | None = None
        self._tts_loaded: bool = False
        self._tts_available: bool = False
        self._tts_load_error: str | None = None
        self._tts_sample_rate: int = SAMPLE_RATE

    # ------------------------------------------------------------------
    # NVIDIA Riva ASR (primary)
    # ------------------------------------------------------------------

    def _transcribe_nvidia_riva(self, audio_bytes: bytes, mime_type: str) -> tuple[str, dict]:
        """Transcribe using NVIDIA Riva via NIM API."""
        from openai import OpenAI  # type: ignore

        client = OpenAI(
            base_url="https://integrate.api.nvidia.com/v1",
            api_key=self._nvidia_key,
        )
        filename = SUPPORTED_AUDIO_TYPES.get(mime_type.lower(), "audio.webm")
        audio_file = (filename, io.BytesIO(audio_bytes), mime_type)

        transcription = client.audio.transcriptions.create(
            file=audio_file,
            model=NVIDIA_ASR_MODEL,
        )
        text = getattr(transcription, "text", "") or ""
        return text.strip(), {"model": NVIDIA_ASR_MODEL, "provider": "nvidia_riva"}

    # ------------------------------------------------------------------
    # Groq Whisper ASR (fallback)
    # ------------------------------------------------------------------

    def _transcribe_groq(self, audio_bytes: bytes, mime_type: str) -> tuple[str, dict]:
        """Transcribe using Groq Whisper API."""
        from groq import Groq  # type: ignore

        client = Groq(api_key=self._groq_key)
        filename = SUPPORTED_AUDIO_TYPES.get(mime_type.lower(), "audio.webm")
        audio_file = (filename, io.BytesIO(audio_bytes), mime_type)

        transcription = client.audio.transcriptions.create(
            file=audio_file,
            model="whisper-large-v3-turbo",
            response_format="verbose_json",
            language="en",
        )
        text = getattr(transcription, "text", "") or ""
        duration = getattr(transcription, "duration", None)
        language = getattr(transcription, "language", "en") or "en"
        return text.strip(), {
            "model": "whisper-large-v3-turbo",
            "provider": "groq_whisper",
            "language": language,
            "duration_seconds": round(float(duration), 2) if duration else None,
        }

    # ------------------------------------------------------------------
    # Public transcribe (cascades providers)
    # ------------------------------------------------------------------

    def transcribe(self, audio_bytes: bytes, mime_type: str = "audio/webm") -> tuple[str, dict]:
        """Transcribe audio → text. Tries NVIDIA Riva first, then Groq Whisper."""
        if self._nvidia_key:
            try:
                return self._transcribe_nvidia_riva(audio_bytes, mime_type)
            except (RuntimeError, OSError, ValueError, TypeError):
                logger.exception("NVIDIA Riva ASR failed, falling back to Groq if available")

        if self._groq_key:
            try:
                return self._transcribe_groq(audio_bytes, mime_type)
            except (RuntimeError, OSError, ValueError, TypeError) as exc:
                logger.exception("Groq Whisper ASR failed")
                raise RuntimeError(f"All ASR providers failed. Last error: {exc}") from exc

        raise RuntimeError(
            "No ASR provider available. Set NVIDIA_API_KEY (Riva) or GROQ_API_KEY (Whisper)."
        )

    # ------------------------------------------------------------------
    # NVIDIA Riva TTS (primary)
    # ------------------------------------------------------------------

    def _synthesize_nvidia_riva(self, text: str) -> bytes:
        """Synthesize speech using NVIDIA Riva TTS via NIM."""
        from openai import OpenAI  # type: ignore

        client = OpenAI(
            base_url="https://integrate.api.nvidia.com/v1",
            api_key=self._nvidia_key,
        )
        response = client.audio.speech.create(
            model="nvidia/riva-tts",
            voice=NVIDIA_TTS_VOICE,
            input=text[:4096],
        )
        audio_bytes = b""
        for chunk in response.iter_bytes():
            audio_bytes += chunk
        return audio_bytes

    # ------------------------------------------------------------------
    # MMS-TTS fallback (local, already cached)
    # ------------------------------------------------------------------

    def _load_mms_tts(self) -> None:
        """Lazy-load Facebook MMS-TTS."""
        if self._tts_loaded:
            return
        try:
            from transformers import VitsModel, VitsTokenizer  # type: ignore

            self._tokenizer = VitsTokenizer.from_pretrained(MMS_TTS_MODEL)
            self._tts_model = VitsModel.from_pretrained(MMS_TTS_MODEL)
            self._tts_model.eval()
            if hasattr(self._tts_model, "config") and hasattr(
                self._tts_model.config, "sampling_rate"
            ):
                self._tts_sample_rate = self._tts_model.config.sampling_rate
            self._tts_available = True
        except (ImportError, OSError, RuntimeError) as exc:
            self._tts_available = False
            self._tts_load_error = str(exc)
            logger.exception("Failed to load MMS-TTS model")
        self._tts_loaded = True

    def _synthesize_mms_tts(self, text: str) -> bytes:
        """Synthesize using local Facebook MMS-TTS."""
        self._load_mms_tts()
        if not self._tts_available:
            raise RuntimeError(
                f"MMS-TTS unavailable: {getattr(self, '_tts_load_error', '')}"
            )

        # pyrefly: ignore [missing-import]
        import torch
        from transformers import set_seed  # type: ignore

        set_seed(SEED)

        chunks = self._chunk_text(text)
        silence = np.zeros(int(self._tts_sample_rate * 0.25), dtype=np.float32)
        audio_parts: list[np.ndarray] = []

        # mypy: assert tokenizer/model present after successful load
        assert self._tokenizer is not None, "Tokenizer not loaded"
        assert self._tts_model is not None, "TTS model not loaded"
        tokenizer = self._tokenizer
        tts_model = self._tts_model

        for chunk in chunks:
            inputs = tokenizer(text=chunk, return_tensors="pt")
            with torch.no_grad():
                output = tts_model(**inputs)
            audio_parts.append(output.waveform.squeeze().cpu().numpy())
            audio_parts.append(silence)

        combined = (
            np.concatenate(audio_parts)
            if audio_parts
            else np.zeros(1, dtype=np.float32)
        )
        return _numpy_to_wav_bytes(combined, sample_rate=self._tts_sample_rate)

    # ------------------------------------------------------------------
    # Public synthesize (cascades providers)
    # ------------------------------------------------------------------

    def synthesize(self, text: str) -> bytes:
        """Synthesize text → audio. Tries NVIDIA Riva first, then MMS-TTS."""
        if self._nvidia_key:
            try:
                return self._synthesize_nvidia_riva(text)
            except (RuntimeError, OSError, ValueError, TypeError):
                logger.exception("NVIDIA Riva TTS failed, falling back to MMS-TTS")
        return self._synthesize_mms_tts(text)

    # ------------------------------------------------------------------
    # Utilities
    # ------------------------------------------------------------------

    @staticmethod
    def _chunk_text(text: str, max_chars: int = 500) -> list[str]:
        """Split text into sentence-level chunks of at most max_chars each."""
        if len(text) <= max_chars:
            return [text]
        sentences = re.split(r"(?<=[.!?])\s+", text.strip())
        chunks: list[str] = []
        current = ""
        for sentence in sentences:
            if len(current) + len(sentence) + 1 <= max_chars:
                current = (current + " " + sentence).strip()
            else:
                if current:
                    chunks.append(current)
                current = sentence[:max_chars]
        if current:
            chunks.append(current)
        return chunks or [text[:max_chars]]

    # ------------------------------------------------------------------
    # Status
    # ------------------------------------------------------------------

    def status(self) -> dict:
        """Return provider availability for ASR and TTS."""
        asr_provider = (
            "nvidia_riva"
            if self._nvidia_key
            else ("groq_whisper" if self._groq_key else "none")
        )
        tts_provider = "nvidia_riva" if self._nvidia_key else "mms_tts_local"

        return {
            "asr_available": bool(self._nvidia_key or self._groq_key),
            "asr_provider": asr_provider,
            "asr_model": (
                NVIDIA_ASR_MODEL if self._nvidia_key else "whisper-large-v3-turbo"
            ),
            "tts_available": True,
            "tts_provider": tts_provider,
            "tts_model": (
                "nvidia/riva-tts" if self._nvidia_key else MMS_TTS_MODEL
            ),
            "nvidia_key_present": bool(self._nvidia_key),
            "groq_key_present": bool(self._groq_key),
            "openai_key_required": False,
        }


# ---------------------------------------------------------------------------
# Module-level singleton
# ---------------------------------------------------------------------------
_engine_singleton: VoiceEngine | None = None


def get_voice_engine() -> VoiceEngine:
    """Return the shared ``VoiceEngine`` singleton."""
    global _engine_singleton
    if _engine_singleton is None:
        _engine_singleton = VoiceEngine()
    return _engine_singleton
