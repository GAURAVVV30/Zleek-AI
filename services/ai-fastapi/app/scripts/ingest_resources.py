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

    domains_dir = ai_fastapi_dir / "app" / "knowledge" / "domains"
    if not domains_dir.exists():
        print(f"Domains folder not found at {domains_dir}")
        return 0

    # Delay importing chromadb until after verifying domains exist to make
    # the script easier to run for scanning/debugging when chromadb isn't installed.
    try:
        import chromadb
    except ImportError:
        logger.warning(
            "chromadb package not available. Install with 'pip install chromadb' to perform ingestion."
        )
        return 0

    # Initialize persistent ChromaDB client
    client = chromadb.PersistentClient(path=str(chroma_dir))

    # Get or create collection
    try:
        collection = client.get_or_create_collection(name="gold_standard_resources")
    except (AttributeError, TypeError) as exc:
        # Fallback if the installed chromadb client uses a different API
        logger.debug("get_or_create_collection not available: %s", exc)
        try:
            collection = client.create_collection(name="gold_standard_resources")
        except Exception:
            logger.exception("Failed to create or obtain 'gold_standard_resources' collection")
            return 0

    total = 0

    for graph_file in sorted(domains_dir.rglob("*_graph.json")):
        domain_id = graph_file.parent.name
        try:
            with open(graph_file, "r", encoding="utf-8") as fh:
                data = json.load(fh)
        except (json.JSONDecodeError, OSError):
            logger.exception("Failed to load %s", graph_file)
            continue

        nodes = data.get("nodes") or data.get("graph") or []
        if isinstance(nodes, dict):
            nodes = nodes.get("nodes", [])

        for n_idx, node in enumerate(nodes):
            node_id = node.get("id") or node.get("node_id") or f"{domain_id}_node_{n_idx}"
            node_name = node.get("name") or node.get("title") or ""
            resources = node.get("resources", []) or []

            for r_idx, res in enumerate(resources):
                res_id = f"{node_id}_res_{r_idx}"
                title = res.get("title") or res.get("name") or ""
                rtype = res.get("type") or res.get("resource_type") or ""
                document = f"{title}\nType: {rtype}".strip()

                metadata = {
                    "domain_id": domain_id,
                    "node_id": node_id,
                    "node_name": node_name,
                    "url": res.get("url"),
                    "provider": res.get("provider"),
                    "authority_score": res.get("authority_score"),
                }

                try:
                    collection.add(documents=[document], metadatas=[metadata], ids=[res_id])
                    total += 1
                except (TypeError, AttributeError) as e:
                    # Signature mismatch between chromadb versions; try client-level API
                    logger.debug("collection.add failed, trying client.add: %s", e)
                    add_func = getattr(client, "add", None)
                    if callable(add_func):
                        try:
                            add_func(
                                collection_name="gold_standard_resources",
                                documents=[document],
                                metadatas=[metadata],
                                ids=[res_id],
                            )
                            total += 1
                        except Exception:
                            logger.exception("Failed to add resource %s from %s", res_id, graph_file)
                    else:
                        logger.error(
                            "chromadb client has no 'add' method; cannot add resource %s from %s",
                            res_id,
                            graph_file,
                        )
                except Exception:
                    logger.exception("Unexpected error adding resource %s from %s", res_id, graph_file)

    logger.info("Ingested %d resources into 'gold_standard_resources' at %s", total, chroma_dir)
    return total


if __name__ == "__main__":
    count = ingest()
    if count:
        logger.info("Ingestion completed successfully.")
    else:
        logger.info("No resources were ingested.")
