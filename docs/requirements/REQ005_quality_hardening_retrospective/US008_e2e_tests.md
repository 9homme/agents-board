# US008 — E2E test specification

**Owner:** tester. Implemented in `tests/e2e/REQ005_quality_hardening_retrospective/US008_stack_smoke.robot`.

## Why e2e

US008 delivers the live e2e stack itself (`docker-compose.yml`, `Makefile`, `Dockerfile`s, seed SQL). The primary validation of US008 is therefore a Robot Framework smoke test that:

1. Assumes the stack is up (`make e2e-up && make e2e-seed` has already run — this is a prerequisite for the test run, not a step inside the test).
2. Asserts the API is reachable and healthy.
3. Asserts the web UI is reachable and renders a page.
4. Asserts an existing Robot suite can be run against the live stack.

The unit/integration-level assertions (Makefile dry-runs, SQL parse checks — IT-US008-001 through IT-US008-007) verify correctness of the infrastructure files themselves. The e2e layer verifies that the assembled stack actually works end-to-end once all the pieces are in place.

This is the one story where e2e is definitionally necessary — without a passing e2e test, US008's core claim ("the stack comes up and Robot runs against it") cannot be verified at any lower layer.

## Scenarios

### E2E-US008-001 — Stack smoke: API server is healthy and returns project list

- **Tag:** US008, smoke
- **Preconditions:**
  - `make e2e-up && make e2e-seed` has completed successfully (prerequisite; NOT a Robot step).
  - `API_BASE_URL` = `http://localhost:8080` (default, per architecture §6.2).
  - Postgres, api-server, and web containers are healthy (confirmed by compose healthcheck `--wait` flag in `e2e-up`).
- **Steps (RequestsLibrary):**
  1. `Create Session    api    ${API_BASE_URL}`
  2. `${resp}=    GET On Session    api    /api/v1/projects`
  3. `Status Should Be    200    ${resp}`
  4. `${body}=    Set Variable    ${resp.json()}`
  5. `Should Contain    ${body}    projects`
- **Expected:**
  - HTTP 200 response.
  - Response body contains the `projects` key (may be an empty list if seeds have not added any projects, or non-empty if baseline seed ran).
- **Cleanup:** none required.
- **Architecture cite:** architecture §6.2 — api-server healthcheck hits `GET /api/v1/projects`; §8 API surface preserved.

### E2E-US008-002 — Stack smoke: web UI is reachable and renders the dashboard

- **Tag:** US008, smoke
- **Preconditions:** same as E2E-US008-001. `WEB_BASE_URL` = `http://localhost:3000`.
- **Steps (Browser library):**
  1. `New Browser    chromium    headless=True`
  2. `New Page    ${WEB_BASE_URL}/`
  3. `Wait For Elements State    css=body    visible    timeout=10s`
  4. `Get Title` — store in `${title}`
  5. `Should Not Be Empty    ${title}`
- **Expected:**
  - Page loads with HTTP 200 (no connection refused, no 5xx from Next.js).
  - The page `<body>` is visible.
  - The page title is not empty (Next.js app is serving HTML).
- **Cleanup:** `Close Browser`.
- **Architecture cite:** architecture §6.2 — web service at port 3000; D-012 containerised Next.js build.

### E2E-US008-003 — Existing Robot suite executes against live stack without connection error

- **Tag:** US008, smoke, regression
- **Preconditions:** same stack. At least one existing REQ Robot suite is available under `tests/e2e/REQ*/`.
- **Notes:** This scenario is NOT run by Robot Framework itself — it is executed by the `make e2e-run` target and its result is captured in the test report. The tester documents in the test report:
  - `make e2e-run REQ=REQ002` (or whichever REQ has the simplest stable tests) exit code.
  - Whether Robot reported `PASS` or `FAIL` for each test.
  - If tests fail due to stale seed data or changed API shapes from prior REQs, that is a per-REQ follow-up — NOT a US008 failure. US008's AC only requires that the tests CAN run (connection-refused errors are gone).
- **Expected outcome to capture:** `robot` process exits without a "Connection refused" error; tests attempt to run and produce PASS/FAIL outcomes (not "no connection / could not reach server" errors).
- **Architecture cite:** architecture §6.6 — existing `.robot` files run via `make e2e-run`; US008 AC "Scenario: existing per-REQ tests/e2e suites still run".
