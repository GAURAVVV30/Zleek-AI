from fastapi import APIRouter

router = APIRouter()


@router.get("/roadmap")
def get_roadmap(domain: str = "software_architect"):
    return {
        "domain": domain,
        "nodes": [
            {"id": "intro", "label": "Intro to Architecture"},
            {"id": "design", "label": "System Design"},
            {"id": "cloud", "label": "Cloud Patterns"},
            {"id": "security", "label": "Security"},
        ],
        "edges": [
            {"source": "intro", "target": "design"},
            {"source": "design", "target": "cloud"},
            {"source": "cloud", "target": "security"},
        ],
    }
