# US004/be_mechanical_noctx_errorlint

**Requirement:** REQ003
**Story:** US004
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** US004_be_unused_handler_test_triage.md
**Worked-by:** be-dev (US004 noctx+errorlint)
**Implements:** US004 acceptance criterion "Specific finding categories are resolved correctly, not papered over" — drives `errorlint` (3) and `noctx` (11) to zero. Total: 14 findings.

## Goal
Apply the two mechanical lint-fix categories: rewrite `err == sql.ErrNoRows` comparisons as `errors.Is(err, sql.ErrNoRows)` (`errorlint`, 3 findings), and add explicit contexts to HTTP test requests and DB pings (`noctx`, 11 findings).

## Scope
- **In:**
  - **`errorlint` (3):** Rewrite every `err == sql.ErrNoRows` (and any other bare error-value comparison the linter flags in the same category) as `errors.Is(err, sql.ErrNoRows)`. Expected sites: repo layer query paths — `services/agent-board/internal/repo/user_story_repo.go`, `services/agent-board/internal/repo/task_repo.go`, and any other repo file the linter flags. Confirm against the live linter run, do not rely on the count alone.
  - **`noctx` (11):**
    - **Test requests:** Replace `httptest.NewRequest(method, url, body)` with `httptest.NewRequestWithContext(ctx, method, url, body)` (use `context.Background()` or the test's existing context). Alternatively use `.WithContext(ctx)` if `NewRequestWithContext` does not fit a particular call site. Expected concentration: `services/agent-board/internal/handler/*_test.go`.
    - **DB ping:** Replace any `db.Ping()` with `db.PingContext(ctx)` carrying the appropriate context (the surrounding HTTP handler context, the long-lived server context in `cmd/*/main.go`, or `context.Background()` if there is no caller-supplied context). Expected sites: `services/agent-board/cmd/api-server/main.go`, `services/agent-board/cmd/mcp-server/main.go`, and any health-check path in `internal/handler/`. Confirm against the live linter run.
- **Out:**
  - All other linter categories (`unused`, `errcheck`, `gocritic`, `gosec`, `revive`) — handled in sibling tasks.
  - Refactoring `httptest` helpers for ergonomics beyond what the lint fix requires.
  - Plumbing a new context type through unrelated call chains. The fix is "add a context to this call", not "redesign context propagation".
  - Edits to `.golangci.yml`.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/repo/user_story_repo.go`
- `services/agent-board/internal/repo/task_repo.go`
- Any other `services/agent-board/internal/repo/*.go` file the live linter flags for `errorlint`
- `services/agent-board/internal/handler/handler_test.go`
- Any other `services/agent-board/internal/handler/*_test.go` file the live linter flags for `noctx` (`task_tools_test.go`, `user_story_tools_test.go`, `project_tools_test.go`, `audit_tools_test.go`, `document_tools_test.go`, `project_handler_test.go` are all candidates)
- `services/agent-board/cmd/api-server/main.go`
- `services/agent-board/cmd/mcp-server/main.go`

(The dev should narrow this list to the exact files the live linter run reports and capture the final list in `## Notes`.)

## Why this is its own task (slicing rationale)
Bundled into one task because both categories are pure mechanical rewrites with effectively no judgement: `==` → `errors.Is`, and `NewRequest` → `NewRequestWithContext`. Splitting them would double the review overhead with no parallelism benefit (they overlap on `handler_test.go` after the prune task lands, so they cannot run in parallel anyway). They follow the `unused` triage task because pruning dead helpers first means `noctx` only rewrites call sites that genuinely matter.

## Test contract
US004 is a quality-refinement story; there is no `US004_be_unit_tests.md`. The contract is the verification commands in the story's "Acceptance criteria":
- `golangci-lint run ./...` inside `services/agent-board/` reports **zero `errorlint` and zero `noctx` findings** after this task. Other categories (`errcheck`, `gocritic`, `gosec`, `revive`) may still be non-zero — they are addressed by the follow-up tasks.
- `go test ./... -race` inside `services/agent-board/` is PASS. Re-run this specifically after the `noctx` batch — the story's "Notes for the team" calls out that adding/changing test contexts can occasionally alter timeout behaviour under `-race`.

## Implementation notes
- Use `golangci-lint run --enable-only=errorlint ./...` and `--enable-only=noctx ./...` to audit each category independently before and after.
- For `httptest.NewRequestWithContext`, prefer the test's existing context if one is in scope; otherwise `context.Background()` is acceptable for a unit/integration test. Do not invent a `context.WithTimeout` to "be safe" — that changes test semantics.
- For `db.PingContext`, in `cmd/*/main.go` the natural choice is the server-lifetime context the rest of the bootstrap already constructs (if any), else `context.Background()`. Do not introduce a new package-level context.
- After applying `noctx` fixes, double-check that any test that was relying on `httptest.NewRequest`'s zero-context behaviour (rare) still asserts what it intended.

## Definition of done
- All listed verification commands pass for this task's scope:
  - `golangci-lint run --enable-only=errorlint ./...` inside `services/agent-board/` reports zero findings.
  - `golangci-lint run --enable-only=noctx ./...` inside `services/agent-board/` reports zero findings.
  - `golangci-lint run ./...` inside `services/agent-board/` still runs (remaining categories from sibling tasks are expected, not a failure of this task).
  - `go test ./... -race` inside `services/agent-board/` is PASS — explicitly re-run after the `noctx` batch.
- (Track: BE) `go vet ./...` clean inside `services/agent-board/`.
- Any `// nolint:errorlint` or `// nolint:noctx` (expected to be none) has a one-line justification, names the linter explicitly, and is not a blanket disable.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Notes

### Scope-shift note from US004_be_unused_handler_test_triage review (2026-05-19, tech-lead)
Re-baseline against the current working tree (post PR #1 `ee98420`) confirms `errorlint` is still 3 findings total and `noctx` is still 11 findings total, but the `errorlint` site map has shifted relative to the story background. The current 3 `errorlint` sites are:
- `services/agent-board/internal/handler/user_story_tools.go:88` — `if err == repo.ErrNotFound` (NEW; introduced by REQ003 handler-layer comparisons in PR #1). Note: lives under `internal/handler/`, NOT `internal/mcp/`.
- `services/agent-board/internal/handler/user_story_tools.go:112` — `if err == repo.ErrNotFound` (NEW; same file)
- `services/agent-board/internal/repo/user_story_repo.go:73` — `if err == sql.ErrNoRows` (the original story-background site that survives)

Two of the three are `repo.ErrNotFound` comparisons in the handler layer, not `sql.ErrNoRows` comparisons in the repo layer. The rewrite is the same shape (`==` → `errors.Is`) but the file list in the task's `## Files touched` section under-counts: add `services/agent-board/internal/handler/user_story_tools.go` when picking this task up. The `task_repo.go` entry can be removed unless the live re-run flags new sites there.

The 11 `noctx` sites are stable and as listed in the story background:
- `services/agent-board/cmd/api-server/main.go:43` — `db.Ping()`
- `services/agent-board/cmd/mcp-server/main.go:29` — `db.Ping()`
- `services/agent-board/internal/handler/handler_test.go:96,144,179,222,256` — `httptest.NewRequest(...)` (5 sites)
- `services/agent-board/internal/handler/project_handler_test.go:38,79,106,148` — `httptest.NewRequest(...)` (4 sites)

(dev appends: final file list from live linter run, any call sites that needed `.WithContext` instead of `NewRequestWithContext`, any test-timing surprises under `-race`)

### Dev implementation notes (2026-05-19, be-dev)

**Pre-fix baseline (live linter re-run):** confirmed 3 errorlint + 11 noctx = 14 findings as documented.

**errorlint fixes (3/3) — files changed:**
- `services/agent-board/internal/handler/user_story_tools.go`: added `"errors"` import; rewrote two `err == repo.ErrNotFound` comparisons at lines 88 and 112 to `errors.Is(err, repo.ErrNotFound)`. Mechanical equivalence — sentinel error comparison, no behaviour change.
- `services/agent-board/internal/repo/user_story_repo.go`: added `"errors"` import; rewrote `err == sql.ErrNoRows` at line 73 (inside `UpdateUserStoryStatus`) to `errors.Is(err, sql.ErrNoRows)`. Mechanical equivalence — same sentinel path, correctly handles wrapped errors.

**noctx fixes (11/11) — files changed:**
- `services/agent-board/cmd/api-server/main.go`: added `"context"` import; replaced `db.Ping()` with `db.PingContext(context.Background())`. Context choice: `context.Background()` — there is no server-lifetime context constructed at this point in the bootstrap sequence; process-startup pings do not require a cancel/deadline.
- `services/agent-board/cmd/mcp-server/main.go`: same change, same rationale.
- `services/agent-board/internal/handler/handler_test.go`: replaced all 5 `httptest.NewRequest(...)` calls with `httptest.NewRequestWithContext(...)`. Two of them (lines 96, 222) previously used the `.WithContext(ctx)` chained form — those were converted to `httptest.NewRequestWithContext(ctx, ...)` using the already-scoped `ctx`. The other three (lines 144, 179, 256) had no ambient context — used `context.Background()`. Context choice rationale: these are pure MCP message handler tests with no timeout semantics; `context.Background()` is the conventional choice for unit tests that do not need cancellation.
- `services/agent-board/internal/handler/project_handler_test.go`: replaced all 4 `httptest.NewRequest(...)` calls with `httptest.NewRequestWithContext(context.Background(), ...)`. Context choice: `context.Background()` — none of the tests had an ambient context, and the handler is stateless; no timeout is needed.

**No `.WithContext` chaining needed** — all 9 httptest call sites converted cleanly to `httptest.NewRequestWithContext(...)`.

**Post-fix verification:**
- `golangci-lint run --enable-only=errorlint ./...` → 0 issues.
- `golangci-lint run --enable-only=noctx ./...` → 0 issues.
- `golangci-lint run ./...` → 11 remaining findings (errcheck 4 + gocritic 5 + gosec 1 + revive 1). Re-baseline confirmed: started at 25, ended at 11.
- `go vet ./...` → clean (no output).
- `go test ./... -race -count=1` run 1 → PASS (domain ok, handler ok, mcp ok, repo ok).
- `go test ./... -race -count=1` run 2 → PASS (domain ok, handler ok, mcp ok, repo ok). No DATA RACE reports in either run.
- `scripts/review/run-gate.sh be services/agent-board` → REVIEW GATE: PASS.
- `scripts/review/run-gate.sh cross` → REVIEW GATE: PASS.

**No nolint directives added.** All 14 findings have clean idiomatic fixes.

## Review log
(tech-lead appends here on each review pass)

### Review pass 1 — 2026-05-19 — verdict: approved

**Verification (re-run by tech-lead, not trusting dev report):**
- `golangci-lint run --enable-only=errorlint ./...` → `0 issues.` exit 0.
- `golangci-lint run --enable-only=noctx ./...` → `0 issues.` exit 0.
- `golangci-lint run ./...` → `11 issues:` exit 1. Per-linter breakdown: `errcheck: 4, gocritic: 5, gosec: 1, revive: 1`. `unused`, `errorlint`, `noctx` all absent — no regression, no category drift.
- `go vet ./...` → clean (exit 0, no output).
- `go test ./... -race -count=1` run 1 → `ok` for domain/handler/mcp/repo; cmd packages have no tests; no `FAIL`, no `DATA RACE`. Exit 0.
- `go test ./... -race -count=1` run 2 (back-to-back) → same result. Exit 0. `noctx` rewrites did not perturb timing under `-race`, as the story flagged might happen.
- `grep -rn "nolint" services/agent-board/` → exit 1 (no matches). Module-wide `// nolint` count: 0. UT-010 hygiene contract still satisfied.
- `scripts/review/run-gate.sh be services/agent-board` → `REVIEW GATE: PASS` exit 0 (gofmt, go vet, golangci-lint --no-config, go test, gosec, govulncheck all PASS).
- `scripts/review/run-gate.sh cross` → `REVIEW GATE: PASS` exit 0 (semgrep, gitleaks PASS).

**Diff-quality observations:**
- Worktree commit `1a383f1` modified exactly 7 files: 6 source + 1 task file. No edits to `architecture.md`, `.golangci.yml`, the test specs (`US004_be_unit_tests.md` et al.), or sibling task files. Scope: tight.
- All 9 `httptest.NewRequestWithContext` rewrites are constructor-only — assertions, recorder, downstream handler calls untouched. Verified by reading diffs in `handler_test.go` and `project_handler_test.go`.
- Context-choice quality: spot-on. `handler_test.go:96` (`TestGetSSEEndpoint`) and `:222` (`TestITFullHandshake`) correctly reuse the in-scope `ctx` from `context.WithCancel(context.Background())` (cancellation propagation preserved); the other 7 sites correctly use `context.Background()` for tests with no cancellation semantics. `cmd/*/main.go` both use `context.Background()` without a timeout — no behaviour change to startup ping.
- `errors.Is` semantics preserved: `repo.ErrNotFound` and `sql.ErrNoRows` are sentinels reached via direct equality in both repo and handler paths; `errors.Is` is a superset that handles wrapping too. No test relied on identity-only comparison (verified by reading the affected handler code path).
- `exitAfterDefer` gocritic findings now appearing on `cmd/api-server/main.go:45` and `cmd/mcp-server/main.go:31` are pre-existing (lines 44/30 at parent `1a383f1^`, shifted by 1 due to the added `"context"` import). Same 2 findings, same 5-count `gocritic` baseline — no new debt introduced.

**Working baseline for the next task** (`US004_be_tail_gocritic_gosec_revive.md`): **11 findings remaining** — errcheck 4 (`message.go:83`, `message.go:105`, `task_repo.go:97`, `user_story_repo.go:66`), gocritic 5 (`api-server/main.go:45` exitAfterDefer, `mcp-server/main.go:31` exitAfterDefer, `task_repo.go:159` sloppyReassign, `user_story_repo.go:86` sloppyReassign, `user_story_repo.go:130` sloppyReassign), gosec 1 (`api-server/main.go:58` G706 log injection), revive 1 (`message.go:15` var-naming sessionId → sessionID). Note: 2 of the 4 errcheck findings (`message.go:83,105` json.Marshal) are NEW vs. the original story background which listed all 4 errcheck as `tx.Rollback` — the tail task will need to handle both shapes; the `tx.Rollback` form-A/form-B guidance in UT-006 still applies for the repo-layer pair.

Streak: 2 approved across this story (US004 unused triage + this one). Clean.
