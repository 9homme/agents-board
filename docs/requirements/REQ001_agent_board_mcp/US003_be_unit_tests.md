# US003 Backend Unit Tests — Document CRUD

## Backend Unit/Integration Tests

| ID | Component | Test Case | Expected Outcome |
|---|---|---|---|
| UT-010 | `internal/repo` | Create document in DB | Record inserted with parent Project ID. |
| UT-011 | `internal/repo` | Get document from DB | Returns correct record or not found. |
| UT-012 | `internal/repo` | Update document in DB | Record updated. |
| UT-013 | `internal/repo` | Delete document in DB | Record removed. |
| UT-014 | `internal/repo` | List documents by Project | Returns all documents for specific project. |
| IT-007 | `internal/handler` | `create_document` tool call | Tool returns serialized document JSON. |
| IT-008 | `internal/handler` | `get_document` tool call | Tool returns serialized document JSON. |
| IT-009 | `internal/handler` | `update_document` tool call | Tool returns serialized document JSON. |
| IT-010 | `internal/handler` | `delete_document` tool call | Tool returns success status. |
| IT-011 | `internal/handler` | `list_documents` tool call | Tool returns list of documents. |

## Mocking Strategy
- Use `sqlmock` for database repository tests.
- Mock the repository in handler tests.
