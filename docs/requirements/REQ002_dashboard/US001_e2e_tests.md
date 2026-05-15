# US001 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ002_dashboard/US001_view_project_dashboard.robot`.

## Why e2e
Verifying the dashboard requires the full Next.js frontend to `api-server` backend to database round-trip. It ensures CORS is configured correctly, environment variables (like `NEXT_PUBLIC_API_BASE_URL`) map correctly in the production build, and database queries execute properly in a real environment.

## Scenarios
### E2E-001 — View Dashboard End-to-End
- **Tag:** US001, smoke, regression
- **Preconditions:** The Next.js frontend, `api-server` backend, and `mcp-server` backend are running. The database is reachable.
- **Steps:**
  1. Connect to the `mcp-server` over SSE and use the `create_project` tool to insert a known test project (e.g. "Dashboard E2E Test").
  2. Open the web dashboard at (WEB_BASE_URL)/.
  3. Wait for the page to load and data to be fetched.
- **Expected:** The UI displays a card with the text "Dashboard E2E Test" and its description, proving the data created via MCP is correctly exposed via the REST API and rendered by the frontend.
- **Cleanup:** None strictly required for read-only view, though in a real environment we might delete the project.