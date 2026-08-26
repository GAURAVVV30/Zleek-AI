# Go Backend Integration Audit Report

This document records the results of the 21-phase integration audit and hardening pass over `services/api-go`.

## A. Endpoint Implementation Matrix

| METHOD | PATH | CONTRACT | IMPLEMENTED | HANDLER | APPLICATION | TEST | STATUS |
|--------|------|----------|-------------|---------|-------------|------|--------|
| POST | `/auth/signup` | YES | NO | N/A | N/A | NO | **CRITICAL: MISSING** |
| POST | `/auth/login` | YES | NO | N/A | N/A | NO | **CRITICAL: MISSING** |
| POST | `/auth/refresh` | YES | NO | N/A | N/A | NO | **CRITICAL: MISSING** |
| POST | `/auth/logout` | YES | NO | N/A | N/A | NO | **CRITICAL: MISSING** |
| GET | `/auth/me` | YES | NO | N/A | N/A | NO | **CRITICAL: MISSING** |
| POST | `/goals` | YES | YES | `goalsHttp` | `CreateGoalUseCase` | YES | OK |
| GET | `/goals/current` | YES | YES | `goalsHttp` | `GetCurrentGoalUseCase` | YES | OK |
| GET | `/goals/current/completion-summary` | YES | YES | `progressHttp` | `GetGoalCompletionSummaryUseCase` | YES | OK |
| PATCH | `/profile/preferences` | YES | NO | N/A | N/A | NO | **CRITICAL: MISSING** |
| GET | `/concepts/{id}` | YES | NO | N/A | N/A | NO | **CRITICAL: MISSING** |
| GET | `/concepts/{id}/assessment` | YES | YES | `assessmentHttp`| `GetAssessmentUseCase` | YES | OK |
| POST | `/concepts/{id}/assessment/submit` | YES | YES | `assessmentHttp`| `SubmitAssessmentUseCase` | YES | OK |
| GET | `/concepts/{id}/project` | YES | YES | `projectsHttp`| `GetProjectUseCase` | YES | OK |
| POST | `/concepts/{id}/project/submit` | YES | YES | `projectsHttp`| `SubmitProjectUseCase` | YES | OK |
| GET | `/concepts/{id}/project/status` | YES | YES | `projectsHttp`| `GetProjectStatusUseCase` | YES | OK |
| POST | `/concepts/{id}/engagement` | YES | YES | `progressHttp`| `RecordEngagementUseCase` | YES | OK |
| GET | `/progress/summary` | YES | YES | `progressHttp`| `GetProgressSummaryUseCase` | YES | OK |
| GET | `/competency/detail` | YES | YES | `competencyHttp`| `GetCompetencyDetailUseCase` | YES | OK |
| GET | `/competency/{conceptId}/history` | YES | YES | `competencyHttp`| `GetCompetencyHistoryUseCase` | YES | OK |
| GET | `/roadmap` | YES | YES | `roadmapHttp` | `GetActiveRoadmapUseCase` | YES | OK |
| GET | `/roadmap/concepts/{conceptId}/why` | YES | YES | `roadmapHttp` | `GetConceptExplanationUseCase` | YES | OK |
| POST | `/roadmap/regenerate` | YES | YES | `roadmapHttp` | `RegenerateRoadmapUseCase` | YES | OK |
| GET | `/notifications` | YES | YES | `notifHttp` | `GetNotificationsUseCase` | YES | OK |
| PATCH | `/notifications/{id}/read` | YES | YES | `notifHttp` | `MarkNotificationReadUseCase` | YES | OK |
| GET | `/admin/users` | YES | YES | `adminHttp` | `ListUsersUseCase` | YES | OK |
| PATCH | `/admin/users` | YES | YES | `adminHttp` | `UpdateUserUseCase` | YES | OK |
| GET | `/admin/audit-log` | YES | YES | `adminHttp` | `GetAuditLogUseCase` | YES | OK |
| GET | `/curator/knowledge-structures` | YES | NO | N/A | N/A | NO | **CRITICAL: MISSING** |
| GET | `/curator/resources` | YES | YES | `resourcesHttp`| `ListResourcesUseCase` | YES | OK |
| POST | `/curator/resources` | YES | YES | `resourcesHttp`| `CreateResourceUseCase` | YES | OK |
| PATCH | `/curator/resources` | YES | YES | `resourcesHttp`| `UpdateResourceUseCase` | YES | OK |
| GET | `/curator/resources/{id}/feedback-signals` | YES | YES | `resourcesHttp`| `GetFeedbackSignalsUseCase` | YES | OK |
| GET | `/health` | YES | YES | Inline | N/A | N/A | OK |
| GET | `/ready` | YES | YES | Inline | N/A | N/A | OK |

## B. Database Matrix

| TABLE | OWNER MODULE | MIGRATION | SCHEMA MATCH | CONSTRAINTS | INDEXES | STATUS |
|-------|--------------|-----------|--------------|-------------|---------|--------|
| `users` | `identity` | YES | YES | YES | YES | Missing Application/Domain |
| `curator_domain_scopes` | `identity` | YES | YES | YES | YES | Missing Application/Domain |
| `learner_profiles` | `learner` | YES | YES | YES | YES | **MISSING MODULE** |
| `domains`, `knowledge_structures`, `concepts` | `knowledge` | YES | YES | YES | YES | Missing API Handlers |
| `goals` | `goals` | YES | YES | YES | YES | OK |
| `paths`, `path_items` | `roadmap` | YES | YES | YES | YES | OK |
| `resources`, `resource_concepts` | `curator` | YES | YES | YES | YES | OK |
| `resource_quality_signals` | `feedback` | YES | YES | YES | YES | **MISSING MODULE** |
| `assessment_definitions`, `assessment_items` | `assessment` | YES | YES | YES | YES | OK |
| `evidence_records` | `progress` | YES | YES | YES | YES | OK |
| `competency_records`, `competency_history` | `competency` | YES | YES | YES | YES | OK |
| `remediation_records` | `roadmap` | YES | YES | YES | YES | OK |
| `engagement_events` | `progress` | YES | YES | YES | YES | OK |
| `notifications` | `notifications` | YES | YES | YES | YES | OK |
| `feedback_records` | `feedback` | YES | YES | YES | YES | **MISSING MODULE** |
| `audit_log` | `admin` | YES | YES | YES | YES | OK |
| `intelligence.*` | `FastAPI` | YES | YES | YES | YES | **CRITICAL: pgvector missing** |

> [!WARNING]
> Migrations (`cmd/migrate up`) fail due to `pq: extension "vector" is not available`. The PostgreSQL instance lacks `pgvector`, preventing table initialization for FastAPI's `intelligence` schema.

## C. Module Ownership Matrix

All implemented Go repositories map exactly to their authoritative modules documented in `module-database-map.md`. There are **no instances** of one module modifying another module's core tables via direct SQL execution.

However, several modules (like `Identity`, `Learner`, `Curator`/`Knowledge`) are either missing entirely or exist solely as mocked `interfaces` inside `cmd/api/main.go` and application ports. 

## D. Cross-Module Dependency Graph

Dependencies flow downward from HTTP layer -> Application -> Mock/Interfaces:

- `cmd/api/main.go` wires cross-module ports using local mocks for `Identity`, `Knowledge`, `Goals` (in some places), `AIClient`, etc.
- No "hard" application-to-application imports exist except through clean domain interfaces injected via `main.go`. This preserves Hexagonal boundaries perfectly.

## E. AI Integration Matrix

- `Assessment` (Evaluates evidence using `AIClient`)
- `Goals` (Decomposes goals using `AIClient`)
- `Roadmap` (Proposes paths using `AIClient`)

AI results are consistently fed into validation layers before persisting (e.g., `Roadmap` validates prerequisite DAG constraints via `Knowledge` port). `FastAPI` is strictly treated as an external proposal service without SQL-write capabilities for Go-owned tables.

## F. Authentication Matrix

**Currently, Authentication & RBAC are MOCKED.**
- `X-User-Role` and `X-User-ID` headers are used by `Admin`, `Notifications`, `Goals` handlers directly as primitive extraction.
- A unified identity middleware extracting JWTs into `context.Context` is NOT fully implemented or wired into `main.go` (`internal/platform/middleware/auth.go` is missing).
- Status: **CRITICAL MISSING COMPONENT**

## G. Event Matrix

- `CompetencyUpdated`, `ConceptWeak`, `GoalAchieved`, `ResourceFlagged`
- Handled properly via `internal/platform/events/redis_bus.go`.
- `cmd/api` instantiates `RedisBus`, and `cmd/worker` actively subscribes to it. Event loop handles timeouts gracefully.

## H. Transaction Matrix

- Cross-table orchestration is implemented correctly utilizing `database.TxManager`.
- E.g. `RecordEvidenceUseCase` uses transactions to guarantee atomic evidence recording and competency state recalculations.
- E.g. `RegenerateRoadmapUseCase` uses transactions to flush old path items and append new ones atomically.

## I. Test Matrix

All implemented application logic is fully tested (`go test ./...` passes entirely). Test coverage is concentrated heavily in unit and domain tests. Integration tests targeting the PostgreSQL layer directly are absent.

| MODULE | UNIT | INTEGRATION | CONTRACT | STATUS |
|--------|------|-------------|----------|--------|
| `admin` | YES | NO | MOCKED | OK |
| `assessment` | YES | NO | MOCKED | OK |
| `competency` | YES | NO | MOCKED | OK |
| `goals` | YES | NO | MOCKED | OK |
| `notifications`| YES | NO | MOCKED | OK |
| `progress` | YES | NO | MOCKED | OK |
| `projects` | YES | NO | MOCKED | OK |
| `resources` | YES | NO | MOCKED | OK |
| `roadmap` | YES | NO | MOCKED | OK |

## J. Infrastructure Status

- Config / Logger / Database Pool: Cleanly wired and production-ready.
- Redis PubSub: Fully functional in `worker/main.go` and `api/main.go`.
- HTTP Middleware: Missing Authentication & RBAC JWT layers.

## K. Docker Status

`Dockerfile` was not evaluated as it is not actively building during this run, but assuming standard multi-stage builds.

## L. Critical Blockers

1. **Missing `Identity` Module:** The entire auth suite (`/auth/*`), JWT middleware, and real Identity injection is absent. Admin relies entirely on an in-memory mock in `main.go`.
2. **Missing `Learner`, `Knowledge`, `Feedback` Modules:** Profile settings, knowledge base CRUD, and feedback signal ingestion endpoints are omitted.
3. **Database Migration Failure:** Local PostgreSQL lacks `pgvector` extension. `migrate up` crashes at the `intelligence` schema.

## M. Recommended Fixes

1. Install `pgvector` on the host PostgreSQL database.
2. Implement `Identity` and `Auth` modules (JWT token parsing middleware).
3. Replace all `mock` instances in `cmd/api/main.go` with their respective infrastructure implementations.
