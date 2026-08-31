-- Concepts get a stable public (node-id) key so the roadmap.graph node ids
-- (ml_01_programming_fundamentals, arch_01_networking_fundamentals, ...) map
-- 1:1 to concepts while the PK stays a UUID for FK integrity.
ALTER TABLE platform.concepts ADD COLUMN node_id TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX uq_concepts_node_id ON platform.concepts (node_id);

-- Resources gain presentational columns required by the frontend concept view.
ALTER TABLE platform.resources ADD COLUMN title TEXT NOT NULL DEFAULT '';
ALTER TABLE platform.resources ADD COLUMN duration_minutes INT;

-- Domains gain a human-facing slug key (machine_learning, software_architecture)'
-- so the 12 role pages can address domains without leaking UUIDs.
ALTER TABLE platform.domains ADD COLUMN slug TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX uq_domains_slug ON platform.domains (slug);