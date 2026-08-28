from fastapi import FastAPI

from app.api.v1 import evaluate, goal, recommendation, resource, roadmap, learning, adaptive

app = FastAPI(title="AI Intelligence Service", version="1.0")

app.include_router(goal.router, prefix="/api/v1/goal", tags=["Goal Understanding"])
app.include_router(roadmap.router, prefix="/api/v1", tags=["roadmap"])
app.include_router(resource.router, prefix="/api/v1", tags=["resource"])
app.include_router(evaluate.router, prefix="/api/v1", tags=["evaluate"])
app.include_router(recommendation.router, prefix="/api/v1/recommendation", tags=["Recommendations"])
app.include_router(learning.router, prefix="/api/v1/learning", tags=["Learning & Evaluation"])
app.include_router(adaptive.router, prefix="/api/v1/adaptive", tags=["Adaptive"])


@app.get("/")
def root() -> dict[str, str]:
    """Return a minimal welcome payload for the service root."""
    return {"message": "AI Intelligence Service is running"}


@app.get("/health")
def health_check() -> dict[str, str]:
    """Return the health status for the AI FastAPI service."""
    return {"status": "healthy", "service": "ai-fastapi"}
