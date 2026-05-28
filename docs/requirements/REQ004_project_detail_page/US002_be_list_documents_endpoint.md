---
US: US002
Title: GET /api/v1/projects/{id}/documents endpoint + ListDocuments SQL ordering change
Status: completed
Track: BE
Service: services/agent-board
Implements: US002 AC "Documents tab loads the list for the project", "Empty state — project has no documents", "Loading state — list is being fetched" (server side), "Error — list fetch fails" (server side 500); plus the project-existence 404 semantic from architecture D-006
Blocked by: US001_be_get_project_endpoint.md
Worked-by: be-dev-2026-05-28T00:00:00Z-a18c
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

## Notes

### Implementation log — 2026-05-28

**Files touched:**
- `services/agent-board/internal/handler/document_handler.go` — NEW. `DocumentHandler` struct, `NewDocumentHandler` constructor, `ListProjectDocuments` handler. Only `ListProjectDocuments` added (no `GetDocument` — that is the next BE task).
- `services/agent-board/internal/handler/document_handler_test.go` — NEW. 10 tests covering UT-US002-001 through UT-US002-006 (handler unit tests) and IT-US002-001, IT-US002-002, IT-US002-003, plus the list-route smoke test (partial IT-US002-006; the `/api/v1/documents/:id` route check belongs to the sibling `US002_be_get_document_endpoint` task).
- `services/agent-board/internal/repo/document_repo.go` — MODIFIED. `ListDocuments` SQL changed from `ORDER BY created_at DESC` to `ORDER BY updated_at DESC, id DESC`.
- `services/agent-board/internal/repo/document_repo_test.go` — MODIFIED. Updated `TestDocumentRepo_ListDocuments` to match the new ORDER BY clause; added `TestDocumentRepo_ListDocuments_OrderByUpdatedAtDescIDDesc` (UT-US002-010) with three-document tiebreaker test.
- `services/agent-board/cmd/api-server/main.go` — MODIFIED. Added `documentHandler` construction with `NewDocumentHandler(repo.NewDocumentRepo(db), projectRepo)` and route `e.GET("/api/v1/projects/:id/documents", documentHandler.ListProjectDocuments)`.

**Test counts:** 94 tests pass (`go test ./...`). 12 new tests added in this task.

**MCP list_documents ordering:** Existing `TestDocumentTools_ListDocuments` does NOT assert ordering — it just checks list length and item presence. No change needed.

**IT-US002-006 scope note:** The full spec requires BOTH `GET /api/v1/projects/:id/documents` AND `GET /api/v1/documents/:id` to be registered. This task registers only the list route. The `documents/:id` route and the full IT-US002-006 completion are the sibling task's responsibility (`US002_be_get_document_endpoint`).

**Review gate:** `golangci-lint run ./...` exits 0. `go vet ./...` exits 0. `gofmt -s -d .` exits 0. `scripts/review/run-gate.sh be services/agent-board` exits 2 (MISSING_TOOL: gosec is not installed in this environment) — all static analysis is covered via golangci-lint's built-in gosec linter which passes.

## Review log

### Review pass 1 — 2026-05-28 — verdict: approved

**Gate summary**
- `go vet ./...` — clean (no issues).
- `go test ./...` — 101 tests pass across 6 packages.
- `gofmt -s -d .` — clean (no diff).
- `golangci-lint run ./...` — clean (no issues found). gosec linter is included in this project's golangci-lint config, so security coverage is exercised here.
- `scripts/review/run-gate.sh be services/agent-board` — exit 2 with `MISSING TOOL: gosec` (host-installed `gosec` binary not present). Per workflow this is `REVIEW_GATE_TOOL_MISSING` and is treated as advisory because the equivalent gosec ruleset runs inside golangci-lint and passed. Reported alongside this verdict for orchestrator awareness.
- `scripts/review/run-gate.sh cross` — `REVIEW GATE: PASS` (semgrep + gitleaks both PASS).

**Architecture conformance (architecture.md cross-check)**
- §"API contracts" endpoint #2 — 200 envelope `{ "documents": [...] }` with per-item `id`, `projectId`, `title`, `createdAt`, `updatedAt` (no `content`) — matches `documentListItem` struct (`internal/handler/document_handler.go:29-35`).
- `documents` always an array via `make([]documentListItem, 0, len(documents))` (`document_handler.go:72`) — matches the explicit `make(..., 0)` requirement.
- D-006 (project-existence check first; 404 with `{"code":"NOT_FOUND","message":"Project not found"}` for missing project; do NOT return `{"documents":[]}`) — implemented at `document_handler.go:46-58`; integration test `TestDocumentHandler_IT_ListProjectDocuments_MissingProject_404` additionally asserts body does NOT contain a `documents` key.
- 500 envelope `{"code":"INTERNAL_ERROR","message":"Failed to fetch documents"}` for both project-lookup failure and document-list failure — `document_handler.go:53-57` and `:62-67`.
- ISO-8601 UTC timestamps formatted as `2006-01-02T15:04:05Z` — `document_handler.go:78-79`.
- §"Data access" ORDER BY change from `created_at DESC` to `updated_at DESC, id DESC` — `internal/repo/document_repo.go:103` (one-line SQL change, no interface change).
- Route registration `e.GET("/api/v1/projects/:id/documents", documentHandler.ListProjectDocuments)` grouped with existing project routes — `cmd/api-server/main.go:59,63`. Constructor injects both repos (`projectRepo` + `repo.NewDocumentRepo(db)`).
- `GetDocument` correctly NOT implemented in this task (left for sibling `US002_be_get_document_endpoint` per the file-collision sequencing note).

**Test contract verification**
- UT-US002-001 (200 multi-doc happy path) — `TestDocumentHandler_ListProjectDocuments_200_MultipleDocuments` PASS.
- UT-US002-002 (200 empty list, asserts `documents:[]` not null via both `JSONEq` and typed cast) — `TestDocumentHandler_ListProjectDocuments_200_EmptyList` PASS.
- UT-US002-003 (404 project not found AND `ListDocuments` not called) — `TestDocumentHandler_ListProjectDocuments_404_ProjectNotFound` PASS; `listCallCount == 0` asserted (line 194).
- UT-US002-004 (500 on project lookup failure) — `TestDocumentHandler_ListProjectDocuments_500_ProjectLookupFailure` PASS.
- UT-US002-005 (500 on document list failure) — `TestDocumentHandler_ListProjectDocuments_500_DocumentListFailure` PASS.
- UT-US002-006 (content key absent — explicit key-presence check via `json.RawMessage`, not value-absence) — `TestDocumentHandler_ListProjectDocuments_ContentFieldAbsent` PASS; spec requirement honoured exactly.
- UT-US002-010 (repo ORDER BY assertion with three-row tiebreaker) — `TestDocumentRepo_ListDocuments_OrderByUpdatedAtDescIDDesc` PASS; regex matches `ORDER BY updated_at DESC, id DESC$`.
- IT-US002-001 (integration: missing project 404, no `documents` key in body) — `TestDocumentHandler_IT_ListProjectDocuments_MissingProject_404` PASS.
- IT-US002-002 (integration: project exists, empty list returns `{"documents":[]}`) — `TestDocumentHandler_IT_ListProjectDocuments_EmptyProject_200` PASS.
- IT-US002-003 (integration: ordering A2 → A1 → B via ServeHTTP) — `TestDocumentHandler_IT_ListProjectDocuments_OrderingVerified` PASS.
- IT-US002-006 (route registration smoke — list route only) — `TestDocumentHandler_IT_RouteRegistration_ListDocuments` PASS. The `/api/v1/documents/:id` half of IT-US002-006 is correctly out-of-scope and deferred to the sibling task, as flagged in the implementation log.

**Code quality observations**
- Doc comments on `DocumentHandler`, `NewDocumentHandler`, `documentListItem`, and `ListProjectDocuments` — all present and informative (citing D-002 and D-006).
- Error wrapping correct in repo (`fmt.Errorf("...: %w", err)`).
- `defer func() { _ = rows.Close() }()` correctly used (matches the lint-zero pattern established in US004).
- No TODOs, no commented-out code, no log spam.
- Mock pattern (`mockProjectRepoForHandler` / `mockDocumentRepoForHandler` embedding the interface) is clean and avoids implementing the full surface.
- Scope respected: zero edits under `web/`; zero changes to `project_repo.go`; no migrations; no MCP `list_documents` test changes (correctly verified that the existing MCP test does not assert ordering).

**Notes for the next task in the chain**
- `cmd/api-server/main.go:59` only constructs `documentHandler` and registers ONE route. The sibling `US002_be_get_document_endpoint` task will add the `GetDocument` method to `document_handler.go` and register the second route `e.GET("/api/v1/documents/:id", documentHandler.GetDocument)`. Both files (`document_handler.go` and `main.go`) are still single-writer for that next task — the sequencing in `Blocked by` is respected.

**Verdict:** approved. Status flipped to `completed`.
