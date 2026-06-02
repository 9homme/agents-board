# US005 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). This story adds 16 test functions across two existing test files. No production code is touched. All tests use `github.com/DATA-DOG/go-sqlmock` (the pattern already used in the existing `*_repo_test.go` files). Write the failing test skeletons first, confirm the right failure mode (nil error returned where non-nil is expected), then write nothing — the existing production code already has the error-wrapping branches you are testing.

Architecture reference: §5 (test backfill matrix — 16 function names, mock shapes, assertion requirements), §5.3 (per-test assertion list), §5.4 (coverage target ≥95%).

## Coverage matrix

| AC scenario | Layer | Test ID | Service / package | Function under test |
|---|---|---|---|---|
| CreateDocument generic DB error | unit | UT-US005-001 | services/agent-board / internal/repo | `DocumentRepo.CreateDocument` |
| GetDocument generic DB error | unit | UT-US005-002 | services/agent-board / internal/repo | `DocumentRepo.GetDocument` |
| UpdateDocument not found (ErrNoRows) | unit | UT-US005-003 | services/agent-board / internal/repo | `DocumentRepo.UpdateDocument` |
| UpdateDocument generic DB error | unit | UT-US005-004 | services/agent-board / internal/repo | `DocumentRepo.UpdateDocument` |
| DeleteDocument generic DB error | unit | UT-US005-005 | services/agent-board / internal/repo | `DocumentRepo.DeleteDocument` |
| ListDocuments query error | unit | UT-US005-006 | services/agent-board / internal/repo | `DocumentRepo.ListDocuments` |
| ListDocuments scan error | unit | UT-US005-007 | services/agent-board / internal/repo | `DocumentRepo.ListDocuments` |
| ListDocuments rows error | unit | UT-US005-008 | services/agent-board / internal/repo | `DocumentRepo.ListDocuments` |
| CreateProject generic DB error | unit | UT-US005-009 | services/agent-board / internal/repo | `ProjectRepo.CreateProject` |
| GetProject generic DB error | unit | UT-US005-010 | services/agent-board / internal/repo | `ProjectRepo.GetProject` |
| UpdateProject not found (ErrNoRows) | unit | UT-US005-011 | services/agent-board / internal/repo | `ProjectRepo.UpdateProject` |
| UpdateProject generic DB error | unit | UT-US005-012 | services/agent-board / internal/repo | `ProjectRepo.UpdateProject` |
| DeleteProject generic DB error | unit | UT-US005-013 | services/agent-board / internal/repo | `ProjectRepo.DeleteProject` |
| ListProjects query error | unit | UT-US005-014 | services/agent-board / internal/repo | `ProjectRepo.ListProjects` |
| ListProjects scan error | unit | UT-US005-015 | services/agent-board / internal/repo | `ProjectRepo.ListProjects` |
| ListProjects rows error | unit | UT-US005-016 | services/agent-board / internal/repo | `ProjectRepo.ListProjects` |

## Unit tests

The 16 test function names below are VERBATIM as specified by architecture §5.1 and §5.2. The dev MUST use these exact names — the test report maps UT-US005-NNN IDs to these function names.

### Document repo — 8 new test functions in `document_repo_test.go`

### UT-US005-001 — `TestDocumentRepo_CreateDocument_GenericError`

- **Service:** `services/agent-board`
- **File:** `internal/repo/document_repo_test.go`
- **Test function name:** `TestDocumentRepo_CreateDocument_GenericError`
- **Given:**
  - `db, mock, _ := sqlmock.New()`
  - `mock.ExpectQuery(/* the INSERT query regex */).WillReturnError(errors.New("db down"))`
  - `repo := NewDocumentRepo(db)` (or however the repo is constructed in this package)
- **When:** call `repo.CreateDocument(context.Background(), <any valid input>)`
- **Then:**
  - Returned error is non-nil.
  - `errors.Is(err, repo.ErrNotFound)` is **false**.
  - Error message contains `"failed to create document"` (substring match).
  - `mock.ExpectationsWereMet()` returns nil.
  - Returned document pointer is nil.
- **Architecture cite:** architecture §5.1 D1; branch at `document_repo.go:43`.

### UT-US005-002 — `TestDocumentRepo_GetDocument_GenericError`

- **Service:** `services/agent-board`
- **File:** `internal/repo/document_repo_test.go`
- **Test function name:** `TestDocumentRepo_GetDocument_GenericError`
- **Given:**
  - `mock.ExpectQuery(/* SELECT query regex */).WillReturnError(errors.New("db down"))` — note: return non-`sql.ErrNoRows` to hit the generic wrap branch (line 65), NOT the `ErrNotFound` mapping.
- **When:** call `repo.GetDocument(context.Background(), "<any-id>")`
- **Then:**
  - Non-nil error.
  - `errors.Is(err, repo.ErrNotFound)` is **false**.
  - Error message contains `"failed to get document"` (substring).
  - `mock.ExpectationsWereMet()` nil.
  - Document pointer is nil.
- **Architecture cite:** architecture §5.1 D2; branch at `document_repo.go:65`.

### UT-US005-003 — `TestDocumentRepo_UpdateDocument_NotFound`

- **Service:** `services/agent-board`
- **File:** `internal/repo/document_repo_test.go`
- **Test function name:** `TestDocumentRepo_UpdateDocument_NotFound`
- **Given:**
  - `mock.ExpectQuery(/* UPDATE query regex */).WillReturnError(sql.ErrNoRows)` — this is the ErrNotFound mapping branch.
- **When:** call `repo.UpdateDocument(context.Background(), "<any-id>", <update input>)`
- **Then:**
  - Non-nil error.
  - `errors.Is(err, repo.ErrNotFound)` is **true**.
  - `mock.ExpectationsWereMet()` nil.
  - Document pointer is nil.
- **Architecture cite:** architecture §5.1 D3; branch at `document_repo.go:84-85`.

### UT-US005-004 — `TestDocumentRepo_UpdateDocument_GenericError`

- **Service:** `services/agent-board`
- **File:** `internal/repo/document_repo_test.go`
- **Test function name:** `TestDocumentRepo_UpdateDocument_GenericError`
- **Given:**
  - `mock.ExpectQuery(/* UPDATE query regex */).WillReturnError(errors.New("db down"))` — non-ErrNoRows to hit the generic wrap at line 87.
- **When:** call `repo.UpdateDocument(context.Background(), "<any-id>", <update input>)`
- **Then:**
  - Non-nil error.
  - `errors.Is(err, repo.ErrNotFound)` is **false**.
  - Error message contains `"failed to update document"` (substring).
  - `mock.ExpectationsWereMet()` nil.
  - Document pointer is nil.
- **Architecture cite:** architecture §5.1 D4; branch at `document_repo.go:87`.

### UT-US005-005 — `TestDocumentRepo_DeleteDocument_GenericError`

- **Service:** `services/agent-board`
- **File:** `internal/repo/document_repo_test.go`
- **Test function name:** `TestDocumentRepo_DeleteDocument_GenericError`
- **Given:**
  - `mock.ExpectExec(/* DELETE exec regex */).WillReturnError(errors.New("db down"))`
- **When:** call `repo.DeleteDocument(context.Background(), "<any-id>")`
- **Then:**
  - Non-nil error.
  - `errors.Is(err, repo.ErrNotFound)` is **false**.
  - Error message contains `"failed to delete document"` (substring).
  - `mock.ExpectationsWereMet()` nil.
- **Architecture cite:** architecture §5.1 D5; `ExecContext` error wrap at `document_repo.go:97`.

### UT-US005-006 — `TestDocumentRepo_ListDocuments_QueryError`

- **Service:** `services/agent-board`
- **File:** `internal/repo/document_repo_test.go`
- **Test function name:** `TestDocumentRepo_ListDocuments_QueryError`
- **Given:**
  - `mock.ExpectQuery(/* SELECT list query regex */).WillReturnError(errors.New("db down"))`
- **When:** call `repo.ListDocuments(context.Background(), "<project-id>")`
- **Then:**
  - Non-nil error.
  - `errors.Is(err, repo.ErrNotFound)` is **false**.
  - Error message contains `"failed to list documents"` (substring).
  - `mock.ExpectationsWereMet()` nil.
  - Returned slice is nil.
- **Architecture cite:** architecture §5.1 D6; `QueryContext` error wrap at `document_repo.go:106`.

### UT-US005-007 — `TestDocumentRepo_ListDocuments_ScanError`

- **Service:** `services/agent-board`
- **File:** `internal/repo/document_repo_test.go`
- **Test function name:** `TestDocumentRepo_ListDocuments_ScanError`
- **Given:**
  - `sqlmock.NewRows([]string{"id", "project_id", "name", ...}).AddRow("not-a-uuid", ...)` — the first column is a UUID destination but receives a value that causes a type-mismatch Scan error.
  - `mock.ExpectQuery(...).WillReturnRows(rows)`
- **When:** call `repo.ListDocuments(context.Background(), "<project-id>")`
- **Then:**
  - Non-nil error.
  - `errors.Is(err, repo.ErrNotFound)` is **false**.
  - Error message contains `"failed to scan document"` or `"error scanning"` (substring — match what the implementation wraps at `document_repo.go:113-114`).
  - `mock.ExpectationsWereMet()` nil.
- **Architecture cite:** architecture §5.1 D7; `rows.Scan` error wrap at `document_repo.go:113-114`.

### UT-US005-008 — `TestDocumentRepo_ListDocuments_RowsErr`

- **Service:** `services/agent-board`
- **File:** `internal/repo/document_repo_test.go`
- **Test function name:** `TestDocumentRepo_ListDocuments_RowsErr`
- **Given:**
  - `sqlmock.NewRows(...).AddRow(<valid row values>).RowError(0, errors.New("rows err"))`
  - `mock.ExpectQuery(...).WillReturnRows(rows)`
- **When:** call `repo.ListDocuments(context.Background(), "<project-id>")`
- **Then:**
  - Non-nil error.
  - `errors.Is(err, repo.ErrNotFound)` is **false**.
  - Error message contains `"error iterating"` (substring — matches the `rows.Err()` wrap at `document_repo.go:119-120`).
  - `mock.ExpectationsWereMet()` nil.
- **Architecture cite:** architecture §5.1 D8; `rows.Err()` wrap at `document_repo.go:119-120`.

---

### Project repo — 8 new test functions in `project_repo_test.go`

### UT-US005-009 — `TestProjectRepo_CreateProject_GenericError`

- **Service:** `services/agent-board`
- **File:** `internal/repo/project_repo_test.go`
- **Test function name:** `TestProjectRepo_CreateProject_GenericError`
- **Given:** `mock.ExpectQuery(...).WillReturnError(errors.New("db down"))`
- **When:** call `repo.CreateProject(context.Background(), <input>)`
- **Then:**
  - Non-nil error. `errors.Is(err, repo.ErrNotFound)` false. Error message contains `"failed to create project"`. `mock.ExpectationsWereMet()` nil. Project pointer nil.
- **Architecture cite:** architecture §5.2 P1; wrap at `project_repo.go:45`.

### UT-US005-010 — `TestProjectRepo_GetProject_GenericError`

- **Service:** `services/agent-board`
- **File:** `internal/repo/project_repo_test.go`
- **Test function name:** `TestProjectRepo_GetProject_GenericError`
- **Given:** `mock.ExpectQuery(...).WillReturnError(errors.New("db down"))` — non-ErrNoRows to hit line 66.
- **When:** call `repo.GetProject(context.Background(), "<any-id>")`
- **Then:**
  - Non-nil error. `errors.Is(err, repo.ErrNotFound)` false. Error message contains `"failed to get project"`. `mock.ExpectationsWereMet()` nil. Project pointer nil.
- **Architecture cite:** architecture §5.2 P2; wrap at `project_repo.go:66`.

### UT-US005-011 — `TestProjectRepo_UpdateProject_NotFound`

- **Service:** `services/agent-board`
- **File:** `internal/repo/project_repo_test.go`
- **Test function name:** `TestProjectRepo_UpdateProject_NotFound`
- **Given:** `mock.ExpectQuery(...).WillReturnError(sql.ErrNoRows)`
- **When:** call `repo.UpdateProject(context.Background(), "<any-id>", <input>)`
- **Then:**
  - Non-nil error. `errors.Is(err, repo.ErrNotFound)` **true**. `mock.ExpectationsWereMet()` nil. Project pointer nil.
- **Architecture cite:** architecture §5.2 P3; `ErrNotFound` mapping at `project_repo.go:85`.

### UT-US005-012 — `TestProjectRepo_UpdateProject_GenericError`

- **Service:** `services/agent-board`
- **File:** `internal/repo/project_repo_test.go`
- **Test function name:** `TestProjectRepo_UpdateProject_GenericError`
- **Given:** `mock.ExpectQuery(...).WillReturnError(errors.New("db down"))`
- **When:** call `repo.UpdateProject(context.Background(), "<any-id>", <input>)`
- **Then:**
  - Non-nil error. `errors.Is(err, repo.ErrNotFound)` false. Error message contains `"failed to update project"`. `mock.ExpectationsWereMet()` nil. Project pointer nil.
- **Architecture cite:** architecture §5.2 P4; wrap at `project_repo.go:87`.

### UT-US005-013 — `TestProjectRepo_DeleteProject_GenericError`

- **Service:** `services/agent-board`
- **File:** `internal/repo/project_repo_test.go`
- **Test function name:** `TestProjectRepo_DeleteProject_GenericError`
- **Given:** `mock.ExpectExec(...).WillReturnError(errors.New("db down"))`
- **When:** call `repo.DeleteProject(context.Background(), "<any-id>")`
- **Then:**
  - Non-nil error. `errors.Is(err, repo.ErrNotFound)` false. Error message contains `"failed to delete project"`. `mock.ExpectationsWereMet()` nil.
- **Architecture cite:** architecture §5.2 P5; wrap at `project_repo.go:97`.

### UT-US005-014 — `TestProjectRepo_ListProjects_QueryError`

- **Service:** `services/agent-board`
- **File:** `internal/repo/project_repo_test.go`
- **Test function name:** `TestProjectRepo_ListProjects_QueryError`
- **Given:** `mock.ExpectQuery(...).WillReturnError(errors.New("db down"))`
- **When:** call `repo.ListProjects(context.Background())`
- **Then:**
  - Non-nil error. `errors.Is(err, repo.ErrNotFound)` false. Error message contains `"failed to list projects"`. `mock.ExpectationsWereMet()` nil. Returned slice nil.
- **Architecture cite:** architecture §5.2 P6; wrap at `project_repo.go:106`.

### UT-US005-015 — `TestProjectRepo_ListProjects_ScanError`

- **Service:** `services/agent-board`
- **File:** `internal/repo/project_repo_test.go`
- **Test function name:** `TestProjectRepo_ListProjects_ScanError`
- **Given:**
  - `sqlmock.NewRows([]string{"id", "name", ...}).AddRow(<type-mismatch value>, ...)` to trigger a `rows.Scan` error.
  - `mock.ExpectQuery(...).WillReturnRows(rows)`
- **When:** call `repo.ListProjects(context.Background())`
- **Then:**
  - Non-nil error. `errors.Is(err, repo.ErrNotFound)` false. Error message contains `"failed to scan project"` or `"error scanning"` (match the wrap at `project_repo.go:113-114`). `mock.ExpectationsWereMet()` nil.
- **Architecture cite:** architecture §5.2 P7.

### UT-US005-016 — `TestProjectRepo_ListProjects_RowsErr`

- **Service:** `services/agent-board`
- **File:** `internal/repo/project_repo_test.go`
- **Test function name:** `TestProjectRepo_ListProjects_RowsErr`
- **Given:**
  - `sqlmock.NewRows(...).AddRow(<valid values>).RowError(0, errors.New("rows err"))`
  - `mock.ExpectQuery(...).WillReturnRows(rows)`
- **When:** call `repo.ListProjects(context.Background())`
- **Then:**
  - Non-nil error. `errors.Is(err, repo.ErrNotFound)` false. Error message contains `"error iterating"`. `mock.ExpectationsWereMet()` nil.
- **Architecture cite:** architecture §5.2 P8; `rows.Err()` wrap at `project_repo.go:119-120`.

---

## Coverage assertion

After all 16 tests pass, run:

```
cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo.out -v
go tool cover -func=/tmp/repo.out
```

Architecture §5.4 requires:
- `project_repo.go` ≥ 95% statement coverage.
- `document_repo.go` ≥ 95% statement coverage.
- Package `internal/repo` total ≥ 95%.
- Any uncovered line must be explicitly enumerated in the test report with a one-line justification (e.g. "panic on impossible condition — not reachable via sqlmock").
