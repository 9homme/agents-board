# US003/be_audit_trail_mcp

**Requirement:** REQ003
**Story:** US003
**Track:** BE
**Service:** services/agent-board
**Status:** in_review
**Blocked by:** US001_be_scaffold_domain_and_migration.md
**Worked-by:** be-dev-2026-05-19T00:00:00Z-a6ce
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
