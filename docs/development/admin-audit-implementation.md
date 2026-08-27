# Admin and Audit Module Implementation

## Module Responsibilities
The `Admin` module provides Role-Based Access Control (RBAC) administrative functions, strictly segregating administrative responsibilities from regular business state operations. It manages high-privilege changes, specifically user roles and status mutations, and maintains an immutable audit trail.

## Database Ownership
- **`users` Table (Read/Update)**: The `users` table is strictly owned by the `Identity` module. 
> [!IMPORTANT]
> **Admin does not own the users table when Identity is the authoritative owner.**
> Admin mutates users exclusively via an application-level port to the `IdentityService`.
- **`audit_log` Table**: The `Admin` module is the authoritative owner of the `platform.audit_log` repository.

## Identity Dependency & Cross-Module Boundaries
Admin exposes user-management capability (like `GET /admin/users` and `PATCH /admin/users`) without establishing its own database connections to the users table. Instead, `Admin` depends on the `IdentityService` application boundary, passing through requests while layering necessary authorization and audit steps over them. (For this iteration, `IdentityService` is mocked at the DI boundary).

## RBAC Behavior
- All `/admin/*` endpoints intercept the `X-User-Role` from identity context.
- If the role is not `admin`, requests are instantly rejected with a `403 Forbidden` standard domain error, explicitly barring learners or curators.

## Safety Constraints
- **Invalid roles/statuses**: Attempts to assign unrecognized roles or non-documented statuses are rejected securely at the application layer.
- **Self-lockout / Self-demotion**: Administrators are explicitly barred by `domain.ErrSelfDemotion` from suspending their own account or removing their own admin role, ensuring at least one admin remains capable of accessing the system.

## Audit Logging & Append-Only Semantics
> [!IMPORTANT]
> **Audit records are append-only.**
The `AuditRepository` infrastructure (`PostgresAuditRepository`) deliberately omits `Update` and `Delete` methods. Administrative actions (like role changes) result in an audit record detailing `BeforeState` and `AfterState` payloads injected via atomic transactions alongside the mutation.

## Endpoints Implemented
Explicitly verified against `go-api.yaml`:
- `GET /admin/users`
- `PATCH /admin/users`
- `GET /admin/audit-log`

## Tests and Verification
- Unit and domain testing verifies RBAC and self-modification safeguards (e.g. asserting `domain.ErrSelfDemotion`).
- Validated via `go test`, `go vet`, and cleanly compiles inside both `cmd/api` and `cmd/worker`.
