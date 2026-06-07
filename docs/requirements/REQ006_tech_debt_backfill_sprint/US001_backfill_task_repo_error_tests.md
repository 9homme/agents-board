# US001 — Backfill `task_repo.go` error-branch tests

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** done

## Story
As a **future contributor changing `services/agent-board/internal/repo/task_repo.go`**, I want **every error branch in `CreateTask`, `GetTask`, `UpdateTask`, `UpdateTaskStatus`, and `ListTasks` to be covered by `sqlmock`-driven tests**, so that a regression in error handling (e.g. dropping a `fmt.Errorf` wrap, removing a `rows.Err()` check, or breaking the `sql.ErrNoRows → ErrNotFound` mapping) fails CI immediately instead of silently shipping.

## Acceptance criteria

- **Scenario: `task_repo_test.go` gains the following test functions (verbatim names)**
  - Given the existing `services/agent-board/internal/repo/task_repo_test.go`
  - When the story is complete
  - Then the following new test functions exist (names are authoritative; tester may add additional cases but MUST NOT rename these):
    1. `TestTaskRepo_CreateTask_GenericError`
    2. `TestTaskRepo_GetTask_GenericError`
    3. `TestTaskRepo_UpdateTask_NotFound`
    4. `TestTaskRepo_UpdateTask_GenericError`
    5. `TestTaskRepo_UpdateTaskStatus_BeginTxError`
    6. `TestTaskRepo_UpdateTaskStatus_NotFound`
    7. `TestTaskRepo_UpdateTaskStatus_UpdateGenericError`
    8. `TestTaskRepo_UpdateTaskStatus_AuditInsertError`
    9. `TestTaskRepo_UpdateTaskStatus_CommitError`
    10. `TestTaskRepo_ListTasks_QueryError`
    11. `TestTaskRepo_ListTasks_ScanError`
    12. `TestTaskRepo_ListTasks_RowsErr`
  - **Note on UpdateTask split.** `task_repo.go:70 UpdateTask` returns both `ErrNotFound` (line 82) AND a generic error (line 84). The split into `_NotFound` + `_GenericError` mirrors REQ005/US005's `document_repo` / `project_repo` precedent.
  - **Note on UpdateTaskStatus.** This method runs inside a transaction (`BeginTx` → `QueryRowContext` → `ExecContext` → `Commit`) with five distinct error exits — five separate tests are required to cover all of them. This is more than the user's "5 functions" framing because `UpdateTaskStatus` is itself a multi-branch method; po-ba accepts the expansion. Tester may consolidate via `t.Run` sub-tests within a parent `TestTaskRepo_UpdateTaskStatus_Errors` if preferred, as long as every branch above is asserted.

- **Scenario: each new test exercises the specific uncovered branch**
  - Given an `sqlmock` DB constructed via `sqlmock.New()`
  - And the mock is configured to return one of:
    - `WillReturnError(errors.New("db down"))` — for `_GenericError`, `_QueryError`, `_BeginTxError`, `_CommitError` cases
    - `WillReturnError(sql.ErrNoRows)` — for `_NotFound` cases (the repo maps this to `repo.ErrNotFound`)
    - `WillReturnRows(...).AddRow(<wrong type>)` — for `_ScanError`
    - `WillReturnRows(...).RowError(rowIdx, err)` — for `_RowsErr`
    - `ExpectExec(...).WillReturnError(err)` — for `_AuditInsertError`
    - `ExpectCommit().WillReturnError(err)` — for `_CommitError`
  - When the corresponding repo method is invoked with valid input arguments
  - Then the test asserts the returned error is non-nil
  - And for `_NotFound` cases specifically: `errors.Is(err, repo.ErrNotFound)` returns true
  - And for non-`_NotFound` cases: `errors.Is(err, repo.ErrNotFound)` returns false
  - And for `_BeginTxError` / `_AuditInsertError` / `_CommitError`: the error message contains the relevant `fmt.Errorf("failed to ...: %w", err)` wrap text (substring match acceptable)
  - And `mock.ExpectationsWereMet()` returns nil

- **Scenario: per-file coverage hits ≥95%**
  - Given `cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo.out -run TestTaskRepo`
  - When `go tool cover -func=/tmp/repo.out | grep task_repo.go` is inspected
  - Then `task_repo.go` shows **≥95% statement coverage** (today's baseline per `docs/tech_debt.md`: `UpdateTask` 62.5%, `UpdateTaskStatus` 71.4%, `CreateTask` 80.0%, `GetTask` 87.5%, `ListTasks` 81.2%)
  - And the only uncovered lines (if any) are documented in the test report with a "genuinely unreachable via sqlmock" justification (see OQ-4 in README)

- **Scenario: existing tests still pass and behaviour is unchanged**
  - Given the production code in `task_repo.go` is **NOT** modified by this story
  - When `cd services/agent-board && go test ./...` runs
  - Then all pre-existing tests pass
  - And all new tests pass
  - And `golangci-lint run ./...` is clean (no new lint issues from the test additions)

- **Scenario: no production-code changes**
  - Given `git diff` of the story's commits
  - When inspected
  - Then **only** `services/agent-board/internal/repo/task_repo_test.go` (and optionally a shared test helper) is modified
  - And `services/agent-board/internal/repo/task_repo.go` is **byte-for-byte unchanged**

## UI / UX flow expectations
**No UI: BE-test only.**

## Out of scope
- **Modifying production repo code.** Test-only story. If a test surfaces an actual bug (e.g. a missing `rows.Err()` check), raise it as `ARCHITECTURE_GAP_FOUND` or a new follow-up story — do not silently fix.
- **`user_story_repo.go`** — that's US002.
- **`audit_repo.go`** — that's US003.
- **`document_repo.go` / `project_repo.go`** — shipped in REQ005/US005.
- **Adding new test helpers** unless they are genuinely shared across ≥3 of the new tests (no premature abstraction).

## Dependencies
- None. Independent of every other US in REQ006.

## Notes for the team

- **REQ005/US005 is the pattern reference.** Re-read `docs/requirements/REQ005_quality_hardening_retrospective/US005_backfill_repo_error_branch_tests.md` — this story is its lineal descendant.
- **Audit reference for baseline numbers.** `docs/tech_debt.md` lines 36–40 give the per-function pre-baseline coverage for `task_repo.go`.
- **`UpdateTaskStatus` is the heaviest test target.** Five distinct error exits inside a transaction — tester should plan for 5 sub-tests (or 5 top-level tests if preferred). The rollback `log.Printf` path (line 99) is fine to leave uncovered; flag it in the test report under OQ-4.
- **Run locally before pushing:** `cd services/agent-board && go test ./internal/repo -cover -v -run TestTaskRepo` should show all new tests pass and per-file coverage ≥95%.
- **`golangci-lint` parity.** Run `golangci-lint run ./...` inside `services/agent-board/` before flipping to `in_review`.

## Sign-off log
(po-ba appends here on each sign-off pass)

### Sign-off pass 1 — 2026-06-07 — verdict: approved
- **Spec review:** All five AC scenarios are covered. The 12 verbatim test-function names in the AC map 1:1 onto UT-001..UT-013 in `US001_be_unit_tests.md` (UT-003 `TestTaskRepo_GetTask_NotFound` is the exhaustiveness addition explicitly permitted by the AC note "tester may add additional cases"). Each error branch in `CreateTask`, `GetTask`, `UpdateTask`, `UpdateTaskStatus` (all 5 transactional exits), and `ListTasks` (query / scan / rows.Err) is exercised with the sqlmock idiom and the assertion set the AC demands (non-nil error; `errors.Is(err, repo.ErrNotFound)` true only for `_NotFound` cases; wrap-prefix substring for begin/audit/commit; `mock.ExpectationsWereMet()` nil). Coverage AC (IT-001) and regression AC (IT-002) are spec'd. The one coverage exemption (`task_repo.go:99` rollback `log.Printf`) is honestly documented under OQ-4 / §4.5 — genuinely sqlmock-unreachable, not a skipped requirement. Pyramid is honest: all unit-level, no e2e padding (BE-test-only story).
- **Result review:** Test report shows 15/15 test IDs PASS, 0 FAIL, no skipped tests. Independently re-verified: `go test ./internal/repo -run TestTaskRepo` → 20 passed; full module `go test ./...` → 301 passed in 7 packages; all 13 verbatim function names present in `task_repo_test.go`; no `t.Skip`/`SkipNow` anywhere in the file. Per-file coverage on `task_repo.go` re-measured: six of seven functions at 100%, `UpdateTaskStatus` at 95.2% — well above the ≥95% AC threshold, with the sole uncovered line being the exempted `:99` rollback log. Production code unchanged confirmed: `git diff main -- services/agent-board/internal/repo/task_repo.go` is empty (tests-only backfill, as required by the "no production-code changes" AC).
- **Routed to:** none — approved. Story set to `done`.
