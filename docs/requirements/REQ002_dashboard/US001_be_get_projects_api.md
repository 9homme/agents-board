# US001/be_get_projects_api

**Requirement:** REQ002
**Story:** US001
**Track:** BE
**Service:** services/agent-board
**Status:** in_progress
**Blocked by:** US001_be_service_rename_and_api_server_scaffold.md
**Worked-by:** be-dev-timestamp-1234
**Implements:** GET /api/v1/projects API contract, UI data flow

## Goal
Implement the `GET /api/v1/projects` REST endpoint in the new API server, reading from the existing `internal/repo` shared repository.

## Scope
- **In:** Adding a new HTTP handler for `/api/v1/projects` returning the exact JSON shape defined in the architecture. Wiring it up in `cmd/api-server/main.go`. Creating or extending `internal/repo` methods to list all projects if not already present.
- **Out:** Any modifications to the database schema or MCP server logic.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/handler/project_handler.go`
- `services/agent-board/internal/handler/project_handler_test.go`
- `services/agent-board/cmd/api-server/main.go`

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US001_be_unit_tests.md`: UT-002, IT-002 (refer to the BE unit test document for exact IDs related to the GET /api/v1/projects endpoint)

## Implementation notes
- The JSON response shape must exactly match `{"projects": [{"id": "...", "name": "...", "description": "...", "createdAt": "...", "updatedAt": "..."}]}`. Ensure field names use camelCase as specified.
- The `internal/repo` likely already has something to fetch projects. Reuse it. If you need a new method, ensure it doesn't break existing MCP usages.
- Standardize the error model to return `{"code": "INTERNAL_ERROR", "message": "Failed to fetch projects"}` on 500 error.

## Definition of done
- All listed tests green.
- `go vet ./...` and `go test ./...` clean inside the task's service module.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved.

## Review log
