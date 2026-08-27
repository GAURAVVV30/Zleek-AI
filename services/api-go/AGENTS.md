# Go Backend Agent Rules

These rules must be followed for every future task performed inside the Go backend (`services/api-go/`).

## Authoritative Documents

The following documents are the sources of truth:
- **Architecture:** `@docs/AI_Learning_Platform_Architecture.md`
- **SRS:** `@docs/AI_Learning_Platform_SRS.md`
- **Database:** `@docs/AI_Learning_Platform_Database_Design.md`
- **Architecture report:** `@docs/AI_Learning_Platform_Architecture_Report.md`
- **Go API implementation map:** `@docs/development/go-api-implementation-map.md`
- **Module/database ownership map:** `@docs/development/module-database-map.md`
- **Go API contract:** `@contracts/openapi/go-api.yaml`
- **AI service contract:** `@contracts/openapi/ai-service.yaml`

Before implementing any future module, these documents must be inspected and used as the source of truth.

## Go Backend Scope

The Go backend is a MODULAR MONOLITH.
The Go service is: `services/api-go/`
FastAPI is a separate intelligence service: `services/ai-fastapi/`
The frontend is: `apps/web/`

For Go backend tasks:
- **MAY READ:** `docs/`, `contracts/`
- **MAY MODIFY:** `services/api-go/`, `contracts/openapi/go-api.yaml` (only when explicitly requested), `docs/development/` (only for backend documentation)
- **MUST NOT MODIFY:** `apps/web/`, `services/ai-fastapi/` unless a future task explicitly requests cross-service contract work.

## Architectural Rule

The fundamental AI boundary is:
Frontend → Go API → Go application layer → aiclient → FastAPI → AI/RAG/ML → aiclient → Go validation → PostgreSQL → Frontend

- FastAPI proposes.
- Go validates.
- Go decides.
- Go persists authoritative business state.
- FastAPI must never directly write Go-owned business tables.

## Module Structure

Every Go domain module under `services/api-go/internal/` must follow:
- `domain/`
- `application/`
- `infrastructure/`
- `interfaces/`

Dependency direction:
`interfaces` → `application` → `domain`
`infrastructure` implements `application`/`domain` ports.

**Rules:**
- `domain` must not depend on `infrastructure`
- `domain` must not depend on HTTP
- `domain` must not depend on PostgreSQL
- `domain` must not depend on Redis
- `domain` must not depend on FastAPI
- `application` must not contain SQL
- `application` must not contain HTTP handling
- HTTP handlers must not contain business logic
- repositories belong in `infrastructure`
- external service adapters belong in `infrastructure`
- `platform` contains cross-cutting technical concerns only

## Platform Rule

Reuse the existing `services/api-go/internal/platform/` including:
- config
- logger
- tracing
- httpserver
- middleware
- database
- cache
- events

- Do not create duplicate PostgreSQL clients.
- Do not create duplicate Redis clients.
- Do not create duplicate HTTP server infrastructure.
- Do not move business logic into platform.

## Database Rule

The canonical database design is: `@docs/AI_Learning_Platform_Database_Design.md`
- Go is the authoritative writer of platform/business state.
- Every database table must have one authoritative Go module owner.
- Do not invent duplicate tables.
- Do not invent columns without first verifying the authoritative documents.
- Do not put SQL inside HTTP handlers.
- Use migrations as the executable representation of the documented schema.

## API Contract Rule

The Go frontend-facing API contract is: `@contracts/openapi/go-api.yaml`
Never silently invent: endpoints, request fields, response fields, status codes, error structures.
If the implementation requires something missing from the contract:
**STOP.** Report the conflict and identify the authoritative source that supports the required change. Do not silently make assumptions.

The Go → FastAPI contract is: `@contracts/openapi/ai-service.yaml`
Do not invent FastAPI endpoints or request/response structures.

## Module Scope Rule

**ONE TASK = ONE MODULE.**
- Do not implement multiple domain modules in one task.
- Do not automatically continue to the next module after completing the current one.

After completing a module:
- run tests
- run gofmt
- run go vet
- run go build
- report changes
- **STOP.** Wait for explicit approval before continuing.

## Implementation Order

Future implementation should generally follow:
1. Platform/foundation
2. Database migrations
3. Identity
4. Learner
5. Knowledge structure
6. AI client
7. Goals
8. Assessment/Diagnostics
9. Roadmap
10. Resources
11. Projects/Storage
12. Competency
13. Progress
14. Notifications
15. Curator
16. Admin/Audit
17. Telemetry/Search
18. Contract/integration/security testing

Do not start a later module while its required dependencies are incomplete.

## Testing Rule

Every implemented feature must have appropriate tests.
Use:
- unit tests for domain/application behavior
- integration tests for database/external infrastructure
- contract tests for API contracts
- security tests where authentication/authorization is involved

Do not consider `go test ./...` passing with zero tests as evidence that a feature is implemented.

## Security Rules

Never:
- log passwords
- log secrets
- return password hashes
- expose internal database errors
- hardcode credentials
- hardcode API keys
- bypass authentication
- bypass authorization
- implement custom cryptography

Use the existing configuration and security abstractions.
