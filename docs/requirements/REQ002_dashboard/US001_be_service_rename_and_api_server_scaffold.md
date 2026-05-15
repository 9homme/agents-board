# US001/be_service_rename_and_api_server_scaffold

**Requirement:** REQ002
**Story:** US001
**Track:** BE
**Service:** services/agent-board
**Status:** in_review
**Blocked by:** 
**Worked-by:** be-dev-20240321T100000Z-abcd
**Implements:** D-001 Split Application Entrypoints

## Goal
Rename the `services/agent-board-mcp` directory to `services/agent-board` and scaffold the `cmd/mcp-server` and `cmd/api-server` executables to support both MCP and REST entrypoints.

## Scope
- **In:** Renaming the `services/agent-board-mcp` directory to `services/agent-board`. Moving `cmd/agent-board-mcp/main.go` to `cmd/mcp-server/main.go`. Creating a basic Echo server at `cmd/api-server/main.go` with CORS support enabled for Next.js frontend origin. Updating `go.mod` module name to reflect `agent-board`.
- **Out:** Implementing the actual `GET /api/v1/projects` REST handlers (handled in subsequent task).

## Files touched (estimated, exclusive)
- `services/agent-board-mcp/` (rename to `services/agent-board/`)
- `services/agent-board/cmd/agent-board-mcp/main.go` (move to `services/agent-board/cmd/mcp-server/main.go`)
- `services/agent-board/cmd/api-server/main.go`
- `services/agent-board/go.mod`

This is a **scaffold task**. Other BE tasks for this story are blocked by this task.

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US001_be_unit_tests.md`: UT-001 (if any for server scaffolding)

## Implementation notes
- Git rename the directory.
- Update `go.mod` to `module agent-board`. You may need to run `go mod tidy` and update imports in existing files to point to the new module path `agent-board/internal/...`.
- `cmd/api-server/main.go` should instantiate an Echo instance, configure CORS to allow all origins or a configured `FRONTEND_URL` environment variable, read the `DATABASE_URL`, and start on port 8080.
- `cmd/mcp-server/main.go` is just the relocated existing code.

## Definition of done
- All listed tests green.
- `go vet ./...` and `go test ./...` clean inside the task's service module.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved.

## Review log
enamed the directory to `services/agent-board` and `cmd/agent-board-mcp` to `cmd/mcp-server` via `git mv`.
- Created `services/agent-board/cmd/api-server/main.go` with Echo and CORS.
- Updated `go.mod` to `module agent-board` and ran `go mod tidy`.
- Replaced all imports of `agent-board-mcp` with `agent-board` in internal go files.
- Gates ran successfully. Note: `US001_be_unit_tests.md` only specifies `UT-001` and `UT-002` for `internal/handler.GetProjects` which are explicitly marked as "out of scope" in the task's instructions. As such, no unit tests were written for scaffolding.

## Review log
