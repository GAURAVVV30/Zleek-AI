from app.api.v1 import (
    adaptive,
    evaluate,
    goal,
    guardrails,
    learning,
    mastery,
    recommendation,
    resource,
    roadmap,
    voice,
)
from app.core.llm_client import LLMClient
from fastapi import FastAPI

app = FastAPI(
    title="AI Intelligence Service — Powered by NVIDIA AI Enterprise Stack",
    version="3.0",
    description=(
        "FastAPI Intelligence Service powering the AI Learning Platform. "
        "Backed by NVIDIA NIM (LLaMA 3.1 70B), NeMo Guardrails (Socratic enforcer), "
        "Riva (real-time voice), NV-Embed (RAG), Bayesian Knowledge Tracing (BKT), "
        "and sentiment-adaptive remediation."
    ),
)

# Existing routers
app.include_router(goal.router, prefix="/api/v1/goal", tags=["Goal Understanding"])
app.include_router(roadmap.router, prefix="/api/v1", tags=["Roadmap"])
app.include_router(resource.router, prefix="/api/v1", tags=["Resource"])
app.include_router(evaluate.router, prefix="/api/v1", tags=["Evaluate"])
app.include_router(recommendation.router, prefix="/api/v1/recommendation", tags=["Recommendations"])
app.include_router(learning.router, prefix="/api/v1/learning", tags=["Learning & Evaluation"])
app.include_router(adaptive.router, prefix="/api/v1/adaptive", tags=["Adaptive"])

# D1: BKT Mastery Model
app.include_router(mastery.router, prefix="/api/v1/mastery", tags=["D1: BKT Mastery"])

# D3: AI Voice Tutor (NVIDIA Riva primary, MMS-TTS fallback)
app.include_router(voice.router, prefix="/api/v1/voice", tags=["D3: Voice Tutor"])

# N2: NeMo Guardrails (Socratic Enforcer)
app.include_router(guardrails.router, prefix="/api/v1/guardrails", tags=["N2: NeMo Guardrails"])


@app.get("/")
def root() -> dict[str, str]:
    """Return a minimal welcome payload for the service root."""
    return {
        "message": "AI Intelligence Service v3 — NVIDIA AI Enterprise Stack enabled",
        "stack": "NIM + NeMo Guardrails + Riva + NV-Embed + BKT + Sentiment",
    }


@app.get("/health")
def health_check() -> dict:
    """Return health status including active LLM provider info."""
    llm = LLMClient()
    llm_status = llm.get_status()

    from app.core.voice_engine import get_voice_engine
    voice_status = get_voice_engine().status()

    from app.core.guardrails_engine import get_guardrails_engine
    guard_status = get_guardrails_engine().status()

    return {
        "status": "healthy",
        "service": "ai-fastapi",
        "version": "3.0",
        "llm": llm_status,
        "voice": {
            "asr_provider": voice_status["asr_provider"],
            "tts_provider": voice_status["tts_provider"],
        },
        "guardrails": {
            "method": guard_status["method"],
            "nvidia_key_present": guard_status["nvidia_key_present"],
        },
        "nvidia_enterprise_stack": bool(llm_status.get("provider") == "nvidia_nim"),
    }
