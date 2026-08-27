# Module Database Ownership Map

| Table Name | Authoritative Go Module | Purpose | Primary Key | Important Foreign Keys | Dependent Modules | FastAPI Access |
|------------|-------------------------|---------|-------------|------------------------|-------------------|----------------|
| `users` | `identity` | Account identity and roles | `id` | None | All | Read |
| `curator_domain_scopes` | `identity` | Curator RBAC | `user_id, domain_id` | `users`, `domains` | `curator`, `knowledge` | Read |
| `learner_profiles` | `learner` | Learner constraints/preferences | `user_id` | `users` | `roadmap` | Read |
| `domains` | `knowledge` | Top-level knowledge areas | `id` | None | `goals`, `curator` | Read |
| `knowledge_structures` | `knowledge` | Versioned curriculum graphs | `id` | `domains`, `users` | `roadmap`, `goals` | Read |
| `concepts` | `knowledge` | Individual skills | `id` | `knowledge_structures` | `roadmap`, `assessment`, `progress` | Read |
| `concept_prerequisites` | `knowledge` | Concept DAG edges | `concept_id, prerequisite_concept_id` | `concepts` | `roadmap` | Read |
| `goals` | `goals` | Learner target state | `id` | `users`, `knowledge_structures` | `roadmap`, `progress` | Read |
| `paths` | `roadmap` | Learner path instance | `id` | `users`, `goals`, `knowledge_structures` | `assessment`, `progress` | Read |
| `path_items` | `roadmap` | Path sequence items | `id` | `paths`, `concepts`, `resources` | `assessment` | Read |
| `resources` | `curator` | Vetted external resources | `id` | `users` | `roadmap`, `feedback` | Read |
| `resource_concepts` | `curator` | Resource to concept mapping | `resource_id, concept_id` | `resources`, `concepts` | `roadmap` | Read |
| `resource_quality_signals` | `feedback` | Aggregate resource quality | `resource_id` | `resources` | `roadmap`, `curator` | Read |
| `assessment_definitions` | `assessment` | Rubrics and quizzes | `id` | `concepts` | `progress` | Read |
| `assessment_items` | `assessment` | Individual questions | `id` | `assessment_definitions` | None | Read |
| `evidence_records` | `progress` | Evaluation results | `id` | `users`, `concepts`, `assessment_definitions`, `path_items` | `progress`, `roadmap` | Read |
| `competency_records` | `competency` | Current concept mastery | `learner_id, concept_id` | `users`, `concepts`, `evidence_records` | `roadmap` | Read |
| `competency_history` | `competency` | Mastery audit log | `id` | `users`, `concepts`, `evidence_records` | None | Read |
| `remediation_records` | `roadmap` | Remediation actions | `id` | `users`, `concepts`, `evidence_records`, `resources` | None | Read |
| `engagement_events` | `progress` | Resource open/view events | `id` | `users`, `path_items` | None | No |
| `notifications` | `notifications` | Learner alerts | `id` | `users` | None | No |
| `feedback_records` | `feedback` | Resource ratings | `id` | `users` | `curator` | No |
| `audit_log` | `admin` | Immutable audit trail | `id` | `users` | None | No |

**Rule:** FASTAPI DOES NOT OWN GO BUSINESS STATE. It only owns `intelligence` schema tables (`concept_embeddings`, `resource_embeddings`, `generation_cache`). FastAPI only reads data from Go schemas for contextual ranking and gap analysis.
