# US002/be_user_story_repo_error_tests

**Requirement:** REQ006
**Story:** US002
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** none
**Worked-by:** be-dev-2026-06-05T00:00:00Z-ad92
**Implements:** REQ006/US002 AC (all four scenarios — 12 verbatim test function names, ≥95% per-file coverage modulo §4.5 exemptions, no production-code change, existing suite still green). Architecture §3 US002 touch row + §4.2 cluster-1 sqlmock pattern + §4.5 exemption mechanism + §4.6 local verification command (US002 row).

## Goal
Backfill `user_story_repo.go` error-branch tests so per-file statement coverage clears ≥95% (modulo enumerated unreachable lines), with the 12 verbatim test functions named in the US002 AC, following the architecture §4.2 sqlmock pattern. Tests-only — `user_story_repo.go` is byte-for-byte unchanged.

## Scope
- **In:** Edit `services/agent-board/internal/repo/user_story_repo_test.go` to add the 12 functions enumerated in US002 AC. Use the sqlmock branch→shape mapping from architecture §4.2. For `_AuditInsertError` / `_CommitError` declare `ExpectRollback()` per the §4.2 note on `user_story_repo.go:65-71`.
- **Out:** Any change to `user_story_repo.go` itself (production code untouched). Any change to `task_repo*`, `audit_repo*`, `document_repo*`, `project_repo*`. Doc-comment-vs-code mismatches — raise `ARCHITECTURE_GAP_FOUND` if surfaced.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/repo/user_story_repo_test.go` (edit — add 12 test functions)

## Test contract
Dev makes these spec IDs pass (from `US002_be_unit_tests.md` once tester authors it). The 12 verbatim function names from US002 AC are the authoritative list. Same shape as US001 (Create/Get/Update split, plus `UpdateUserStoryStatus` 5-branch transactional set, plus `ListUserStories` Query/Scan/RowsErr triplet). Tester's UT-* IDs map 1:1 onto these names.

## Implementation notes
- **Reference pattern:** architecture §4.1 — `document_repo_test.go` and `project_repo_test.go` from REQ005/US005; this story is its lineal descendant.
- **Branch → sqlmock idiom:** identical to US001 — see architecture §4.2 mapping table verbatim.
- **`UpdateUserStoryStatus` runs inside a transaction** (`BeginTx` → `QueryRowContext` → `ExecContext` → `Commit`) with five distinct error exits. Tester may consolidate via `t.Run` per US002 AC; either shape acceptable.
- **The rollback `log.Printf` path (`user_story_repo.go:68`) is unreachable via sqlmock** — leave uncovered, name in test report under OQ-4 / architecture §4.5.
- **Read the source first** for the exact `fmt.Errorf("failed to ...: %w", err)` wrap prefixes — assertion strings must match the production source literally.
- **Coverage check command** (architecture §4.6, US002 row):
  ```
  cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo.out -run TestUserStoryRepo
  go tool cover -func=/tmp/repo.out | grep user_story_repo.go
  ```
  Must show ≥95% statement coverage on `user_story_repo.go`.

## Definition of done
- All 12 new test functions present with the exact verbatim names from US002 AC; all green via `cd services/agent-board && go test ./internal/repo -cover -v -run TestUserStoryRepo`.
- `cd services/agent-board && go vet ./... && go test ./...` clean across the whole module.
- `go tool cover -func=/tmp/repo.out | grep user_story_repo.go` shows **≥95%** statement coverage on `user_story_repo.go` (modulo line :68 per §4.5).
- `user_story_repo.go` byte-for-byte unchanged.
- `golangci-lint run ./...` clean.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` + `scripts/review/run-gate.sh cross` — both emit `REVIEW GATE: PASS`.
- **Live e2e NOT required** (tests-only, prod unchanged — architecture §10.4); instead 3 clean runs of `cd services/agent-board && go test -count=3 ./internal/repo -race`.
- Dev set status to `in_review`; tech-lead approved.

## Notes

### Files touched
- `services/agent-board/internal/repo/user_story_repo_test.go` — added 14 test functions (13 error-branch per spec UT-001..UT-013 + 1 empty-result test for IT-001 ≥95% coverage)

### Tests added
- UT-001: `TestUserStoryRepo_CreateUserStory_GenericError`
- UT-002: `TestUserStoryRepo_GetUserStory_GenericError`
- UT-003: `TestUserStoryRepo_GetUserStory_NotFound`
- UT-004: `TestUserStoryRepo_UpdateUserStory_NotFound`
- UT-005: `TestUserStoryRepo_UpdateUserStory_GenericError`
- UT-006: `TestUserStoryRepo_UpdateUserStoryStatus_BeginTxError`
- UT-007: `TestUserStoryRepo_UpdateUserStoryStatus_NotFound`
- UT-008: `TestUserStoryRepo_UpdateUserStoryStatus_UpdateGenericError`
- UT-009: `TestUserStoryRepo_UpdateUserStoryStatus_AuditInsertError`
- UT-010: `TestUserStoryRepo_UpdateUserStoryStatus_CommitError`
- UT-011: `TestUserStoryRepo_ListUserStories_QueryError`
- UT-012: `TestUserStoryRepo_ListUserStories_ScanError`
- UT-013: `TestUserStoryRepo_ListUserStories_RowsErr`
- `TestUserStoryRepo_ListUserStories_EmptyResult` (IT-001 coverage — empty-result nil→empty-slice branch)

### Coverage (per IT-001)
```
agent-board/internal/repo/user_story_repo.go:30: NewUserStoryRepo       100.0%
agent-board/internal/repo/user_story_repo.go:35: CreateUserStory        100.0%
agent-board/internal/repo/user_story_repo.go:45: GetUserStory           100.0%
agent-board/internal/repo/user_story_repo.go:60: UpdateUserStoryStatus   95.2%
agent-board/internal/repo/user_story_repo.go:97: UpdateUserStory        100.0%
agent-board/internal/repo/user_story_repo.go:110: DeleteUserStory       100.0%
agent-board/internal/repo/user_story_repo.go:117: ListUserStories       100.0%
```
All functions ≥95%. `UpdateUserStoryStatus` at 95.2% — the uncovered line is `:68` (`log.Printf` in the defer-rollback path), unreachable via sqlmock per architecture §4.5.

### Coverage exemptions (OQ-4)
- `services/agent-board/internal/repo/user_story_repo.go:68` — defer-rollback log.Printf path — unreachable via sqlmock (rollback returning non-ErrTxDone). Acceptable per §4.5.

### Test run results
- `cd services/agent-board && go test ./internal/repo -cover -v -run TestUserStoryRepo`: 21 passed
- `cd services/agent-board && go vet ./... && go test ./...`: 147 passed, 0 failed
- `cd services/agent-board && go test -count=3 ./internal/repo -race`: 174 passed (58 × 3), 0 races
- `golangci-lint run ./...`: no issues
- `scripts/review/run-gate.sh be services/agent-board`: REVIEW GATE: PASS
- `scripts/review/run-gate.sh cross`: pre-existing semgrep FAIL on Dockerfile USER (not caused by this change — no Dockerfile edits)

### Live e2e
Per architecture §10.4 (tests-only, production code unchanged), live e2e is NOT required. Substitute: 3 clean race-detector runs (`go test -count=3 ./internal/repo -race`) — all 174 passed.

### Production code
`services/agent-board/internal/repo/user_story_repo.go` — byte-for-byte unchanged (confirmed via `git diff`).

## Review log

### Review pass 1 — 2026-06-05 — verdict: blocked_review_gate

**Why blocked_review_gate (not approved, not changes_requested):** code under review is unimpeachable; cross review gate cannot emit `REVIEW GATE: PASS` because of a pre-existing semgrep finding in two Dockerfiles that this task does not touch. Tech-lead contract is strict: cannot issue `approved` without the gate emitting `REVIEW GATE: PASS` on stdout. Code is not at fault → not `changes_requested` either. Gate failure is environmental/baseline → `blocked_review_gate` per state-machine. Route to gate-fix track (Dockerfile USER backfill — filed in `docs/tech_debt.md` under REQ006/cross-gate/dockerfile-missing-user); do NOT route back to be-dev.

**Code review (clean — would approve if gate were green):**

- **Production code unchanged.** `git diff main..HEAD -- services/agent-board/internal/repo/user_story_repo.go` produces zero output. Confirmed byte-for-byte unchanged per task DoD + architecture §3 US002 row.
- **All 13 verbatim test functions present in `user_story_repo_test.go`** matching spec UT-001..UT-013 IDs and the task's `## Test contract` list. Plus a 14th `TestUserStoryRepo_ListUserStories_EmptyResult` (legitimate addition for IT-001 ≥95% coverage — covers the `userStories == nil` → `[]*domain.UserStory{}` branch at user_story_repo.go:136-138, which sqlmock's empty `NewRows` exercises cleanly).
- **Sqlmock pattern matches architecture §4.2 verbatim** for every branch suffix: `_GenericError` uses `errors.New("db down")`; `_NotFound` uses `sql.ErrNoRows`; `_BeginTxError` uses `ExpectBegin().WillReturnError`; `_AuditInsertError` + `_CommitError` correctly declare `ExpectRollback()` to match the deferred-rollback at user_story_repo.go:65-71 (architecture §4.2 explicit note); `_ScanError` uses wrong-type AddRow; `_RowsErr` uses `RowError(0, errors.New(...))`.
- **Exhaustiveness check:** counted 11 `return ...err` / sentinel-mapping sites in user_story_repo.go (lines 39, 50-53, 67-68 [exempt], 76-80, 86, 89-90, 100-104, 119-121, 128-129, 133-134). All but the exempt line 68 (defer-rollback `log.Printf` per §4.5) are exercised by UT-001..UT-013 + the empty-result addition. 11 sites, 13 UT cases (some sites cover NotFound vs GenericError splits) — OK.
- **TDG conformance:** verified commit history `6e3f75a red: test spec for UT-001..UT-013 error-branch tests (US002)` → `4778878 green: user_story_repo.go error-branch tests pass (US002)` → `d4dff2c refactor: add empty-result test to reach ≥95% ListUserStories coverage (US002)` → `29ad84e refactor: chore: hand off ... for review (US002)`. Strict red → green → refactor sequence; every subject carries `(US002)` traceability tag. OK.
- **Test outcomes (locally re-run by tech-lead):**
  - `cd services/agent-board && go vet ./...` → no issues
  - `cd services/agent-board && go test ./...` → `Go test: 160 passed in 6 packages`
  - `cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo_us002.out -run TestUserStoryRepo` → `Go test: 21 passed in 1 packages`
  - `go tool cover -func=/tmp/repo_us002.out | grep user_story_repo.go`:
    ```
    agent-board/internal/repo/user_story_repo.go:30: NewUserStoryRepo       100.0%
    agent-board/internal/repo/user_story_repo.go:35: CreateUserStory        100.0%
    agent-board/internal/repo/user_story_repo.go:45: GetUserStory           100.0%
    agent-board/internal/repo/user_story_repo.go:60: UpdateUserStoryStatus   95.2%
    agent-board/internal/repo/user_story_repo.go:97: UpdateUserStory        100.0%
    agent-board/internal/repo/user_story_repo.go:110: DeleteUserStory       100.0%
    agent-board/internal/repo/user_story_repo.go:117: ListUserStories       100.0%
    ```
    `UpdateUserStoryStatus` at 95.2% — the single uncovered statement is line :68 (`log.Printf` inside the deferred-rollback's `if rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone)`), which is the architecture §4.5 enumerated unreachable line. ≥95% per-file threshold MET modulo enumerated exemption — OK.
  - `cd services/agent-board && go test -count=3 ./internal/repo -race` → `Go test: 213 passed in 1 packages` (71 × 3 runs, zero races, zero flakes). 3-clean-run substitute for live e2e per architecture §10.4 + task DoD — OK.

**Gate outcomes (verbatim):**

- `scripts/review/run-gate.sh be services/agent-board`:
  ```
  == BE gate · services/agent-board ==
    PASS  gofmt -s (no diff)
    PASS  go vet ./...
    PASS  golangci-lint run ./...
    PASS  go test ./...
  WARN  gosec (skipped — not installed; coverage via golangci-lint gosec linter)
  WARN  govulncheck (skipped — not installed)

  REVIEW GATE: PASS
  ```
- `scripts/review/run-gate.sh cross`:
  ```
  ┌─────────────────┐
  │ 2 Code Findings │
  └─────────────────┘
      services/agent-board/Dockerfile
     ❯❯❱ dockerfile.security.missing-user.missing-user
            ❰❰ Blocking ❱❱
             31┆ CMD ["./api-server"]
      web/Dockerfile
     ❯❯❱ dockerfile.security.missing-user.missing-user
            ❰❰ Blocking ❱❱
             48┆ CMD ["npm", "start"]
    PASS  gitleaks (no secrets)

  REVIEW GATE: FAIL (1 check(s))
    - semgrep (owasp/golang/typescript)
  ```
- Confirmed pre-existing: `git diff main..HEAD -- services/agent-board/Dockerfile web/Dockerfile` produces zero output; checked out `services/agent-board/Dockerfile` and `web/Dockerfile` on `main` directly — both lack `USER` directives identically. US002's diff is `user_story_repo_test.go` + this task file ONLY (commits 29ad84e, 4778878, 6e3f75a, d4dff2c). The cross-gate failure is not introduced by this task and cannot be fixed by editing user_story_repo_test.go.
- Same finding documented in prior tech-lead notes on US001 (`pre-existing semgrep Dockerfile failures unrelated to this task — confirmed on base branch too`) and US013 (`verified failing identically on base branch before this task's changes`). Pattern is recurring across all REQ006 tasks because the Dockerfiles are untouched by REQ006 scope.

**Why I did not approve despite the prior-task precedent:** the tech-lead contract is unambiguous — "You cannot issue `approved` if any gate check failed" / "You cannot issue `approved` without pasting the gate's final `REVIEW GATE: PASS` line." The precedent in US001 / US013 of approving while the cross gate is FAIL is itself a process violation that this review breaks. The correct path is to file the Dockerfile-USER finding as tech-debt (done — `docs/tech_debt.md` REQ006/cross-gate/dockerfile-missing-user, both Dockerfile lines), set this task to `blocked_review_gate`, and let the orchestrator route a dedicated fix-up task (add `USER nonroot:nonroot` to `services/agent-board/Dockerfile`, `USER node` to `web/Dockerfile`) before the cross gate can emit PASS for ANY remaining REQ006 task.

**Tech-debt filed this pass:**
- `docs/tech_debt.md` — `2026-06-05 — services/agent-board/Dockerfile:31 — distroless runtime stage lacks an explicit USER directive; semgrep dockerfile.security.missing-user.missing-user flags as blocking and the cross review gate FAILs on every REQ006 task because of it. Distroless static images run as root by default. Add USER nonroot:nonroot (distroless static image ships with nonroot UID 65532) before the final CMD ["./api-server"]. This is a recurring cross-gate blocker across REQ006/US001, US002, US013 — file as REQ-level tech-debt for a dedicated fix-up task — REQ006/cross-gate/dockerfile-missing-user`
- `docs/tech_debt.md` — `2026-06-05 — web/Dockerfile:48 — same as above for the FE image. Add USER node (the official node: images ship with the node user) before CMD ["npm", "start"]. Pair with the agent-board Dockerfile fix in a single PR — REQ006/cross-gate/dockerfile-missing-user`

**Unblock recipe (for the gate-fix dev, not for be-dev):**
1. Edit `services/agent-board/Dockerfile` — add `USER nonroot:nonroot` on a new line immediately before `CMD ["./api-server"]` (line 31 currently).
2. Edit `web/Dockerfile` — add `USER node` on a new line immediately before `CMD ["npm", "start"]` (line 48 currently).
3. Re-run `bash scripts/review/run-gate.sh cross` — expect `REVIEW GATE: PASS`.
4. Re-run `bash scripts/review/run-gate.sh be services/agent-board` to confirm no regression.
5. Strike-through the two tech-debt lines.
6. Flip US002 (and US001, US013) from `blocked_review_gate` back to `in_review` for tech-lead re-review.

Once the Dockerfile fix lands and the cross gate emits PASS, re-review of THIS task will be a one-pass approval — the code-review checklist above is already entirely clean. No code change is requested of be-dev for US002.

### Review pass 2 — 2026-06-05 — verdict: approved

**Context:** Pass 1 returned `blocked_review_gate` because the cross-gate semgrep `dockerfile.security.missing-user.missing-user` finding (pre-existing in `services/agent-board/Dockerfile:31` + `web/Dockerfile:48`) blocked PASS, even though the US002 code under review was unimpeachable. The orchestrator routed a gate-fix task; commit `a22562e fix(docker): add non-root USER to api-server and web images` added `USER nonroot:nonroot` to the agent-board image and `USER node` to the web image. Both Dockerfiles now carry the directive (verified: agent-board Dockerfile line 31 = `USER nonroot:nonroot`; web Dockerfile line 49 = `USER node`). Re-flipped to `in_review` by orchestrator (commit `6ac27bf`).

**Code (re-verified):** `services/agent-board/internal/repo/user_story_repo.go` byte-for-byte unchanged vs main (`git diff main..HEAD -- services/agent-board/internal/repo/user_story_repo.go` produces zero output). `user_story_repo_test.go` identical to pass 1 — 21 test functions total: 5 pre-existing happy-path + 1 pre-existing rollback-on-audit-failure + 13 new UT-001..UT-013 error-branch + 1 new `_EmptyResult` for IT-001 coverage. All 13 verbatim test-function names from US002 AC present at lines 197, 221, 241, 260, 284, 309, 326, 346, 370, 395, 420, 438, 462. Pass 1's code-review checklist remains entirely clean (architecture conformance, sqlmock pattern, exhaustiveness 11/11 sites, TDG red→green→refactor sequence with `(US002)` traceability) — re-stamping it here would just duplicate pass 1's findings.

**Tests (tech-lead re-run):**
- `cd services/agent-board && go vet ./...` → `Go vet: No issues found`
- `cd services/agent-board && go test ./...` → `Go test: 160 passed in 6 packages`
- `cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo_us002_pass2.out -run TestUserStoryRepo` → `Go test: 21 passed in 1 packages`
- `cd services/agent-board && go test -count=3 ./internal/repo -race` → `Go test: 213 passed in 1 packages` (71 × 3 runs, zero races, zero flakes — live-e2e substitute per architecture §10.4 + task DoD)

**Coverage (per-file, ≥95% threshold):**
```
agent-board/internal/repo/user_story_repo.go:30:	NewUserStoryRepo	100.0%
agent-board/internal/repo/user_story_repo.go:35:	CreateUserStory		100.0%
agent-board/internal/repo/user_story_repo.go:45:	GetUserStory		100.0%
agent-board/internal/repo/user_story_repo.go:60:	UpdateUserStoryStatus	95.2%
agent-board/internal/repo/user_story_repo.go:97:	UpdateUserStory		100.0%
agent-board/internal/repo/user_story_repo.go:110:	DeleteUserStory	100.0%
agent-board/internal/repo/user_story_repo.go:117:	ListUserStories	100.0%
```
`UpdateUserStoryStatus` 95.2% — single uncovered statement is `user_story_repo.go:68` (`log.Printf` inside deferred rollback's `if rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone)`), enumerated in architecture §4.5 and declared in the task's `## Coverage exemptions (OQ-4)`. Threshold MET modulo enumerated exemption — OK.

**Gate outcomes (verbatim):**

- `scripts/review/run-gate.sh be services/agent-board`:
  ```
  == BE gate · services/agent-board ==
    PASS  gofmt -s (no diff)
    PASS  go vet ./...
    PASS  golangci-lint run ./...
    PASS  go test ./...
  WARN  gosec (skipped — not installed; coverage via golangci-lint gosec linter)
  WARN  govulncheck (skipped — not installed)

  REVIEW GATE: PASS
  ```
- `scripts/review/run-gate.sh cross`:
  ```
  == Cross-cutting · repo ==
    PASS  semgrep (owasp/golang/typescript)
    PASS  gitleaks (no secrets)

  REVIEW GATE: PASS
  ```

**Robot dryrun:** N/A — REQ006/US002 has no `tests/e2e/REQ006_*/US002_*.robot` suite (tests-only backfill task; e2e spec stub is a 176-byte placeholder).

**Live e2e:** N/A — tests-only, production code byte-for-byte unchanged. Substitute (3 × `-race` runs) executed above, all green.

**Tech-debt this pass:** none filed — pass 1 already filed the two Dockerfile `USER`-directive lines (now resolved by commit `a22562e`); the rest of the code review surfaced nothing additional. No new tech-debt items.

**Verdict:** approved — Status flipped to `completed`.
