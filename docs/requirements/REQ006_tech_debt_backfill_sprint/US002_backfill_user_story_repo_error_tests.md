# US002 — Backfill `user_story_repo.go` error-branch tests

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** in_signoff

## Story
As a **future contributor changing `services/agent-board/internal/repo/user_story_repo.go`**, I want **every error branch in `CreateUserStory`, `GetUserStory`, `UpdateUserStory`, `UpdateUserStoryStatus`, and `ListUserStories` to be covered by `sqlmock`-driven tests**, so that a regression in error handling (e.g. dropping the `sql.ErrNoRows → ErrNotFound` mapping or breaking the transactional `UpdateUserStoryStatus` audit-write) fails CI immediately instead of silently shipping.

## Acceptance criteria

- **Scenario: `user_story_repo_test.go` gains the following test functions (verbatim names)**
  - Given the existing `services/agent-board/internal/repo/user_story_repo_test.go`
  - When the story is complete
  - Then the following new test functions exist (names authoritative; tester may add more):
    1. `TestUserStoryRepo_CreateUserStory_GenericError`
    2. `TestUserStoryRepo_GetUserStory_GenericError`
    3. `TestUserStoryRepo_UpdateUserStory_NotFound`
    4. `TestUserStoryRepo_UpdateUserStory_GenericError`
    5. `TestUserStoryRepo_UpdateUserStoryStatus_BeginTxError`
    6. `TestUserStoryRepo_UpdateUserStoryStatus_NotFound`
    7. `TestUserStoryRepo_UpdateUserStoryStatus_UpdateGenericError`
    8. `TestUserStoryRepo_UpdateUserStoryStatus_AuditInsertError`
    9. `TestUserStoryRepo_UpdateUserStoryStatus_CommitError`
    10. `TestUserStoryRepo_ListUserStories_QueryError`
    11. `TestUserStoryRepo_ListUserStories_ScanError`
    12. `TestUserStoryRepo_ListUserStories_RowsErr`
  - **Note on UpdateUserStory split.** `user_story_repo.go:97 UpdateUserStory` returns both `ErrNotFound` (line 102) AND a generic error (line 104). Same pattern as REQ005/US005's `_NotFound` + `_GenericError` split.
  - **Note on UpdateUserStoryStatus.** Same transactional shape as `task_repo.UpdateTaskStatus` (`BeginTx` → `QueryRowContext` → `ExecContext` → `Commit`). Five distinct error exits — five tests. Tester may use `t.Run` sub-tests under a parent `TestUserStoryRepo_UpdateUserStoryStatus_Errors` if preferred.

- **Scenario: each new test exercises the specific uncovered branch**
  - Given an `sqlmock` DB constructed via `sqlmock.New()`
  - And the mock is configured (per branch) per the same pattern as US001 (`WillReturnError` / `sql.ErrNoRows` / wrong-type `AddRow` / `RowError` / `ExpectExec().WillReturnError` / `ExpectCommit().WillReturnError`)
  - When the corresponding repo method is invoked with valid input arguments
  - Then the test asserts the returned error is non-nil
  - And for `_NotFound` cases: `errors.Is(err, repo.ErrNotFound)` returns true
  - And for non-`_NotFound` cases: `errors.Is(err, repo.ErrNotFound)` returns false
  - And `mock.ExpectationsWereMet()` returns nil

- **Scenario: per-file coverage hits ≥95%**
  - Given `cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo.out -run TestUserStoryRepo`
  - When `go tool cover -func=/tmp/repo.out | grep user_story_repo.go` is inspected
  - Then `user_story_repo.go` shows **≥95% statement coverage** (today's baseline per `docs/tech_debt.md`: `UpdateUserStory` 57.1%, `UpdateUserStoryStatus` 76.2%, `ListUserStories` 76.5%, `CreateUserStory` 80.0%, `GetUserStory` 87.5%)
  - And the only uncovered lines (if any) are documented in the test report with a "genuinely unreachable via sqlmock" justification (OQ-4 in README)

- **Scenario: existing tests still pass and behaviour is unchanged**
  - Given the production code in `user_story_repo.go` is **NOT** modified by this story
  - When `cd services/agent-board && go test ./...` runs
  - Then all pre-existing tests pass
  - And all new tests pass
  - And `golangci-lint run ./...` is clean

- **Scenario: no production-code changes**
  - Given `git diff` of the story's commits
  - When inspected
  - Then **only** `services/agent-board/internal/repo/user_story_repo_test.go` (and optionally a shared test helper) is modified
  - And `services/agent-board/internal/repo/user_story_repo.go` is **byte-for-byte unchanged**

## UI / UX flow expectations
**No UI: BE-test only.**

## Out of scope
- **Modifying production repo code.** If a backfill test surfaces a real bug, raise a new story.
- **`task_repo.go`** — US001.
- **`audit_repo.go`** — US003.
- **`document_repo.go` / `project_repo.go`** — REQ005/US005.

## Dependencies
- None. Independent of every other US in REQ006.

## Notes for the team

- **Same shape as US001** but mirrored onto user stories. Tester may copy the US001 test structure file-for-file and substitute names — the two repos have the same interface shape.
- **Audit reference for baseline numbers.** `docs/tech_debt.md` lines 41–45 give per-function pre-baseline.
- **`UpdateUserStoryStatus` rollback `log.Printf` path** (line 68) is fine to leave uncovered; flag in test report.
- **Run locally before pushing:** `cd services/agent-board && go test ./internal/repo -cover -v -run TestUserStoryRepo`.

## Sign-off log
(po-ba appends here on each sign-off pass)
