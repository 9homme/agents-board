# US002 — Backend unit & integration test specification
# `user_story_repo.go` error-branch tests

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify` + `github.com/DATA-DOG/go-sqlmock`. Tests live in `services/agent-board/internal/repo/user_story_repo_test.go`. Production code in `user_story_repo.go` is **byte-for-byte unchanged** — if a test surfaces a real bug, raise `ARCHITECTURE_GAP_FOUND`.

## Coverage matrix

| AC scenario | Layer | Test ID | Package | Function under test |
|---|---|---|---|---|
| `CreateUserStory` query fails | unit | UT-001 | `internal/repo` | `UserStoryRepo.CreateUserStory` |
| `GetUserStory` query fails | unit | UT-002 | `internal/repo` | `UserStoryRepo.GetUserStory` |
| `GetUserStory` returns no rows → ErrNotFound | unit | UT-003 | `internal/repo` | `UserStoryRepo.GetUserStory` |
| `UpdateUserStory` returns no rows → ErrNotFound | unit | UT-004 | `internal/repo` | `UserStoryRepo.UpdateUserStory` |
| `UpdateUserStory` query fails (non-NotFound) | unit | UT-005 | `internal/repo` | `UserStoryRepo.UpdateUserStory` |
| `UpdateUserStoryStatus` BeginTx fails | unit | UT-006 | `internal/repo` | `UserStoryRepo.UpdateUserStoryStatus` |
| `UpdateUserStoryStatus` QueryRowContext returns no rows → ErrNotFound | unit | UT-007 | `internal/repo` | `UserStoryRepo.UpdateUserStoryStatus` |
| `UpdateUserStoryStatus` QueryRowContext fails (non-NotFound) | unit | UT-008 | `internal/repo` | `UserStoryRepo.UpdateUserStoryStatus` |
| `UpdateUserStoryStatus` audit ExecContext fails → rollback | unit | UT-009 | `internal/repo` | `UserStoryRepo.UpdateUserStoryStatus` |
| `UpdateUserStoryStatus` Commit fails | unit | UT-010 | `internal/repo` | `UserStoryRepo.UpdateUserStoryStatus` |
| `ListUserStories` QueryContext fails | unit | UT-011 | `internal/repo` | `UserStoryRepo.ListUserStories` |
| `ListUserStories` rows.Scan type-mismatch | unit | UT-012 | `internal/repo` | `UserStoryRepo.ListUserStories` |
| `ListUserStories` rows.Err() after iteration | unit | UT-013 | `internal/repo` | `UserStoryRepo.ListUserStories` |
| per-file coverage ≥95% | integration | IT-001 | `internal/repo` | `user_story_repo.go` all functions |
| full suite still passes | integration | IT-002 | `internal/repo` | `go test ./...` |

## Unit tests

### UT-001 — `TestUserStoryRepo_CreateUserStory_GenericError`
- **Service:** `services/agent-board`
- **Function under test:** `UserStoryRepo.CreateUserStory`
- **Given:**
  ```go
  db, mock, _ := sqlmock.New()
  r := repo.NewUserStoryRepo(db)
  mock.ExpectQuery(`INSERT INTO user_stories`).
      WillReturnError(errors.New("db down"))
  ```
- **When:** `r.CreateUserStory(context.Background(), validUserStory)`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `errors.Is(err, repo.ErrNotFound)` is `false`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_GenericError` branch

---

### UT-002 — `TestUserStoryRepo_GetUserStory_GenericError`
- **Service:** `services/agent-board`
- **Function under test:** `UserStoryRepo.GetUserStory`
- **Given:**
  ```go
  mock.ExpectQuery(`SELECT .* FROM user_stories WHERE`).
      WithArgs("us-id-1").
      WillReturnError(errors.New("db down"))
  ```
- **When:** `r.GetUserStory(context.Background(), "us-id-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `errors.Is(err, repo.ErrNotFound)` is `false`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_GenericError` branch

---

### UT-003 — `TestUserStoryRepo_GetUserStory_NotFound`

**Note:** Required by exhaustiveness mandate — `GetUserStory` has a `sql.ErrNoRows → ErrNotFound` mapping site. The AC says "tester may add additional cases"; this is a mandatory additional case.

- **Service:** `services/agent-board`
- **Function under test:** `UserStoryRepo.GetUserStory`
- **Given:**
  ```go
  mock.ExpectQuery(`SELECT .* FROM user_stories WHERE`).
      WithArgs("us-id-1").
      WillReturnError(sql.ErrNoRows)
  ```
- **When:** `r.GetUserStory(context.Background(), "us-id-1")`
- **Then:**
  - `result` is `nil`
  - `errors.Is(err, repo.ErrNotFound)` is `true`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_NotFound` branch; `user_story_repo.go` sql.ErrNoRows mapping

---

### UT-004 — `TestUserStoryRepo_UpdateUserStory_NotFound`
- **Service:** `services/agent-board`
- **Function under test:** `UserStoryRepo.UpdateUserStory`
- **Given:**
  ```go
  mock.ExpectQuery(`UPDATE user_stories SET`).
      WillReturnError(sql.ErrNoRows)
  ```
- **When:** `r.UpdateUserStory(context.Background(), validUserStory)`
- **Then:**
  - `result` is `nil`
  - `errors.Is(err, repo.ErrNotFound)` is `true`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_NotFound` branch; `user_story_repo.go:102`

---

### UT-005 — `TestUserStoryRepo_UpdateUserStory_GenericError`
- **Service:** `services/agent-board`
- **Function under test:** `UserStoryRepo.UpdateUserStory`
- **Given:**
  ```go
  mock.ExpectQuery(`UPDATE user_stories SET`).
      WillReturnError(errors.New("db down"))
  ```
- **When:** `r.UpdateUserStory(context.Background(), validUserStory)`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `errors.Is(err, repo.ErrNotFound)` is `false`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_GenericError` branch; `user_story_repo.go:104`

---

### UT-006 — `TestUserStoryRepo_UpdateUserStoryStatus_BeginTxError`
- **Service:** `services/agent-board`
- **Function under test:** `UserStoryRepo.UpdateUserStoryStatus`
- **Given:**
  ```go
  mock.ExpectBegin().WillReturnError(errors.New("begin fail"))
  ```
- **When:** `r.UpdateUserStoryStatus(context.Background(), "us-id-1", "in_development", "user-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `err.Error()` contains the wrap text from the source (confirm exact prefix by reading `user_story_repo.go`)
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_BeginTxError` branch

---

### UT-007 — `TestUserStoryRepo_UpdateUserStoryStatus_NotFound`
- **Service:** `services/agent-board`
- **Function under test:** `UserStoryRepo.UpdateUserStoryStatus`
- **Given:**
  ```go
  mock.ExpectBegin()
  mock.ExpectQuery(`UPDATE user_stories SET status`).
      WillReturnError(sql.ErrNoRows)
  mock.ExpectRollback()
  ```
- **When:** `r.UpdateUserStoryStatus(context.Background(), "us-id-1", "done", "user-1")`
- **Then:**
  - `result` is `nil`
  - `errors.Is(err, repo.ErrNotFound)` is `true`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_NotFound` within transactional path; `user_story_repo.go:65-71`

---

### UT-008 — `TestUserStoryRepo_UpdateUserStoryStatus_UpdateGenericError`
- **Service:** `services/agent-board`
- **Function under test:** `UserStoryRepo.UpdateUserStoryStatus`
- **Given:**
  ```go
  mock.ExpectBegin()
  mock.ExpectQuery(`UPDATE user_stories SET status`).
      WillReturnError(errors.New("db down"))
  mock.ExpectRollback()
  ```
- **When:** `r.UpdateUserStoryStatus(context.Background(), "us-id-1", "done", "user-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `errors.Is(err, repo.ErrNotFound)` is `false`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_UpdateGenericError` (transactional) branch

---

### UT-009 — `TestUserStoryRepo_UpdateUserStoryStatus_AuditInsertError`
- **Service:** `services/agent-board`
- **Function under test:** `UserStoryRepo.UpdateUserStoryStatus`
- **Given:**
  ```go
  mock.ExpectBegin()
  mock.ExpectQuery(`UPDATE user_stories SET status`).
      WillReturnRows(sqlmock.NewRows(userStoryCols).AddRow(/* happy user story row */))
  mock.ExpectExec(`INSERT INTO status_audit_trail`).
      WillReturnError(errors.New("audit fail"))
  mock.ExpectRollback()
  ```
- **When:** `r.UpdateUserStoryStatus(context.Background(), "us-id-1", "done", "user-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `err.Error()` contains the `fmt.Errorf` wrap text from the source (confirm exact prefix by reading `user_story_repo.go`)
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_AuditInsertError` branch

---

### UT-010 — `TestUserStoryRepo_UpdateUserStoryStatus_CommitError`
- **Service:** `services/agent-board`
- **Function under test:** `UserStoryRepo.UpdateUserStoryStatus`
- **Given:**
  ```go
  mock.ExpectBegin()
  mock.ExpectQuery(`UPDATE user_stories SET status`).
      WillReturnRows(sqlmock.NewRows(userStoryCols).AddRow(/* happy user story row */))
  mock.ExpectExec(`INSERT INTO status_audit_trail`).
      WillReturnResult(sqlmock.NewResult(1, 1))
  mock.ExpectCommit().WillReturnError(errors.New("commit fail"))
  ```
- **When:** `r.UpdateUserStoryStatus(context.Background(), "us-id-1", "done", "user-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_CommitError` branch

---

### UT-011 — `TestUserStoryRepo_ListUserStories_QueryError`
- **Service:** `services/agent-board`
- **Function under test:** `UserStoryRepo.ListUserStories`
- **Given:**
  ```go
  mock.ExpectQuery(`SELECT .* FROM user_stories WHERE`).
      WillReturnError(errors.New("db down"))
  ```
- **When:** `r.ListUserStories(context.Background(), "project-id-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_QueryError` branch

---

### UT-012 — `TestUserStoryRepo_ListUserStories_ScanError`
- **Service:** `services/agent-board`
- **Function under test:** `UserStoryRepo.ListUserStories`
- **Given:**
  ```go
  // Pass a wrong type in the first column to force Scan failure
  cols := []string{"id", "project_id", "title", "description", "status", "created_at", "updated_at"}
  mock.ExpectQuery(`SELECT .* FROM user_stories WHERE`).
      WillReturnRows(sqlmock.NewRows(cols).AddRow(12345 /* wrong type for string id */, ...))
  ```
- **When:** `r.ListUserStories(context.Background(), "project-id-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_ScanError` branch

---

### UT-013 — `TestUserStoryRepo_ListUserStories_RowsErr`
- **Service:** `services/agent-board`
- **Function under test:** `UserStoryRepo.ListUserStories`
- **Given:**
  ```go
  cols := []string{"id", "project_id", "title", "description", "status", "created_at", "updated_at"}
  mock.ExpectQuery(`SELECT .* FROM user_stories WHERE`).
      WillReturnRows(sqlmock.NewRows(cols).AddRow(/* valid row */).RowError(0, errors.New("rows err")))
  ```
- **When:** `r.ListUserStories(context.Background(), "project-id-1")`
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
  cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo.out -run TestUserStoryRepo
  go tool cover -func=/tmp/repo.out | grep user_story_repo.go
  ```
- **Expect:** `user_story_repo.go` total statement coverage ≥95%.
- **Acceptable uncovered lines:** `user_story_repo.go:68` — the `defer rollback log.Printf` line (fires only when `tx.Rollback()` itself returns a non-`sql.ErrTxDone` error under sqlmock). Document in test report under OQ-4.

### IT-002 — full suite regression
- **Service:** `services/agent-board`
- **Command:** `cd services/agent-board && go test ./... && golangci-lint run ./...`
- **Expect:** all pre-existing tests still pass; no new lint issues.

## Coverage exemptions

- `user_story_repo.go:68` — `defer rollback log.Printf` inside `UpdateUserStoryStatus` — unreachable via sqlmock (rollback returning non-ErrTxDone). Acceptable per architecture.md §4.5.
