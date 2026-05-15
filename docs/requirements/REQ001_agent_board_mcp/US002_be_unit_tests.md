# US002 Backend Unit Tests — Project CRUD

## Backend Unit/Integration Tests

| ID | Component | Test Case | Expected Outcome |
|---|---|---|---|
| UT-005 | `internal/repo` | Create project in DB | Record inserted, timestamps generated, ID returned. |
| UT-006 | `internal/repo` | Get project from DB | Returns correct record or not found error. |
| UT-007 | `internal/repo` | Update project in DB | Record updated, `updated_at` refreshed. |
| UT-008 | `internal/repo` | Delete project in DB | Record removed; children (docs/stories) cascadingly removed. |
| UT-009 | `internal/repo` | List projects in DB | Returns all projects. |
| IT-002 | `internal/handler` | `create_project` tool call | Tool returns serialized project JSON matching contract. |
| IT-003 | `internal/handler` | `get_project` tool call | Tool returns serialized project JSON or error if missing. |
| IT-004 | `internal/handler` | `update_project` tool call | Tool returns serialized updated project JSON. |
| IT-005 | `internal/handler` | `delete_project` tool call | Tool returns success status. |
| IT-006 | `internal/handler` | `list_projects` tool call | Tool returns list of projects. |

## Mocking Strategy
- Use `sqlmock` for database repository tests.
- Mock the repository in handler tests to isolate tool routing from DB logic.
