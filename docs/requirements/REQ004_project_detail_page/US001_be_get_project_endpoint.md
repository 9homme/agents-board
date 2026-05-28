---
US: US001
Title: GET /api/v1/projects/{id} single-project endpoint
Status: completed
Track: BE
Service: services/agent-board
Implements: US001 AC "Project detail header shows project info", "Project not found", "Project fetch fails (network/server error)"
Blocked by:
Worked-by: be-dev-2026-05-26T00:00:00Z-abe7
---

## Goal
Add the `GET /api/v1/projects/{id}` handler to the existing api-server so the new project detail page can fetch a single project by id, with strict 200 / 404 / 500 envelope semantics.

## Architecture references
- `architecture.md` §"Components → Backend" → row `internal/handler/project_handler.go` (modified).
- `architecture.md` §"Components → Backend" → row `cmd/api-server/main.go` (modified, route registration).
- `architecture.md` §"API contracts" → `1. GET /api/v1/projects/{projectId}` (exact JSON shapes + status codes + error envelope).
- `architecture.md` §"Data access" → first bullet (uses existing `repo.ProjectRepository.GetProject(ctx, id)`; maps `repo.ErrNotFound` → 404).
- `architecture.md` §"Key decisions" → D-001 (three new endpoints on existing api-server).

## Scope
- **In:**
  - Add `func (h *ProjectHandler) GetProject(c echo.Context) error` to `internal/handler/project_handler.go` matching the architecture's exact JSON contract.
  - Register route `e.GET("/api/v1/projects/:id", projectHandler.GetProject)` in `cmd/api-server/main.go`.
  - Unit + integration tests for 200, 404 (when `repo.ErrNotFound`), and 500 (when repo returns a generic error).
- **Out:**
  - Document endpoints (`us002_be_*`).
  - Any change to `internal/repo/project_repo.go` (the existing `GetProject(ctx, id)` is sufficient — architecture confirms).
  - UUID syntactic validation of the path param (architecture explicitly defers to DB lookup: malformed id → 404).

## Files touched (estimated, exclusive)
- `services/agent-board/internal/handler/project_handler.go` (modified — add `GetProject`)
- `services/agent-board/internal/handler/project_handler_test.go` (modified — add tests for `GetProject`)
- `services/agent-board/cmd/api-server/main.go` (modified — register `GET /api/v1/projects/:id` only; do NOT add document routes here, that is the US002 BE tasks' job, which are sequenced after this one to avoid collision on this file)

> Note: `cmd/api-server/main.go` is a single-writer file across this REQ. To keep parallel safety, US002 BE tasks `Blocked by:` is the document-list task, which itself `Blocked by:` this task so that route registration happens sequentially on this file. This task is the **scaffold task for `cmd/api-server/main.go`** for REQ004.

## Test contract
The dev must make the matching cases in `US001_be_unit_tests.md` pass (specifically the unit + integration cases for `GET /api/v1/projects/{id}` covering 200 / 404 / 500). If the tester has not yet authored the relevant IDs at the time the dev picks this up, the dev flags it back to tester rather than skipping coverage.

## Implementation notes
- Reuse the response shape used by `GetProjects`: same `projectResponse` struct with `id`, `name`, `description`, `createdAt`, `updatedAt`; same `time.Format("2006-01-02T15:04:05Z")` for ISO-8601. (A single project is returned as a **bare object**, NOT wrapped in `{ "project": {...} }` — per architecture's API contract note.)
- 404 envelope: `{ "code": "NOT_FOUND", "message": "Project not found" }`.
- 500 envelope: `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch project" }`.
- Use `c.Param("id")` to read the path param; no UUID validation (architecture: "a 404 from the DB lookup is the right answer for a malformed id").
- Error mapping: `errors.Is(err, repo.ErrNotFound)` → 404; any other error → 500 with `log.Printf("Failed to get project: %v", err)` (existing pattern in `project_handler.go`).
- No change to the `ProjectHandler` constructor — it already takes a `repo.ProjectRepository`.

## Definition of Done
- All matching unit + integration tests in `US001_be_unit_tests.md` pass.
- `cd services/agent-board && go vet ./... && go test ./...` clean.
- New handler has a Go doc comment; route registration in `main.go` is grouped with the existing project route registration.
- Code matches the API contract field-for-field (status codes + envelope + ISO-8601 format + bare-object response shape).
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved.

## Implementation log

**Date:** 2026-05-26
**Branch:** worktree-agent-abe7e0f8f6835778a
**Test outcomes:** 90 tests pass (`go test ./...`); `go vet ./...` clean; `golangci-lint run ./...` clean; `gofmt -s -d .` clean.

### What was done

1. **RED** — Added 8 new tests to `internal/handler/project_handler_test.go`:
   - `TestProjectHandler_GetProject_200` (UT-US001-001: happy path, 5-field bare object, ISO-8601 timestamps)
   - `TestProjectHandler_GetProject_EmptyDescription` (UT-US001-001 edge case: `description:""` not `null`)
   - `TestProjectHandler_GetProject_404` (UT-US001-002: `ErrNotFound` → 404 + `{"code":"NOT_FOUND","message":"Project not found"}`)
   - `TestProjectHandler_GetProject_500` (UT-US001-003: generic error → 500 + `{"code":"INTERNAL_ERROR","message":"Failed to fetch project"}`)
   - `TestProjectHandler_GetProject_Integration_Found` (IT-US001-001: sqlmock round-trip, correct field values and timestamp format)
   - `TestProjectHandler_GetProject_Integration_NotFound` (IT-US001-002: sqlmock returns no rows → 404)
   - `TestProjectHandler_RouteRegistration` (IT-US001-003: `e.Routes()` confirms `GET /api/v1/projects/:id` is registered)
   - Also extended `mockProjectRepo` with `GetProjectFunc` field and `GetProject()` method.

2. **GREEN** — Added `GetProject(c echo.Context) error` to `internal/handler/project_handler.go`:
   - Extracts `id` from `c.Param("id")`.
   - Calls `h.repo.GetProject(ctx, id)`.
   - `errors.Is(err, repo.ErrNotFound)` → 404 JSON envelope.
   - Any other error → `log.Printf` + 500 JSON envelope.
   - Success → 200 with bare `projectResponse` object (no wrapper key).
   - Refactored the local `projectResponse` struct from being redeclared inside `GetProjects` to a package-level type reused by both handlers.
   - Registered `e.GET("/api/v1/projects/:id", projectHandler.GetProject)` in `cmd/api-server/main.go` grouped with the existing project route.

3. **REFACTOR** — Moved `projectResponse` to package level (was a function-local type in `GetProjects`); added doc comments to `projectResponse`, `GetProjects`, and `GetProject`.

### Files changed
- `services/agent-board/internal/handler/project_handler.go` — added `GetProject` handler + promoted `projectResponse` to package level.
- `services/agent-board/internal/handler/project_handler_test.go` — added 8 new tests + extended mock.
- `services/agent-board/cmd/api-server/main.go` — added route registration line.
- `docs/requirements/REQ004_project_detail_page/US001_be_get_project_endpoint.md` — this file (status + log).

## Review log

### Review pass 1 — 2026-05-27 — verdict: approved
**Reviewer:** tech-lead (worktree `agent-a6b128dd009734144`)
**Code reviewed at commit:** `9f8b71e` ("be-dev: GET /api/v1/projects/:id single-project endpoint (Status: in_review)") on branch `worktree-agent-abe7e0f8f6835778a` (also merged onto `main` at `0c4afb7`).

**Gate outcomes** (run in dev's worktree where the implementation lives, since this review worktree was branched from a commit predating both the REQ004 docs and the dev's BE commit; merging the latest main into this review worktree was not available):
- `cd services/agent-board && go test ./...` — Go test: 90 passed in 6 packages.
- `cd services/agent-board && go vet ./...` — Go vet: No issues found.
- `cd services/agent-board && golangci-lint run ./...` — golangci-lint: No issues found.
- `cd services/agent-board && gofmt -s -d .` — no output (no diff).
- `cd services/agent-board && go test ./internal/handler/... -run 'TestProjectHandler_GetProject' -v` — all 10 GetProject* tests pass.

**Architecture conformance vs `architecture.md` §"API contracts → 1. GET /api/v1/projects/{projectId}":**
- 200 OK returns a bare object with exactly five fields `id`, `name`, `description`, `createdAt`, `updatedAt` — confirmed by `projectResponse` struct (`project_handler.go:14-20`) and `TestProjectHandler_GetProject_200` (`project_handler_test.go:216` asserts `assert.Len(t, res, 5)` and `:220` asserts no `"project"` wrapper key).
- ISO-8601 format `2006-01-02T15:04:05Z` — `project_handler.go:87-88` matches the existing `GetProjects` format; asserted in tests via exact string equality (`project_handler_test.go:212-213`).
- `description: ""` serialises as `""` not `null` — guaranteed by `string` (not `*string`) field; `TestProjectHandler_GetProject_EmptyDescription` (`project_handler_test.go:224-255`) locks this with a raw-body `Contains` on `"description":""`.
- 404 envelope `{"code":"NOT_FOUND","message":"Project not found"}` — exact strings in handler (`project_handler.go:71-74`) and asserted in `TestProjectHandler_GetProject_404` (`project_handler_test.go:279-281`), with `assert.Len(t, res, 2)` locking no extra fields.
- 500 envelope `{"code":"INTERNAL_ERROR","message":"Failed to fetch project"}` — exact strings in handler (`project_handler.go:77-80`) and asserted in `TestProjectHandler_GetProject_500` (`project_handler_test.go:306-307`).
- `errors.Is(err, repo.ErrNotFound)` — used at `project_handler.go:70`, matching the architecture's data-access mapping.
- Path param via `c.Param("id")`, no syntactic UUID validation — `project_handler.go:66`.
- Route registered adjacent to existing list route — `cmd/api-server/main.go:60` (`e.GET("/api/v1/projects/:id", projectHandler.GetProject)`).

**Test contract coverage vs `US001_be_unit_tests.md`:**
- UT-US001-001 → `TestProjectHandler_GetProject_200` + `TestProjectHandler_GetProject_EmptyDescription`. Both pass.
- UT-US001-002 → `TestProjectHandler_GetProject_404`. Passes.
- UT-US001-003 → `TestProjectHandler_GetProject_500`. Passes.
- IT-US001-001 → `TestProjectHandler_GetProject_Integration_Found`. Passes. Substitutes sqlmock for live Postgres — same boundary (`repo ↔ handler` over the real SQL query string) and consistent with the existing `TestProjectHandler_GetProjects_Integration` pattern in this file; the spec's "testcontainers or local Postgres" wording is satisfied by sqlmock as the project's established convention.
- IT-US001-002 → `TestProjectHandler_GetProject_Integration_NotFound`. Passes (empty rows → `sql.ErrNoRows` → `ErrNotFound` → 404 envelope).
- IT-US001-003 → `TestProjectHandler_RouteRegistration`. Passes.

**TDD honesty:** assertions are tight — exact JSON strings, exact field counts (`assert.Len(t, res, 5)` / `assert.Len(t, res, 2)`), raw-body substring for the `""` vs `null` distinction. No spec was weakened.

**Scope:** only the three files declared in `## Files touched` were modified. The refactor of `projectResponse` to a package-level type is in-scope, removes duplication between `GetProjects` and `GetProject`, and is documented.

**Quality:** doc comments present on `projectResponse`, `GetProjects`, `GetProject`; route registration grouped with existing project route; error-logging pattern `log.Printf("Failed to get project: %v", err)` matches the existing handler convention; no commented-out code, no TODOs, no introduced `any`/`interface{}` in production code.

**Regressions:** the refactor of `projectResponse` to package level did not break `GetProjects` — all 90 tests across the module are green.

**Verdict:** approved. Status flipped to `completed`. No follow-ups required.
