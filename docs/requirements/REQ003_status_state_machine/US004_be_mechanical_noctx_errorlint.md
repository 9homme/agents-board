# US004/be_mechanical_noctx_errorlint

**Requirement:** REQ003
**Story:** US004
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** US004_be_unused_handler_test_triage.md
**Worked-by:** 
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

## Review log
(tech-lead appends here on each review pass)
