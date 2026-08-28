# Competency and Progress Module Implementation

## Modules Overview
Following architectural reconciliation, the modules are structured with distinct responsibilities:
- **Competency (`internal/competency`)**: Authoritative owner of the learner's mastery state. Owns `competency_records` and `competency_history`.
- **Progress (`internal/progress`)**: Owner of evidence, engagement, and progress aggregation. Owns `evidence_records` and `engagement_events`.

## Transaction Boundaries and Evidence Flow
The architecture requires evidence to be immutably recorded simultaneously with the competency state transition. 

To satisfy both domain boundaries and atomicity without leaking SQL across packages:
1. `Assessment` module validates the evaluation and emits an `AssessmentDomain.Evidence` struct to an application boundary port.
2. `Progress` module's `RecordEvidenceUseCase` initiates a database transaction (`pgx.Tx`).
3. `Progress` inserts the `evidence_record`.
4. `Progress` calls an application-level port exposed by `Competency` (`CompetencyService.UpdateWithEvidence`), passing the `tx` object along with the deterministically mapped `Result` (competent, weak, inconclusive).
5. `Competency` maps the result to a strict `CompetencyState` and inserts/updates `competency_records` and appends to `competency_history` using the same `tx`.
6. `Progress` commits the transaction.

If any failure occurs, the transaction rolls back, preventing orphaned evidence or unrecorded state transitions.

## Engagement vs Competency
> [!IMPORTANT]
> **Engagement is not competency evidence.**

Engagement events (e.g., viewing a resource) are processed by `POST /concepts/{id}/engagement` inside the Progress module. These are recorded immutably in `engagement_events` but *never* trigger a competency update. Competency is only updated through the deterministic Evidence transaction flow initiated by Assessment validations.

## AI Boundary
> [!IMPORTANT]
> **AI proposes; Go validates and persists authoritative competency state.**

AI is not directly invoked by the Competency or Progress modules in this step, as the AI contract (`POST /v1/evaluate`) was already implemented inside the `Assessment` module boundary. The `Progress` and `Competency` modules only receive deterministically constrained enums (`competent`, `weak`, `inconclusive`) and numeric scores mapped by Go after Go has validated the AI's proposal.

## Endpoints Implemented
- `GET /competency/detail?conceptId={id}`: Retrieves authoritative current mastery state.
- `GET /competency/{conceptId}/history`: Retrieves mastery audit log.
- `GET /progress/summary`: Aggregates in-progress vs competent counts.
- `GET /goals/current/completion-summary`: Integrates with the Goals module to summarize progress against a target.
- `POST /concepts/{id}/engagement`: Records stateless learner interaction.

## Authentication and Security
- Cross-user access is prevented natively. `X-Learner-ID` is extracted contextually (simulating middleware behavior).
- Endpoints do not blindly accept unverified parameters for cross-user aggregation.

## Tests
- Added deterministic state transition and evidence rollback tests in `application_test.go`.
- Validated compile boundaries. No SQL/pgx leaks exist in domain layers.
