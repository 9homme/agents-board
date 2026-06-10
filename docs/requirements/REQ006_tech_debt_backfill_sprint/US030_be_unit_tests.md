# US030 — Backend unit & integration test specification
# `task_tools.go` error-mapping tests

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live in `services/agent-board/internal/handler/task_tools_test.go`. Production code in `task_tools.go` is **byte-for-byte unchanged**.

**Key semantic for `task_tools.go`:** repo errors are **wrapped** with `fmt.Errorf("failed to <op> task: %w", err)`. Assert via `assert.Contains(t, err.Error(), "failed to <op> task:")`. `_NotFound` cases return a fresh error — assert `err.Error()` contains `"task not found"` AND `errors.Is(err, repo.ErrNotFound)` is `false`. Status-transition errors use the format `"invalid transition from <from> to <to>"`.

**Harness shape (architecture.md §4.3):** real `mcp.NewToolRegistry()`, hand-written `MockTaskRepo` (or `testify/mock`), `handler.RegisterTaskTools(registry, mockRepo)`, retrieve via `registry.GetTool(name)`, invoke with `json.RawMessage`.

## Coverage matrix

| AC scenario | Layer | Test ID | Function under test |
|---|---|---|---|
| `RegisterTaskTools` registers all 5 tools | unit | UT-001 | `RegisterTaskTools` |
| `create_task` invalid JSON | unit | UT-002 | `handleCreateTask` |
| `create_task` missing userStoryId or title | unit | UT-003 | `handleCreateTask` |
| `create_task` default status when omitted | unit | UT-004 | `handleCreateTask` |
| `create_task` invalid initial status | unit | UT-005 | `handleCreateTask` |
| `create_task` repo error (wrapped) | unit | UT-006 | `handleCreateTask` |
| `get_task` invalid JSON | unit | UT-007 | `handleGetTask` |
| `get_task` empty id | unit | UT-008 | `handleGetTask` |
| `get_task` repo ErrNotFound | unit | UT-009 | `handleGetTask` |
| `get_task` repo generic error (wrapped) | unit | UT-010 | `handleGetTask` |
| `update_task` invalid JSON | unit | UT-011 | `handleUpdateTask` |
| `update_task` empty id | unit | UT-012 | `handleUpdateTask` |
| `update_task` initial Get returns ErrNotFound | unit | UT-013 | `handleUpdateTask` |
| `update_task` initial Get generic error (wrapped) | unit | UT-014 | `handleUpdateTask` |
| `update_task` invalid status transition | unit | UT-015 | `handleUpdateTask` |
| `update_task` status change + field update error | unit | UT-016 | `handleUpdateTask` |
| `update_task` status change + UpdateTaskStatus error | unit | UT-017 | `handleUpdateTask` |
| `update_task` no status change + UpdateTask error | unit | UT-018 | `handleUpdateTask` |
| `update_task` status change happy path | unit | UT-019 | `handleUpdateTask` |
| `delete_task` invalid JSON | unit | UT-020 | `handleDeleteTask` |
| `delete_task` empty id | unit | UT-021 | `handleDeleteTask` |
| `delete_task` repo error (wrapped) | unit | UT-022 | `handleDeleteTask` |
| `list_tasks` invalid JSON | unit | UT-023 | `handleListTasks` |
| `list_tasks` missing userStoryId | unit | UT-024 | `handleListTasks` |
| `list_tasks` repo error (wrapped) | unit | UT-025 | `handleListTasks` |
| per-file coverage ≥95% | integration | IT-001 | `task_tools.go` all functions |
| full suite still passes | integration | IT-002 | `go test ./...` |

## Unit tests

### UT-001 — `TestRegisterTaskTools_RegistersAllFiveTools`
- **Function under test:** `RegisterTaskTools`
- **Given:**
  ```go
  registry := mcp.NewToolRegistry()
  mockRepo := &MockTaskRepo{}
  handler.RegisterTaskTools(registry, mockRepo)
  ```
- **Then:**
  - `registry.GetTool("create_task")` returns `(handler, true)`
  - `registry.GetTool("get_task")` returns `(handler, true)`
  - `registry.GetTool("update_task")` returns `(handler, true)`
  - `registry.GetTool("delete_task")` returns `(handler, true)`
  - `registry.GetTool("list_tasks")` returns `(handler, true)`
  - Unknown name returns `(nil, false)`
- **Architecture cite:** architecture.md §4.3 `_RegistersAll*Tools`; tech_debt.md line 57 (`RegisterTaskTools` at 67.4%)

---

### UT-002 — `TestCreateTaskTool_InvalidArguments`
- **When:** `tool(ctx, json.RawMessage("not-valid-json"))`
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-003 — `TestCreateTaskTool_MissingUserStoryIDOrTitle`
- **Given:** valid JSON but `userStoryId` is empty OR `title` is empty
- **Then:** `err.Error()` contains `"userStoryId and title are required"`

---

### UT-004 — `TestCreateTaskTool_DefaultStatusWhenOmitted`
- **Given:**
  ```go
  var capturedTask *domain.Task
  mockRepo.CreateTaskFunc = func(_ context.Context, t *domain.Task) (*domain.Task, error) {
      capturedTask = t
      return t, nil
  }
  ```
- **When:** `tool(ctx, json.RawMessage(`{"userStoryId": "us-1", "title": "Do thing"}`))` (no `status` field)
- **Then:**
  - `capturedTask.Status` equals `domain.TaskStatusPending` (or the string `"pending"` — confirm exact value from domain)
  - `err` is `nil`
- **Architecture cite:** US030 AC `_DefaultStatusWhenOmitted`

---

### UT-005 — `TestCreateTaskTool_InvalidInitialStatus`
- **Given:** body has `status` set to `"in_progress"` (non-pending initial status)
- **When:** `tool(ctx, json.RawMessage(`{"userStoryId": "us-1", "title": "T", "status": "in_progress"}`))` 
- **Then:**
  - `err.Error()` contains `"invalid initial status:"`
- **Architecture cite:** US030 AC `_InvalidInitialStatus`; `domain.NewTask` validation

---

### UT-006 — `TestCreateTaskTool_RepoError`
- **Given:** `CreateTaskFunc` returns `errors.New("db down")`
- **Then:** `err.Error()` contains `"failed to create task:"`

---

### UT-007 — `TestGetTaskTool_InvalidArguments`
- **When:** malformed JSON
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-008 — `TestGetTaskTool_EmptyID`
- **When:** `tool(ctx, json.RawMessage(`{"id": ""}`))` 
- **Then:** `err.Error()` contains `"id is required"`

---

### UT-009 — `TestGetTaskTool_NotFound`
- **Given:** `GetTaskFunc` returns `repo.ErrNotFound`
- **Then:**
  - `err.Error()` contains `"task not found"`
  - `errors.Is(err, repo.ErrNotFound)` is `false`

---

### UT-010 — `TestGetTaskTool_GenericError`
- **Given:** `GetTaskFunc` returns `errors.New("db down")`
- **Then:** `err.Error()` contains `"failed to get task:"`

---

### UT-011 — `TestUpdateTaskTool_InvalidArguments`
- **When:** malformed JSON
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-012 — `TestUpdateTaskTool_EmptyID`
- **When:** `tool(ctx, json.RawMessage(`{"id": ""}`))` 
- **Then:** `err.Error()` contains `"id is required"`

---

### UT-013 — `TestUpdateTaskTool_NotFoundOnInitialGet`
- **Given:** `GetTaskFunc` returns `repo.ErrNotFound`
- **Then:**
  - `err.Error()` contains `"task not found"`
  - `errors.Is(err, repo.ErrNotFound)` is `false`

---

### UT-014 — `TestUpdateTaskTool_GenericErrorOnInitialGet`
- **Given:** `GetTaskFunc` returns `errors.New("db down")`
- **Then:** `err.Error()` contains `"failed to get task:"`

---

### UT-015 — `TestUpdateTaskTool_InvalidStatusTransition`
- **Given:**
  - `GetTaskFunc` returns existing task with `Status = "pending"`
  - Body requests `Status = "done"` (an invalid direct transition pending → done if the domain requires intermediate states — confirm from `domain.TaskStatusPending.IsValidTransition("done")`)
- **Then:**
  - `err.Error()` matches `"invalid transition from pending to done"` (or the actual from/to pair for the tested transition — use the format `"invalid transition from <from> to <to>"`)
- **Architecture cite:** architecture.md §4.3 `_InvalidStatusTransition` branch

---

### UT-016 — `TestUpdateTaskTool_StatusChange_FieldUpdateError`
- **Given:**
  - `GetTaskFunc` returns existing task with valid from-status
  - Body requests valid status transition AND a `title` update
  - `UpdateTaskFunc` (the field-update call) returns `errors.New("db down")`
- **Then:** `err.Error()` contains `"failed to update task fields:"`

---

### UT-017 — `TestUpdateTaskTool_StatusChange_UpdateTaskStatusError`
- **Given:**
  - `GetTaskFunc` returns existing task
  - Body requests valid status transition, no other field changes
  - `UpdateTaskStatusFunc` returns `errors.New("db down")`
- **Then:** `err.Error()` contains `"failed to update task status:"`

---

### UT-018 — `TestUpdateTaskTool_NoStatusChange_RepoUpdateError`
- **Given:**
  - `GetTaskFunc` returns existing task
  - Body requests no status change (or omits `status`), but has field updates
  - `UpdateTaskFunc` returns `errors.New("db down")`
- **Then:** `err.Error()` contains `"failed to update task:"`

---

### UT-019 — `TestUpdateTaskTool_StatusChange_HappyPath`
- **Given:**
  - `GetTaskFunc` returns existing task with `Status = "pending"`
  - Body requests valid status transition (`status = "in_progress"`)
  - `UpdateTaskStatusFunc` returns an updated task with `Status = "in_progress"`
- **Then:**
  - `err` is `nil`
  - `result` is a `TaskResponse` (or equivalent struct) with `Status = "in_progress"`
- **Architecture cite:** US030 AC `_StatusChange_HappyPath`

---

### UT-020 — `TestDeleteTaskTool_InvalidArguments`
- **When:** malformed JSON
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-021 — `TestDeleteTaskTool_EmptyID`
- **When:** `tool(ctx, json.RawMessage(`{"id": ""}`))` 
- **Then:** `err.Error()` contains `"id is required"`

---

### UT-022 — `TestDeleteTaskTool_RepoError`
- **Given:** `DeleteTaskFunc` returns `errors.New("db down")`
- **Then:** `err.Error()` contains `"failed to delete task:"`

---

### UT-023 — `TestListTasksTool_InvalidArguments`
- **When:** malformed JSON
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-024 — `TestListTasksTool_MissingUserStoryID`
- **Given:** valid JSON but `userStoryId` is empty
- **Then:** `err.Error()` indicates `userStoryId` is required (confirm exact wording from `task_tools.go`)

---

### UT-025 — `TestListTasksTool_RepoError`
- **Given:** `ListTasksFunc` returns `errors.New("db down")`
- **Then:** `err.Error()` contains `"failed to list tasks:"`

## Integration tests

### IT-001 — per-file coverage ≥95%
- **Command:**
  ```
  cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out \
      -run "TestRegisterTaskTools|Test(Create|Get|Update|Delete|List)Task(s?)Tool"
  go tool cover -func=/tmp/handler.out | grep task_tools.go
  ```
- **Expect:** `task_tools.go` total statement coverage ≥95%.

### IT-002 — full suite regression
- **Command:** `cd services/agent-board && go test ./... && golangci-lint run ./...`
- **Expect:** all pre-existing tests pass; no new lint issues.

## Coverage exemptions

None anticipated. If any line is genuinely unreachable, document under OQ-4.
