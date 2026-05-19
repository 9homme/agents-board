# US004/be_unused_handler_test_triage

**Requirement:** REQ003
**Story:** US004
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** 
**Worked-by:** be-dev (US004 unused triage)
**Implements:** US004 acceptance criterion "Scenario: `unused` cluster in handler_test.go is triaged first" — drives the `unused` linter category to zero (9 findings).

## Goal
Re-baseline the lint findings against the current working tree, then drive the `unused` linter category (9 findings, mostly concentrated in `internal/handler/handler_test.go`) to zero by **pruning** dead helpers rather than suppressing them, except where a helper provably has a caller the linter cannot see (in which case use an explicit `// nolint:unused` with a one-line justification).

## Scope
- **In:**
  - **First action — re-baseline.** Before touching any code, run `golangci-lint run ./...` from inside `services/agent-board/` and confirm the 34-finding / per-linter breakdown in `US004_tech_debt_lint_zero.md` still matches the current working tree. If it drifted, capture the new per-linter counts in this task's `## Notes` and proceed against the actual current state.
  - Triage every `unused` finding (9 expected). Most are in `services/agent-board/internal/handler/handler_test.go`; others may surface elsewhere — fix them wherever the linter reports them.
  - Default action: **prune** the dead helper. Some of these helpers are believed to be a self-inflicted artifact of the recent SSE-race fix (commit `18ca51e` "fix(handler): synchronise SSE test goroutines before reading recorder") and likely have no remaining caller.
  - Only suppress with `// nolint:unused // <one-line justification>` if the helper is genuinely callable from a build the linter does not see (e.g. an integration-only build tag) and removal would break that build.
- **Out:**
  - All other linter categories (`noctx`, `errorlint`, `errcheck`, `gocritic`, `gosec`, `revive`) — handled in subsequent US004 tasks.
  - Any production-code behaviour change. Pruning test helpers must not weaken what the remaining tests assert.
  - Edits to `.golangci.yml`.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/handler/handler_test.go` (primary — most `unused` findings live here)
- Any other `_test.go` file under `services/agent-board/` where the re-baseline reveals a stray `unused` finding (capture the actual list in `## Notes`)

## Why this is its own task (slicing rationale)
The `unused` cluster is triaged first because pruning dead helpers shrinks the diff for every subsequent task and reduces the chance that a later mechanical fix (e.g. `noctx` rewriting `httptest.NewRequest` call sites) silently re-introduces a reference to a helper that should have been deleted. Bundling this with the `noctx` work would mix judgement-heavy pruning with mechanical rewrites and obscure both during review.

## Test contract
US004 is a quality-refinement story; there is no `US004_be_unit_tests.md`. The contract is the verification commands in the story's "Acceptance criteria":
- `golangci-lint run ./...` inside `services/agent-board/` reports **zero `unused` findings** after this task (other categories may still be non-zero — they are addressed by the follow-up tasks in this story).
- `go test ./... -race` inside `services/agent-board/` is PASS (no new failures, no new race reports) relative to the baseline.

## Implementation notes
- Run the linter scoped to the category before and after, e.g. `golangci-lint run --enable-only=unused ./...`, to make the delta auditable.
- For each finding, decide prune-vs-suppress on the spot and document the choice in this task's `## Notes` (one line per finding: "pruned" or "suppressed: <reason>").
- If pruning a helper exposes the fact that a test was silently doing nothing meaningful (e.g. the helper was the only thing being called), flag that in `## Notes` — do NOT silently delete the assertion-free test; surface it so tech-lead and tester can decide.
- Re-run `go test ./... -race` after the prune pass specifically; pruning a `defer wg.Wait()`-style helper can change test timing.

## Definition of done
- All listed verification commands pass for this task's scope:
  - `golangci-lint run --enable-only=unused ./...` inside `services/agent-board/` reports zero findings.
  - `golangci-lint run ./...` inside `services/agent-board/` still runs (will report the remaining categories — that is expected, not a failure of this task).
  - `go test ./... -race` inside `services/agent-board/` is PASS.
- (Track: BE) `go vet ./...` clean inside `services/agent-board/`.
- Every remaining `// nolint:unused` (if any) has a one-line justification on the same or preceding line, names `unused` explicitly, and is not a blanket `// nolint` or file-level disable.
- No production-code behaviour change; only test helpers / dead code touched.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0, and `scripts/review/run-gate.sh cross` exits 0. The dev should run these locally before flipping to `in_review` — tech-lead will rerun them and reject on any failure.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Notes

### Re-baseline outcome — 2026-05-19

Re-baseline run from inside `services/agent-board/` on the current working tree (post-merge of PR #1 `ee98420`):

| Linter | Story baseline (2026-05-19) | Re-baseline actual | Drift |
|---|---|---|---|
| `noctx` | 11 | 11 | none |
| `unused` | 9 | **0** | **-9 (already zero)** |
| `gocritic` | 5 | 5 | none |
| `errcheck` | 4 | 4 | none |
| `errorlint` | 3 | 3 | none |
| `gosec` | 1 | 1 | none |
| `revive` | 1 | 1 | none |
| **Total** | **34** | **25** | **-9** |

**Key finding:** The `unused` cluster (9 findings) was already resolved by PR #1 (`ee98420 Merge pull request #1 from 9homme/tech-debt/lint-baseline-and-sse-race-fix`). The merge commit substantially rewrote `internal/handler/handler_test.go` (+103 lines net) as part of the SSE race-fix and the REQ003 implementation work, which eliminated all dead helpers that the `unused` linter had flagged. The `unused` category is confirmed zero by `golangci-lint run --enable-only=unused ./...` (0 issues, exit 0).

### Per-finding triage (UT-004)

Since `unused` was already zero on re-baseline, there were no findings to triage — no code changes required. This task is a no-op prune pass: the dead helpers were removed as part of the merged PR. Zero suppressions added (zero `// nolint:unused` directives in the codebase).

**Verification: `golangci-lint run --enable-only=unused ./...`** → `0 issues.` (exit 0).

### Suppression hygiene (UT-010)

`grep -rn "nolint" services/agent-board/` returns zero lines. No `// nolint` directives exist anywhere in the service. Hygiene requirement trivially satisfied.

### Verification outputs

All commands run from inside `services/agent-board/`:

1. `golangci-lint run --enable-only=unused ./...` → `0 issues.` exit 0. **PASS**
2. `golangci-lint run ./...` → 25 findings (errcheck: 4, errorlint: 3, gocritic: 5, gosec: 1, noctx: 11, revive: 1). Exit 1. **Expected — remaining categories belong to follow-up tasks. Count ≤ baseline (25 vs 34 expected, -9 from already-resolved `unused`).**
3. `go test ./... -race` → all packages `ok`, exit 0. **PASS**
4. `go vet ./...` → no output, exit 0. **PASS**
5. `scripts/review/run-gate.sh be services/agent-board` → REVIEW GATE: PASS (exit 0). **PASS**
6. `scripts/review/run-gate.sh cross` → REVIEW GATE: PASS (exit 0). **PASS**

### Files touched

None — this task is a verified no-op because the `unused` cluster was already eliminated by the previously merged PR. No source files were modified.

### Flagged for tech-lead and tester attention

- The `errcheck` linter now flags **two new sites** not in the original baseline: `internal/handler/message.go:83` and `internal/handler/message.go:105` (`json.Marshal` return unchecked). These were introduced by the REQ003 implementation in PR #1. They are `errcheck` findings, out of scope for this task but worth noting as a no-regression check. The errcheck count (4) still matches the story baseline despite these new sites, suggesting the `.golangci.yml` `errcheck` exclusion for `_test.go` files absorbs some older baseline findings that are no longer present. Follow-up task for `errcheck` should address all 4 current sites including these two in `message.go`.
- The `errorlint` baseline showed 3 findings against `sql.ErrNoRows`; the re-baseline shows 3 findings but 2 are now against `repo.ErrNotFound` in `user_story_tools.go` (lines 88, 112) and 1 against `sql.ErrNoRows` in `user_story_repo.go` (line 73). The REQ003 implementation added handler-layer `== repo.ErrNotFound` comparisons. Follow-up `errorlint` task should address all 3 current sites.

## Review log

### Review pass 1 — 2026-05-19 — verdict: approved

Verified no-op. The `unused` cluster (9 findings in the story baseline) was already eliminated by the previously merged PR #1 (`ee98420`, SSE-race fix + REQ003 implementation), which rewrote `internal/handler/handler_test.go`. The dev modified zero source files; the only commit on this task's branch (`cee0cef`) touches just this task file. The story's "Notes for the team" explicitly authorises capturing drift and proceeding against the current state, which is what the dev did.

**Re-confirmed verification (run by tech-lead, not trusting captured output):**
- `golangci-lint run --enable-only=unused ./...` inside `services/agent-board/` → `0 issues.` exit 0. PASS.
- `golangci-lint run ./...` inside `services/agent-board/` → 25 issues, exit 1. `unused` absent. PASS (expected — remaining categories belong to follow-up tasks).
- `go test ./... -race -count=1` inside `services/agent-board/` → all packages `ok`, exit 0, no DATA RACE. PASS.
- `go vet ./...` inside `services/agent-board/` → clean, exit 0. PASS.
- `scripts/review/run-gate.sh be services/agent-board` → `REVIEW GATE: PASS` (exit 0). PASS. (All checks PASS: gofmt -s, go vet, golangci-lint, go test, gosec, govulncheck.)
- `scripts/review/run-gate.sh cross` → `REVIEW GATE: PASS` (exit 0). PASS. (semgrep + gitleaks.)
- `git diff HEAD -- docs/requirements/REQ003_status_state_machine/architecture.md .golangci.yml` → empty. Architecture and lint config untouched.
- `grep -rn "nolint" services/agent-board/` → 0 lines. Suppression hygiene (UT-010) trivially satisfied.

**Working baseline for tasks 2–4** (re-confirmed per-linter breakdown of the remaining 25 findings):

| Linter | Count | Sites |
|---|---|---|
| `noctx` | 11 | `cmd/api-server/main.go:43`, `cmd/mcp-server/main.go:29` (db.Ping); `internal/handler/handler_test.go:96,144,179,222,256` + `internal/handler/project_handler_test.go:38,79,106,148` (httptest.NewRequest) |
| `gocritic` | 5 | `cmd/api-server/main.go:44`, `cmd/mcp-server/main.go:30` (exitAfterDefer, log.Fatalf after defer); `internal/repo/task_repo.go:159`, `internal/repo/user_story_repo.go:85,129` (sloppyReassign on err) |
| `errcheck` | 4 | `internal/handler/message.go:83,105` (json.Marshal unchecked — NEW from PR #1); `internal/repo/task_repo.go:97`, `internal/repo/user_story_repo.go:65` (`_ = tx.Rollback()` flagged because `check-blank: true`) |
| `errorlint` | 3 | `internal/handler/user_story_tools.go:88,112` (`== repo.ErrNotFound` — NEW from PR #1); `internal/repo/user_story_repo.go:73` (`== sql.ErrNoRows`) |
| `gosec` | 1 | `cmd/api-server/main.go:57` (G706, log-injection on port) |
| `revive` | 1 | `internal/handler/message.go:15` (var-naming: `sessionId` → `sessionID`) |
| **Total** | **25** | unused absent |

**Coverage acknowledgement:** UT-004 (zero `unused` findings) is satisfied. UT-001 (zero overall) and UT-002 (race PASS) are story-wide gates this task contributes to; both remain story-level until the chain completes. UT-010 (suppression hygiene) is vacuously satisfied with zero `nolint` directives.

**Scope-shift notes patched into downstream tasks** (rather than relayed via orchestrator, since editing task files is within tech-lead's scope and lowers handoff friction):
- `US004_be_errcheck_rollback_discard.md` — `## Notes` now lists the 4 actual sites including the two new `json.Marshal` sites in `internal/handler/message.go` (not just the original `*_repo.go` rollback sites). The handler-layer sites need a different per-site decision (handle vs explicit-discard) than the rollback pattern.
- `US004_be_mechanical_noctx_errorlint.md` — `## Notes` now lists the 3 actual `errorlint` sites including the two new `repo.ErrNotFound` comparisons in `internal/handler/user_story_tools.go` (note: under `internal/handler/`, NOT `internal/mcp/` as the orchestrator's brief stated). The 11 `noctx` sites are stable and enumerated.

**Verified no-op due to upstream drift** — acknowledged. No code or test changes by this task; closing as approved.
