# US006/be_task_tools_error_mapping_tests

**Requirement:** REQ006
**Story:** US006
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** none
**Worked-by:**
**Implements:** REQ006/US006 AC (all scenarios — 25 verbatim test function names lifting `RegisterTaskTools` from 67.4%, including the 5 distinct status-change branches of `UpdateTaskTool` + the no-status-change branch, ≥95% per-file coverage modulo §4.5 exemptions, no production-code change). Architecture §3 US006 touch row + §4.3 cluster-2 mock-repo pattern + §4.5 exemption mechanism + §4.6 local verification command (US006 row).

## Goal
Backfill `task_tools.go` IT-* error-mapping tests so per-file statement coverage clears ≥95%, with the 25 verbatim test functions named in US006 AC. `UpdateTaskTool` carries the bulk (5 distinct status-change branches + a no-status-change branch). Tests-only.

## Scope
- **In:** Edit `services/agent-board/internal/handler/task_tools_test.go` to add 25 test functions per US006 AC. Use a hand-written `MockTaskRepo` per architecture §4.3 (or extend an existing one if already present in the file). `testify/mock` acceptable but do not mix styles within the file.
- **Out:** Any change to `task_tools.go`. Any change to siblings (`project_tools*`, `document_tools*`, `user_story_tools*`, `message*`).

## Files touched (estimated, exclusive)
- `services/agent-board/internal/handler/task_tools_test.go` (edit — add `MockTaskRepo` if absent + 25 test functions)

## Test contract
Dev makes the 25 verbatim test-function names from US006 AC pass. Coverage includes:
- `TestRegisterTaskTools_RegistersAllFiveTools` (lifts the register function).
- The Create / Get / Update / Delete / List error matrix per architecture §4.3.
- `UpdateTaskTool`'s 5 distinct status-change branches + the `_StatusUnchanged` branch (these are the bulk).
- `_InvalidStatusTransition` (architecture §4.3 mapping — Get returns existing status A, invoke with `Status` field = status B where `existing.IsValidTransition(*req.Status)` is false).

Tester's `US006_be_unit_tests.md` IT-* IDs map 1:1.

## Implementation notes
- **`task_tools.go` WRAPS most repo errors** with `fmt.Errorf("failed to <op> task: %w", err)` (architecture §4.3 "assertion nuance" — first bullet). Assertion idiom: `assert.Contains(t, err.Error(), "failed to <op>")`.
- **`_NotFound` returns `fmt.Errorf("task not found")`** (fresh). Assert via `assert.Contains(t, err.Error(), "task not found")` AND `assert.False(t, errors.Is(err, repo.ErrNotFound))`.
- **Status-transition tests:** populate the mock's `GetTaskFunc` to return an existing task with a known status, then invoke `UpdateTaskTool` with body containing a `status` field that triggers each of the 5 valid transitions (covering the `UpdateTaskStatus`-repo path), plus a no-status-change body (covering the `UpdateTask`-repo path), plus an invalid transition (covering the `IsValidTransition` false branch).
- **Mock-repo shape (architecture §4.3):** hand-written `MockTaskRepo` embedding `repo.TaskRepository`. Func fields: `CreateTaskFunc`, `GetTaskFunc`, `UpdateTaskFunc`, `UpdateTaskStatusFunc`, `DeleteTaskFunc`, `ListTasksFunc`.
- **Read the source FIRST** for the exact wrap-prefix strings (`"failed to create task"`, `"failed to update task"`, etc.) and validation strings — assertions must match `task_tools.go` literally.
- **Coverage check command** (architecture §4.6, US006 row):
  ```
  cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out \
      -run "TestRegisterTaskTools|Test(Create|Get|Update|Delete|List)Task(s?)Tool"
  go tool cover -func=/tmp/handler.out | grep task_tools.go
  ```
  Must show ≥95% statement coverage on `task_tools.go`.

## Definition of done
- All 25 new test functions present with US006 AC's verbatim names; all green via the local verification command above.
- `cd services/agent-board && go vet ./... && go test ./...` clean across the whole module.
- `task_tools.go` ≥95% statement coverage (modulo any §4.5 exemptions named in the test report).
- `task_tools.go` byte-for-byte unchanged.
- `golangci-lint run ./...` clean.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` + `scripts/review/run-gate.sh cross` both `REVIEW GATE: PASS`.
- **Live e2e NOT required** (tests-only); instead 3 clean runs of `cd services/agent-board && go test -count=3 ./internal/handler -race`.
- Dev set status to `in_review`; tech-lead approved.

## Review log
