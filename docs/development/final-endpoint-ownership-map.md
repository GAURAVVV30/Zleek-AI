# Final Endpoint Ownership and Implementation Audit

This document verifies the API surface requested by the frontend against the actual architectural owners, implementation status in Go/FastAPI, and database boundaries.

## 1. Complete Endpoint Matrix

| POST | `/diagnostics/start` | GO → FASTAPI ORCHESTRATION | NO | NO | Unknown (Conflict) | NOT IMPLEMENTED |
| POST | `/diagnostics/{sessionId}/answer` | GO → FASTAPI ORCHESTRATION | NO | NO | Unknown (Conflict) | NOT IMPLEMENTED |
| GET | `/diagnostics/{sessionId}/results` | GO → FASTAPI ORCHESTRATION | NO | NO | Unknown (Conflict) | NOT IMPLEMENTED |
| **ROADMAP** | | | | | | |
| GET | `/roadmap` | GO | YES | NO | `paths` (`roadmap`) | IMPLEMENTED |
| GET | `/roadmap/concepts/{id}/why` | GO → FASTAPI ORCHESTRATION | YES | YES (`ai-service.yaml`) | None (Stateless/AI) | IMPLEMENTED |
| POST | `/roadmap/regenerate` | GO → FASTAPI ORCHESTRATION | YES | YES (`ai-service.yaml`) | `paths` (`roadmap`) | IMPLEMENTED |
| **LEARNING** | | | | | | |
| GET | `/concepts/{id}` | GO | NO | NO | `concepts` (`knowledge`) | NOT IMPLEMENTED |
| GET | `/concepts/{id}/resources/alternate`| GO → FASTAPI ORCHESTRATION | NO | YES (`ai-service.yaml`) | `resources` (`curator`) | NOT IMPLEMENTED |
| GET | `/concepts/{id}/resources/{resId}/why`| GO → FASTAPI ORCHESTRATION | NO | YES (`ai-service.yaml`) | None (Stateless/AI) | NOT IMPLEMENTED |
| POST | `/concepts/{id}/engagement` | GO | YES | NO | `engagement_events` (`progress`) | IMPLEMENTED |
| POST | `/resources/{resourceId}/feedback` | GO | NO | NO | `feedback_records` (`feedback`) | NOT IMPLEMENTED |
| **ASSESSMENT / PROJECT**| | | | | | |
| GET | `/concepts/{id}/assessment` | GO | YES | NO | `assessment_definitions` (`assessment`) | IMPLEMENTED |
| POST | `/concepts/{id}/assessment/submit`| GO → FASTAPI ORCHESTRATION | YES | YES (`ai-service.yaml`) | `evidence_records` (`progress`) | IMPLEMENTED |
| GET | `/concepts/{id}/project` | GO | YES | NO | `assessment_definitions` (`assessment`) | IMPLEMENTED |
| POST | `/concepts/{id}/project/submit` | GO | YES | NO | `evidence_records` (`progress`) | IMPLEMENTED |
| GET | `/concepts/{id}/project/status` | GO | YES | NO | `evidence_records` (`progress`) | IMPLEMENTED |
| **PROGRESS / COMPETENCY / NOTIFICATIONS**| | | | | | |
| GET | `/progress/summary` | GO | YES | NO | `evidence_records` (`progress`) | IMPLEMENTED |
| GET | `/competency/detail` | GO | YES | NO | `competency_records` (`competency`) | IMPLEMENTED |
| GET | `/competency/{conceptId}/history` | GO | YES | NO | `competency_history` (`competency`) | IMPLEMENTED |
| GET | `/goals/current/completion-summary` | GO | YES | NO | `goals` / `progress` | IMPLEMENTED |
| GET | `/notifications` | GO | YES | NO | `notifications` (`notifications`) | IMPLEMENTED |
| PATCH | `/notifications/{id}/read` | GO | YES | NO | `notifications` (`notifications`) | IMPLEMENTED |
| **CURATOR** | | | | | | |
| GET | `/curator/knowledge-structures` | GO | NO | NO | `knowledge_structures` (`knowledge`) | NOT IMPLEMENTED |
| POST | `/curator/knowledge-structures` | GO | NO | NO | `knowledge_structures` (`knowledge`) | NOT IMPLEMENTED |
| PATCH | `/curator/knowledge-structures` | GO | NO | NO | `knowledge_structures` (`knowledge`) | NOT IMPLEMENTED |
| POST | `/curator/knowledge-structures/validate`| GO → FASTAPI ORCHESTRATION | NO | YES (`ai-service.yaml`) | None (Stateless/AI) | NOT IMPLEMENTED |
| GET | `/curator/resources` | GO | YES | NO | `resources` (`curator`) | IMPLEMENTED |
| POST | `/curator/resources` | GO | YES | NO | `resources` (`curator`) | IMPLEMENTED |
| PATCH | `/curator/resources` | GO | YES | NO | `resources` (`curator`) | IMPLEMENTED |
| GET | `/curator/resources/{id}/feedback-signals`| GO | YES | NO | `resource_quality_signals` (`feedback`) | IMPLEMENTED |
| **ADMIN** | | | | | | |
| GET | `/admin/users` | GO | YES | NO | `users` (`identity`) | IMPLEMENTED |
| PATCH | `/admin/users` | GO | YES | NO | `users`, `audit_log` (`identity`, `admin`) | IMPLEMENTED |
| GET | `/admin/audit-log` | GO | YES | NO | `audit_log` (`admin`) | IMPLEMENTED |
| **OPTIONAL ADDITIONS** | | | | | | |
| GET | `/storage/upload-url` | GO | NO | NO | None (Infrastructure) | DEFERRED |
| PATCH | `/profile/settings` | GO | NO | NO | `users` (`identity`) | DEFERRED |
| POST | `/auth/change-password` | GO | NO | NO | `users` (`identity`) | DEFERRED |
| DELETE | `/profile/account` | GO | NO | NO | `users` (`identity`) | DEFERRED |
| GET | `/search` | GO → FASTAPI ORCHESTRATION | NO | YES (`ai-service.yaml`) | `intelligence.*` (`FastAPI`) | DEFERRED |
| POST | `/telemetry/events` | GO | NO | NO | None (Infrastructure) | DEFERRED |
| GET | `/metrics` | GO | NO | NO | None (Infrastructure) | DEFERRED |
| **SYSTEM** | | | | | | |
| GET | `/health` | GO | YES | NO | None | IMPLEMENTED |
| GET | `/ready` | GO | YES | NO | Platform DB Connection | IMPLEMENTED |

---

## 2. Missing Go Endpoints
The following endpoints exist in `go-api.yaml` and are required by the frontend but have no handlers, domain logic, or database adapters implemented in `services/api-go`:
- `/auth/signup`, `/auth/login`, `/auth/refresh`, `/auth/logout`, `/auth/me` (Identity module completely missing)
- `/domains`, `/concepts/{id}` (Knowledge structure reads missing)
- `/concepts/{id}/resources/alternate`, `/concepts/{id}/resources/{resId}/why` (Resource AI features missing)
- `/resources/{resourceId}/feedback` (Feedback ingestion missing)
- `/curator/knowledge-structures/*` (Curator Knowledge mutation missing)
- `/profile/preferences` (Learner preferences missing)
- `/diagnostics/*` (Diagnostic flow missing)

## 3. Incorrectly Assigned Endpoints / Conflicts
- **`/diagnostics/*` Conflict**: The frontend requests `/diagnostics/start`, `/diagnostics/{sessionId}/answer`, and `/diagnostics/{sessionId}/results`. While present in `go-api.yaml`, there is zero documentation of a Diagnostics data model in `AI_Learning_Platform_Database_Design.md`, and it is completely omitted from the `module-database-map.md`. If Go is to persist diagnostic sessions, an architectural update to the schema is required.
- **`/search` Conflict**: Search traditionally belongs to FastAPI as it interfaces with `intelligence.concept_embeddings` and `intelligence.resource_embeddings`. However, `go-api.yaml` exposes `GET /search`. This indicates a "GO → FASTAPI ORCHESTRATION" relay pattern, rather than direct frontend-to-FastAPI access.

## 4. Intelligence (formerly FastAPI-owned) Endpoints
The AI intelligence layer documented in `ai-service.yaml` has been ported into the Go backend and is served in-process at exact `/api/v1/*` paths (mirroring the FastAPI route set). Go's domain routes remain the front-door; these `/api/v1/*` intelligence endpoints are the direct AI capability surface:
- `POST /api/v1/goal/analyze`
- `GET /api/v1/roadmap`, `POST /api/v1/roadmap`, `GET /api/v1/roadmap/list`
- `GET /api/v1/resource`
- `POST /api/v1/evaluate`
- `POST /api/v1/recommendation/personalize-roadmap`
- `POST /api/v1/learning/lesson`, `POST /api/v1/learning/evaluate`
- `POST /api/v1/adaptive/next-action`
- `POST /api/v1/mastery/update`, `POST /api/v1/mastery/update-incremental`, `GET /api/v1/mastery/params/{node_id}`
- `GET /api/v1/guardrails/status`, `POST /api/v1/guardrails/check`
- `GET /api/v1/voice/status`, `POST /api/v1/voice/transcribe`, `POST /api/v1/voice/synthesize`, `POST /api/v1/voice/tutor-session`
- `GET /api/v1/health`

These endpoints degrade without LLM/NVIDIA/Groq keys exactly as FastAPI did (guardrail blocks, BKT/adaptive logic, keyword domain matching and deterministic fallbacks still work).

## 5. Shared Endpoints
There are no endpoints where Go and FastAPI jointly manage HTTP writing. Go is exclusively the front-door (API Gateway), maintaining a strict orchestrator pattern where it proxies AI requests to FastAPI, validates them, and writes to Postgres.

## 6. Optional Endpoints That Should Be Implemented Now
None. Given the severity of the missing core capabilities (Auth, Knowledge mapping, Learner profiles), implementing optional telemetry, search, or storage URLs would be a misuse of resources.

## 7. Optional Endpoints That Should Be Deferred
- `GET /storage/upload-url`
- `PATCH /profile/settings`
- `POST /auth/change-password`
- `DELETE /profile/account`
- `GET /search`
- `POST /telemetry/events`
- `GET /metrics`

## 8. Critical Integration Blockers
1. **Missing Identity Module**: Auth endpoints `/auth/*` are missing. The entire platform currently relies on mocked authentication (hardcoded headers). It is impossible to launch integration tests against the frontend without a real `users` table flow.
2. **Missing Knowledge Management**: `/curator/knowledge-structures` mutations and basic `/domains` reads are omitted. Without these, the core curriculum DAG cannot be modified.
3. **Missing Feedback Ingestion**: `/resources/{resourceId}/feedback` is not implemented, preventing learners from providing the signal needed for the curator dashboard and AI remediation.

## 9. Recommended Implementation Order
To unblock full-stack integration safely, the following sequence is recommended:
1. **Identity & Auth**: Implement `internal/identity` with JWT issuing and validation.
2. **Learner Profiles**: Implement `internal/learner` to handle `/profile/preferences`.
3. **Knowledge Base CRUD**: Implement `internal/knowledge` for basic concept/domain lookups and curator mutations.
4. **Feedback**: Implement `internal/feedback` to capture resource ratings.
5. **Diagnostics**: Clarify architecture and DB design for `/diagnostics/*` before proceeding.
6. **Frontend Integration**: Once auth and core data structures are stable.
