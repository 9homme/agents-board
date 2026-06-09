# US004 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`.

## Coverage matrix
| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| Get project user stories returns 200 with list + task count | integration | IT-001 | services/agent-board / internal/handler | `GET /api/v1/projects/{id}/user-stories` |
| Returns 404 for missing project | integration | IT-002 | services/agent-board / internal/handler | `GET /api/v1/projects/{id}/user-stories` |
| Returns 500 for malformed (non-UUID) project ID | integration | IT-003 | services/agent-board / internal/handler | `GET /api/v1/projects/{id}/user-stories` |
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

### IT-003 — Returns 500 for malformed (non-UUID) project ID
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ DB
- **Setup:** A real test DB (testcontainers or a shared test instance). Pass the literal string `"invalid-uuid"` as the path parameter.
- **Endpoint exercised:** `GET /api/v1/projects/invalid-uuid/user-stories`
- **Rationale:** Postgres rejects a non-UUID string passed to a `uuid`-typed column with "invalid input syntax for type uuid". This is not `sql.ErrNoRows`, so `GetProject` wraps it as a generic error (not `ErrNotFound`). The handler falls through to the 500 branch — identical to the sibling `ListProjectDocuments` handler. There is no UUID validation gate before the repo call; adding one would be a handler code change outside this story's scope. This behavior is consistent with the existing codebase pattern.
- **Expect:** Status `500`. Body is `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch user stories" }`.

### IT-004 — Returns 500 on repository failure
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo
- **Setup:** Inject a mock `ProjectRepository` whose `GetProject` returns a non-`ErrNotFound` error (e.g. `errors.New("db connection lost")`), wired into the handler under test via `httptest`.
- **Endpoint exercised:** `GET /api/v1/projects/{valid-uuid}/user-stories`
- **Expect:** Status `500`. Body is `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch user stories" }`.

## Spec change log
### Revision 1 — 2024-03-XX — driver: po-ba sign-off pass
- committed IT-003 to 404 specifically per architecture contract.
- added IT-004 to cover 500 Internal Error from repository failure.
- expanded UT-001 into distinct UT-001, UT-003, UT-004, UT-005 for query, scan, and rows iteration errors.

### Revision 2 — 2026-06-08 — driver: tech-lead review gap finding
- changed IT-003 — changed expected status from 404 to 500 and updated description name from "Returns 404 for invalid project ID format" to "Returns 500 for malformed (non-UUID) project ID". Rationale: live Postgres testing confirmed that an invalid UUID causes "invalid input syntax for type uuid" — a DB-level type error that is NOT `sql.ErrNoRows`. `GetProject` wraps it as a generic error (not `ErrNotFound`); the handler falls to the 500 branch. This is consistent with the sibling `ListProjectDocuments` handler. The architecture's 404 contract says "project does not exist" — a malformed input is semantically distinct and there is no UUID validation gate in the production code. Aligning the spec to real behavior is correct; adding a validation gate would require an out-of-scope handler code change and would leave the sibling handler inconsistent.
- changed IT-004 — corrected the expected response body `message` from `"Internal server error"` to `"Failed to fetch user stories"` to match the architecture contract and the handler's actual error string. Also tightened the setup description to use a mock `ProjectRepository` via `httptest` rather than vague "intercept repo".