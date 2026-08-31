"""Domain graph loading and traversal utilities for learning paths."""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import networkx as nx

BASE_DIR = Path(__file__).resolve().parent.parent
DEFAULT_DOMAIN_PATH = BASE_DIR / "knowledge" / "domains"


class GraphEngine:
    """Load domain skill graphs as directed acyclic graphs and derive learning order.

    Each domain graph file is expected to have a top-level ``nodes`` list where each
    node contains an ``id`` and a ``prerequisites`` dictionary with a ``hard`` list of
    dependency node IDs. The engine validates that each domain graph is a DAG and
    exposes a personalized path function that returns the remaining learning steps in
    strict dependency order.
    """

    def __init__(self, domains_dir: str | Path | None = None) -> None:
        self.domains_dir = Path(domains_dir) if domains_dir is not None else DEFAULT_DOMAIN_PATH
        self.domain_graphs: dict[str, nx.DiGraph] = {}
        self.domain_metadata: dict[str, dict[str, Any]] = {}
        self._node_lookup: dict[str, dict[str, dict[str, Any]]] = {}
        self._load_domains()

    def _load_domains(self) -> None:
        """Scan the domains folder for JSON graph files and build each graph."""
        if not self.domains_dir.exists():
            raise FileNotFoundError(f"Domain directory not found: {self.domains_dir}")

        graph_files = sorted(self.domains_dir.rglob("*_graph.json"))
        if not graph_files:
            raise FileNotFoundError(
                f"No domain graph files found in '{self.domains_dir}'. "
                "Expected files ending with '_graph.json'."
            )

        valid_graph_count = 0

        for graph_file in graph_files:
            try:
                with graph_file.open("r", encoding="utf-8") as file_handle:
                    payload = json.load(file_handle)
            except json.JSONDecodeError as exc:
                raise ValueError(f"Invalid JSON in domain file '{graph_file}': {exc}") from exc

            try:
                domain_id = self._extract_domain_id(payload, graph_file)
            except ValueError:
                continue

            valid_graph_count += 1
            graph = nx.DiGraph()
            graph.graph["domain_id"] = domain_id
            graph.graph["domain_name"] = payload.get("domain_name")
            graph.graph["source_file"] = str(graph_file)

            nodes = payload.get("nodes", [])
            if not isinstance(nodes, list):
                raise TypeError(f"Domain '{domain_id}' has an invalid 'nodes' payload; expected a list.")

            for node in nodes:
                if not isinstance(node, dict):
                    raise TypeError(f"Domain '{domain_id}' contains a non-object node entry in '{graph_file}'.")

                node_id = node.get("id")
                if not node_id or not isinstance(node_id, str):
                    raise ValueError(
                        f"Domain '{domain_id}' contains a node without a valid string 'id' field."
                    )

                if graph.has_node(node_id):
                    raise ValueError(f"Duplicate node ID '{node_id}' found in domain '{domain_id}'.")

                graph.add_node(node_id, **node)
                self._node_lookup.setdefault(domain_id, {})[node_id] = node

            for node in nodes:
                node_id = node.get("id")
                prerequisites = node.get("prerequisites", {})
                if not isinstance(prerequisites, dict):
                    raise TypeError(
                        f"Node '{node_id}' in domain '{domain_id}' has an invalid 'prerequisites' payload."
                    )

                hard_prereqs = prerequisites.get("hard", [])
                if not isinstance(hard_prereqs, list):
                    raise TypeError(
                        f"Node '{node_id}' in domain '{domain_id}' has a non-list 'prerequisites.hard' value."
                    )

                for dependency in hard_prereqs:
                    if not isinstance(dependency, str):
                        raise TypeError(
                            f"Node '{node_id}' in domain '{domain_id}' contains a non-string prerequisite ID."
                        )
                    if dependency not in graph.nodes:
                        raise ValueError(
                            f"Node '{node_id}' in domain '{domain_id}' references missing prerequisite '{dependency}'."
                        )
                    if dependency == node_id:
                        raise ValueError(
                            f"Node '{node_id}' in domain '{domain_id}' is self-referencing as a prerequisite."
                        )
                    graph.add_edge(dependency, node_id)

            if not nx.is_directed_acyclic_graph(graph):
                raise ValueError(
                    f"Circular dependency detected in domain '{domain_id}'. "
                    f"The graph must be a valid DAG before it can be used for personalized learning paths."
                )

            self.domain_graphs[domain_id] = graph
            self.domain_metadata[domain_id] = payload

        if valid_graph_count == 0:
            raise FileNotFoundError(
                f"No valid domain graph payloads were found under '{self.domains_dir}'. "
                "Each graph file must contain a top-level 'domain_id' field."
            )

    @staticmethod
    def _extract_domain_id(payload: dict[str, Any], graph_file: Path) -> str:
        """Return the domain ID from the payload or raise a clear error."""
        domain_id = payload.get("domain_id")
        if not isinstance(domain_id, str) or not domain_id.strip():
            raise ValueError(f"Missing or invalid 'domain_id' in file '{graph_file.name}'.")
        return domain_id.strip()

    def get_personalized_path(self, domain_id: str, completed_nodes: list[str]) -> list[dict]:
        """Return the remaining learning path in a strict dependency order for a domain.

        The method performs a topological sort for the selected domain, removes any
        nodes already marked as complete, and returns the remaining node dictionaries.

        Args:
            domain_id: Unique domain identifier, matching a loaded graph file.
            completed_nodes: A list of node IDs already mastered by the learner.

        Returns:
            A list of remaining node dictionaries in execution order.

        Raises:
            ValueError: If the domain is unknown or the graph cannot be resolved.
        """
        if domain_id not in self.domain_graphs:
            valid_domains = ", ".join(sorted(self.domain_graphs.keys())) or "none"
            raise ValueError(f"Domain '{domain_id}' does not exist. Available domains: {valid_domains}.")

        graph = self.domain_graphs[domain_id]
        completed_set = {str(node_id) for node_id in completed_nodes}

        ordered_ids = list(nx.topological_sort(graph))
        remaining_ids = [node_id for node_id in ordered_ids if node_id not in completed_set]

        ordered_nodes: list[dict[str, Any]] = []
        for node_id in remaining_ids:
            if node_id not in self._node_lookup.get(domain_id, {}):
                continue
            ordered_nodes.append(dict(self._node_lookup[domain_id][node_id]))

        return ordered_nodes

    def get_graph(self, domain_id: str) -> nx.DiGraph:
        """Return the NetworkX directed graph for a given domain."""
        if domain_id not in self.domain_graphs:
            raise ValueError(f"Domain '{domain_id}' does not exist.")
        return self.domain_graphs[domain_id]

    def list_domains(self) -> list[str]:
        """Return the available domain IDs in the loaded graph set."""
        return sorted(self.domain_graphs.keys())


if __name__ == "__main__":
    engine = GraphEngine()
    for domain_id in engine.list_domains():
        print(f"{domain_id}: {len(engine.get_graph(domain_id).nodes)} nodes")
