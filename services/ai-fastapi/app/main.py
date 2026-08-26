from fastapi import FastAPI

app = FastAPI(title="AI Service")

@app.get("/v1/health")
def health_check():
    return {"status": "ok"}
