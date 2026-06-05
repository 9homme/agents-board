# US009 — Backend unit & integration test specification
# `internal/mcp/server.go` ToolRegistry + Session message tests

**For BE Dev:** these are the tests you write FIRST (TDD red). Create `services/agent-board/internal/mcp/server_test.go` from scratch (place all 15 test functions here, or split across `tool_registry_test.go` + extensions to the existing `session_test.go` — test names are authoritative regardless of file placement). Production code in `server.go` is **byte-for-byte unchanged**.

**Harness shape (architecture.md §4.4):** bare-struct construction — `mcp.NewToolRegistry()`, `mcp.NewSessionManager()`. No DB, no HTTP. Pure in-memory state with mutex coordination. Invoke methods directly; assert on return values and struct state.

**Context-cancellation choice (architecture.md §4.4):** for `TestSession_ReceiveMessage_ContextCancelled`, this spec uses `context.WithCancel` and calls `cancel()` before invoking `ReceiveMessage`. The assertion is `errors.Is(err, context.Canceled)`.

## Coverage matrix

| AC scenario | Layer | Test ID | Function under test |
|---|---|---|---|
| `NewToolRegistry` returns empty registry | unit | UT-001 | `NewToolRegistry` |
| `RegisterTool` adds a handler | unit | UT-002 | `ToolRegistry.RegisterTool` |
| `RegisterTool` overwrites prior handler | unit | UT-003 | `ToolRegistry.RegisterTool` |
| `GetTool` on unknown name returns false | unit | UT-004 | `ToolRegistry.GetTool` |
| `GetTool` on known name returns handler | unit | UT-005 | `ToolRegistry.GetTool` |
| `ListTools` returns all registered names (unordered) | unit | UT-006 | `ToolRegistry.ListTools` |
| `ListTools` on empty registry returns empty slice | unit | UT-007 | `ToolRegistry.ListTools` |
| Concurrent RegisterTool + GetTool + ListTools — no data race | unit | UT-008 | `ToolRegistry` (concurrent) |
| `QueueMessage` happy path — message received | unit | UT-009 | `Session.QueueMessage` + `ReceiveMessage` |
| `QueueMessage` on full queue returns error | unit | UT-010 | `Session.QueueMessage` |
| `ReceiveMessage` happy path | unit | UT-011 | `Session.ReceiveMessage` |
| `ReceiveMessage` context cancelled | unit | UT-012 | `Session.ReceiveMessage` |
| `RemoveSession` removes session | unit | UT-013 | `SessionManager.RemoveSession` |
| `RemoveSession` on unknown ID is a no-op | unit | UT-014 | `SessionManager.RemoveSession` |
| Concurrent CreateSession + RemoveSession — no data race | unit | UT-015 | `SessionManager` (concurrent) |
| package-level coverage ≥95% | integration | IT-001 | `internal/mcp` package |
| full suite still passes under -race | integration | IT-002 | `go test ./...` |

## Unit tests

### UT-001 — `TestNewToolRegistry_ReturnsEmptyRegistry`
- **Function under test:** `NewToolRegistry`
- **Given:** `registry := mcp.NewToolRegistry()`
- **When:** `names := registry.ListTools()`
- **Then:** `assert.Empty(t, names)` (length 0)
- **Architecture cite:** architecture.md §4.4; US009 AC item 1

---

### UT-002 — `TestToolRegistry_RegisterTool_AddsHandler`
- **Function under test:** `ToolRegistry.RegisterTool`
- **Given:**
  ```go
  registry := mcp.NewToolRegistry()
  noop := func(ctx context.Context, args json.RawMessage) (interface{}, error) { return nil, nil }
  registry.RegisterTool("my_tool", noop)
  ```
- **When:** `h, ok := registry.GetTool("my_tool")`
- **Then:**
  - `ok` is `true`
  - `h` is non-nil
- **Architecture cite:** US009 AC item 2

---

### UT-003 — `TestToolRegistry_RegisterTool_OverwritesPriorHandler`
- **Function under test:** `ToolRegistry.RegisterTool` (overwrite)
- **Given:**
  ```go
  registry := mcp.NewToolRegistry()
  first := func(ctx context.Context, _ json.RawMessage) (interface{}, error) { return "first", nil }
  second := func(ctx context.Context, _ json.RawMessage) (interface{}, error) { return "second", nil }
  registry.RegisterTool("tool", first)
  registry.RegisterTool("tool", second)
  ```
- **When:** `h, ok := registry.GetTool("tool")`; `result, _ := h(context.Background(), nil)`
- **Then:**
  - `ok` is `true`
  - `result` equals `"second"` (latest handler wins)
- **Architecture cite:** US009 AC item 3

---

### UT-004 — `TestToolRegistry_GetTool_UnknownNameReturnsFalse`
- **Function under test:** `ToolRegistry.GetTool`
- **Given:** `registry := mcp.NewToolRegistry()` (empty)
- **When:** `h, ok := registry.GetTool("nonexistent")`
- **Then:**
  - `ok` is `false`
  - `h` is `nil`
- **Architecture cite:** US009 AC item 4

---

### UT-005 — `TestToolRegistry_GetTool_KnownNameReturnsHandler`
- **Function under test:** `ToolRegistry.GetTool`
- **Given:** registry with one registered tool that returns `"hello"`
- **When:** retrieve handler and invoke it
- **Then:** `result` equals `"hello"`, `err` is nil
- **Architecture cite:** US009 AC item 5

---

### UT-006 — `TestToolRegistry_ListTools_ReturnsAllRegisteredNames`
- **Function under test:** `ToolRegistry.ListTools`
- **Given:**
  ```go
  registry := mcp.NewToolRegistry()
  registry.RegisterTool("alpha", noop)
  registry.RegisterTool("beta", noop)
  registry.RegisterTool("gamma", noop)
  ```
- **When:** `names := registry.ListTools()`
- **Then:**
  - `assert.Len(t, names, 3)`
  - `assert.ElementsMatch(t, names, []string{"alpha", "beta", "gamma"})` — unordered membership check (NOT sorted `assert.Equal`)
- **Important note for test report:** the `ListTools` doc-comment claims "lexicographic order" but the implementation does NOT sort. The unordered check (`ElementsMatch`) is intentional; document this doc-comment vs. code mismatch in the test report under OQ-4 as a follow-up tech-debt line.
- **Architecture cite:** architecture.md §4.4 special case; US009 AC item 6; architecture.md §13.1 R-5

---

### UT-007 — `TestToolRegistry_ListTools_EmptyAfterNoRegistrations`
- **Function under test:** `ToolRegistry.ListTools`
- **Given:** fresh registry, no registrations
- **When:** `names := registry.ListTools()`
- **Then:** `assert.Empty(t, names)`
- **Architecture cite:** US009 AC item 7

---

### UT-008 — `TestToolRegistry_ConcurrentRegisterAndGet`
- **Function under test:** `ToolRegistry` thread-safety
- **Given:**
  ```go
  registry := mcp.NewToolRegistry()
  var wg sync.WaitGroup
  const N = 100
  wg.Add(N * 3) // 100 goroutines for Register, 100 for Get, 100 for List
  ```
- **When:** 100 goroutines each call `registry.RegisterTool(...)`, 100 call `registry.GetTool(...)`, 100 call `registry.ListTools()` — all interleaved
- **Then:** test must pass cleanly under `go test -race ./internal/mcp`
- **Architecture cite:** architecture.md §4.4 concurrent test note; US009 AC item 8; must use `sync.WaitGroup` + ≥100 goroutines

---

### UT-009 — `TestSession_QueueMessage_HappyPath`
- **Function under test:** `Session.QueueMessage` + `Session.ReceiveMessage`
- **Given:**
  ```go
  sm := mcp.NewSessionManager()
  sess := sm.CreateSession()
  msg := []byte(`{"test": true}`)
  ```
- **When:**
  ```go
  err1 := sess.QueueMessage(msg)
  received, err2 := sess.ReceiveMessage(context.Background())
  ```
- **Then:**
  - `err1` is nil
  - `err2` is nil
  - `received` equals `msg`
- **Architecture cite:** US009 AC item 9

---

### UT-010 — `TestSession_QueueMessage_FullReturnsError`
- **Function under test:** `Session.QueueMessage` (queue-full branch)
- **Given:**
  ```go
  sm := mcp.NewSessionManager()
  sess := sm.CreateSession()
  // Fill the channel to capacity (channel capacity is 100 — confirmed from server.go:59)
  for i := 0; i < 100; i++ {
      require.NoError(t, sess.QueueMessage([]byte("filler")))
  }
  ```
- **When:** `err := sess.QueueMessage([]byte("overflow"))`
- **Then:**
  - `err` is non-nil
  - `err.Error()` equals `"message queue full"`
- **Architecture cite:** architecture.md §4.4 `TestSession_QueueMessage_FullReturnsError`; US009 AC item 10; `server.go:59` capacity=100

---

### UT-011 — `TestSession_ReceiveMessage_HappyPath`
- **Function under test:** `Session.ReceiveMessage`
- **Given:** `sess.QueueMessage([]byte("hello"))`
- **When:** `msg, err := sess.ReceiveMessage(context.Background())`
- **Then:**
  - `err` is nil
  - `msg` equals `[]byte("hello")`
- **Architecture cite:** US009 AC item 11

---

### UT-012 — `TestSession_ReceiveMessage_ContextCancelled`
- **Function under test:** `Session.ReceiveMessage` (context cancel path)
- **Given:**
  ```go
  ctx, cancel := context.WithCancel(context.Background())
  cancel() // cancel BEFORE calling ReceiveMessage
  ```
- **When:** `msg, err := sess.ReceiveMessage(ctx)`
- **Then:**
  - `msg` is nil
  - `errors.Is(err, context.Canceled)` is `true`
- **Architecture cite:** architecture.md §4.4 context-cancel choice; US009 AC item 12

---

### UT-013 — `TestSessionManager_RemoveSession_RemovesSession`
- **Function under test:** `SessionManager.RemoveSession`
- **Given:**
  ```go
  sm := mcp.NewSessionManager()
  sess := sm.CreateSession()
  sessID := sess.ID
  sm.RemoveSession(sessID)
  ```
- **When:** `_, ok := sm.GetSession(sessID)`
- **Then:** `ok` is `false`
- **Architecture cite:** US009 AC item 13; tech_debt.md line 69 (`RemoveSession` at 0%)

---

### UT-014 — `TestSessionManager_RemoveSession_UnknownIDIsNoop`
- **Function under test:** `SessionManager.RemoveSession`
- **Given:**
  ```go
  sm := mcp.NewSessionManager()
  sess := sm.CreateSession() // other session that must remain intact
  ```
- **When:** `sm.RemoveSession("completely-nonexistent-id")` — must not panic
- **Then:**
  - No panic (test completes normally)
  - `_, ok := sm.GetSession(sess.ID)` → `ok` is `true` (the other session is still there)
- **Architecture cite:** US009 AC item 14

---

### UT-015 — `TestSessionManager_RemoveSession_ConcurrentSafe`
- **Function under test:** `SessionManager` concurrent safety
- **Given:**
  ```go
  sm := mcp.NewSessionManager()
  const N = 100
  var wg sync.WaitGroup
  wg.Add(N * 2)
  ```
- **When:** N goroutines each call `sm.CreateSession()` and N goroutines each call `sm.RemoveSession(...)` with the same session IDs — interleaved
- **Then:** test must pass cleanly under `go test -race ./internal/mcp`
- **Architecture cite:** architecture.md §4.4 concurrent test note; US009 AC item 15

## Integration tests

### IT-001 — package-level coverage ≥95%
- **Command:**
  ```
  cd services/agent-board && go test ./internal/mcp -coverprofile=/tmp/mcp.out
  go tool cover -func=/tmp/mcp.out
  ```
- **Expect:** `internal/mcp` package total statement coverage ≥95%.

### IT-002 — full suite regression + race detector
- **Command:**
  ```
  cd services/agent-board && go test ./internal/mcp -cover -race -v
  cd services/agent-board && go test ./... && golangci-lint run ./...
  ```
- **Expect:** all tests pass; zero data-race reports; no lint issues.

## Coverage exemptions

None anticipated. If any line in `server.go` is genuinely unreachable via the bare-struct harness, document under OQ-4. The `ListTools` doc-comment mismatch (lexicographic order claim vs. unsorted implementation) is flagged as a tech-debt note in the test report, NOT an uncovered line.
