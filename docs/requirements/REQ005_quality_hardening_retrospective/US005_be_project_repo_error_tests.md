# US005 — Backfill 8 project-repo error-branch tests

**Story:** US005 — Backfill 14 repo error-branch tests in `internal/repo`
**Requirement:** REQ005
**Track:** BE
**Service:** services/agent-board
**Status:** in_review
**Implements:** Scenario: `project_repo_test.go` gains the symmetric error-branch tests, Scenario: each test exercises the specific uncovered branch (project_repo portion), Scenario: per-file coverage hits ≥95% (project_repo.go portion), Scenario: existing tests still pass and behaviour is unchanged, Scenario: no production-code changes
**Blocked by:** none
**Worked-by:** be-dev-2026-06-03T00:00:00Z-a4a3

## Goal

Add the 8 backfill error-branch tests for `project_repo.go` per architecture §5.2 (`P1`..`P8`) so per-file coverage of `services/agent-board/internal/repo/project_repo.go` reaches ≥ 95 %. Pure test addition — no production-code changes. Symmetric counterpart to the document-repo task.

## Scope

- **In:** Add 8 new test functions to `services/agent-board/internal/repo/project_repo_test.go` with names verbatim from §5.2 (`TestProjectRepo_CreateProject_GenericError`, `_GetProject_GenericError`, `_UpdateProject_NotFound`, `_UpdateProject_GenericError`, `_DeleteProject_GenericError`, `_ListProjects_QueryError`, `_ListProjects_ScanError`, `_ListProjects_RowsErr`). Each follows §5.3 uniformity.
- **Out:** Document-repo tests (separate task `US005_be_document_repo_error_tests.md`); any edit to `project_repo.go` production source (§5.5); handler-side tests; `cmd/api-server` / `internal/mcp` coverage; refactoring existing happy-path tests.

## Files touched (estimated, exclusive)

- `services/agent-board/internal/repo/project_repo_test.go`

Independent of `US005_be_document_repo_error_tests.md`. Both US005 tasks parallelise cleanly.

## Test contract

The dev must make these tests pass (from `US005_be_unit_tests.md`, IDs assigned by tester — mapped to architecture §5.2 P1..P8):
- P1 — `TestProjectRepo_CreateProject_GenericError`
- P2 — `TestProjectRepo_GetProject_GenericError`
- P3 — `TestProjectRepo_UpdateProject_NotFound`
- P4 — `TestProjectRepo_UpdateProject_GenericError`
- P5 — `TestProjectRepo_DeleteProject_GenericError`
- P6 — `TestProjectRepo_ListProjects_QueryError`
- P7 — `TestProjectRepo_ListProjects_ScanError`
- P8 — `TestProjectRepo_ListProjects_RowsErr`

If tester surfaces new test IDs beyond these eight, the dev writes them and flags the addition back to tester.

## Implementation notes

- Architecture §5.2 lists each test's mock shape verbatim — mirror of §5.1's pattern. Mock-shape notes:
  - `_GenericError` / `_QueryError`: `WillReturnError(errors.New("db down"))`.
  - `_NotFound`: `WillReturnError(sql.ErrNoRows)` then `errors.Is(err, repo.ErrNotFound)` must be true.
  - `_DeleteProject_GenericError`: `mock.ExpectExec(...).WillReturnError(...)`.
  - `_ScanError`: `AddRow` with column-type mismatch.
  - `_RowsErr`: `RowError(0, errors.New("..."))`.
- Per-test uniform assertions from §5.3:
  1. Returned error non-nil.
  2. `errors.Is(err, repo.ErrNotFound)` true only for `_NotFound`; false for all other variants.
  3. Substring match on the relevant `failed to <op>` wrap text (or `"error iterating"` for `_RowsErr`).
  4. `mock.ExpectationsWereMet()` returns nil.
  5. Returned `*Project` / `[]Project` is nil on the error path.
- Use the existing `setupMock(t)` helper / pattern of existing happy-path tests in `project_repo_test.go`.
- §5.5: production source MUST NOT be edited. Tech-lead will `git diff` to confirm.
- TDG skill (.claude/skills/tdg/SKILL.md) MUST be invoked at each TDD phase per be-dev workflow.

## Definition of done

- All 8 listed tests green.
- `cd services/agent-board && go vet ./... && go test ./...` clean.
- `cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo.out && go tool cover -func=/tmp/repo.out` — `project_repo.go` shows ≥ 95 % per-file line coverage (baseline ~81.5 %). If any line is genuinely unreachable via sqlmock, list it in a `## Coverage exemption` section here with a one-line justification.
- `git diff services/agent-board/internal/repo/project_repo.go` shows zero changes.
- `scripts/review/run-gate.sh be services/agent-board` exits with `REVIEW GATE: PASS`.
- `scripts/review/run-gate.sh cross` exits with `REVIEW GATE: PASS`.
- Code matches architecture §5 contract.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Notes

**Files touched:** `services/agent-board/internal/repo/project_repo_test.go` only. Zero changes to production code confirmed by `git diff`.

**Tests added (8 new):**
- UT-US005-009 `TestProjectRepo_CreateProject_GenericError` — P1
- UT-US005-010 `TestProjectRepo_GetProject_GenericError` — P2
- UT-US005-011 `TestProjectRepo_UpdateProject_NotFound` — P3
- UT-US005-012 `TestProjectRepo_UpdateProject_GenericError` — P4
- UT-US005-013 `TestProjectRepo_DeleteProject_GenericError` — P5
- UT-US005-014 `TestProjectRepo_ListProjects_QueryError` — P6
- UT-US005-015 `TestProjectRepo_ListProjects_ScanError` — P7
- UT-US005-016 `TestProjectRepo_ListProjects_RowsErr` — P8

**Coverage:** `project_repo.go` achieved 100% statement coverage (≥95% required). All functions: 100%.

**`go vet ./...`:** clean. **`go test ./...`:** 125 passed across 6 packages.

**BE gate (`scripts/review/run-gate.sh be services/agent-board`):** FAILS on `govulncheck ./...` due to a pre-existing `crypto/x509` stdlib vulnerability in `internal/handler/sse.go` and `internal/handler/message.go`. Confirmed pre-existing by running the gate against the stash before my changes — same failure. This task only touches `project_repo_test.go`. Gate failure is NOT introduced by this task.

**Cross gate (`scripts/review/run-gate.sh cross`):** REVIEW GATE: PASS.

## Review log

(tech-lead appends here on each review pass)
