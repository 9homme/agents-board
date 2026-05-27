---
US: US002
Title: GET /api/v1/projects/{id}/documents endpoint + ListDocuments SQL ordering change
Status: pending
Track: BE
Service: services/agent-board
Implements: US002 AC "Documents tab loads the list for the project", "Empty state — project has no documents", "Loading state — list is being fetched" (server side), "Error — list fetch fails" (server side 500); plus the project-existence 404 semantic from architecture D-006
Blocked by: US001_be_get_project_endpoint.md
Worked-by:
---

## Goal
Add the `GET /api/v1/projects/{id}/documents` handler that returns the project's documents (metadata-only — no `content`) ordered by `updated_at DESC, id DESC`, returning 404 when the project itself does not exist. Implements the underlying repo SQL ordering change in the same task because the API ordering AC is meaningless without it.

## Architecture references
- `architecture.md` §"Components → Backend" → row `internal/handler/document_handler.go` (new, both list + get handlers — this task implements `ListProjectDocuments` only; the `GetDocument` half is the next BE task, sequenced via Blocked by so both halves don't write to this file in parallel).
- `architecture.md` §"Components → Backend" → row `internal/repo/document_repo.go` (modified — `ListDocuments` SQL ordering change from `ORDER BY created_at DESC` to `ORDER BY updated_at DESC, id DESC`).
- `architecture.md` §"Components → Backend" → row `cmd/api-server/main.go` (modified — register `GET /api/v1/projects/:id/documents` and construct `documentHandler` with `repo.NewDocumentRepo(db)`).
- `architecture.md` §"API contracts" → endpoint #2 `GET /api/v1/projects/{projectId}/documents`.
- `architecture.md` §"Data access" → second bullet (step 1 verify project exists via `repo.ProjectRepository.GetProject`; step 2 `repo.DocumentRepository.ListDocuments`; ordering diff explicit).
- `architecture.md` §"Key decisions" → D-002 (split metadata-list from full-content detail) and D-006 (404 vs `{ documents: [] }` for missing project).
- `architecture.md` §"Risks & open questions" → MCP `list_documents` ordering change is acceptable behavior shift (called out for human awareness; no MCP test change required).

## Scope
- **In:**
  - Create `services/agent-board/internal/handler/document_handler.go` with:
    - A `DocumentHandler` struct holding `documentRepo repo.DocumentRepository` AND `projectRepo repo.ProjectRepository` (needed for the project-existence check per D-006).
    - A `NewDocumentHandler(documentRepo, projectRepo)` constructor.
    - `func (h *DocumentHandler) ListProjectDocuments(c echo.Context) error` — verifies project exists via `projectRepo.GetProject(ctx, projectID)` (404 envelope on `ErrNotFound`); then `documentRepo.ListDocuments(ctx, projectID)`; returns `{ "documents": [...] }` with metadata only (NO `content` field), `documents` always an array (use `make([]documentListItem, 0)`); each item carries `id`, `projectId`, `title`, `createdAt`, `updatedAt` formatted as ISO-8601 UTC `2006-01-02T15:04:05Z`. (The `GetDocument` handler is added in the next BE task.)
  - Modify `services/agent-board/internal/repo/document_repo.go` `ListDocuments` SQL: change `ORDER BY created_at DESC` to `ORDER BY updated_at DESC, id DESC`. No interface change.
  - Update / add tests in `services/agent-board/internal/repo/document_repo_test.go` to assert the new ordering (insert ≥3 fixtures with varying `updated_at` and same `updated_at` plus different `id` to prove the tiebreaker).
  - Add handler tests (unit + integration via `httptest`) covering 200 with non-empty list, 200 with empty list `{ "documents": [] }`, 404 on unknown project, and 500 on a downstream repo error.
  - Modify `cmd/api-server/main.go`: construct `documentHandler := handler.NewDocumentHandler(repo.NewDocumentRepo(db), projectRepo)`; register `e.GET("/api/v1/projects/:id/documents", documentHandler.ListProjectDocuments)`.
- **Out:**
  - The single-document endpoint (`GET /api/v1/documents/{id}`) — `us002_be_get_document_endpoint`.
  - Any change to `internal/repo/project_repo.go`.
  - Stripping `content` from the repo SELECT statement (architecture: "the over-fetch cost is acceptable at current scale" — explicitly out of scope).
  - Adding the composite index `(project_id, updated_at DESC, id DESC)` (architecture: explicitly NOT in scope for this requirement).
  - Updating MCP `list_documents` tests for the new ordering — if the existing test currently asserts the old `created_at DESC` order, fix that test in this task; if it doesn't, no change required. The behavior change is intentional per architecture D-001 consequences.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/handler/document_handler.go` (new — only `ListProjectDocuments` and constructor in this task; the next task adds `GetDocument` to the same file)
- `services/agent-board/internal/handler/document_handler_test.go` (new — tests for `ListProjectDocuments`)
- `services/agent-board/internal/repo/document_repo.go` (modified — `ListDocuments` ORDER BY change)
- `services/agent-board/internal/repo/document_repo_test.go` (modified — add ordering assertions)
- `services/agent-board/cmd/api-server/main.go` (modified — add `documentHandler` construction + one route registration)
- `services/agent-board/internal/handler/handler_test.go` (potentially modified if it holds shared fixtures the new tests reuse — dev's call)
- Possibly `services/agent-board/internal/handler/document_tools_test.go` — only if the existing MCP `list_documents` test asserts the OLD ordering and now needs updating; if not, leave untouched.

> Sequenced after `US001_be_get_project_endpoint.md` because both write to `cmd/api-server/main.go`. Sequenced before `US002_be_get_document_endpoint.md` for the same reason and because both write to `internal/handler/document_handler.go`.

## Test contract
The dev must make the matching cases in `US002_be_unit_tests.md` pass — covering: 200 happy path with ordering assertion (`updated_at DESC, id DESC`), 200 empty-list `{ "documents": [] }` (never `null`), 404 on missing project with envelope `{ "code": "NOT_FOUND", "message": "Project not found" }`, 500 on repo error with envelope `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch documents" }`, and the repo-level ordering assertion. (UT-* / IT-* IDs assigned by tester.) If the tester has not yet authored the relevant IDs at the time the dev picks this up, the dev flags it back to tester rather than skipping coverage.

## Implementation notes
- Mirror the existing `GetProjects` shape: a private response struct per item with json tags `id`, `projectId`, `title`, `createdAt`, `updatedAt` (NO `content`); `time.Format("2006-01-02T15:04:05Z")` for both timestamps.
- The handler ignores `domain.Document.Content` even though the repo loads it (architecture explicitly accepts this over-fetch).
- 404 path: when `repo.ProjectRepository.GetProject` returns `repo.ErrNotFound`, respond 404 with `{ "code": "NOT_FOUND", "message": "Project not found" }`. Architecture D-006 is explicit that this is the project-not-found body even on the documents path.
- 500 path: any other error → `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch documents" }` + `log.Printf("Failed to list documents: %v", err)`.
- Initialise the slice with `make([]documentListItem, 0)` so an empty response serialises as `[]`, not `null` (architecture explicit + existing pattern in `GetProjects`).
- The repo SQL change is one-line. The test must cover both axes: (a) two rows with different `updated_at` → newer first; (b) two rows with the same `updated_at` → larger `id` first (tiebreaker).

## Definition of Done
- All matching unit + integration tests in `US002_be_unit_tests.md` pass.
- `cd services/agent-board && go vet ./... && go test ./...` clean.
- New handler struct + constructor + method have Go doc comments; route registration in `main.go` is grouped with the existing project routes.
- API contract field-for-field correct (status codes + envelope + ISO-8601 format + `documents` always an array).
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved.

## Review log
(left for tech-lead review pass entries)
