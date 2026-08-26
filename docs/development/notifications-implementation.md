# Notifications Module Implementation

## Module Responsibilities
The `Notifications` module is the authoritative owner of the `platform.notifications` table. It does not own or modify any core business state (e.g., goals, competency, or roadmap). Instead, it acts as a decoupled consumer of Domain Events emitted by business modules via the platform event bus.

## Event Sources and Schemas
The following events are consumed:
- `CompetencyUpdated`: Triggers when a learner's competency state changes.
- `ConceptWeak`: Triggers when an assessment or evaluation identifies a weakness.
- `GoalAchieved`: Triggers when all prerequisites for a goal are satisfied.
- `ResourceFlagged`: Triggers when a resource is reported.

Schemas for these events were explicitly documented as struct payloads inside the `internal/platform/events` package.

## Redis Event Bus and Worker Behavior
To satisfy the architectural requirement for asynchronous event consumption (and "Redis pub/sub" delivery):
- Implemented `events.RedisBus` referencing the `go-redis/v9` client.
- The `api` service (in `cmd/api/main.go`) initializes the `RedisBus` for future publishing by the business modules.
- The `worker` service (in `cmd/worker/main.go`) initializes the `Notifications` domain logic and subscribes to the `RedisBus`.
- When an event occurs, the worker deserializes it and calls the `Notifications` `EventHandler`, which persists a new notification row.

> [!IMPORTANT]
> **Notifications consume domain events and do not own business state.** 
> The worker is isolated from HTTP handling.

## Endpoints Implemented
Matched exactly against `go-api.yaml`:
- `GET /notifications` (Retrieves the learner's notification history).
- `PATCH /notifications/{id}/read` (Marks a specific notification as read).

## Authentication
`X-User-ID` is parsed from headers (typically set by identity middleware) to ensure the notifications are strictly scoped to the authenticated learner. Marking another user's notification as read explicitly returns an `Unauthorized` domain error.

## Tests and Verification
- Domain tests mocked the internal storage to simulate event injection (`HandleCompetencyUpdated` etc) and boundary protections.
- `go vet`, `go test`, and `go build` passed successfully for both `api` and `worker` binaries.
