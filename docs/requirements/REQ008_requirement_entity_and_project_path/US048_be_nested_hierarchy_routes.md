# US048/be_nested_hierarchy_routes

**Requirement:** REQ008
**Story:** US048
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** US044_be_requirement_schema_migration_domain
**Worked-by:**
**Implements:** US048, D-009 (full canonical hierarchy; flat routes removed), API contracts §6–§11, Breaking-changes (B) removed-routes table

## Goal
Migrate the read API to the full canonical hierarchy `Project → Requirement → UserStory → Task` (and `Project → Requirement → Document`): register the 6 new nested routes with full ownership-chain guards, add `requirementId` to story/task/document response shapes, and delete the 8 flat/shorthand routes from the router.

## Scope
- **In:**
  - Register 6 new nested GET routes in `cmd/api-server/main.go` and **delete** the 8 flat/shorthand routes.
  - Add handler methods (or adapt existing ones) for the nested routes, each performing the full ownership-chain guard top-down, reusing existing fetch + response-mapping logic.
  - Add `requirementId` to the user-story list/detail and document list/detail response shapes (story/document items). Task shapes are unchanged.
  - Repo support for requirement-scoped listing + chain lookups: `RequirementRepo.GetRequirement` (for `requirement.project_id == :pid`), `UserStoryRepo.ListByRequirement` + ensure `GetUserStory` returns `RequirementID`, `DocumentRepo.ListByRequirement` + `GetDocument` returns `RequirementID`. Add SELECT columns for `requirement_id` where needed.
- **Out:**
  - Write verbs (POST/PUT/DELETE) — read surface only.
  - Project create / path (US045 `be_project_create_with_path`).
  - MCP tools (US045 `be_requirement_mcp_tools`) — MCP is unaffected by route changes.
  - FE client migration to the new paths (US046/US047 own their own FE fetchers).

## Files touched (estimated, exclusive)
- `services/agent-board/cmd/api-server/main.go` (modify — remove 8 routes, add 6; wire repos/handlers)
- `services/agent-board/internal/handler/user_story_handler.go` (modify — requirement-scoped list + `requirementId` in item shape)
- `services/agent-board/internal/handler/user_story_detail_handler.go` (modify — hierarchy detail + tasks under chain guard; `requirementId` in detail shape)
- `services/agent-board/internal/handler/document_handler.go` (modify — requirement-scoped list/detail + `requirementId`)
- `services/agent-board/internal/handler/requirement_handler.go` (modify — add `GetRequirement`-based chain helper if hosted here; coordinate with US045 owner of this file)
- `services/agent-board/internal/repo/user_story_repo.go` (modify — `ListByRequirement`; `GetUserStory`/list SELECT add `requirement_id`)
- `services/agent-board/internal/repo/document_repo.go` (modify — `ListByRequirement`; SELECTs add `requirement_id`)
- `services/agent-board/internal/repo/requirement_repo.go` (modify — `GetRequirement` if not added by US045)
- corresponding `*_test.go` for each handler/repo above

**Shared-file note / collision warning:**
- `cmd/api-server/main.go` is also edited by US045 `be_requirement_repo_and_list_api` (adds the §4 GET) and `be_project_create_with_path` (adds §3 POST). This task removes 8 routes and adds 6. **High collision on `main.go`.** Recommend the orchestrator run this task after both US045 route-adding tasks have merged, OR accept a merge resolution. The §4 `GET /api/v1/projects/:pid/requirements` route added by US045 is KEPT (it is contract §4, canonical) — do not delete it here.
- `internal/handler/requirement_handler.go` and `internal/repo/requirement_repo.go` are created by US045 `be_requirement_repo_and_list_api`. This task may need to ADD a `GetRequirement` method to the repo and a chain helper. **Coordinate: prefer adding `GetRequirement` in the US045 repo task** (note it there) so this task only consumes it. If absent at pickup, add it here.
- `user_story_repo.go`/`document_repo.go` SELECT changes here do NOT overlap the INSERT changes in US045 `be_requirement_mcp_tools` (different methods), but both touch the same files — sequence or accept merge.

## Architecture extract

### Decision D-009 — Full canonical entity hierarchy; flat routes removed
Register every nested-resource read endpoint **only** under its full hierarchy path and **delete** all flat/shorthand route registrations from `main.go`. Each nested handler validates the complete ownership chain (`requirement.project_id == :pid`, `userStory.requirement_id == :rid`, `userStory.project_id == :pid`, `task.user_story_id == :usid`, `document.requirement_id == :rid`); any link mismatch → **404** with the existing not-found envelope, indistinguishable from a true not-found (no cross-resource leakage). Response bodies unchanged from the old flat endpoints plus the already-planned `requirementId`; **only the URL changes**. Guard order is fixed (validate parents before returning the child) so every failure collapses to a single 404.

### Routes REMOVED (delete from `cmd/api-server/main.go`)
```
GET /api/v1/projects/:id/user-stories     → replaced by §6
GET /api/v1/projects/:id/documents        → replaced by §10
GET /api/v1/user-stories/:id              → replaced by §7
GET /api/v1/user-stories/:id/tasks        → replaced by §8
GET /api/v1/tasks/:id                     → replaced by §9
GET /api/v1/documents/:id                 → replaced by §11
GET /api/v1/requirements/:rid/user-stories (intermediate draft — if registered) → §6
GET /api/v1/requirements/:rid/documents    (intermediate draft — if registered) → §10
```
Note: the live `main.go` currently registers `projects/:id/documents`, `documents/:id`, `projects/:id/user-stories`, `user-stories/:id`, `user-stories/:id/tasks` — remove all of these. The intermediate `requirements/:rid/...` draft routes are not present in the current `main.go`; remove only if found. `tasks/:id` is not currently registered (no top-level task route in live main.go) — remove only if present.

### Contract §6 — GET /api/v1/projects/:pid/requirements/:rid/user-stories
- **Path params:** `pid`, `rid`. **Ownership chain:** requirement exists and `requirement.project_id == :pid`; any mismatch → **404** (`NOT_FOUND`), indistinguishable.
- **200 OK** — item shape includes `requirementId`, order `createdAt` DESC:
```json
{
  "userStories": [
    {
      "id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
      "projectId": "11111111-1111-1111-1111-111111111111",
      "requirementId": "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f",
      "title": "Add item to basket",
      "description": "",
      "status": "in_progress",
      "taskCount": 3,
      "createdAt": "2026-06-02T09:00:00Z",
      "updatedAt": "2026-06-02T09:00:00Z"
    }
  ]
}
```
Empty → `{ "userStories": [] }`.
- **404**: `{ "code": "NOT_FOUND", "message": "Requirement not found" }`
- **500**: `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch user stories" }`

### Contract §7 — GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid (detail)
- **Ownership chain:** fetch story by `usid`; verify `userStory.requirement_id == :rid` AND `userStory.project_id == :pid`. Any mismatch / not-found → **404**.
- **200 OK** — bare story detail object (no `taskCount`):
```json
{
  "id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  "projectId": "11111111-1111-1111-1111-111111111111",
  "requirementId": "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f",
  "title": "Add item to basket",
  "description": "",
  "status": "in_progress",
  "createdAt": "2026-06-02T09:00:00Z",
  "updatedAt": "2026-06-02T09:00:00Z"
}
```
- **404**: `{ "code": "NOT_FOUND", "message": "User story not found" }`
- **500**: `{ "code": "INTERNAL_ERROR", "message": "Internal server error" }`

### Contract §8 — GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks
- **Ownership chain:** `requirement.project_id == :pid`, `userStory.requirement_id == :rid`, `userStory.project_id == :pid` before returning tasks. Any mismatch / not-found → **404**.
- **200 OK** — existing task item shape (unchanged; no `requirementId` on tasks):
```json
{
  "tasks": [
    {
      "id": "dddddddd-dddd-dddd-dddd-dddddddddddd",
      "userStoryId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
      "title": "be_basket_repo",
      "description": "",
      "status": "pending",
      "track": "BE",
      "createdAt": "2026-06-02T09:00:00Z",
      "updatedAt": "2026-06-02T09:00:00Z"
    }
  ]
}
```
Empty → `{ "tasks": [] }`.
- **404**: `{ "code": "NOT_FOUND", "message": "User story not found" }`
- **500**: `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch tasks" }`
> Note: the architect's example includes a `track` field. The live `taskResponse` shape (`internal/handler/user_story_detail_handler.go`) currently has `id,userStoryId,title,description,status,createdAt,updatedAt` (no `track`). **Keep the task body byte-for-byte identical to the existing flat endpoint** — D-009 says "task shape unchanged". Do NOT add `track` if the existing shape lacks it; the architect's example is illustrative of the existing shape, not a new field. If the tester's spec asserts `track`, flag `ARCHITECTURE_TEST_CONFLICT`.

### Contract §9 — GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid (detail)
- **Ownership chain:** `requirement.project_id == :pid`, `userStory.requirement_id == :rid`, `userStory.project_id == :pid`, AND `task.user_story_id == :usid`. Any mismatch / not-found → **404**.
- **200 OK** — bare task detail object (existing shape, unchanged):
```json
{
  "id": "dddddddd-dddd-dddd-dddd-dddddddddddd",
  "userStoryId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  "title": "be_basket_repo",
  "description": "",
  "status": "pending",
  "track": "BE",
  "createdAt": "2026-06-02T09:00:00Z",
  "updatedAt": "2026-06-02T09:00:00Z"
}
```
- **404**: `{ "code": "NOT_FOUND", "message": "Task not found" }`
- **500**: `{ "code": "INTERNAL_ERROR", "message": "Internal server error" }`
> Same `track` caveat as §8 — keep the existing flat task-detail body byte-for-byte. Note: there is currently no top-level `tasks/:id` route in live `main.go`; this nested detail route reuses the task fetch + the existing `taskResponse` mapping.

### Contract §10 — GET /api/v1/projects/:pid/requirements/:rid/documents
- **Ownership chain:** `requirement.project_id == :pid`; any mismatch / not-found → **404**.
- **200 OK** — metadata-only item shape (no `content`), includes `requirementId`, order `updatedAt` DESC, `id` DESC (existing convention):
```json
{
  "documents": [
    {
      "id": "cccccccc-cccc-cccc-cccc-cccccccccccc",
      "projectId": "11111111-1111-1111-1111-111111111111",
      "requirementId": "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f",
      "title": "README",
      "createdAt": "2026-06-02T09:00:00Z",
      "updatedAt": "2026-06-02T09:00:00Z"
    }
  ]
}
```
Empty → `{ "documents": [] }`.
- **404**: `{ "code": "NOT_FOUND", "message": "Requirement not found" }`
- **500**: `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch documents" }`

### Contract §11 — GET /api/v1/projects/:pid/requirements/:rid/documents/:docid (detail)
- **Ownership chain:** fetch doc by `docid`; verify `document.requirement_id == :rid` AND `document.project_id == :pid`. Any mismatch / not-found → **404**.
- **200 OK** — full document object **including `content`** and `requirementId`:
```json
{
  "id": "cccccccc-cccc-cccc-cccc-cccccccccccc",
  "projectId": "11111111-1111-1111-1111-111111111111",
  "requirementId": "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f",
  "title": "README",
  "content": "# README\n...",
  "createdAt": "2026-06-02T09:00:00Z",
  "updatedAt": "2026-06-02T09:00:00Z"
}
```
- **404**: `{ "code": "NOT_FOUND", "message": "Document not found" }`
- **500**: `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch document" }`

### Conventions / error model
Error envelope `{ "code", "message" }`; codes `NOT_FOUND` / `INTERNAL_ERROR` (no new codes). Timestamps ISO-8601 UTC `2006-01-02T15:04:05Z`. List arrays never `null`.

### Existing handler/repo logic to reuse
- `UserStoryHandler.GetProjectUserStories` already does project-existence-then-list; adapt to requirement scope (verify requirement→project, then `ListByRequirement`). Add `RequirementID` to `userStoryListItem` + `userStoryDetailResponse`.
- `UserStoryHandler.GetUserStory` / `GetUserStoryTasks` (`user_story_detail_handler.go`) — wrap with chain guard.
- `DocumentHandler.ListProjectDocuments` / `GetDocument` — adapt to requirement scope + chain guard; add `RequirementID` to `documentListItem` and to `mapDocumentToResponse` output.
- Repos: add `ListByRequirement(ctx, requirementID)` to user story + document repos (WHERE `requirement_id = $1`, same ORDER BY as today); add `requirement_id` to SELECT column lists + Scan targets in `GetUserStory`, `ListUserStoriesWithTaskCount`, `GetDocument`, `ListDocuments`. `GetRequirement(ctx, id)` on the requirement repo for the chain check.

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US048_be_unit_tests.md`: UT/IT IDs covering — each §6–§11 happy 200 (correct body incl. `requirementId` where applicable, scoped correctly); the ownership-mismatch 404 permutations (project missing; requirement in wrong project; story in wrong requirement; task in wrong story; document in wrong requirement; resource id missing); each removed route no longer resolves; 500 paths. Keep mismatch permutations at handler-via-`httptest` level (tester guidance).
- Flag any spec gaps (e.g. `track` field) back to tester / `ARCHITECTURE_TEST_CONFLICT` if applicable.

## Implementation notes
- Use path param names exactly: `:pid`, `:rid`, `:usid`, `:tid`, `:docid`.
- Guard helper: load project → load requirement & assert `project_id == :pid` → load child & assert membership. Collapse every failure to the resource's not-found envelope. A `repo.ErrNotFound` anywhere in the chain → 404.
- Do not introduce a new handler struct or error code (D-009).

## Definition of done
- All listed tests green.
- `go vet ./...` and `go test ./...` clean inside `services/agent-board`.
- Coverage ≥80% on each new/modified production `.go` file in `## Files touched`, or a written `## Coverage exemption`.
- No new public exports without a doc comment.
- Code matches the `## Architecture extract` (§6–§11 JSON + 404 envelopes verbatim; 8 routes removed; 6 added).
- Review gate green (BE + cross; paste `REVIEW GATE: PASS` into `## Notes`).
- `robot --dryrun tests/e2e/REQ008_*/` parses (paste output into `## Notes`).
- Dev set status to `in_review` and reported back.

## Notes

## Review log
