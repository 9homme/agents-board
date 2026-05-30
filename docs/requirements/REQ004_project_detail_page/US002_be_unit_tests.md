# US002 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live next to the code they exercise inside `services/agent-board/`.

## Coverage matrix

| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| List documents — happy path (multiple) | unit | UT-US002-001 | services/agent-board / internal/handler | `handler.ListProjectDocuments(c echo.Context)` |
| List documents — empty list (project exists, zero docs) | unit | UT-US002-002 | services/agent-board / internal/handler | `handler.ListProjectDocuments(c echo.Context)` |
| List documents — project not found 404 | unit | UT-US002-003 | services/agent-board / internal/handler | `handler.ListProjectDocuments(c echo.Context)` |
| List documents — 500 on project lookup failure | unit | UT-US002-004 | services/agent-board / internal/handler | `handler.ListProjectDocuments(c echo.Context)` |
| List documents — 500 on document lookup failure | unit | UT-US002-005 | services/agent-board / internal/handler | `handler.ListProjectDocuments(c echo.Context)` |
| List documents — content field absent from response | unit | UT-US002-006 | services/agent-board / internal/handler | `handler.ListProjectDocuments(c echo.Context)` |
| Get document — happy path | unit | UT-US002-007 | services/agent-board / internal/handler | `handler.GetDocument(c echo.Context)` |
| Get document — 404 not found | unit | UT-US002-008 | services/agent-board / internal/handler | `handler.GetDocument(c echo.Context)` |
| Get document — 500 internal error | unit | UT-US002-009 | services/agent-board / internal/handler | `handler.GetDocument(c echo.Context)` |
| Repo: ListDocuments orders by updated_at DESC, id DESC | unit | UT-US002-010 | services/agent-board / internal/repo | `repo.DocumentRepository.ListDocuments` |
| List documents — integration round-trip (D-006: missing project → 404 not empty list) | integration | IT-US002-001 | services/agent-board | `GET /api/v1/projects/{id}/documents` |
| List documents — integration: project exists, empty list returns `{"documents":[]}` | integration | IT-US002-002 | services/agent-board | `GET /api/v1/projects/{id}/documents` |
| List documents — integration: ordering verified | integration | IT-US002-003 | services/agent-board | `GET /api/v1/projects/{id}/documents` |
| Get document — integration round-trip | integration | IT-US002-004 | services/agent-board | `GET /api/v1/documents/{id}` |
| Get document — integration 404 | integration | IT-US002-005 | services/agent-board | `GET /api/v1/documents/{id}` |
| Routes registered on Echo | integration | IT-US002-006 | services/agent-board / cmd/api-server | `main.go` router registration |

## Unit tests

### UT-US002-001 — ListProjectDocuments handler: 200 with multiple documents
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Function under test:** `handler.ListProjectDocuments(c echo.Context) error`
- **Given:**
  - A mock `repo.ProjectRepository.GetProject` that returns a valid `*domain.Project` (id = `"123e4567-e89b-12d3-a456-426614174000"`).
  - A mock `repo.DocumentRepository.ListDocuments` that returns two `*domain.Document` values (in the order the SQL already returns them — repo layer owns ordering):
    1. `ID="d111aaaa-1111-1111-1111-111111111111"`, `ProjectID="123e4567-e89b-12d3-a456-426614174000"`, `Title="Architecture overview"`, `CreatedAt=2026-05-18T08:30:00Z`, `UpdatedAt=2026-05-20T09:45:00Z`
    2. `ID="d222bbbb-2222-2222-2222-222222222222"`, `ProjectID="123e4567-e89b-12d3-a456-426614174000"`, `Title="Onboarding guide"`, `CreatedAt=2026-05-15T11:00:00Z`, `UpdatedAt=2026-05-19T16:20:00Z`
  - An Echo context with path param `:id` = `"123e4567-e89b-12d3-a456-426614174000"`.
- **When:** call `handler.ListProjectDocuments(c)`
- **Then:**
  - HTTP status 200
  - Response body is exactly:
    ```json
    {
      "documents": [
        {
          "id": "d111aaaa-1111-1111-1111-111111111111",
          "projectId": "123e4567-e89b-12d3-a456-426614174000",
          "title": "Architecture overview",
          "createdAt": "2026-05-18T08:30:00Z",
          "updatedAt": "2026-05-20T09:45:00Z"
        },
        {
          "id": "d222bbbb-2222-2222-2222-222222222222",
          "projectId": "123e4567-e89b-12d3-a456-426614174000",
          "title": "Onboarding guide",
          "createdAt": "2026-05-15T11:00:00Z",
          "updatedAt": "2026-05-19T16:20:00Z"
        }
      ]
    }
    ```
  - The `documents` array is present and is an array (not `null`).
  - Per-item fields: `id`, `projectId`, `title`, `createdAt`, `updatedAt`. The `content` field is absent.
- **Architecture cite:** API contract §2 `GET /api/v1/projects/{projectId}/documents`, 200 OK.

### UT-US002-002 — ListProjectDocuments handler: 200 empty list (project exists, no documents)
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Function under test:** `handler.ListProjectDocuments(c echo.Context) error`
- **Given:**
  - A mock `repo.ProjectRepository.GetProject` that returns a valid `*domain.Project`.
  - A mock `repo.DocumentRepository.ListDocuments` that returns an empty slice (`[]domain.Document{}`).
  - An Echo context with path param `:id` = `"123e4567-e89b-12d3-a456-426614174000"`.
- **When:** call `handler.ListProjectDocuments(c)`
- **Then:**
  - HTTP status 200
  - Response body is exactly `{"documents":[]}` — the `documents` key maps to an empty JSON array (never `null`).
- **Notes:** The BE must initialize the slice with `make([]documentListItem, 0)` before marshalling (existing pattern from `GetProjects`).
- **Architecture cite:** API contract §2, note "documents is always an array (never null). Empty list → `{ "documents": [] }`".

### UT-US002-003 — ListProjectDocuments handler: 404 project not found (D-006)
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Function under test:** `handler.ListProjectDocuments(c echo.Context) error`
- **Given:**
  - A mock `repo.ProjectRepository.GetProject` that returns `repo.ErrNotFound`.
  - `repo.DocumentRepository.ListDocuments` should NOT be called (verify with mock assertion).
  - An Echo context with path param `:id` = `"no-such-project"`.
- **When:** call `handler.ListProjectDocuments(c)`
- **Then:**
  - HTTP status 404
  - Response body is exactly:
    ```json
    { "code": "NOT_FOUND", "message": "Project not found" }
    ```
  - `ListDocuments` mock was NOT called (assert zero invocations).
- **Architecture cite:** API contract §2, 404 Not Found; D-006 — "do NOT return `{ "documents": [] }` for an unknown project".

### UT-US002-004 — ListProjectDocuments handler: 500 on project lookup failure
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Function under test:** `handler.ListProjectDocuments(c echo.Context) error`
- **Given:**
  - A mock `repo.ProjectRepository.GetProject` that returns a generic `errors.New("connection pool exhausted")`.
  - An Echo context with path param `:id` = `"any-id"`.
- **When:** call `handler.ListProjectDocuments(c)`
- **Then:**
  - HTTP status 500
  - Response body is exactly:
    ```json
    { "code": "INTERNAL_ERROR", "message": "Failed to fetch documents" }
    ```
- **Architecture cite:** API contract §2, 500 Internal Server Error.

### UT-US002-005 — ListProjectDocuments handler: 500 on document list failure
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Function under test:** `handler.ListProjectDocuments(c echo.Context) error`
- **Given:**
  - A mock `repo.ProjectRepository.GetProject` that returns a valid `*domain.Project`.
  - A mock `repo.DocumentRepository.ListDocuments` that returns a generic error.
  - An Echo context with path param `:id` = `"123e4567-e89b-12d3-a456-426614174000"`.
- **When:** call `handler.ListProjectDocuments(c)`
- **Then:**
  - HTTP status 500
  - Response body is exactly `{"code":"INTERNAL_ERROR","message":"Failed to fetch documents"}`.
- **Architecture cite:** API contract §2, 500 Internal Server Error.

### UT-US002-006 — ListProjectDocuments handler: content field absent from response items
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Function under test:** `handler.ListProjectDocuments(c echo.Context) error`
- **Given:**
  - Mock repo returns one document where the domain struct has `Content = "# Very long markdown…"`.
  - An Echo context with path param `:id` set to the project's id.
- **When:** call `handler.ListProjectDocuments(c)` and unmarshal the response body.
- **Then:**
  - The unmarshalled item does NOT contain a `content` key (use `json.RawMessage` or map-based assertion to confirm key absence, not just value absence).
- **Architecture cite:** API contract §2 — "content is intentionally absent from this response shape".

### UT-US002-007 — GetDocument handler: 200 happy path
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Function under test:** `handler.GetDocument(c echo.Context) error`
- **Given:**
  - A mock `repo.DocumentRepository.GetDocument` returning `*domain.Document`:
    - `ID="d111aaaa-1111-1111-1111-111111111111"`, `ProjectID="123e4567-e89b-12d3-a456-426614174000"`, `Title="Architecture overview"`, `Content="# Architecture\n\nThis project uses…\n\n\`\`\`mermaid\ngraph TD; A-->B;\n\`\`\`\n"`, `CreatedAt=2026-05-18T08:30:00Z`, `UpdatedAt=2026-05-20T09:45:00Z`.
  - An Echo context with path param `:id` = `"d111aaaa-1111-1111-1111-111111111111"`.
- **When:** call `handler.GetDocument(c)`
- **Then:**
  - HTTP status 200
  - Response body is exactly:
    ```json
    {
      "id": "d111aaaa-1111-1111-1111-111111111111",
      "projectId": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Architecture overview",
      "content": "# Architecture\n\nThis project uses…\n\n```mermaid\ngraph TD; A-->B;\n```\n",
      "createdAt": "2026-05-18T08:30:00Z",
      "updatedAt": "2026-05-20T09:45:00Z"
    }
    ```
  - All six fields present (`id`, `projectId`, `title`, `content`, `createdAt`, `updatedAt`).
  - `content` is a string (may be `""`); never `null`.
- **Edge cases to also cover:**
  - `content` is `""` — must serialize as `""` not `null`.
- **Architecture cite:** API contract §3 `GET /api/v1/documents/{documentId}`, 200 OK.

### UT-US002-008 — GetDocument handler: 404 not found
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Function under test:** `handler.GetDocument(c echo.Context) error`
- **Given:**
  - A mock `repo.DocumentRepository.GetDocument` returning `repo.ErrNotFound`.
  - An Echo context with path param `:id` = `"no-such-doc"`.
- **When:** call `handler.GetDocument(c)`
- **Then:**
  - HTTP status 404
  - Response body is exactly:
    ```json
    { "code": "NOT_FOUND", "message": "Document not found" }
    ```
- **Architecture cite:** API contract §3, 404 Not Found.

### UT-US002-009 — GetDocument handler: 500 internal error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Function under test:** `handler.GetDocument(c echo.Context) error`
- **Given:**
  - A mock `repo.DocumentRepository.GetDocument` returning a generic error.
  - An Echo context with path param `:id` = `"any-id"`.
- **When:** call `handler.GetDocument(c)`
- **Then:**
  - HTTP status 500
  - Response body is exactly:
    ```json
    { "code": "INTERNAL_ERROR", "message": "Failed to fetch document" }
    ```
- **Architecture cite:** API contract §3, 500 Internal Server Error.

### UT-US002-010 — Repo: ListDocuments SQL orders by updated_at DESC, id DESC
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Function under test:** `repo.DocumentRepository.ListDocuments(ctx, projectID)`
- **Given:**
  - A test DB (or SQL mock capturing the query string) with three documents for the same project id:
    - Doc A: `updated_at = 2026-05-20T10:00:00Z`, `id = "aaaa0001-..."`
    - Doc B: `updated_at = 2026-05-19T10:00:00Z`, `id = "bbbb0002-..."`
    - Doc C: `updated_at = 2026-05-20T10:00:00Z`, `id = "cccc0003-..."` (same `updated_at` as A — tiebreaker test)
- **When:** call `ListDocuments(ctx, projectID)`
- **Then:**
  - The returned slice has length 3.
  - Order is: C (same timestamp as A, but `cccc > aaaa` as strings so id DESC puts C before A), then A, then B.
    - Equivalently: `updated_at DESC` first, then `id DESC` for docs sharing the same `updated_at`.
- **Notes:** This test directly verifies the SQL change from `ORDER BY created_at DESC` to `ORDER BY updated_at DESC, id DESC` specified in the architecture (§"Data access").
- **Architecture cite:** §"Data access" — `ORDER BY updated_at DESC, id DESC`.

## Integration tests

### IT-US002-001 — GET /api/v1/projects/{id}/documents — missing project returns 404 not empty list (D-006)
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ real test DB
- **Setup:** Test DB with no project matching the requested id. No documents inserted.
- **Endpoint exercised:** `GET /api/v1/projects/00000000-0000-0000-0000-000000000000/documents`
- **Expect:**
  - 404 status
  - Body exactly `{"code":"NOT_FOUND","message":"Project not found"}`
  - Body does NOT contain `"documents"` key.
- **Architecture cite:** D-006; API contract §2, 404.

### IT-US002-002 — GET /api/v1/projects/{id}/documents — project exists, zero documents returns `{"documents":[]}`
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ real test DB
- **Setup:**
  - Insert project with `id = "123e4567-e89b-12d3-a456-426614174001"`.
  - Do NOT insert any documents for that project.
- **Endpoint exercised:** `GET /api/v1/projects/123e4567-e89b-12d3-a456-426614174001/documents`
- **Expect:**
  - 200 status
  - Body is exactly `{"documents":[]}` — array present, array is empty, array is not `null`.
- **Architecture cite:** API contract §2 — "documents is always an array"; D-006 distinguishing from missing project.

### IT-US002-003 — GET /api/v1/projects/{id}/documents — ordering verified
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ real test DB
- **Setup:**
  - Insert project with id `"123e4567-e89b-12d3-a456-426614174002"`.
  - Insert three documents:
    - Doc B: `title="B"`, `updated_at = 2026-05-19T10:00:00Z`, `id` lexicographically lower.
    - Doc A1: `title="A1"`, `updated_at = 2026-05-20T10:00:00Z`, `id` lexicographically lower (tiebreaker set 1).
    - Doc A2: `title="A2"`, `updated_at = 2026-05-20T10:00:00Z`, `id` lexicographically higher (tiebreaker set 2).
- **Endpoint exercised:** `GET /api/v1/projects/123e4567-e89b-12d3-a456-426614174002/documents`
- **Expect:**
  - 200 status
  - `documents[0].title = "A2"` (higher `id` wins tiebreaker when `updated_at` is equal)
  - `documents[1].title = "A1"`
  - `documents[2].title = "B"` (oldest `updated_at` last)
- **Architecture cite:** API contract §2 — "Order: updatedAt desc, then id desc as a stable tiebreaker".

### IT-US002-004 — GET /api/v1/documents/{id} — found
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ real test DB
- **Setup:**
  - Insert project and a document with known `id`, `title`, `content = "# Hello"`.
- **Endpoint exercised:** `GET /api/v1/documents/{id}`
- **Expect:**
  - 200 status
  - Body contains `id`, `projectId`, `title`, `content = "# Hello"`, `createdAt`, `updatedAt` — all six fields; `createdAt`/`updatedAt` in `2006-01-02T15:04:05Z` format.
- **Teardown:** Delete inserted rows.

### IT-US002-005 — GET /api/v1/documents/{id} — not found
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ real test DB
- **Setup:** Test DB with no document matching the requested id.
- **Endpoint exercised:** `GET /api/v1/documents/00000000-0000-0000-0000-000000000000`
- **Expect:**
  - 404 status
  - Body exactly `{"code":"NOT_FOUND","message":"Document not found"}`.

### IT-US002-006 — Route registration smoke test for both new document routes
- **Service:** `services/agent-board`
- **Boundary:** `cmd/api-server/main.go` router instantiation
- **Setup:** Construct the Echo router as done in `main.go`.
- **Expect:**
  - Route `GET /api/v1/projects/:id/documents` is present in `e.Routes()`.
  - Route `GET /api/v1/documents/:id` is present in `e.Routes()`.
- **Notes:** Low-cost guard; does not exercise handler logic.
