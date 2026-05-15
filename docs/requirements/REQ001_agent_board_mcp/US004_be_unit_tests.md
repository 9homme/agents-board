# US004 Backend Unit Tests — User Story CRUD

## Backend Unit/Integration Tests

| ID | Component | Test Case | Expected Outcome |
|---|---|---|---|
| UT-015 | `internal/repo` | Create user story in DB | Record inserted with parent Project ID and status. |
| UT-016 | `internal/repo` | Get user story from DB | Returns correct record or not found. |
| UT-017 | `internal/repo` | Update user story in DB | Record updated including status. |
| UT-018 | `internal/repo` | Delete user story in DB | Record removed; children (tasks) cascadingly removed. |
| UT-019 | `internal/repo` | List user stories by Project | Returns all user stories for specific project. |
| IT-012 | `internal/handler` | `create_user_story` tool call | Tool returns serialized user story JSON. |
| IT-013 | `internal/handler` | `get_user_story` tool call | Tool returns serialized user story JSON. |
| IT-014 | `internal/handler` | `update_user_story` tool call | Tool returns serialized user story JSON. |
| IT-015 | `internal/handler` | `delete_user_story` tool call | Tool returns success status. |
| IT-016 | `internal/handler` | `list_user_stories` tool call | Tool returns list of user stories. |

## Mocking Strategy
- Use `sqlmock` for database repository tests.
- Mock the repository in handler tests.
