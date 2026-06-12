# US018 — Replace unbounded `context.Background()` at DB-ping sites with signal-cancellable lifecycle context

**Requirement:** REQ005 — quality hardening retrospective
**Status:** draft

## Story
As an **operator running `api-server` or `mcp-server`**, I want the **DB ping at startup to honour a timeout AND to cancel on SIGTERM**, so that a stuck network round-trip during boot does not hang the process forever and that a `kill -TERM` during boot terminates cleanly instead of waiting on an unbounded context.

## Acceptance criteria

- **Scenario: DB ping respects a bounded timeout**
  - Given `services/agent-board/cmd/api-server/main.go` is started
  - When `db.PingContext(...)` is invoked
  - Then the context passed to `PingContext` is derived from `context.WithTimeout` with a deadline of 5 seconds (default; configurable via env var `DB_PING_TIMEOUT_SECONDS` is optional — see "Notes for the team")
  - And on timeout, `run()` returns an error that wraps the timeout error (e.g. `fmt.Errorf("db ping timed out: %w", err)`)
  - And the corresponding `defer cancel()` is called before `run()` returns

- **Scenario: DB ping cancels on SIGTERM / SIGINT during startup**
  - Given `api-server` is in the middle of `db.PingContext(...)` (simulated by pointing `DATABASE_URL` at a host that accepts TCP but never completes the handshake, OR by using a controllable mock)
  - When the process receives `SIGTERM` (or `SIGINT`)
  - Then the ping context is cancelled by the signal handler
  - And `PingContext` returns `context.Canceled` (or a wrap thereof)
  - And `run()` returns the cancellation error
  - And the process exits within 1 second of the signal (no orphaned ping goroutine, no zombie DB connection)

- **Scenario: same fix applied to `mcp-server`**
  - Given `services/agent-board/cmd/mcp-server/main.go:37` also calls `db.PingContext(context.Background())`
  - When the same pattern is applied to `mcp-server`
  - Then both servers' AC above hold for `mcp-server` identically

- **Scenario: happy path unchanged**
  - Given a healthy Postgres reachable at `DATABASE_URL`
  - When `api-server` (or `mcp-server`) starts
  - Then `PingContext` succeeds well within the 5-second timeout
  - And the server proceeds to register handlers and `e.Start(...)` / equivalent as today
  - And there is no observable behaviour change in the happy path beyond the new lifecycle wiring

- **Scenario: lifecycle context is the signal-cancellable parent**
  - Given the signal handler is wired via `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` (or equivalent)
  - When `run()` later spawns work that needs the same lifecycle (currently the only DB-touching site is `PingContext`)
  - Then `PingContext`'s context is the signal-cancellable context with a `WithTimeout` derived from it
  - And future work in `run()` can also derive from the same lifecycle parent without rewiring

- **Scenario: no `context.Background()` remains in production code at the two named sites**
  - Given a `grep -n 'context.Background()' services/agent-board/cmd/api-server/main.go services/agent-board/cmd/mcp-server/main.go`
  - When the story is complete
  - Then those two files contain zero `context.Background()` calls
  - And the audit-flagged line `main.go:52` and `mcp-server/main.go:37` are replaced with the signal-cancellable + timeout-bounded variant

## UI / UX flow expectations

**No UI:** operator-facing reliability change. The "flow" is: operator runs `./api-server` (or `./mcp-server`); during boot the DB ping either succeeds, times out cleanly, or aborts cleanly on signal. Visible only via process exit code and log line.

## Out of scope
- **Sweeping `context.Background()` in tests.** All matches in `*_test.go` files (every `internal/handler/*_test.go` and `internal/mcp/*_test.go` match in the audit grep) are correct — tests legitimately use `context.Background()` for sqlmock and httptest setup. Per D-003 this story does NOT touch test files.
- **Graceful HTTP shutdown.** Wiring SIGTERM to `e.Shutdown(ctx)` for in-flight requests is a separate concern. This story only fixes the boot-time ping. If the architect wants to bundle them, raise via `ARCHITECTURE_GAP_FOUND`.
- **Refactoring `main()` / `run()` structure.** Add the minimum signal-context wiring; do not rewrite the file.
- **Adding a new env var `DB_PING_TIMEOUT_SECONDS`.** Optional — see Notes. Architect / tech-lead may decide hard-coded 5s is enough.

## Dependencies
- None. Independent of every other US in REQ005.

## Notes for the team

- **Two known sites** (verified via grep on 2026-05-30 HEAD):
  - `services/agent-board/cmd/api-server/main.go:52` — `if err := db.PingContext(context.Background()); err != nil {`
  - `services/agent-board/cmd/mcp-server/main.go:37` — `if err := db.PingContext(context.Background()); err != nil {`
  - All other `context.Background()` grep hits in `services/` are in `*_test.go` and are out of scope per D-003.
- **Suggested skeleton** (illustrative, not prescriptive — architect / tech-lead designs the final shape):
  ```go
  ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
  defer stop()

  pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  if err := db.PingContext(pingCtx); err != nil {
      return fmt.Errorf("db ping failed: %w", err)
  }
  ```
- **Env var question** — `DB_PING_TIMEOUT_SECONDS` is OPTIONAL. If added, parse with `strconv.Atoi`, fall back to 5 on parse failure or unset. If not added, hard-code 5s with a `// TODO(REQ005): make configurable if ops needs it` comment.
- **Test approach** (tester will detail):
  - Unit-test extraction of the timeout/signal-handling logic into a helper that can be tested without spinning up a real DB.
  - Or: integration-style test using `sqlmock` + a `driver.Conn` whose `Ping(ctx)` blocks until `<-ctx.Done()`.
  - Asserting SIGTERM cancellation in-process is fiddly but doable via `syscall.Kill(syscall.Getpid(), syscall.SIGTERM)` from a goroutine.
- **Architecture note:** if the system-architect decides to factor the signal-cancellable context construction into a shared helper (e.g. `internal/lifecycle/context.go`), that's a valid choice — both `cmd/api-server` and `cmd/mcp-server` benefit. Not required by AC; just an option.

## Sign-off log
(po-ba appends here on each sign-off pass)
