# US005 Backend Unit Tests — Task CRUD

## Backend Unit/Integration Tests

| ID | Component | Test Case | Expected Outcome |
|---|---|---|---|
| UT-020 | `internal/repo` | Create task in DB | Record inserted with parent User Story ID. |
| UT-021 | `internal/repo` | Get task from DB | Returns correct record or not found. |
| UT-022 | `internal/repo` | Update task in DB | Record updated. |
| UT-023 | `internal/repo` | Delete task in DB | Record removed. |
| UT-024 | `internal/repo` | List tasks by User Story | Returns all tasks for specific user story. |
| IT-017 | `internal/handler` | `create_task` tool call | Tool returns serialized task JSON. |
| IT-018 | `internal/handler` | `get_task` tool call | Tool returns serialized task JSON. |
| IT-019 | `internal/handler` | `update_task` tool call | Tool returns serialized task JSON. |
| IT-020 | `internal/handler` | `delete_task` tool call | Tool returns success status. |
| IT-021 | `internal/handler` | `list_tasks` tool call | Tool returns list of tasks. |

## Mocking Strategy
- Use `sqlmock` for database repository tests.
- Mock the repository in handler tests.
