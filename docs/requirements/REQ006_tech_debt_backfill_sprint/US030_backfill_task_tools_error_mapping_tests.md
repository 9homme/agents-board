# US030 — Backfill `task_tools.go` error-mapping tests

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** done

## Story
As a **future contributor changing `services/agent-board/internal/handler/task_tools.go`**, I want **every repo-error → MCP-error-envelope mapping AND every status-transition guard in `RegisterTaskTools` and its 5 underlying tool closures to be covered by integration tests**, so that a regression (e.g. swallowing `repo.ErrNotFound`, dropping the `existing.IsValidTransition` check, or breaking the audit-trail-coupled `UpdateTaskStatus` call) fails CI immediately.

## Acceptance criteria

- **Scenario: `task_tools_test.go` gains the following test functions (verbatim names)**
  - Given the existing `services/agent-board/internal/handler/task_tools_test.go`
  - When the story is complete
  - Then the following new test functions exist:
    1. `TestRegisterTaskTools_RegistersAllFiveTools` — covers `RegisterTaskTools` 67.4% by asserting the five tool names resolve.
    2. `TestCreateTaskTool_InvalidArguments`
    3. `TestCreateTaskTool_MissingUserStoryIDOrTitle` — empty userStoryId OR empty title → "userStoryId and title are required".
    4. `TestCreateTaskTool_DefaultStatusWhenOmitted` — req.Status omitted → defaults to `domain.TaskStatusPending` (verify the stored task's Status field).
    5. `TestCreateTaskTool_InvalidInitialStatus` — req.Status = "in_progress" (non-pending) → `domain.NewTask` returns err → handler wraps as `"invalid initial status: <err>"`.
    6. `TestCreateTaskTool_RepoError` — `CreateTask` returns err → wraps as `"failed to create task: <err>"`.
    7. `TestGetTaskTool_InvalidArguments`
    8. `TestGetTaskTool_EmptyID`
    9. `TestGetTaskTool_NotFound` — `repo.ErrNotFound` → `"task not found"`.
    10. `TestGetTaskTool_GenericError` — wraps as `"failed to get task: <err>"`.
    11. `TestUpdateTaskTool_InvalidArguments`
    12. `TestUpdateTaskTool_EmptyID`
    13. `TestUpdateTaskTool_NotFoundOnInitialGet`
    14. `TestUpdateTaskTool_GenericErrorOnInitialGet` — wraps as `"failed to get task: <err>"`.
    15. `TestUpdateTaskTool_InvalidStatusTransition` — req.Status != existing.Status AND `IsValidTransition` returns false → error contains `"invalid transition from <from> to <to>"`.
    16. `TestUpdateTaskTool_StatusChange_FieldUpdateError` — status change valid, title/description also provided, but `UpdateTask` (the field-update call) returns err → wraps as `"failed to update task fields: <err>"`.
    17. `TestUpdateTaskTool_StatusChange_UpdateTaskStatusError` — status change valid, `UpdateTaskStatus` returns err → wraps as `"failed to update task status: <err>"`.
    18. `TestUpdateTaskTool_NoStatusChange_RepoUpdateError` — no status change, `UpdateTask` returns err → wraps as `"failed to update task: <err>"`.
    19. `TestUpdateTaskTool_StatusChange_HappyPath` — status change valid, no extra field changes, `UpdateTaskStatus` returns updated task → response shape `TaskResponse` with new status.
    20. `TestDeleteTaskTool_InvalidArguments`
    21. `TestDeleteTaskTool_EmptyID`
    22. `TestDeleteTaskTool_RepoError` — wraps as `"failed to delete task: <err>"`.
    23. `TestListTasksTool_InvalidArguments`
    24. `TestListTasksTool_MissingUserStoryID`
    25. `TestListTasksTool_RepoError` — wraps as `"failed to list tasks: <err>"`.
  - **Status-transition coverage is the load-bearing job of this story.** `task_tools.go:96 UpdateTaskTool` has the most branches of any handler in the file — the invalid-transition, status-change-with-field-error, status-change-with-status-error, and no-status-change-with-update-error paths must all be tested explicitly.

- **Scenario: each new test exercises the specific uncovered branch**
  - Given a mock `repo.TaskRepository`
  - When the corresponding tool closure is invoked with a `json.RawMessage`
  - Then assertions per the test name:
    - `_InvalidArguments` → error contains `"invalid arguments"`
    - `_MissingUserStoryIDOrTitle` → error contains `"userStoryId and title are required"`
    - `_EmptyID` → error contains `"id is required"`
    - `_NotFound` → error message equals or contains `"task not found"`
    - `_GenericError` / `_RepoError` / `_*Error` → error contains the exact wrap prefix from the source
    - `_InvalidStatusTransition` → error message matches `"invalid transition from <from-status> to <to-status>"`
    - `_DefaultStatusWhenOmitted` → mock receives a `*domain.Task` with `Status == "pending"`
    - `_HappyPath` → returned `interface{}` is a `TaskResponse` with the expected fields (ID, Status, etc.)

- **Scenario: per-file coverage hits ≥95%**
  - Given `cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out -run "TestRegisterTaskTools|Test(Create|Get|Update|Delete|List)Task(s?)Tool"`
  - When `go tool cover -func=/tmp/handler.out | grep task_tools.go` is inspected
  - Then `task_tools.go` shows **≥95% statement coverage** (baseline per `docs/tech_debt.md` line 57: `RegisterTaskTools` 67.4%)
  - And the only uncovered lines (if any) are documented in the test report under OQ-4

- **Scenario: existing tests still pass and behaviour is unchanged**
  - Given `task_tools.go` is **NOT** modified by this story
  - When `cd services/agent-board && go test ./...` runs
  - Then all pre-existing tests pass
  - And all new tests pass
  - And `golangci-lint run ./...` is clean

- **Scenario: no production-code changes**
  - Given `git diff` of the story's commits
  - When inspected
  - Then **only** `services/agent-board/internal/handler/task_tools_test.go` (and optionally a shared mock-repo helper) is modified
  - And `services/agent-board/internal/handler/task_tools.go` is **byte-for-byte unchanged**

## UI / UX flow expectations
**No UI: BE-test only.**

## Out of scope
- **Modifying handler production code.** Tests-only.
- **`project_tools.go` / `document_tools.go` / `user_story_tools.go` / `message.go`** — US028, US029, US031, US032.
- **`domain.Task.IsValidTransition` unit tests** — already covered in `internal/domain`.

## Dependencies
- None. Independent.

## Notes for the team

- **`UpdateTaskTool` is the heaviest test target in the cluster.** Five distinct status-change branches plus the no-status-change branch. Tester should sub-group with `t.Run` if helpful.
- **Audit reference.** `docs/tech_debt.md` line 57 for the `RegisterTaskTools` baseline.
- **Mock task repo must support all 5 interface methods.** Hand-written mock is probably 40 lines; `testify/mock` is fine too.
- **Status-transition string format.** `fmt.Errorf("invalid transition from %s to %s", ...)` — assertion should match this format exactly.
- **Run locally before pushing:** `cd services/agent-board && go test ./internal/handler -cover -v -run "TestRegisterTaskTools|Test(Create|Get|Update|Delete|List)Task(s?)Tool"`.

## Sign-off log
(po-ba appends here on each sign-off pass)

### Sign-off pass 1 — 2026-06-07 — verdict: approved
- **Spec review:** All 5 AC scenarios covered. The 25 verbatim test-function names in AC scenario 1 map 1:1 to UT-001..UT-025 in `US030_be_unit_tests.md`. The load-bearing `UpdateTaskTool` branch matrix (invalid-transition, status-change-field-error, status-change-status-error, no-status-change-update-error, status-change-happy) is each explicitly covered (UT-015..UT-019). Per-branch assertion semantics (wrap-prefix `"failed to <op> task:"`, fresh `"task not found"` with `errors.Is(...)==false`, `"invalid transition from <from> to <to>"` format) are specified in the spec. Coverage-target AC (IT-001 ≥95%) and no-prod-change AC (scenario 5) both have explicit integration cases. No spec gaps.
- **Result review:** All 27 test IDs (UT-001..UT-025 + IT-001 + IT-002) report PASS in `US030_test_report.md`; 0 fail, 0 skipped. Independently verified, not just trusted:
  - `git diff HEAD -- internal/handler/task_tools.go` → empty (production file byte-for-byte unchanged) — satisfies AC scenario 5 and the "no production-code changes" requirement.
  - Coverage command reproduced: `mapTaskToResponse` 100.0%, `RegisterTaskTools` 95.3% — clears the ≥95% target (AC scenario 3). The 2 extra coverage-helper tests are legitimate (not in AC but needed to hit 95% under the narrow coverage-regex run); they do not weaken any spec case.
  - All 25 verbatim AC test names confirmed present in `task_tools_test.go`; `grep` for `t.Skip`/`Skip(` returned 0 — no skipped or tagged-out cases.
  - tech-lead review pass 1 approved with both review gates `REVIEW GATE: PASS` and `go test -count=3 -race` clean.
- **Routed to:** none (approved).
