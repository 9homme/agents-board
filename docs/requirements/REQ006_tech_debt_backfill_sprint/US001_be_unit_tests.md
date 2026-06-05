# US001 — Backend unit & integration test specification
# `task_repo.go` error-branch tests

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify` + `github.com/DATA-DOG/go-sqlmock`. Tests live in `services/agent-board/internal/repo/task_repo_test.go`. Production code in `task_repo.go` is **byte-for-byte unchanged** — if a test surfaces a real bug, raise `ARCHITECTURE_GAP_FOUND`.

## Coverage matrix

| AC scenario | Layer | Test ID | Package | Function under test |
|---|---|---|---|---|
| `CreateTask` query fails | unit | UT-001 | `internal/repo` | `TaskRepo.CreateTask` |
| `GetTask` query fails | unit | UT-002 | `internal/repo` | `TaskRepo.GetTask` |
| `GetTask` returns no rows → ErrNotFound | unit | UT-003 | `internal/repo` | `TaskRepo.GetTask` |
| `UpdateTask` returns no rows → ErrNotFound | unit | UT-004 | `internal/repo` | `TaskRepo.UpdateTask` |
| `UpdateTask` query fails (non-NotFound) | unit | UT-005 | `internal/repo` | `TaskRepo.UpdateTask` |
| `UpdateTaskStatus` BeginTx fails | unit | UT-006 | `internal/repo` | `TaskRepo.UpdateTaskStatus` |
| `UpdateTaskStatus` QueryRowContext returns no rows → ErrNotFound | unit | UT-007 | `internal/repo` | `TaskRepo.UpdateTaskStatus` |
| `UpdateTaskStatus` QueryRowContext fails (non-NotFound) | unit | UT-008 | `internal/repo` | `TaskRepo.UpdateTaskStatus` |
| `UpdateTaskStatus` audit ExecContext fails → rollback | unit | UT-009 | `internal/repo` | `TaskRepo.UpdateTaskStatus` |
| `UpdateTaskStatus` Commit fails | unit | UT-010 | `internal/repo` | `TaskRepo.UpdateTaskStatus` |
| `ListTasks` QueryContext fails | unit | UT-011 | `internal/repo` | `TaskRepo.ListTasks` |
| `ListTasks` rows.Scan type-mismatch | unit | UT-012 | `internal/repo` | `TaskRepo.ListTasks` |
| `ListTasks` rows.Err() after iteration | unit | UT-013 | `internal/repo` | `TaskRepo.ListTasks` |
| per-file coverage ≥95% | integration | IT-001 | `internal/repo` | `task_repo.go` all functions |
| full suite still passes | integration | IT-002 | `internal/repo` | `go test ./...` |

## Unit tests

### UT-001 — `TestTaskRepo_CreateTask_GenericError`
- **Service:** `services/agent-board`
- **Function under test:** `TaskRepo.CreateTask`
- **Given:**
  ```go
  db, mock, _ := sqlmock.New()
  r := repo.NewTaskRepo(db)
  mock.ExpectQuery(`INSERT INTO tasks`).
      WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
      WillReturnError(errors.New("db down"))
  ```
- **When:** `r.CreateTask(context.Background(), validTask)`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `errors.Is(err, repo.ErrNotFound)` is `false`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_GenericError` branch; `task_repo.go` CreateTask error site

---

### UT-002 — `TestTaskRepo_GetTask_GenericError`
- **Service:** `services/agent-board`
- **Function under test:** `TaskRepo.GetTask`
- **Given:**
  ```go
  mock.ExpectQuery(`SELECT .* FROM tasks WHERE`).
      WithArgs("task-id-1").
      WillReturnError(errors.New("db down"))
  ```
- **When:** `r.GetTask(context.Background(), "task-id-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `errors.Is(err, repo.ErrNotFound)` is `false`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_GenericError` branch

---

### UT-003 — `TestTaskRepo_GetTask_NotFound`

**Note:** This test is NOT in the AC's required verbatim list but is required by the exhaustiveness mandate — `task_repo.go`'s `GetTask` has a `sql.ErrNoRows → ErrNotFound` mapping site. The AC note says "tester may add additional cases"; this is a required additional case per exhaustiveness.

- **Service:** `services/agent-board`
- **Function under test:** `TaskRepo.GetTask`
- **Given:**
  ```go
  mock.ExpectQuery(`SELECT .* FROM tasks WHERE`).
      WithArgs("task-id-1").
      WillReturnError(sql.ErrNoRows)
  ```
- **When:** `r.GetTask(context.Background(), "task-id-1")`
- **Then:**
  - `result` is `nil`
  - `errors.Is(err, repo.ErrNotFound)` is `true`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_NotFound` branch; `task_repo.go` sql.ErrNoRows mapping

---

### UT-004 — `TestTaskRepo_UpdateTask_NotFound`
- **Service:** `services/agent-board`
- **Function under test:** `TaskRepo.UpdateTask`
- **Given:**
  ```go
  mock.ExpectQuery(`UPDATE tasks SET`).
      WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
      WillReturnError(sql.ErrNoRows)
  ```
- **When:** `r.UpdateTask(context.Background(), validTask)`
- **Then:**
  - `result` is `nil`
  - `errors.Is(err, repo.ErrNotFound)` is `true`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_NotFound` branch; `task_repo.go:82`

---

### UT-005 — `TestTaskRepo_UpdateTask_GenericError`
- **Service:** `services/agent-board`
- **Function under test:** `TaskRepo.UpdateTask`
- **Given:**
  ```go
  mock.ExpectQuery(`UPDATE tasks SET`).
      WillReturnError(errors.New("db down"))
  ```
- **When:** `r.UpdateTask(context.Background(), validTask)`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `errors.Is(err, repo.ErrNotFound)` is `false`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_GenericError` branch; `task_repo.go:84`

---

### UT-006 — `TestTaskRepo_UpdateTaskStatus_BeginTxError`
- **Service:** `services/agent-board`
- **Function under test:** `TaskRepo.UpdateTaskStatus`
- **Given:**
  ```go
  mock.ExpectBegin().WillReturnError(errors.New("begin fail"))
  ```
- **When:** `r.UpdateTaskStatus(context.Background(), "task-id-1", "in_progress", "user-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `err.Error()` contains the wrap text from the source (e.g. `"failed to begin transaction"` — confirm exact prefix by reading `task_repo.go`)
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_BeginTxError` branch

---

### UT-007 — `TestTaskRepo_UpdateTaskStatus_NotFound`
- **Service:** `services/agent-board`
- **Function under test:** `TaskRepo.UpdateTaskStatus`
- **Given:**
  ```go
  mock.ExpectBegin()
  mock.ExpectQuery(`UPDATE tasks SET status`).
      WillReturnError(sql.ErrNoRows)
  mock.ExpectRollback()
  ```
- **When:** `r.UpdateTaskStatus(context.Background(), "task-id-1", "done", "user-1")`
- **Then:**
  - `result` is `nil`
  - `errors.Is(err, repo.ErrNotFound)` is `true`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_NotFound` within transactional path; `task_repo.go:82`

---

### UT-008 — `TestTaskRepo_UpdateTaskStatus_UpdateGenericError`
- **Service:** `services/agent-board`
- **Function under test:** `TaskRepo.UpdateTaskStatus`
- **Given:**
  ```go
  mock.ExpectBegin()
  mock.ExpectQuery(`UPDATE tasks SET status`).
      WillReturnError(errors.New("db down"))
  mock.ExpectRollback()
  ```
- **When:** `r.UpdateTaskStatus(context.Background(), "task-id-1", "done", "user-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `errors.Is(err, repo.ErrNotFound)` is `false`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_UpdateGenericError` (transactional) branch

---

### UT-009 — `TestTaskRepo_UpdateTaskStatus_AuditInsertError`
- **Service:** `services/agent-board`
- **Function under test:** `TaskRepo.UpdateTaskStatus`
- **Given:**
  ```go
  mock.ExpectBegin()
  mock.ExpectQuery(`UPDATE tasks SET status`).
      WillReturnRows(sqlmock.NewRows(taskCols).AddRow(/* happy task row */))
  mock.ExpectExec(`INSERT INTO status_audit_trail`).
      WillReturnError(errors.New("audit fail"))
  mock.ExpectRollback()
  ```
- **When:** `r.UpdateTaskStatus(context.Background(), "task-id-1", "done", "user-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `err.Error()` contains the `fmt.Errorf` wrap text from the source (e.g. `"failed to insert audit"` — confirm exact prefix by reading `task_repo.go`)
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_AuditInsertError` branch; `task_repo.go` ExecContext audit site

---

### UT-010 — `TestTaskRepo_UpdateTaskStatus_CommitError`
- **Service:** `services/agent-board`
- **Function under test:** `TaskRepo.UpdateTaskStatus`
- **Given:**
  ```go
  mock.ExpectBegin()
  mock.ExpectQuery(`UPDATE tasks SET status`).
      WillReturnRows(sqlmock.NewRows(taskCols).AddRow(/* happy task row */))
  mock.ExpectExec(`INSERT INTO status_audit_trail`).
      WillReturnResult(sqlmock.NewResult(1, 1))
  mock.ExpectCommit().WillReturnError(errors.New("commit fail"))
  ```
- **When:** `r.UpdateTaskStatus(context.Background(), "task-id-1", "done", "user-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `err.Error()` contains the `fmt.Errorf` wrap text from the source (e.g. `"failed to commit"` — confirm exact prefix by reading `task_repo.go`)
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_CommitError` branch

---

### UT-011 — `TestTaskRepo_ListTasks_QueryError`
- **Service:** `services/agent-board`
- **Function under test:** `TaskRepo.ListTasks`
- **Given:**
  ```go
  mock.ExpectQuery(`SELECT .* FROM tasks WHERE`).
      WillReturnError(errors.New("db down"))
  ```
- **When:** `r.ListTasks(context.Background(), "user-story-id-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_QueryError` branch

---

### UT-012 — `TestTaskRepo_ListTasks_ScanError`
- **Service:** `services/agent-board`
- **Function under test:** `TaskRepo.ListTasks`
- **Given:**
  ```go
  // Pass a wrong type in the first column to force Scan failure
  cols := []string{"id", "user_story_id", "title", "description", "status", "created_at", "updated_at"}
  mock.ExpectQuery(`SELECT .* FROM tasks WHERE`).
      WillReturnRows(sqlmock.NewRows(cols).AddRow(12345 /* wrong type for string id */, ...))
  ```
- **When:** `r.ListTasks(context.Background(), "user-story-id-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_ScanError` branch

---

### UT-013 — `TestTaskRepo_ListTasks_RowsErr`
- **Service:** `services/agent-board`
- **Function under test:** `TaskRepo.ListTasks`
- **Given:**
  ```go
  cols := []string{"id", "user_story_id", "title", "description", "status", "created_at", "updated_at"}
  mock.ExpectQuery(`SELECT .* FROM tasks WHERE`).
      WillReturnRows(sqlmock.NewRows(cols).AddRow(/* valid row */).RowError(0, errors.New("rows err")))
  ```
- **When:** `r.ListTasks(context.Background(), "user-story-id-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_RowsErr` branch

## Integration tests

### IT-001 — per-file coverage ≥95%
- **Service:** `services/agent-board`
- **Command:**
  ```
  cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo.out -run TestTaskRepo
  go tool cover -func=/tmp/repo.out | grep task_repo.go
  ```
- **Expect:** `task_repo.go` total statement coverage ≥95%.
- **Acceptable uncovered lines:** `task_repo.go:99` — the `defer rollback log.Printf` line (fires only when `tx.Rollback()` itself returns a non-`sql.ErrTxDone` error under sqlmock, which is not reachable). Document in test report under OQ-4.

### IT-002 — full suite regression
- **Service:** `services/agent-board`
- **Command:** `cd services/agent-board && go test ./... && golangci-lint run ./...`
- **Expect:** all pre-existing tests still pass; no new lint issues.

## Coverage exemptions

- `task_repo.go:99` — `defer rollback log.Printf` inside `UpdateTaskStatus` — the log.Printf fires only when `tx.Rollback()` itself returns an error other than `sql.ErrTxDone`; sqlmock's rollback behaviour does not surface this combination realistically. Acceptable per architecture.md §4.5.
