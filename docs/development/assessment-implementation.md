# Assessment Module Implementation

## Module Responsibilities
The Assessment module (`internal/assessment`) is responsible for providing assessment definitions/items to learners and orchestrating the submission and evaluation of those assessments. It strictly relies on the architectural paradigm where FastAPI proposes evaluation grades, Go validates them, and the result is forwarded to the Evidence boundary without Assessment mutating competency states directly.

## Owned Tables
As verified by the schema mappings, Assessment owns:
- `platform.assessment_definitions`
- `platform.assessment_items`

## Endpoints
- `GET /concepts/{id}/assessment`
  Retrieves an assessment and its items. Answer keys are explicitly stripped from the response.
- `POST /concepts/{id}/assessment/submit`
  Accepts a JSON payload (the learner's answers), delegates evaluation to the AI Service, validates the result, and records evidence.

## Assessment Lifecycle & AI Evaluation Flow
1. Submission data is sent by the authenticated learner.
2. The `SubmitAssessmentUseCase` receives the payload and checks if the concept and assessment exist.
3. The AI client is invoked (`POST /v1/evaluate`) with the `submissionData` and the assessment's `rubric`.
4. The AI responds with an `EvaluationResult` containing a numeric score, a numeric confidence, and an enum `Result` (e.g. `competent`, `weak`).
5. Go validates the response constraints (e.g., Score between 0-100, Result matches the strict set). Invalid or hallucinated responses are blocked and return a domain error.
6. The evidence payload is generated and handed off.

## Evidence and Competency Boundary
The Assessment module **does not** import or write to `platform.evidence_records`, `platform.competency_records`, or `platform.competency_history`. Instead, an application-layer port called `EvidenceService` is defined:
```go
type EvidenceService interface {
	RecordEvidence(ctx context.Context, evidence *domain.Evidence) error
}
```
In `cmd/api/main.go`, a mock `progress.MockEvidenceService` handles this hand-off. When the Progress module is fully implemented, it will provide the real implementation of this interface. 

## Authentication and Error Behavior
- Simulated Learner ID context mapping is utilized inside the HTTP handler to extract `X-Learner-ID`.
- Domain-level errors are strictly mapped to `400 Bad Request` or `404 Not Found` to prevent SQL/internal stack leaks.

## Testing and Verification
Unit tests in `application/assessment_test.go` verify the validation flows, specifically ensuring hallucinated/invalid AI evaluations are rejected and not recorded as evidence.
All builds and formatting checks passed.
