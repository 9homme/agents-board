---
US: US002
Title: GET /api/v1/documents/{id} single-document endpoint
Status: pending
Track: BE
Service: services/agent-board
Implements: US002 AC "Selecting a document loads its content into the previewer" (server side), "Deep-link to a specific document" (server side), "Deep-link to a document that doesn't exist for this project" (404 surface), "Loading state — content is being fetched" (server side), "Error — content fetch fails" (server side 500)
Blocked by: US002_be_list_documents_endpoint.md
Worked-by:
---

## Goal
Add the `GET /api/v1/documents/{id}` handler that returns a single document including its full markdown `content`, with strict 200 / 404 / 500 envelope semantics matching the architecture's API contract.

## Architecture references
- `architecture.md` §"Components → Backend" → row `internal/handler/document_handler.go` (modified — add `GetDocument` to the same file the previous BE task created).
- `architecture.md` §"Components → Backend" → row `cmd/api-server/main.go` (modified — register `GET /api/v1/documents/:id`).
- `architecture.md` §"API contracts" → endpoint #3 `GET /api/v1/documents/{documentId}`.
- `architecture.md` §"Data access" → third bullet (uses existing `repo.DocumentRepository.GetDocument`; maps `repo.ErrNotFound` → 404; returns `content` directly).
- `architecture.md` §"Key decisions" → D-002 (split metadata-list from full-content detail).

## Scope
- **In:**
  - Add `func (h *DocumentHandler) GetDocument(c echo.Context) error` to `internal/handler/document_handler.go` (the file created by the prior BE task).
  - Register route `e.GET("/api/v1/documents/:id", documentHandler.GetDocument)` in `cmd/api-server/main.go`.
  - Handler tests (unit + integration via `httptest`) covering 200 with all fields present (including `content`), 404 when `repo.ErrNotFound`, and 500 on a generic repo error.
- **Out:**
  - The list endpoint (`GET /api/v1/projects/{id}/documents`) — already implemented by the prior BE task.
  - Any change to `internal/repo/document_repo.go` (existing `GetDocument(ctx, id)` is sufficient).
  - Schema migrations.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/handler/document_handler.go` (modified — add `GetDocument` method)
- `services/agent-board/internal/handler/document_handler_test.go` (modified — add tests for `GetDocument`)
- `services/agent-board/cmd/api-server/main.go` (modified — add one route registration)

> Sequenced after `US002_be_list_documents_endpoint.md` because both write to `internal/handler/document_handler.go` AND to `cmd/api-server/main.go`.

## Test contract
The dev must make the matching cases in `US002_be_unit_tests.md` pass — covering: 200 happy path with full response body (`id`, `projectId`, `title`, `content`, `createdAt`, `updatedAt` all present + ISO-8601 timestamps), 200 with `content == ""` (architecture: content MAY be empty string, never null), 404 with envelope `{ "code": "NOT_FOUND", "message": "Document not found" }`, 500 with envelope `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch document" }`. (UT-* / IT-* IDs assigned by tester.) If the tester has not yet authored the relevant IDs at the time the dev picks this up, the dev flags it back to tester rather than skipping coverage.

## Implementation notes
- Response shape is a **bare document object** (not wrapped in `{ "document": {...} }`) — consistent with the project singular endpoint per architecture's API-contract note on plural-collection vs singular-resource convention.
- Build a private `documentResponse` struct with json tags `id`, `projectId`, `title`, `content`, `createdAt`, `updatedAt`. Map from `*domain.Document` with `time.Format("2006-01-02T15:04:05Z")` for both timestamps.
- Error mapping: `errors.Is(err, repo.ErrNotFound)` → 404 with `{ "code": "NOT_FOUND", "message": "Document not found" }`; any other error → 500 with `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch document" }` + `log.Printf("Failed to get document: %v", err)`.
- Use `c.Param("id")` for the path param; no UUID validation (architecture: malformed id → 404 from DB lookup).
- The 200 body's `content` field is a raw markdown string and MAY be `""`. Do not coerce empty to null; do not omit the field.

## Definition of Done
- All matching unit + integration tests in `US002_be_unit_tests.md` pass.
- `cd services/agent-board && go vet ./... && go test ./...` clean.
- New method has a Go doc comment.
- API contract field-for-field correct (status codes + envelope + ISO-8601 format + bare-object response shape including `content`).
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved.

## Review log
(left for tech-lead review pass entries)
