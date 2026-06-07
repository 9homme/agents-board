# US003 — Backfill `audit_repo.go` error-branch tests

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** done

## Story
As a **future contributor changing `services/agent-board/internal/repo/audit_repo.go`**, I want **every error branch in `getAuditTrail` (and its two public callers `GetTaskAuditTrail` / `GetUserStoryAuditTrail`) to be covered by `sqlmock`-driven tests**, so that a regression in the audit-trail read path (e.g. dropping a `fmt.Errorf` wrap or removing a `rows.Err()` check) fails CI immediately instead of silently shipping.

## Acceptance criteria

- **Scenario: `audit_repo_test.go` gains the following test functions (verbatim names)**
  - Given the existing `services/agent-board/internal/repo/audit_repo_test.go`
  - When the story is complete
  - Then the following new test functions exist:
    1. `TestAuditRepo_GetTaskAuditTrail_QueryError`
    2. `TestAuditRepo_GetTaskAuditTrail_ScanError`
    3. `TestAuditRepo_GetTaskAuditTrail_RowsErr`
    4. `TestAuditRepo_GetUserStoryAuditTrail_QueryError`
    5. `TestAuditRepo_GetUserStoryAuditTrail_ScanError`
    6. `TestAuditRepo_GetUserStoryAuditTrail_RowsErr`
  - **Note on coverage shape.** `getAuditTrail` (the private helper at `audit_repo.go:31`) has exactly three error branches — Query/Scan/RowsErr. The two public methods (`GetTaskAuditTrail` at line 54, `GetUserStoryAuditTrail` at line 60) are thin wrappers that pass a different `entityType` string. Covering the three error branches through BOTH wrappers (6 tests total) gives full coverage of both the helper AND both public entry points, AND verifies the `entity_type` argument is passed correctly to the underlying query.

- **Scenario: each new test exercises the specific uncovered branch**
  - Given an `sqlmock` DB constructed via `sqlmock.New()`
  - And the mock is configured per branch:
    - `WillReturnError(errors.New("db down"))` — for `_QueryError`
    - `WillReturnRows(...).AddRow(<wrong type for entry.ID>)` — for `_ScanError`
    - `WillReturnRows(...).RowError(rowIdx, err)` — for `_RowsErr`
  - When the corresponding public method (`GetTaskAuditTrail` or `GetUserStoryAuditTrail`) is invoked with a valid entity ID string
  - Then the test asserts the returned error is non-nil
  - And the error message contains the relevant `fmt.Errorf` wrap text:
    - `_QueryError` cases: error contains `"failed to query audit trail"` (line 34)
    - `_ScanError` cases: error contains `"failed to scan audit trail entry"` (line 42)
    - `_RowsErr` cases: error contains `"error iterating audit trail"` (line 47)
  - And the returned `[]*domain.StatusAuditLog` is nil
  - And `mock.ExpectationsWereMet()` returns nil
  - And the `ExpectQuery` matched a SQL string with `WHERE entity_type = $1 AND entity_id = $2` AND the first argument was either `"task"` (for `GetTaskAuditTrail` tests) or `"user_story"` (for `GetUserStoryAuditTrail` tests)

- **Scenario: per-file coverage hits ≥95%**
  - Given `cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo.out -run TestAuditRepo`
  - When `go tool cover -func=/tmp/repo.out | grep audit_repo.go` is inspected
  - Then `audit_repo.go` shows **≥95% statement coverage** (today's baseline per `docs/tech_debt.md` line 46: `getAuditTrail` at 78.6%)
  - And the only uncovered lines (if any) are documented in the test report under OQ-4

- **Scenario: existing tests still pass and behaviour is unchanged**
  - Given the production code in `audit_repo.go` is **NOT** modified by this story
  - When `cd services/agent-board && go test ./...` runs
  - Then all pre-existing tests pass
  - And all new tests pass
  - And `golangci-lint run ./...` is clean

- **Scenario: no production-code changes**
  - Given `git diff` of the story's commits
  - When inspected
  - Then **only** `services/agent-board/internal/repo/audit_repo_test.go` (and optionally a shared test helper) is modified
  - And `services/agent-board/internal/repo/audit_repo.go` is **byte-for-byte unchanged**

## UI / UX flow expectations
**No UI: BE-test only.**

## Out of scope
- **Modifying production audit-repo code.** Tests-only. If a backfill test surfaces a real bug, raise a new story.
- **`task_repo.go`** — US001.
- **`user_story_repo.go`** — US002.
- **`document_repo.go` / `project_repo.go`** — REQ005/US005.
- **Audit trail write-side coverage.** The audit trail is **written** by `task_repo.UpdateTaskStatus` and `user_story_repo.UpdateUserStoryStatus` — those write paths are covered in US001 and US002 respectively. US003 covers only the read path.

## Dependencies
- None. Independent of every other US in REQ006.

## Notes for the team

- **Smallest cluster-1 story.** Audit repo is a single private helper + two public wrappers, so only 6 tests required. Should be fast to land.
- **Audit reference for baseline number.** `docs/tech_debt.md` line 46: `getAuditTrail` at 78.6%.
- **The entity-type argument assertion matters.** The `_QueryError` tests for `GetTaskAuditTrail` vs `GetUserStoryAuditTrail` are otherwise identical — the ONLY difference is the `"task"` vs `"user_story"` argument passed to the query. Asserting via `ExpectQuery(...).WithArgs("task", taskID)` etc. ensures we have not silently swapped the entity-type passthrough.
- **Run locally before pushing:** `cd services/agent-board && go test ./internal/repo -cover -v -run TestAuditRepo`.

## Sign-off log
(po-ba appends here on each sign-off pass)

### Sign-off pass 1 — 2026-06-07 — verdict: approved
- **Spec review:** All 5 AC scenarios map to test cases in `US003_be_unit_tests.md`. The 6 verbatim test-function names (AC scenario 1) map 1:1 onto UT-001..UT-006. The branch-specific assertions (AC scenario 2 — non-nil error, exact `fmt.Errorf` wrap text per branch, nil result, `ExpectationsWereMet`, and the `WithArgs("task"|"user_story", id)` entity-type passthrough) are present in each UT-* case. Coverage ≥95% (AC scenario 3) is IT-001; full-suite regression + lint (AC scenarios 4 & 5) is IT-002. The three error branches (Query/Scan/RowsErr) × both public callers = 6 cases, which exhaustively covers `getAuditTrail` and both wrappers — pyramid is honest (all unit/integration; correctly no e2e for a tests-only backfill).
- **Result review:** `US003_test_report.md` reports 8 test IDs, 8 PASS, 0 FAIL, 0 skipped. Independently verified: `go test ./internal/repo -run TestAuditRepo` → 9 passed; `go tool cover` shows `audit_repo.go` at 100% on all four functions (`NewAuditRepo`, `getAuditTrail`, `GetTaskAuditTrail`, `GetUserStoryAuditTrail`) — exceeds the ≥95% target (baseline 78.6%). Production code unchanged confirmed independently: `git diff` of `internal/repo/audit_repo.go` against the story baseline is empty (tests-only). The 7-col vs 6-col `auditCols` note was correctly resolved by the dev to match the real 6-column production SELECT and validated at code review — descriptive spec context, not a contract gap, so no tester revision required.
- **Routed to:** none (approved).
