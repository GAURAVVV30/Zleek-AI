"""AI Voice Tutor API — ASR transcription, TTS synthesis, and full tutor session.

TTS is powered by Microsoft SpeechT5 (local, FREE, no API key needed).
ASR is powered by Groq Whisper (free tier, existing GROQ_API_KEY).

Endpoints
---------
POST /voice/transcribe         — Upload audio → text transcript (Groq Whisper)
POST /voice/synthesize         — Send text → WAV audio (local SpeechT5)
POST /voice/tutor-session      — Full loop: audio → grade → BKT → spoken feedback
GET  /voice/status             — Check ASR/TTS capability availability
"""

import base64
from typing import Any

from app.core.evaluator import AssessmentEvaluator
from app.core.graph_engine import GraphEngine
from app.core.llm_client import LLMClient
from app.core.voice_engine import get_voice_engine
from fastapi import APIRouter, File, Form, HTTPException, UploadFile
from fastapi.responses import Response
from pydantic import BaseModel, Field

router = APIRouter()

# Module-level File() defaults to satisfy ruff B008 (avoid function call in
# argument defaults). Assign once and reuse in route signatures.
_TRANSCRIBE_FILE = File(..., description="Audio file: .wav, .mp3, .webm, .ogg, .flac")
_TUTOR_FILE = File(..., description="Learner's spoken audio answer.")

# ---------------------------------------------------------------------------
# Singleton evaluator (avoids reload on every request)
# ---------------------------------------------------------------------------
_graph_engine: GraphEngine | None = None
_evaluator: AssessmentEvaluator | None = None


def _get_evaluator() -> AssessmentEvaluator:
    global _graph_engine, _evaluator
    if _evaluator is None:
        _graph_engine = GraphEngine()
        _evaluator = AssessmentEvaluator(
            llm_client=LLMClient(),
            graph_engine=_graph_engine,
        )
    return _evaluator


# ---------------------------------------------------------------------------
# Request models
# ---------------------------------------------------------------------------

class SynthesizeRequest(BaseModel):
    text: str = Field(..., description="The text to convert to speech.", max_length=2000)


# ---------------------------------------------------------------------------
# Routes
# ---------------------------------------------------------------------------

@router.get("/status", summary="Check voice capability availability")
def voice_status() -> dict[str, Any]:
    """Return whether ASR (Groq Whisper) and TTS (local SpeechT5) are available.

    Neither requires an OpenAI key — both are fully free.
    """
    engine = get_voice_engine()
    return engine.status()


@router.post("/transcribe", summary="Transcribe audio to text using Groq Whisper")
async def transcribe_audio(
    file: UploadFile = _TRANSCRIBE_FILE,
) -> dict[str, Any]:
    """Convert spoken audio to text using Groq's Whisper API (free tier).

    Uses the existing GROQ_API_KEY — no new credentials required.

    Returns:
        - ``transcript``       — Transcribed text
        - ``duration_seconds`` — Audio duration (if available from API)
        - ``language``         — Detected language
        - ``model``            — ASR model used (whisper-large-v3-turbo)
        - ``filename``         — Original upload filename
    """
    engine = get_voice_engine()

    audio_bytes = await file.read()
    if not audio_bytes:
        raise HTTPException(status_code=400, detail="Uploaded audio file is empty.")

    mime_type = file.content_type or "audio/webm"

    try:
        transcript, meta = engine.transcribe(audio_bytes, mime_type=mime_type)
    except RuntimeError as exc:
        raise HTTPException(status_code=503, detail=str(exc))

    if not transcript:
        raise HTTPException(status_code=422, detail="Whisper returned an empty transcript. Check audio quality.")

    return {
        "transcript": transcript,
        "filename": file.filename,
        **meta,
    }


@router.post("/synthesize", summary="Convert text to WAV audio using local SpeechT5 (free)")
def synthesize_text(request: SynthesizeRequest) -> Response:
    """Convert feedback text to speech using Microsoft SpeechT5.

    Runs entirely locally — no API key required, completely free.
    Returns a WAV audio stream (Content-Type: audio/wav, 16kHz mono).
    """
    engine = get_voice_engine()

    try:
        wav_bytes = engine.synthesize(request.text)
    except RuntimeError as exc:
        raise HTTPException(
            status_code=503,
            detail=str(exc),
        )

    return Response(
        content=wav_bytes,
        media_type="audio/wav",
        headers={"Content-Disposition": "inline; filename=feedback.wav"},
    )


@router.post(
    "/tutor-session",
    summary="Full spoken tutoring loop: audio → grade → BKT → spoken feedback (all free)",
)
async def tutor_session(
    file: UploadFile = _TUTOR_FILE,
    domain_id: str = Form(...),
    node_id: str = Form(...),
    attempt_history: str | None = Form(
        None,
        description='JSON-encoded list of previous binary responses, e.g. "[1, 0, 1]"',
    ),
    return_audio: bool = Form(True),
) -> dict[str, Any]:
    """Execute the full AI voice tutoring loop in a single API call.

    **Completely free — no OpenAI key needed.**

    Flow:
        1. Transcribe learner's spoken audio (Groq Whisper — free tier)
        2. Evaluate transcribed text (LLM + sentiment tone adaptation)
        3. Update BKT mastery probability
        4. Synthesize Socratic feedback as spoken audio (local SpeechT5)

    Returns:
        - ``transcript``     — Transcribed student answer
        - ``evaluation``     — Full evaluation dict (score, feedback, sentiment, bkt)
        - ``audio_base64``   — Base64-encoded WAV of spoken feedback (if return_audio=True)
        - ``audio_mime``     — MIME type of returned audio (audio/wav)
        - ``audio_available``— Whether TTS synthesis succeeded
    """
    engine = get_voice_engine()
    evaluator = _get_evaluator()

    # ---- Step 1: Transcribe ----
    audio_bytes = await file.read()
    if not audio_bytes:
        raise HTTPException(status_code=400, detail="Audio file is empty.")

    mime_type = file.content_type or "audio/webm"
    try:
        transcript, asr_meta = engine.transcribe(audio_bytes, mime_type=mime_type)
    except RuntimeError as exc:
        raise HTTPException(status_code=503, detail=f"ASR failed: {exc}")

    if not transcript:
        raise HTTPException(status_code=422, detail="Empty transcript — check audio quality.")

    # ---- Step 2: Parse optional attempt history ----
    parsed_history: list | None = None
    if attempt_history:
        import json
        try:
            parsed = json.loads(attempt_history)
        except json.JSONDecodeError:
            raise HTTPException(
                status_code=400,
                detail="attempt_history must be valid JSON (e.g. '[1, 0, 1]').",
            )
        if not isinstance(parsed, list):
            raise HTTPException(
                status_code=400,
                detail="attempt_history must be a JSON array of 0/1 integers.",
            )
        parsed_history = parsed

    # ---- Step 3: Evaluate (LLM + sentiment + BKT) ----
    try:
        evaluation = evaluator.evaluate_submission(
            domain_id=domain_id,
            node_id=node_id,
            student_answer=transcript,
            attempt_history=parsed_history,
        )
    except ValueError as exc:
        raise HTTPException(status_code=404, detail=str(exc))

    if evaluation.get("error"):
        raise HTTPException(status_code=502, detail=evaluation["error"])

    # ---- Step 4: Synthesize spoken feedback (local SpeechT5) ----
    audio_base64: str | None = None
    audio_available = False

    if return_audio:
        feedback_text = evaluation.get("feedback", "")
        bkt_info = evaluation.get("bkt", {})
        p_mastery = bkt_info.get("p_mastery", 0)
        mastered = bkt_info.get("mastered", False)

        # Construct a spoken message that includes BKT context
        if mastered:
            spoken = (
                f"Excellent work! You have demonstrated mastery of this concept. "
                f"Your mastery probability is now {round(p_mastery * 100)} percent. "
                f"{feedback_text}"
            )
        else:
            spoken = (
                f"Your current mastery level is {round(p_mastery * 100)} percent. "
                f"{feedback_text}"
            )
            if evaluation.get("remediation_hint"):
                spoken += f" Here is a helpful hint: {evaluation['remediation_hint']}"

        try:
            wav_bytes = engine.synthesize(spoken)
            audio_base64 = base64.b64encode(wav_bytes).decode("utf-8")
            audio_available = True
        except RuntimeError:
            audio_available = False

    return {
        "transcript": transcript,
        "asr_meta": asr_meta,
        "evaluation": evaluation,
        "audio_base64": audio_base64,
        "audio_mime": "audio/wav" if audio_available else None,
        "audio_available": audio_available,
        "tts_provider": "microsoft_speecht5_local",
    }
