# US045/be_project_create_with_path

**Requirement:** REQ008
**Story:** US045
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** US045_be_requirement_repo_and_list_api
**Worked-by:**
**Implements:** US045, D-3 (path validation), D-3b (path uniqueness), D-006 (path required), D-008 (MCP create_project requires path), API contract §1/§2 (projects gain `path`), §3 (`POST /api/v1/projects`)

## Goal
Add server-side path validation (`os.Stat`/`IsDir`) and persistence so a project is created with a required, validated, unique local `path` over both the new HTTP `POST /api/v1/projects` and the MCP `create_project` tool, and surface `path` on all project read responses.

## Scope
- **In:**
  - New `internal/fsutil` package — `os.Stat`/`IsDir` path-validation helper (isolated for unit testing; no suggestion/autocomplete logic).
  - `internal/repo/project_repo.go` — `CreateProject` accepts required `path`; INSERT gains `path`; SELECTs (`GetProject`, `ListProjects`, `UpdateProject` RETURNING) gain `path`; surface sentinels `ErrDuplicatePath` (→409) and `ErrInvalidPath` (→400). Detect Postgres unique-violation on `uq_projects_path` → `ErrDuplicatePath`.
  - `internal/handler/project_handler.go` — add `path` to `projectResponse`; new `CreateProject` handler with inline `fsutil` path validation + 400/409/201 envelopes.
  - `cmd/api-server/main.go` — register `POST /api/v1/projects`.
  - `internal/handler/project_tools.go` — MCP `create_project` gains required, validated, unique `path` (D-008), reusing `fsutil` + the repo sentinels.
- **Out:**
  - Requirement repo / list HTTP (US045 `be_requirement_repo_and_list_api`).
  - Requirement MCP tools + story/document `requirement_id` (US045 `be_requirement_mcp_tools`).
  - Hierarchy route migration (US048).

## Files touched (estimated, exclusive)
- `services/agent-board/internal/fsutil/fsutil.go` (new)
- `services/agent-board/internal/fsutil/fsutil_test.go` (new)
- `services/agent-board/internal/repo/project_repo.go` (modify)
- `services/agent-board/internal/repo/project_repo_test.go` (modify)
- `services/agent-board/internal/handler/project_handler.go` (modify)
- `services/agent-board/internal/handler/project_handler_test.go` (modify)
- `services/agent-board/internal/handler/project_tools.go` (modify)
- `services/agent-board/internal/handler/project_tools_test.go` (modify)
- `services/agent-board/cmd/api-server/main.go` (modify — add ONE POST route)

**Shared-file note:** `cmd/api-server/main.go` is also edited by `be_requirement_repo_and_list_api` (GET route) and US048 (route removal/addition). All three add/remove distinct lines; sequence or accept a small merge. `project_repo.go`/`project_handler.go`/`project_tools.go` are owned exclusively by this task within REQ008.

## Architecture extract

### Decision D-3 — Path validation
The backend **stores the path string and verifies via `os.Stat` that it exists on disk and is a directory** (`IsDir()`); otherwise the create request is rejected (400). Path is **required** (NOT NULL) — absent/blank `path` → 400.

### Decision D-3b — Path uniqueness
A project's `path` must be **unique** across projects — a duplicate path returns **409**.

### Decision D-006 — `path` required everywhere
`path` is **required and non-blank at the API** — the create endpoint rejects absent/blank `path` with 400. No path-less projects.

### Decision D-008 — `create_project` MCP tool must also require a validated `path`
The MCP `create_project` tool gains a **required, non-blank** `path` input, validated (`os.Stat` + `IsDir`, uniqueness) on the same code path as the HTTP create handler (shared via the `fsutil` helper + repo sentinels `ErrInvalidPath`/`ErrDuplicatePath`). The DB `DEFAULT ''` is retained only to satisfy the ALTER on the legacy row; never a reachable value via either create path.

### Conventions
- Base prefix `/api/v1`. JSON. Error envelope `{ "code", "message" }`. New code: `DUPLICATE_PATH` (409). Existing: `VALIDATION_ERROR` (400), `NOT_FOUND` (404), `INTERNAL_ERROR` (500). Timestamps ISO-8601 UTC `2006-01-02T15:04:05Z`.
- **Logging:** Do **not** log full filesystem paths at info level in the create handler — log counts/codes instead.

### Contract §1 — GET /api/v1/projects (MODIFIED: now includes `path`)
- **200 OK** — each project object **gains a `path` field** (`string`, always present, non-empty directory path):
```json
{
  "projects": [
    {
      "id": "11111111-1111-1111-1111-111111111111",
      "name": "agents-board",
      "description": "",
      "path": "/Users/me/workspace/agents-board",
      "createdAt": "2026-06-01T09:00:00Z",
      "updatedAt": "2026-06-01T09:00:00Z"
    }
  ]
}
```
- **500**: `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch projects" }`

### Contract §2 — GET /api/v1/projects/:pid (MODIFIED: now includes `path`)
- **200 OK** — bare project object **plus `path`**:
```json
{
  "id": "11111111-1111-1111-1111-111111111111",
  "name": "agents-board",
  "description": "",
  "path": "/Users/me/workspace/agents-board",
  "createdAt": "2026-06-01T09:00:00Z",
  "updatedAt": "2026-06-01T09:00:00Z"
}
```
- **404**: `{ "code": "NOT_FOUND", "message": "Project not found" }`
- **500**: `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch project" }`

### Contract §3 — POST /api/v1/projects (NEW HTTP handler)
- **Request body:**
```json
{
  "name": "agents-board",
  "description": "",
  "path": "/Users/me/workspace/agents-board"
}
```
Field rules: `name` string **required, non-blank** (trimmed); `description` string **optional** (default `""`); `path` string **required, non-blank** — must exist on disk (`os.Stat`) and be a directory (`IsDir()`), and must be unique across projects.
- **201 Created** — bare project object including `path`:
```json
{
  "id": "33333333-3333-3333-3333-333333333333",
  "name": "agents-board",
  "description": "",
  "path": "/Users/me/workspace/agents-board",
  "createdAt": "2026-06-09T11:00:00Z",
  "updatedAt": "2026-06-09T11:00:00Z"
}
```
- **400 Bad Request** — blank/absent `name` or `path`, OR `path` does not exist / is not a directory (persists nothing):
```json
{ "code": "VALIDATION_ERROR", "message": "path is required" }
{ "code": "VALIDATION_ERROR", "message": "path does not exist or is not a directory" }
{ "code": "VALIDATION_ERROR", "message": "name is required" }
```
- **409 Conflict** — `path` already linked to another project (persists nothing):
```json
{ "code": "DUPLICATE_PATH", "message": "path already linked to another project" }
```
- **500**: `{ "code": "INTERNAL_ERROR", "message": "Failed to create project" }`
- **Idempotency:** none. Uniqueness on `path` enforced at the DB level (constraint `uq_projects_path`) and surfaced as 409. The FE distinguishes 400 (`VALIDATION_ERROR`) vs 409 (`DUPLICATE_PATH`).

### MCP `create_project` (D-008)
Existing tool (`project_tools.go`) sets only `name`/`description`. Add required `path`: trim + non-blank check (else tool error), validate via `fsutil` (`os.Stat`+`IsDir`, else tool error), pass to repo; map `ErrDuplicatePath` → a duplicate-path tool error. Keep existing camelCase input keys (`name`, `description`); add `path` as a new input key.

### Data model (created by US044 — read/write here)
```sql
ALTER TABLE projects ADD COLUMN path TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD CONSTRAINT uq_projects_path UNIQUE (path);
```
`domain.Project.Path string \`json:"path"\`` already exists (US044).

### Existing code to extend
- `project_repo.go` `CreateProject` currently: `INSERT INTO projects (name, description) VALUES ($1,$2) RETURNING id, created_at, updated_at`. Change to include `path`: `INSERT INTO projects (name, description, path) VALUES ($1,$2,$3) RETURNING id, created_at, updated_at`. On error, detect pq/pgx unique-violation (SQLSTATE `23505`, constraint `uq_projects_path`) and return `ErrDuplicatePath`.
- `GetProject`/`ListProjects` SELECTs: add `path` to the column list and `Scan` target.
- `project_handler.go` `projectResponse`: add `Path string \`json:"path"\``; populate it in `GetProjects`/`GetProject` and the new `CreateProject`.

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US045_be_unit_tests.md`: UT/IT IDs covering — `fsutil` validate (exists+dir → ok; missing → err; file-not-dir → err; blank → err); §3 POST 201 happy path (response includes `path`); §3 400 blank/absent name; §3 400 blank/absent path; §3 400 path missing-on-disk / not-a-directory; §3 409 duplicate path; §3 500 repo error; §1/§2 GET responses now include `path`; MCP `create_project` requires + validates `path` (success; blank → error; invalid → error; duplicate → error).
- Flag any spec gaps back to tester.

## Implementation notes
- `fsutil` signature suggestion: `func ValidateDir(path string) error` returning a sentinel (e.g. `ErrPathNotDir`) the handler maps to the 400 message `path does not exist or is not a directory`; the handler/tool does the blank-check before calling `fsutil` so the blank case yields `path is required`.
- Wire in `main.go`: `e.POST("/api/v1/projects", projectHandler.CreateProject)`.
- Bind the request body with echo's `c.Bind(&req)`; trim `name`; treat missing JSON keys as blank.

## Definition of done
- All listed tests green.
- `go vet ./...` and `go test ./...` clean inside `services/agent-board`.
- Coverage ≥80% on each new/modified production `.go` file in `## Files touched`, or a written `## Coverage exemption`.
- No new public exports without a doc comment.
- Code matches the `## Architecture extract` (§1/§2/§3 JSON + 400/409 bodies verbatim; D-008 MCP parity).
- No full paths logged at info level.
- Review gate green (BE + cross; paste `REVIEW GATE: PASS` into `## Notes`).
- `robot --dryrun tests/e2e/REQ008_*/` parses (paste output into `## Notes`).
- Dev set status to `in_review` and reported back.

## Notes

## Review log
