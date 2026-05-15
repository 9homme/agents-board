# REQ001 — Agent Board MCP

## Summary
Develop a minimal Jira/Confluence-like platform for agents. Agents communicate via the Model Context Protocol (MCP). The initial scope is an MCP server (SSE/HTTP) built with the Go Echo framework and PostgreSQL for persistence, providing CRUD operations for Projects, Documents, User Stories, and Tasks.

## Business Goal
Enable AI agents to seamlessly read and write project management entities (Projects, User Stories, Tasks, Documents) directly through the standardized MCP protocol, turning the system into a native tool for AI assistants.

## Confirmed Decisions
- Technology stack: Go with Echo framework, PostgreSQL for DB.
- Transport: MCP over SSE (Server-Sent Events) and HTTP.
- Hierarchy:
  - Project (Root)
    - Document (Child of Project)
    - User Story (Child of Project)
      - Task (Child of User Story)
- No custom frontend UI in this phase; the interface is the MCP protocol itself.

## Stories
- [US001 — MCP server setup](US001_mcp_server_setup.md)
- [US002 — Project CRUD](US002_project_crud.md)
- [US003 — Document CRUD](US003_document_crud.md)
- [US004 — User Story CRUD](US004_user_story_crud.md)
- [US005 — Task CRUD](US005_task_crud.md)

## Tasks
| US ID | Task ID | Track | Service | Name | Status | Blocked By |
|-------|---------|-------|---------|------|--------|------------|
| US001 | be_mcp_server | BE | services/agent-board-mcp | US001_be_mcp_server.md | pending | |
| US002 | be_schema_and_project_repo | BE | services/agent-board-mcp | US002_be_schema_and_project_repo.md | pending | |
| US002 | be_project_mcp_tools | BE | services/agent-board-mcp | US002_be_project_mcp_tools.md | pending | US001_be_mcp_server.md, US002_be_schema_and_project_repo.md |
| US003 | be_document_crud | BE | services/agent-board-mcp | US003_be_document_crud.md | pending | US001_be_mcp_server.md, US002_be_schema_and_project_repo.md |
| US004 | be_user_story_crud | BE | services/agent-board-mcp | US004_be_user_story_crud.md | pending | US001_be_mcp_server.md, US002_be_schema_and_project_repo.md |
| US005 | be_task_crud | BE | services/agent-board-mcp | US005_be_task_crud.md | pending | US001_be_mcp_server.md, US002_be_schema_and_project_repo.md |
| US006 | be_quality_refinements | BE | services/agent-board-mcp | US006_be_quality_refinements.md | pending | |
