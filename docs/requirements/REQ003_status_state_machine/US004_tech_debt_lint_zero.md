# US004 — Tech debt: drive golangci-lint to zero in services/agent-board

**Requirement:** REQ003 — status_state_machine
**Status:** done

## User Story
As the engineering team (and the future AI agents who will edit this codebase), I want `golangci-lint run ./...` to exit clean inside `services/agent-board/` against the project's real ruleset, so that the lint gate is a trustworthy source of truth and stops silently bypassing 34 real findings every time CI / a dev runs it.

## Background

Verbatim source material from the engineering session that surfaced this debt:

> Now that the real ruleset actually runs, golangci-lint surfaces 34 real findings in `services/agent-board/` that were silently bypassed before. Breakdown:
>
> | Linter | Count | Notable spots |
> |---|---|---|
> | `noctx` | 11 | `httptest.NewRequest` calls + `db.Ping` usages (should be `PingContext`) |
> | `unused` | 9 | Mostly in `internal/handler/handler_test.go` — likely some helpers from a recent race fix the linter thinks aren't used in `_test.go` builds |
> | `gocritic` | 5 | `sloppyReassign` etc. in repo layer |
> | `errcheck` | 4 | Ignored `tx.Rollback()` returns in `user_story_repo` / `task_repo` |
> | `errorlint` | 3 | `err == sql.ErrNoRows` instead of `errors.Is` |
> | `gosec` | 1 | Log-injection warning in a `cmd/*/main.go` |
> | `revive` | 1 | (single finding) |

Total: **34 findings** across 7 linters. The goal of this story is to drive that number to zero, or to an explicit and individually-justified set of `// nolint:<linter>` suppressions where a finding is a genuine false positive.

## Acceptance criteria
- **Scenario: Lint exits clean**
  - Given the migrated v2 `.golangci.yml` at the repo root
  - When I run `golangci-lint run ./...` from inside `services/agent-board/`
  - Then the command exits with status 0 and reports zero findings.
- **Scenario: Any remaining suppressions are explicit and justified**
  - Given a finding the team has decided to keep (e.g. an unavoidable false positive)
  - When I look at the suppression site
  - Then there is an explicit `// nolint:<linter>` directive naming the linter(s) being suppressed, with a one-line justification comment on the same or preceding line. No blanket `// nolint` and no file-level disables.
- **Scenario: `unused` cluster in handler_test.go is triaged first**
  - Given the 9 `unused` findings (most concentrated in `internal/handler/handler_test.go`)
  - When this story is worked
  - Then the `unused` cluster is reviewed before any other linter category, because some of these helpers are likely a self-inflicted artifact of the recent SSE-race fix and may simply be prunable rather than suppressible. Pruning is preferred to suppressing where the helper genuinely has no caller.
- **Scenario: No behaviour change in production code under -race**
  - Given the changes made to clear the lint findings
  - When I run `go test ./... -race` inside `services/agent-board/`
  - Then the result is PASS with no new failures or data-race reports relative to the pre-story baseline.
- **Scenario: Specific finding categories are resolved correctly, not papered over**
  - `noctx` (11): `httptest.NewRequest` call sites use `httptest.NewRequestWithContext` (or pass a context via `.WithContext`), and `db.Ping` is replaced with `db.PingContext` carrying the appropriate context.
  - `errcheck` (4): ignored `tx.Rollback()` returns in `user_story_repo` / `task_repo` are either handled or explicitly discarded with a documented justification (rollback-after-error is the conventional safe-discard, but it must still be explicit).
  - `errorlint` (3): `err == sql.ErrNoRows` comparisons are rewritten as `errors.Is(err, sql.ErrNoRows)`.
  - `gocritic` (5): `sloppyReassign` and friends in the repo layer are rewritten in idiomatic form.
  - `gosec` (1): the log-injection warning in `cmd/*/main.go` is addressed by sanitising the logged value or by `// nolint:gosec` with a justification if the input is provably safe.
  - `revive` (1) and any remaining single findings: addressed in idiomatic Go.

## UI / UX flow expectations
No UI: backend-only quality refinement, no user-facing surface.

## Out of scope
- Any feature work or new behaviour in `services/agent-board/`.
- Any architectural changes (the architecture in `architecture.md` is locked).
- Any changes under `web/` or `tests/e2e/`.
- Migrating or modifying `.golangci.yml` itself — the migrated v2 ruleset is the input to this story, not its output. If a rule needs to change, that is a separate decision and a separate story.
- Other services under `services/` (the baseline of 34 findings is scoped to `services/agent-board/` only).
- Upgrading dependencies (e.g. pgx, Echo) — out of scope; that was REQ001 US006.

## Dependencies
- REQ003 US001, US002, US003 (all `done`) — this story cleans up debt in the same service those stories built.
- The migrated v2 `.golangci.yml` must remain in place at the repo root.

## Notes for the team
- **Lint baseline source:** The 34-finding baseline was captured on **2026-05-19** against the **working-tree state** of `.golangci.yml` at that date. The migrated v2 ruleset is currently uncommitted in the working tree, so there is no committed SHA to anchor the baseline against. The first action when picking up this story should be to re-run `golangci-lint run ./...` inside `services/agent-board/` and confirm the per-linter counts still match (or document any drift) before beginning fixes.
- **Suggested order of work:**
  1. Triage the `unused` cluster (9) — prune dead helpers in `internal/handler/handler_test.go` first; this likely shrinks the diff fastest.
  2. Mechanical fixes: `errorlint` (3) → `errors.Is`. `noctx` (11) → `httptest.NewRequestWithContext` and `db.PingContext`.
  3. `errcheck` rollback ignores (4) — explicit discard with a justification comment.
  4. `gocritic` (5), `gosec` (1), `revive` (1) — case-by-case.
- **Watch out for:** changing test contexts via `noctx` fixes can occasionally alter timeout behaviour under `-race`. Re-run `go test ./... -race` after that batch specifically.
- **Definition of done for sign-off:** test report must include the output of `golangci-lint run ./...` (zero findings) AND `go test ./... -race` (PASS), both run inside `services/agent-board/`.

## Sign-off log
(po-ba appends here on each sign-off pass)

### Sign-off pass 1 — 2026-05-19 — verdict: approved
**Reviewed by:** po-ba (sign-off mode)
**Story status flipped to:** `done`

**Story-level gate (verbatim from the test report — the artefact the "Definition of done for sign-off" demands):**

```
$ cd services/agent-board && golangci-lint run ./...
0 issues.
$ echo $?
0
```

```
$ cd services/agent-board && go test ./... -race -count=1
?   	agent-board/cmd/api-server	[no test files]
?   	agent-board/cmd/mcp-server	[no test files]
ok  	agent-board/internal/domain	1.418s
ok  	agent-board/internal/handler	1.875s
ok  	agent-board/internal/mcp	2.038s
ok  	agent-board/internal/repo	2.282s
$ echo $?
0
```

Both Definition-of-done-for-sign-off gates pass cleanly. The 34-finding baseline from 2026-05-19 is now zero; `-race` is clean across all four packages with no `FAIL` lines and no `DATA RACE` blocks.

**Spec review (UT-001..UT-010 coverage in `US004_be_unit_tests.md`):**
- All 10 UT-* cases (UT-001 lint-clean / UT-002 race-clean / UT-003 noctx / UT-004 unused / UT-005 gocritic / UT-006 errcheck / UT-007 errorlint / UT-008 gosec / UT-009 revive / UT-010 suppression hygiene) are individually green in the test report, with per-linter `--enable-only=<linter>` cross-checks each returning `0 issues.`. The tester's static-analysis spec is the right shape for a no-behaviour-change story; no unit/integration cases were dropped, skipped, or silently weakened.
- `US004_fe_unit_tests.md` and `US004_e2e_tests.md` are appropriate N/A stubs — this is a backend quality refinement with no UI surface and no API contract change. The N/A decision is honoured correctly.

**AC scenario coverage (all 5 met):**
1. *Lint exits clean* — `golangci-lint run ./...` exits 0, zero findings. PASS.
2. *Any remaining suppressions are explicit and justified* — exactly 1 directive in the entire service (`cmd/api-server/main.go:75 //nolint:gosec`), names the linter, justification on the preceding line ("G706: safePort has all control chars (< 0x20 and DEL) stripped above; log injection is not possible."), not blanket, not file-level. Sanitisation was applied first (`strings.Map` strips runes `< 0x20` and DEL) — the suppression is the necessary secondary measure because gosec's taint analysis cannot follow `strings.Map`. Genuine technical reason, not boilerplate. PASS.
3. *`unused` cluster in handler_test.go is triaged first* — task 1 re-baselined and discovered the 9 `unused` findings were already eliminated by merged PR #1 (`ee98420`, SSE-race fix); zero `nolint:unused` directives added; tester drift acknowledgement followed correctly. PASS.
4. *No behaviour change in production code under -race* — both `-race` runs (back-to-back, twice each pass) PASS across all 4 packages, exit 0, no DATA RACE. The `noctx` rewrites (which the story flagged as the risky batch) did not perturb timing. PASS.
5. *Specific finding categories resolved correctly, not papered over* — `noctx` (11) via `httptest.NewRequestWithContext` + `db.PingContext`; `errcheck` (4) via real error-handling (2× `json.Marshal` → log+500 in `message.go`; 2× `tx.Rollback` → Form B `if rbErr := ...; rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone)`); `errorlint` (3) via `errors.Is`; `gocritic` (5) via `exitAfterDefer` `main()`→`run() error` split + 3 `sloppyReassign` `err :=` rewrites; `gosec` (1) sanitise + justified suppress; `revive` (1) `sessionId`→`sessionID` local rename with the query-param string literal preserved (no API break). PASS.

**Scope discipline:** Cumulative service-code diff `ee98420`..`912724e` is 8 files / +73 / -33 LOC, all inside `services/agent-board/`. Confirmed no edits to `architecture.md`, `.golangci.yml`, `web/`, `tests/e2e/`, or other services — exactly as the story's "Out of scope" section required.

**Tester coverage acknowledgement:** The 10 UT-* static-analysis cases in `US004_be_unit_tests.md` (including the per-linter `--enable-only` verification commands and the UT-010 grep-contract for suppression hygiene) were the right shape for a quality-refinement story with no new behaviour, and the test report demonstrates each one as PASS with a per-linter cross-check table. Good spec hygiene; no spec gap.

**Provenance — the four task review commits referenced in the test report (chain integrity verified):**
- `5dc20c1` — `US004_be_unused_handler_test_triage.md` review pass 1 approved (verified no-op due to PR #1 drift)
- `5a0857a` — `US004_be_mechanical_noctx_errorlint.md` review pass 1 approved (`noctx` 11 + `errorlint` 3 → 0)
- `5b55e6e` — `US004_be_errcheck_rollback_discard.md` review pass 1 approved (`errcheck` 4 → 0)
- `912724e` — `US004_be_tail_gocritic_gosec_revive.md` review pass 1 approved (`gocritic` 5 + `gosec` 1 + `revive` 1 → 0; owns the story-wide exit-zero gate)

All four tasks are `Status: completed`. Four consecutive `approved` tech-lead verdicts across the chain; circuit breaker never engaged.

**Spot-check from po-ba (independent of test report):** `grep -rn "nolint" services/agent-board/` returns exactly the single declared `//nolint:gosec` line at `cmd/api-server/main.go:75`, with the justification comment present on line 74 of `main.go`. The suppression inventory in the test report matches the working tree exactly.

**Routed to:** none — story closed.
