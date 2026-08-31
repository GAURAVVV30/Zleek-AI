from typing import List, Optional, Dict, Any

import json
from pathlib import Path

import logging

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

router = APIRouter()


class Node(BaseModel):
    id: str
    label: str


class Edge(BaseModel):
    source: str
    target: str


class RoadmapRequest(BaseModel):
    domain: Optional[str] = None
    nodes: Optional[List[Node]] = None
    edges: Optional[List[Edge]] = None


SAMPLE_ROADMAP = {
    "domain": "software_architect",
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


# persistent store for submitted roadmaps
APP_DIR = Path(__file__).resolve().parents[2]
DATA_DIR = APP_DIR / "data"
DATA_DIR.mkdir(parents=True, exist_ok=True)
DATA_FILE = DATA_DIR / "roadmaps.json"

logger = logging.getLogger("ai_learning_platform.roadmap")


def _normalize_domain(domain: str) -> str:
    return domain.strip().lower().replace(" ", "_")


def _load_store() -> Dict[str, Any]:
    if not DATA_FILE.exists():
        return {}
    try:
        return json.loads(DATA_FILE.read_text(encoding="utf-8"))
    except Exception:
        logger.exception("Failed to load roadmap store from %s", DATA_FILE)
        return {}


def _save_store(store: Dict[str, Any]) -> None:
    tmp = DATA_FILE.with_suffix(".tmp")
    tmp.write_text(json.dumps(store, indent=2), encoding="utf-8")
    tmp.replace(DATA_FILE)


@router.get("/roadmap")
def get_roadmap(domain: str = "software_architect"):
    """Return a stored roadmap for `domain` if present, otherwise a sample."""
    store = _load_store()
    key = _normalize_domain(domain)
    if key in store:
        return store[key]

    result = SAMPLE_ROADMAP.copy()
    result["domain"] = domain
    return result


@router.get("/roadmap/list")
def list_roadmaps():
    """Debug endpoint: list saved roadmap domain keys."""
    store = _load_store()
    return {"domains": list(store.keys())}


@router.post("/roadmap")
def post_roadmap(payload: RoadmapRequest):
    """Accept a roadmap-like JSON payload, validate it, persist, and return it.

    The submitted roadmap is saved under a normalized domain key and will be
    returned by subsequent GET `/roadmap?domain=...` calls.
    """
    domain = (payload.domain or "custom").strip()
    key = _normalize_domain(domain)

    nodes = [n.dict() for n in (payload.nodes or [])]
    edges = [e.dict() for e in (payload.edges or [])]

    if not nodes and not edges:
        # nothing to persist: return sample but preserve domain
        fallback = SAMPLE_ROADMAP.copy()
        fallback["domain"] = domain
        return fallback

    record = {"domain": domain, "nodes": nodes, "edges": edges}

    store = _load_store()
    store[key] = record
    try:
        _save_store(store)
    except Exception as exc:  # pragma: no cover - runtime guard
        logger.exception("Failed to save roadmap for domain %s: %s", key, exc)
        raise HTTPException(status_code=500, detail="Failed to persist roadmap")

    return record
