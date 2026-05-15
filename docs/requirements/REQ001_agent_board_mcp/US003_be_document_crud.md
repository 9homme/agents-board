# US003/be_document_crud

**Requirement:** REQ001
**Story:** US003
**Track:** BE
**Service:** services/agent-board-mcp
**Status:** in_review
**Blocked by:** US001_be_mcp_server.md, US002_be_schema_and_project_repo.md
**Worked-by:** be-dev-20231010-ABCD
**Implements:** Document Tools JSON Schemas, Document Data Model

## Goal
Implement the Document repository and register the MCP tools for Document CRUD operations.

## Scope
- **In:** `internal/domain/document.go`, `internal/repo/document_repo.go`, `internal/handler/document_tools.go`.
- **Out:** Database schema migration (already handled).

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US003_be_unit_tests.md`: all Document-related UT-* and IT-*.

## Implementation notes
- Implement repository methods for Document CRUD.
- Register tools: `create_document`, `get_document`, `update_document`, `delete_document`, `list_documents`.
- Ensure responses exactly match the Document Tools JSON Schemas in the architecture.
- Handle `projectId` foreign key validations appropriately.

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Review log

### Review pass 1 — 2024-05-18 — verdict: approved
- All tests (`go test ./...` and `go vet ./...`) passed.
- Implementation matches the architecture's exact JSON shape for Document Tools.
- Good use of `domain.Document` and appropriate public doc comments.
