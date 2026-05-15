# US005/be_task_crud

**Requirement:** REQ001
**Story:** US005
**Track:** BE
**Service:** services/agent-board-mcp
**Status:** completed
**Blocked by:** US001_be_mcp_server.md, US002_be_schema_and_project_repo.md
**Worked-by:** be-dev
**Implements:** Task Tools JSON Schemas, Task Data Model

## Goal
Implement the Task repository and register the MCP tools for Task CRUD operations.

## Scope
- **In:** `internal/domain/task.go`, `internal/repo/task_repo.go`, `internal/handler/task_tools.go`.
- **Out:** Database schema migration (already handled).

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US005_be_unit_tests.md`: all Task-related UT-* and IT-*.

## Implementation notes
- Implement repository methods for Task CRUD.
- Register tools: `create_task`, `get_task`, `update_task`, `delete_task`, `list_tasks`.
- Ensure responses exactly match the Task Tools JSON Schemas in the architecture.
- Handle `userStoryId` foreign key appropriately.

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Review log
### Review pass 1 — 2024-05-18 — verdict: approved
- `go test ./...` passed.
- `go vet ./...` clean.
- Code matches the `architecture.md` schemas exactly.
- Great job handling the JSON responses and database queries.
