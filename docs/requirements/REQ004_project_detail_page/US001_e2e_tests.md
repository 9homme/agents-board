# US001 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ004_project_detail_page/US001_navigate.robot`.

## Why e2e

The following scenario cannot be verified at the BE or FE component level alone:

1. The full browser navigation from the dashboard card click through to the detail page requires the Next.js router to actually change the URL (`/projects/{id}`) and for the new page component to mount, which JSDOM-based component tests cannot verify end-to-end.
2. The tab-switching → URL persistence → browser-refresh cycle requires a real browser to verify that `?tab=user-stories` survives an actual page reload (not just a mock router state change).

The XSS, race-cancellation, error envelopes, and rendering concerns are all verified at the unit layer. The e2e layer covers only observable user journeys through the running stack.

## Scenarios

### E2E-US001-001 — Dashboard click-through to detail page then tab switch
- **Tag:** US001, smoke, regression
- **Preconditions:**
  - Next.js frontend running at `${WEB_BASE_URL}` (default `http://localhost:3000`).
  - `api-server` running at `${API_BASE_URL}` (default `http://localhost:8080`).
  - MCP server running (for data setup via `mcp_keywords.resource`).
  - Database reachable.
- **Setup (data):**
  1. Via MCP `create_project` tool: create a project with a unique name (e.g. `"REQ004 E2E Project <random>"`) and description `"E2E test project description"`.
  2. Record the returned `projectId`.
- **Steps:**
  1. Open `${WEB_BASE_URL}/` (the dashboard).
  2. Wait for the project card with the created project name to be visible.
  3. Click the project card.
  4. Assert the current URL is `${WEB_BASE_URL}/projects/{projectId}`.
  5. Wait for the page heading (`<h1>`) to contain the project name.
  6. Assert that two tabs are visible: "Documents" and "User Stories".
  7. Assert the "Documents" tab is active (has the active visual state; no `?tab=` in the URL or `?tab=documents`).
  8. Click the "User Stories" tab.
  9. Assert the URL now contains `?tab=user-stories`.
  10. Assert the tab content area shows the verbatim text `"Coming soon — user stories will appear here in a future release."`.
  11. Click the "Documents" tab.
  12. Assert the URL no longer contains `tab=user-stories` (either `?tab=documents` or `?tab=` absent).
- **Expected:**
  - Card navigation works in a real browser.
  - URL reflects tab state after tab switches.
  - User Stories placeholder text is exactly as specified.
- **Cleanup:** None strictly required (project remains in DB; test project names are randomised to avoid collisions).

### E2E-US001-002 — Direct URL to detail page with `?tab=user-stories` survives browser refresh
- **Tag:** US001, regression
- **Preconditions:** same as E2E-US001-001; use the same `projectId` from setup.
- **Setup (data):** reuse the project created in E2E-US001-001 suite setup, or create a new one.
- **Steps:**
  1. Navigate directly to `${WEB_BASE_URL}/projects/{projectId}?tab=user-stories`.
  2. Wait for page load.
  3. Assert "User Stories" tab is active.
  4. Assert the verbatim placeholder text is visible.
  5. Reload the browser page (`Reload`).
  6. Wait for page load.
  7. Assert "User Stories" tab is still active after reload.
  8. Assert placeholder text is still visible.
- **Expected:**
  - URL-driven tab state is preserved across browser refresh (not just in-memory React state).
- **Cleanup:** none.
