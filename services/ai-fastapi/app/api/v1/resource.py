from fastapi import APIRouter, HTTPException

router = APIRouter()


@router.get("/resource")
def get_resource(domain: str = "software_architect", topic: str = "system design"):
    if not topic:
        raise HTTPException(status_code=400, detail="topic is required")

    return {
        "domain": domain,
        "topic": topic,
        "resource": {
            "title": "System Design Primer",
            "type": "tutorial",
            "url": "https://github.com/donnemartin/system-design-primer",
            "summary": "A curated guide for architecture, scalability, and design trade-offs.",
        },
    }
