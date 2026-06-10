# US045/be_requirement_mcp_tools

**Requirement:** REQ008
**Story:** US045
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** US045_be_requirement_repo_and_list_api
**Worked-by:**
**Implements:** US045, D-004 (MCP-only create), D-007 (MCP-only update), API contract §5 (`create_requirement`/`list_requirements`/`update_requirement`), §12 (`create_user_story` BREAKING), §13 (`create_document` BREAKING)

## Goal
Add the requirement MCP tools (`create_requirement`, `list_requirements`, `update_requirement`) and fix the two breaking MCP create tools (`create_user_story`, `create_document`) to require `requirement_id`, so po-ba/agents can register and advance requirements and story/document creation keeps working after the NOT NULL migration.

## Scope
- **In:**
  - New `internal/handler/requirement_tools.go` — `RegisterRequirementTools(registry, requirementRepo, projectRepo)` registering `create_requirement`, `list_requirements`, `update_requirement`.
  - Add `requirement_id` to `create_user_story` (`internal/handler/user_story_tools.go`) + repo INSERT (`internal/repo/user_story_repo.go:47`).
  - Add `requirement_id` to `create_document` (`internal/handler/document_tools.go`) + repo INSERT (`internal/repo/document_repo.go:31`).
  - Register the new tools in `cmd/mcp-server/main.go`.
- **Out:**
  - HTTP requirement list (US045 `be_requirement_repo_and_list_api` — already creates the `RequirementRepository` this task uses; do NOT recreate `requirement_repo.go`).
  - Project create / path validation (US045 `be_project_create_with_path`).
  - `update_user_story`/`update_document`/`update_project`/status-transition/delete tools (unaffected — they do not touch `requirement_id`).

## Files touched (estimated, exclusive)
- `services/agent-board/internal/handler/requirement_tools.go` (new)
- `services/agent-board/internal/handler/requirement_tools_test.go` (new)
- `services/agent-board/internal/handler/user_story_tools.go` (modify — add `requirement_id` to `create_user_story`)
- `services/agent-board/internal/handler/user_story_tools_test.go` (modify)
- `services/agent-board/internal/handler/document_tools.go` (modify — add `requirement_id` to `create_document`)
- `services/agent-board/internal/handler/document_tools_test.go` (modify)
- `services/agent-board/internal/repo/user_story_repo.go` (modify — INSERT gains `requirement_id`)
- `services/agent-board/internal/repo/document_repo.go` (modify — INSERT gains `requirement_id`)
- `services/agent-board/cmd/mcp-server/main.go` (modify — register `RegisterRequirementTools`; this task is the single writer of mcp-server main.go for REQ008. US045 `be_project_create_with_path` also edits `project_tools.go` but NOT mcp-server/main.go, so no collision there.)

**Shared-file note:** `internal/repo/user_story_repo.go` and `document_repo.go` `Create*` INSERTs change here. US048 reads via these repos' SELECTs but does NOT touch the INSERTs, so no collision. `requirement_repo.go` is owned by the blocking task (`be_requirement_repo_and_list_api`); import its `RequirementRepository` — do not redefine it.

## Architecture extract

### Decision D-004 — Requirement create via MCP only
Create and list via MCP tools (`create_requirement`, `list_requirements`). po-ba calls `create_requirement` at end of Phase 1 after writing REQ docs to disk.

### Decision D-007 — Requirement update via MCP only (no HTTP PATCH)
Add an MCP `update_requirement` tool (partial update: `status`, `name`, `description`). No HTTP PATCH. **No state-machine enforcement** — `Requirement.status` is a plain stored enum; any of `draft|in_progress|done` may be set from any other.

### Contract §5 — MCP tools `create_requirement`, `list_requirements`, `update_requirement`

**`create_requirement`**
- **Input:** `project_id` (string, uuid, required), `name` (string, required non-blank), `description` (string, optional, default `""`), `status` (string, optional, one of `"draft"|"in_progress"|"done"`, default `"draft"`)
- **Output (success):** the created requirement object (same shape as the list item — see §4 below)
- **Errors:** project not found → tool error; blank name → tool error; invalid status → tool error

**`list_requirements`**
- **Input:** `project_id` (string, uuid, required)
- **Output (success):** `{ "requirements": [...] }` — same shape as `GET /api/v1/projects/:pid/requirements`
- **Errors:** project not found → tool error

**`update_requirement`**
- **Input:** `requirement_id` (string, uuid, **required**); `status` (string, optional, one of `"draft"|"in_progress"|"done"`); `name` (string, optional, non-blank if provided); `description` (string, optional). At least one mutable field SHOULD be provided; an all-empty update is a no-op returning the current object.
- **Behaviour:** partial update — only provided fields change. `status` validated against the enum; invalid value → tool error. **No state-machine enforcement.** `name`, if provided, is trimmed and must be non-blank. `updated_at` bumped on any change.
- **Output (success):** the updated requirement object (same shape as the list item):
```json
{
  "id": "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f",
  "projectId": "11111111-1111-1111-1111-111111111111",
  "name": "Default",
  "description": "",
  "status": "in_progress",
  "createdAt": "2026-06-09T10:00:00Z",
  "updatedAt": "2026-06-09T12:30:00Z"
}
```
- **Errors:** requirement not found → tool error; invalid status value → tool error; blank `name` when provided → tool error.

### Requirement object shape (shared with §4)
```json
{
  "id": "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f",
  "projectId": "11111111-1111-1111-1111-111111111111",
  "name": "Default",
  "description": "",
  "status": "draft",
  "createdAt": "2026-06-09T10:00:00Z",
  "updatedAt": "2026-06-09T10:00:00Z"
}
```

### Contract §12 — MCP `create_user_story` — BREAKING: must supply `requirement_id`
- **Why this breaks:** current INSERT `INSERT INTO user_stories (project_id, title, description, status)` (`user_story_repo.go:47`) omits `requirement_id`; after migration `000003` sets `user_stories.requirement_id NOT NULL`, **every existing call fails** with a NOT NULL violation. Story creation is MCP-only — this is the sole write path.
- **Input (modified):** `project_id` (uuid, required); `requirement_id` (uuid, **required — NEW**); `title` (required); `description` (optional); `status` (optional, default `"draft"`).
- **Validation:** `requirement_id` required and non-blank; SHOULD belong to `project_id` (requirement's `project_id` must equal the supplied `project_id`) — mismatch → tool error `requirement does not belong to project`. Repo INSERT becomes `INSERT INTO user_stories (project_id, requirement_id, title, description, status)`.
- **Output (success):** the created story object now including `requirementId`:
```json
{
  "id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  "projectId": "11111111-1111-1111-1111-111111111111",
  "requirementId": "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f",
  "title": "Add item to basket",
  "description": "",
  "status": "draft",
  "createdAt": "2026-06-09T12:00:00Z",
  "updatedAt": "2026-06-09T12:00:00Z"
}
```
- **Errors:** missing `project_id`/`requirement_id`/`title` → tool error; requirement not found or not in project → tool error; invalid initial status → tool error (existing rule retained — only `draft` initial status via `domain.NewUserStory`).

### Contract §13 — MCP `create_document` — BREAKING: must supply `requirement_id`
- **Why this breaks:** current INSERT `INSERT INTO documents (project_id, title, content)` (`document_repo.go:31`) omits `requirement_id`; after migration `000003` sets `documents.requirement_id NOT NULL`, every call fails. Document creation is MCP-only.
- **Input (modified):** `project_id` (uuid, required); `requirement_id` (uuid, **required — NEW**); `title` (required); `content` (optional).
- **Validation:** `requirement_id` required and non-blank; SHOULD belong to `project_id` (mismatch → tool error). Repo INSERT becomes `INSERT INTO documents (project_id, requirement_id, title, content)`.
- **Output (success):** the created document object now including `requirementId`:
```json
{
  "id": "cccccccc-cccc-cccc-cccc-cccccccccccc",
  "projectId": "11111111-1111-1111-1111-111111111111",
  "requirementId": "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f",
  "title": "README",
  "content": "# README\n...",
  "createdAt": "2026-06-09T12:00:00Z",
  "updatedAt": "2026-06-09T12:00:00Z"
}
```
- **Errors:** missing `project_id`/`requirement_id`/`title` → tool error; requirement not found or not in project → tool error.

### Repo `RequirementRepository` (created by the blocking task)
`ListByProject(ctx, projectID) ([]*domain.Requirement, error)` and `Create(ctx, *domain.Requirement) (*domain.Requirement, error)`. For `update_requirement`, add an `Update`/`UpdateRequirement` method to the same repo if not present (e.g. `UPDATE requirements SET name=$1, description=$2, status=$3, updated_at=NOW() WHERE id=$4 RETURNING ...`) and a `GetRequirement(ctx, id)` for the not-found / project-membership checks — coordinate with the blocking task so methods aren't duplicated; if missing, add them to `requirement_repo.go` here (single file, no other task writes it after the blocking task).

### MCP tool registration patterns to mirror
- Existing tools register via `registry.RegisterTool("name", func(ctx, args json.RawMessage)(interface{}, error){...})` — see `user_story_tools.go`/`task_tools.go`. Note existing tools read camelCase keys (`projectId`) from the args, but REQ008 contract specifies snake_case input keys (`project_id`, `requirement_id`). **Follow the architecture's snake_case input keys for the NEW requirement tools** and for the new `requirement_id` field on the create tools; keep existing camelCase keys (`projectId`, `title`, etc.) on `create_user_story`/`create_document` unchanged so you don't break current callers — add `requirement_id` as a new key.
- Register in `cmd/mcp-server/main.go` alongside the others: `handler.RegisterRequirementTools(toolRegistry, requirementRepo, projectRepo)` (construct `requirementRepo := repo.NewRequirementRepo(db)`).

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US045_be_unit_tests.md`: UT/IT IDs covering — `create_requirement` happy path (status defaults to `draft`), explicit status, validation (blank name / bad status / unknown project); `list_requirements` happy + unknown-project error; `update_requirement` partial update (status/name/description), bad-status error, not-found error, no-op all-empty; `create_user_story` requires `requirement_id` (success populates `requirementId`; missing/blank → error; requirement-not-in-project → error); `create_document` same for `requirement_id`.
- Flag any spec gaps back to tester.

## Implementation notes
- Reuse `domain.NewUserStory` for the initial-status check in `create_user_story` (existing behaviour). Set `RequirementID` on the domain struct before the repo INSERT.
- For requirement→project membership validation, fetch the requirement (or project's requirements) and compare `projectId`.
- Bump `updated_at` via `NOW()` in the UPDATE for `update_requirement`.

## Definition of done
- All listed tests green.
- `go vet ./...` and `go test ./...` clean inside `services/agent-board`.
- Coverage ≥80% on each new/modified production `.go` file in `## Files touched`, or a written `## Coverage exemption`.
- No new public exports without a doc comment.
- Code matches the `## Architecture extract` (§5/§12/§13 I/O verbatim; snake_case for new keys).
- Review gate green (BE + cross; paste `REVIEW GATE: PASS` into `## Notes`).
- `robot --dryrun tests/e2e/REQ008_*/` parses (paste output into `## Notes`).
- Dev set status to `in_review` and reported back.

## Notes

## Review log
