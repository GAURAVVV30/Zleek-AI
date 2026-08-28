# Goals Module Implementation

This document describes the implementation of the Goals Module.

## Architectural Flow
The Goals module strictly enforces the required paradigm: **FastAPI proposes -> Go validates -> Go persists**.

1. The HTTP Handler (`interfaces/http`) receives a request to create a new goal.
2. The Application Use Case (`CreateGoalUseCase`) orchestrates the boundary crossings:
   - First, it calls `aiclient.Client.ProposeGoalMapping()` to query FastAPI. The AI maps the unstructured goal text to a `knowledge_structure_id`.
   - Second, it calls `knowledge.Service.ValidateStructure()` to confirm the proposed structure ID exists within Go's authoritative platform database schema.
   - Third, if validation passes, it delegates to the Infrastructure layer to save the Goal entity.
3. The Infrastructure Layer (`postgres_goal_repository`) saves the entity directly to the `platform.goals` PostgreSQL table using `pgxpool`.

## Supported Endpoints
- `POST /goals`: Creates a new learner goal by mapping free-text intent to a knowledge structure via AI and saving it to the database.
- `GET /goals/current`: Retrieves the active goal for the authenticated learner.

## Database
- Schema: `platform.goals`
- Columns Managed: `id`, `learner_id`, `goal_text`, `knowledge_structure_id`, `status`, `achieved_at`, `created_at`
- Existing active goals are automatically marked as `abandoned` when a new goal is successfully verified and created.

## Module Ownership
- Code Path: `services/api-go/internal/goals/`
- The `aiclient` boundary and `knowledge` boundary interfaces were defined within the Goals Application layer and mock stubs were created to support integration in absence of the other fully completed modules.
