# US001/be_task_state_machine

**Requirement:** REQ003
**Story:** US001
**Track:** BE
**Service:** services/agent-board
**Status:** completed
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

### Review pass 1 — 2026-05-19 — verdict: approved
- Tests: `go vet ./...` clean; `go test ./...` inside `services/agent-board` — all packages PASS (domain, handler, mcp, repo).
- Review gate (BE): `REVIEW GATE: PASS` — gofmt -s, go vet, golangci-lint, go test, gosec, govulncheck all PASS.
- Review gate (cross): `REVIEW GATE: PASS` — semgrep (owasp/golang/typescript), gitleaks all PASS.
- Test contract satisfied: IT-001 covered by `TestTaskTools_UpdateTask_InvalidTransition` (handler/task_tools_test.go:233) — `pending->completed` rejected with `invalid transition` error, repo not called. UT-005 enforcement re-verified at handler layer via `TestTaskTools_CreateTask_EnforcesInitialStatus` (UT-005b, task_tools_test.go:326). Repo transactional behavior covered by `TestTaskRepo_UpdateTaskStatus` and `TestTaskRepo_UpdateTaskStatus_RollbackOnAuditFail` (task_repo_test.go:141, :173). UT-001..UT-005 domain tests carried over from the scaffold task and still pass.
- D-001 conformance: transition validation is in `internal/domain` (`Task.IsValidTransition`); handler calls `existing.IsValidTransition` before persisting (task_tools.go:122). No business logic leaked into transport.
- D-003 conformance: `UpdateTaskStatus` (task_repo.go:90) uses one `BeginTx`, UPDATEs task status then INSERTs into `status_audit_trail`, rolls back on any failure, commits last. Audit INSERT columns (`entity_id, entity_type, from_status, to_status`) match the architecture data model exactly; `entity_type` hard-set to `"task"`.
- `create_task` enforces initial `pending` via `domain.NewTask` — non-pending status rejected before any DB call (task_tools.go:60).
- OUT-OF-SCOPE EDIT (noted, non-blocking): `cmd/api-server/main.go:18` changed `middleware.Logger()` -> `middleware.RequestLogger()`. This file is not in the task's `## Files touched`. The dev disclosed it in `## Notes`; it is a one-line lint-fix required to pass the gate and the orchestrator's REQ003 merge already resolved this file in favor of main. Approved with the understanding that future drive-by lint fixes should be routed as a separate scaffold/chore task.
- FOLLOW-UP (non-blocking): in `update_task`, a combined status + title/description change performs a non-transactional `UpdateTask` then a separate `UpdateTaskStatus` transaction (task_tools.go:130-148). Status+audit atomicity (the task's actual requirement) is preserved; field+status atomicity is not, but it is outside this task's scope and the architecture. Consider consolidating if a future story requires it.
- Verdict: approved. Status flipped to `completed`.
