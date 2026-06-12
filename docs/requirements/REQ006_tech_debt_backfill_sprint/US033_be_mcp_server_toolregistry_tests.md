# US033/be_mcp_server_toolregistry_tests

**Requirement:** REQ006
**Story:** US033
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** none
**Worked-by:** be-dev-2026-06-06T00:00:00Z-a9f3
**Implements:** REQ006/US033 AC (all scenarios — 15 verbatim test function names covering `NewToolRegistry`, `ToolRegistry.RegisterTool` / `GetTool` / `ListTools` / concurrent register-and-get, `Session.QueueMessage` / `ReceiveMessage`, `SessionManager.RemoveSession`, ≥95% per-file coverage modulo §4.5 exemptions, no production-code change). Architecture §3 US033 touch row + §4.4 cluster-3 bare-struct test pattern + §4.5 exemption mechanism + §4.6 local verification command (US033 row).

## Goal
Create `services/agent-board/internal/mcp/server_test.go` (NEW file) and add 15 verbatim test functions per US033 AC, covering `ToolRegistry`, `Session.QueueMessage` / `ReceiveMessage`, and `SessionManager.RemoveSession`. Bare-struct testing — no DB, no HTTP. Tests-only. **Race-clean under `go test -race`** is mandatory for the concurrent tests.

## Scope
- **In:** Create the new file `services/agent-board/internal/mcp/server_test.go` (architecture §3 US033 row authoritative). Add 15 test functions per US033 AC. Per architecture §3 note the tester MAY split into `tool_registry_test.go` + `session_test.go` additions (existing `session_test.go` is tiny and may be merged into). The test-function NAMES are authoritative regardless of file placement; the dev follows whatever split the tester locks in `US033_be_unit_tests.md`.
- **In:** `TestToolRegistry_ConcurrentRegisterAndGet` and `TestSessionManager_RemoveSession_ConcurrentSafe` MUST run cleanly under `go test -race ./internal/mcp` (architecture §4.4 special case).
- **Out:** Any change to `server.go`. **The `ListTools` doc-comment vs. code mismatch ("lexicographic order" — code does NOT sort) is NOT silently fixed** — flag it in the test report under OQ-4 per architecture §3 US033 row + §4.4 + R-5; use unordered membership check (`assert.ElementsMatch`) in `TestToolRegistry_ListTools_ReturnsAllRegisteredNames`.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/mcp/server_test.go` (NEW file)

(If the tester elects to split into `tool_registry_test.go` and to add into `session_test.go`, this list expands accordingly — dev follows tester's spec layout. Update `## Files touched` at claim time if the spec lands differently.)

## Test contract
Dev makes the 15 verbatim test-function names from US033 AC pass. Coverage hits:
- `TestNewToolRegistry_*` family.
- `TestToolRegistry_RegisterTool_*`, `TestToolRegistry_GetTool_*`, `TestToolRegistry_ListTools_*`, `TestToolRegistry_ConcurrentRegisterAndGet`.
- `TestSession_QueueMessage_*` (including `_FullReturnsError`), `TestSession_ReceiveMessage_*` (including `_ContextCancelled`).
- `TestSessionManager_RemoveSession_*` (including `_ConcurrentSafe`).

Tester's `US033_be_unit_tests.md` UT-* IDs map 1:1 onto these names.

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
- **`TestToolRegistry_ListTools_ReturnsAllRegisteredNames`:** the production code does NOT sort despite the doc comment. **Assertion: `assert.ElementsMatch` (unordered membership), NOT `assert.Equal` on a sorted list.** DO NOT silently fix the doc-comment-vs-code mismatch — flag as tech-debt in the test report (architecture §4.4 + §3 US033 row + R-5).
- **Run with race:** `cd services/agent-board && go test -race -v ./internal/mcp` is the local verification command (architecture §4.6, US033 row).
- **Coverage check command** (architecture §4.6, US033 row):
  ```
  cd services/agent-board && go test ./internal/mcp -cover -race -v
  ```
  Per-file coverage on `server.go` must hit ≥95% modulo any §4.5 exemptions.

## Definition of done
- All 15 new test functions present with US033 AC's verbatim names; all green via `cd services/agent-board && go test -race -cover -v ./internal/mcp`.
- `cd services/agent-board && go vet ./... && go test ./...` clean across the whole module.
- `server.go` ≥95% statement coverage (modulo any §4.5 exemptions named in the test report — including the `ListTools` doc-comment-mismatch flagged in the test report).
- `server.go` byte-for-byte unchanged.
- `golangci-lint run ./...` clean.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` + `scripts/review/run-gate.sh cross` both `REVIEW GATE: PASS`.
- **Race-clean assertion:** `cd services/agent-board && go test -race ./internal/mcp` passes; in particular the two concurrent tests do NOT report data races.
- **Live e2e NOT required** (tests-only); instead 3 clean runs of `cd services/agent-board && go test -count=3 -race ./internal/mcp`.
- Dev set status to `in_review`; tech-lead approved.

## Notes

### Files touched
- `services/agent-board/internal/mcp/server_test.go` (NEW — 226 lines, 15 test functions)

### Tests added (all 15 verbatim UT-* names per US033_be_unit_tests.md)
| UT-ID | Test function | Result |
|---|---|---|
| UT-001 | `TestNewToolRegistry_ReturnsEmptyRegistry` | PASS |
| UT-002 | `TestToolRegistry_RegisterTool_AddsHandler` | PASS |
| UT-003 | `TestToolRegistry_RegisterTool_OverwritesPriorHandler` | PASS |
| UT-004 | `TestToolRegistry_GetTool_UnknownNameReturnsFalse` | PASS |
| UT-005 | `TestToolRegistry_GetTool_KnownNameReturnsHandler` | PASS |
| UT-006 | `TestToolRegistry_ListTools_ReturnsAllRegisteredNames` | PASS |
| UT-007 | `TestToolRegistry_ListTools_EmptyAfterNoRegistrations` | PASS |
| UT-008 | `TestToolRegistry_ConcurrentRegisterAndGet` | PASS |
| UT-009 | `TestSession_QueueMessage_HappyPath` | PASS |
| UT-010 | `TestSession_QueueMessage_FullReturnsError` | PASS |
| UT-011 | `TestSession_ReceiveMessage_HappyPath` | PASS |
| UT-012 | `TestSession_ReceiveMessage_ContextCancelled` | PASS |
| UT-013 | `TestSessionManager_RemoveSession_RemovesSession` | PASS |
| UT-014 | `TestSessionManager_RemoveSession_UnknownIDIsNoop` | PASS |
| UT-015 | `TestSessionManager_RemoveSession_ConcurrentSafe` | PASS |

### Coverage
`server.go` coverage: **100.0%** (all 10 exported functions at 100%). Requirement was ≥95%.

### Race detector
`go test -count=3 -race ./internal/mcp/` — 3/3 clean runs. No data races detected.

### Review gates
- `scripts/review/run-gate.sh be services/agent-board` → **REVIEW GATE: PASS**
- `scripts/review/run-gate.sh cross` → **REVIEW GATE: PASS**

### OQ-4 tech-debt flag (per spec §4.4 + §13.1 R-5)
`ListTools` doc-comment in `server.go:116` claims "lexicographic order" but the implementation iterates a map (unordered). `TestToolRegistry_ListTools_ReturnsAllRegisteredNames` uses `assert.ElementsMatch` (unordered membership) intentionally. The doc-comment vs. code mismatch is NOT fixed here — it is flagged as a follow-up tech-debt item (OQ-4). Resolution options: sort in the implementation, or correct the doc-comment to "unspecified order".

### Production code
`server.go` is byte-for-byte unchanged (no production code modified).

## Review log

### Review pass 1 — 2026-06-07 — verdict: approved

**Verdict: APPROVED.** All DoD criteria met.

**Test contract (15/15 verbatim names present + 2 IT-*):**
- All 15 UT-* function names from `US033_be_unit_tests.md` present in `server_test.go` with verbatim AC names. Confirmed `grep -cE "^func Test" server_test.go` = 15 (16th test `TestSessionCreationAndMessageQueuing` is the pre-existing `session_test.go` test, left intact).
- IT-001 (≥95% coverage) and IT-002 (full suite + race) both satisfied.

**`go test -race ./internal/mcp/... -v`:** all 16 tests `--- PASS`, no `DATA RACE`. Summary: `ok  agent-board/internal/mcp  1.334s`.

**Flake check — 3 consecutive `go test -count=1 -race ./internal/mcp/...`:**
- RUN 1: `ok  agent-board/internal/mcp  1.407s`
- RUN 2: `ok  agent-board/internal/mcp  1.315s`
- RUN 3: `ok  agent-board/internal/mcp  1.295s`
- 3/3 clean, zero data races. (Live e2e N/A per task DoD — tests-only; 3-clean-run race check is the substitute gate.)

**Full suite — `go test ./...`:** all 7 packages `ok` (api-server, mcp-server, config, domain, handler, mcp, repo).

**Per-file coverage — `go tool cover -func=/tmp/cov009.out`:**
- `server.go` — **100.0%** of statements (every one of the 10 functions: QueueMessage, ReceiveMessage, NewSessionManager, CreateSession, GetSession, RemoveSession, NewToolRegistry, RegisterTool, GetTool, ListTools all at 100.0%). Package total 100.0%. Requirement ≥95% — exceeded. No `## Coverage exemption` needed.

**Review gate:**
```
== BE gate · services/agent-board ==
  PASS  gofmt -s (no diff)
  PASS  go vet ./...
  PASS  golangci-lint run ./...
  PASS  go test ./...
WARN  gosec (skipped — not installed; coverage via golangci-lint gosec linter)
WARN  govulncheck (skipped — not installed)
REVIEW GATE: PASS
```
```
== Cross-cutting · repo ==
  PASS  semgrep (owasp/golang/typescript)
  PASS  gitleaks (no secrets)
REVIEW GATE: PASS
```
Both gates emitted `REVIEW GATE: PASS` (exit 0). The gosec/govulncheck WARN-skips are gate-internal non-fatal fallbacks (gosec coverage retained via golangci-lint's gosec linter) — the gate itself emitted PASS, so this is a clean gate result, not a `blocked_review_gate` tooling condition.

**OQ-4 (ListTools doc-comment vs code mismatch) — correctly handled, NOT silently fixed:**
- `server.go:115` doc-comment claims "lexicographic order" but `ListTools` iterates a map (unordered). Production code byte-for-byte unchanged (confirmed `git diff HEAD -- server.go` empty).
- `TestToolRegistry_ListTools_ReturnsAllRegisteredNames` (UT-006) uses `assert.ElementsMatch` (unordered membership), NOT a sorted `assert.Equal` — correct per spec.
- OQ-4 flagged in task `## Notes` AND in the `server_test.go` file header comment (lines 4-8). Tech-debt line filed below.

**Spec exhaustiveness (anti-REQ005 branch check):** every SUT branch maps to a spec case:
- `QueueMessage` (2 branches: send-success / queue-full) → UT-009, UT-010.
- `ReceiveMessage` (2 branches: ctx.Done / msg) → UT-012, UT-011.
- `RemoveSession` (present-ID / unknown-ID noop / concurrent) → UT-013, UT-014, UT-015.
- `GetTool` (found / not-found / concurrent) → UT-005, UT-004, UT-008.
- `RegisterTool` (add / overwrite / concurrent) → UT-002, UT-003, UT-008.
- `ListTools` (populated / empty / concurrent) → UT-006, UT-007, UT-008.
- Constructors → UT-001. No uncovered branch — no SPEC_GAP_FOUND.

**TDG conformance:** test file committed in `1db8836` `red: test spec for all 15 ToolRegistry+Session+SessionManager tests (US033)` — `red:` prefix + `(US033)` tag, conformant. Tests-only task against an already-existing, already-correct SUT: a single `red:` commit (tests that pass green against unchanged production code) is the correct shape; no production code to write, so no separate `green:` cycle is expected.

**Production code:** `server.go` byte-for-byte unchanged (`git diff HEAD` empty). Scope respected — only `server_test.go` (NEW) added.

**Tech-debt:** one non-blocking finding filed to `docs/tech_debt.md` (OQ-4 doc-comment/code mismatch carry-forward).
