# US032 — Backend unit & integration test specification
# `message.go` error-routing tests

**For BE Dev:** these are the tests you write FIRST (TDD red). Create `services/agent-board/internal/handler/message_test.go` from scratch (it does not currently exist). Production code in `message.go` is **byte-for-byte unchanged**.

**Harness decision (architecture.md §8, D-010):**
- `HandleMessage` — tested via `httptest.NewRequest` + `httptest.NewRecorder` against an `echo.New()` server with `HandleMessage` mounted at `POST /message`.
- `sendError` and `sendToolResultError` — unexported. **Chosen approach: path (b)** — test these helpers indirectly through `HandleMessage` routing paths that exercise them (items 4, 5, 6, 7 in the AC). For UT-010, UT-011, UT-012, UT-013 (direct queue assertions + queue-full), additionally create `services/agent-board/internal/handler/handler_internal_test.go` with `package handler` that re-exports the two helpers as `var SendError = (*Handler).sendError` and `var SendToolResultError = (*Handler).sendToolResultError`. This is the idiomatic Go internal-test-helper pattern for external test packages; the file is a `_test.go` file and does NOT constitute a production-code change per architecture.md §8.3 note.

**Shared test setup (copy into test file — architecture.md §8.2):**
```go
package handler_test

func newTestHandler(t *testing.T) (*handler.Handler, *mcp.Session, *mcp.ToolRegistry) {
    t.Helper()
    sm := mcp.NewSessionManager()
    tr := mcp.NewToolRegistry()
    h := handler.NewHandler(sm, tr)
    sess := sm.CreateSession()
    return h, sess, tr
}

func postMessage(t *testing.T, h *handler.Handler, sessionID string, body []byte) *httptest.ResponseRecorder {
    t.Helper()
    e := echo.New()
    e.POST("/message", h.HandleMessage)
    req := httptest.NewRequest(http.MethodPost, "/message?sessionId="+sessionID, bytes.NewReader(body))
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()
    e.ServeHTTP(rec, req)
    return rec
}
```

**Queue-full setup (architecture.md §8.4):**
```go
for i := 0; i < 100; i++ {
    err := sess.QueueMessage([]byte("filler"))
    require.NoError(t, err)
}
// next QueueMessage call inside HandleMessage / sendError will return err
```

## Coverage matrix

| AC scenario | Layer | Test ID | Function under test |
|---|---|---|---|
| Missing sessionId query param | unit | UT-001 | `HandleMessage` |
| Invalid (unknown) sessionId | unit | UT-002 | `HandleMessage` |
| Malformed JSON body | unit | UT-003 | `HandleMessage` |
| Non-tools/call method routes to sendError | unit | UT-004 | `HandleMessage` → `sendError` |
| Wrong JSON-RPC version routes to sendError | unit | UT-005 | `HandleMessage` → `sendError` |
| Tool not found routes to sendToolResultError | unit | UT-006 | `HandleMessage` → `sendToolResultError` |
| Tool execution error routes to sendToolResultError | unit | UT-007 | `HandleMessage` → `sendToolResultError` |
| Happy path — tool returns value, session queued | unit | UT-008 | `HandleMessage` |
| QueueMessage fails → HTTP 500 | unit | UT-009 | `HandleMessage` |
| sendError queues JSONRPCResponse + returns echo error | unit | UT-010 | `sendError` (via re-exported helper) |
| sendToolResultError queues JSONRPCResponse + returns echo error | unit | UT-011 | `sendToolResultError` (via re-exported helper) |
| sendError with queue full — logs, still returns echo error | unit | UT-012 | `sendError` queue-full path |
| sendToolResultError with queue full — logs, still returns echo error | unit | UT-013 | `sendToolResultError` queue-full path |
| per-file coverage ≥95% (modulo json.Marshal fallbacks) | integration | IT-001 | `message.go` all functions |
| full suite still passes | integration | IT-002 | `go test ./...` |

## Unit tests

### UT-001 — `TestHandleMessage_MissingSessionID`
- **Function under test:** `HandleMessage`
- **Given:** test handler via `newTestHandler`
- **When:** `postMessage(t, h, "", validJSONRPCBody)`
- **Then:**
  - `rec.Code` is `400`
  - `rec.Body.String()` contains `"sessionId is required"`
- **Architecture cite:** architecture.md §8; US032 AC item 1

---

### UT-002 — `TestHandleMessage_InvalidSessionID`
- **Function under test:** `HandleMessage`
- **When:** `postMessage(t, h, "does-not-exist", validJSONRPCBody)`
- **Then:**
  - `rec.Code` is `400`
  - `rec.Body.String()` contains `"invalid sessionId"`
- **Architecture cite:** US032 AC item 2

---

### UT-003 — `TestHandleMessage_InvalidJSONPayload`
- **Function under test:** `HandleMessage`
- **When:** `postMessage(t, h, sess.ID, []byte("not-json"))`
- **Then:**
  - `rec.Code` is `400`
  - `rec.Body.String()` contains `"invalid JSON-RPC payload"`
- **Architecture cite:** US032 AC item 3

---

### UT-004 — `TestHandleMessage_NonToolsCallMethod`
- **Function under test:** `HandleMessage` → routes to `sendError`
- **When:** `postMessage(t, h, sess.ID, []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`))` 
- **Then:**
  - `rec.Code` satisfies the response contract (HTTP 200 per D-010 — `sendError` returns `*echo.HTTPError` with status 200, which Echo serialises as 200)
  - The session queue has one message containing `"InvalidRequest"` (or the MCP code constant) and message `"Invalid request"` — OR the response body via Echo reflects those fields (read the source to determine whether the serialised body is available on rec or in the queue)
- **Architecture cite:** architecture.md §8.3; US032 AC item 4; `mcp.InvalidRequest`

---

### UT-005 — `TestHandleMessage_WrongJSONRPCVersion`
- **Function under test:** `HandleMessage` → routes to `sendError`
- **When:** body has `jsonrpc: "1.0"` and `method: "tools/call"`
- **Then:** same assertion as UT-004 (sendError path, `InvalidRequest`)
- **Architecture cite:** US032 AC item 5

---

### UT-006 — `TestHandleMessage_ToolNotFound`
- **Function under test:** `HandleMessage` → routes to `sendToolResultError`
- **Given:** no tools registered in `tr` (registry is empty)
- **When:** `postMessage(t, h, sess.ID, validToolsCallBody("nonexistent_tool"))`
- **Then:**
  - Response (via rec or session queue) contains `"Tool not found"`
- **Architecture cite:** US032 AC item 6

---

### UT-007 — `TestHandleMessage_ToolExecutionError`
- **Function under test:** `HandleMessage` → routes to `sendToolResultError`
- **Given:**
  ```go
  tr.RegisterTool("my_tool", func(_ context.Context, _ json.RawMessage) (interface{}, error) {
      return nil, errors.New("boom")
  })
  ```
- **When:** `postMessage(t, h, sess.ID, validToolsCallBody("my_tool"))`
- **Then:**
  - Response contains `"boom"` (tool error message surfaced)
- **Architecture cite:** US032 AC item 7

---

### UT-008 — `TestHandleMessage_HappyPath`
- **Function under test:** `HandleMessage`
- **Given:**
  ```go
  tr.RegisterTool("echo_tool", func(_ context.Context, args json.RawMessage) (interface{}, error) {
      return map[string]string{"answer": "42"}, nil
  })
  ```
- **When:** `postMessage(t, h, sess.ID, validToolsCallBodyWithID("echo_tool", 99))`
- **Then:**
  - `rec.Code` is `200`
  - Session queue has received exactly one message
  - Decoded queue message is a `JSONRPCResponse` with `ID = 99`, `Result.Content[0].Type == "text"`, and `Result.Content[0].Text` contains the marshalled `{"answer":"42"}` string
- **Architecture cite:** US032 AC item 8

---

### UT-009 — `TestHandleMessage_QueueMessageFails`
- **Function under test:** `HandleMessage` queue-full path
- **Given:**
  ```go
  // pre-fill queue to capacity
  for i := 0; i < 100; i++ {
      require.NoError(t, sess.QueueMessage([]byte("filler")))
  }
  tr.RegisterTool("any_tool", func(_ context.Context, _ json.RawMessage) (interface{}, error) {
      return map[string]string{"ok": "yes"}, nil
  })
  ```
- **When:** `postMessage(t, h, sess.ID, validToolsCallBody("any_tool"))`
- **Then:**
  - `rec.Code` is `500`
  - `rec.Body.String()` contains `"failed to queue message"`
- **Architecture cite:** architecture.md §8.4; US032 AC item 9

---

### UT-010 — `TestSendError_QueuesAndReturnsEchoError`
- **Function under test:** `sendError` (via `handler_internal_test.go` re-export `handler.SendError`)
- **Given:** `h, sess, _ := newTestHandler(t)` (queue empty)
- **When:** `result := handler.SendError(h, sess, 5, mcp.InvalidRequest, "Invalid request")`
- **Then:**
  - `result` is `*echo.HTTPError` with `Code == 200`
  - `sess` queue has one message; decode it as `JSONRPCResponse`; assert `Error.Code == mcp.InvalidRequest` AND `Error.Message == "Invalid request"` AND `ID == 5`
- **Architecture cite:** architecture.md §8.3 path-(a) approach; US032 AC item 10

---

### UT-011 — `TestSendToolResultError_QueuesAndReturnsEchoError`
- **Function under test:** `sendToolResultError` (via re-export)
- **When:** `result := handler.SendToolResultError(h, sess, 7, "boom")`
- **Then:**
  - `result` is `*echo.HTTPError` with `Code == 200`
  - Queue message decodes to a `JSONRPCResponse` with `Result.IsError == true`, `Result.Content[0].Type == "text"`, `Result.Content[0].Text == "boom"`, and `ID == 7`
- **Architecture cite:** US032 AC item 11

---

### UT-012 — `TestSendError_QueueFailure_LogsButReturnsEchoError`
- **Function under test:** `sendError` queue-full path
- **Given:** pre-fill queue to capacity (100 messages)
- **When:** `result := handler.SendError(h, sess, 1, mcp.InvalidRequest, "msg")`
- **Then:**
  - `result` is non-nil `*echo.HTTPError` (NOT nil — the error is still returned to the caller)
  - Queue is still full (the 101st QueueMessage failed silently with a log)
  - (The `log.Printf("failed to queue error message: ...")` path is exercised — confirmed by coverage)
- **Architecture cite:** architecture.md §8.3; US032 AC item 12; `message.go:46` log.Printf path

---

### UT-013 — `TestSendToolResultError_QueueFailure_LogsButReturnsEchoError`
- **Function under test:** `sendToolResultError` queue-full path
- **Given:** pre-fill queue to capacity
- **When:** `result := handler.SendToolResultError(h, sess, 1, "msg")`
- **Then:**
  - `result` is non-nil `*echo.HTTPError`
  - `log.Printf("failed to queue tool result error: ...")` path is exercised
- **Architecture cite:** architecture.md §8.3; US032 AC item 13; `message.go:64` log.Printf path

## Integration tests

### IT-001 — per-file coverage ≥95%
- **Command:**
  ```
  cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out \
      -run "TestHandleMessage|TestSendError|TestSendToolResultError"
  go tool cover -func=/tmp/handler.out | grep message.go
  ```
- **Expect:** `message.go` total statement coverage ≥95%, OR `≥95% modulo the two unreachable json.Marshal error fallbacks`.
- **Acceptable uncovered lines (OQ-4):**
  - `message.go:46` — `mcp.InternalError` fallback on `json.Marshal(response)` failure inside `sendError`. The `JSONRPCResponse` struct only contains `string`, `int`, `interface{}` fields with marshallable contents. Unreachable without injecting a non-marshallable tool result (e.g. `chan int`). Acceptable to leave uncovered; document in test report.
  - `message.go:64` — symmetric `mcp.InternalError` fallback in `sendToolResultError`. Same rationale.

### IT-002 — full suite regression
- **Command:** `cd services/agent-board && go test ./... && golangci-lint run ./...`
- **Expect:** all pre-existing tests pass; no new lint issues.

## Coverage exemptions

- `message.go:46` — `json.Marshal` failure fallback in `sendError` — unreachable for marshallable `JSONRPCResponse` struct without injecting a non-marshallable value (e.g. `chan int` tool return). Acceptable per architecture.md §4.5.
- `message.go:64` — symmetric fallback in `sendToolResultError` — same rationale. Acceptable per architecture.md §4.5.
