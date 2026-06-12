# US018 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live in `services/agent-board/cmd/api-server/main_test.go` (new file) and `services/agent-board/cmd/mcp-server/main_test.go` (new file), one package per binary. Do NOT add tests to `internal/` for this story — the pattern under test is in the `cmd/*/main.go` files.

Both test packages use `sqlmock` to simulate a blocking or failing DB driver — no real Postgres needed.

Architecture reference: §3 (signal-cancellable lifecycle context contract), §3.4 (`run()` shape), §3.5 (defer order), §3.7 (testability notes).

## Coverage matrix

| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| SIGTERM cancels lifecycle context (api-server) | unit | UT-US018-001 | services/agent-board / cmd/api-server | `signal.NotifyContext` cancellation on SIGTERM |
| SIGTERM cancels lifecycle context (mcp-server) | unit | UT-US018-002 | services/agent-board / cmd/mcp-server | `signal.NotifyContext` cancellation on SIGTERM |
| DB ping respects 5 s timeout (api-server) | integration | IT-US018-001 | services/agent-board / cmd/api-server | `db.PingContext` times out with `context.DeadlineExceeded` |
| DB ping respects 5 s timeout (mcp-server) | integration | IT-US018-002 | services/agent-board / cmd/mcp-server | `db.PingContext` times out with `context.DeadlineExceeded` |
| Defer order: cancel before stop before db.Close (api-server) | unit | UT-US018-003 | services/agent-board / cmd/api-server | deferred call order verified via call-order recorder |
| Defer order: cancel before stop before db.Close (mcp-server) | unit | UT-US018-004 | services/agent-board / cmd/mcp-server | deferred call order verified via call-order recorder |
| No `context.Background()` at the two named sites | unit | UT-US018-005 | services/agent-board / cmd/{api-server,mcp-server} | static grep assertion |
| Happy path: PingContext succeeds, no error returned (api-server) | integration | IT-US018-003 | services/agent-board / cmd/api-server | `run()` proceeds past ping with healthy DB mock |

## Unit tests

### UT-US018-001 — SIGTERM cancels lifecycle context (api-server)

- **Service:** `services/agent-board`
- **Package under test:** `cmd/api-server` (test file: `cmd/api-server/main_test.go`)
- **Function under test:** lifecycle context cancellation via `signal.NotifyContext`
- **Given:**
  - In a `t.Run` subtest, build a lifecycle context: `ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)`.
  - Start a goroutine that waits 50 ms then calls `syscall.Kill(syscall.Getpid(), syscall.SIGTERM)`.
- **When:** the goroutine fires SIGTERM
- **Then:**
  - `<-ctx.Done()` returns within 1 second (use `select` with a `time.After(1 * time.Second)` case that calls `t.Fatal("context not cancelled within timeout")`).
  - `ctx.Err()` is `context.Canceled`.
- **Cleanup:** `stop()` in a `defer` to reset signal handlers.
- **Edge cases:** repeat with SIGINT to confirm both signals are handled (can be a sub-subtest).
- **Architecture cite:** architecture §3.2 — signals handled: `os.Interrupt`, `syscall.SIGTERM`; §3.7 testability notes.

### UT-US018-002 — SIGTERM cancels lifecycle context (mcp-server)

- **Service:** `services/agent-board`
- **Package under test:** `cmd/mcp-server` (test file: `cmd/mcp-server/main_test.go`)
- **Function under test:** same `signal.NotifyContext` pattern as api-server
- **Given/When/Then:** identical to UT-US018-001 but instantiated in the `cmd/mcp-server` package.
- **Note:** these are two separate test files in two separate packages — do NOT share a test helper that would require a new `internal/lifecycle/` package (per architecture D-008, no shared helper).
- **Architecture cite:** architecture §3.4 — "The two files do NOT share a helper"; §2 US018 row for mcp-server.

### UT-US018-003 — Defer order: cancel → stop → db.Close (api-server)

- **Service:** `services/agent-board`
- **Package under test:** `cmd/api-server`
- **Function under test:** `run()` deferred cleanup order
- **Given:**
  - Create a call-order recorder: `var order []string` protected by a mutex.
  - Replace `defer stop()` and `defer cancel()` and `defer db.Close()` with equivalents that append `"stop"`, `"cancel"`, `"close"` to the recorder in their mock implementations.
  - Since `run()` is not directly injectable, this test may exercise the pattern at a lower level: assert that the declaration order in source is `db.Close` first, then `stop`, then `cancel` by either (a) using `go/ast` to parse the file in the test and verify the `defer` statement ordering, or (b) extracting the lifecycle setup into a testable helper (a private `func setupLifecycle(...) (context.Context, context.CancelFunc, context.CancelFunc)` that does not require an `internal/` package — it lives in `main_test.go` or an unexported function in the same package).
- **When:** the lifecycle setup function runs and its cleanup functions are deferred
- **Then:**
  - Cleanup call order recorded is `["cancel", "stop", "close"]` (Go LIFO: last declared runs first; declaration order must be close → stop → cancel).
- **Notes:** If the AST-parse approach is used, the test does not require running `run()` at all — it is a static correctness check. Document the chosen approach in the test file.
- **Architecture cite:** architecture §3.5 — defer order: `cancel` (innermost) → `stop` → `db.Close`.

### UT-US018-004 — Defer order: cancel → stop → db.Close (mcp-server)

- **Service:** `services/agent-board`
- **Package under test:** `cmd/mcp-server`
- **Function under test:** `run()` deferred cleanup order
- **Given/When/Then:** identical pattern to UT-US018-003, applied to `cmd/mcp-server/main_test.go`.
- **Architecture cite:** architecture §3.5; §2 US018 row for mcp-server.

### UT-US018-005 — No `context.Background()` at the two named production sites

- **Service:** `services/agent-board`
- **Package under test:** `cmd/api-server`, `cmd/mcp-server`
- **Function under test:** static source assertion (grep-style)
- **Given:** the source files `cmd/api-server/main.go` and `cmd/mcp-server/main.go`
- **When:** file contents are read in the test (use `os.ReadFile`)
- **Then:**
  - Neither file contains the substring `db.PingContext(context.Background())`.
  - Each file DOES contain `db.PingContext(pingCtx)` (or equivalent derived context variable name — match the variable name used in the `run()` implementation).
- **Notes:** This is a regression guard. If a future change accidentally reverts the context, this test catches it without needing a runtime DB.
- **Architecture cite:** US018 AC "Scenario: no `context.Background()` remains in production code at the two named sites".

## Integration tests

### IT-US018-001 — DB ping times out with `context.DeadlineExceeded` (api-server)

- **Service:** `services/agent-board`
- **Package under test:** `cmd/api-server` (test file: `cmd/api-server/main_test.go`)
- **Boundary:** `db.PingContext(pingCtx)` against a `sqlmock`-backed driver that blocks until context cancellation
- **Setup:**
  - Open a `sqlmock` DB: `db, mock, err := sqlmock.New()`.
  - Configure the mock's `Ping` to block: implement a custom `driver.Pinger` that blocks on `<-ctx.Done()` then returns `ctx.Err()`.
    - OR: use `mock.ExpectPing().WillDelayFor(10 * time.Second)` if the sqlmock version supports it; the timeout is 5 s so a 10 s delay will always trigger the deadline.
  - Build a `pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)` and `defer cancel()`.
- **When:** `db.PingContext(pingCtx)` is called
- **Then:**
  - The call returns within 6 seconds (a 1-second grace above the 5 s deadline is acceptable for test stability).
  - The returned error satisfies `errors.Is(err, context.DeadlineExceeded)` — either directly or via `fmt.Errorf("db ping failed: %w", err)` wrapping.
- **Notes:** The architecture hard-codes 5 s (D-013). The test exercises the exact timeout value; use 5*time.Second as the timeout and assert the round-trip completes well within 6 s.
- **Architecture cite:** architecture §3.3 — 5 s timeout; §3.4 `run()` shape; §3.7.

### IT-US018-002 — DB ping times out with `context.DeadlineExceeded` (mcp-server)

- **Service:** `services/agent-board`
- **Package under test:** `cmd/mcp-server`
- **Boundary:** same as IT-US018-001
- **Setup/When/Then:** identical to IT-US018-001 but instantiated in `cmd/mcp-server/main_test.go`.
- **Architecture cite:** architecture §2 US018 row (mcp-server line at `main.go:37`).

### IT-US018-003 — Happy path: PingContext succeeds, `run()` proceeds (api-server)

- **Service:** `services/agent-board`
- **Package under test:** `cmd/api-server`
- **Boundary:** `db.PingContext(pingCtx)` against a `sqlmock` that returns immediately with no error
- **Setup:**
  - Open a `sqlmock` DB.
  - `mock.ExpectPing()` (returns nil immediately).
  - Build `pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)` and `defer cancel()`.
- **When:** `db.PingContext(pingCtx)` is called
- **Then:**
  - No error returned.
  - `mock.ExpectationsWereMet()` returns nil.
  - The context has NOT expired (assertion: `pingCtx.Err() == nil` immediately after the successful ping).
- **Architecture cite:** US018 AC "Scenario: happy path unchanged"; architecture §3.4.
