# US003/be_audit_repo_error_tests

**Requirement:** REQ006
**Story:** US003
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** none
**Worked-by:** be-dev-2026-06-05T00-00-00Z-a5c2
**Implements:** REQ006/US003 AC (all scenarios — 6 verbatim test function names covering 3 error branches × 2 public callers of `getAuditTrail`, ≥95% per-file coverage, no production-code change). Architecture §3 US003 touch row + §4.2 cluster-1 sqlmock pattern + §4.6 local verification command (US003 row).

## Goal
Backfill `audit_repo.go` error-branch tests so per-file statement coverage clears ≥95%, with the 6 verbatim test functions named in the US003 AC (3 error branches × 2 callers — typically `GetTaskAuditTrail` and `GetUserStoryAuditTrail`), following architecture §4.2. Tests-only.

## Scope
- **In:** Edit `services/agent-board/internal/repo/audit_repo_test.go` to add 6 test functions per US003 AC. Use sqlmock `_QueryError` / `_ScanError` / `_RowsErr` shapes from architecture §4.2 against the shared `getAuditTrail` plumbing.
- **Out:** Any change to `audit_repo.go`. Any change to `task_repo*`, `user_story_repo*`, etc.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/repo/audit_repo_test.go` (edit — add 6 test functions)

## Test contract
Dev makes the 6 verbatim test-function names from US003 AC pass (typically the cross-product of {`_QueryError`, `_ScanError`, `_RowsErr`} × {`GetTaskAuditTrail`, `GetUserStoryAuditTrail`}). Tester's UT-* IDs in `US003_be_unit_tests.md` map 1:1 onto these names.

## Implementation notes
- **Reference pattern:** architecture §4.1 / §4.2 — same shape as US001 / US002, but the SUT is the `ListXxx` family because `getAuditTrail` is a query-then-scan-loop method.
- **Branch idioms used:**
  - `_QueryError` → `mock.ExpectQuery(...).WillReturnError(errors.New("db down"))`.
  - `_ScanError` → `mock.ExpectQuery(...).WillReturnRows(sqlmock.NewRows(cols).AddRow(<wrong type>))`.
  - `_RowsErr` → `mock.ExpectQuery(...).WillReturnRows(sqlmock.NewRows(cols).AddRow(<valid>).RowError(0, errors.New("rows err")))`.
- **Read the source first** for the column-name list passed into `sqlmock.NewRows(cols)` and any wrap-prefix strings — they must match production literally.
- **Coverage check command** (architecture §4.6, US003 row):
  ```
  cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo.out -run TestAuditRepo
  go tool cover -func=/tmp/repo.out | grep audit_repo.go
  ```
  Must show ≥95% statement coverage on `audit_repo.go`.

## Definition of done
- All 6 new test functions present with US003 AC's verbatim names; all green via `cd services/agent-board && go test ./internal/repo -cover -v -run TestAuditRepo`.
- `cd services/agent-board && go vet ./... && go test ./...` clean across the whole module.
- `audit_repo.go` ≥95% statement coverage (modulo any §4.5 exemptions named in the test report).
- `audit_repo.go` byte-for-byte unchanged.
- `golangci-lint run ./...` clean.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` + `scripts/review/run-gate.sh cross` both `REVIEW GATE: PASS`.
- **Live e2e NOT required** (tests-only); instead 3 clean runs of `cd services/agent-board && go test -count=3 ./internal/repo -race`.
- Dev set status to `in_review`; tech-lead approved.

## Notes

### Files touched
- `services/agent-board/internal/repo/audit_repo_test.go` — added 6 test functions (UT-001..UT-006)

### Tests added
- `TestAuditRepo_GetTaskAuditTrail_QueryError` (UT-001)
- `TestAuditRepo_GetTaskAuditTrail_ScanError` (UT-002)
- `TestAuditRepo_GetTaskAuditTrail_RowsErr` (UT-003)
- `TestAuditRepo_GetUserStoryAuditTrail_QueryError` (UT-004)
- `TestAuditRepo_GetUserStoryAuditTrail_ScanError` (UT-005)
- `TestAuditRepo_GetUserStoryAuditTrail_RowsErr` (UT-006)

### Coverage result
`audit_repo.go` lifted from 78.6% (getAuditTrail) to 100% statement coverage.
All functions: `NewAuditRepo` 100%, `getAuditTrail` 100%, `GetTaskAuditTrail` 100%, `GetUserStoryAuditTrail` 100%.

### Live e2e
Not required per task DoD (tests-only story). Substituted with 3 clean race-detector runs:
`go test -count=3 ./internal/repo -race` — 231 passed, 0 failed, 0 races.

### Review gate
`scripts/review/run-gate.sh be services/agent-board` — REVIEW GATE: PASS
`scripts/review/run-gate.sh cross` — REVIEW GATE: PASS

### Spec note (ScanError column mismatch)
US003_be_unit_tests.md UT-002/UT-005 show 7-column `auditCols` including `changed_by`. The actual production `audit_repo.go` SELECT and `domain.StatusAuditLog` have only 6 columns (`id, entity_id, entity_type, from_status, to_status, changed_at` — no `changed_by`). Tests were written to match production. The `_ScanError` trigger uses `bool` in the `changed_at` (time.Time) column position — same pattern as `task_repo_test.go:485`. This is not an ARCHITECTURE_GAP_FOUND; the spec column list is descriptive context, not a contract.

### Production code
`audit_repo.go` is byte-for-byte unchanged (confirmed via `git diff`).

## Review log

### Review pass 1 — 2026-06-06 — verdict: approved

All 6 verbatim test functions present and passing (UT-001..UT-006).

**Tests:** `go test ./internal/repo/... -v -run TestAuditRepo` — 9 passed, 0 failed.
**Full module:** `go test ./...` — 210 passed, 0 failed across 6 packages.

**Coverage (`audit_repo.go` per-file, all functions):**
```
agent-board/internal/repo/audit_repo.go:25:  NewAuditRepo            100.0%
agent-board/internal/repo/audit_repo.go:31:  getAuditTrail           100.0%
agent-board/internal/repo/audit_repo.go:54:  GetTaskAuditTrail       100.0%
agent-board/internal/repo/audit_repo.go:60:  GetUserStoryAuditTrail  100.0%
```
All functions at 100% — well above the ≥95% threshold.

**Production code:** `git diff HEAD -- internal/repo/audit_repo.go` — no output (byte-for-byte unchanged, confirmed).

**Cross gate:** `scripts/review/run-gate.sh cross` — REVIEW GATE: PASS (semgrep PASS, gitleaks PASS).

**Error branch exhaustiveness:** 3 `return err` sites in `getAuditTrail` (`QueryContext` line 34, `rows.Scan` line 42, `rows.Err()` line 47) × 2 public callers = 6 branches. Spec provides 6 UT-* cases (UT-001..UT-006) — fully covered.

**TDG discipline:** commits on branch follow red→green→refactor cycles with `(US003)` traceability tags. The `refactor: chore:` prefix variant is a known pre-existing pattern already filed in `docs/tech_debt.md` (2026-06-05 entry).

**Spec note:** dev correctly noted the 7-column vs 6-column discrepancy between the spec's `auditCols` and the actual production SELECT. The 6-column production schema (`id, entity_id, entity_type, from_status, to_status, changed_at`) matches `audit_repo.go:29` exactly. Dev implemented ScanError using `bool` in the `changed_at` position — same pattern as `task_repo_test.go:485`. This is not a spec gap; the spec's column list was descriptive context.

**e2e:** Not required per DoD (tests-only story). Dev reported 3 clean race-detector runs (`go test -count=3 ./internal/repo -race` — 231 passed, 0 failed, 0 races).

**Tech-debt:** none filed this pass.
