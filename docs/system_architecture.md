# System Architecture — agent-board (living contract)

**Purpose:** Single source of truth for what is currently implemented and deployed. Future `system-architect` invocations read this instead of source. Update it whenever a REQ is merged.

**Scope:** Reflects approved/locked contract state. Includes through **REQ008** (Requirement entity + project local-path linking, approved 2026-06-10).

## Topology

Single Go module `agent-board` at `services/agent-board/` exposing **two binaries** over one shared Postgres DB and one shared `internal/` package set (`domain`, `repo`, `handler`, `mcp`, `config`, `migrate`).

| Binary | Cmd | Transport | Surface |
|---|---|---|---|
| api-server | `cmd/api-server/main.go` | Echo HTTP/JSON, REST | Read-only `GET` endpoints for the Next.js frontend |
| mcp-server | `cmd/mcp-server/main.go` | Echo HTTP + MCP over SSE (JSON-RPC 2.0) | Full CRUD + status transitions + audit trail for agents |

- **Migrations** run idempotently at api-server startup only (embedded `migrations.FS` via `internal/migrate`). mcp-server does NOT run migrations; it pings DB then serves.
- **DB driver:** `pgx/v5/stdlib` via `database/sql`.
- **Startup (both):** `config.ResolveDBURL()` → `sql.Open` → bounded 5s `PingContext` (signal-cancellable, SIGINT/SIGTERM) → serve.

---

## HTTP Endpoints (api-server)

All under base prefix `/api/v1`. Read endpoints (`GET`) plus one write endpoint (`POST /api/v1/projects`, added REQ008). Path params are UUID strings.

**REQ008 — full canonical entity hierarchy.** Every nested read endpoint lives under its full ownership path `Project → Requirement → UserStory → Task` (and `Project → Requirement → Document`). There are **no flat/shorthand routes** — the old flat routes were removed from `main.go` (breaking). Each nested route validates the complete ownership chain and returns **404** on any link mismatch (indistinguishable from a true not-found — no cross-resource leakage).

| Method | Path | Handler | Request body | Response shape (200) | Status codes |
|---|---|---|---|---|---|
| GET | `/api/v1/projects` | `ProjectHandler.GetProjects` | — | `{"projects":[projectObj]}` (never null; each incl. `path`) | 200, 500 |
| GET | `/api/v1/projects/:pid` | `ProjectHandler.GetProject` | — | bare `projectObj` (incl. `path`) | 200, 404, 500 |
| POST | `/api/v1/projects` | `ProjectHandler.CreateProject` | `{name(req), description?, path(req)}` | bare `projectObj` (incl. `path`) | 201, 400, 409, 500 |
| GET | `/api/v1/projects/:pid/requirements` | `RequirementHandler` (list) | — | `{"requirements":[requirementObj]}` (never null; ORDER BY created_at ASC) | 200, 404, 500 |
| GET | `/api/v1/projects/:pid/requirements/:rid/user-stories` | `UserStoryHandler` (hierarchy list) | — | `{"userStories":[userStoryListItem]}` (never null; incl. `taskCount`, `requirementId`) | 200, 404, 500 |
| GET | `/api/v1/projects/:pid/requirements/:rid/user-stories/:usid` | `UserStoryHandler` (hierarchy detail) | — | bare `userStoryObj` (**no taskCount**; incl. `requirementId`) | 200, 404, 500 |
| GET | `/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks` | `TaskHandler` (hierarchy list) | — | `{"tasks":[taskObj]}` (never null; `[]` for story-with-no-tasks) | 200, 404, 500 |
| GET | `/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid` | `TaskHandler` (hierarchy detail) | — | bare `taskObj` | 200, 404, 500 |
| GET | `/api/v1/projects/:pid/requirements/:rid/documents` | `DocumentHandler` (hierarchy list) | — | `{"documents":[docListItem]}` (never null; metadata only, **no content**; incl. `requirementId`) | 200, 404, 500 |
| GET | `/api/v1/projects/:pid/requirements/:rid/documents/:docid` | `DocumentHandler` (hierarchy detail) | — | bare `documentObj` (**includes content**; incl. `requirementId`) | 200, 404, 500 |

Notes:
- **Ownership-chain validation (REQ008):** each nested route verifies the full chain before returning the child (`requirement.project_id == :pid`, `userStory.requirement_id == :rid`, `userStory.project_id == :pid`, `task.user_story_id == :usid`, `document.requirement_id == :rid`). Any mismatch or missing parent → 404, indistinguishable from a true not-found.
- **REMOVED in REQ008 (breaking, deleted from `main.go`):** `GET /api/v1/projects/:id/user-stories` → use `.../requirements/:rid/user-stories`; `GET /api/v1/projects/:id/documents` → `.../requirements/:rid/documents`; `GET /api/v1/user-stories/:id` → `.../user-stories/:usid`; `GET /api/v1/user-stories/:id/tasks` → `.../user-stories/:usid/tasks`; `GET /api/v1/tasks/:id` → `.../tasks/:tid`; `GET /api/v1/documents/:id` → `.../documents/:docid`. All HTTP callers (FE client, e2e) must use the hierarchy paths. MCP tools are unaffected (they call repos directly, no URLs).
- 404 body: `{"code":"NOT_FOUND","message":"<entity> not found"}`. 500 body: `{"code":"INTERNAL_ERROR","message":"..."}`. `POST /projects` validation: 400 `{"code":"VALIDATION_ERROR",...}` (blank name/path, or path missing/not a directory); 409 `{"code":"DUPLICATE_PATH",...}` (path already linked).
- `POST /projects` validates `path` server-side via `os.Stat` + `IsDir()` and enforces uniqueness (DB `uq_projects_path`). Path validation reads the **api-server host filesystem** (not the browser's).

### HTTP response object shapes
```
projectObj          = {id, name, description, path, createdAt, updatedAt}   // REQ008: path always present (non-empty dir)
requirementObj      = {id, projectId, name, description, status, createdAt, updatedAt}   // status: draft|in_progress|done
documentObj         = {id, projectId, requirementId, title, content, createdAt, updatedAt}   // detail: includes content
docListItem         = {id, projectId, requirementId, title, createdAt, updatedAt}            // list: NO content
userStoryObj        = {id, projectId, requirementId, title, description, status, createdAt, updatedAt}   // detail: NO taskCount
userStoryListItem   = {id, projectId, requirementId, title, description, status, taskCount, createdAt, updatedAt}
taskObj             = {id, userStoryId, title, description, status, track, createdAt, updatedAt}
```
All fields are JSON strings except `taskCount` (number). All timestamps are strings (see API Conventions).

---

## MCP Tools (mcp-server)

Transport: client opens `GET /sse` (gets a session), then `POST /message?sessionId=<id>` with a JSON-RPC 2.0 `tools/call` request. Tool result is returned both in the POST response body and pushed over SSE.

- **JSON-RPC request:** `{"jsonrpc":"2.0","id":<any>,"method":"tools/call","params":{"name":"<tool>","arguments":<object>}}`
- **Success result:** `{"jsonrpc":"2.0","id":<id>,"result":{"content":[{"type":"text","text":"<JSON-stringified tool output>"}]}}` — **tool output is JSON-encoded into the `text` field**, not a structured object.
- **Tool-level error:** `result.isError=true`, `content[0].text=<error message>`.
- **Protocol error** (bad jsonrpc/method, tool not found): `{"...","error":{"code":<int>,"message":"..."}}`. Codes: ParseError -32700, InvalidRequest -32600, MethodNotFound -32601, InternalError (see `mcp/types.go`).
- **Transport errors** (pre-RPC): `POST /message` returns HTTP 400 `{"error":"sessionId is required"|"invalid sessionId"|"invalid JSON-RPC payload"}`.

| Tool | Input fields | Output shape (the `text` JSON) | Notes |
|---|---|---|---|
| `create_project` | `projectId`/`name`(req, trimmed≠""), `description`, `path`(**req REQ008** — validated os.Stat+IsDir, unique) | `projectObj` (incl `path`) | **REQ008 BREAKING:** `path` now required + validated; sentinels `ErrInvalidPath`(→error), `ErrDuplicatePath`(→error) |
| `get_project` | `id`(req) | `projectObj` (incl `path`) | err "project not found" |
| `update_project` | `id`(req), `name`?, `description`? (pointers) | `projectObj` (incl `path`) | partial; empty name rejected if provided; path not editable |
| `delete_project` | `id`(req) | `{"success":true}` | cascades requirements/documents/stories |
| `list_projects` | — | `{"projects":[projectObj]}` (never null; each incl `path`) | — |
| `create_requirement` | `project_id`(req), `name`(req, non-blank), `description`?, `status`?(default `draft`, one of `draft\|in_progress\|done`) | `requirementObj` | **NEW REQ008**; po-ba calls at end of Phase 1; project-not-found/blank-name/invalid-status → error |
| `list_requirements` | `project_id`(req) | `{"requirements":[requirementObj]}` (never null; ORDER BY created_at ASC) | **NEW REQ008**; project-not-found → error |
| `update_requirement` | `requirement_id`(req), `status`?, `name`?, `description`? (pointers) | `requirementObj` | **NEW REQ008**; partial; status enum-validated, **no state-machine enforcement**; all-empty = no-op; not-found/invalid-status/blank-name → error |
| `create_document` | `projectId`(req), `requirement_id`(**req REQ008**), `title`(req), `content` | `DocumentResponse` (incl content + `requirementId`) | **REQ008 BREAKING:** `requirement_id` required; must belong to `projectId`; mismatch → error |
| `get_document` | `id`(req) | `DocumentResponse` (incl `requirementId`) | err "document not found" |
| `update_document` | `id`(req), `title`?, `content`? (pointers) | `DocumentResponse` (incl `requirementId`) | partial; does not change parent |
| `delete_document` | `id`(req) | `{"success":true}` | — |
| `list_documents` | `projectId`(req) | `{"documents":[DocumentResponse]}` (incl `requirementId`) | **includes content** (differs from REST list) |
| `create_user_story` | `projectId`(req), `requirement_id`(**req REQ008**), `title`(req), `description`, `status`(default `draft`) | `UserStoryResponse` (incl `requirementId`) | **REQ008 BREAKING:** `requirement_id` required; must belong to `projectId` (mismatch → error); initial status must be `draft` (domain.NewUserStory) |
| `get_user_story` | `id`(req) | `UserStoryResponse` (incl `requirementId`) | err "user story not found" |
| `update_user_story` | `id`(req), `title`?, `description`?, `status`? | `UserStoryResponse` (incl `requirementId`) | status change validated against transition table + transactional audit insert; does not change parent |
| `delete_user_story` | `id`(req) | `{"success":true}` | cascades tasks |
| `list_user_stories` | `projectId`(req) | `{"userStories":[UserStoryResponse]}` (never null; incl `requirementId`) | no taskCount |
| `create_task` | `userStoryId`(req), `title`(req), `description`, `status`(default `pending`) | `TaskResponse` | initial status must be `pending` (domain.NewTask) |
| `get_task` | `id`(req) | `TaskResponse` | err "task not found" |
| `update_task` | `id`(req), `title`?, `description`?, `status`? | `TaskResponse` | status change validated + transactional audit insert; **REQ008/US049:** may set `blocked_review_gate` from `in_review`/`changes_requested` |
| `delete_task` | `id`(req) | `{"success":true}` | — |
| `list_tasks` | `userStoryId`(req) | `{"tasks":[TaskResponse]}` | — |
| `get_task_audit_trail` | `taskId`(req) | `{"auditTrail":[AuditLogResponse]}` | chronological ASC |
| `get_user_story_audit_trail` | `userStoryId`(req) | `{"auditTrail":[AuditLogResponse]}` | chronological ASC |

MCP response object shapes (note: MCP uses RFC3339 timestamps, REST uses a fixed-Z format — see Conventions):
```
projectObj          = domain.Project marshalled directly: {id, name, description, path, createdAt, updatedAt} (RFC3339)   // REQ008: path
requirementObj      = {id, projectId, name, description, status, createdAt, updatedAt}   // REQ008; status: draft|in_progress|done
DocumentResponse    = {id, projectId, requirementId, title, content, createdAt, updatedAt}
UserStoryResponse   = {id, projectId, requirementId, title, description, status, createdAt, updatedAt}
TaskResponse        = {id, userStoryId, title, description, status, createdAt, updatedAt}
AuditLogResponse    = {id, entityId, entityType, fromStatus, toStatus, changedAt}
```

---

## Database Schema

Postgres. Migrations: `000001_init_schema`, `000002_status_audit_trail`, `000003_requirement_entity` (REQ008). All PKs are `UUID DEFAULT gen_random_uuid()`. All `created_at`/`updated_at` are `TIMESTAMPTZ NOT NULL DEFAULT NOW()`.

### `projects`
| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, default gen_random_uuid() |
| name | VARCHAR(255) | NOT NULL |
| description | TEXT | nullable |
| path | TEXT | **NOT NULL** (REQ008); UNIQUE via `uq_projects_path`. DB DEFAULT `''` exists only to satisfy NOT NULL on the ALTER for the legacy row — never reachable via API (create paths require non-blank validated path) |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

Constraint: `uq_projects_path UNIQUE (path)` (REQ008).

### `requirements` (REQ008)
| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK, default gen_random_uuid() |
| project_id | UUID | NOT NULL, FK → projects(id) ON DELETE CASCADE |
| name | VARCHAR(255) | NOT NULL |
| description | TEXT | NOT NULL DEFAULT '' |
| status | VARCHAR(50) | NOT NULL DEFAULT 'draft', CHECK (status IN ('draft','in_progress','done')) |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

Index: `idx_requirements_project_id (project_id)`. New level between `projects` and `user_stories`/`documents`. No state-machine enforcement on `status` (plain stored enum). Backfill (migration 000003): one `'Default'` requirement per existing project, then existing `user_stories`/`documents` re-parented under it (zero data loss).

### `documents`
| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK |
| project_id | UUID | NOT NULL, FK → projects(id) ON DELETE CASCADE (retained, denormalised) |
| requirement_id | UUID | **NOT NULL** (REQ008, after backfill), FK → requirements(id) ON DELETE CASCADE |
| title | VARCHAR(255) | NOT NULL |
| content | TEXT | nullable |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

Indexes: `idx_documents_project_id (project_id)`, `idx_documents_requirement_id (requirement_id)` (REQ008).

### `user_stories`
| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK |
| project_id | UUID | NOT NULL, FK → projects(id) ON DELETE CASCADE (retained, denormalised) |
| requirement_id | UUID | **NOT NULL** (REQ008, after backfill), FK → requirements(id) ON DELETE CASCADE |
| title | VARCHAR(255) | NOT NULL |
| description | TEXT | nullable |
| status | VARCHAR(50) | NOT NULL |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

Indexes: `idx_user_stories_project_id (project_id)`, `idx_user_stories_requirement_id (requirement_id)` (REQ008).

### `tasks`
| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK |
| user_story_id | UUID | NOT NULL, FK → user_stories(id) ON DELETE CASCADE |
| title | VARCHAR(255) | NOT NULL |
| description | TEXT | nullable |
| status | VARCHAR(50) | NOT NULL |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

Index: `idx_tasks_user_story_id (user_story_id)`.

### `status_audit_trail`
| Column | Type | Constraints |
|---|---|---|
| id | UUID | PK |
| entity_id | UUID | NOT NULL (no FK — polymorphic) |
| entity_type | VARCHAR(50) | NOT NULL; `'task'` or `'user_story'` |
| from_status | VARCHAR(50) | NOT NULL |
| to_status | VARCHAR(50) | NOT NULL |
| changed_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() |

Index: `idx_status_audit_trail_entity (entity_type, entity_id, changed_at ASC)`.

Audit rows are written **only** by `UpdateTaskStatus` / `UpdateUserStoryStatus`, inside the same transaction as the status UPDATE (atomic, D-003).

---

## Domain Types (`internal/domain`)

```go
type Project struct {
    ID, Name, Description, Path string   // REQ008: Path (non-pointer, required)
    CreatedAt, UpdatedAt       time.Time
}
type Requirement struct {                // REQ008
    ID, ProjectID, Name, Description, Status string   // Status: draft | in_progress | done
    CreatedAt, UpdatedAt                     time.Time
}
type Document struct {
    ID, ProjectID, RequirementID, Title, Content string   // REQ008: RequirementID
    CreatedAt, UpdatedAt                         time.Time
}
type UserStory struct {
    ID, ProjectID, RequirementID, Title, Description, Status string   // REQ008: RequirementID
    CreatedAt, UpdatedAt                                     time.Time
}
type Task struct {
    ID, UserStoryID, Title, Description, Status string
    CreatedAt, UpdatedAt                        time.Time
}
type StatusAuditLog struct {
    ID, EntityID, EntityType, FromStatus, ToStatus string  // EntityType: "task" | "user_story"
    ChangedAt                                       time.Time
}
```
JSON tags are camelCase on all of the above (`projectId`, `userStoryId`, `createdAt`, etc.).

### Status state machines (`domain/status_machine.go`)
Constructors enforce initial status: `NewTask` requires `pending`; `NewUserStory` requires `draft` (else `ErrInvalidInitialStatus`). `IsValidTransition` enforces:

- **Task statuses:** `pending → in_progress → in_review → {completed | changes_requested}`; `changes_requested → {in_progress | in_review | completed | blocked_circuit_breaker}`. **REQ008/US049:** `in_review → blocked_review_gate` and `changes_requested → blocked_review_gate` (set via MCP `update_task` when the review-gate tooling fails — distinct from `blocked_circuit_breaker`, which is a code-review failure). Constant `TaskStatusBlockedReviewGate = "blocked_review_gate"`. No DB migration (status column is TEXT; enforcement is Go-domain-only). Terminal: `completed`, `blocked_circuit_breaker`, `blocked_review_gate`.
- **UserStory statuses:** `draft → in_development → in_signoff → {done | changes_requested}`; `changes_requested → {in_development | in_signoff | done | blocked_circuit_breaker}`. Terminal: `done`, `blocked_circuit_breaker`.
- Errors: `ErrInvalidStatusTransition`, `ErrInvalidInitialStatus`.

---

## Repositories (`internal/repo`)

Shared sentinel: `repo.ErrNotFound` (mapped from `sql.ErrNoRows`).

| Repo | Constructor | Interface methods |
|---|---|---|
| project | `NewProjectRepo` → `ProjectRepository` | Create (REQ008: requires `path`; sentinels `ErrDuplicatePath`→409, `ErrInvalidPath`→400), Get, Update, Delete, ListProjects (ORDER BY created_at DESC) |
| requirement | `NewRequirementRepo` → `RequirementRepository` (REQ008) | Create, List(projectID) (ORDER BY created_at ASC), Get, Update — backs both HTTP list and MCP `create/list/update_requirement` |
| document | `NewDocumentRepo` → `DocumentRepository` | Create (REQ008: INSERT incl `requirement_id`), Get, Update, Delete, ListDocuments(projectID), **ListDocumentsByRequirement** (REQ008) (ORDER BY updated_at DESC, id DESC) |
| user story | `NewUserStoryRepo` → `*UserStoryRepo` | Create (REQ008: INSERT incl `requirement_id`), Get, Update, **UpdateUserStoryStatus** (txn + audit), Delete, ListUserStories, **ListUserStoriesWithTaskCount** (LEFT JOIN tasks, COUNT, ORDER BY created_at DESC), **ListUserStoriesByRequirement** (REQ008) |
| task | `NewTaskRepo` → `TaskRepository` | Create, Get, Update, **UpdateTaskStatus** (txn + audit), Delete, ListTasks(userStoryID) (ORDER BY created_at DESC) |
| audit | `NewAuditRepo` → `AuditRepository` | GetTaskAuditTrail, GetUserStoryAuditTrail (both ORDER BY changed_at ASC) |

REQ008: path validation (`os.Stat` + `IsDir`) lives in `internal/fsutil`, shared by the HTTP `CreateProject` handler and the MCP `create_project` tool.

`ListUserStoriesWithTaskCount` returns `[]*UserStoryWithCount{ domain.UserStory; TaskCount int }` — backs the REST list endpoint's `taskCount`.

---

## API Conventions

- **Base prefix (REST):** `/api/v1`. MCP transport paths are `/sse` and `/message` (no prefix).
- **Error envelope (REST):** `{"code":"<CODE>","message":"<human text>"}`. Codes in use: `NOT_FOUND` (404), `INTERNAL_ERROR` (500), and (REQ008, on `POST /api/v1/projects`) `VALIDATION_ERROR` (400), `DUPLICATE_PATH` (409).
- **Error model (MCP):** tool failures → `result.isError=true` with message in `content[0].text`; protocol failures → JSON-RPC `error` object; transport failures → HTTP 400 `{"error":"..."}`.
- **List wrapping:** every list is wrapped in a named key and is **never null** (`{"projects":[]}`, `{"documents":[]}`, `{"userStories":[]}`, `{"tasks":[]}`, `{"auditTrail":[]}`). Bare object for single-entity GETs.
- **Field casing:** camelCase everywhere (`projectId`, `userStoryId`, `taskCount`, `createdAt`).
- **Timestamp format — INCONSISTENT, by design of current code (note for the architect):**
  - **REST handlers** format with `"2006-01-02T15:04:05Z"` — a literal `Z` suffix, no offset. `user-stories` detail/tasks handlers additionally call `.UTC()` first; `projects`/`documents` handlers do NOT call `.UTC()` (they apply the `Z` literal to whatever zone the time carries).
  - **MCP tools** format with `time.RFC3339` (real offset, e.g. `+00:00`/`Z`).
  - MCP `list_projects`/`get_project` marshal `domain.Project` directly → `time.Time` default JSON (RFC3339Nano).
  - This divergence is existing behavior; flag before relying on a single canonical timestamp format.
- **content field divergence:** REST document **list** omits `content`; REST document **detail** and **all MCP document tools** (incl `list_documents`) include `content`.
- **taskCount divergence:** present only on REST `user-stories` **list**; absent on REST detail and all MCP user-story tools.

---

## Configuration & Ops

- **DB connection (`config.ResolveDBURL`):** `DATABASE_URL` is the **sole** accepted var (REQ006/US010, D-006). `DB_URL` is rejected at startup with an explicit rename/disambiguate error; missing both → "DATABASE_URL environment variable is required".
- **`PORT`:** default `8080` (both binaries). api-server sanitizes control chars before logging (G706).
- **`FRONTEND_URL`:** api-server CORS allow-origin (default `*`). Allowed headers: Origin, Content-Type, Accept.
- **Middleware:** both use `RequestLogger` + `Recover` (Echo). mcp-server logs method/uri/status.
- **Migrations:** embedded, run by api-server at boot only (`internal/migrate` + `migrations.FS`), idempotent (D-001/D-002).

---

## Frontend (`web/`)

Next.js Pages Router, CSR-only. All backend calls go through `web/lib/api/`; types in `web/lib/api/types.ts` mirror the REST shapes above field-for-field. MSW handlers under `web/test/msw/` reflect these JSON shapes. (The frontend consumes the **REST** surface only — not MCP.)

---

## Maintenance note for the architect
Update this file **at architecture approval time** (after human approves `architecture.md`), not at merge. The approved architecture is the locked contract — update HTTP Endpoints, MCP Tools, DB Schema, Domain Types, and Conventions here immediately after approval so the next REQ's architect reads accurate state.

Last updated: 2026-06-10T03:58:00Z (REQ008 — Requirement entity + project local-path linking; incl. US049 `blocked_review_gate` task status).
