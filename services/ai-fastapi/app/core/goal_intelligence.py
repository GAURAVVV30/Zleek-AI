from typing import Any

from app.core.graph_engine import GraphEngine
from app.core.llm_client import LLMClient


class GoalAnalyzer:
    """Map a user's career goal text to a domain and extract prior skills.

    Uses `LLMClient.generate_structured_json` to produce a deterministic
    JSON object containing `mapped_domain_id`, `extracted_prior_skills`,
    and `reasoning`.
    """

    def __init__(self, llm_client: LLMClient, graph_engine: GraphEngine) -> None:
        self.llm = llm_client
        self.graph = graph_engine

    def analyze_user_intent(self, user_text: str) -> dict[str, Any]:
        """Analyze free-form user intent and map to a domain.

        Args:
            user_text: The user's goal or intent as a short paragraph.

        Returns:
            A parsed dict with keys: `mapped_domain_id` (str),
            `extracted_prior_skills` (list[str]), and `reasoning` (str).
            On failure returns a dict with an `error` key and `raw` text.
        """

        domains: list[str] = self.graph.list_domains()
        domains_text = ", ".join(domains) if domains else ""

        system_prompt = (
            "You are a career mapping assistant. The user will state their goal. "
            f"You must map their intent to one of the following available domain IDs: {domains_text}. "
            "You must also extract any technical skills they claim to already know."
        )

        response_schema = {
            "mapped_domain_id": {"type": "string"},
            "extracted_prior_skills": {"type": "array"},
            "reasoning": {"type": "string"},
        }

        result = self.llm.generate_structured_json(system_prompt=system_prompt, user_prompt=user_text, response_schema=response_schema)

        if isinstance(result, dict) and result.get("error"):
            return {"error": result.get("error"), "raw": result.get("raw")}

        # Validate shape
        mapped = result.get("mapped_domain_id") if isinstance(result, dict) else None
        skills = result.get("extracted_prior_skills") if isinstance(result, dict) else None
        reasoning = result.get("reasoning") if isinstance(result, dict) else None

        if not isinstance(mapped, str):
            return {"error": "Invalid or missing 'mapped_domain_id' from LLM.", "raw": result}

        if mapped not in domains:
            # allow mapped suggestions that are close — but warn
            return {"mapped_domain_id": mapped, "extracted_prior_skills": skills or [], "reasoning": reasoning or "", "warning": "Mapped domain is not in available domains."}

        if not isinstance(skills, list):
            skills = []

        if not isinstance(reasoning, str):
            reasoning = ""

        return {"mapped_domain_id": mapped, "extracted_prior_skills": skills, "reasoning": reasoning}
