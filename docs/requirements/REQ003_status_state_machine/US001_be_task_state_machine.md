# US001/be_task_state_machine

**Requirement:** REQ003
**Story:** US001
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** US001_be_scaffold_domain_and_migration.md
**Worked-by:** 
**Implements:** US001, D-003, part of US003

## Goal
Enforce the Task status state machine on creation and updates, and transactionally write the state transitions to the audit trail.

## Scope
- **In:** Update `internal/repo` and `internal/handler` for Tasks to validate state transitions and initial state, and write to the audit trail on change.
- **Out:** Reading from the audit trail (handled in US003). Story status state machine.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/repo/task_repo.go`
- `services/agent-board/internal/repo/task_repo_test.go`
- `services/agent-board/internal/handler/task_tools.go`
- `services/agent-board/internal/handler/task_tools_test.go`

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US001_be_unit_tests.md`: Check specific UT and IT IDs for the task state machine and audit writing.

## Implementation notes
- Handlers should use the domain logic from `status_machine.go` to validate transitions/initial state before applying them.
- Updates must be transactional (update task status and insert into `status_audit_trail` in one DB transaction).
- `create_task` should ensure the initial status defaults to or is explicitly set to `pending`.

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Review log
