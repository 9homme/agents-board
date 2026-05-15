# US005 — Task CRUD

**Requirement:** REQ001 — agent_board_mcp
**Status:** done

## Story
As an MCP client, I want to perform CRUD operations on Tasks under a User Story, so that I can break down requirements into actionable work items.

## Acceptance criteria
- **Scenario: Create a Task**
  - Given an existing User Story ID
  - When I call the `create_task` tool with User Story ID, title, description, and status
  - Then the Task is saved in the database linked to the User Story
  - And the tool returns the Task ID
- **Scenario: Read a Task**
  - Given an existing Task ID
  - When I call the `get_task` tool
  - Then the tool returns the Task details and its parent User Story ID
- **Scenario: Update a Task**
  - Given an existing Task ID
  - When I call the `update_task` tool to change its status or details
  - Then the Task is successfully updated
- **Scenario: Delete a Task**
  - Given an existing Task ID
  - When I call the `delete_task` tool
  - Then the Task is removed from the database

## UI / UX flow expectations
No UI: Operations exposed via MCP.

## Out of scope
- Assignment to specific users (keep it simple for now).

## Dependencies
- US001 (MCP server setup)
- US004 (User Story CRUD)

## Notes for the team
- Tasks are children of User Stories. They are not directly linked to Projects (they inherit the Project context via their parent User Story).

## Sign-off log

### Sign-off pass 1 — 2026-05-15 — verdict: approved
- **Spec review:** All ACs are well covered by BE unit, integration, and E2E tests. No gaps found.
- **Result review:** All tests are passing.
- **Routed to:** none
