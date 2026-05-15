# US001 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live next to the code they exercise inside the relevant `services/<service-name>/` module.

## Coverage matrix
| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| Successfully load project list | unit | UT-001 | services/agent-board / internal/handler | `handler.GetProjects(...)` |
| Error state | unit | UT-002 | services/agent-board / internal/handler | `handler.GetProjects(...)` |
| Successfully load project list | integration | IT-001 | services/agent-board | `GET /api/v1/projects` |

## Unit tests
### UT-001 — Successfully load project list
- **Service:** `services/agent-board`
- **Function under test:** `internal/handler.GetProjects`
- **Given:** A mocked repository that returns a list of projects.
- **When:** call the HTTP handler for `GET /api/v1/projects`.
- **Then:** returns 200 OK, and the JSON response contains the list of projects matching the architecture contract.
- **Edge cases to also cover:** Repository returning an empty list (empty state).
- **Architecture cite:** API contract row `GET /api/v1/projects`, 200 OK response.

### UT-002 — Error state
- **Service:** `services/agent-board`
- **Function under test:** `internal/handler.GetProjects`
- **Given:** A mocked repository that returns an error.
- **When:** call the HTTP handler for `GET /api/v1/projects`.
- **Then:** returns 500 Internal Server Error, and the JSON response matches the `INTERNAL_ERROR` shape in the architecture.
- **Architecture cite:** API contract row `GET /api/v1/projects`, 500 Internal Server Error.

## Integration tests
### IT-001 — Fetch projects end-to-end (DB)
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ in-memory/test DB
- **Setup:** Initialize the test DB, insert some test projects directly via the repository. Initialize the `api-server` router.
- **Endpoint exercised:** `GET /api/v1/projects`
- **Request body:** None
- **Expect:** 200 OK, body matches the JSON contract and contains the inserted projects.
- **Teardown:** Clean up test DB.