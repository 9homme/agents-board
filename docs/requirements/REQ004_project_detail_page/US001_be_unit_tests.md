# US001 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live next to the code they exercise inside `services/agent-board/`.

## Coverage matrix

| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| GET project — happy path | unit | UT-US001-001 | services/agent-board / internal/handler | `handler.GetProject(c echo.Context)` |
| GET project — 404 not found | unit | UT-US001-002 | services/agent-board / internal/handler | `handler.GetProject(c echo.Context)` |
| GET project — 500 internal error | unit | UT-US001-003 | services/agent-board / internal/handler | `handler.GetProject(c echo.Context)` |
| GET project — integration round-trip | integration | IT-US001-001 | services/agent-board | `GET /api/v1/projects/{id}` |
| GET project — integration 404 | integration | IT-US001-002 | services/agent-board | `GET /api/v1/projects/{id}` |
| Route registered on Echo | integration | IT-US001-003 | services/agent-board / cmd/api-server | `main.go` router registration |

## Unit tests

### UT-US001-001 — GetProject handler: 200 happy path
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Function under test:** `handler.GetProject(c echo.Context) error`
- **Given:**
  - A mock `repo.ProjectRepository` that returns a `*domain.Project` with:
    - `ID`: `"123e4567-e89b-12d3-a456-426614174000"`
    - `Name`: `"E-commerce Website"`
    - `Description`: `"A new online store for electronics"`
    - `CreatedAt`: a `time.Time` value representing `2026-05-20T10:00:00Z`
    - `UpdatedAt`: a `time.Time` value representing `2026-05-20T10:00:00Z`
  - An Echo context with path param `:id` = `"123e4567-e89b-12d3-a456-426614174000"`
- **When:** call `handler.GetProject(c)`
- **Then:**
  - HTTP status 200
  - Response body is exactly:
    ```json
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "E-commerce Website",
      "description": "A new online store for electronics",
      "createdAt": "2026-05-20T10:00:00Z",
      "updatedAt": "2026-05-20T10:00:00Z"
    }
    ```
  - All six fields present (`id`, `name`, `description`, `createdAt`, `updatedAt`); no extra fields.
  - Timestamps formatted as `2006-01-02T15:04:05Z` (Go `time.Format` layout, matching existing `GetProjects` handler).
  - Response is a bare object (NOT wrapped in `{ "project": {...} }`).
- **Edge cases to also cover:**
  - `description` is `""` (empty string) — must serialize as `""` not `null`.
- **Architecture cite:** API contract §1 `GET /api/v1/projects/{projectId}`, 200 OK; note on bare singular-resource shape.

### UT-US001-002 — GetProject handler: 404 not found
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Function under test:** `handler.GetProject(c echo.Context) error`
- **Given:**
  - A mock `repo.ProjectRepository` that returns `repo.ErrNotFound` for any id.
  - An Echo context with path param `:id` = `"no-such-id"`
- **When:** call `handler.GetProject(c)`
- **Then:**
  - HTTP status 404
  - Response body is exactly:
    ```json
    { "code": "NOT_FOUND", "message": "Project not found" }
    ```
  - No other fields in the body.
- **Architecture cite:** API contract §1 `GET /api/v1/projects/{projectId}`, 404 Not Found.

### UT-US001-003 — GetProject handler: 500 internal error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Function under test:** `handler.GetProject(c echo.Context) error`
- **Given:**
  - A mock `repo.ProjectRepository` that returns a generic `errors.New("db connection refused")` for any id.
  - An Echo context with path param `:id` = `"any-id"`
- **When:** call `handler.GetProject(c)`
- **Then:**
  - HTTP status 500
  - Response body is exactly:
    ```json
    { "code": "INTERNAL_ERROR", "message": "Failed to fetch project" }
    ```
- **Architecture cite:** API contract §1 `GET /api/v1/projects/{projectId}`, 500 Internal Server Error.

## Integration tests

### IT-US001-001 — GET /api/v1/projects/{id} — found
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ real test DB (testcontainers or local Postgres configured via `DATABASE_URL` env var in the test)
- **Setup:**
  - Start (or connect to) a test-scoped Postgres instance.
  - Insert a project row directly via SQL or repo helper:
    - `id` = `"123e4567-e89b-12d3-a456-426614174000"`
    - `name` = `"Integration Test Project"`
    - `description` = `"desc"`
    - `created_at`, `updated_at` = `2026-05-20T10:00:00Z`
  - Boot the Echo router with the real repo connected to the test DB.
- **Endpoint exercised:** `GET /api/v1/projects/123e4567-e89b-12d3-a456-426614174000`
- **Request body:** none
- **Expect:**
  - 200 status
  - Body `id` = `"123e4567-e89b-12d3-a456-426614174000"`, `name` = `"Integration Test Project"`
  - All five fields present; `createdAt` and `updatedAt` in `2006-01-02T15:04:05Z` format.
- **Teardown:** Delete inserted row or drop test schema.

### IT-US001-002 — GET /api/v1/projects/{id} — not found
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ real test DB
- **Setup:** Test DB with no row matching the requested id.
- **Endpoint exercised:** `GET /api/v1/projects/00000000-0000-0000-0000-000000000000`
- **Request body:** none
- **Expect:**
  - 404 status
  - Body exactly `{"code":"NOT_FOUND","message":"Project not found"}`
- **Teardown:** none required (no data inserted).

### IT-US001-003 — Route registration smoke test
- **Service:** `services/agent-board`
- **Boundary:** `cmd/api-server/main.go` router instantiation
- **Setup:** Construct the Echo router as done in `main.go` (without starting the HTTP listener; use `httptest.NewServer` or verify routes with `e.Routes()`).
- **Endpoint exercised:** Confirm `GET /api/v1/projects/:id` is present in `e.Routes()`.
- **Expect:** Route entry with `Method=GET`, `Path=/api/v1/projects/:id` exists in the router's route table.
- **Notes:** This is a low-cost guard against accidental omission of the `e.GET(...)` call. Does NOT exercise the handler logic (covered by UT-US001-001/002/003).
