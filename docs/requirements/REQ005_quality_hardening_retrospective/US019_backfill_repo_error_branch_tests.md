# US019 — Backfill 14 repo error-branch tests in `internal/repo`

**Requirement:** REQ005 — quality hardening retrospective
**Status:** draft

## Story
As a **future contributor changing `internal/repo/project_repo.go` or `document_repo.go`**, I want the **generic DB-error wrap branches, `rows.Scan` failures, and `rows.Err()` checks to be covered by `sqlmock`-driven tests**, so that a regression in error-handling (e.g. dropping a `fmt.Errorf` wrap or removing a `rows.Err()` check) fails CI immediately instead of silently shipping.

## Acceptance criteria

- **Scenario: `document_repo_test.go` gains 7 new error-branch tests**
  - Given the existing `services/agent-board/internal/repo/document_repo_test.go`
  - When the story is complete
  - Then the following test functions exist (names taken verbatim from REQ004 audit §4.4):
    1. `TestDocumentRepo_CreateDocument_GenericError`
    2. `TestDocumentRepo_GetDocument_GenericError`
    3. `TestDocumentRepo_UpdateDocument_NotFound`
    4. `TestDocumentRepo_UpdateDocument_GenericError`
    5. `TestDocumentRepo_DeleteDocument_GenericError`
    6. `TestDocumentRepo_ListDocuments_QueryError`
    7. `TestDocumentRepo_ListDocuments_ScanError`
    8. `TestDocumentRepo_ListDocuments_RowsErr`
  - **Note:** the audit lists 7 functions but UpdateDocument is split into two (`_NotFound` and `_GenericError`) per the audit narrative. That makes 8 for `document_repo` and 8 for `project_repo` if the same split applies — total can be 14 or 16 depending on how the symmetric Update split is counted. See "Notes for the team" — tester / tech-lead settles the exact count, AC requires AT LEAST the names listed.

- **Scenario: `project_repo_test.go` gains the symmetric error-branch tests**
  - Given the existing `services/agent-board/internal/repo/project_repo_test.go`
  - When the story is complete
  - Then the following test functions exist (names mirror the document set, per audit §4.4):
    1. `TestProjectRepo_CreateProject_GenericError`
    2. `TestProjectRepo_GetProject_GenericError`
    3. `TestProjectRepo_UpdateProject_NotFound`
    4. `TestProjectRepo_UpdateProject_GenericError`
    5. `TestProjectRepo_DeleteProject_GenericError`
    6. `TestProjectRepo_ListProjects_QueryError`
    7. `TestProjectRepo_ListProjects_ScanError`
    8. `TestProjectRepo_ListProjects_RowsErr`

- **Scenario: each test exercises the specific uncovered branch**
  - For each new test, the AC requires (Given/When/Then below applies generically to all 16):
    - Given an `sqlmock` DB
    - And the mock is configured to return either: `WillReturnError(errors.New("db down"))` (for QueryError / GenericError / DeleteError cases), or a row whose column type mismatches the destination (for ScanError), or `RowError(rowIdx, err)` (for RowsErr), or `sql.ErrNoRows` (for `_NotFound`)
    - When the corresponding repo method is invoked
    - Then the test asserts the returned error is non-nil
    - And the error message contains the relevant `fmt.Errorf("failed to <op> ...: %w", err)` wrap text (substring match acceptable)
    - And for `_NotFound` cases specifically: `errors.Is(err, repo.ErrNotFound)` returns true
    - And for `_GenericError` cases specifically: `errors.Is(err, repo.ErrNotFound)` returns false
    - And `mock.ExpectationsWereMet()` returns nil

- **Scenario: per-file coverage hits ≥95%**
  - Given `cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo.out`
  - When `go tool cover -func=/tmp/repo.out` is inspected
  - Then `project_repo.go` and `document_repo.go` each show ≥95% statement coverage (today's baseline per audit §1.1 is 62.5–88.9% per function, ~81.5% overall package)
  - And package-level `internal/repo` total is ≥95%
  - And the only uncovered lines are those that genuinely cannot be reached by sqlmock (call them out explicitly in the test report)

- **Scenario: existing tests still pass and behaviour is unchanged**
  - Given the production code in `project_repo.go` and `document_repo.go` is NOT modified by this story (test-only addition)
  - When `go test ./...` runs in `services/agent-board`
  - Then all pre-existing tests pass
  - And all 16 new tests pass
  - And `golangci-lint run ./...` is clean (no new lint issues from the test additions)

- **Scenario: no production-code changes**
  - Given a `git diff` of the story's commits
  - When inspected
  - Then only `*_test.go` files (and possibly a new test helper) are modified — no edits to `project_repo.go` or `document_repo.go`

## UI / UX flow expectations

**No UI:** test-coverage debt cleanup. Flow: future contributor edits a repo method, CI runs, regression caught — no user-visible behaviour change today.

## Out of scope
- **Modifying the production repo code.** This is a test-only story. If a test reveals an actual bug (e.g. a missing `rows.Err()` check), raise it as an `ARCHITECTURE_GAP_FOUND` or a new story — do not silently fix in this scope.
- **Backfilling handler-side tests.** Handlers are already at 100% per audit §1.1.
- **Refactoring existing tests** that already exist for happy paths. Add new tests; do not rewrite.
- **Adding `cmd/api-server` or `internal/mcp` coverage** (those are the packages dragging total coverage to 65.5%). Those have lower ROI and need their own design pass — out of this story.

## Dependencies
- None. Independent of every other US in REQ005.

## Notes for the team

- **Audit reference is authoritative.** `REQ004_quality_audit.md` §4.4 spells out each function name and the exact `sqlmock` mock pattern needed (`WillReturnError`, `AddRow` with wrong type, `RowError(rowIdx, err)`). Re-read it before writing tests.
- **Test-count nuance.** The audit's prose says "14 backfill tests" but enumerating both `Update_NotFound` AND `Update_GenericError` (symmetric for both repos) gives 16 functions. po-ba accepts either count — AC pins the AT-LEAST list above. If tech-lead / tester wants to fold both Update branches into one parametrised test, that's fine as long as both branches are asserted.
- **Coverage threshold of 95%.** Picked because audit projects "~95%" with these additions. If tester can show a specific line that is genuinely unreachable via sqlmock, document it in the test report with a justification and the threshold becomes "≥95% modulo enumerated unreachable lines."
- **No new test helpers required** unless the tests share enough setup to justify one. Each test is short — `sqlmock`-driven repo error tests typically run 15-25 lines each.
- **Run locally before pushing:** `cd services/agent-board && go test ./internal/repo -cover -v` should show all 16 new tests pass and the per-file coverage number meets AC.

## Sign-off log
(po-ba appends here on each sign-off pass)
