# US003/be_audit_trail_mcp

**Requirement:** REQ003
**Story:** US003
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** US001_be_scaffold_domain_and_migration.md
**Worked-by:** be-dev-2026-05-19T00:00:00Z-a337
**Implements:** US003, API contracts for get_task_audit_trail and get_user_story_audit_trail

## Goal
Add new MCP tools to fetch the state change audit trail for tasks and user stories.

## Scope
- **In:** `internal/repo` queries to fetch audit logs by entity ID and type, new MCP handlers `get_task_audit_trail` and `get_user_story_audit_trail`, and registration of these tools.
- **Out:** Writing audit logs (handled in US001 and US002 tasks).

## Files touched (estimated, exclusive)
- `services/agent-board/internal/repo/audit_repo.go`
- `services/agent-board/internal/repo/audit_repo_test.go`
- `services/agent-board/internal/handler/audit_tools.go`
- `services/agent-board/internal/handler/audit_tools_test.go`
- `services/agent-board/internal/mcp/server.go`
- `services/agent-board/cmd/mcp-server/main.go` (added by orchestrator on review pass 1 — wire `RegisterAuditTools` at runtime)

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US003_be_unit_tests.md`: Check specific UT and IT IDs for fetching audit trails.

## Implementation notes
- Add `GetAuditTrail` methods to a new `AuditRepo` or the existing repos (a dedicated `AuditRepo` is fine).
- The new tools `get_task_audit_trail` and `get_user_story_audit_trail` must return the exact JSON payload specified in `architecture.md`.
- Ensure the result is ordered chronologically as specified in the schema definition (using the index).

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Notes
### Files touched
- `services/agent-board/internal/repo/audit_repo.go` — New `AuditRepository` interface + `auditRepo` implementation with `GetTaskAuditTrail` and `GetUserStoryAuditTrail` methods querying `status_audit_trail` ordered by `changed_at ASC`.
- `services/agent-board/internal/repo/audit_repo_test.go` — IT-004/IT-005 at repo layer using `go-sqlmock`.
- `services/agent-board/internal/handler/audit_tools.go` — `AuditLogResponse` type + `RegisterAuditTools` registering `get_task_audit_trail` and `get_user_story_audit_trail` MCP tools matching architecture JSON shapes exactly.
- `services/agent-board/internal/handler/audit_tools_test.go` — IT-003/IT-004/IT-005 at handler layer with mock repos.
- `services/agent-board/internal/mcp/server.go` — Added `ListTools()` method to `ToolRegistry`.

### Tests added
- `TestAuditRepo_GetTaskAuditTrail` (IT-004 repo layer)
- `TestAuditRepo_GetUserStoryAuditTrail` (IT-005 repo layer)
- `TestAuditRepo_GetTaskAuditTrail_Empty`
- `TestAuditTools_GetTaskAuditTrail` (IT-004 handler layer)
- `TestAuditTools_GetTaskAuditTrail_Empty`
- `TestAuditTools_GetTaskAuditTrail_MissingTaskID`
- `TestAuditTools_GetUserStoryAuditTrail` (IT-005 handler layer)
- `TestAuditTools_GetUserStoryAuditTrail_MissingID`
- `TestAuditTools_NoAuditOnInvalidTaskTransition` (IT-003)
- `TestAuditTools_NoAuditOnInvalidUserStoryTransition` (IT-003)
- `TestAuditTools_GetTaskAuditTrail_RepoError`

### Follow-up
- `cmd/mcp-server/main.go` needs `handler.RegisterAuditTools(toolRegistry, repo.NewAuditRepo(db))` to wire the tools at runtime. This file was not in the task's `Files touched` list and was not modified per task scope. Tech-lead should verify or add a separate task for this wiring.

## Review log

### Review pass 1 — 2026-05-19 — verdict: changes_requested

**Tests & gates (all green):**
- `go vet ./...` — clean. `go test ./...` inside `services/agent-board` — all packages PASS (domain, handler, mcp, repo).
- BE review gate: `REVIEW GATE: PASS` — gofmt -s, go vet, golangci-lint, go test, gosec, govulncheck all PASS.
- Cross review gate: `REVIEW GATE: PASS` — semgrep (owasp/golang/typescript), gitleaks all PASS.

**Test contract — satisfied:**
- IT-004 (repo): `TestAuditRepo_GetTaskAuditTrail` (audit_repo_test.go:14) — chronological order verified.
- IT-005 (repo): `TestAuditRepo_GetUserStoryAuditTrail` (audit_repo_test.go:52).
- IT-004 (handler): `TestAuditTools_GetTaskAuditTrail` (audit_tools_test.go:32) — exact JSON shape verified.
- IT-005 (handler): `TestAuditTools_GetUserStoryAuditTrail` (audit_tools_test.go:121).
- IT-003: `TestAuditTools_NoAuditOnInvalidTaskTransition` / `TestAuditTools_NoAuditOnInvalidUserStoryTransition` (audit_tools_test.go:185, :224).
- JSON contract conformance: `AuditLogResponse` (audit_tools.go:17-24) matches architecture lines 82-94 / 105-117 field-for-field; `changedAt` formatted RFC3339; `entityType` carried through ('task' / 'user_story'); `auditTrail` empty slice (not nil) on no results. No deviation.
- Chronological ordering: `ORDER BY changed_at ASC` (audit_repo.go:29) — matches architecture index intent (line 133).
- Scope discipline: `git diff` confirms only the 5 declared `## Files touched` paths were modified.

**Required change (blocking):**
1. `cmd/mcp-server/main.go` — the two new tools are never registered with the running server. `RegisterAuditTools` (audit_tools.go:39) is implemented but never called. `main.go:53-56` registers project/document/user-story/task tools but not audit tools, so `get_task_audit_trail` and `get_user_story_audit_trail` resolve to "Tool not found" in `HandleMessage` (message.go:34-37) at runtime — the tools are unreachable.
   - This is incomplete delivery, not a separate integration task: this task's own `## Scope: In` (line 16) explicitly includes "**and registration of these tools**", the architecture lists "Add new MCP tools ..." as `internal/handler` responsibility and "Provide MCP tools to retrieve the audit trail" as in-scope (architecture lines 8, 51, 71), US003 AC "Retrieve audit trail" requires the backend to answer the query, and no separate wiring task exists in the REQ003 README task table. The US001/US002 precedent registered their tools in `main.go` (lines 55-56).
   - Fix: add `handler.RegisterAuditTools(toolRegistry, repo.NewAuditRepo(db))` to `cmd/mcp-server/main.go` after line 56. `cmd/mcp-server/main.go` must be added to `## Files touched` — the orchestrator will widen the list before re-spawning the dev.
   - The dev correctly left the file untouched given the original `## Files touched` list and disclosed the gap in `## Notes` — good discipline. The remediation is to widen scope, not to penalize the disclosure.

**Verdict:** changes_requested. Implementation is otherwise correct, contract-faithful, well-tested, and gate-clean; the single remaining gap is runtime wiring of the delivered tools.

### Rework — 2026-05-19 — response to Review pass 1

**Finding addressed:**
- `cmd/mcp-server/main.go` — added `handler.RegisterAuditTools(toolRegistry, repo.NewAuditRepo(db))` on line 57, immediately after the existing `RegisterTaskTools` call. This wires `get_task_audit_trail` and `get_user_story_audit_trail` into the live MCP server, making them reachable at runtime. No other changes were made.

**Verification:**
- `go vet ./...` — clean.
- `go build ./...` — clean.
- `go test ./...` — all packages PASS (domain, handler, mcp, repo).
- BE review gate: `REVIEW GATE: PASS` (gofmt -s, go vet, golangci-lint, go test, gosec, govulncheck).
- Cross review gate: `REVIEW GATE: PASS` (semgrep, gitleaks).

### Review pass 2 — 2026-05-19 — verdict: approved

**Pass-1 finding — resolved:**
- `cmd/mcp-server/main.go:57` now calls `handler.RegisterAuditTools(toolRegistry, repo.NewAuditRepo(db))`, placed immediately after `RegisterTaskTools` (line 56) and before `NewHandler` (line 59). Correct registry variable (`toolRegistry`), correct db handle (`db`), correct constructor (`repo.NewAuditRepo`). `get_task_audit_trail` and `get_user_story_audit_trail` are now reachable at runtime via `HandleMessage`.

**Scope discipline:**
- Rework commit `c6a5ee1` adds exactly one line to `cmd/mcp-server/main.go` and nothing else. `git log -p` confirms no other files touched in the rework. `cmd/mcp-server/main.go` was widened into `## Files touched` (commit `b402451`). No drive-by edits.

**Tests & gates — all green:**
- `go vet ./...` — clean. `go build ./...` — clean. `go test ./...` inside `services/agent-board` — all packages PASS (domain, handler, mcp, repo).
- BE review gate (`scripts/review/run-gate.sh be services/agent-board`):
  ```
  == BE gate · services/agent-board ==
    PASS  gofmt -s (no diff)
    PASS  go vet ./...
    PASS  golangci-lint run ./...
    PASS  go test ./...
    PASS  gosec ./... (security)
    PASS  govulncheck ./...

  REVIEW GATE: PASS
  ```
- Cross review gate (`scripts/review/run-gate.sh cross`):
  ```
  == Cross-cutting · repo ==
    PASS  semgrep (owasp/golang/typescript)
    PASS  gitleaks (no secrets)

  REVIEW GATE: PASS
  ```

**Verdict:** approved. The sole pass-1 blocker (runtime wiring of the audit tools) is fully resolved with a minimal, correct one-line change. Implementation remains contract-faithful (`AuditLogResponse` matches architecture lines 82-94 / 105-117 field-for-field, `ORDER BY changed_at ASC`), test contract (IT-003/IT-004/IT-005) satisfied, no regressions. Streak reset on approval.
