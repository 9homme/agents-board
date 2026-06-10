# US047 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ008_requirement_entity_and_project_path/US047_requirement_navigation.robot`.

## Why e2e

The component tests (FCT-047-*) cover all state branches (loading/error/empty/success), hook cleanup, URL param management, and API client shapes in isolation. The e2e cases are justified because they require the full FE↔BE↔DB round-trip with real seeded data:

- **E2E-047-001 (golden path):** Opening a real project detail page, seeing the linked path in the header, and the requirements list all sourced from the running server — confirms the full data flow from migration → DB → handler → FE.
- **E2E-047-002:** Clicking a requirement and verifying the user stories tab updates to show that requirement's stories — a multi-step UI interaction requiring real seeded data.
- **E2E-047-003 (realistic failure):** Requirements request fails (or returns empty) → empty/error state renders gracefully — proves error boundaries don't break the whole page.

URL-param, keyboard navigation, and component-level concerns are proven at FCT level and are not promoted.

## Scenarios

### E2E-047-001 — Project detail page shows path in header and requirements list (golden path)
- **Tag:** US047, smoke
- **Preconditions:**
  - Stack up. Seed data includes at least one project with `path` set and at least one requirement (the Default from migration backfill).
  - `${WEB_BASE_URL}` = `http://localhost:3000`; `${API_BASE_URL}` = `http://localhost:8080`.
- **Steps:**
  1. `GET ${API_BASE_URL}/api/v1/projects` → capture first project `${PROJ_ID}` and `${PROJ_PATH}`.
  2. `New Page    ${WEB_BASE_URL}/projects/${PROJ_ID}`
  3. Wait for the project header to be visible (timeout=10s).
  4. Assert `${PROJ_PATH}` text is visible somewhere in the page header area.
  5. Assert a requirements list element is visible.
  6. Assert at least one requirement item is present (the "Default" requirement).
- **Expected:** Path visible in header; requirements list not empty.
- **Cleanup:** None (read-only).
- **Architecture cite:** US047 AC "Project shows its requirements"; "Linked path visible"; §4

### E2E-047-002 — Selecting a requirement scopes the user stories tab
- **Tag:** US047, regression
- **Preconditions:**
  - A project with at least one requirement and one user story under that requirement (created via MCP in suite setup).
- **Steps:**
  1. Create project via MCP → `${PROJ_ID}`.
  2. Create requirement via MCP → `${REQ_ID}`.
  3. Create user story via MCP with `requirement_id = ${REQ_ID}` → title `"E2E US047 Story"`.
  4. `New Page    ${WEB_BASE_URL}/projects/${PROJ_ID}`
  5. Wait for requirements list; click on the requirement item.
  6. Wait for the User Stories tab to update (or navigate to it).
  7. Assert `"E2E US047 Story"` appears in the user stories list.
- **Expected:** Clicking the requirement causes the user stories tab to show that requirement's stories.
- **Cleanup:** MCP `delete_project` (cascades).
- **Architecture cite:** US047 AC "Drill into a requirement"; §6 user-stories scoped to requirement

### E2E-047-003 — Empty requirements state renders "No requirements yet"
- **Tag:** US047, regression
- **Preconditions:** A freshly created project with no requirements yet (note: the migration backfill creates a Default requirement for existing projects, so this must be a brand-new project created post-migration with no auto-requirement).
- **Note:** If the server always creates a Default requirement for new projects (architecture says it does NOT — creation is MCP-only and does not auto-create requirements), then a new project created via POST /api/v1/projects should have zero requirements.
- **Steps:**
  1. `POST ${API_BASE_URL}/api/v1/projects` with valid path → `${PROJ_ID}`.
  2. `New Page    ${WEB_BASE_URL}/projects/${PROJ_ID}`
  3. Wait for requirements area (timeout=10s).
  4. Assert empty state text matching "no requirements" is visible (case-insensitive).
- **Expected:** Empty state renders without error; project header and path are still visible.
- **Cleanup:** MCP `delete_project ${PROJ_ID}`.
- **Architecture cite:** US047 AC "Empty requirements"; §4 "Empty project → requirements: []"
