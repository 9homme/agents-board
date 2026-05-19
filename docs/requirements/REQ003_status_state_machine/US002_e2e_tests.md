# US002 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ003_status_state_machine/US002_story_state_machine.robot`.

## Why e2e
While domain logic validates transitions, testing the full lifecycle end-to-end ensures the MCP `update_user_story` tool correctly marshals these errors back to the client (`isError: true` and the text message) over the live HTTP/SSE connection, rather than crashing or returning an opaque 500 error.

## Scenarios
### E2E-001 — Valid story state machine transitions
- **Tag:** US002, regression
- **Preconditions:** Seed a valid Project via MCP tools.
- **Steps:** (HTTP via RequestsLibrary + MCP SSE)
  1. Create a user story with status `draft`. (Should succeed)
  2. Update the story status to `in_development`. (Should succeed)
  3. Update the story status to `in_signoff`. (Should succeed)
  4. Update the story status to `changes_requested`. (Should succeed)
  5. Update the story status to `in_development`. (Should succeed)
  6. Update the story status to `in_signoff`, then to `done`. (Should succeed)
- **Expected:** Each update returns a successful MCP tool response with the updated story state.

### E2E-002 — Invalid story state machine transition rejected
- **Tag:** US002, regression
- **Preconditions:** Seed a valid Project.
- **Steps:** 
  1. Create a user story with status `draft`.
  2. Try to update the story status to `done` directly.
- **Expected:** The `update_user_story` tool returns `isError: true` with a descriptive error message indicating an invalid transition.

### E2E-003 — Enforce initial state on story creation
- **Tag:** US002
- **Preconditions:** Seed a valid Project.
- **Steps:**
  1. Try to create a user story with status `done`.
- **Expected:** Either the story is successfully created but forced to `draft`, or the creation is rejected with an error. (Check the created story's state or the response error).
