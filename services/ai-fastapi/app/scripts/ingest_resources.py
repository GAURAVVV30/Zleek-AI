#!/usr/bin/env python3
"""Ingest resources from domain graph JSON files into a persistent ChromaDB collection.

Usage: run this script from the ai-fastapi project or directly with Python.
"""
import json
import logging
from pathlib import Path

logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO)


def find_ai_fastapi_root(script_path: Path) -> Path:
    # Walk up parents to find the ai-fastapi directory name, fallback to parents[2]
    p = script_path.resolve()
    for parent in p.parents:
        if parent.name == "ai-fastapi":
            return parent
    # Fallback: assume layout services/ai-fastapi/app/scripts
    return script_path.resolve().parents[2]


def ingest(chroma_folder_name: str = "chroma_data") -> int:
    script_path = Path(__file__)
    ai_fastapi_dir = find_ai_fastapi_root(script_path)

    chroma_dir = ai_fastapi_dir / chroma_folder_name
    chroma_dir.mkdir(parents=True, exist_ok=True)

    gold_json_file = ai_fastapi_dir / "app" / "knowledge" / "gold_tier_resources.json"
    if not gold_json_file.exists():
        logger.error(f"Gold Tier dataset not found at {gold_json_file}")
        return 0

    try:
        import chromadb
    except ImportError:
        logger.warning("chromadb package not available.")
        return 0

    client = chromadb.PersistentClient(path=str(chroma_dir))
    try:
        collection = client.get_or_create_collection(name="gold_standard_resources")
    except Exception:
        collection = client.create_collection(name="gold_standard_resources")

    total = 0
    with open(gold_json_file, "r", encoding="utf-8") as fh:
        dataset = json.load(fh)

    for role_id, role_data in dataset.items():
        if role_id == "data_engineer":
            continue
        role_name = role_data.get("role_name", role_id)
        for module in role_data.get("modules", []):
            module_id = module.get("module_id")
            module_number = module.get("module_number")
            module_name = module.get("module_name")
            resources = module.get("resources", {})

            for rtype in ["video", "documentation", "hands_on"]:
                r_list = resources.get(rtype, [])
                for idx, res in enumerate(r_list):
                    res_id = res.get("id") or f"{module_id}_{rtype}_{idx}"
                    title = res.get("title") or ""
                    desc = res.get("description") or ""
                    url = res.get("url") or ""
                    doc = f"{title}\nCategory: {rtype}\nDescription: {desc}\nURL: {url}".strip()

                    metadata = {
                        "domain_id": role_id,
                        "role": role_name,
                        "node_id": module_id,
                        "module_number": module_number,
                        "module_name": module_name,
                        "resource_type": rtype,
                        "title": title,
                        "description": desc,
                        "url": url,
                        "provider": res.get("provider", ""),
                        "authority_score": 1.0,
                    }

                    try:
                        collection.add(documents=[doc], metadatas=[metadata], ids=[res_id])
                        total += 1
                    except Exception as e:
                        logger.debug("Failed adding resource %s: %s", res_id, e)

    logger.info("Ingested %d resources into 'gold_standard_resources' at %s", total, chroma_dir)
    return total


if __name__ == "__main__":
    count = ingest()
    if count:
        logger.info("Ingestion completed successfully.")
    else:
        logger.info("No resources were ingested.")

