# Roadmap Orchestrator Implementation

## Module Responsibilities
The `Roadmap` module is the authoritative orchestrator of the learner's learning path. It generates, retrieves, and persists the learner's `paths` and `path_items`. The orchestration bridges information from Goals, Knowledge, Competency, Progress, and Resources, using the AI Client for initial path proposals, but strictly enforcing domain validations before persistence.

## Owned Tables
- `platform.paths`
- `platform.path_items`

Note: `remediation_records` logic was initially identified as part of the roadmap domain, but full implementation logic for remediation creation relies on progress triggers which are orchestrated outside the basic regeneration flow. It remains under the Roadmap domain for future integration.

## Regeneration Flow
1. Fetch the learner's active `Goal`.
2. Fetch the learner's current `Competency` state.
3. Send this context via `AIClientService` to FastAPI to get a proposed path of concepts and resources.
4. **Validate**:
    - Deduplicate concept items.
    - Validate the proposed sequence respects prerequisite rules via `KnowledgeService`.
    - Validate any selected resources are published/available via `ResourcesService`.
5. **Persist Transactionally**:
    - Deactivate any existing active paths for this goal.
    - Insert the new `paths` and `path_items`.
6. Any failure in validation immediately rejects the AI proposal, ensuring Go remains authoritative.

## Cross-Module Interfaces Used
The orchestration is strictly decoupled using ports:
- `GoalsService` -> to verify goal bounds.
- `KnowledgeService` -> to verify prerequisite DAG.
- `CompetencyService` -> to determine which concepts are already competent.
- `ResourcesService` -> to verify the availability of resources.
- `AIClientService` -> to communicate with the FastAPI intelligence layer.

## Concurrency and Transaction Behavior
- **Transaction**: The old path is deactivated, and the new path with its path items are inserted atomically using `database.TxManager` -> `pgx.Tx`. This guarantees we never end up with a partial roadmap.
- **Concurrency**: While there's no distributed lock implemented for concurrent generation yet, the transactional update ensures that race conditions result in either both succeeding consecutively or a database constraint violation on `one_active_path_per_goal` index, preserving integrity.

## Authentication
`X-User-ID` is parsed from headers (typically set by identity middleware) to ensure the roadmap is strictly localized to the authenticated learner.

## Tests Added
- Tests for `RegenerateRoadmapUseCase` simulate the orchestration boundary using mocks for all 5 dependencies.
- A test specifically checks `ErrInvalidPrerequisite` if the AI proposes a path that fails the Knowledge service prerequisite validation, proving Go rejects invalid AI proposals.

## Build/Test Results
- `go vet ./internal/roadmap/...` - clean
- `go test -v ./internal/roadmap/...` - passed
- `go build -o /dev/null ./cmd/api` - succeeded

## AI Explainability
- The `GET /roadmap/concepts/{conceptId}/why` endpoint acts as a passthrough to the AI Client's explanation capabilities without polluting Go's persistent state with ephemeral explanations.
