# US004/be_user_story_crud

**Requirement:** REQ001
**Story:** US004
**Track:** BE
**Service:** services/agent-board-mcp
**Status:** completed
**Blocked by:** US001_be_mcp_server.md, US002_be_schema_and_project_repo.md
**Worked-by:** be-dev
**Implements:** User Story Tools JSON Schemas, User Story Data Model

## Goal
Implement the User Story repository and register the MCP tools for User Story CRUD operations.

## Scope
- **In:** `internal/domain/user_story.go`, `internal/repo/user_story_repo.go`, `internal/handler/user_story_tools.go`.
- **Out:** Database schema migration (already handled).

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US004_be_unit_tests.md`: all User Story-related UT-* and IT-*.

## Implementation notes
- Implement repository methods for User Story CRUD.
- Register tools: `create_user_story`, `get_user_story`, `update_user_story`, `delete_user_story`, `list_user_stories`.
- Ensure responses exactly match the User Story Tools JSON Schemas in the architecture.

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Review log
### Review pass 1 — 2024-05-18 — verdict: approved
- Checked internal/handler/user_story_tools.go schemas against architecture JSON specifications. They match.
- Checked domain types and SQL data structures.
- All tests green.

## Notes
- Created `internal/domain/user_story.go`.
- Created `internal/repo/user_story_repo.go` and `internal/repo/user_story_repo_test.go` (implementing UT-015 to UT-019).
- Created `internal/handler/user_story_tools.go` and `internal/handler/user_story_tools_test.go` (implementing IT-012 to IT-016).
- Registered the tools in `cmd/agent-board-mcp/main.go`.
- Tests pass (`go vet` and `go test` clean).
