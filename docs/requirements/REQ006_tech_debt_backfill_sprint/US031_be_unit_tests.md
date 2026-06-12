# US031 — Backend unit & integration test specification
# `user_story_tools.go` error-mapping tests

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live in `services/agent-board/internal/handler/user_story_tools_test.go`. Production code in `user_story_tools.go` is **byte-for-byte unchanged**.

**Key semantic for `user_story_tools.go`:** repo errors are **passed through** (NOT wrapped with `fmt.Errorf`) on most paths. Use `errors.Is(returnedErr, mockErr)` for generic-error assertions, NOT substring match on a wrap prefix. `_NotFound` cases return a fresh error `errors.New("user story not found")` — assert `err.Error()` contains `"user story not found"` AND `errors.Is(err, repo.ErrNotFound)` is `false`. Status-transition errors use the format `"invalid transition from <from> to <to>"`. **This passthrough behaviour is the key difference from `task_tools.go` — read the source before writing assertions.**

**Harness shape (architecture.md §4.3):** real `mcp.NewToolRegistry()`, hand-written `MockUserStoryRepo` (or `testify/mock`), `handler.RegisterUserStoryTools(registry, mockRepo)`, retrieve via `registry.GetTool(name)`, invoke with `json.RawMessage`.

## Coverage matrix

| AC scenario | Layer | Test ID | Function under test |
|---|---|---|---|
| `RegisterUserStoryTools` registers all 5 tools | unit | UT-001 | `RegisterUserStoryTools` |
| `create_user_story` invalid JSON | unit | UT-002 | `handleCreateUserStory` |
| `create_user_story` missing projectId or title | unit | UT-003 | `handleCreateUserStory` |
| `create_user_story` default status when omitted | unit | UT-004 | `handleCreateUserStory` |
| `create_user_story` invalid initial status | unit | UT-005 | `handleCreateUserStory` |
| `create_user_story` repo error (passthrough) | unit | UT-006 | `handleCreateUserStory` |
| `get_user_story` invalid JSON | unit | UT-007 | `handleGetUserStory` |
| `get_user_story` missing id | unit | UT-008 | `handleGetUserStory` |
| `get_user_story` repo ErrNotFound | unit | UT-009 | `handleGetUserStory` |
| `get_user_story` repo generic error (passthrough) | unit | UT-010 | `handleGetUserStory` |
| `update_user_story` invalid JSON | unit | UT-011 | `handleUpdateUserStory` |
| `update_user_story` missing id | unit | UT-012 | `handleUpdateUserStory` |
| `update_user_story` initial Get returns ErrNotFound | unit | UT-013 | `handleUpdateUserStory` |
| `update_user_story` initial Get generic error (passthrough) | unit | UT-014 | `handleUpdateUserStory` |
| `update_user_story` invalid status transition | unit | UT-015 | `handleUpdateUserStory` |
| `update_user_story` UpdateUserStoryStatus error (passthrough) | unit | UT-016 | `handleUpdateUserStory` |
| `update_user_story` post-status field update error (passthrough) | unit | UT-017 | `handleUpdateUserStory` |
| `update_user_story` status change happy path no extra fields | unit | UT-018 | `handleUpdateUserStory` |
| `update_user_story` status change happy path with extra fields | unit | UT-019 | `handleUpdateUserStory` |
| `update_user_story` no status change + UpdateUserStory error | unit | UT-020 | `handleUpdateUserStory` |
| `delete_user_story` invalid JSON | unit | UT-021 | `handleDeleteUserStory` |
| `delete_user_story` missing id | unit | UT-022 | `handleDeleteUserStory` |
| `delete_user_story` repo error (passthrough) | unit | UT-023 | `handleDeleteUserStory` |
| `list_user_stories` invalid JSON | unit | UT-024 | `handleListUserStories` |
| `list_user_stories` missing projectId | unit | UT-025 | `handleListUserStories` |
| `list_user_stories` repo error (passthrough) | unit | UT-026 | `handleListUserStories` |
| `list_user_stories` empty slice returns `{"userStories":[]}` | unit | UT-027 | `handleListUserStories` |
| per-file coverage ≥95% | integration | IT-001 | `user_story_tools.go` all functions |
| full suite still passes | integration | IT-002 | `go test ./...` |

## Unit tests

### UT-001 — `TestRegisterUserStoryTools_RegistersAllFiveTools`
- **Function under test:** `RegisterUserStoryTools`
- **Given:**
  ```go
  registry := mcp.NewToolRegistry()
  mockRepo := &MockUserStoryRepo{}
  handler.RegisterUserStoryTools(registry, mockRepo)
  ```
- **Then:**
  - `registry.GetTool("create_user_story")` returns `(handler, true)`
  - `registry.GetTool("get_user_story")` returns `(handler, true)`
  - `registry.GetTool("update_user_story")` returns `(handler, true)`
  - `registry.GetTool("delete_user_story")` returns `(handler, true)`
  - `registry.GetTool("list_user_stories")` returns `(handler, true)`
  - Unknown name returns `(nil, false)`
- **Architecture cite:** architecture.md §4.3 `_RegistersAll*Tools`; tech_debt.md line 58 (`RegisterUserStoryTools` at 63.5%)

---

### UT-002 — `TestCreateUserStoryTool_InvalidArguments`
- **When:** `tool(ctx, json.RawMessage("not-valid-json"))`
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-003 — `TestCreateUserStoryTool_MissingProjectIDOrTitle`
- **Given:** valid JSON but `projectId` is empty OR `title` is empty
- **Then:** `err.Error()` contains `"missing required fields"`
- **Architecture cite:** US031 AC exact validation string

---

### UT-004 — `TestCreateUserStoryTool_DefaultStatusWhenOmitted`
- **Given:**
  ```go
  var capturedStory *domain.UserStory
  mockRepo.CreateUserStoryFunc = func(_ context.Context, s *domain.UserStory) (*domain.UserStory, error) {
      capturedStory = s
      return s, nil
  }
  ```
- **When:** body has `projectId` and `title` but no `status`
- **Then:**
  - `capturedStory.Status` equals `domain.UserStoryStatusDraft` (or `"draft"` — confirm from domain)
  - `err` is `nil`

---

### UT-005 — `TestCreateUserStoryTool_InvalidInitialStatus`
- **Given:** body has `status = "in_signoff"` (non-draft initial status)
- **Then:** `err.Error()` contains `"invalid initial status:"`

---

### UT-006 — `TestCreateUserStoryTool_RepoError`
- **Given:** `CreateUserStoryFunc` returns `mockErr := errors.New("db down")`
- **Then:** `errors.Is(returnedErr, mockErr)` is `true` (passthrough — NO wrap)

---

### UT-007 — `TestGetUserStoryTool_InvalidArguments`
- **When:** malformed JSON
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-008 — `TestGetUserStoryTool_MissingID`
- **Given:** valid JSON but `id` is empty
- **Then:** `err.Error()` contains `"missing id"`
- **Architecture cite:** US031 AC exact validation string

---

### UT-009 — `TestGetUserStoryTool_NotFound`
- **Given:** `GetUserStoryFunc` returns `repo.ErrNotFound`
- **Then:**
  - `err.Error()` contains `"user story not found"`
  - `errors.Is(err, repo.ErrNotFound)` is `false`

---

### UT-010 — `TestGetUserStoryTool_GenericError`
- **Given:** `GetUserStoryFunc` returns `mockErr := errors.New("db down")`
- **Then:** `errors.Is(returnedErr, mockErr)` is `true` (passthrough)

---

### UT-011 — `TestUpdateUserStoryTool_InvalidArguments`
- **When:** malformed JSON
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-012 — `TestUpdateUserStoryTool_MissingID`
- **Given:** valid JSON but `id` is empty
- **Then:** `err.Error()` contains `"missing id"`

---

### UT-013 — `TestUpdateUserStoryTool_NotFoundOnInitialGet`
- **Given:** `GetUserStoryFunc` returns `repo.ErrNotFound`
- **Then:**
  - `err.Error()` contains `"user story not found"`
  - `errors.Is(err, repo.ErrNotFound)` is `false`

---

### UT-014 — `TestUpdateUserStoryTool_GenericErrorOnInitialGet`
- **Given:** `GetUserStoryFunc` returns `mockErr := errors.New("db down")`
- **Then:** `errors.Is(returnedErr, mockErr)` is `true` (passthrough)

---

### UT-015 — `TestUpdateUserStoryTool_InvalidStatusTransition`
- **Given:**
  - `GetUserStoryFunc` returns existing story with `Status = "draft"`
  - Body requests `Status = "done"` (invalid direct transition — confirm from `domain.UserStory.IsValidTransition`)
- **Then:** `err.Error()` contains `"invalid transition from draft to done"` (or the tested pair)

---

### UT-016 — `TestUpdateUserStoryTool_StatusChange_UpdateUserStoryStatusError`
- **Given:**
  - `GetUserStoryFunc` returns story with valid from-status
  - Body requests valid status transition
  - `UpdateUserStoryStatusFunc` returns `mockErr := errors.New("db down")`
- **Then:** `errors.Is(returnedErr, mockErr)` is `true` (passthrough)

---

### UT-017 — `TestUpdateUserStoryTool_StatusChange_PostStatusFieldUpdateError`
- **Given:**
  - `GetUserStoryFunc` returns story
  - Body requests valid status transition AND a `title` or `description` update
  - `UpdateUserStoryStatusFunc` returns a happy updated story
  - `UpdateUserStoryFunc` (the post-status field save) returns `mockErr := errors.New("db down")`
- **Then:** `errors.Is(returnedErr, mockErr)` is `true` (passthrough)

---

### UT-018 — `TestUpdateUserStoryTool_StatusChange_HappyPath_NoExtraFields`
- **Given:**
  - `GetUserStoryFunc` returns story with valid from-status
  - Body requests valid status transition, no extra field changes
  - `UpdateUserStoryStatusFunc` returns an updated story
- **Then:**
  - `err` is `nil`
  - `result` is a `UserStoryResponse` with the new status

---

### UT-019 — `TestUpdateUserStoryTool_StatusChange_HappyPath_WithExtraFields`
- **Given:**
  - `GetUserStoryFunc` returns story with valid from-status
  - Body requests valid status transition AND `title` + `description` changes
  - `UpdateUserStoryStatusFunc` returns happy; `UpdateUserStoryFunc` returns the final updated story
- **Then:**
  - `err` is `nil`
  - `result` is a `UserStoryResponse` with both new status and updated title/description

---

### UT-020 — `TestUpdateUserStoryTool_NoStatusChange_RepoUpdateError`
- **Given:**
  - `GetUserStoryFunc` returns story
  - Body has no `status` field (or same status), has field updates
  - `UpdateUserStoryFunc` returns `mockErr := errors.New("db down")`
- **Then:** `errors.Is(returnedErr, mockErr)` is `true` (passthrough)

---

### UT-021 — `TestDeleteUserStoryTool_InvalidArguments`
- **When:** malformed JSON
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-022 — `TestDeleteUserStoryTool_MissingID`
- **Given:** valid JSON but `id` is empty
- **Then:** `err.Error()` contains `"missing id"`

---

### UT-023 — `TestDeleteUserStoryTool_RepoError`
- **Given:** `DeleteUserStoryFunc` returns `mockErr := errors.New("db down")`
- **Then:** `errors.Is(returnedErr, mockErr)` is `true` (passthrough)

---

### UT-024 — `TestListUserStoriesTool_InvalidArguments`
- **When:** malformed JSON
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-025 — `TestListUserStoriesTool_MissingProjectID`
- **Given:** valid JSON but `projectId` is empty
- **Then:** `err.Error()` contains `"missing projectId"` (confirm exact wording from `user_story_tools.go`)

---

### UT-026 — `TestListUserStoriesTool_RepoError`
- **Given:** `ListUserStoriesFunc` returns `mockErr := errors.New("db down")`
- **Then:** `errors.Is(returnedErr, mockErr)` is `true` (passthrough)

---

### UT-027 — `TestListUserStoriesTool_EmptySliceReturnsEmptyArray`
- **Given:** `ListUserStoriesFunc` returns `nil` or empty slice
- **Then:**
  - `err` is `nil`
  - `result` is a map with key `"userStories"` whose value is a non-nil slice of length 0
  - (NOT nil under the `"userStories"` key — verify `make([]UserStoryResponse, 0, ...)` semantics)
- **Architecture cite:** US031 AC `_EmptySliceReturnsEmptyArray`

## Integration tests

### IT-001 — per-file coverage ≥95%
- **Command:**
  ```
  cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out \
      -run "TestRegisterUserStoryTools|Test(Create|Get|Update|Delete|List)UserStor(y|ies)Tool"
  go tool cover -func=/tmp/handler.out | grep user_story_tools.go
  ```
- **Expect:** `user_story_tools.go` total statement coverage ≥95%.

### IT-002 — full suite regression
- **Command:** `cd services/agent-board && go test ./... && golangci-lint run ./...`
- **Expect:** all pre-existing tests pass; no new lint issues.

## Coverage exemptions

None anticipated. If any line is genuinely unreachable, document under OQ-4.
