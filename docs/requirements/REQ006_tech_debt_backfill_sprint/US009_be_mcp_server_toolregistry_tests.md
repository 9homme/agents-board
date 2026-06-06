# US009/be_mcp_server_toolregistry_tests

**Requirement:** REQ006
**Story:** US009
**Track:** BE
**Service:** services/agent-board
**Status:** in_progress
**Blocked by:** none
**Worked-by:** be-dev-2026-06-06T00:00:00Z-a9f3
**Implements:** REQ006/US009 AC (all scenarios — 15 verbatim test function names covering `NewToolRegistry`, `ToolRegistry.RegisterTool` / `GetTool` / `ListTools` / concurrent register-and-get, `Session.QueueMessage` / `ReceiveMessage`, `SessionManager.RemoveSession`, ≥95% per-file coverage modulo §4.5 exemptions, no production-code change). Architecture §3 US009 touch row + §4.4 cluster-3 bare-struct test pattern + §4.5 exemption mechanism + §4.6 local verification command (US009 row).

## Goal
Create `services/agent-board/internal/mcp/server_test.go` (NEW file) and add 15 verbatim test functions per US009 AC, covering `ToolRegistry`, `Session.QueueMessage` / `ReceiveMessage`, and `SessionManager.RemoveSession`. Bare-struct testing — no DB, no HTTP. Tests-only. **Race-clean under `go test -race`** is mandatory for the concurrent tests.

## Scope
- **In:** Create the new file `services/agent-board/internal/mcp/server_test.go` (architecture §3 US009 row authoritative). Add 15 test functions per US009 AC. Per architecture §3 note the tester MAY split into `tool_registry_test.go` + `session_test.go` additions (existing `session_test.go` is tiny and may be merged into). The test-function NAMES are authoritative regardless of file placement; the dev follows whatever split the tester locks in `US009_be_unit_tests.md`.
- **In:** `TestToolRegistry_ConcurrentRegisterAndGet` and `TestSessionManager_RemoveSession_ConcurrentSafe` MUST run cleanly under `go test -race ./internal/mcp` (architecture §4.4 special case).
- **Out:** Any change to `server.go`. **The `ListTools` doc-comment vs. code mismatch ("lexicographic order" — code does NOT sort) is NOT silently fixed** — flag it in the test report under OQ-4 per architecture §3 US009 row + §4.4 + R-5; use unordered membership check (`assert.ElementsMatch`) in `TestToolRegistry_ListTools_ReturnsAllRegisteredNames`.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/mcp/server_test.go` (NEW file)

(If the tester elects to split into `tool_registry_test.go` and to add into `session_test.go`, this list expands accordingly — dev follows tester's spec layout. Update `## Files touched` at claim time if the spec lands differently.)

## Test contract
Dev makes the 15 verbatim test-function names from US009 AC pass. Coverage hits:
- `TestNewToolRegistry_*` family.
- `TestToolRegistry_RegisterTool_*`, `TestToolRegistry_GetTool_*`, `TestToolRegistry_ListTools_*`, `TestToolRegistry_ConcurrentRegisterAndGet`.
- `TestSession_QueueMessage_*` (including `_FullReturnsError`), `TestSession_ReceiveMessage_*` (including `_ContextCancelled`).
- `TestSessionManager_RemoveSession_*` (including `_ConcurrentSafe`).

Tester's `US009_be_unit_tests.md` UT-* IDs map 1:1 onto these names.

## Implementation notes
- **Test file structure (architecture §4.4 — copy-paste-able):**
  ```go
  func TestToolRegistry_RegisterTool_PopulatesRegistry(t *testing.T) {
      registry := mcp.NewToolRegistry()
      registry.RegisterTool("foo", func(ctx context.Context, args json.RawMessage) ([]byte, error) {
          return []byte(`"ok"`), nil
      })
      handler, ok := registry.GetTool("foo")
      assert.True(t, ok)
      assert.NotNil(t, handler)
  }
  ```
- **`TestSession_QueueMessage_FullReturnsError`** (architecture §4.4 special case): pre-fill EXACTLY 100 messages (channel capacity hard-coded at `server.go:59` as `make(chan []byte, 100)`), assert 101st `QueueMessage` returns `errors.New("message queue full")`.
- **`TestSession_ReceiveMessage_ContextCancelled`** (architecture §4.4): `ctx, cancel := context.WithCancel(...)`; `cancel()`; assert the returned error matches `context.Canceled` via `errors.Is`. Alternative `context.WithTimeout(ctx, 1*time.Millisecond)` for `DeadlineExceeded` is equally acceptable — tester picks ONE.
- **Concurrent tests:** use `sync.WaitGroup` + ≥100 goroutines. `ToolRegistry` uses `sync.RWMutex` internally so properly written tests will not race; the test exists to GUARD against regressions that drop the lock. Same for `SessionManager.RemoveSession_ConcurrentSafe`.
- **`TestToolRegistry_ListTools_ReturnsAllRegisteredNames`:** the production code does NOT sort despite the doc comment. **Assertion: `assert.ElementsMatch` (unordered membership), NOT `assert.Equal` on a sorted list.** DO NOT silently fix the doc-comment-vs-code mismatch — flag as tech-debt in the test report (architecture §4.4 + §3 US009 row + R-5).
- **Run with race:** `cd services/agent-board && go test -race -v ./internal/mcp` is the local verification command (architecture §4.6, US009 row).
- **Coverage check command** (architecture §4.6, US009 row):
  ```
  cd services/agent-board && go test ./internal/mcp -cover -race -v
  ```
  Per-file coverage on `server.go` must hit ≥95% modulo any §4.5 exemptions.

## Definition of done
- All 15 new test functions present with US009 AC's verbatim names; all green via `cd services/agent-board && go test -race -cover -v ./internal/mcp`.
- `cd services/agent-board && go vet ./... && go test ./...` clean across the whole module.
- `server.go` ≥95% statement coverage (modulo any §4.5 exemptions named in the test report — including the `ListTools` doc-comment-mismatch flagged in the test report).
- `server.go` byte-for-byte unchanged.
- `golangci-lint run ./...` clean.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` + `scripts/review/run-gate.sh cross` both `REVIEW GATE: PASS`.
- **Race-clean assertion:** `cd services/agent-board && go test -race ./internal/mcp` passes; in particular the two concurrent tests do NOT report data races.
- **Live e2e NOT required** (tests-only); instead 3 clean runs of `cd services/agent-board && go test -count=3 -race ./internal/mcp`.
- Dev set status to `in_review`; tech-lead approved.

## Review log
