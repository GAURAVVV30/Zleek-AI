# Go API Implementation Map

| Method | Path | Go Module | Use Case | DB Tables | AI | Auth | RBAC |
|--------|------|-----------|----------|-----------|-----|------|------|
| POST | `/auth/register` | `identity` | UC-01 | `users`, `learner_profiles` | No | No | None |
| POST | `/auth/login` | `identity` | UC-02 | `users` | No | No | None |
| POST | `/auth/logout` | `identity` | UC-02 | (Redis sessions) | No | Yes | Learner/Curator/Admin |
| GET | `/profile` | `learner` | UC-02 | `learner_profiles` | No | Yes | Learner |
| PATCH | `/profile` | `learner` | UC-01 | `learner_profiles` | No | Yes | Learner |
| POST | `/goals` | `goals` | UC-03 | `goals` | Yes | Yes | Learner |
| GET | `/goals/active` | `goals` | UC-02 | `goals` | No | Yes | Learner |
| POST | `/diagnostics/start` | `assessment` | UC-04 | (Redis temp state) | No | Yes | Learner |
| POST | `/diagnostics/{sessionId}/answer` | `assessment` | UC-04 | (Redis temp state) | No | Yes | Learner |
| GET | `/diagnostics/{sessionId}/results` | `assessment` | UC-04 | `evidence_records`, `competency_records` | Yes | Yes | Learner |
| GET | `/path/active` | `roadmap` | UC-02 | `paths`, `path_items` | Yes | Yes | Learner |
| POST | `/path/items/{itemId}/engage` | `roadmap` | UC-06 | `engagement_events` | No | Yes | Learner |
| GET | `/path/items/{itemId}/explanation` | `roadmap` | UC-11 | None | Yes | Yes | Learner |
| GET | `/assessments/{conceptId}` | `assessment` | UC-07 | `assessment_definitions`, `assessment_items` | Yes | Yes | Learner |
| POST | `/assessments/{assessmentId}/submit` | `assessment` | UC-07 | `evidence_records`, `competency_records` | Yes | Yes | Learner |
| GET | `/storage/upload-url` | `assessment` | UC-12 | None | No | Yes | Learner |
| POST | `/projects/submit` | `assessment` | UC-12 | `evidence_records` | Yes | Yes | Learner |
| GET | `/competency` | `progress` | UC-02 | `competency_records` | No | Yes | Learner |
| GET | `/progress` | `progress` | UC-13 | `competency_records` | No | Yes | Learner |
| POST | `/resources/{resourceId}/feedback` | `feedback` | UC-10 | `feedback_records` | No | Yes | Learner |
| GET | `/notifications` | `notifications` | UC-01 | `notifications` | No | Yes | Learner |
| GET | `/domains` | `knowledge` | UC-14 | `domains` | No | Yes | Curator |
| POST | `/curator/knowledge-structures` | `knowledge` | UC-14 | `knowledge_structures`, `concepts`, `concept_prerequisites` | No | Yes | Curator |
| POST | `/curator/knowledge-structures/validate` | `knowledge` | UC-14 | None | No | Yes | Curator |
| GET | `/curator/resources/candidates` | `curator` | UC-14 | `resources` | No | Yes | Curator |
| POST | `/curator/resources/{resourceId}/publish` | `curator` | UC-14 | `resources`, `resource_concepts` | No | Yes | Curator |
| POST | `/curator/resources/{resourceId}/retire` | `curator` | UC-14 | `resources` | No | Yes | Curator |
| POST | `/telemetry/events` | `platform/events` | UC-06 | `engagement_events` | No | Yes | Learner |

## Module Dependencies
`identity`
  ↓
`knowledge` (provides domain structure)
  ↓
`learner` (depends on identity)
  ↓
`goals` (depends on learner, knowledge)
  ↓
`aiclient` (orchestrates AI logic for gaps, ranking)
  ↓
`roadmap` (depends on goals, knowledge, aiclient)
  ↓
`assessment` (depends on roadmap, knowledge, aiclient)
  ↓
`progress` (depends on assessment evidence)
  ↓
`feedback` (depends on resources, roadmap)
  ↓
`notifications` (depends on progress, roadmap)
  ↓
`curator` (depends on knowledge, resources)
