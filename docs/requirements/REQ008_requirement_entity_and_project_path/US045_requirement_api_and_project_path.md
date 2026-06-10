# US045 — Requirement read API (HTTP list + MCP create) + project path on create

**Requirement:** REQ008 — Requirement entity + project local-path linking
**Status:** draft

## Story
As a frontend client and as the po-ba agent, I want an HTTP API to list Requirements under a project and to create a Project with a validated required local path, plus MCP tools to create and list Requirements, so that the web can navigate the new REQ level (read-only for requirements) and register projects by path, while requirement records are created by po-ba via MCP at the end of Phase 1.

## Acceptance criteria
- **Scenario: List requirements for a project**
  - Given a project with one or more requirements
  - When the client calls `GET /api/v1/projects/{projectId}/requirements`
  - Then it returns 200 with the project's requirements (each: `id`, `projectId`, `name`, `description`, `status`, `createdAt`, `updatedAt`), ordered deterministically (e.g. `createdAt` ascending)
- **Scenario: List requirements for unknown project**
  - Given a projectId that does not exist
  - When the client calls `GET /api/v1/projects/{projectId}/requirements`
  - Then it returns 404 with a structured error body
- **Scenario: No HTTP create endpoint for requirements**
  - Given the requirement-create capability is MCP-only
  - When the client attempts an HTTP `POST` to create a requirement under a project
  - Then no such HTTP route exists (requirement creation is not exposed over HTTP; the web is read-only for requirements)
- **Scenario: Create a requirement via the `create_requirement` MCP tool**
  - Given an existing project
  - When the po-ba agent calls the `create_requirement` MCP tool with `project_id` + `name` (and optional `description`, `status`)
  - Then the requirement is persisted with the correct `project_id` and `status` defaulting to `draft` when omitted, and the created requirement object is returned
- **Scenario: Create a requirement via MCP with explicit status**
  - Given an existing project
  - When the agent calls `create_requirement` with a `status` of `draft`, `in_progress`, or `done`
  - Then the requirement is persisted at that status
- **Scenario: `create_requirement` validation**
  - Given a `create_requirement` call
  - When `name` is missing/blank, `status` is outside the allowed enum, or `project_id` does not exist
  - Then the tool returns an error and persists nothing
- **Scenario: List requirements via the `list_requirements` MCP tool**
  - Given a project with one or more requirements
  - When the agent calls `list_requirements` with that `project_id`
  - Then it returns the project's requirements (same shape as the HTTP list); an unknown `project_id` returns a tool error
- **Scenario: Create a project with a valid local path**
  - Given a local path that exists on disk and is a directory
  - When the client POSTs a project create with `name` + `path`
  - Then the backend confirms via `os.Stat` that the path exists and is a directory, and returns 201 with the created project including its `path`
- **Scenario: Create a project with a missing or blank path**
  - Given a project create request whose `path` is absent or blank
  - When the client POSTs the project create
  - Then it returns 400 with a structured validation error (`path is required`) and persists nothing — path-less projects are not allowed
- **Scenario: Create a project with an invalid path**
  - Given a `path` that is non-empty but does not exist on disk, or exists but is not a directory
  - When the client POSTs a project create
  - Then `os.Stat` validation fails and it returns 400 with a structured validation error indicating the path problem, and persists nothing
- **Scenario: Duplicate path rejected**
  - Given a `path` already linked to an existing project
  - When the client POSTs a project create with the same `path`
  - Then it returns 409 (conflict) and persists nothing
- **Scenario: Listings reflect re-parenting after migration**
  - Given the migration from US044 ran
  - When the client lists requirements for a migrated project
  - Then the "Default" Requirement is returned and its User Stories / Documents are reachable scoped to it

## UI / UX flow expectations
No UI: this story delivers backend HTTP endpoints + MCP tools + JSON contracts only. The web consumes the HTTP endpoints in US046 (Add Project) and US047 (navigation); the MCP tools are called by the po-ba agent, not the web. The exact JSON request/response shapes and status-code bodies are defined by the system-architect in `architecture.md`.

## Out of scope
- Web forms / navigation (US046, US047).
- HTTP create/update/delete of requirements — requirement **creation is MCP-only**; only HTTP **list** is exposed this REQ.
- Editing a project's path after creation.
- Filesystem autocomplete / directory suggestions — removed this REQ (path is typed manually; see US046).

## Dependencies
- US044 (data model + migration + domain types).

## Notes for the team
- HTTP endpoints needed (architect locked exact paths/shapes in `architecture.md`):
  - `GET /api/v1/projects/:id/requirements` — list requirements for a project (read-only).
  - `POST /api/v1/projects` accepting `name` + **required** `path`.
  - `GET /api/v1/requirements/:reqId/user-stories` and `GET /api/v1/requirements/:reqId/documents` — requirement-scoped listings needed by US047.
- MCP tools needed (architect locked exact I/O in `architecture.md`):
  - `create_requirement` (input: `project_id`, `name`, optional `description`, optional `status`) — called by po-ba at end of Phase 1 to register the REQ DB record after writing docs to disk.
  - `list_requirements` (input: `project_id`) — agents read a project's requirements. Shares the underlying repository with the HTTP list handler.
- There is **no** `POST /api/v1/projects/:id/requirements` HTTP endpoint (D-004). Requirement creation is MCP-only; the web is view-only for requirements.
- `path` is **required and non-blank** at the API (D-006). Absent/blank `path` → 400 (`path is required`). Existence/is-directory validation is server-authoritative via `os.Stat` (exists + `IsDir()`); uniqueness is enforced at the DB level (US044) and surfaced as 409.
- `Requirement.status` accepts only `draft` | `in_progress` | `done`; reject other values (HTTP would not apply — creation is MCP, which returns a tool error for invalid status).
- **Security:** path validation reads the api-server host filesystem. Treat paths as sensitive — do not log full paths at info level (log counts/codes). No filesystem-suggestion endpoint exists this REQ, so there is no broad directory-listing exposure.

## Sign-off log
(po-ba appends here on each sign-off pass)
