# US001/be_scaffold_domain_and_migration

**Requirement:** REQ003
**Story:** US001
**Track:** BE
**Service:** services/agent-board
**Status:** in_progress
**Blocked by:** 
**Worked-by:** be-dev
**Implements:** D-001, D-002, Database Schema

## Goal
Scaffold the database migration and domain logic for the status state machines and audit trails.

## Scope
- **In:** PostgreSQL migrations for `status_audit_trail` table, domain models for state machine logic (Valid transitions for Task and User Story), and domain model for Audit Logs.
- **Out:** Updating handlers or repos to actually use this logic.

## Files touched (estimated, exclusive)
- `services/agent-board/migrations/000002_status_audit_trail.up.sql`
- `services/agent-board/migrations/000002_status_audit_trail.down.sql`
- `services/agent-board/internal/domain/status_machine.go`
- `services/agent-board/internal/domain/status_machine_test.go`
- `services/agent-board/internal/domain/audit_log.go`

This task is a scaffold task. Other tasks in the same requirement should `Blocked by:` this task so it runs solo before parallelising the rest.

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US001_be_unit_tests.md` and `US002_be_unit_tests.md`: Relevant unit tests for domain validation (check exact IDs in the test files).

## Implementation notes
- The `status_audit_trail` table schema is exactly as defined in `architecture.md`.
- `status_machine.go` should contain logic to check `IsValidTaskTransition(from, to)` and `IsValidStoryTransition(from, to)`.
- It should also have logic to validate the initial states: `pending` for Task, `draft` for Story.

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0, and `scripts/review/run-gate.sh cross` exits 0. The dev should run these locally before flipping to `in_review` — tech-lead will rerun them and reject on any failure.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Review log
