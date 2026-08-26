-- Drop Intelligence schema objects
DROP TABLE IF EXISTS intelligence.generation_cache;
DROP TABLE IF EXISTS intelligence.resource_embeddings;
DROP TABLE IF EXISTS intelligence.concept_embeddings;

-- Drop Platform schema objects
DROP TABLE IF EXISTS platform.audit_log;
DROP TABLE IF EXISTS platform.feedback_records;
DROP TABLE IF EXISTS platform.notifications;
DROP TABLE IF EXISTS platform.engagement_events;
DROP TABLE IF EXISTS platform.remediation_records;
DROP TABLE IF EXISTS platform.competency_history;
DROP TABLE IF EXISTS platform.competency_records;
DROP TABLE IF EXISTS platform.evidence_records;
DROP TABLE IF EXISTS platform.assessment_items;
DROP TABLE IF EXISTS platform.assessment_definitions;
DROP TABLE IF EXISTS platform.resource_quality_signals;
DROP TABLE IF EXISTS platform.resource_concepts;
DROP TABLE IF EXISTS platform.path_items;
DROP TABLE IF EXISTS platform.resources;
DROP TABLE IF EXISTS platform.paths;
DROP TABLE IF EXISTS platform.goals;
DROP TABLE IF EXISTS platform.concept_prerequisites;
DROP TABLE IF EXISTS platform.concepts;
DROP TABLE IF EXISTS platform.knowledge_structures;
ALTER TABLE platform.curator_domain_scopes DROP CONSTRAINT IF EXISTS fk_curator_domains;
DROP TABLE IF EXISTS platform.domains;
DROP TABLE IF EXISTS platform.curator_domain_scopes;
DROP TABLE IF EXISTS platform.learner_profiles;
DROP TABLE IF EXISTS platform.users;

-- Drop Extensions and Schemas
DROP EXTENSION IF EXISTS pgcrypto;
DROP EXTENSION IF EXISTS vector;
DROP SCHEMA IF EXISTS intelligence;
DROP SCHEMA IF EXISTS platform;
