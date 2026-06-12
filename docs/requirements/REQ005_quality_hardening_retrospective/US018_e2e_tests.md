# US018 — E2E test specification

**Owner:** tester.

## Why e2e does not apply

US018 fixes signal-handling and DB-ping context in `cmd/api-server/main.go` and `cmd/mcp-server/main.go`. The changes are process-lifecycle concerns:

- **SIGTERM cancellation** is a process-signal event. Robot Framework's Browser and RequestsLibrary have no mechanism to send POSIX signals to a running container or to observe whether a goroutine's context was cancelled by a signal. The integration tests (IT-US018-001/002 using `syscall.Kill` from within the process) are the correct layer.
- **DB-ping timeout** is a boot-time concern. By the time the docker-compose stack (US022) is healthy (postgres and api-server both report healthy), the ping has already succeeded. There is no e2e-observable difference between a 5 s bounded ping and an unbounded ping for the happy path.
- **No-`context.Background()` regression guard** is a source-code assertion, not a runtime observable.

Verifying signal lifecycle correctly requires process-level control that is out of scope for the Robot / Browser / Requests layer. The full correctness contract is covered by UT-US018-001 through UT-US018-005 and IT-US018-001 through IT-US018-003.

**Verdict: No e2e scenarios. Signal lifecycle is a process-level concern; integration tests under IT-US018-* exercise it.**
