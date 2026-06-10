# US048 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ008_requirement_entity_and_project_path/US048_fix_nested_resource_endpoints.robot`.

## Why e2e

The ownership-chain mismatch permutations and all 8 removed-route assertions are proven at the IT-048-* integration level. The e2e cases are limited to:

- **E2E-048-001 (golden path per hierarchy leaf):** One 200 round-trip for each leaf endpoint through the real running stack (user-story detail, task detail, document detail), confirming route registration, chain validation, and response shape all work end-to-end.
- **E2E-048-002 (one representative chain-mismatch 404):** A single ownership-mismatch 404 at the e2e layer to confirm the chain guard is wired in the running server (not just in httptest).
- **E2E-048-003 (removed routes return 404):** Assert all 8 removed routes no longer resolve in the real running server. This is specifically e2e because it verifies the deployed `main.go` route registration, not just an httptest harness.

The US048 story itself says: "keep ownership-mismatch permutations at the unit/integration level — a small set of e2e cases is sufficient." We comply.

## Scenarios

### E2E-048-001 — Full hierarchy leaf endpoints return 200 (golden path)
- **Tag:** US048, smoke
- **Preconditions:**
  - Stack up. `API_BASE_URL = http://localhost:8080`.
  - Suite setup creates: project P, requirement R (project_id=P), user story S (requirement_id=R, project_id=P), task T (user_story_id=S), document D (requirement_id=R, project_id=P).
- **Sub-cases (run in sequence, same fixtures):**

  **Sub-case A — user-story detail:**
  1. `GET ${API_BASE_URL}/api/v1/projects/${P}/requirements/${R}/user-stories/${S}`
  2. Assert HTTP 200.
  3. Assert body has `id=${S}`, `projectId=${P}`, `requirementId=${R}`, no `taskCount`.

  **Sub-case B — task detail:**
  1. `GET ${API_BASE_URL}/api/v1/projects/${P}/requirements/${R}/user-stories/${S}/tasks/${T}`
  2. Assert HTTP 200.
  3. Assert body has `id=${T}`, `userStoryId=${S}` — no `requirementId` on task.

  **Sub-case C — document detail:**
  1. `GET ${API_BASE_URL}/api/v1/projects/${P}/requirements/${R}/documents/${D}`
  2. Assert HTTP 200.
  3. Assert body has `id=${D}`, `projectId=${P}`, `requirementId=${R}`, `content` field present.

- **Cleanup:** MCP `delete_project ${P}` (cascades to all children).
- **Architecture cite:** §7, §9, §11

### E2E-048-002 — Chain mismatch returns 404 (no leakage)
- **Tag:** US048, regression
- **Preconditions:**
  - Suite setup creates: project P1, project P2, requirement R (project_id=P2), user story S (requirement_id=R).
- **Steps:**
  1. `GET ${API_BASE_URL}/api/v1/projects/${P1}/requirements/${R}/user-stories` (R belongs to P2, not P1).
  2. Assert HTTP 404.
  3. Assert body `{"code":"NOT_FOUND","message":"Requirement not found"}` — does NOT expose user stories of P2.
- **Cleanup:** MCP delete P1 and P2.
- **Architecture cite:** §6 "chain mismatch → 404, indistinguishable"; D-009 "no cross-resource leakage"

### E2E-048-003 — All 8 removed flat routes return non-200
- **Tag:** US048, regression
- **Preconditions:** Stack up with any project UUID (can use a dummy UUID — we are asserting route absence, not data).
- **Steps (HTTP RequestsLibrary — each assertion is a separate request):**
  1. `GET /api/v1/projects/11111111-1111-1111-1111-111111111111/user-stories` → assert NOT 200 (expect 404 or 405).
  2. `GET /api/v1/projects/11111111-1111-1111-1111-111111111111/documents` → assert NOT 200.
  3. `GET /api/v1/user-stories/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa` → assert NOT 200.
  4. `GET /api/v1/user-stories/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/tasks` → assert NOT 200.
  5. `GET /api/v1/tasks/dddddddd-dddd-dddd-dddd-dddddddddddd` → assert NOT 200.
  6. `GET /api/v1/documents/cccccccc-cccc-cccc-cccc-cccccccccccc` → assert NOT 200.
  7. `GET /api/v1/requirements/b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f/user-stories` → assert NOT 200.
  8. `GET /api/v1/requirements/b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f/documents` → assert NOT 200.
- **Expected:** All 8 requests return a status code other than 200. The old handlers do NOT execute — no data is returned.
- **Cleanup:** None.
- **Architecture cite:** Breaking changes — "REMOVED HTTP routes"; D-009 "no backward compatibility"
