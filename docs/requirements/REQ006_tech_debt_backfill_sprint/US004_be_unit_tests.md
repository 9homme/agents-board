# US004 — Backend unit & integration test specification
# `project_tools.go` error-mapping tests

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live in `services/agent-board/internal/handler/project_tools_test.go`. Production code in `project_tools.go` is **byte-for-byte unchanged**.

**Key semantic for `project_tools.go`:** most repo errors are **passed through** (not wrapped with `fmt.Errorf`). Use `errors.Is(returnedErr, mockErr)` for generic-error assertions, not substring match on a wrap prefix. `_NotFound` cases return a fresh `errors.New("project not found")` — assert via `assert.Contains(t, err.Error(), "project not found")` AND `assert.False(t, errors.Is(err, repo.ErrNotFound))`.

**Harness shape (architecture.md §4.3):** construct a real `mcp.NewToolRegistry()`, instantiate a hand-written `MockProjectRepo` struct (or `testify/mock` — pick once per file), call `handler.RegisterProjectTools(registry, mockRepo)`, retrieve the tool closure via `registry.GetTool("tool-name")`, invoke it with `json.RawMessage`, assert the `(interface{}, error)` pair. No `httptest` — these are pure function calls.

## Coverage matrix

| AC scenario | Layer | Test ID | Function under test |
|---|---|---|---|
| `RegisterProjectTools` registers all 5 tools | unit | UT-001 | `RegisterProjectTools` |
| `create_project` invalid JSON | unit | UT-002 | `handleCreateProject` |
| `create_project` empty name | unit | UT-003 | `handleCreateProject` |
| `create_project` repo error (passthrough) | unit | UT-004 | `handleCreateProject` |
| `get_project` invalid JSON | unit | UT-005 | `handleGetProject` |
| `get_project` empty id | unit | UT-006 | `handleGetProject` |
| `get_project` repo ErrNotFound | unit | UT-007 | `handleGetProject` |
| `get_project` repo generic error (passthrough) | unit | UT-008 | `handleGetProject` |
| `update_project` invalid JSON | unit | UT-009 | `handleUpdateProject` |
| `update_project` empty id | unit | UT-010 | `handleUpdateProject` |
| `update_project` initial Get returns ErrNotFound | unit | UT-011 | `handleUpdateProject` |
| `update_project` initial Get returns generic error | unit | UT-012 | `handleUpdateProject` |
| `update_project` name provided but empty | unit | UT-013 | `handleUpdateProject` |
| `update_project` UpdateProject repo error | unit | UT-014 | `handleUpdateProject` |
| `delete_project` invalid JSON | unit | UT-015 | `handleDeleteProject` |
| `delete_project` empty id | unit | UT-016 | `handleDeleteProject` |
| `delete_project` repo error (passthrough) | unit | UT-017 | `handleDeleteProject` |
| `list_projects` repo error (passthrough) | unit | UT-018 | `handleListProjects` |
| per-file coverage ≥95% | integration | IT-001 | `project_tools.go` all functions |
| full suite still passes | integration | IT-002 | `go test ./...` |

## Unit tests

### UT-001 — `TestRegisterProjectTools_RegistersAllFiveTools`
- **Function under test:** `RegisterProjectTools`
- **Given:**
  ```go
  registry := mcp.NewToolRegistry()
  mockRepo := &MockProjectRepo{}
  handler.RegisterProjectTools(registry, mockRepo)
  ```
- **When:** each expected tool name is queried
- **Then:**
  - `registry.GetTool("create_project")` returns `(handler, true)` — handler is non-nil
  - `registry.GetTool("get_project")` returns `(handler, true)`
  - `registry.GetTool("update_project")` returns `(handler, true)`
  - `registry.GetTool("delete_project")` returns `(handler, true)`
  - `registry.GetTool("list_projects")` returns `(handler, true)`
  - `registry.GetTool("nonexistent_tool")` returns `(nil, false)`
- **Architecture cite:** architecture.md §4.3 `_RegistersAll*Tools` branch; tech_debt.md line 55 (`RegisterProjectTools` at 0%)

---

### UT-002 — `TestHandleCreateProject_InvalidArguments`
- **Function under test:** `handleCreateProject` (via `create_project` tool)
- **Given:** `mockRepo` not needed (unmarshal fails before any repo call)
- **When:** `tool(ctx, json.RawMessage("not-valid-json"))`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `err.Error()` contains `"invalid arguments"`
- **Architecture cite:** architecture.md §4.3 `_InvalidArguments` branch

---

### UT-003 — `TestHandleCreateProject_EmptyName`
- **Function under test:** `handleCreateProject`
- **Given:** valid JSON but `name` is empty string or whitespace-only
- **When:** `tool(ctx, json.RawMessage(`{"name": ""}`))` (or whitespace variant)
- **Then:**
  - `result` is `nil`
  - `err.Error()` contains `"name is required and cannot be empty"`
- **Architecture cite:** architecture.md §4.3 `_EmptyName` branch

---

### UT-004 — `TestHandleCreateProject_RepoError`
- **Function under test:** `handleCreateProject`
- **Given:**
  ```go
  mockErr := errors.New("db down")
  mockRepo.CreateProjectFunc = func(ctx context.Context, p *domain.Project) (*domain.Project, error) {
      return nil, mockErr
  }
  ```
- **When:** `tool(ctx, json.RawMessage(`{"name": "My Project"}`))` 
- **Then:**
  - `result` is `nil`
  - `errors.Is(returnedErr, mockErr)` is `true` (passthrough — no wrap)
- **Architecture cite:** architecture.md §4.3 `_GenericError` passthrough; architecture.md §12.4

---

### UT-005 — `TestHandleGetProject_InvalidArguments`
- **Function under test:** `handleGetProject`
- **Given:** malformed JSON
- **When:** `tool(ctx, json.RawMessage("not-valid-json"))`
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-006 — `TestHandleGetProject_EmptyID`
- **Function under test:** `handleGetProject`
- **Given:** valid JSON but `id` is empty
- **When:** `tool(ctx, json.RawMessage(`{"id": ""}`))` 
- **Then:** `err.Error()` contains `"id is required"`

---

### UT-007 — `TestHandleGetProject_NotFound`
- **Function under test:** `handleGetProject`
- **Given:**
  ```go
  mockRepo.GetProjectFunc = func(ctx context.Context, id string) (*domain.Project, error) {
      return nil, repo.ErrNotFound
  }
  ```
- **When:** `tool(ctx, json.RawMessage(`{"id": "proj-1"}`))` 
- **Then:**
  - `result` is `nil`
  - `err.Error()` contains `"project not found"`
  - `errors.Is(err, repo.ErrNotFound)` is `false` (fresh error — sentinel NOT preserved)
- **Architecture cite:** architecture.md §4.3 `_NotFound` branch; architecture.md §12.4

---

### UT-008 — `TestHandleGetProject_GenericError`
- **Function under test:** `handleGetProject`
- **Given:**
  ```go
  mockErr := errors.New("db down")
  mockRepo.GetProjectFunc = func(_ context.Context, _ string) (*domain.Project, error) {
      return nil, mockErr
  }
  ```
- **When:** `tool(ctx, json.RawMessage(`{"id": "proj-1"}`))` 
- **Then:** `errors.Is(returnedErr, mockErr)` is `true` (passthrough)

---

### UT-009 — `TestHandleUpdateProject_InvalidArguments`
- **Function under test:** `handleUpdateProject`
- **When:** `tool(ctx, json.RawMessage("not-valid-json"))`
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-010 — `TestHandleUpdateProject_EmptyID`
- **Function under test:** `handleUpdateProject`
- **When:** `tool(ctx, json.RawMessage(`{"id": ""}`))` 
- **Then:** `err.Error()` contains `"id is required"`

---

### UT-011 — `TestHandleUpdateProject_NotFoundOnInitialGet`
- **Function under test:** `handleUpdateProject`
- **Given:**
  ```go
  mockRepo.GetProjectFunc = func(_ context.Context, _ string) (*domain.Project, error) {
      return nil, repo.ErrNotFound
  }
  ```
- **When:** `tool(ctx, json.RawMessage(`{"id": "proj-1"}`))` 
- **Then:**
  - `err.Error()` contains `"project not found"`
  - `errors.Is(err, repo.ErrNotFound)` is `false`

---

### UT-012 — `TestHandleUpdateProject_GenericErrorOnInitialGet`
- **Function under test:** `handleUpdateProject`
- **Given:** `GetProjectFunc` returns generic error
- **Then:** `errors.Is(returnedErr, mockErr)` is `true` (passthrough)

---

### UT-013 — `TestHandleUpdateProject_EmptyNameWhenProvided`
- **Function under test:** `handleUpdateProject`
- **Given:** `GetProjectFunc` returns a valid project; request body has `name` field set to empty/whitespace
- **When:** `tool(ctx, json.RawMessage(`{"id": "proj-1", "name": " "}`))` 
- **Then:** `err.Error()` contains `"name cannot be empty if provided"`

---

### UT-014 — `TestHandleUpdateProject_RepoUpdateError`
- **Function under test:** `handleUpdateProject`
- **Given:**
  - `GetProjectFunc` returns a valid project (happy)
  - `UpdateProjectFunc` returns `errors.New("db down")`
- **When:** `tool(ctx, json.RawMessage(`{"id": "proj-1", "name": "New Name"}`))` 
- **Then:** `errors.Is(returnedErr, mockErr)` is `true` (passthrough)

---

### UT-015 — `TestHandleDeleteProject_InvalidArguments`
- **Function under test:** `handleDeleteProject`
- **When:** malformed JSON
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-016 — `TestHandleDeleteProject_EmptyID`
- **Function under test:** `handleDeleteProject`
- **When:** `tool(ctx, json.RawMessage(`{"id": ""}`))` 
- **Then:** `err.Error()` contains `"id is required"`

---

### UT-017 — `TestHandleDeleteProject_RepoError`
- **Function under test:** `handleDeleteProject`
- **Given:** `DeleteProjectFunc` returns `errors.New("db down")`
- **Then:** `errors.Is(returnedErr, mockErr)` is `true` (passthrough)

---

### UT-018 — `TestHandleListProjects_RepoError`
- **Function under test:** `handleListProjects`
- **Given:** `ListProjectsFunc` returns `errors.New("db down")`
- **When:** `tool(ctx, json.RawMessage(`{}`))`
- **Then:** `errors.Is(returnedErr, mockErr)` is `true` (passthrough)

## Integration tests

### IT-001 — per-file coverage ≥95%
- **Service:** `services/agent-board`
- **Command:**
  ```
  cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out \
      -run "TestHandle(Create|Get|Update|Delete|List)Project|TestRegisterProjectTools"
  go tool cover -func=/tmp/handler.out | grep project_tools.go
  ```
- **Expect:** `project_tools.go` total statement coverage ≥95%.

### IT-002 — full suite regression
- **Command:** `cd services/agent-board && go test ./... && golangci-lint run ./...`
- **Expect:** all pre-existing tests pass; no new lint issues.

## Coverage exemptions

None anticipated. 18 tests on a 6-function file with no transactional branches should reach >95%. If any line is genuinely unreachable, document under OQ-4.
