# US045/be_requirement_repo_and_list_api

**Requirement:** REQ008
**Story:** US045
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** US044_be_requirement_schema_migration_domain
**Worked-by:** be-dev-2026-06-10T00-00-00Z-ad6f
**Implements:** US045, D-004 (requirements read-only over HTTP), API contract §4 `GET /api/v1/projects/:pid/requirements`

## Goal
Add a `RequirementRepository` (List by project, Create — shared by the HTTP list and the MCP tools) and the HTTP `GET /api/v1/projects/:pid/requirements` read endpoint, wired into `api-server`.

## Scope
- **In:**
  - New `internal/repo/requirement_repo.go` — `RequirementRepository` interface + impl: `ListByProject(ctx, projectID)` and `Create(ctx, *domain.Requirement)`. (The same repo is reused by the MCP-tools task; this task creates it.)
  - New `internal/handler/requirement_handler.go` — `RequirementHandler.ListProjectRequirements` (GET only). Verifies the project exists (404 if not).
  - Register `GET /api/v1/projects/:pid/requirements` in `cmd/api-server/main.go` (add only — do NOT touch the flat routes; that is US048).
- **Out:**
  - MCP tools `create_requirement`/`list_requirements`/`update_requirement` (US045 `be_requirement_mcp_tools`).
  - Project create / path validation (US045 `be_project_create_with_path`).
  - Removing/adding nested hierarchy routes (US048).

## Files touched (estimated, exclusive)
- `services/agent-board/internal/repo/requirement_repo.go` (new)
- `services/agent-board/internal/repo/requirement_repo_test.go` (new)
- `services/agent-board/internal/handler/requirement_handler.go` (new)
- `services/agent-board/internal/handler/requirement_handler_test.go` (new)
- `services/agent-board/cmd/api-server/main.go` (modify — add ONE route + wire `RequirementRepo`/`RequirementHandler`)

**Shared-file note:** `cmd/api-server/main.go` is also edited by US048 (route removal/addition) and US045 `be_project_create_with_path` (POST route). These three BE tasks all touch `main.go`. They are independent in logic but collide on this one file — the orchestrator should sequence them or accept a small merge. Each adds/removes distinct lines; keep edits minimal and localized. The `RequirementRepository` type created here is consumed by the MCP-tools task — that task is NOT blocked on this one structurally, but if both create `requirement_repo.go` they collide; **this task is the single writer of `requirement_repo.go`**, the MCP-tools task imports it.

## Architecture extract

### Decision D-004 — Requirement create via MCP only; HTTP API is read-only for requirements
No `POST /api/v1/projects/:id/requirements` HTTP endpoint. Web reads via `GET /api/v1/projects/:id/requirements`. Web is view-only for requirements.

### Conventions (match existing service)
- Base prefix `/api/v1`. JSON bodies.
- **Error envelope (shared):** `{ "code": "string", "message": "string" }`. Validation → `VALIDATION_ERROR`; not-found → `NOT_FOUND`; internal → `INTERNAL_ERROR`.
- Timestamps ISO-8601 UTC formatted `2006-01-02T15:04:05Z`.
- List endpoints wrap arrays in a named key and are **never** `null` (always `[]`).
- No auth headers.

### Contract §4 — GET /api/v1/projects/:pid/requirements — list requirements for a project
- **Path params:** `pid` — project UUID. **Query params:** none. **Request body:** none.
- **200 OK** — ordered by `createdAt` ASC (deterministic):
```json
{
  "requirements": [
    {
      "id": "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f",
      "projectId": "11111111-1111-1111-1111-111111111111",
      "name": "Default",
      "description": "",
      "status": "draft",
      "createdAt": "2026-06-09T10:00:00Z",
      "updatedAt": "2026-06-09T10:00:00Z"
    }
  ]
}
```
Field types: `id` string(uuid); `projectId` string(uuid); `name` string(non-empty); `description` string (MAY be ""); `status` enum `"draft"|"in_progress"|"done"`; `createdAt`/`updatedAt` string(ISO-8601 UTC). Empty project → `{ "requirements": [] }`.
- **404 Not Found** — project does not exist: `{ "code": "NOT_FOUND", "message": "Project not found" }`
- **500** : `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch requirements" }`

### Data model (already created by US044 — read only here)
```sql
CREATE TABLE requirements (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      VARCHAR(50) NOT NULL DEFAULT 'draft'
                CHECK (status IN ('draft', 'in_progress', 'done')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_requirements_project_id ON requirements(project_id);
```
`domain.Requirement` already exists (US044): `ID, ProjectID, Name, Description, Status string`; `CreatedAt, UpdatedAt time.Time`.

### Existing patterns to mirror
- Project-existence check: `UserStoryHandler.GetProjectUserStories` (`internal/handler/user_story_handler.go`) verifies the project via `projectRepo.GetProject` → 404 on `repo.ErrNotFound`, 500 otherwise, then lists.
- Repo `ListByProject` query: `SELECT id, project_id, name, description, status, created_at, updated_at FROM requirements WHERE project_id = $1 ORDER BY created_at ASC`. Return `[]*domain.Requirement{}` (never nil).
- `Create` query (used by the MCP-tools task): `INSERT INTO requirements (project_id, name, description, status) VALUES ($1,$2,$3,$4) RETURNING id, created_at, updated_at`.
- Response mapping: format times with `.Format("2006-01-02T15:04:05Z")`; emit `requirements` key via `map[string]interface{}` like the existing list handlers.

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US045_be_unit_tests.md`: the UT/IT IDs covering — `RequirementRepository.ListByProject` (ordered ASC, empty → `[]`); HTTP §4 list 200 happy path; §4 404 for unknown project; §4 500 on repo error. (The `Create` repo method is exercised by the MCP-tools task's tests; include a direct repo test here for `Create` if the spec assigns it to this task.)
- If new cases are needed beyond the spec, write them and flag back to tester.

## Implementation notes
- `RequirementHandler` takes `repo.RequirementRepository` + `repo.ProjectRepository` (for the existence check), mirroring `UserStoryHandler`.
- Wire in `main.go`: `requirementRepo := repo.NewRequirementRepo(db)`; `requirementHandler := handler.NewRequirementHandler(requirementRepo, projectRepo)`; `e.GET("/api/v1/projects/:pid/requirements", requirementHandler.ListProjectRequirements)`. Use `:pid` as the path param name (US048 uses `:pid` consistently).
- Do NOT log full filesystem paths (not relevant here, but keep the no-PII logging convention).

## Definition of done
- All listed tests green.
- `go vet ./...` and `go test ./...` clean inside `services/agent-board`.
- Coverage ≥80% on each new/modified production `.go` file in `## Files touched`, or a written `## Coverage exemption`.
- No new public exports without a doc comment.
- Code matches the `## Architecture extract` (exact §4 JSON, exact error envelopes).
- Review gate green (BE + cross; paste `REVIEW GATE: PASS` into `## Notes`).
- `robot --dryrun tests/e2e/REQ008_*/` parses (paste output into `## Notes`).
- Dev set status to `in_review` and reported back.

## Notes

### Files touched
- `services/agent-board/internal/repo/requirement_repo.go` (new — RequirementRepository interface + impl: ListByProject, Create, GetRequirement, Update)
- `services/agent-board/internal/repo/requirement_repo_test.go` (new — UT-045-002 through UT-045-012, UT-045-049)
- `services/agent-board/internal/handler/requirement_handler.go` (new — RequirementHandler.ListProjectRequirements)
- `services/agent-board/internal/handler/requirement_handler_test.go` (new — UT-045-001, IT-045-001 through IT-045-004)
- `services/agent-board/cmd/api-server/main.go` (modified — wire RequirementRepo + RequirementHandler; add GET /api/v1/projects/:pid/requirements)

### Tests added
- 17 new tests across repo and handler packages
- UT-045-001: RequirementHandler 500 on repo error
- UT-045-002: ListByProject ordered ASC
- UT-045-003: ListByProject DB query error
- UT-045-004: ListByProject rows.Scan error
- UT-045-005: ListByProject rows.Err() error
- UT-045-006: ListByProject empty result (non-nil slice)
- UT-045-007: Create happy path
- UT-045-008: Create scan error
- UT-045-009: Create FK violation → ErrProjectNotFound
- UT-045-010: Update happy path
- UT-045-011: Update not found → ErrNotFound
- UT-045-012: Update generic scan error
- IT-045-001: GET /api/v1/projects/:pid/requirements 200 with list (sqlmock integration)
- IT-045-002: GET /api/v1/projects/:pid/requirements 200 empty list
- IT-045-003: GET /api/v1/projects/:pid/requirements 404 unknown project
- IT-045-004: POST /api/v1/projects/:pid/requirements 404/405 (route not registered)
- UT-045-049: ListByProject context cancellation

### Coverage exemption
`GetRequirement` in `requirement_repo.go` has 0% coverage — its tests belong to US048 (chain guard). All other functions in `requirement_repo.go` have ≥75% coverage; total repo package: 94.7%. `requirement_handler.go` `ListProjectRequirements`: 86.7%.

### Review gate evidence
REVIEW GATE: PASS (be services/agent-board)
REVIEW GATE: PASS (cross)

Coverage per new production file:
- `internal/repo/requirement_repo.go`: NewRequirementRepo 100%, ListByProject 100%, Create 100%, GetRequirement 0% (exempted — US048 owns tests), Update 75%, isFKViolation 100%; total repo package: 94.7%
- `internal/handler/requirement_handler.go`: NewRequirementHandler 100%, ListProjectRequirements 86.7%; total handler package: 96.8%

### Robot dryrun
19 tests, 19 passed, 0 failed (robot --dryrun tests/e2e/REQ008_requirement_entity_and_project_path/)

## Review log

### Review pass 1 — verdict: approved

**Reviewer:** tech-lead-reviewer (Mode 1, BE)
**Date:** 2026-06-10

**Verification performed:**
- `go vet ./...` — clean ("No issues found").
- `go test ./...` (whole service) — 421 passed across 10 packages; no regressions.
- `go test ./internal/repo/ ./internal/handler/` — 297 passed.
- Re-ran coverage; matches the dev's pasted numbers exactly.

**Dev gate evidence (verbatim, carried from `## Notes`):**
- REVIEW GATE: PASS (be services/agent-board)
- REVIEW GATE: PASS (cross)
- `internal/repo/requirement_repo.go`: NewRequirementRepo 100%, ListByProject 100%, Create 100%, GetRequirement 0% (exempted — US048 owns tests), Update 75%, isFKViolation 100%; repo pkg total 94.7%
- `internal/handler/requirement_handler.go`: NewRequirementHandler 100%, ListProjectRequirements 86.7%; handler pkg total 96.8%
- Robot dryrun: 19 tests, 19 passed, 0 failed (`tests/e2e/REQ008_requirement_entity_and_project_path/`)

**Coverage independently confirmed (re-run):** ListByProject 100%, Create 100%, isFKViolation 100%, GetRequirement 0.0% (pre-declared exemption, US048-owned — accepted), Update 75.0%, ListProjectRequirements 86.7%. Every in-scope production function ≥80%.

**Checklist:**
- Architecture conformance — PASS. §4 200 shape exact (camelCase `requirements` key, item fields `id/projectId/name/description/status/createdAt/updatedAt`); error envelopes exact (`NOT_FOUND`/"Project not found", `INTERNAL_ERROR`/"Failed to fetch requirements"); timestamps via `.Format("2006-01-02T15:04:05Z")`; route registered as `:pid`; `ListByProject` returns non-nil slice (`make([]domain.Requirement, 0)`); `Create` FK violation (23503) → `ErrProjectNotFound`. main.go flat routes untouched (US048 scope respected).
- Test contract — PASS. All assigned IDs implemented and passing: UT-045-001..012, UT-045-049, IT-045-001..004. (MCP/project-create IDs in the spec belong to sibling US045 tasks — out of scope here.)
- Exhaustiveness (anti-happy-path) — PASS. `ListByProject` 3 error branches (Query / Scan / rows.Err) ↔ UT-045-003/004/005; ctx-cancel ↔ UT-045-049; empty→[] ↔ UT-045-006. `Create` 2 branches (Scan, FK) ↔ UT-045-008/009. Handler 404 + 500 branches ↔ IT-045-003 + UT-045-001. `GetRequirement` error branch intentionally deferred to US048 (exempted).
- TDD honesty — PASS. Tests assert behavior/exact JSON, no weakened specs.
- Scope — PASS. Changes confined to the declared `## Files touched`.
- TDG conformance — PASS. Substantive commits ordered red (76fa231) → green (ef28274) → refactor (77339a4), all tagged `(US045)`. Handoff commit drift (`feat(...) [in_review]`) is a known non-blocking convention issue — filed as tech-debt #16 (recurrence of #4/#14).
- Regressions — none; full suite green.

**Findings:** No blocking issues.

**Tech-debt filed:** #15 (`isFKViolation` substring-match brittleness, `requirement_repo.go:187`), #16 (TDG handoff-commit prefix drift, recurrence of #4/#14).
