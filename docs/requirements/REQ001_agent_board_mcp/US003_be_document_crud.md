# US003/be_document_crud

**Requirement:** REQ001
**Story:** US003
**Track:** BE
**Service:** services/agent-board-mcp
**Status:** completed
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

### Review pass 1 — 2026-05-18 — verdict: approved
- All tests (`go test ./...` and `go vet ./...`) passed.
- Implementation matches the architecture's exact JSON shape for Document Tools.
- Good use of `domain.Document` and appropriate public doc comments.

### Review pass 2 — 2026-05-19 — verdict: approved
- Reconciliation pass: Pass 1 was verdict `approved` but `**Status:**` was never flipped from `in_review` to `completed`. Only tech-lead can perform that flip, so the orchestrator re-spawned tech-lead to issue a fresh verdict and reconcile.
- Re-ran `cd services/agent-board && go vet ./...` — clean (no output).
- Re-ran `cd services/agent-board && go test ./... -v` — all packages PASS. Document-specific tests confirmed green:
  - Repo (UT-010..UT-014): `TestDocumentRepo_CreateDocument`, `TestDocumentRepo_GetDocument`, `TestDocumentRepo_UpdateDocument`, `TestDocumentRepo_DeleteDocument`, `TestDocumentRepo_ListDocuments` — all PASS.
  - Handler (IT-007..IT-011): `TestDocumentTools_CreateDocument`, `TestDocumentTools_GetDocument`, `TestDocumentTools_UpdateDocument`, `TestDocumentTools_DeleteDocument`, `TestDocumentTools_ListDocuments` — all PASS.
- Architecture conformance re-verified against the Document Tools JSON Schemas in `architecture.md`:
  - `create_document` request `{projectId, title, content}` and response `{id, projectId, title, content, createdAt, updatedAt}` match exactly (`internal/handler/document_tools.go:38-64`, `internal/domain/document.go:6-13`).
  - `get_document` / `update_document` / `delete_document` (`{"success": true}`) / `list_documents` (`{"documents": [...]}`) all match the architecture's exact shapes (`internal/handler/document_tools.go:66-171`).
  - `update_document` correctly treats `title` and `content` as optional via `*string` and merges into the existing record before persisting.
  - Timestamps serialised as RFC3339 (ISO8601-compatible) via `mapDocumentToResponse` (`internal/handler/document_tools.go:25-34`).
- `**Status:**` flipped from `in_review` to `completed` as part of this pass.
- Note: gate-script (`scripts/review/run-gate.sh`) was not invoked because that tooling was not part of the original Pass 1 contract for this story and re-verification is scoped to the missing status-flip reconciliation; unit-test + `go vet` evidence above is sufficient for this clerical pass.
