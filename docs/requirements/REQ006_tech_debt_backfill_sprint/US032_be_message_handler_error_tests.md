# US032/be_message_handler_error_tests

**Requirement:** REQ006
**Story:** US032
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** none
**Worked-by:** be-dev-2026-06-06T00:00:00Z-a6a4
**Implements:** REQ006/US032 AC (all scenarios — 13 verbatim test function names for `HandleMessage` / `sendError` / `sendToolResultError`, ≥95% per-file coverage modulo §4.5 exemptions, no production-code change). Architecture §3 US032 touch row + §4.3 cluster-2 patterns (where applicable) + §8 message.go test harness shape (D-010) + §4.5 exemption mechanism + §4.6 local verification command (US032 row).

## Goal
Create `services/agent-board/internal/handler/message_test.go` (NEW file) and add 13 verbatim test functions per US032 AC, using the **httptest + Echo** harness for `HandleMessage` and **direct call** (with the `internal_test.go` re-export OR indirect routing) for `sendError` / `sendToolResultError`, per architecture §8 / D-010. Tests-only.

## Scope
- **In:** Create the new file `services/agent-board/internal/handler/message_test.go`. Add the 13 test functions per US032 AC. Use real `mcp.SessionManager` + real `mcp.ToolRegistry` with controllable tool registrations (architecture §8.1 / D-010 explicitly rules out introducing a `SessionManager` interface).
- **In (conditional):** If indirect routing through `HandleMessage` cannot reach all coverage targets for `sendError` / `sendToolResultError`, the dev MAY add `services/agent-board/internal/handler/handler_internal_test.go` (package `handler`, NOT `handler_test`) re-exporting `var SendError = (*Handler).sendError` and `var SendToolResultError = (*Handler).sendToolResultError` per architecture §8.3 path-(a). This `_test.go` file is permitted under D-012's no-new-production-code rule because it is a test file. The dev SHOULD prefer path-(b) (indirect routing through `HandleMessage`) first; the re-export is the fallback only if coverage gaps remain.
- **Out:** Any change to `message.go`. Any change to siblings. Do NOT introduce a `SessionManager` interface for mocking (architecture §8.1 / D-010 explicit rejection).

## Files touched (estimated, exclusive)
- `services/agent-board/internal/handler/message_test.go` (NEW file)
- `services/agent-board/internal/handler/handler_internal_test.go` (NEW file — CONDITIONAL, only if path-b leaves coverage gaps per architecture §8.3)

## Test contract
Dev makes the 13 verbatim test-function names from US032 AC pass. Architecture §8 / D-010 fixes the harness shape — the 13 functions split:
- 4–5 functions exercise `HandleMessage` end-to-end via httptest + Echo (e.g. wrong JSON-RPC version, non-`tools/call` method, missing session, valid tool call happy path, tool call returning error).
- 2 functions exercise `sendError` (queue success + queue-full failure).
- 2 functions exercise `sendToolResultError` (queue success + queue-full failure).
- Remaining functions cover `_QueueMessageFails`, `_UnknownTool`, `_InvalidArguments`, `_NotifyToolError`, etc. per the AC verbatim list.

Tester's `US032_be_unit_tests.md` IT-* IDs map 1:1 onto these names.

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
- **Coverage check command** (architecture §4.6, US032 row):
  ```
  cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out \
      -run "TestHandleMessage|TestSendError|TestSendToolResultError"
  go tool cover -func=/tmp/handler.out | grep message.go
  ```
  Must show ≥95% statement coverage on `message.go` (modulo §8.5 / §4.5 `json.Marshal` fallback exemptions named in the test report).

## Definition of done
- All 13 new test functions present with US032 AC's verbatim names; all green via the local verification command above.
- `cd services/agent-board && go vet ./... && go test ./...` clean across the whole module.
- `message.go` ≥95% statement coverage (modulo any §4.5 / §8.5 exemptions named in the test report — `message.go:46` and `message.go:64` `json.Marshal` fallbacks are acceptable exemptions if not chased).
- `message.go` byte-for-byte unchanged.
- `golangci-lint run ./...` clean.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` + `scripts/review/run-gate.sh cross` both `REVIEW GATE: PASS`.
- **Live e2e NOT required** (tests-only); instead 3 clean runs of `cd services/agent-board && go test -count=3 ./internal/handler -race`.
- Dev set status to `in_review`; tech-lead approved.

## Notes

### Files touched
- `services/agent-board/internal/handler/message_test.go` — NEW file, 14 test functions (13 AC + 1 optional chan-int coverage chase for message.go:46)
- `services/agent-board/internal/handler/handler_internal_test.go` — NEW file, re-exports `sendError`/`sendToolResultError` as `handler.SendError`/`handler.SendToolResultError` for UT-010..UT-013 direct-call tests (architecture §8.3 path-a)

### Tests added
- UT-001 `TestHandleMessage_MissingSessionID` — PASS
- UT-002 `TestHandleMessage_InvalidSessionID` — PASS
- UT-003 `TestHandleMessage_InvalidJSONPayload` — PASS
- UT-004 `TestHandleMessage_NonToolsCallMethod` — PASS
- UT-005 `TestHandleMessage_WrongJSONRPCVersion` — PASS
- UT-006 `TestHandleMessage_ToolNotFound` — PASS
- UT-007 `TestHandleMessage_ToolExecutionError` — PASS
- UT-008 `TestHandleMessage_HappyPath` — PASS
- UT-009 `TestHandleMessage_QueueMessageFails` — PASS
- UT-010 `TestSendError_QueuesAndReturnsEchoError` — PASS
- UT-011 `TestSendToolResultError_QueuesAndReturnsEchoError` — PASS
- UT-012 `TestSendError_QueueFailure_LogsButReturnsEchoError` — PASS
- UT-013 `TestSendToolResultError_QueueFailure_LogsButReturnsEchoError` — PASS
- Optional: `TestHandleMessage_NonMarshalableToolResult` — PASS (drives message.go:46)

### Coverage results (`go test ./internal/handler -coverprofile`)
- `HandleMessage`: 96.3% (26/27 stmts) — message.go:64 json.Marshal fallback exempt (§4.5)
- `sendError`: 75.0% (3/4 stmts) — message.go:84-86 json.Marshal fallback exempt (§4.5)
- `sendToolResultError`: 75.0% (3/4 stmts) — message.go:109-111 json.Marshal fallback exempt (§4.5)

### Coverage exemptions (OQ-4)
- `services/agent-board/internal/handler/message.go:84-86` — `json.Marshal` failure fallback in `sendError` — `JSONRPCResponse` struct contains only marshallable fields (string/int/interface{} of string/int). Unreachable without non-marshallable content. Acceptable per architecture.md §4.5.
- `services/agent-board/internal/handler/message.go:109-111` — symmetric fallback in `sendToolResultError` — same rationale. Acceptable per architecture.md §4.5.
- `services/agent-board/internal/handler/message.go:63-64` — `json.Marshal(resp)` fallback after tool result already serialised to `[]byte` — the `JSONRPCResponse` at this point contains only string fields; unreachable. Acceptable per architecture.md §4.5 / §8.5.

### Race check (3 clean runs)
`cd services/agent-board && go test -count=3 ./internal/handler -race` — 366 tests passed, 0 race conditions.

### Full suite
`cd services/agent-board && go test ./...` — 224 tests, 0 failures.

### Review gates
- `scripts/review/run-gate.sh be services/agent-board` — REVIEW GATE: PASS
- `scripts/review/run-gate.sh cross` — REVIEW GATE: PASS

### Live e2e
Not required per DoD (tests-only story). Race-checked 3 clean runs substituted per DoD item.

### Production code
`message.go` is byte-for-byte unchanged. No new production symbols introduced.

## Review log

### Review pass 1 — 2026-06-07 — verdict: approved

**Test contract (all 13 verbatim AC names + 1 optional, all PASS):**
- UT-001 `TestHandleMessage_MissingSessionID` · UT-002 `TestHandleMessage_InvalidSessionID` · UT-003 `TestHandleMessage_InvalidJSONPayload` · UT-004 `TestHandleMessage_NonToolsCallMethod` · UT-005 `TestHandleMessage_WrongJSONRPCVersion` · UT-006 `TestHandleMessage_ToolNotFound` · UT-007 `TestHandleMessage_ToolExecutionError` · UT-008 `TestHandleMessage_HappyPath` · UT-009 `TestHandleMessage_QueueMessageFails` · UT-010 `TestSendError_QueuesAndReturnsEchoError` · UT-011 `TestSendToolResultError_QueuesAndReturnsEchoError` · UT-012 `TestSendError_QueueFailure_LogsButReturnsEchoError` · UT-013 `TestSendToolResultError_QueueFailure_LogsButReturnsEchoError` — plus optional `TestHandleMessage_NonMarshalableToolResult`. `go test ./internal/handler -v` → `ok agent-board/internal/handler`, 14 PASS / 0 FAIL.

**Architecture conformance (§8 / D-010):**
- Harness is httptest+Echo for `HandleMessage`, direct-call via `handler_internal_test.go` re-exports for `sendError`/`sendToolResultError` — exactly the §8.3 path-(a) carve-out. No `SessionManager` interface introduced (real `mcp.NewSessionManager()` / `mcp.NewToolRegistry()`); D-010 / §8.1 honored.
- `message.go` byte-for-byte unchanged: `git diff HEAD -- services/agent-board/internal/handler/message.go` empty (exit 0). No new production symbols.

**Spec-exhaustiveness (anti-REQ005):** 13 reachable error/state branches in `message.go` (missing/invalid sessionId, bind error, wrong version/method→sendError, tool-not-found→sendToolResultError, tool-exec-error, queue-fail in HandleMessage, happy path, sendError queue-success+queue-full, sendToolResultError queue-success+queue-full) — each maps 1:1 to a UT-* case. The only uncovered branches are the three `json.Marshal` fallbacks (message.go:46, :64, :84-86, :110-113) — all named exempt in spec §coverage-exemptions and architecture §4.5/§8.5. No spec gap.

**Coverage (`go tool cover -func` on message.go):**
```
agent-board/internal/handler/message.go:14:  HandleMessage         96.3%
agent-board/internal/handler/message.go:74:  sendError             75.0%
agent-board/internal/handler/message.go:95:  sendToolResultError   75.0%
```
Raw block profile confirms the ONLY uncovered blocks are `message.go:84.16,87.3` and `message.go:110.16,113.3` (the json.Marshal fallbacks — exempt §4.5/§8.5). The `log.Printf` queue-full paths (`message.go:88.56,90.3` and `:114.56,116.3`) are count=1 (COVERED by UT-012/UT-013 — not faked). sendError/sendToolResultError below 95% is fully accounted for by the named exemptions; HandleMessage 96.3% ≥ 95%.

**TDG conformance:** US032 commits follow red→green→refactor with `(US032)` tags — `58943fe red: test spec for all 13 ... (US032)`, `47f137d green: all 14 message handler tests pass — production code unchanged (US032)`, `089ffc9 refactor: ... (US032)`, `e911619 refactor: chore: hand off US032 ... (US032)`. All tdg-prefixed, correct ordering.

**Race (DoD substitute for live-e2e; tests-only story):** `go test -count=3 ./internal/handler -race` → `ok agent-board/internal/handler`, race clean across 3 runs.

**Gates:**
- `scripts/review/run-gate.sh be services/agent-board` → exit 0, `REVIEW GATE: PASS` (gofmt/vet/golangci-lint/go-test PASS; gosec+govulncheck WARN-skipped — gosec coverage retained via golangci-lint gosec linter).
- `scripts/review/run-gate.sh cross` → exit 0, `REVIEW GATE: PASS` (semgrep PASS, gitleaks PASS).
- Full module: `go test ./...` → all 7 packages `ok`.
- No Robot e2e suites exist for REQ006 (`tests/e2e/REQ006*` absent) — dryrun/live-e2e N/A; tests-only story per DoD.

**Tech-debt:** none filed this pass.

**Verdict: approved → Status: completed.**
