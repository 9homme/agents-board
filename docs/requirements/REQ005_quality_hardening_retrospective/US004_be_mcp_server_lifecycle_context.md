# US004 — Signal-cancellable lifecycle context for `mcp-server` DB ping

**Story:** US004 — Replace unbounded `context.Background()` at DB-ping sites with signal-cancellable lifecycle context
**Requirement:** REQ005
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Implements:** Scenario: same fix applied to `mcp-server`, Scenario: DB ping respects a bounded timeout (mcp-server), Scenario: DB ping cancels on SIGTERM / SIGINT during startup (mcp-server), Scenario: happy path unchanged (mcp-server), Scenario: lifecycle context is the signal-cancellable parent (mcp-server), Scenario: no `context.Background()` remains in production code at the two named sites (mcp-server portion)
**Blocked by:** none
**Worked-by:** _(none)_

## Goal

Apply the same signal-cancellable lifecycle context pattern as the api-server task to `services/agent-board/cmd/mcp-server/main.go:37`. After this task, `mcp-server` boot honours SIGTERM/SIGINT during the ping, bounds the ping at 5 seconds, and cleans up deferred resources per architecture §3.5.

## Scope

- **In:** Edit `services/agent-board/cmd/mcp-server/main.go` `run()` body around line 37 to add `ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` + `defer stop()` + `pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)` + `defer cancel()` + pass `pingCtx` to `db.PingContext`. Add necessary stdlib imports. Add the `// TODO(REQ005): make configurable if ops needs it` comment per §3.3. Add a new `cmd/mcp-server/main_test.go` with the signal-handler and timeout-branch tests per §3.7.
- **Out:** `api-server` changes (separate task `US004_be_api_server_lifecycle_context.md`); shared `internal/lifecycle/` helper (rejected by D-008); graceful HTTP shutdown (D-009); env var `DB_PING_TIMEOUT_SECONDS` (D-013); test-file `context.Background()` sweeps (D-003).

## Files touched (estimated, exclusive)

- `services/agent-board/cmd/mcp-server/main.go`
- `services/agent-board/cmd/mcp-server/main_test.go` (new)

Independent of `US004_be_api_server_lifecycle_context.md` — different files entirely. Per D-008, NO shared helper, so no `internal/` overlap. Both US004 tasks parallelise cleanly.

## Test contract

The dev must make these tests pass (from `US004_be_unit_tests.md`, IDs assigned by tester):
- Mirror of the signal-handler subtest used for api-server, scoped to `cmd/mcp-server`.
- Mirror of the timeout-branch subtest (sqlmock blocking `Ping`), scoped to `cmd/mcp-server`.
- Happy-path subtest for `mcp-server`.
- Defer-order assertion mirroring api-server's.

If tester surfaces new test IDs beyond these, the dev writes them and flags the addition back to tester.

## Implementation notes

- Architecture §3.4 is the verbatim pattern — copy it. Do not invent variants. Two services intentionally duplicate (D-008).
- Defer declaration order per §3.5: existing `defer db.Close()` stays first; `defer stop()` after signal-context setup; `defer cancel()` after the `WithTimeout` setup. LIFO at exit becomes `cancel → stop → db.Close`.
- Wrap ping errors as `fmt.Errorf("db ping failed: %w", err)` (matching api-server's wrap text exactly, since the architecture treats them as one logical pattern).
- No change to MCP transport setup, `/sse` / `/message` handler registration, or other mcp-specific wiring.
- Imports to add: `os/signal`, `syscall`, `time` — confirm none already imported before duplicating.
- TDG skill (.claude/skills/tdg/SKILL.md) MUST be invoked at each TDD phase per be-dev workflow.

## Definition of done

- All listed tests green.
- `cd services/agent-board && go vet ./... && go test ./...` clean.
- `cd services/agent-board && go test -coverprofile=/tmp/cov.out ./... && go tool cover -func=/tmp/cov.out` — `cmd/mcp-server/main.go` clears ≥ 80 % line coverage OR a `## Coverage exemption` block here justifies the gap (same caveat as api-server task — boot wiring is partially exemptable; new ping/signal branches must be covered).
- `grep -n 'context.Background()' services/agent-board/cmd/mcp-server/main.go` returns zero hits.
- `scripts/review/run-gate.sh be services/agent-board` exits with `REVIEW GATE: PASS`.
- `scripts/review/run-gate.sh cross` exits with `REVIEW GATE: PASS`.
- Code matches architecture §3 contract end-to-end; no divergence from api-server's source pattern other than file location.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Review log

(tech-lead appends here on each review pass)
