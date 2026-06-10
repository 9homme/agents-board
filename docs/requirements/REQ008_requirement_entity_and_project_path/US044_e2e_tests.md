# US044 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ008_requirement_entity_and_project_path/US044_introduce_requirement_entity.robot`.

## Why e2e

US044 is a pure data-model / migration story with no UI and no HTTP endpoints of its own. The unit and integration tests (IT-044-001 through IT-044-010) verify every schema and backfill detail. The e2e layer adds one case: confirming that after the migration the server is alive and the project-list HTTP endpoint returns projects with the new `path` field present (confirming the schema change propagated to the read path). This cannot be proven at the unit/integration layer because it requires the full running stack (migration applied + server started + HTTP endpoint responding).

## Scenarios

### E2E-044-001 — Running server returns projects with `path` field after migration
- **Tag:** US044, smoke
- **Preconditions:**
  - `make e2e-up` has been run; `agent-board` service is up with migration 000003 applied.
  - `make e2e-seed` has been run (or the baseline SQL inserts at least one project).
  - `API_BASE_URL` points to `http://localhost:8080`.
- **Steps (HTTP via RequestsLibrary):**
  1. `GET ${API_BASE_URL}/api/v1/projects`
  2. Assert HTTP 200.
  3. Parse the JSON body; take the first project object from `projects[]`.
  4. Assert the key `"path"` is present on the project object (the field exists).
  5. Assert the value of `"path"` is a string (not null, not absent).
- **Expected:** The response confirms `projects[0].path` is a string — the migration has applied and the read path returns the new column.
- **Cleanup:** None (read-only).
- **Architecture cite:** §1 GET /api/v1/projects — `path` field always present

### E2E-044-002 — Default requirement exists for a seeded project after migration
- **Tag:** US044, regression
- **Preconditions:**
  - Stack up with migration 000003 applied and seed data that includes at least one project with at least one user story (REQ000_baseline.sql seed covers this).
  - `API_BASE_URL` = `http://localhost:8080`.
- **Steps (HTTP via RequestsLibrary):**
  1. `GET ${API_BASE_URL}/api/v1/projects` → capture the first project's `id` as `${PROJECT_ID}`.
  2. `GET ${API_BASE_URL}/api/v1/projects/${PROJECT_ID}/requirements`
  3. Assert HTTP 200.
  4. Assert body has key `"requirements"` containing at least one item.
  5. Assert first requirement has `"name": "Default"` and `"status": "draft"`.
  6. Assert first requirement has `"projectId": "${PROJECT_ID}"`.
- **Expected:** The seeded project has a "Default" requirement created by the migration backfill.
- **Cleanup:** None.
- **Architecture cite:** US044 AC "Existing projects auto-migrated into a Default Requirement"; §4 requirements list shape
