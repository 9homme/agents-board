# US004/be_tail_gocritic_gosec_revive

**Requirement:** REQ003
**Story:** US004
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** US004_be_errcheck_rollback_discard.md
**Worked-by:** be-dev (US004 tail)
**Implements:** US004 final acceptance criterion "Scenario: Lint exits clean" — drives the remaining tail (`gocritic` 5, `gosec` 1, `revive` 1 = 7 findings) to zero AND verifies the overall `golangci-lint run ./...` exit-zero gate.

## Goal
Resolve the remaining case-by-case lint findings — `gocritic` (5, `sloppyReassign` etc. in the repo layer), `gosec` (1, log-injection in a `cmd/*/main.go`), and `revive` (1) — and then confirm the story-level acceptance: `golangci-lint run ./...` exits 0 with zero findings inside `services/agent-board/`.

## Scope
- **In:**
  - **`gocritic` (5):** Rewrite the flagged sites in idiomatic Go. Expected concentration: repo layer (`sloppyReassign` and friends in `services/agent-board/internal/repo/*.go`). Confirm exact sites against the live linter run. Common `gocritic` rewrites: `sloppyReassign` → declare the variable in the right place; `appendAssign` / `appendCombine` → fold appends; `commentFormatting` → fix comment shape. Choose the idiomatic Go form, not a clever one.
  - **`gosec` (1):** Address the log-injection warning in `services/agent-board/cmd/api-server/main.go` or `services/agent-board/cmd/mcp-server/main.go` (the story names "a `cmd/*/main.go`" — find it via the live linter run). Two acceptable resolutions:
    1. **Sanitise** the logged value: strip or escape newlines / control characters before logging (`strings.ReplaceAll(v, "\n", " ")` and similar), or log a structured field rather than a formatted string.
    2. **Suppress with `// nolint:gosec // <one-line justification>`** only if the input is provably safe (e.g. it's a hard-coded constant, or a value already validated by a typed parser earlier in the same function). The justification must explain WHY it is safe, not just that the dev believes it is.
  - **`revive` (1):** Address the single `revive` finding. The repo's `revive` rule set is enabled for `exported`, `var-naming`, `package-comments`, `context-as-argument`, `error-return`, `error-naming` (see `.golangci.yml`). The fix shape depends on which sub-rule fires — typically a missing doc comment on an exported symbol, a misnamed error variable, or a misordered context argument. Fix per the rule.
  - **Final verification gate:** After this task's fixes land, run `golangci-lint run ./...` inside `services/agent-board/` and confirm **zero findings overall** — this is the story-level acceptance criterion and this is the task that proves it.
- **Out:**
  - Re-opening any decision made in the earlier US004 tasks. If a finding from a previous category re-surfaces (e.g. a `noctx` regression because a new call site was added), flag it back rather than silently re-fixing it under this task.
  - Refactoring the repo layer beyond what `gocritic` literally requires.
  - Edits to `.golangci.yml`.

## Files touched (estimated, exclusive)
- One or more files under `services/agent-board/internal/repo/` (for `gocritic`)
- `services/agent-board/cmd/api-server/main.go` OR `services/agent-board/cmd/mcp-server/main.go` (for `gosec` — exact file from the live linter run)
- One file under `services/agent-board/internal/` (for `revive` — exact file from the live linter run)

(The dev should narrow this list to the exact files the live linter run reports and capture the final list in `## Notes`.)

## Why this is its own task (slicing rationale)
The tail is bundled because each individual category is a single-digit count and would not justify a standalone task, and because this is the task that owns the final story-level acceptance gate (`golangci-lint run ./...` exits 0 overall). Runs last in the chain so that any regression introduced by an earlier task is caught here and routed back via `changes_requested` on whichever task introduced it, rather than silently absorbed.

## Test contract
US004 is a quality-refinement story; there is no `US004_be_unit_tests.md`. The contract is the verification commands in the story's "Acceptance criteria" and the story's "Definition of done for sign-off":
- **Final story-level acceptance:** `golangci-lint run ./...` inside `services/agent-board/` exits 0 with **zero findings overall** (not just for this task's three categories — every linter must be clean).
- `go test ./... -race` inside `services/agent-board/` is PASS.

## Implementation notes
- Use `golangci-lint run --enable-only=gocritic ./...`, `--enable-only=gosec ./...`, and `--enable-only=revive ./...` to audit each category independently before and after.
- For the log-injection finding, prefer **sanitisation** to suppression. A `// nolint:gosec` should be the last resort and must name `gosec` explicitly with a justification.
- After this task lands, run the full `golangci-lint run ./...` (no `--enable-only` filter) inside `services/agent-board/` and capture the exit code + the (expected empty) finding list verbatim in `## Notes`. This is the artefact the sign-off test report will quote.

## Definition of done
- **Story-level final gate met:**
  - `golangci-lint run ./...` inside `services/agent-board/` exits 0 and reports zero findings. **This is the story's primary acceptance criterion and the gate this task owns.**
  - `go test ./... -race` inside `services/agent-board/` is PASS, no new failures, no new race reports relative to the pre-story baseline.
- This task's own per-category gates:
  - `golangci-lint run --enable-only=gocritic ./...` reports zero findings.
  - `golangci-lint run --enable-only=gosec ./...` reports zero findings.
  - `golangci-lint run --enable-only=revive ./...` reports zero findings.
- (Track: BE) `go vet ./...` clean inside `services/agent-board/`.
- Every remaining `// nolint:<linter>` directive across the service has a one-line justification on the same or preceding line, names the linter(s) explicitly, and is not a blanket `// nolint` or file-level disable.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev captured the verbatim output of `golangci-lint run ./...` (exit 0, zero findings) in `## Notes` so the test report can quote it.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Notes

### Baseline confirmation (2026-05-19)
Re-baseline confirmed 7 findings exactly as specified: gocritic=5, gosec=1, revive=1. No drift.

### Per-finding fix descriptions (7 entries)

1. **gocritic exitAfterDefer — `cmd/api-server/main.go:45`**
   Refactored `main()` into `main()` + `run() error`. The `run()` function owns the `defer db.Close()` and all startup logic; it returns errors rather than calling `log.Fatalf`. `main()` calls `run()` and uses `os.Exit(1)` on non-nil result. The `defer db.Close()` now unwinds correctly as part of `run()`'s stack.

2. **gocritic exitAfterDefer — `cmd/mcp-server/main.go:31`**
   Same `run() error` split applied. `main()` is now a thin wrapper; `run()` owns DB open, defer-close, ping, server registration, and `e.Start()`. Startup semantics (same port, same SSE/message routes, same tool registrations) are preserved exactly.

3. **gocritic sloppyReassign — `internal/repo/task_repo.go:162`**
   Rewrote `if err = rows.Err(); err != nil {` → `if err := rows.Err(); err != nil {`. Safe: the outer `err` (from `QueryContext`) is not used after this point; the function's only remaining `return` is `return tasks, nil`.

4. **gocritic sloppyReassign — `internal/repo/user_story_repo.go:89`**
   Rewrote `if err = tx.Commit(); err != nil {` → `if err := tx.Commit(); err != nil {`. Safe: the deferred rollback closure captures the outer `err` (last set by `tx.ExecContext` at line 84); on a commit failure, the outer `err` is nil so the defer does not attempt a rollback — correct, since a failed commit leaves the tx in a terminal state (`sql.ErrTxDone`).

5. **gocritic sloppyReassign — `internal/repo/user_story_repo.go:133`**
   Rewrote `if err = rows.Err(); err != nil {` → `if err := rows.Err(); err != nil {`. Same analysis as finding 3 — no subsequent use of outer `err`.

6. **gosec G706 — `cmd/api-server/main.go:58`**
   **Choice: sanitise + suppress with justification.**
   `port` is read from the `PORT` environment variable. While PORT is typically digits-only, gosec's taint analysis treats any env-var-derived string as tainted. Applied defence-in-depth sanitisation: `strings.Map` strips all control characters (runes < 0x20 and DEL/0x7f) producing `safePort`. Gosec's taint analysis cannot prove the sanitisation statically (it still flags the derived variable), so a `//nolint:gosec` directive was added on the `log.Printf` line with a justification comment on the immediately preceding line explaining that `safePort` has all control chars stripped. The server still starts on `:port` (the original value, not `safePort`) — the sanitisation is logging-only.

7. **revive var-naming — `internal/handler/message.go:15`**
   Renamed local variable `sessionId` → `sessionID` (Go acronym convention). All four references in `message.go` updated. The `sessionId` string literal in the query param (`c.QueryParam("sessionId")`) and in error message strings is unchanged — those are API-level strings, not Go identifiers.

### gosec sanitise-vs-suppress decision (finding 6)
Sanitisation was applied first (strip control chars via `strings.Map`). The suppression was added as a secondary measure because gosec's taint analysis does not follow the transformation and continues to flag the sanitised variable. The `//nolint:gosec` directive names `gosec` explicitly and carries a justification comment on the preceding line. The justification explains WHY the value is safe (control chars stripped), not merely that the developer believes it is.

### Verbatim `golangci-lint run ./...` output (exit 0, zero findings)

```
$ cd services/agent-board && golangci-lint run ./...
0 issues.
```

Exit status: 0.

### Verbatim `go test ./... -race` summary lines (both runs)

Run 1:
```
?   	agent-board/cmd/api-server	[no test files]
?   	agent-board/cmd/mcp-server	[no test files]
ok  	agent-board/internal/domain	1.801s
ok  	agent-board/internal/handler	2.296s
ok  	agent-board/internal/mcp	2.553s
ok  	agent-board/internal/repo	3.042s
```

Run 2:
```
?   	agent-board/cmd/api-server	[no test files]
?   	agent-board/cmd/mcp-server	[no test files]
ok  	agent-board/internal/domain	1.323s
ok  	agent-board/internal/handler	2.061s
ok  	agent-board/internal/mcp	1.579s
ok  	agent-board/internal/repo	2.257s
```

Both runs: all packages ok, no FAIL lines, no DATA RACE.

### Final suppression inventory

1 suppression across the entire service:

| File | Line | Linter | Directive | Justification |
|---|---|---|---|---|
| `cmd/api-server/main.go:75` | 75 | `gosec` | `//nolint:gosec` | Preceding line (74): "G706: safePort has all control chars (< 0x20 and DEL) stripped above; log injection is not possible." |

No blanket `// nolint` directives. No file-level disables.

### Files touched (actual)
- `services/agent-board/cmd/api-server/main.go` (exitAfterDefer refactor + gosec sanitisation + suppression)
- `services/agent-board/cmd/mcp-server/main.go` (exitAfterDefer refactor)
- `services/agent-board/internal/repo/task_repo.go` (sloppyReassign fix, line 162)
- `services/agent-board/internal/repo/user_story_repo.go` (sloppyReassign fixes, lines 89 and 133)
- `services/agent-board/internal/handler/message.go` (revive var-naming fix, line 15)
- `docs/requirements/REQ003_status_state_machine/US004_be_tail_gocritic_gosec_revive.md` (this file)

## Review log
(tech-lead appends here on each review pass)

### Review pass 1 — 2026-05-19 — verdict: approved

**Story-level acceptance gate (the headline this task owns):**

```
$ cd services/agent-board && golangci-lint run ./...
0 issues.
```
Exit status: 0. **US004 story-level acceptance criterion "Scenario: Lint exits clean" is now met.** Baseline drove from 34 → 0 across all 7 linter categories (noctx, unused, gocritic, errcheck, errorlint, gosec, revive).

**Per-category audits (re-run by reviewer, not trusted from dev report):**
- `golangci-lint run --enable-only=gocritic ./...` → `0 issues.` exit 0.
- `golangci-lint run --enable-only=gosec ./...` → `0 issues.` exit 0.
- `golangci-lint run --enable-only=revive ./...` → `0 issues.` exit 0.
- `go vet ./...` → clean, exit 0.

**Race tests (run twice, both PASS, no DATA RACE):**

Run 1 / Run 2 — both:
```
?   	agent-board/cmd/api-server	[no test files]
?   	agent-board/cmd/mcp-server	[no test files]
ok  	agent-board/internal/domain
ok  	agent-board/internal/handler
ok  	agent-board/internal/mcp
ok  	agent-board/internal/repo
```
Exit 0 both runs; `grep -E "FAIL|DATA RACE"` returned zero matches.

**Mandatory review gate scripts:**
- `scripts/review/run-gate.sh be services/agent-board` → `REVIEW GATE: PASS` exit 0 (gofmt, go vet, golangci-lint, go test, gosec, govulncheck all PASS).
- `scripts/review/run-gate.sh cross` → `REVIEW GATE: PASS` exit 0 (semgrep, gitleaks PASS).

**Behavioural-preservation checks (one bullet per code change):**
- `cmd/api-server/main.go` `main()`→`run() error` refactor: identical CORS config, same `DATABASE_URL` env-var read, same `/api/v1/projects` route, same `PORT` defaulting, server still binds on `:" + port` (line 76) using the original un-sanitised value — `safePort` is logging-only as claimed. `defer db.Close()` now lives inside `run()` so it unwinds before `main`'s `os.Exit(1)`. No new globals, no timeouts changed, no new env-var reads.
- `cmd/mcp-server/main.go` `main()`→`run() error` refactor: identical tool registrations (Project / Document / UserStory / Task / Audit), identical `/sse` GET and `/message` POST routes, same `DB_URL` env, same `PORT` defaulting, same `:"+port` bind. The pre-existing `log.Fatal` on missing `DB_URL` (line 28) is before any `defer` so does not re-trigger `exitAfterDefer` — correctly left as-is.
- `cmd/api-server/main.go:68-75` gosec G706 sanitisation: `strings.Map` callback strips runes `< 0x20` AND `r == 0x7f` (DEL) as required; `safePort` used ONLY on the `log.Printf` line (75); `e.Start(":" + port)` at line 76 uses the original `port` — server actually listens on the intended port. `//nolint:gosec` directive names `gosec` explicitly and the justification on line 74 explains WHY (control chars stripped), not merely that the dev believes it is safe. Sanitise-first approach was correct; the suppression is the unavoidable secondary measure because gosec's taint analysis does not follow `strings.Map`.
- `internal/repo/task_repo.go:162` `if err :=` rewrite: `ListTasks` — no transaction, only `defer rows.Close()`. Outer `err` from `QueryContext` (line 147) is never used after the `rows.Err()` check; the only subsequent statement is `return tasks, nil`. Shadowing is behaviour-neutral. Safe.
- `internal/repo/user_story_repo.go:89` `if err := tx.Commit()` rewrite (highest-risk site, verified end-to-end): the deferred rollback closure (lines 65–71) captures outer `err`. At line 89 the outer `err` is guaranteed `nil` (set at line 84 by `tx.ExecContext`; line 85–87 returns on non-nil). If commit fails, the inner `err` is returned directly (line 90); the outer `err` remains `nil`, so the deferred closure's `if err != nil` correctly skips rollback — which is the right semantics, since a failed `tx.Commit()` already terminates the transaction (`sql.ErrTxDone` on rollback attempt). The rollback safety net `TestUserStoryRepo_UpdateUserStoryStatus_RollbackOnAuditFailure` (in `internal/repo`) PASSes under both race runs. Dev's analysis matches reality.
- `internal/repo/user_story_repo.go:133` `if err :=` rewrite: `ListUserStories` — same shape as `task_repo.go:162` (no tx, `rows.Close()` defer, outer `err` from `QueryContext` not used after `rows.Err()` check). Safe.
- `internal/handler/message.go:15` `sessionId` → `sessionID` rename: purely local. The query-param string literal `"sessionId"` (line 15 `c.QueryParam("sessionId")`) is unchanged — API contract preserved. Error message strings "sessionId is required" / "invalid sessionId" unchanged (lines 17, 22). No exported field, struct tag, JSON tag, or wire identifier touched. All four in-function references updated consistently (15, 16, 20).

**Final suppression inventory (UT-010 satisfied):**
- `grep -rn "nolint" services/agent-board/` returns exactly 1 match:
  - `cmd/api-server/main.go:75 — //nolint:gosec` (linter named, justification on preceding line 74, not blanket, not file-level).
- `grep -rn "// nolint$" services/agent-board/` and `grep -rn "//nolint$" services/agent-board/` both return 0 lines.

**gosec sanitise+nolint choice:** Acceptable. Sanitisation was applied substantively (`strings.Map` strips control chars and DEL), satisfying the AC's preferred Form A; the `//nolint:gosec` is a necessary secondary measure because gosec's taint analysis cannot prove transformation safety statically. Justification is genuine (explains the stripping) rather than boilerplate.

**Scope discipline:** Cumulative US004 diff against pre-story baseline (`ee98420`..`9769d8d`): 8 service files changed, +73 / -33 LOC. No `architecture.md`, no `.golangci.yml`, no `US004_*_tests.md` spec files, no sibling-task `.md` files modified beyond legitimate `## Notes` and `## Review log` appends.

**Streak:** 4th consecutive `approved` across the US004 chain (`be_unused_handler_test_triage` → `be_mechanical_noctx_errorlint` → `be_errcheck_rollback_discard` → `be_tail_gocritic_gosec_revive`). The chain is now closed; Phase 3c (test report capture) can begin.
