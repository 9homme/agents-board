# US011/be_errcheck_rollback_discard

**Requirement:** REQ003
**Story:** US011
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** US011_be_mechanical_noctx_errorlint.md
**Worked-by:** be-dev (US011 errcheck)
**Implements:** US011 acceptance criterion "Specific finding categories are resolved correctly, not papered over" — drives `errcheck` (4) to zero, specifically the ignored `tx.Rollback()` returns in `user_story_repo` / `task_repo`.

## Goal
Make every ignored `tx.Rollback()` return in the repo layer explicit: assign to `_` with a one-line justification comment that documents why the rollback error is being deliberately discarded (the conventional "rollback-after-error: the original error is what the caller cares about" pattern). Where `errcheck` flags a non-rollback site, evaluate per case — discard with justification, or actually handle the error, whichever is correct.

## Scope
- **In:**
  - All 4 `errcheck` findings, expected concentration: ignored `tx.Rollback()` returns inside `services/agent-board/internal/repo/user_story_repo.go` and `services/agent-board/internal/repo/task_repo.go`. Confirm exact sites against the live linter run; if the linter flags additional repo files, fix them too.
  - The accepted pattern for rollback-after-error is an explicit discard plus a justification comment, e.g.:
    ```go
    if err != nil {
        _ = tx.Rollback() // rollback after caller-visible error; surface the original error
        return fmt.Errorf("...: %w", err)
    }
    ```
    A bare `_ = tx.Rollback()` with no comment is **not acceptable** — the comment is the justification the story requires.
  - Note the lint config (`.golangci.yml`) enables `errcheck.check-blank: true`, so `_ = tx.Rollback()` alone still trips the linter. Use a `// nolint:errcheck // rollback after error; original error is what the caller cares about` directive on the rollback line, OR call `tx.Rollback()` inside an `if rbErr := tx.Rollback(); rbErr != nil { ... }` block that at minimum logs `rbErr`. Pick per site; the choice and reasoning go in `## Notes`.
  - For any non-rollback `errcheck` finding the live linter reports (e.g. an ignored `rows.Close()` outside the `bodyclose`/`sqlclosecheck` scope), fix it correctly — handle, log, or explicit-discard-with-justification.
- **Out:**
  - All other linter categories (`unused`, `noctx`, `errorlint`, `gocritic`, `gosec`, `revive`) — handled in sibling tasks.
  - Refactoring the repo layer's transaction helpers. The fix is local: at each flagged call site.
  - Edits to `.golangci.yml`.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/repo/user_story_repo.go`
- `services/agent-board/internal/repo/task_repo.go`
- Possibly the matching `*_test.go` files **only if** a tx-rollback site exists in test code that the linter flags — capture the final list in `## Notes` after the live re-baseline.

## Why this is its own task (slicing rationale)
Separated from the mechanical task because `errcheck` is judgement-bearing, not mechanical: each rollback ignore needs an individual call about discard-with-justification vs. log-and-continue, and an inline justification comment that a reviewer will read. Mixing that into the `noctx` / `errorlint` mechanical pass would force the reviewer to context-switch per finding and increase the risk of a rubber-stamp "looks fine" on a discard that should have been a log. Runs after the mechanical task because `noctx` may rewrite the same repo files (`task_repo.go`, `user_story_repo.go`) for unrelated reasons; serialising avoids merge conflicts.

## Test contract
US011 is a quality-refinement story; there is no `US011_be_unit_tests.md`. The contract is the verification commands in the story's "Acceptance criteria":
- `golangci-lint run ./...` inside `services/agent-board/` reports **zero `errcheck` findings** after this task. Other categories (`gocritic`, `gosec`, `revive`) may still be non-zero — they are addressed by the follow-up task.
- `go test ./... -race` inside `services/agent-board/` is PASS, including the existing repo rollback tests (`TestTaskRepo_UpdateTaskStatus_RollbackOnAuditFail` and equivalents in `user_story_repo_test.go`).

## Implementation notes
- Use `golangci-lint run --enable-only=errcheck ./...` to audit before and after.
- Per-site decision matrix (document each choice in `## Notes`):
  - **Rollback inside the `if err != nil` branch of an already-failing operation** → explicit discard with `// nolint:errcheck // rollback after error; ...` justification. This is the conventional safe-discard.
  - **Rollback in a `defer`** → wrap in `if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) { log.Printf(...) }` — `sql.ErrTxDone` is the expected case when the tx was committed.
  - **Anything else** (e.g. an ignored non-rollback error) → handle it, do not discard.
- Do not change the surrounding error-wrap semantics. The caller's view of the original error must not change.

## Definition of done
- All listed verification commands pass for this task's scope:
  - `golangci-lint run --enable-only=errcheck ./...` inside `services/agent-board/` reports zero findings.
  - `golangci-lint run ./...` inside `services/agent-board/` still runs (remaining `gocritic` / `gosec` / `revive` categories from the follow-up task are expected, not a failure here).
  - `go test ./... -race` inside `services/agent-board/` is PASS.
- (Track: BE) `go vet ./...` clean inside `services/agent-board/`.
- Every `// nolint:errcheck` has a one-line justification on the same or preceding line, names `errcheck` explicitly, and is not a blanket disable.
- No change to caller-visible error semantics: the error returned to the caller is the original underlying error, not the rollback error.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Notes

### Scope-shift note from US011_be_unused_handler_test_triage review (2026-05-19, tech-lead)
Re-baseline against the current working tree (post PR #1 `ee98420`) confirms `errcheck` is still 4 findings total, but the site map has shifted relative to the story background. The current 4 sites are:
- `services/agent-board/internal/handler/message.go:83` — `respBytes, _ := json.Marshal(resp)` (NEW; introduced by REQ003 implementation in PR #1)
- `services/agent-board/internal/handler/message.go:105` — `respBytes, _ := json.Marshal(resp)` (NEW; same origin)
- `services/agent-board/internal/repo/task_repo.go:97` — `_ = tx.Rollback()` (the conventional rollback-after-error site; still flagged because `.golangci.yml` enables `errcheck.check-blank: true`)
- `services/agent-board/internal/repo/user_story_repo.go:65` — `_ = tx.Rollback()` (same)

The two `message.go` sites are NOT `tx.Rollback()` — they are unchecked `json.Marshal` returns and need a different per-site decision (handle the marshal error and respond 500, or explicit-discard with justification if the input was just-validated). Apply the per-site decision matrix in `## Implementation notes` to all 4 sites; the two `message.go` sites fall into the "anything else — handle it, do not discard" bucket by default unless there's a documented reason `json.Marshal` of this specific struct cannot fail.

Add `services/agent-board/internal/handler/message.go` to the `## Files touched` list when picking this task up. The original list (`user_story_repo.go`, `task_repo.go`) remains valid for the two rollback sites.

### Dev notes (2026-05-19, be-dev US011 errcheck)

**Re-baseline result:** `golangci-lint run --enable-only=errcheck ./...` from `services/agent-board/` confirmed exactly 4 findings matching the scope-shift note verbatim:
- `internal/handler/message.go:83` — `respBytes, _ := json.Marshal(resp)`
- `internal/handler/message.go:105` — `respBytes, _ := json.Marshal(resp)`
- `internal/repo/task_repo.go:97` — `_ = tx.Rollback()`
- `internal/repo/user_story_repo.go:66` — `_ = tx.Rollback()`

No drift from the documented inventory.

**Files touched (final):**
- `services/agent-board/internal/handler/message.go`
- `services/agent-board/internal/repo/task_repo.go`
- `services/agent-board/internal/repo/user_story_repo.go`

**Per-site decisions:**

1. **`handler/message.go:83` — `sendError` / `respBytes, _ := json.Marshal(resp)`**
   - Decision: Handle the error (Form B equivalent — "anything else, handle it").
   - Reasoning: `mcp.JSONRPCResponse` contains an `interface{}` ID field; while in practice all inbound IDs are JSON-primitive, marshal could fail for pathological inputs. More importantly, the task spec explicitly places non-rollback sites in the "handle it, do not discard" bucket. Fix: assign the error, check it, log and return a 500 echo error if marshal fails; otherwise proceed to `session.QueueMessage` as before.
   - Caller semantics: unchanged — successful path still queues and returns `echo.NewHTTPError(StatusOK, resp)`.

2. **`handler/message.go:105` — `sendToolResultError` / `respBytes, _ := json.Marshal(resp)`**
   - Decision: Handle the error (same rationale as site 1).
   - Fix: same pattern — log marshal failure, return 500 without queuing.

3. **`repo/task_repo.go:97` — `defer` rollback inside `UpdateTaskStatus`**
   - Decision: Form B — `if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) { log.Printf(...) }`.
   - Reasoning: The rollback is inside a `defer func()` guarded by `if err != nil`. The task spec prescribes Form B for deferred rollbacks. The `sql.ErrTxDone` guard is necessary because if the commit succeeded, `err` is nil so the defer won't execute the rollback — but having the guard makes it robust if the guard condition is ever modified. The `TestTaskRepo_UpdateTaskStatus_RollbackOnAuditFail` test uses `sqlmock.ExpectRollback()` and still passes because `tx.Rollback()` is still called.
   - Added `"log"` import.
   - Caller semantics: unchanged — original error is still returned.

4. **`repo/user_story_repo.go:66` — `defer` rollback inside `UpdateUserStoryStatus`**
   - Decision: Form B — same pattern as site 3.
   - Reasoning: Same as site 3. The `TestUserStoryRepo_UpdateUserStoryStatus_RollbackOnAuditFailure` test still passes because `tx.Rollback()` is still called.
   - Added `"log"` import.
   - Caller semantics: unchanged.

**Verification results:**
- `golangci-lint run --enable-only=errcheck ./...` → 0 issues (exit 0). PASS.
- `golangci-lint run ./...` → 7 issues: gocritic=5, gosec=1, revive=1. errcheck=0 confirmed. Expected remainder intact.
- `go test ./... -race -count=1` run 1 → all packages ok, no DATA RACE. PASS.
- `go test ./... -race -count=1` run 2 → all packages ok, no DATA RACE. PASS.
- `go vet ./...` → clean (no output). PASS.
- `scripts/review/run-gate.sh be services/agent-board` → exit 0. PASS.
- `scripts/review/run-gate.sh cross` → exit 0. PASS.

**Suppression inventory:** 0 nolint directives. All 4 sites fixed via actual error handling — no suppressions required.

## Review log
(tech-lead appends here on each review pass)

### Review pass 1 — 2026-05-19 — verdict: approved

**Verification (re-run by tech-lead, not trusting dev report):**
- `golangci-lint run --enable-only=errcheck ./...` → `0 issues.` exit 0. UT-006 PASS.
- `golangci-lint run ./...` → `7 issues:` exit 1. Per-linter breakdown: `gocritic: 5, gosec: 1, revive: 1`. `errcheck`, `errorlint`, `noctx`, `unused` all absent — no regression across previously-cleared categories. Exact remaining inventory for task 4:
  - `cmd/api-server/main.go:45` `exitAfterDefer` (gocritic)
  - `cmd/mcp-server/main.go:31` `exitAfterDefer` (gocritic)
  - `internal/repo/task_repo.go:162` `sloppyReassign` (gocritic)
  - `internal/repo/user_story_repo.go:89` `sloppyReassign` (gocritic)
  - `internal/repo/user_story_repo.go:133` `sloppyReassign` (gocritic)
  - `cmd/api-server/main.go:58` `G706` log injection (gosec)
  - `internal/handler/message.go:15` `var-naming sessionId → sessionID` (revive)
- `go vet ./...` → clean (exit 0, no output).
- `go test ./... -race -count=1` run 1 → `ok` for domain/handler/mcp/repo; cmd packages have no tests; no `FAIL`, no `DATA RACE`. Exit 0. UT-002 PASS.
- `go test ./... -race -count=1` run 2 (back-to-back) → identical result, exit 0. No timing perturbation.
- `go test ./internal/repo/... -race -run "Rollback" -v` → `TestTaskRepo_UpdateTaskStatus_RollbackOnAuditFail PASS`, `TestUserStoryRepo_UpdateUserStoryStatus_RollbackOnAuditFailure PASS`. `sqlmock.ExpectRollback()` satisfied — Form B rewrite still drives the rollback call, no "unexpected" or "missing" mock errors.
- `scripts/review/run-gate.sh be services/agent-board` → `REVIEW GATE: PASS` exit 0 (gofmt, go vet, golangci-lint (gate runs `--no-config` defaults), go test, gosec, govulncheck all PASS).
- `scripts/review/run-gate.sh cross` → `REVIEW GATE: PASS` exit 0 (semgrep owasp/golang/typescript/react, gitleaks all PASS).
- `grep -rn "tx\.Rollback()" services/agent-board/internal/repo/` → both matches are inside `if rbErr :=` blocks. UT-006 grep contract satisfied.
- `grep -rn "nolint" services/agent-board/` → exit 1 (no matches). Module-wide `// nolint` count remains **0**. UT-010 hygiene contract intact.

**Diff-quality observations:**
- Worktree commit `1de1e75` modified exactly 4 files: 3 source (`handler/message.go`, `repo/task_repo.go`, `repo/user_story_repo.go`) + 1 task file. No edits to `architecture.md`, `.golangci.yml`, test files, sibling task files, or anything in `cmd/`. Scope: tight, matches the dev's claim verbatim.
- Test files (`*_test.go` under `internal/repo/` and `internal/handler/`) are byte-for-byte unchanged across the commit — verified via `git diff 5a0857a..1de1e75 -- *_test.go`. No test weakening.
- **`message.go` error-handling quality:** the dev chose Form B / handle-the-error consistently across both `sendError` and `sendToolResultError`. The 500 fallback uses `echo.NewHTTPError(http.StatusInternalServerError, "internal error")` with a plain string body — no recursive call back into `json.Marshal` on the same value, so the error path cannot infinitely recurse. The log lines use `%v` on the marshaling error itself (an internal Go-runtime value, not user input), so no log-injection vector is introduced — gosec G706 in task 4 stays scoped to its existing `cmd/api-server/main.go:58` site. External contract of both functions on the success path is byte-identical to pre-change.
- **Form B choice for the rollback sites:** correct call, consistently applied. Both repo functions use a `defer func() { if err != nil { ... } }` pattern, so Form B is the right pick per UT-006 ("preferred if rollback failures should be visible in logs", and the only safe form for deferred rollbacks where `sql.ErrTxDone` is a legitimate non-error). The `sql.ErrTxDone` guard is correct — it suppresses the noise for already-committed transactions while surfacing genuine rollback failures. Log strings include operation context (`task tx` / `user_story tx`) without exposing row data, IDs, or secrets.
- **Logger choice:** standard library `log.Printf` — idiomatic for this codebase. `internal/handler/project_handler.go:28` already uses `log.Printf` for an analogous "failed sub-operation, surface to operator" log. No structured logger is in use anywhere under `services/agent-board/internal/`, so matching `log` is correct.
- **Caller-visible error semantics:** preserved exactly. `UpdateTaskStatus` still returns `fmt.Errorf("failed to update task status: %w", err)` / `ErrNotFound` / `fmt.Errorf("failed to insert audit log: %w", err)` / `fmt.Errorf("failed to commit transaction: %w", err)` chains as before. `UpdateUserStoryStatus` still returns the `ErrNotFound` sentinel via the `err = ErrNotFound` assignment then `return nil, err`. The rollback log is fire-and-forget; it does not shadow, replace, or wrap the original error the caller cares about.

**Working baseline for the next (final) task** (`US011_be_tail_gocritic_gosec_revive.md`): **7 findings remaining** — gocritic 5, gosec 1, revive 1 — fully listed above with file:line. Task 4 must drive this to 0 and satisfy UT-001 (lint exits clean), UT-005, UT-008, UT-009, UT-010.

Streak: **3 consecutive approved across this story** (unused triage → noctx/errorlint mechanical → errcheck). Clean.
