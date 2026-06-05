# US009 — Backfill `internal/mcp/server.go` ToolRegistry + Session message tests

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** draft

## Story
As a **future contributor changing `services/agent-board/internal/mcp/server.go`**, I want **every method on `ToolRegistry` (`NewToolRegistry`, `RegisterTool`, `GetTool`, `ListTools`), the `Session.QueueMessage` queue-full branch, the `Session.ReceiveMessage` context-cancel branch, and `SessionManager.RemoveSession` to be covered by unit tests**, so that a regression in MCP plumbing (e.g. dropping the channel-based backpressure, returning `(nil, nil)` from `GetTool` on a missing tool, or leaking sessions after `RemoveSession`) fails CI immediately.

## Acceptance criteria

- **Scenario: `server_test.go` (or `tool_registry_test.go` + `session_test.go` extensions) gains the following test functions (verbatim names)**
  - Given the existing `services/agent-board/internal/mcp/session_test.go` (the only existing mcp test file per `ls` output)
  - When the story is complete
  - Then the following new test functions exist (placement: tester chooses one file or splits — names authoritative regardless of file):
    1. `TestNewToolRegistry_ReturnsEmptyRegistry` — covers `NewToolRegistry` 0% by asserting `ListTools()` returns an empty slice.
    2. `TestToolRegistry_RegisterTool_AddsHandler` — covers `RegisterTool` 0% by registering a noop handler and asserting `GetTool` returns `(handler, true)`.
    3. `TestToolRegistry_RegisterTool_OverwritesPriorHandler` — registering a second handler under the same name replaces the first (latest wins).
    4. `TestToolRegistry_GetTool_UnknownNameReturnsFalse` — covers `GetTool` 0% by querying an unregistered name and asserting `(nil, false)`.
    5. `TestToolRegistry_GetTool_KnownNameReturnsHandler` — already implicitly covered by UT-2 but call out explicitly: returned handler, when invoked, produces the registered behaviour.
    6. `TestToolRegistry_ListTools_ReturnsAllRegisteredNames` — covers `ListTools` 0% by registering three names and asserting the returned slice has length 3 and contains each name (unordered membership check — current implementation does NOT sort despite the doc-comment claim; flag in test report as a discrepancy if so).
    7. `TestToolRegistry_ListTools_EmptyAfterNoRegistrations` — fresh registry → empty slice.
    8. `TestToolRegistry_ConcurrentRegisterAndGet` — spawns N goroutines that interleave `RegisterTool` + `GetTool` + `ListTools`; the test passes if no data race is reported under `go test -race`.
    9. `TestSession_QueueMessage_HappyPath` — buffered channel has capacity; `QueueMessage` returns nil and `ReceiveMessage` returns the message.
    10. `TestSession_QueueMessage_FullReturnsError` — covers `QueueMessage` 66.7% by pre-filling the channel to capacity (100), then asserting next `QueueMessage` returns `errors.New("message queue full")`.
    11. `TestSession_ReceiveMessage_HappyPath` — `QueueMessage(b)` then `ReceiveMessage(ctx)` returns `(b, nil)`.
    12. `TestSession_ReceiveMessage_ContextCancelled` — covers `ReceiveMessage` 66.7% by cancelling the context BEFORE / DURING the call; asserts return is `(nil, context.Canceled)` (or `context.DeadlineExceeded` if using `WithTimeout`).
    13. `TestSessionManager_RemoveSession_RemovesSession` — covers `RemoveSession` 0% by creating a session via `CreateSession`, calling `RemoveSession(session.ID)`, then asserting `GetSession(session.ID)` returns `(nil, false)`.
    14. `TestSessionManager_RemoveSession_UnknownIDIsNoop` — calling `RemoveSession("nonexistent")` does not panic and leaves other sessions intact.
    15. `TestSessionManager_RemoveSession_ConcurrentSafe` — N goroutines creating + removing sessions; passes under `-race`.

- **Scenario: each new test exercises the specific uncovered branch**
  - Given the relevant struct (`*ToolRegistry` or `*Session` or `*SessionManager`) constructed directly
  - When the method under test is invoked per the test name
  - Then assertions:
    - For `_HappyPath`: positive return values match expectations
    - For `_FullReturnsError`: returned error has message `"message queue full"`
    - For `_ContextCancelled`: returned error matches `context.Canceled` (or `context.DeadlineExceeded` per implementation choice) via `errors.Is`
    - For `_UnknownNameReturnsFalse` / `_UnknownIDIsNoop`: bool flag is `false`, no panic
    - For `_Concurrent*`: tests must be runnable under `go test -race ./internal/mcp` with no race reports

- **Scenario: package-level coverage hits ≥95%**
  - Given `cd services/agent-board && go test ./internal/mcp -coverprofile=/tmp/mcp.out`
  - When `go tool cover -func=/tmp/mcp.out` is inspected
  - Then `internal/mcp` package shows **≥95% statement coverage**
  - And the only uncovered lines (if any) are documented in the test report under OQ-4

- **Scenario: existing tests still pass and behaviour is unchanged**
  - Given `server.go` is **NOT** modified by this story
  - When `cd services/agent-board && go test ./...` runs
  - Then all pre-existing tests pass
  - And all new tests pass
  - And `golangci-lint run ./...` is clean
  - And `cd services/agent-board && go test ./internal/mcp -race` passes (covers the `_Concurrent*` tests)

- **Scenario: no production-code changes**
  - Given `git diff` of the story's commits
  - When inspected
  - Then **only** test files under `services/agent-board/internal/mcp/` (new `server_test.go` and/or `tool_registry_test.go`, plus optional edits to `session_test.go`) are added/modified
  - And `services/agent-board/internal/mcp/server.go` is **byte-for-byte unchanged**

## UI / UX flow expectations
**No UI: BE-test only.**

## Out of scope
- **Modifying `mcp/server.go` production code.** Tests-only.
- **`CreateSession` / `GetSession`** — already covered (per `session_test.go` existing).
- **The `types.go` JSON-RPC types** — not in the tech-debt list.
- **SSE wire-format tests** — not in scope.

## Dependencies
- None. Independent.

## Notes for the team

- **`ListTools` doc-comment lies.** The comment says "in lexicographic order" but the implementation does NOT sort. AC scenario for UT-6 deliberately uses unordered membership check. If tester finds this annoying enough to warrant either fixing the implementation OR fixing the comment, raise as a NEW follow-up story — do not silently fix in this scope.
- **Race tests must be runnable.** Tests UT-8, UT-15 (and ideally a queue-concurrency one too if the tester wants to add it) MUST pass under `go test -race`. The current `ToolRegistry` uses `sync.RWMutex` and the `SessionManager` uses `sync.RWMutex` — both should be race-clean.
- **Channel capacity matters.** `Session.messages` is `make(chan []byte, 100)`. UT-10 pre-fills exactly 100 messages then expects the 101st `QueueMessage` to return the queue-full error. This is the load-bearing assertion for the queue-full branch.
- **Context cancellation for UT-12.** Use `context.WithCancel(context.Background())` and call `cancel()` before invoking `ReceiveMessage`. Alternative: `context.WithTimeout(ctx, 1*time.Millisecond)` to exercise `DeadlineExceeded`. Either is acceptable; pick one in the spec.
- **Audit reference.** `docs/tech_debt.md` lines 65–71 for the 7 sub-threshold functions.
- **Run locally before pushing:** `cd services/agent-board && go test ./internal/mcp -cover -race -v`.

## Sign-off log
(po-ba appends here on each sign-off pass)
