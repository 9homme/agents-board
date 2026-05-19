# US004/be_errcheck_rollback_discard

**Requirement:** REQ003
**Story:** US004
**Track:** BE
**Service:** services/agent-board
**Status:** in_review
**Blocked by:** US004_be_mechanical_noctx_errorlint.md
**Worked-by:** be-dev (US004 errcheck)
**Implements:** US004 acceptance criterion "Specific finding categories are resolved correctly, not papered over" — drives `errcheck` (4) to zero, specifically the ignored `tx.Rollback()` returns in `user_story_repo` / `task_repo`.

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
US004 is a quality-refinement story; there is no `US004_be_unit_tests.md`. The contract is the verification commands in the story's "Acceptance criteria":
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

### Scope-shift note from US004_be_unused_handler_test_triage review (2026-05-19, tech-lead)
Re-baseline against the current working tree (post PR #1 `ee98420`) confirms `errcheck` is still 4 findings total, but the site map has shifted relative to the story background. The current 4 sites are:
- `services/agent-board/internal/handler/message.go:83` — `respBytes, _ := json.Marshal(resp)` (NEW; introduced by REQ003 implementation in PR #1)
- `services/agent-board/internal/handler/message.go:105` — `respBytes, _ := json.Marshal(resp)` (NEW; same origin)
- `services/agent-board/internal/repo/task_repo.go:97` — `_ = tx.Rollback()` (the conventional rollback-after-error site; still flagged because `.golangci.yml` enables `errcheck.check-blank: true`)
- `services/agent-board/internal/repo/user_story_repo.go:65` — `_ = tx.Rollback()` (same)

The two `message.go` sites are NOT `tx.Rollback()` — they are unchecked `json.Marshal` returns and need a different per-site decision (handle the marshal error and respond 500, or explicit-discard with justification if the input was just-validated). Apply the per-site decision matrix in `## Implementation notes` to all 4 sites; the two `message.go` sites fall into the "anything else — handle it, do not discard" bucket by default unless there's a documented reason `json.Marshal` of this specific struct cannot fail.

Add `services/agent-board/internal/handler/message.go` to the `## Files touched` list when picking this task up. The original list (`user_story_repo.go`, `task_repo.go`) remains valid for the two rollback sites.

### Dev notes (2026-05-19, be-dev US004 errcheck)

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
