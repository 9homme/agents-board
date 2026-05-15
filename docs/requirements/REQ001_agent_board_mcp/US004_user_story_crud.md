# US004 — User Story CRUD

**Requirement:** REQ001 — agent_board_mcp
**Status:** done

## Story
As an MCP client, I want to perform CRUD operations on User Stories under a Project, so that I can manage agile requirements.

## Acceptance criteria
- **Scenario: Create a User Story**
  - Given an existing Project ID
  - When I call the `create_user_story` tool with Project ID, title, description, and status
  - Then the User Story is saved in the database linked to the Project
  - And the tool returns the User Story ID
- **Scenario: Read a User Story**
  - Given an existing User Story ID
  - When I call the `get_user_story` tool
  - Then the tool returns the User Story details including its parent Project ID
- **Scenario: Update a User Story**
  - Given an existing User Story ID
  - When I call the `update_user_story` tool with new details or status
  - Then the database record is updated accordingly
- **Scenario: Delete a User Story**
  - Given an existing User Story ID
  - When I call the `delete_user_story` tool
  - Then the User Story and all its child Tasks are deleted from the database

## UI / UX flow expectations
No UI: Operations are provided as MCP tools.

## Out of scope
- Story points estimation logic (store as simple field if needed, but not required to aggregate).

## Dependencies
- US001 (MCP server setup)
- US002 (Project CRUD)

## Notes for the team
- User Stories are direct children of Projects.
- Must cascade delete Tasks when a User Story is deleted.

## Sign-off log

### Sign-off pass 1 — 2026-05-15 — verdict: approved
- **Spec review:** All ACs are well covered by BE unit, integration, and E2E tests. No gaps found.
- **Result review:** All tests are passing.
- **Routed to:** none
