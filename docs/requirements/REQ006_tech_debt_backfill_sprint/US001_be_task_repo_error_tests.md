# US001/be_task_repo_error_tests

**Requirement:** REQ006
**Story:** US001
**Track:** BE
**Service:** services/agent-board
**Status:** in_review
**Blocked by:** none
**Worked-by:** be-dev-2026-06-05T00:00:00Z-aa86
**Implements:** REQ006/US001 AC (all four scenarios — 12 verbatim test function names, ≥95% per-file coverage modulo §4.5 exemptions, no production-code change, existing suite still green). Architecture §3 US001 touch row + §4.2 cluster-1 sqlmock pattern + §4.5 exemption mechanism + §4.6 local verification command (US001 row).

## Goal
Backfill `task_repo.go` error-branch tests so per-file statement coverage clears ≥95% (modulo enumerated unreachable lines), with the 12 verbatim test functions named in the US001 AC, following the architecture §4.2 sqlmock pattern. Tests-only — `task_repo.go` is byte-for-byte unchanged.

## Scope
- **In:** Edit `services/agent-board/internal/repo/task_repo_test.go` to add the 12 functions enumerated in §4.6 (and in US001 AC). Use the sqlmock branch→shape mapping from architecture §4.2 verbatim. For `_AuditInsertError` / `_CommitError` declare `ExpectRollback()` per the §4.2 note on `task_repo.go:96-102`.
- **Out:** Any change to `task_repo.go` itself (production code untouched). Any change to `user_story_repo*`, `audit_repo*`, `document_repo*`, `project_repo*`. Any new shared helper unless used by ≥3 of the new tests in this file (no premature abstraction). Doc-comment-vs-code mismatches — if surfaced, raise as `ARCHITECTURE_GAP_FOUND`, do not silently patch.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/repo/task_repo_test.go` (edit — add 12 test functions)

## Test contract
Dev makes these spec IDs pass (from `US001_be_unit_tests.md` once tester authors it). The 12 verbatim function names from US001 AC are the authoritative list:
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

Tester's `US001_be_unit_tests.md` IDs (UT-* / IT-*) map 1:1 onto these names. If tester adds more cases (e.g. a CreateTask scan-error variant), the dev writes them too and the additional names appear in the spec.

## Implementation notes
- **Reference pattern:** architecture §4.1 — re-read `services/agent-board/internal/repo/project_repo_test.go` and `document_repo_test.go` from REQ005/US005 for the canonical shape.
- **Branch → sqlmock idiom** (architecture §4.2, copy-pasteable):
  - `_GenericError` → `WillReturnError(errors.New("db down"))` on the relevant `ExpectQuery`.
  - `_NotFound` → `WillReturnError(sql.ErrNoRows)`; assert `errors.Is(err, repo.ErrNotFound)`.
  - `_ScanError` → `WillReturnRows(sqlmock.NewRows(cols).AddRow(<wrong type>))`.
  - `_RowsErr` → `WillReturnRows(sqlmock.NewRows(cols).AddRow(...).RowError(0, errors.New("rows err")))`.
  - `_BeginTxError` → `ExpectBegin().WillReturnError(errors.New("begin fail"))`.
  - `_AuditInsertError` → `ExpectBegin()` + happy query + `ExpectExec(<audit>).WillReturnError(...)` + `ExpectRollback()`.
  - `_CommitError` → `ExpectBegin()` + happy query + happy exec + `ExpectCommit().WillReturnError(...)`.
  - `_UpdateGenericError` (transactional) → `ExpectBegin()` + `ExpectQuery(<update>).WillReturnError(errors.New("update fail"))` + `ExpectRollback()`.
- **Assertions per branch:**
  - All branches: `assert.Error(t, err)` AND `assert.NoError(t, mock.ExpectationsWereMet())`.
  - `_NotFound` branches: `assert.ErrorIs(t, err, repo.ErrNotFound)`.
  - Non-`_NotFound` branches: `assert.False(t, errors.Is(err, repo.ErrNotFound))`.
  - `_BeginTxError` / `_AuditInsertError` / `_CommitError`: `assert.Contains(t, err.Error(), "<wrap prefix>")` — read the source for the exact `fmt.Errorf("failed to ...: %w", err)` strings before locking the assertion.
- **`UpdateTaskStatus` is the heaviest target** — five distinct error exits inside a transaction. Tester may consolidate via `t.Run` sub-tests within a parent `TestTaskRepo_UpdateTaskStatus_Errors` per US001 AC note; either shape is acceptable as long as every branch above is asserted.
- **The rollback `log.Printf` path (`task_repo.go:99`) is unreachable via sqlmock** — leave uncovered and name it in the test report under OQ-4 / architecture §4.5.
- **Coverage check command** (architecture §4.6, US001 row):
  ```
  cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo.out -run TestTaskRepo
  go tool cover -func=/tmp/repo.out | grep task_repo.go
  ```
  Must show ≥95% statement coverage on `task_repo.go`.

## Definition of done
- All 12 new test functions present with exact names; all green via `cd services/agent-board && go test ./internal/repo -cover -v -run TestTaskRepo`.
- `cd services/agent-board && go vet ./... && go test ./...` clean across the whole module (no regression elsewhere).
- `go tool cover -func=/tmp/repo.out | grep task_repo.go` shows **≥95%** statement coverage on `task_repo.go` (modulo `task_repo.go:99` rollback-log line per §4.5 — name it in the dev's task `## Notes` for the eventual test report).
- `task_repo.go` is byte-for-byte unchanged (`git diff services/agent-board/internal/repo/task_repo.go` is empty).
- `golangci-lint run ./...` clean inside `services/agent-board/`.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0 with `REVIEW GATE: PASS`; `scripts/review/run-gate.sh cross` exits 0 with `REVIEW GATE: PASS`.
- **Live e2e + 3-clean-run flake check NOT required for this story** — it is tests-only, production code unchanged (architecture §10.4). Equivalent assertion: `cd services/agent-board && go test -count=3 ./internal/repo -race` clean three runs.
- Dev set status to `in_review` and reported back; tech-lead approved (`completed`).

## Notes

### Files touched
- `services/agent-board/internal/repo/task_repo_test.go` — added 13 test functions (12 required verbatim + UT-003 `TestTaskRepo_GetTask_NotFound` per exhaustiveness mandate)
- `docs/requirements/REQ006_tech_debt_backfill_sprint/US001_be_task_repo_error_tests.md` — status flip to in_review, notes added
- `docs/requirements/REQ006_tech_debt_backfill_sprint/US001_be_unit_tests.md` — copied into worktree (was untracked in main workspace)
- `docs/requirements/REQ006_tech_debt_backfill_sprint/architecture.md` — copied into worktree (was untracked in main workspace)

### Tests added
13 new test functions covering all required error branches in `task_repo.go`:
- `TestTaskRepo_CreateTask_GenericError` (UT-001)
- `TestTaskRepo_GetTask_GenericError` (UT-002)
- `TestTaskRepo_GetTask_NotFound` (UT-003, exhaustiveness addition)
- `TestTaskRepo_UpdateTask_NotFound` (UT-004)
- `TestTaskRepo_UpdateTask_GenericError` (UT-005)
- `TestTaskRepo_UpdateTaskStatus_BeginTxError` (UT-006)
- `TestTaskRepo_UpdateTaskStatus_NotFound` (UT-007)
- `TestTaskRepo_UpdateTaskStatus_UpdateGenericError` (UT-008)
- `TestTaskRepo_UpdateTaskStatus_AuditInsertError` (UT-009)
- `TestTaskRepo_UpdateTaskStatus_CommitError` (UT-010)
- `TestTaskRepo_ListTasks_QueryError` (UT-011)
- `TestTaskRepo_ListTasks_ScanError` (UT-012)
- `TestTaskRepo_ListTasks_RowsErr` (UT-013)

### Test results
- `go test ./internal/repo -run TestTaskRepo`: 20 passed (7 pre-existing + 13 new)
- `go test ./... `: 146 passed in 6 packages
- `go vet ./...`: no issues
- `golangci-lint run ./...`: no issues
- `go test -count=3 ./internal/repo -race`: 171 passed (57 × 3 runs, no races)
- BE review gate: `REVIEW GATE: PASS`
- Cross gate: pre-existing semgrep Dockerfile failures unrelated to this task (confirmed on base branch too)

### Coverage (task_repo.go)
- `NewTaskRepo`: 100%
- `CreateTask`: 100%
- `GetTask`: 100%
- `UpdateTask`: 100%
- `UpdateTaskStatus`: 95.2% — line 99 (`log.Printf` in deferred rollback error handler) not covered per architecture §4.5 exemption (OQ-4). This line fires only when `tx.Rollback()` itself returns a non-`sql.ErrTxDone` error, which sqlmock does not produce in this scenario.
- `DeleteTask`: 100%
- `ListTasks`: 100%

### Production code change
`task_repo.go`: byte-for-byte unchanged (`git diff services/agent-board/internal/repo/task_repo.go` is empty).

### E2E
Not required per task DoD (tests-only story, architecture §10.4). Equivalent 3-clean-run race check passed: 171 tests, no races.

## Review log
(tech-lead appends here on each review pass)
