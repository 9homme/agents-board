# US046 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ008_requirement_entity_and_project_path/US046_add_project_by_local_path.robot`.

## Why e2e

The component tests (FCT-046-*) cover all UI states, client-side validation, and API client behavior in isolation using MSW. The e2e cases below are justified because they require the full FE↔BE round-trip with a real server that runs real `os.Stat` validation against the host filesystem:

- **E2E-046-001 (golden path):** Full user journey from clicking "Add Project" through form submission to seeing the new project in the list — requires the real POST /api/v1/projects endpoint and a real directory.
- **E2E-046-002 (realistic failure):** Server rejects a non-existent path (400) and the FE renders the error inline — this cross-layer behavior cannot be proven by component tests which use MSW.
- **E2E-046-003:** Duplicate path 409 → inline error — the FE correctly distinguishes DUPLICATE_PATH from VALIDATION_ERROR and shows the right message.

The loading state, button-disabled, auto-fill, double-submit, and abort cases are all provable at FCT level and are NOT promoted to e2e.

## Scenarios

### E2E-046-001 — Add Project golden path: form opens, submits, project appears in list
- **Tag:** US046, smoke
- **Preconditions:**
  - Stack up (`make e2e-up`, `make e2e-seed`). Web at `${WEB_BASE_URL}` (`http://localhost:3000`). API at `${API_BASE_URL}` (`http://localhost:8080`).
  - `/tmp` directory exists on the api-server host (always true on Linux/macOS containers).
  - No existing project with `path = "/tmp"`.
- **Steps (Browser library):**
  1. `New Page    ${WEB_BASE_URL}`
  2. Click the "Add Project" button (`role=button >> text=/add project/i`).
  3. Wait for the dialog to appear (`role=dialog` visible, timeout=5s).
  4. Fill the path field with `/tmp` (`Fill Text    [placeholder*=path] | [aria-label*=path]    /tmp`).
  5. Verify the name field auto-fills with a non-empty value (the basename of `/tmp` is `tmp`).
  6. Click the submit button.
  7. Wait for the dialog to close (`role=dialog` hidden, timeout=10s).
  8. Verify a project card/row with name containing `"tmp"` (or the manually-set name) appears in the dashboard list.
- **Expected:** Dialog closes; new project visible in list.
- **Cleanup:** Call `DELETE` via MCP or HTTP to remove the test project.
- **Architecture cite:** §3 201; US046 "Successful create"

### E2E-046-002 — Add Project: server rejects non-existent path → inline error shown
- **Tag:** US046, regression
- **Preconditions:** Stack up. Dashboard page loaded.
- **Steps (Browser library):**
  1. `New Page    ${WEB_BASE_URL}`
  2. Click "Add Project".
  3. Wait for dialog.
  4. Fill path with `/tmp/this-path-does-not-exist-e2e-x99z`.
  5. Name auto-fills (e.g. `"x99z"`).
  6. Click submit.
  7. Wait for inline error text to appear (timeout=5s).
- **Expected:**
  - An inline error message matching "path does not exist" or "not a directory" is visible within the dialog.
  - The dialog is still open (form not closed).
  - No new project appears in the dashboard list.
- **Cleanup:** None (nothing persisted).
- **Architecture cite:** §3 400 VALIDATION_ERROR; US046 "Server validation error surfaced"

### E2E-046-003 — Add Project: duplicate path returns 409 → correct inline error
- **Tag:** US046, regression
- **Preconditions:**
  - A project with `path = "/tmp"` already exists (created in test setup via POST or seed).
- **Steps (Browser library):**
  1. `New Page    ${WEB_BASE_URL}`
  2. Click "Add Project".
  3. Fill path with `/tmp`.
  4. Click submit.
  5. Wait for inline error.
- **Expected:**
  - Inline error text contains "already linked to another project" (the 409 DUPLICATE_PATH message).
  - Dialog still open.
  - No second project created.
- **Cleanup:** Delete the setup project.
- **Architecture cite:** §3 409 DUPLICATE_PATH; US046 "Server validation error surfaced"
