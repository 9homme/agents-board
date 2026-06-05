# US008 — Backfill `message.go` error-routing tests

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** draft

## Story
As a **future contributor changing `services/agent-board/internal/handler/message.go`**, I want **every error-routing branch in `HandleMessage`, `sendError`, and `sendToolResultError` to be covered by integration tests against `httptest`**, so that a regression (e.g. dropping the `sessionId` validation, returning the wrong JSON-RPC error code, or queue-failure not surfacing 500) fails CI immediately.

## Acceptance criteria

- **Scenario: `message_test.go` is created and gains the following test functions (verbatim names)**
  - Given `services/agent-board/internal/handler/message_test.go` does not currently exist (per `ls` output)
  - When the story is complete
  - Then the file exists and contains the following new test functions:
    1. `TestHandleMessage_MissingSessionID` — query param `sessionId` empty → HTTP 400 with body `{"error":"sessionId is required"}`.
    2. `TestHandleMessage_InvalidSessionID` — `sessionId=does-not-exist` → HTTP 400 with body `{"error":"invalid sessionId"}`.
    3. `TestHandleMessage_InvalidJSONPayload` — `c.Bind` fails (malformed JSON) → HTTP 400 with body `{"error":"invalid JSON-RPC payload"}`.
    4. `TestHandleMessage_NonToolsCallMethod` — body has `jsonrpc:"2.0"` and `method:"initialize"` → routes through `sendError` with code `mcp.InvalidRequest` and message `"Invalid request"`.
    5. `TestHandleMessage_WrongJSONRPCVersion` — body has `jsonrpc:"1.0"`, `method:"tools/call"` → same `sendError` path as above.
    6. `TestHandleMessage_ToolNotFound` — valid request, `toolRegistry.GetTool` returns `(nil, false)` → routes through `sendToolResultError` with message `"Tool not found"`.
    7. `TestHandleMessage_ToolExecutionError` — registered tool returns `(nil, errors.New("boom"))` → routes through `sendToolResultError` with message `"boom"`.
    8. `TestHandleMessage_HappyPath` — registered tool returns a marshallable value → HTTP 200 with a `JSONRPCResponse` containing `Result.Content[0].Text` matching the marshalled value, AND `session.QueueMessage` was called once.
    9. `TestHandleMessage_QueueMessageFails` — `session.QueueMessage` returns err (queue full) → HTTP 500 with body `{"error":"failed to queue message"}`.
    10. `TestSendError_QueuesAndReturnsEchoError` — direct call to `(*Handler).sendError(session, id, code, message)` → queues a marshalled `JSONRPCResponse` with `Error: {Code: code, Message: message}` on the session AND returns an `*echo.HTTPError` with status 200 whose `Message` is the response struct.
    11. `TestSendToolResultError_QueuesAndReturnsEchoError` — direct call to `(*Handler).sendToolResultError(session, id, message)` → queues a marshalled `JSONRPCResponse` with `Result: {IsError: true, Content: [{Type:"text", Text: message}]}` AND returns an `*echo.HTTPError` with status 200.
    12. `TestSendError_QueueFailure_LogsButReturnsEchoError` — session queue is full → `session.QueueMessage` returns err; function still returns the `*echo.HTTPError` (does NOT propagate the queue err to the caller); covers the `log.Printf("failed to queue error message: ...")` path.
    13. `TestSendToolResultError_QueueFailure_LogsButReturnsEchoError` — symmetric to above for `sendToolResultError` (covers the `log.Printf("failed to queue tool result error: ...")` path).
  - **Note on `json.Marshal` failure paths (lines 64, 84, 110).** `json.Marshal` failure on a `JSONRPCResponse` is genuinely unreachable from happy code paths (the struct only contains `string`, `int`, and `interface{}` fields with marshallable values). The `mcp.InternalError` fallback branches at lines 46 and 64 of `HandleMessage` are NOT in the AC's required-coverage list. If tester is willing to inject a non-marshallable value via a custom `interface{}` (e.g. a channel) returned from a stub tool, that is welcome but optional. If left uncovered, document in test report under OQ-4.

- **Scenario: each new test exercises the specific branch**
  - Given an Echo test server constructed via `echo.New()` with `Handler.HandleMessage` registered on `POST /message`
  - And a `Handler` constructed with a mock `sessionManager` (or a real one with a known session pre-created) and a real `mcp.ToolRegistry` with controllable tool registrations
  - And `httptest.NewRequest` / `httptest.NewRecorder` are used to invoke the route
  - When the test posts the relevant body to `/message?sessionId=<value>`
  - Then assertions match the test name:
    - For HTTP 400/500 cases: `rec.Code` matches AND `rec.Body.String()` contains the expected error key
    - For `sendError` / `sendToolResultError` cases: the returned `*echo.HTTPError` has the expected `Code` (status 200 in both cases by current code), AND the session queue contains the expected serialised `JSONRPCResponse`
    - For `_HappyPath`: response body is a `JSONRPCResponse` with `ID` matching the request ID, `Result.Content[0].Type == "text"`, `Result.Content[0].Text == marshalled(tool_return_value)`, AND the session queue received the same bytes

- **Scenario: per-file coverage hits ≥95%**
  - Given `cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out -run "TestHandleMessage|TestSendError|TestSendToolResultError"`
  - When `go tool cover -func=/tmp/handler.out | grep message.go` is inspected
  - Then `message.go` shows **≥95% statement coverage**, OR documented as `≥95% modulo the two unreachable `json.Marshal` error fallbacks` (baselines per `docs/tech_debt.md` lines 59–61: `HandleMessage` 70.4%, `sendError` 0%, `sendToolResultError` 0%)
  - And the unreachable lines are explicitly named in the test report under OQ-4

- **Scenario: existing tests still pass and behaviour is unchanged**
  - Given `message.go` is **NOT** modified by this story
  - When `cd services/agent-board && go test ./...` runs
  - Then all pre-existing tests pass
  - And all new tests pass
  - And `golangci-lint run ./...` is clean

- **Scenario: no production-code changes**
  - Given `git diff` of the story's commits
  - When inspected
  - Then **only** `services/agent-board/internal/handler/message_test.go` (new file) and optionally a shared test helper is created/modified
  - And `services/agent-board/internal/handler/message.go` is **byte-for-byte unchanged**

## UI / UX flow expectations
**No UI: BE-test only.**

## Out of scope
- **Modifying handler production code.** Tests-only.
- **`project_tools.go` / `document_tools.go` / `task_tools.go` / `user_story_tools.go`** — US004–US007.
- **SSE endpoint (`sse.go`)** — not in the tech-debt list; out of REQ006 scope.
- **The two genuinely unreachable `json.Marshal` error paths** — documented under OQ-4 if not covered.

## Dependencies
- None. Independent.

## Notes for the team

- **`message_test.go` does not exist today.** Confirmed via `ls services/agent-board/internal/handler/`. This story creates it from scratch.
- **Test harness shape (OQ-5 in README).** This story uses `httptest` + Echo because `HandleMessage` is invoked through Echo's `Context`. The two helpers (`sendError`, `sendToolResultError`) take a `*mcp.Session` directly and can be tested without an HTTP request — see the direct-call AC entries.
- **Session manager mocking.** Tester may either (a) use the real `mcp.SessionManager` and pre-create a known session, or (b) introduce an interface and a mock. po-ba prefers (a) for simplicity — it is the same approach the existing handler tests already use. Tester's call.
- **Queue-full scenario.** `Session.QueueMessage` returns err when the channel is full (capacity 100). Tester fills the channel by writing 100 dummy messages, then triggers the test path.
- **Audit reference.** `docs/tech_debt.md` lines 59–61 for the three baseline numbers (all 0% except `HandleMessage`).
- **Run locally before pushing:** `cd services/agent-board && go test ./internal/handler -cover -v -run "TestHandleMessage|TestSendError|TestSendToolResultError"`.

## Sign-off log
(po-ba appends here on each sign-off pass)
