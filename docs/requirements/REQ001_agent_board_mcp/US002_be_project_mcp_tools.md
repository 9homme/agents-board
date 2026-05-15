# US002/be_project_mcp_tools

**Requirement:** REQ001
**Story:** US002
**Track:** BE
**Service:** services/agent-board-mcp
**Status:** completed
**Blocked by:** US001_be_mcp_server.md, US002_be_schema_and_project_repo.md
**Worked-by:** be-dev-2024-05-15T12:00:00Z-abcd
**Implements:** Project Tools JSON Schemas (`create_project`, `get_project`, `update_project`, `delete_project`, `list_projects`)

## Goal
Implement the MCP tool handlers for Project CRUD and register them with the MCP server.

## Scope
- **In:** `internal/handler/project_tools.go`, updates to wire the tools in `cmd/agent-board-mcp/main.go`.
- **Out:** Project database repository (already done), UI pagination.

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US002_be_unit_tests.md`: UT-* for Project MCP tool request parsing and response formatting.

## Implementation notes
- Tools must strictly validate the input arguments against the JSON schemas in the architecture.
- Format responses exactly as defined in the architecture (Result Content JSON).
- Return an MCP error response (text block with `isError: true`) for validation failures or not found.
- Connect to the `ProjectRepo` interface.

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Review log

## Notes
- Files touched:
  - `services/agent-board-mcp/internal/handler/project_tools.go`
  - `services/agent-board-mcp/internal/handler/project_tools_test.go`
  - `services/agent-board-mcp/cmd/agent-board-mcp/main.go`
  - `services/agent-board-mcp/internal/handler/document_tools_test.go` (fixed a bug in a parallel task's test assertion to ensure `go test` passes)
- Tests added: 5 IT-* test cases (IT-002 to IT-006) for MCP tool handlers.
- All tests pass successfully. Code strict adheres to the API schema in architecture.
