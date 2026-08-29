import json
from pathlib import Path


BASE_DIR = Path(__file__).resolve().parent.parent
DOMAIN_PATH = BASE_DIR / "knowledge" / "domains"


def load_rag_resources(domain: str = "software_architect"):
    rag_file = DOMAIN_PATH / domain / "resources_rag.json"
    if not rag_file.exists():
        return {"domain": domain, "resources": []}

    with rag_file.open("r", encoding="utf-8") as f:
        return json.load(f)


class ResourceRetriever:
    """Retrieve resources from a persistent ChromaDB collection.

    This class connects to a local persistent `chromadb.PersistentClient`
    storing data under the `chroma_data` folder at the project root
    (the parent of the `app` directory). It exposes a simple method to
    fetch resources for a given `node_id`.
    """

    def __init__(self, chroma_folder: str = "chroma_data", collection_name: str = "gold_standard_resources"):
        """Initialize the retriever and connect to the ChromaDB collection.

        Args:
            chroma_folder: Relative folder name under the ai-fastapi root where ChromaDB stores persistent files.
            collection_name: The name of the collection to query.
        """
        try:
            import chromadb  # imported lazily so the module can be used even when chromadb isn't installed
        except Exception as exc:
            raise RuntimeError("chromadb is required for ResourceRetriever. Install with 'pip install chromadb'") from exc

        ai_fastapi_root = BASE_DIR.parent
        chroma_dir = ai_fastapi_root / chroma_folder
        chroma_dir.mkdir(parents=True, exist_ok=True)

        # Create a persistent client and obtain the collection.
        self._client = chromadb.PersistentClient(path=str(chroma_dir))
        try:
            self._collection = self._client.get_or_create_collection(name=collection_name)
        except Exception:
            self._collection = self._client.create_collection(name=collection_name)

    def get_resources_for_node(self, node_id: str, n_results: int = 3) -> list[dict]:
        """Return up to `n_results` resources for the given `node_id`.

        The method queries ChromaDB using a `where` filter for an exact
        metadata match on `node_id`. Results are normalized to a list of
        dictionaries containing at least `id`, `title`, `url`, and `provider`.

        Args:
            node_id: The node identifier to filter resources by.
            n_results: Maximum number of resources to return.

        Returns:
            A list of resource dicts: {"id", "title", "url", "provider", "metadata"}.
        """

        # Query by metadata. Use collection.query if available.
        try:
            query_result = self._collection.query(where={"node_id": node_id}, n_results=n_results)
            ids = query_result.get("ids", [])
            metadatas = query_result.get("metadatas", [])
            documents = query_result.get("documents", [])
        except Exception:
            # Fallback: try a simple get with where filter (older/newer SDK differences)
            try:
                get_result = self._collection.get(where={"node_id": node_id})
                ids = get_result.get("ids", [])[:n_results]
                metadatas = get_result.get("metadatas", [])[:n_results]
                documents = get_result.get("documents", [])[:n_results]
            except Exception:
                return []

        results: list[dict] = []
        for idx, doc in enumerate(documents):
            meta = metadatas[idx] if idx < len(metadatas) else {}
            res_id = ids[idx][0] if ids and isinstance(ids[0], list) and idx < len(ids) and ids[idx] else (ids[idx] if idx < len(ids) else None)

            # Document format from ingestion: "<title>\nType: <type>".
            title = doc.split("\n", 1)[0] if isinstance(doc, str) else ""
            results.append({
                "id": res_id,
                "title": title,
                "url": meta.get("url"),
                "provider": meta.get("provider"),
                "metadata": meta,
            })

        return results
