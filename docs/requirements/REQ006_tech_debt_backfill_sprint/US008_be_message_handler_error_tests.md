# US008/be_message_handler_error_tests

**Requirement:** REQ006
**Story:** US008
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** none
**Worked-by:**
**Implements:** REQ006/US008 AC (all scenarios — 13 verbatim test function names for `HandleMessage` / `sendError` / `sendToolResultError`, ≥95% per-file coverage modulo §4.5 exemptions, no production-code change). Architecture §3 US008 touch row + §4.3 cluster-2 patterns (where applicable) + §8 message.go test harness shape (D-010) + §4.5 exemption mechanism + §4.6 local verification command (US008 row).

## Goal
Create `services/agent-board/internal/handler/message_test.go` (NEW file) and add 13 verbatim test functions per US008 AC, using the **httptest + Echo** harness for `HandleMessage` and **direct call** (with the `internal_test.go` re-export OR indirect routing) for `sendError` / `sendToolResultError`, per architecture §8 / D-010. Tests-only.

## Scope
- **In:** Create the new file `services/agent-board/internal/handler/message_test.go`. Add the 13 test functions per US008 AC. Use real `mcp.SessionManager` + real `mcp.ToolRegistry` with controllable tool registrations (architecture §8.1 / D-010 explicitly rules out introducing a `SessionManager` interface).
- **In (conditional):** If indirect routing through `HandleMessage` cannot reach all coverage targets for `sendError` / `sendToolResultError`, the dev MAY add `services/agent-board/internal/handler/handler_internal_test.go` (package `handler`, NOT `handler_test`) re-exporting `var SendError = (*Handler).sendError` and `var SendToolResultError = (*Handler).sendToolResultError` per architecture §8.3 path-(a). This `_test.go` file is permitted under D-012's no-new-production-code rule because it is a test file. The dev SHOULD prefer path-(b) (indirect routing through `HandleMessage`) first; the re-export is the fallback only if coverage gaps remain.
- **Out:** Any change to `message.go`. Any change to siblings. Do NOT introduce a `SessionManager` interface for mocking (architecture §8.1 / D-010 explicit rejection).

## Files touched (estimated, exclusive)
- `services/agent-board/internal/handler/message_test.go` (NEW file)
- `services/agent-board/internal/handler/handler_internal_test.go` (NEW file — CONDITIONAL, only if path-b leaves coverage gaps per architecture §8.3)

## Test contract
Dev makes the 13 verbatim test-function names from US008 AC pass. Architecture §8 / D-010 fixes the harness shape — the 13 functions split:
- 4–5 functions exercise `HandleMessage` end-to-end via httptest + Echo (e.g. wrong JSON-RPC version, non-`tools/call` method, missing session, valid tool call happy path, tool call returning error).
- 2 functions exercise `sendError` (queue success + queue-full failure).
- 2 functions exercise `sendToolResultError` (queue success + queue-full failure).
- Remaining functions cover `_QueueMessageFails`, `_UnknownTool`, `_InvalidArguments`, `_NotifyToolError`, etc. per the AC verbatim list.

Tester's `US008_be_unit_tests.md` IT-* IDs map 1:1 onto these names.

## Implementation notes
- **Harness boilerplate (architecture §8.2 — copy-paste-able):**
  ```go
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
- **Queue-full setup (architecture §8.4):** pre-fill the session's message channel to capacity (100) so the next `QueueMessage` returns `errors.New("message queue full")`:
  ```go
  for i := 0; i < 100; i++ {
      require.NoError(t, sess.QueueMessage([]byte("filler")))
  }
  ```
- **Stub tool registration:** for tests exercising tool-error paths, register a one-off stub tool against `tr` (e.g. `tr.RegisterTool("failing-tool", func(ctx, args) ([]byte, error) { return nil, errors.New("boom") })`) and invoke via the JSON-RPC body in the request.
- **`json.Marshal` fallbacks at `message.go:46` and `message.go:64`** — reachable only if a registered tool returns a non-marshallable value (e.g. `chan int`). Per architecture §8.5 + §4.5 these are **acceptable to leave uncovered IF flagged in the test report**. Optional to chase: register a stub tool returning `make(chan int)` to drive these lines.
- **Path-(a) re-export file (CONDITIONAL):** if needed, the file `handler_internal_test.go` lives in `package handler` (internal test package) and contains nothing more than two `var` re-exports of the unexported method values. This is the idiomatic Go pattern for testing private methods from `handler_test`. Architecture §8.3 explicitly authorises this carve-out under D-012 (`_test.go` files don't count as production code).
- **Coverage check command** (architecture §4.6, US008 row):
  ```
  cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out \
      -run "TestHandleMessage|TestSendError|TestSendToolResultError"
  go tool cover -func=/tmp/handler.out | grep message.go
  ```
  Must show ≥95% statement coverage on `message.go` (modulo §8.5 / §4.5 `json.Marshal` fallback exemptions named in the test report).

## Definition of done
- All 13 new test functions present with US008 AC's verbatim names; all green via the local verification command above.
- `cd services/agent-board && go vet ./... && go test ./...` clean across the whole module.
- `message.go` ≥95% statement coverage (modulo any §4.5 / §8.5 exemptions named in the test report — `message.go:46` and `message.go:64` `json.Marshal` fallbacks are acceptable exemptions if not chased).
- `message.go` byte-for-byte unchanged.
- `golangci-lint run ./...` clean.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` + `scripts/review/run-gate.sh cross` both `REVIEW GATE: PASS`.
- **Live e2e NOT required** (tests-only); instead 3 clean runs of `cd services/agent-board && go test -count=3 ./internal/handler -race`.
- Dev set status to `in_review`; tech-lead approved.

## Review log
