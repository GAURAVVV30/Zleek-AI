# Projects Module Implementation

## Module Responsibilities
The `Projects` module provides an orchestration layer to facilitate project-based learning. Following the strict database constraint that projects do not have their own tables, this module acts purely as an application-level router, retrieving project details from the `Assessment` module and persisting project submissions to the `Progress/Evidence` module. 

## Project Representation
Projects are not stored in dedicated `projects` or `project_submissions` tables. Instead:
- **Definitions**: Projects are sourced from `assessment_definitions` where `type = 'project'`, owned by the `Assessment` module.
- **Submissions**: Submissions are recorded as `evidence_records`, owned by the `Progress` module.

## Endpoints Implemented
Matched exactly against `go-api.yaml`:
- `GET /concepts/{id}/project`
- `POST /concepts/{id}/project/submit`
- `GET /concepts/{id}/project/status`

## Cross-Module Boundaries
### Assessment Boundary
The module relies on an `AssessmentService` port to fetch `AssessmentDefinition` details for a concept, enforcing that the `type` is genuinely a project. It maps this data into a distinct `Project` domain struct.

### Evidence Boundary
The module relies on an `EvidenceService` port to ingest submissions as `ProjectSubmission` records. It does not perform SQL inserts into `evidence_records` directly, maintaining strict ownership boundaries. Note that:
> [!IMPORTANT]
> **Project submission is not automatically competency achievement.** The submission creates a pending evidence record which requires asynchronous evaluation (either by AI or Curator) before affecting the learner's competency state.

### Storage Boundary
A `StorageService` port validates the existence and integrity of the artifact references provided in the submission metadata. The file upload process (`GET /storage/upload-url`) is intentionally excluded from the domain logic of this module to maintain decoupling; it relies on secure URL/URI reference passing.

## Review and Status Lifecycle
The `GET /concepts/{id}/project/status` endpoint queries the `EvidenceService` for the latest submission status, returning states such as `pending`, `reviewed`, etc. This aligns with the asynchronous evaluation flow.

## Authentication
`X-User-ID` is parsed from headers (typically set by identity middleware) to ensure the project submissions and statuses are strictly scoped to the authenticated learner.

## Tests and Verification
- Application use-case tests mimic the orchestration boundary using mocks for all 3 dependencies (`Assessment`, `Evidence`, `Storage`).
- Explicit tests verify rejection behavior for invalid artifact references and non-existent project concepts.
- `go vet`, `go test`, and `go build` passed successfully.
