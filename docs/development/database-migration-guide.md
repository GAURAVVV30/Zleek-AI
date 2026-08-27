# Database Migration Guide

This document describes how to manage database schemas for the Learning Platform Go Backend.

## Tooling
We use [`golang-migrate`](https://github.com/golang-migrate/migrate) to manage database migrations.
The migration runner is built into the Go backend repository at `cmd/migrate/main.go`.

## Prerequisites
- A running PostgreSQL instance.
- The `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB` environment variables must be configured in `.env` or exported.

## Creating a Migration
Migrations are stored in the `services/api-go/migrations/` directory.

To create a new migration, use the `migrate` CLI tool from `golang-migrate` (if installed globally) or simply create the files manually:
```bash
touch migrations/000002_add_new_table.up.sql
touch migrations/000002_add_new_table.down.sql
```
Make sure the filenames follow the `{version}_{title}.{up|down}.sql` format.

**Rules for Migrations:**
- Do not write business logic into SQL files.
- Ensure every business table is assigned an authoritative Go module ownership in the `module-database-map.md`.
- Ensure migrations apply cleanly to the `platform` or `intelligence` schemas depending on ownership. FastAPI owns `intelligence`, but Go manages the schema creations here.

## Running Migrations
Build and run the migration binary from the `services/api-go` directory:

```bash
# Build the migrate binary
go build -o bin/migrate ./cmd/migrate

# Run all UP migrations
./bin/migrate up

# Run DOWN migrations (rollback)
./bin/migrate down

# Check current migration version
./bin/migrate version
```

## Database Ownership Boundary
The Go service is the authoritative owner of the `platform` schema, which contains all business state (users, competency, roadmap, evidence).
The FastAPI service only owns the `intelligence` schema (embeddings, ML cache).
FastAPI MUST NEVER write to the `platform` schema. Go manages the migration files for both schemas to ensure the structure is deployable via a single pipeline.
