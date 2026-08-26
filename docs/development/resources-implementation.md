# Resources Module Implementation

## Module Responsibilities
The `Resources` module is the authoritative owner of the learning platform's resource catalog. It manages resource metadata, content URLs, quality signals, and the mappings between resources and knowledge concepts.

## Owned Tables
- `platform.resources`
- `platform.resource_concepts`
- `platform.resource_quality_signals`

> [!IMPORTANT]
> **Feedback Boundary:** As per the `module-database-map.md`, `feedback_records` is owned by a separate `Feedback` module. The `Resources` module does *not* own or directly manage raw feedback submissions.

## Resource Lifecycle
Resources follow a strict lifecycle managed by the Go domain model:
1. **candidate**: Initial state when a resource is proposed or uploaded.
2. **published**: State after a Curator validates and approves the resource.
3. **retired**: Resource is no longer active but retained for historical records.
4. **flagged**: Resource requires attention due to reports or automated checks.

Transitions are enforced securely (e.g. transitioning to `published` mandates setting `curated_by` and `curated_at` metadata).

## Concept Relationship
A Resource can be mapped to multiple Knowledge Concepts via `resource_concepts`. This relationship is created atomically within a transaction when a resource is added.
To prevent invalid references without breaking modularity, the `Resources` application layer validates concepts via an injected `ConceptValidationService` port, which is implemented in `main.go` to delegate to the `Knowledge` module.

## Curator Authorization
Endpoints modifying or retrieving non-public resource state (e.g., `POST /curator/resources`) are secured using HTTP handler middleware checks enforcing `X-User-Role` is `curator` or `admin`.

## Quality Signals
`resource_quality_signals` stores the aggregated feedback metrics. The `Resources` module provides the API endpoint to read these aggregated signals, but raw feedback is processed asynchronously or by the `Feedback` module.

## AI Ranking Boundary
> [!IMPORTANT]
> **Go owns authoritative resource state.** AI ranking is strictly advisory and must be validated by Go. 
> The FastAPI service reads resource data, generates embeddings (`intelligence` schema), and provides rankings. However, FastAPI is strictly prohibited from modifying `resources` or `resource_concepts`. Any future AI-driven roadmap generation will consume AI candidate lists, but Go will validate and filter those candidates against the authoritative, published catalog before serving them to the learner.

## File Upload / Storage Dependency
Resource creation currently accepts URLs. Full file upload support requires the `Storage` service (`GET /storage/upload-url`) which generates presigned S3 URLs, allowing the frontend to upload directly to object storage and submit the final S3 URL to this `Resources` module.

## Tests and Verification
- **Unit Tests:** `internal/resources/application/resources_test.go` verifies lifecycle transitions (e.g., candidate -> published requires curator identity) and application use cases.
- **Commands run:** `go vet ./internal/resources/...`, `go test -v ./internal/resources/...`, and `go build ./cmd/api`. All checks passed successfully.
