# US001/be_task_state_machine

**Requirement:** REQ003
**Story:** US001
**Track:** BE
**Service:** services/agent-board
**Status:** in_review
**Blocked by:** US001_be_scaffold_domain_and_migration.md
**Worked-by:** be-dev-2026-05-19T00:00:00Z-a7b5
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

## Notes

### Files touched
- `services/agent-board/internal/repo/task_repo.go` — Added `UpdateTaskStatus(ctx, id, fromStatus, toStatus)` to `TaskRepository` interface and `taskRepo` implementation. The method uses a single DB transaction to: (1) UPDATE task status, (2) INSERT into `status_audit_trail`. Rolls back on any failure.
- `services/agent-board/internal/repo/task_repo_test.go` — Added `TestTaskRepo_UpdateTaskStatus` (happy path) and `TestTaskRepo_UpdateTaskStatus_RollbackOnAuditFail` (rollback test).
- `services/agent-board/internal/handler/task_tools.go` — Updated `create_task` to use `domain.NewTask` (enforces `pending` initial status). Updated `update_task` to: validate state transition via `existing.IsValidTransition(newStatus)` before persisting; call `UpdateTaskStatus` (transactional) when status changes; call `UpdateTask` for non-status field changes only.
- `services/agent-board/internal/handler/task_tools_test.go` — Added `TestTaskTools_UpdateTask_InvalidTransition` (IT-001), `TestTaskTools_UpdateTask_ValidTransition` (IT-001b), `TestTaskTools_CreateTask_EnforcesInitialStatus` (UT-005b). Updated IT-019 (`TestTaskTools_UpdateTask`) to test title-only update (no status change) since status changes now route through `UpdateTaskStatus`.
- `services/agent-board/cmd/api-server/main.go` — Fixed pre-existing deprecated `middleware.Logger()` -> `middleware.RequestLogger()` to pass golangci-lint gate.

### Tests added
- `TestTaskRepo_UpdateTaskStatus` — repo layer, TX happy path
- `TestTaskRepo_UpdateTaskStatus_RollbackOnAuditFail` — repo layer, TX rollback
- `TestTaskTools_UpdateTask_InvalidTransition` — handler layer IT-001: invalid transition returns error
- `TestTaskTools_UpdateTask_ValidTransition` — handler layer IT-001b: valid transition calls UpdateTaskStatus
- `TestTaskTools_CreateTask_EnforcesInitialStatus` — handler layer UT-005b: non-pending initial status rejected

### Test counts
- `go test ./...` inside `services/agent-board`: 29 tests, all PASS.
- BE gate: PASS. Cross gate: PASS.

### Notes
- The existing domain tests (UT-001 through UT-005 for `IsValidTransition` and `NewTask`) already existed from the scaffold task and continue to pass.
- IT-019 was updated to test title-only updates (no status change) to correctly reflect the new behavior where status changes go through `UpdateTaskStatus` transactionally. The prior test was testing a valid transition (pending->in_progress) combined with a title update, which now routes differently.

## Review log
