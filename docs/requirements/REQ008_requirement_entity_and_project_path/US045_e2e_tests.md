# US045 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ008_requirement_entity_and_project_path/US045_requirement_api_and_project_path.robot`.

## Why e2e

The unit and integration tests (IT-045-*) cover all handler/repo/fsutil code paths. The following e2e cases are justified because they require the complete running stack (real DB + real filesystem + running api-server process) and exercise the full I/O path that cannot be faked in unit tests:

- **E2E-045-001:** POST /api/v1/projects with a real disk directory — verifies fsutil `os.Stat` runs against the real filesystem inside the container and the 201 response is correct.
- **E2E-045-002/003:** 400 on missing/bad path — proves the validation actually fires end-to-end (not just in handler logic under httptest).
- **E2E-045-004:** 409 duplicate path — proves the DB constraint and error mapping are wired through.
- **E2E-045-005:** GET /api/v1/projects/:pid/requirements lists requirements — proves the full handler → repo → DB → JSON round-trip including the MCP create that seeded the data.
- **E2E-045-006:** MCP create_requirement then list via HTTP — proves the shared repository code path used by both MCP and HTTP serves consistent data.
- **E2E-045-007:** MCP create_user_story with requirement_id — proves the BREAKING CHANGE (requirement_id now required) is wired correctly in the running server.
- **E2E-045-008:** MCP create_user_story without requirement_id — proves the tool error path is wired.

## Scenarios

### E2E-045-001 — POST /api/v1/projects 201 with real directory path
- **Tag:** US045, smoke
- **Preconditions:**
  - Stack up (`make e2e-up`). `API_BASE_URL = http://localhost:8080`.
  - A real directory exists on the api-server filesystem. Use `/tmp` (always present) as the test path — or create a unique subdirectory via a setup keyword.
- **Steps (HTTP via RequestsLibrary):**
  1. `POST ${API_BASE_URL}/api/v1/projects` with body `{"name": "E2E Project ${random}", "description": "", "path": "/tmp/e2e-proj-${random}"}` where `${random}` is a unique suffix. Create the directory first via MCP or a pre-seed script if needed. (Alternatively use `/tmp` itself as a guaranteed directory.)
  2. Assert HTTP 201.
  3. Assert body has `id` (non-empty string), `name`, `path == "/tmp"` (or the created dir), `createdAt`, `updatedAt`.
- **Cleanup:** MCP `delete_project` with the returned `id`.
- **Architecture cite:** §3 201

### E2E-045-002 — POST /api/v1/projects 400 — missing path field
- **Tag:** US045, regression
- **Steps:**
  1. `POST ${API_BASE_URL}/api/v1/projects` with body `{"name": "No Path Project"}`.
  2. Assert HTTP 400.
  3. Assert body `{"code":"VALIDATION_ERROR","message":"path is required"}`.
- **Cleanup:** None (nothing persisted).
- **Architecture cite:** §3 400

### E2E-045-003 — POST /api/v1/projects 400 — path not on disk
- **Tag:** US045, regression
- **Steps:**
  1. `POST ${API_BASE_URL}/api/v1/projects` with body `{"name": "Bad Path", "path": "/tmp/definitely-does-not-exist-e2e-x99z"}`.
  2. Assert HTTP 400.
  3. Assert body `{"code":"VALIDATION_ERROR","message":"path does not exist or is not a directory"}`.
- **Architecture cite:** §3 400

### E2E-045-004 — POST /api/v1/projects 409 — duplicate path
- **Tag:** US045, regression
- **Preconditions:** One project already exists with `path = "/tmp"` (or create one as setup).
- **Steps:**
  1. Create first project via POST with `path = "/tmp"` (or use a seed). Record project `id` as `${PROJ_ID_1}`.
  2. `POST ${API_BASE_URL}/api/v1/projects` again with same `path = "/tmp"`.
  3. Assert HTTP 409.
  4. Assert body `{"code":"DUPLICATE_PATH","message":"path already linked to another project"}`.
- **Cleanup:** Delete `${PROJ_ID_1}` via MCP.
- **Architecture cite:** §3 409

### E2E-045-005 — GET /api/v1/projects/:pid/requirements — returns Default requirement post-migration
- **Tag:** US045, smoke
- **Preconditions:** Stack up with migration 000003 applied. At least one seeded project with a Default requirement.
- **Steps:**
  1. `GET ${API_BASE_URL}/api/v1/projects` → take first project `${PROJ_ID}`.
  2. `GET ${API_BASE_URL}/api/v1/projects/${PROJ_ID}/requirements`.
  3. Assert HTTP 200.
  4. Assert body has `"requirements"` key with array (not null).
  5. Assert first requirement has `"projectId": "${PROJ_ID}"`, `"name"`, `"status"` (one of draft/in_progress/done), `"createdAt"`, `"updatedAt"`.
- **Architecture cite:** §4 200

### E2E-045-006 — GET /api/v1/projects/:pid/requirements — 404 for unknown project
- **Tag:** US045, regression
- **Steps:**
  1. `GET ${API_BASE_URL}/api/v1/projects/00000000-0000-0000-0000-000000000000/requirements`.
  2. Assert HTTP 404.
  3. Assert body `{"code":"NOT_FOUND","message":"Project not found"}`.
- **Architecture cite:** §4 404

### E2E-045-007 — MCP create_user_story now requires requirement_id (BREAKING CHANGE)
- **Tag:** US045, smoke
- **Preconditions:** Stack up with migration 000003 applied.
- **Steps (MCP via SSE):**
  1. `create_project` MCP with name + valid path → `${PROJ_ID}`.
  2. `create_requirement` MCP with `project_id = ${PROJ_ID}`, `name = "E2E REQ"` → `${REQ_ID}`.
  3. `create_user_story` MCP with `project_id = ${PROJ_ID}`, `requirement_id = ${REQ_ID}`, `title = "E2E Story"`.
  4. Assert tool returns success with `requirementId = ${REQ_ID}`.
  5. Verify via HTTP: `GET ${API_BASE_URL}/api/v1/projects/${PROJ_ID}/requirements/${REQ_ID}/user-stories` → assert the story appears.
- **Cleanup:** MCP `delete_project` (cascades).
- **Architecture cite:** §12 BREAKING CHANGE; §6 list user stories

### E2E-045-008 — MCP create_user_story without requirement_id returns tool error
- **Tag:** US045, regression
- **Preconditions:** A project exists.
- **Steps (MCP via SSE):**
  1. `create_project` MCP → `${PROJ_ID}`.
  2. Call `create_user_story` with `project_id = ${PROJ_ID}`, `title = "No Req Story"` (no `requirement_id`).
  3. Assert tool returns an error result (not a success JSON with an id).
  4. Verify no user story was created: `GET /api/v1/projects/${PROJ_ID}/requirements` would return zero requirements, so no stories can exist.
- **Cleanup:** MCP `delete_project`.
- **Architecture cite:** §12 "missing requirement_id → tool error"
