CREATE SCHEMA IF NOT EXISTS platform;
CREATE SCHEMA IF NOT EXISTS intelligence;
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ===================== IDENTITY =====================

CREATE TABLE platform.users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    role            TEXT NOT NULL CHECK (role IN ('learner','curator','admin')),
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','suspended')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE platform.learner_profiles (
    user_id             UUID PRIMARY KEY REFERENCES platform.users(id) ON DELETE RESTRICT,
    time_availability   TEXT,
    format_preference   TEXT,
    prior_experience    TEXT,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE platform.curator_domain_scopes (
    user_id     UUID NOT NULL REFERENCES platform.users(id) ON DELETE RESTRICT,
    domain_id   UUID NOT NULL, -- references platform.domains(id), deferred definition
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, domain_id)
);

-- ===================== KNOWLEDGE =====================

CREATE TABLE platform.domains (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    description TEXT
);

-- Add foreign key constraint that was deferred
ALTER TABLE platform.curator_domain_scopes ADD CONSTRAINT fk_curator_domains FOREIGN KEY (domain_id) REFERENCES platform.domains(id) ON DELETE RESTRICT;

CREATE TABLE platform.knowledge_structures (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain_id       UUID NOT NULL REFERENCES platform.domains(id) ON DELETE RESTRICT,
    version         INT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published','deprecated')),
    supersedes_id   UUID REFERENCES platform.knowledge_structures(id),
    created_by      UUID NOT NULL REFERENCES platform.users(id) ON DELETE RESTRICT,
    published_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (domain_id, version)
);

CREATE TABLE platform.concepts (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_structure_id  UUID NOT NULL REFERENCES platform.knowledge_structures(id) ON DELETE RESTRICT,
    name                    TEXT NOT NULL,
    description             TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (knowledge_structure_id, name)
);

CREATE TABLE platform.concept_prerequisites (
    concept_id              UUID NOT NULL REFERENCES platform.concepts(id) ON DELETE RESTRICT,
    prerequisite_concept_id UUID NOT NULL REFERENCES platform.concepts(id) ON DELETE RESTRICT,
    PRIMARY KEY (concept_id, prerequisite_concept_id),
    CHECK (concept_id <> prerequisite_concept_id)
);

-- ===================== GOALS & PATHS =====================

CREATE TABLE platform.goals (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    learner_id              UUID NOT NULL REFERENCES platform.users(id) ON DELETE RESTRICT,
    goal_text               TEXT NOT NULL,
    knowledge_structure_id  UUID NOT NULL REFERENCES platform.knowledge_structures(id) ON DELETE RESTRICT,
    status                  TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','achieved','abandoned')),
    achieved_at             TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE platform.paths (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    learner_id              UUID NOT NULL REFERENCES platform.users(id) ON DELETE RESTRICT,
    goal_id                 UUID NOT NULL REFERENCES platform.goals(id) ON DELETE RESTRICT,
    knowledge_structure_id  UUID NOT NULL REFERENCES platform.knowledge_structures(id) ON DELETE RESTRICT,
    status                  TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','completed')),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX one_active_path_per_goal ON platform.paths (goal_id) WHERE status = 'active';

CREATE TABLE platform.resources (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url                 TEXT NOT NULL UNIQUE,
    source              TEXT,
    author              TEXT,
    resource_type       TEXT NOT NULL,
    difficulty          TEXT,
    authority_score     NUMERIC,
    provenance_note     TEXT,
    status              TEXT NOT NULL DEFAULT 'candidate'
                            CHECK (status IN ('candidate','published','retired','flagged')),
    last_checked_at     TIMESTAMPTZ,
    freshness_status    TEXT NOT NULL DEFAULT 'unverified'
                            CHECK (freshness_status IN ('fresh','stale','unverified')),
    curated_by          UUID REFERENCES platform.users(id),
    curated_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (status <> 'published' OR (curated_by IS NOT NULL AND curated_at IS NOT NULL))
);

CREATE TABLE platform.path_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    path_id         UUID NOT NULL REFERENCES platform.paths(id) ON DELETE CASCADE,
    concept_id      UUID NOT NULL REFERENCES platform.concepts(id) ON DELETE RESTRICT,
    resource_id     UUID REFERENCES platform.resources(id) ON DELETE RESTRICT,
    sequence_order  INT NOT NULL,
    state           TEXT NOT NULL DEFAULT 'locked'
                        CHECK (state IN ('locked','available','in_progress','weak_evidence','competent')),
    is_remediation  BOOLEAN NOT NULL DEFAULT false,
    inserted_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(path_id, concept_id, is_remediation, sequence_order)
);
CREATE INDEX ON platform.path_items (path_id, sequence_order);

CREATE TABLE platform.resource_concepts (
    resource_id     UUID NOT NULL REFERENCES platform.resources(id) ON DELETE RESTRICT,
    concept_id      UUID NOT NULL REFERENCES platform.concepts(id) ON DELETE RESTRICT,
    relevance_note  TEXT,
    PRIMARY KEY (resource_id, concept_id)
);
CREATE INDEX ON platform.resource_concepts (concept_id);

CREATE TABLE platform.resource_quality_signals (
    resource_id         UUID PRIMARY KEY REFERENCES platform.resources(id) ON DELETE CASCADE,
    avg_rating          NUMERIC,
    feedback_count      INT NOT NULL DEFAULT 0,
    outcome_correlation NUMERIC,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ===================== ASSESSMENT & EVIDENCE =====================

CREATE TABLE platform.assessment_definitions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    concept_id      UUID NOT NULL REFERENCES platform.concepts(id) ON DELETE RESTRICT,
    type            TEXT NOT NULL CHECK (type IN ('quiz','scenario','project')),
    rubric          JSONB,
    version         INT NOT NULL,
    generated_by    TEXT NOT NULL CHECK (generated_by IN ('expert','ai')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (concept_id, version)
);

CREATE TABLE platform.assessment_items (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assessment_definition_id    UUID NOT NULL REFERENCES platform.assessment_definitions(id) ON DELETE CASCADE,
    prompt                      TEXT NOT NULL,
    item_type                   TEXT NOT NULL,
    answer_key                  JSONB,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE platform.evidence_records (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    learner_id                  UUID NOT NULL REFERENCES platform.users(id) ON DELETE RESTRICT,
    concept_id                  UUID NOT NULL REFERENCES platform.concepts(id) ON DELETE RESTRICT,
    assessment_definition_id    UUID NOT NULL REFERENCES platform.assessment_definitions(id) ON DELETE RESTRICT,
    path_item_id                UUID REFERENCES platform.path_items(id) ON DELETE RESTRICT,
    submission_data             JSONB NOT NULL,
    score                       NUMERIC,
    confidence                  NUMERIC,
    evaluator_type              TEXT NOT NULL CHECK (evaluator_type IN ('ai','curator')),
    evaluated_by                UUID REFERENCES platform.users(id),
    model_version               TEXT,
    result                      TEXT NOT NULL CHECK (result IN ('competent','weak','inconclusive')),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON platform.evidence_records (learner_id, concept_id, created_at DESC);

-- ===================== COMPETENCY =====================

CREATE TABLE platform.competency_records (
    learner_id      UUID NOT NULL REFERENCES platform.users(id) ON DELETE RESTRICT,
    concept_id      UUID NOT NULL REFERENCES platform.concepts(id) ON DELETE RESTRICT,
    state           TEXT NOT NULL DEFAULT 'not_started'
                        CHECK (state IN ('not_started','in_progress','weak_evidence','competent')),
    last_evidence_id UUID REFERENCES platform.evidence_records(id),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (learner_id, concept_id)
);

CREATE TABLE platform.competency_history (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    learner_id      UUID NOT NULL REFERENCES platform.users(id) ON DELETE RESTRICT,
    concept_id      UUID NOT NULL REFERENCES platform.concepts(id) ON DELETE RESTRICT,
    previous_state  TEXT,
    new_state       TEXT NOT NULL,
    evidence_id     UUID NOT NULL REFERENCES platform.evidence_records(id) ON DELETE RESTRICT,
    changed_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON platform.competency_history (learner_id, concept_id, changed_at DESC);

-- ===================== REMEDIATION =====================

CREATE TABLE platform.remediation_records (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    learner_id                  UUID NOT NULL REFERENCES platform.users(id) ON DELETE RESTRICT,
    concept_id                  UUID NOT NULL REFERENCES platform.concepts(id) ON DELETE RESTRICT,
    triggered_by_evidence_id    UUID NOT NULL REFERENCES platform.evidence_records(id) ON DELETE RESTRICT,
    remediation_resource_id     UUID REFERENCES platform.resources(id),
    attempt_number              INT NOT NULL DEFAULT 1,
    status                      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','resolved','escalated')),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at                 TIMESTAMPTZ
);

-- ===================== OTHER (Notifications, Engagement, Feedback, Audit) =====================

CREATE TABLE platform.engagement_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    learner_id      UUID NOT NULL REFERENCES platform.users(id) ON DELETE RESTRICT,
    path_item_id    UUID NOT NULL REFERENCES platform.path_items(id) ON DELETE RESTRICT,
    event_type      TEXT NOT NULL CHECK (event_type IN ('resource_opened','marked_reviewed')),
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE platform.notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    learner_id      UUID NOT NULL REFERENCES platform.users(id) ON DELETE RESTRICT,
    event_type      TEXT NOT NULL,
    payload         JSONB,
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE platform.feedback_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    learner_id      UUID NOT NULL REFERENCES platform.users(id) ON DELETE RESTRICT,
    target_type     TEXT NOT NULL CHECK (target_type IN ('resource','path_decision')),
    target_id       UUID NOT NULL,
    rating          NUMERIC,
    comment         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE platform.audit_log (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id            UUID NOT NULL REFERENCES platform.users(id) ON DELETE RESTRICT,
    action              TEXT NOT NULL,
    target_entity_type  TEXT NOT NULL,
    target_entity_id    UUID NOT NULL,
    before_state        JSONB,
    after_state         JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON platform.audit_log (target_entity_type, target_entity_id, created_at DESC);
CREATE INDEX ON platform.audit_log USING GIN (before_state);
CREATE INDEX ON platform.audit_log USING GIN (after_state);

-- ===================== INTELLIGENCE SCHEMA (FastAPI ML artifacts) =====================

CREATE TABLE intelligence.concept_embeddings (
    concept_id      UUID PRIMARY KEY,
    embedding       VECTOR(1536) NOT NULL,
    model_version   TEXT NOT NULL,
    generated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE intelligence.resource_embeddings (
    resource_id     UUID PRIMARY KEY,
    embedding       VECTOR(1536) NOT NULL,
    model_version   TEXT NOT NULL,
    generated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Note: Indexes on vectors usually applied after table has enough data. Adding them here for completeness.
CREATE INDEX ON intelligence.resource_embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
CREATE INDEX ON intelligence.concept_embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Table for generation cache not explicitly typed but implied.
CREATE TABLE intelligence.generation_cache (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cache_key       TEXT NOT NULL UNIQUE,
    payload         JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
