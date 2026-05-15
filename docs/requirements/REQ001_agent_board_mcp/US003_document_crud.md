# US003 — Document CRUD

**Requirement:** REQ001 — agent_board_mcp
**Status:** done

## Story
As an MCP client, I want to perform CRUD operations on Documents under a Project, so that I can store and retrieve knowledge base articles or architectural decisions.

## Acceptance criteria
- **Scenario: Create a Document**
  - Given an existing Project ID and an active MCP connection
  - When I call the `create_document` tool with the Project ID, title, and content
  - Then the Document is saved in the database linked to the Project
  - And the tool returns the newly created Document ID
- **Scenario: Read a Document**
  - Given an existing Document ID
  - When I call the `get_document` tool
  - Then the tool returns the Document title, content, and parent Project ID
- **Scenario: Update a Document**
  - Given an existing Document ID
  - When I call the `update_document` tool with new content or title
  - Then the Document is updated in the database
- **Scenario: Delete a Document**
  - Given an existing Document ID
  - When I call the `delete_document` tool
  - Then the Document is removed from the database

## UI / UX flow expectations
No UI: Documents are managed entirely through MCP tools via the AI agent.

## Out of scope
- Rich text formatting validation or Markdown parsing (store as raw string).
- Version history of documents.

## Dependencies
- US001 (MCP server setup)
- US002 (Project CRUD - needed for foreign key constraints)

## Notes for the team
- Documents are direct children of Projects.

## Sign-off log

### Sign-off pass 1 — 2026-05-15 — verdict: approved
- **Spec review:** All ACs are well covered by BE unit, integration, and E2E tests. No gaps found.
- **Result review:** All tests are passing.
- **Routed to:** none
