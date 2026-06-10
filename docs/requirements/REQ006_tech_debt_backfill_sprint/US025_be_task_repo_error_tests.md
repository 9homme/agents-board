# US025/be_task_repo_error_tests

**Requirement:** REQ006
**Story:** US025
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** none
**Worked-by:** be-dev-2026-06-05T00:00:00Z-aa86
**Implements:** REQ006/US025 AC (all four scenarios — 12 verbatim test function names, ≥95% per-file coverage modulo §4.5 exemptions, no production-code change, existing suite still green). Architecture §3 US025 touch row + §4.2 cluster-1 sqlmock pattern + §4.5 exemption mechanism + §4.6 local verification command (US025 row).

## Goal
Backfill `task_repo.go` error-branch tests so per-file statement coverage clears ≥95% (modulo enumerated unreachable lines), with the 12 verbatim test functions named in the US025 AC, following the architecture §4.2 sqlmock pattern. Tests-only — `task_repo.go` is byte-for-byte unchanged.

## Scope
- **In:** Edit `services/agent-board/internal/repo/task_repo_test.go` to add the 12 functions enumerated in §4.6 (and in US025 AC). Use the sqlmock branch→shape mapping from architecture §4.2 verbatim. For `_AuditInsertError` / `_CommitError` declare `ExpectRollback()` per the §4.2 note on `task_repo.go:96-102`.
- **Out:** Any change to `task_repo.go` itself (production code untouched). Any change to `user_story_repo*`, `audit_repo*`, `document_repo*`, `project_repo*`. Any new shared helper unless used by ≥3 of the new tests in this file (no premature abstraction). Doc-comment-vs-code mismatches — if surfaced, raise as `ARCHITECTURE_GAP_FOUND`, do not silently patch.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/repo/task_repo_test.go` (edit — add 12 test functions)

## Test contract
Dev makes these spec IDs pass (from `US025_be_unit_tests.md` once tester authors it). The 12 verbatim function names from US025 AC are the authoritative list:
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

Tester's `US025_be_unit_tests.md` IDs (UT-* / IT-*) map 1:1 onto these names. If tester adds more cases (e.g. a CreateTask scan-error variant), the dev writes them too and the additional names appear in the spec.

## Implementation notes
- **Reference pattern:** architecture §4.1 — re-read `services/agent-board/internal/repo/project_repo_test.go` and `document_repo_test.go` from REQ005/US029 for the canonical shape.
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
- **`UpdateTaskStatus` is the heaviest target** — five distinct error exits inside a transaction. Tester may consolidate via `t.Run` sub-tests within a parent `TestTaskRepo_UpdateTaskStatus_Errors` per US025 AC note; either shape is acceptable as long as every branch above is asserted.
- **The rollback `log.Printf` path (`task_repo.go:99`) is unreachable via sqlmock** — leave uncovered and name it in the test report under OQ-4 / architecture §4.5.
- **Coverage check command** (architecture §4.6, US025 row):
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
- `docs/requirements/REQ006_tech_debt_backfill_sprint/US025_be_task_repo_error_tests.md` — status flip to in_review, notes added
- `docs/requirements/REQ006_tech_debt_backfill_sprint/US025_be_unit_tests.md` — copied into worktree (was untracked in main workspace)
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

### Review pass 1 — 2026-06-05 — verdict: blocked_review_gate

**Code-quality review on the task itself is clean and would otherwise be `approved`.** The blocker is the cross-gate failure described below; once the gate-fix track clears the pre-existing Dockerfile findings, re-running the gate against this unchanged code should immediately flip to `approved` (no rework needed on US025).

- **BE gate:** `scripts/review/run-gate.sh be services/agent-board` → exit 0; final line verbatim — `REVIEW GATE: PASS`. (gofmt -s, go vet, golangci-lint, go test ./... all green; gosec/govulncheck soft-warn skipped per host config — same baseline as REQ005/US029.)
- **Cross gate:** `scripts/review/run-gate.sh cross` → exit 1; final line — `REVIEW GATE: FAIL (1 check(s))` — failing check is `semgrep (owasp/golang/typescript)` with TWO `dockerfile.security.missing-user.missing-user` findings (`services/agent-board/Dockerfile:31` + `web/Dockerfile:48`). These findings are PRE-EXISTING — `git log -- services/agent-board/Dockerfile web/Dockerfile` shows the last touch on either file is `7b836c5` (REQ005/US032 — May 2026), well before REQ006 was opened. US025's `Scope: In` is exactly one file (`services/agent-board/internal/repo/task_repo_test.go`); the Dockerfiles are outside US025's lane. The cross gate cannot emit PASS regardless of what US025 does.
- **Per-file coverage on `task_repo.go`** (architecture §4.6 + spec IT-001):
  ```
  agent-board/internal/repo/task_repo.go:31:		NewTaskRepo		100.0%
  agent-board/internal/repo/task_repo.go:36:		CreateTask		100.0%
  agent-board/internal/repo/task_repo.go:51:		GetTask			100.0%
  agent-board/internal/repo/task_repo.go:70:		UpdateTask		100.0%
  agent-board/internal/repo/task_repo.go:91:		UpdateTaskStatus	95.2%
  agent-board/internal/repo/task_repo.go:133:		DeleteTask		100.0%
  agent-board/internal/repo/task_repo.go:140:		ListTasks		100.0%
  ```
  Six of seven functions at 100%; `UpdateTaskStatus` at 95.2% (one uncovered line is `task_repo.go:99` — the `log.Printf` inside the `defer` rollback path, sqlmock-unreachable). Exempt per architecture §4.5 (OQ-4) and spec IT-001's `## Coverage exemptions`. Threshold ≥95% met.
- **Test contract conformance:** all 12 verbatim function names from `## Test contract` present in `task_repo_test.go` (lines 228, 253, 292, 317, 342, 360, 380, 401, 429, 457, 475, 495). Plus UT-003 `TestTaskRepo_GetTask_NotFound` per spec's exhaustiveness mandate. `go test ./internal/repo -run TestTaskRepo` → 20 passed (7 pre-existing + 13 new).
- **Sqlmock pattern conformance (architecture §4.2):** every branch shape audited against §4.2 — `_GenericError` uses `WillReturnError(errors.New("db down"))`; `_NotFound` uses `WillReturnError(sql.ErrNoRows)`; `_BeginTxError` uses `ExpectBegin().WillReturnError(...)`; `_AuditInsertError` and `_CommitError` declare `ExpectRollback()` per §4.2's note about `task_repo.go:96-102`; `_UpdateGenericError` declares `ExpectRollback()` (transactional path with the same defer); `_ScanError` uses `AddRow(...wrong-type...)` (bool in `created_at` time.Time column); `_RowsErr` uses `.RowError(0, errors.New("rows err"))`. All correct.
- **Wrap-prefix assertions vs source:** confirmed `task_repo.go:94` → `"failed to begin transaction"`, `task_repo.go:122` → `"failed to insert audit log"`, `task_repo.go:126` → `"failed to commit transaction"`. UT-006/009/010 `assert.Contains` strings match the source verbatim.
- **Production code byte-for-byte unchanged:** `git log fb4e882..HEAD -- services/agent-board/internal/repo/task_repo.go` returns empty. Confirmed.
- **Full suite regression:** `cd services/agent-board && go test ./...` → `Go test: 160 passed in 6 packages`. No regression elsewhere.
- **Race + 3-clean-run flake check:** `go test -count=3 ./internal/repo -race` → `Go test: 213 passed in 1 packages` (71 cases × 3 runs, no races). Equivalent to the live-e2e 3-clean-run flake check (story is tests-only per architecture §10.4; DoD line 71 names this substitution explicitly).
- **Test spec exhaustiveness audit (anti-REQ005/US029 check):** counted `task_repo.go` error branches vs UT-* cases — `CreateTask` 1 err site → UT-001; `GetTask` 2 (sql.ErrNoRows + generic) → UT-002 + UT-003; `UpdateTask` 2 (sql.ErrNoRows + generic) → UT-004 + UT-005; `UpdateTaskStatus` 5 (begin + sql.ErrNoRows + generic + audit-exec + commit) → UT-006 + UT-007 + UT-008 + UT-009 + UT-010; `DeleteTask` trivial passthrough (no err-branch test required); `ListTasks` 3 (query + scan + rows.Err) → UT-011 + UT-012 + UT-013. 13 branches → 13 UT cases. Exact 1:1 match. No SPEC_GAP_FOUND.
- **TDG conformance:** commit chain `fb4e882..a3fe1a8` → `red:` (fb4e882) → `green:` (873cf2b) → `refactor: chore:` (b1dcdc4) → `refactor: chore:` (a3fe1a8). Red-before-green ordering OK; every commit carries the `(US025)` traceability tag. The two `refactor: chore:` subjects are mildly off-shape (chore-like work is being labeled as refactor — should arguably use a dedicated `chore:` prefix if the TDG skill supports it). Not blocking — filed as tech-debt.
- **Dev hand-off `## Notes` evidence:** dev recorded `REVIEW GATE: PASS` for BE gate and explicitly disclosed pre-existing semgrep Dockerfile failures (lines 104–105) — honest hand-off, no rationalisation.

**Verdict rationale.** Per `.claude/agents/tech-lead.md` strict reading: "You cannot issue `approved` if any gate check failed... [use] `blocked_review_gate` when the gate, coverage tooling, or `robot --dryrun` could not run cleanly through to a clear PASS/FAIL — i.e. the gate or tooling is at fault, not the code." The cross-gate semgrep check IS working correctly and IS reporting two real (pre-existing) Dockerfile security findings — that is the gate behaving as intended. The code US025 introduces is NOT at fault (US025 touches only `task_repo_test.go`; Dockerfiles are outside `Scope: In`). This is the "gate cannot emit PASS for reasons unrelated to this task's code" branch: `blocked_review_gate` is the only valid verdict. `changes_requested` would be wrong (US025's code is not the defect); `approved` is forbidden (the gate did not emit `REVIEW GATE: PASS` on cross).

**Orchestrator routing.** Do NOT re-route to be-dev as a US025 rework. The unblock is one of:

1. **Preferred:** spawn a dedicated gate-fix track task to add `USER non-root` (or equivalent) to both Dockerfiles — one-line patch per file. After merge, re-run the cross gate against this unchanged code and the verdict flips to `approved`.
2. **Alternative:** if the orchestrator + human decide Dockerfile hardening is out of scope for REQ006, add a documented carve-out to `scripts/review/run-gate.sh` (semgrep `--exclude=Dockerfile` or similar) with a comment citing the deferral REQ. Re-run cross gate, approve US025.

Tech-debt findings already filed to `docs/tech_debt.md` so they aren't lost regardless of which path the orchestrator takes.

**Tech-debt: filed this pass.** See `docs/tech_debt.md` 2026-06-05 entries (two Dockerfile findings + one TDG prefix-shape observation).

### Review pass 2 — 2026-06-05 — verdict: approved

Re-review after the orchestrator's gate-fix patch landed `USER nonroot:nonroot` on `services/agent-board/Dockerfile:31` and `USER node` on `web/Dockerfile:49`. The code under review (`task_repo.go` + `task_repo_test.go`) is byte-for-byte unchanged from pass 1 — re-confirmed `git diff main -- services/agent-board/internal/repo/task_repo.go` is empty. All pass-1 findings (which were "would-be approved modulo the cross-gate blocker") carry forward unchanged. Only the gate-blocker is re-verified here.

- **BE gate:** `scripts/review/run-gate.sh be services/agent-board` → exit 0; final line verbatim — `REVIEW GATE: PASS`. (gofmt -s, go vet, golangci-lint, go test ./... all green. gosec / govulncheck soft-skip per host config, identical to pass 1 and to REQ005/US029 baseline — not a regression.)
- **Cross gate:** `scripts/review/run-gate.sh cross` → exit 0; final line verbatim — `REVIEW GATE: PASS`. Semgrep (owasp/golang/typescript) and gitleaks both PASS. The pre-existing `dockerfile.security.missing-user.missing-user` findings that blocked pass 1 are now resolved by the orchestrator's gate-fix patch (confirmed `USER nonroot:nonroot` at `services/agent-board/Dockerfile:31` and `USER node` at `web/Dockerfile:49`).
- **Per-file coverage on `task_repo.go`** (architecture §4.6, US025 row):
  ```
  agent-board/internal/repo/task_repo.go:31:		NewTaskRepo		100.0%
  agent-board/internal/repo/task_repo.go:36:		CreateTask		100.0%
  agent-board/internal/repo/task_repo.go:51:		GetTask			100.0%
  agent-board/internal/repo/task_repo.go:70:		UpdateTask		100.0%
  agent-board/internal/repo/task_repo.go:91:		UpdateTaskStatus	95.2%
  agent-board/internal/repo/task_repo.go:133:		DeleteTask		100.0%
  agent-board/internal/repo/task_repo.go:140:		ListTasks		100.0%
  ```
  Threshold ≥95% met across the board. The one uncovered statement is `task_repo.go:99` (`log.Printf` in the deferred rollback error handler) — exempt per architecture §4.5 / OQ-4 and spec IT-001's `## Coverage exemptions` (sqlmock cannot make `tx.Rollback()` itself return a non-`sql.ErrTxDone` error).
- **Race + 3-clean-run flake check (DoD line 71 substitution):** `cd services/agent-board && go test -count=3 ./internal/repo -race` → `Go test: 213 passed in 1 packages` (71 cases × 3 runs, zero failures, zero races). Equivalent to the live-e2e 3-clean-run flake check; the story is tests-only per architecture §10.4 so live e2e is not required.
- **Full-module regression check:** `cd services/agent-board && go test ./...` → `Go test: 160 passed in 6 packages`. No regression elsewhere.
- **Production code byte-for-byte unchanged:** `git log main..HEAD -- services/agent-board/internal/repo/task_repo.go` empty; `git diff main -- services/agent-board/internal/repo/task_repo.go` empty. Confirmed.
- **All pass-1 conformance checks carry forward unchanged:** 12 verbatim function names present (+ exhaustiveness UT-003), sqlmock pattern §4.2 conformance, wrap-prefix assertions vs source, test-spec exhaustiveness (13 branches → 13 UT cases, 1:1 match, no SPEC_GAP_FOUND), TDG red-before-green ordering with `(US025)` traceability tags.

**Verdict rationale.** Both gates emit `REVIEW GATE: PASS`. Coverage threshold met. Three clean race runs. Code review on the test code itself was already clean in pass 1 (`would otherwise be approved`). The blocker was the cross-gate Dockerfile finding outside US025's scope; the orchestrator's gate-fix track has cleared it. There is no remaining reason to withhold approval. **approved.**

**Tech-debt: none filed this pass.** All findings from pass 1 are already on the ledger (two Dockerfile findings — now resolved by the gate-fix track and can be struck through on next sweep; one TDG prefix-shape observation — remains). No new findings surfaced by the unchanged code in this pass.

