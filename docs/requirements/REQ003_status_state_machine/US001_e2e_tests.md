# US001 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ003_status_state_machine/US001_task_state_machine.robot`.

## Why e2e
While domain logic validates transitions, testing the full lifecycle end-to-end ensures the MCP `update_task` tool correctly marshals these errors back to the client (`isError: true` and the text message) over the live HTTP/SSE connection, rather than crashing or returning an opaque 500 error.

## Scenarios
### E2E-001 — Valid task state machine transitions
- **Tag:** US001, regression
- **Preconditions:** Seed a valid Project and User Story via MCP tools.
- **Steps:** (HTTP via RequestsLibrary + MCP SSE)
  1. Create a task with status `pending`. (Should succeed)
  2. Update the task status to `in_progress`. (Should succeed)
  3. Update the task status to `in_review`. (Should succeed)
  4. Update the task status to `changes_requested`. (Should succeed)
  5. Update the task status to `in_progress`. (Should succeed)
  6. Update the task status to `completed`. (Should succeed)
- **Expected:** Each update returns a successful MCP tool response with the updated task state.

### E2E-002 — Invalid task state machine transition rejected
- **Tag:** US001, regression
- **Preconditions:** Seed a valid Project and User Story.
- **Steps:** 
  1. Create a task with status `pending`.
  2. Try to update the task status to `completed` directly.
- **Expected:** The `update_task` tool returns `isError: true` with a descriptive error message indicating an invalid transition.

### E2E-003 — Enforce initial state on task creation
- **Tag:** US001
- **Preconditions:** Seed a valid Project and User Story.
- **Steps:**
  1. Try to create a task with status `completed`.
- **Expected:** Either the task is successfully created but forced to `pending`, or the creation is rejected with an error. (Check the created task's state or the response error).
