# US005 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`.

## Coverage matrix
| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| Get user story detail returns 200 | integration | IT-001 | services/agent-board / internal/handler | `GET /api/v1/user-stories/{id}` |
| Get user story detail returns 404 | integration | IT-002 | services/agent-board / internal/handler | `GET /api/v1/user-stories/{id}` |
| Get user story tasks returns 200 with list | integration | IT-003 | services/agent-board / internal/handler | `GET /api/v1/user-stories/{id}/tasks` |
| Get user story tasks returns empty list | integration | IT-004 | services/agent-board / internal/handler | `GET /api/v1/user-stories/{id}/tasks` |
| Get user story tasks returns 404 for missing story | integration | IT-005 | services/agent-board / internal/handler | `GET /api/v1/user-stories/{id}/tasks` |
| Get user story detail returns 500 on repo error | integration | IT-006 | services/agent-board / internal/handler | `GET /api/v1/user-stories/{id}` |
| Get user story tasks returns 500 on repo error | integration | IT-007 | services/agent-board / internal/handler | `GET /api/v1/user-stories/{id}/tasks` |
| Handler GET story maps ErrNotFound to 404 | unit | UT-001 | services/agent-board / internal/handler | `HandleGetUserStory` |
| Handler GET story maps generic error to 500 | unit | UT-002 | services/agent-board / internal/handler | `HandleGetUserStory` |
| Handler GET tasks maps ErrNotFound to 404 | unit | UT-003 | services/agent-board / internal/handler | `HandleGetUserStoryTasks` |
| Handler GET tasks maps generic error to 500 | unit | UT-004 | services/agent-board / internal/handler | `HandleGetUserStoryTasks` |

## Unit tests
### UT-001 — Handler GET story maps ErrNotFound to 404
- **Service:** `services/agent-board`
- **Function under test:** `internal/handler.HandleGetUserStory` (or similar naming)
- **Given:** A mocked repo that returns `ErrNotFound`.
- **When:** GET `/api/v1/user-stories/{id}` is called.
- **Then:** Handler responds with `404 Not Found` and the correct JSON error body.

### UT-002 — Handler GET story maps generic error to 500
- **Service:** `services/agent-board`
- **Function under test:** `internal/handler.HandleGetUserStory`
- **Given:** A mocked repo that returns a generic `errors.New("db down")`.
- **When:** GET `/api/v1/user-stories/{id}` is called.
- **Then:** Handler responds with `500 Internal Server Error` and the correct JSON error body.

### UT-003 — Handler GET tasks maps ErrNotFound to 404
- **Service:** `services/agent-board`
- **Function under test:** `internal/handler.HandleGetUserStoryTasks`
- **Given:** A mocked repo that returns `ErrNotFound`.
- **When:** GET `/api/v1/user-stories/{id}/tasks` is called.
- **Then:** Handler responds with `404 Not Found` and the correct JSON error body.

### UT-004 — Handler GET tasks maps generic error to 500
- **Service:** `services/agent-board`
- **Function under test:** `internal/handler.HandleGetUserStoryTasks`
- **Given:** A mocked repo that returns a generic error.
- **When:** GET `/api/v1/user-stories/{id}/tasks` is called.
- **Then:** Handler responds with `500 Internal Server Error` and the correct JSON error body.

## Integration tests
### IT-001 — Get user story detail returns 200
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ DB
- **Setup:** A test DB with a user story.
- **Endpoint exercised:** `GET /api/v1/user-stories/{id}`
- **Expect:** Status `200`. Body matches `{"id", "projectId", "title", "description", "status", "createdAt", "updatedAt"}`. Note: `taskCount` is NOT present.
- **Teardown:** Clean up DB rows.

### IT-002 — Get user story detail returns 404
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ DB
- **Setup:** A test DB with no matching user story.
- **Endpoint exercised:** `GET /api/v1/user-stories/{missing-id}`
- **Expect:** Status `404`. Body is `{ "code": "NOT_FOUND", "message": "User story not found" }`.

### IT-003 — Get user story tasks returns 200 with list
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ DB
- **Setup:** A test DB with a user story and associated tasks.
- **Endpoint exercised:** `GET /api/v1/user-stories/{id}/tasks`
- **Expect:** Status `200`. Body matches `{"tasks":[{id, userStoryId, title, description, status, createdAt, updatedAt}]}`. Ordered by `createdAt DESC`.
- **Teardown:** Clean up DB rows.

### IT-004 — Get user story tasks returns empty list
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ DB
- **Setup:** A test DB with a user story that has 0 tasks.
- **Endpoint exercised:** `GET /api/v1/user-stories/{id}/tasks`
- **Expect:** Status `200`. Body is `{"tasks":[]}`.

### IT-005 — Get user story tasks returns 404 for missing story
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ DB
- **Setup:** A test DB with no matching user story.
- **Endpoint exercised:** `GET /api/v1/user-stories/{missing-id}/tasks`
- **Expect:** Status `404`. Body is `{ "code": "NOT_FOUND", "message": "User story not found" }`.

### IT-006 — Get user story detail returns 500 on repo error
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo
- **Setup:** Mock or intercept repo to force an error.
- **Endpoint exercised:** `GET /api/v1/user-stories/{id}`
- **Expect:** Status `500`. Body is `{ "code": "INTERNAL_ERROR", "message": "Internal server error" }`.

### IT-007 — Get user story tasks returns 500 on repo error
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo
- **Setup:** Mock or intercept repo to force an error.
- **Endpoint exercised:** `GET /api/v1/user-stories/{id}/tasks`
- **Expect:** Status `500`. Body is `{ "code": "INTERNAL_ERROR", "message": "Internal server error" }`.

## Coverage exemption
Skipped UTs for repo because `GetUserStory` and `ListTasks` are marked as unchanged/existing in the architecture, so they are already covered. Only handler integration tests and handler unit tests are needed for the new routes.

## Spec change log
### Revision 1 — 2024-03-XX — driver: po-ba sign-off pass
- Added IT-006 and IT-007 to test 500 error scenarios for new endpoints.
- Added UT-001 through UT-004 to explicitly cover handler error mapping paths (ErrNotFound -> 404, generic -> 500).
