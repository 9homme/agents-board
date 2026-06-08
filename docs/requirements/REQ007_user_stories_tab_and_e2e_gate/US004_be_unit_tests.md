# US004 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`.

## Coverage matrix
| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| Get project user stories returns 200 with list + task count | integration | IT-001 | services/agent-board / internal/handler | `GET /api/v1/projects/{id}/user-stories` |
| Returns 404 for missing project | integration | IT-002 | services/agent-board / internal/handler | `GET /api/v1/projects/{id}/user-stories` |
| Returns 404 for invalid project ID format | integration | IT-003 | services/agent-board / internal/handler | `GET /api/v1/projects/{id}/user-stories` |
| Returns 500 on repository failure | integration | IT-004 | services/agent-board / internal/handler | `GET /api/v1/projects/{id}/user-stories` |
| Repo list user stories with task count | unit | UT-001 | services/agent-board / internal/repo | `ListUserStoriesWithTaskCount` |
| Repo returns empty list when no stories exist | unit | UT-002 | services/agent-board / internal/repo | `ListUserStoriesWithTaskCount` |
| Repo returns error on query execution failure | unit | UT-003 | services/agent-board / internal/repo | `ListUserStoriesWithTaskCount` |
| Repo returns error on row scan failure | unit | UT-004 | services/agent-board / internal/repo | `ListUserStoriesWithTaskCount` |
| Repo returns error on rows iteration failure | unit | UT-005 | services/agent-board / internal/repo | `ListUserStoriesWithTaskCount` |

## Unit tests
### UT-001 — Repo list user stories with task count
- **Service:** `services/agent-board`
- **Function under test:** `internal/repo.ListUserStoriesWithTaskCount`
- **Given:** A mocked `sqlx.DB` (via sqlmock). Two stories for a project, one with 2 tasks, one with 0 tasks.
- **When:** call `ListUserStoriesWithTaskCount`
- **Then:** returns no error and the list of stories with the correct `taskCount` aggregate, ordered by `created_at DESC`.

### UT-002 — Repo returns empty list when no stories exist
- **Service:** `services/agent-board`
- **Function under test:** `internal/repo.ListUserStoriesWithTaskCount`
- **Given:** A mocked `sqlx.DB` that returns no rows.
- **When:** call `ListUserStoriesWithTaskCount`
- **Then:** returns an empty slice and no error.

### UT-003 — Repo returns error on query execution failure
- **Service:** `services/agent-board`
- **Function under test:** `internal/repo.ListUserStoriesWithTaskCount`
- **Given:** A mocked `sqlx.DB` that returns an error on query execution.
- **When:** call `ListUserStoriesWithTaskCount`
- **Then:** returns the exact error.

### UT-004 — Repo returns error on row scan failure
- **Service:** `services/agent-board`
- **Function under test:** `internal/repo.ListUserStoriesWithTaskCount`
- **Given:** A mocked `sqlx.DB` returning a row that causes `Scan` to fail (e.g. type mismatch).
- **When:** call `ListUserStoriesWithTaskCount`
- **Then:** returns the scan error.

### UT-005 — Repo returns error on rows iteration failure
- **Service:** `services/agent-board`
- **Function under test:** `internal/repo.ListUserStoriesWithTaskCount`
- **Given:** A mocked `sqlx.DB` returning rows with `rows.Err()` set to an error after iteration.
- **When:** call `ListUserStoriesWithTaskCount`
- **Then:** returns the iteration error.

## Integration tests
### IT-001 — Get project user stories returns 200 with list
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ DB
- **Setup:** A test DB with a project, two user stories, and tasks linked to them.
- **Endpoint exercised:** `GET /api/v1/projects/{id}/user-stories`
- **Expect:** Status `200`. Body matches `{"userStories":[{id, projectId, title, description, status, taskCount, createdAt, updatedAt}]}`. `taskCount` must match the seed. Ordered by `createdAt DESC`.
- **Teardown:** Clean up DB rows.

### IT-002 — Returns 404 for missing project
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ DB
- **Setup:** A test DB with no projects matching the requested ID.
- **Endpoint exercised:** `GET /api/v1/projects/{missing-id}/user-stories`
- **Expect:** Status `404`. Body is `{ "code": "NOT_FOUND", "message": "Project not found" }`.

### IT-003 — Returns 404 for invalid project ID format
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo
- **Endpoint exercised:** `GET /api/v1/projects/invalid-uuid/user-stories`
- **Expect:** Status `404`. Body is `{ "code": "NOT_FOUND", "message": "Project not found" }`.

### IT-004 — Returns 500 on repository failure
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo
- **Setup:** Mock or intercept repo to force an error.
- **Endpoint exercised:** `GET /api/v1/projects/{id}/user-stories`
- **Expect:** Status `500`. Body is `{ "code": "INTERNAL_ERROR", "message": "Internal server error" }`.

## Spec change log
### Revision 1 — 2024-03-XX — driver: po-ba sign-off pass
- committed IT-003 to 404 specifically per architecture contract.
- added IT-004 to cover 500 Internal Error from repository failure.
- expanded UT-001 into distinct UT-001, UT-003, UT-004, UT-005 for query, scan, and rows iteration errors.