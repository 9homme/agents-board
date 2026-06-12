# US010 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ003_status_state_machine/US010_state_change_audit_trail.robot`.

## Why e2e
Retrieving the audit trail involves verifying the new MCP tools (`get_task_audit_trail`, `get_user_story_audit_trail`) correctly format the JSON response and that the end-to-end flow from updating an entity to querying its history works consistently across the system.

## Scenarios
### E2E-001 — Retrieve task audit trail after valid transitions
- **Tag:** US010, regression
- **Preconditions:** Seed a valid Project, User Story, and Task via MCP tools.
- **Steps:** (HTTP via RequestsLibrary + MCP SSE)
  1. Create a task with status `pending`.
  2. Update the task status to `in_progress`.
  3. Update the task status to `in_review`.
  4. Call the `get_task_audit_trail` MCP tool for the task.
- **Expected:** The `get_task_audit_trail` response contains at least two entries (from `pending` to `in_progress`, and `in_progress` to `in_review`), returned in chronological order.

### E2E-002 — Retrieve story audit trail after valid transitions
- **Tag:** US010, regression
- **Preconditions:** Seed a valid Project and User Story via MCP tools.
- **Steps:**
  1. Create a user story with status `draft`.
  2. Update the story status to `in_development`.
  3. Call the `get_user_story_audit_trail` MCP tool for the story.
- **Expected:** The response contains at least one entry (from `draft` to `in_development`).

### E2E-003 — Audit record not created on invalid transition
- **Tag:** US010
- **Preconditions:** Seed a valid Project, User Story, and Task.
- **Steps:**
  1. Create a task with status `pending`.
  2. Attempt to update the task to `completed` (invalid transition). This should fail.
  3. Call the `get_task_audit_trail` MCP tool for the task.
- **Expected:** The audit trail is empty, meaning the failed transition was not recorded.
