# US002 — Project CRUD

**Requirement:** REQ001 — agent_board_mcp
**Status:** done

## Story
As an MCP client, I want to perform CRUD operations on Projects, so that I can manage the top-level containers for all project-related work.

## Acceptance criteria
- **Scenario: Create a Project**
  - Given an active MCP connection
  - When I call the `create_project` tool with a name and description
  - Then the project is saved in the PostgreSQL database
  - And the tool returns the newly created Project ID
- **Scenario: Read a Project**
  - Given an existing Project ID
  - When I call the `get_project` tool with the ID
  - Then the tool returns the Project details (name, description, timestamps)
- **Scenario: Update a Project**
  - Given an existing Project ID
  - When I call the `update_project` tool with new details
  - Then the Project is updated in the database
  - And the tool returns the updated details
- **Scenario: Delete a Project**
  - Given an existing Project ID
  - When I call the `delete_project` tool with the ID
  - Then the Project and all its cascading children (Documents, User Stories) are deleted from the database
  - And the tool returns a success status

## UI / UX flow expectations
No UI: This feature exposes CRUD operations as MCP tools for AI agents.

## Out of scope
- Pagination for list endpoints (unless required for basic functionality).

## Dependencies
- US001 (MCP server setup must be complete)
- PostgreSQL database setup

## Notes for the team
- Ensure data integrity and cascading deletes are handled properly in the DB schema.
- Projects are the root entities.

## Sign-off log

### Sign-off pass 1 — 2026-05-15 — verdict: approved
- **Spec review:** All ACs are well covered by BE unit, integration, and E2E tests. No gaps found.
- **Result review:** All tests are passing.
- **Routed to:** none
