# US001 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ007_user_stories_tab_and_e2e_gate/US001_migrations_at_startup.robot`.

## Why e2e
Migrations run at server boot and block traffic until complete. E2E verifies that the running stack exposes a healthy, migrated API endpoint immediately (without requiring a separate seed or migration step).

## Scenarios
### E2E-US001-001 — API starts and serves requests immediately on a fresh migrated database
- **Tag:** US001, smoke
- **Preconditions:** The stack is up (api-server running, database accessible).
- **Steps:** 
  1. Send `GET (API_BASE_URL)/api/v1/projects` via RequestsLibrary.
- **Expected:** Returns status `200` without database undefined-table errors, proving the application tables (like `projects`) exist due to startup migrations.
- **Cleanup:** None.
