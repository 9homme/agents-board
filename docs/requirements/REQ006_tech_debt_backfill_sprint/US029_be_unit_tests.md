# US029 — Backend unit & integration test specification
# `document_tools.go` error-mapping tests

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live in `services/agent-board/internal/handler/document_tools_test.go`. The existing `MockDocumentRepo` at the top of that file should be reused. Production code in `document_tools.go` is **byte-for-byte unchanged**.

**Key semantic for `document_tools.go`:** repo errors are **wrapped** with `fmt.Errorf("failed to <op> document: %w", err)`. Assert via `assert.Contains(t, err.Error(), "failed to <op> document:")`. `_NotFound` cases return a fresh error — assert `err.Error()` contains `"document not found"` AND `errors.Is(err, repo.ErrNotFound)` is `false`.

**Harness shape (architecture.md §4.3):** real `mcp.NewToolRegistry()`, reuse existing `MockDocumentRepo`, call `handler.RegisterDocumentTools(registry, mockRepo)`, retrieve via `registry.GetTool("tool-name")`, invoke with `json.RawMessage`.

## Coverage matrix

| AC scenario | Layer | Test ID | Function under test |
|---|---|---|---|
| `RegisterDocumentTools` registers all 5 tools | unit | UT-001 | `RegisterDocumentTools` |
| `create_document` invalid JSON | unit | UT-002 | `handleCreateDocument` |
| `create_document` missing projectId or title | unit | UT-003 | `handleCreateDocument` |
| `create_document` repo error (wrapped) | unit | UT-004 | `handleCreateDocument` |
| `get_document` invalid JSON | unit | UT-005 | `handleGetDocument` |
| `get_document` empty id | unit | UT-006 | `handleGetDocument` |
| `get_document` repo ErrNotFound | unit | UT-007 | `handleGetDocument` |
| `get_document` repo generic error (wrapped) | unit | UT-008 | `handleGetDocument` |
| `update_document` invalid JSON | unit | UT-009 | `handleUpdateDocument` |
| `update_document` empty id | unit | UT-010 | `handleUpdateDocument` |
| `update_document` initial Get returns ErrNotFound | unit | UT-011 | `handleUpdateDocument` |
| `update_document` initial Get generic error (wrapped) | unit | UT-012 | `handleUpdateDocument` |
| `update_document` UpdateDocument repo error (wrapped) | unit | UT-013 | `handleUpdateDocument` |
| `delete_document` invalid JSON | unit | UT-014 | `handleDeleteDocument` |
| `delete_document` empty id | unit | UT-015 | `handleDeleteDocument` |
| `delete_document` repo error (wrapped) | unit | UT-016 | `handleDeleteDocument` |
| `list_documents` invalid JSON | unit | UT-017 | `handleListDocuments` |
| `list_documents` missing projectId | unit | UT-018 | `handleListDocuments` |
| `list_documents` repo error (wrapped) | unit | UT-019 | `handleListDocuments` |
| `list_documents` empty slice returns `{"documents":[]}` | unit | UT-020 | `handleListDocuments` |
| per-file coverage ≥95% | integration | IT-001 | `document_tools.go` all functions |
| full suite still passes | integration | IT-002 | `go test ./...` |

## Unit tests

### UT-001 — `TestRegisterDocumentTools_RegistersAllFiveTools`
- **Function under test:** `RegisterDocumentTools`
- **Given:**
  ```go
  registry := mcp.NewToolRegistry()
  mockRepo := &MockDocumentRepo{} // existing mock at top of document_tools_test.go
  handler.RegisterDocumentTools(registry, mockRepo)
  ```
- **Then:**
  - `registry.GetTool("create_document")` returns `(handler, true)`
  - `registry.GetTool("get_document")` returns `(handler, true)`
  - `registry.GetTool("update_document")` returns `(handler, true)`
  - `registry.GetTool("delete_document")` returns `(handler, true)`
  - `registry.GetTool("list_document")` (or `"list_documents"` — confirm exact name from `document_tools.go`) returns `(handler, true)`
  - `registry.GetTool("nonexistent_tool")` returns `(nil, false)`
- **Architecture cite:** architecture.md §4.3 `_RegistersAll*Tools` branch; tech_debt.md line 56 (`RegisterDocumentTools` at 69.2%)

---

### UT-002 — `TestCreateDocumentTool_InvalidArguments`
- **When:** `tool(ctx, json.RawMessage("not-valid-json"))`
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-003 — `TestCreateDocumentTool_MissingProjectIDOrTitle`
- **Given:** valid JSON but `projectId` is empty OR `title` is empty
- **Then:** `err.Error()` contains `"projectId and title are required"`

---

### UT-004 — `TestCreateDocumentTool_RepoError`
- **Given:**
  ```go
  mockErr := errors.New("db down")
  mockRepo.CreateDocumentFunc = func(_ context.Context, _ *domain.Document) (*domain.Document, error) {
      return nil, mockErr
  }
  ```
- **Then:**
  - `result` is `nil`
  - `err.Error()` contains `"failed to create document:"`
- **Architecture cite:** architecture.md §4.3 `_GenericError` wrapped; document_tools.go wrap line

---

### UT-005 — `TestGetDocumentTool_InvalidArguments`
- **When:** `tool(ctx, json.RawMessage("not-valid-json"))`
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-006 — `TestGetDocumentTool_EmptyID`
- **When:** `tool(ctx, json.RawMessage(`{"id": ""}`))` 
- **Then:** `err.Error()` contains `"id is required"`

---

### UT-007 — `TestGetDocumentTool_NotFound`
- **Given:** `GetDocumentFunc` returns `repo.ErrNotFound`
- **Then:**
  - `err.Error()` contains `"document not found"`
  - `errors.Is(err, repo.ErrNotFound)` is `false`

---

### UT-008 — `TestGetDocumentTool_GenericError`
- **Given:** `GetDocumentFunc` returns `errors.New("db down")`
- **Then:** `err.Error()` contains `"failed to get document:"`

---

### UT-009 — `TestUpdateDocumentTool_InvalidArguments`
- **When:** malformed JSON
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-010 — `TestUpdateDocumentTool_EmptyID`
- **When:** `tool(ctx, json.RawMessage(`{"id": ""}`))` 
- **Then:** `err.Error()` contains `"id is required"`

---

### UT-011 — `TestUpdateDocumentTool_NotFoundOnInitialGet`
- **Given:** `GetDocumentFunc` returns `repo.ErrNotFound`
- **Then:**
  - `err.Error()` contains `"document not found"`
  - `errors.Is(err, repo.ErrNotFound)` is `false`

---

### UT-012 — `TestUpdateDocumentTool_GenericErrorOnInitialGet`
- **Given:** `GetDocumentFunc` returns `errors.New("db down")`
- **Then:** `err.Error()` contains `"failed to get document:"`

---

### UT-013 — `TestUpdateDocumentTool_UpdateRepoError`
- **Given:**
  - `GetDocumentFunc` returns a valid document (happy)
  - `UpdateDocumentFunc` returns `errors.New("db down")`
- **Then:** `err.Error()` contains `"failed to update document:"`

---

### UT-014 — `TestDeleteDocumentTool_InvalidArguments`
- **When:** malformed JSON
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-015 — `TestDeleteDocumentTool_EmptyID`
- **When:** `tool(ctx, json.RawMessage(`{"id": ""}`))` 
- **Then:** `err.Error()` contains `"id is required"`

---

### UT-016 — `TestDeleteDocumentTool_RepoError`
- **Given:** `DeleteDocumentFunc` returns `errors.New("db down")`
- **Then:** `err.Error()` contains `"failed to delete document:"`

---

### UT-017 — `TestListDocumentsTool_InvalidArguments`
- **When:** malformed JSON
- **Then:** `err.Error()` contains `"invalid arguments"`

---

### UT-018 — `TestListDocumentsTool_MissingProjectID`
- **Given:** valid JSON but `projectId` is empty
- **Then:** `err.Error()` contains `"projectId"` and indicates it is required (confirm exact wording from `document_tools.go`)

---

### UT-019 — `TestListDocumentsTool_RepoError`
- **Given:** `ListDocumentsFunc` returns `errors.New("db down")`
- **Then:** `err.Error()` contains `"failed to list documents:"`

---

### UT-020 — `TestListDocumentsTool_EmptySliceReturnsEmptyDocumentsArray`
- **Given:** `ListDocumentsFunc` returns `nil` (or empty slice `[]*domain.Document{}`)
- **When:** `tool(ctx, json.RawMessage(`{"projectId": "proj-1"}`))` 
- **Then:**
  - `err` is `nil`
  - `result` is a `map[string]interface{}` with key `"documents"` whose value is a non-nil slice of length 0
  - (NOT a nil value under the `"documents"` key)
- **Architecture cite:** US029 AC `_EmptySliceReturnsEmptyDocumentsArray` scenario

## Integration tests

### IT-001 — per-file coverage ≥95%
- **Command:**
  ```
  cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out \
      -run "TestRegisterDocumentTools|Test(Create|Get|Update|Delete|List)Document(s?)Tool"
  go tool cover -func=/tmp/handler.out | grep document_tools.go
  ```
- **Expect:** `document_tools.go` total statement coverage ≥95%.

### IT-002 — full suite regression
- **Command:** `cd services/agent-board && go test ./... && golangci-lint run ./...`
- **Expect:** all pre-existing tests pass; no new lint issues.

## Coverage exemptions

None anticipated. If any line is genuinely unreachable, document under OQ-4.
