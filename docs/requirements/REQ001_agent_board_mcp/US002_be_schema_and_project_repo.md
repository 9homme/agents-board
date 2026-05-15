# US002/be_schema_and_project_repo

**Requirement:** REQ001
**Story:** US002
**Track:** BE
**Service:** services/agent-board-mcp
**Status:** completed
**Blocked by:** 
**Worked-by:** be-dev-20231024-1234
**Implements:** PostgreSQL Schema (all tables), D-002, Project Data Model

## Goal
Create the comprehensive database migration for all entities and implement the Project repository for CRUD operations.

## Scope
- **In:** `migrations/`, `internal/domain/project.go`, `internal/repo/db.go`, `internal/repo/project_repo.go`.
- **Out:** Exposing the MCP tools for Project (handled in a separate task).

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US002_be_unit_tests.md`: IT-* tests for database migrations and Project repository CRUD operations.

## Implementation notes
- Add `DB_URL` configuration handling in the main entrypoint and pass db connection to repos.
- Write the PostgreSQL schema in `migrations/` exactly as specified in the architecture, including `ON DELETE CASCADE` for all foreign keys.
- Implement standard CRUD methods in `project_repo.go` using `database/sql` or `sqlx` (preferred).
- Ensure standard Go error wrapping (`fmt.Errorf("...: %w", err)`).

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Review log
### Review pass 1 — 2024-05-16 — verdict: approved
- `go vet` and `go test ./...` pass clean.
- `migrations/000001_init_schema.up.sql` matches the architecture perfectly including `ON DELETE CASCADE`.
- Repository methods use standard Go errors wrapping. Approved.

## Notes
- Created `internal/domain/project.go` with exact JSON tag names and properties from the architecture.
- Added db schema to `migrations/000001_init_schema.up.sql` and down scripts.
- Implemented `ProjectRepository` in `internal/repo/project_repo.go` with tests passing `UT-005` to `UT-009`.
- Added database connection and `DB_URL` handling to `cmd/agent-board-mcp/main.go`.
