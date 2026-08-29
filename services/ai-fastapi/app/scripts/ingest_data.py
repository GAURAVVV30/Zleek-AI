from pathlib import Path
import json

ROOT = Path(__file__).resolve().parent.parent
DOMAIN_DIR = ROOT / "knowledge" / "domains"


def ingest_domain_json(domain: str):
    domain_path = DOMAIN_DIR / domain
    domain_path.mkdir(parents=True, exist_ok=True)

    sample_data = {
        "nodes": [],
        "edges": [],
    }

    for filename in ["skill_graph.json", "resources_rag.json", "assessments.json"]:
        file_path = domain_path / filename
        if not file_path.exists():
            file_path.write_text(json.dumps(sample_data if "skill_graph" in filename or "assessments" in filename else {"resources": []}, indent=2), encoding="utf-8")

    print(f"Created sample metadata for domain: {domain}")


if __name__ == "__main__":
    ingest_domain_json("software_architect")
    ingest_domain_json("machine_learning")
