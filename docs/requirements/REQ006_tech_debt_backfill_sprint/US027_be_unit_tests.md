# US027 — Backend unit & integration test specification
# `audit_repo.go` error-branch tests

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify` + `github.com/DATA-DOG/go-sqlmock`. Tests live in `services/agent-board/internal/repo/audit_repo_test.go`. Production code in `audit_repo.go` is **byte-for-byte unchanged** — if a test surfaces a real bug, raise `ARCHITECTURE_GAP_FOUND`.

The private helper `getAuditTrail` has exactly three error branches — Query, Scan, RowsErr. Both public callers `GetTaskAuditTrail` and `GetUserStoryAuditTrail` are thin wrappers that differ only in the `entity_type` argument passed to the query. Covering each branch through BOTH callers gives 6 tests total and verifies the entity-type passthrough is correct.

## Coverage matrix

| AC scenario | Layer | Test ID | Package | Function under test |
|---|---|---|---|---|
| `GetTaskAuditTrail` → `getAuditTrail` QueryContext fails | unit | UT-001 | `internal/repo` | `AuditRepo.GetTaskAuditTrail` |
| `GetTaskAuditTrail` → `getAuditTrail` rows.Scan type-mismatch | unit | UT-002 | `internal/repo` | `AuditRepo.GetTaskAuditTrail` |
| `GetTaskAuditTrail` → `getAuditTrail` rows.Err() after iteration | unit | UT-003 | `internal/repo` | `AuditRepo.GetTaskAuditTrail` |
| `GetUserStoryAuditTrail` → `getAuditTrail` QueryContext fails | unit | UT-004 | `internal/repo` | `AuditRepo.GetUserStoryAuditTrail` |
| `GetUserStoryAuditTrail` → `getAuditTrail` rows.Scan type-mismatch | unit | UT-005 | `internal/repo` | `AuditRepo.GetUserStoryAuditTrail` |
| `GetUserStoryAuditTrail` → `getAuditTrail` rows.Err() after iteration | unit | UT-006 | `internal/repo` | `AuditRepo.GetUserStoryAuditTrail` |
| per-file coverage ≥95% | integration | IT-001 | `internal/repo` | `audit_repo.go` |
| full suite still passes | integration | IT-002 | `internal/repo` | `go test ./...` |

## Unit tests

### UT-001 — `TestAuditRepo_GetTaskAuditTrail_QueryError`
- **Service:** `services/agent-board`
- **Function under test:** `AuditRepo.GetTaskAuditTrail` (delegates to `getAuditTrail`)
- **Given:**
  ```go
  db, mock, _ := sqlmock.New()
  r := repo.NewAuditRepo(db)
  mock.ExpectQuery(`SELECT .* FROM status_audit_trail WHERE entity_type`).
      WithArgs("task", "task-id-1").
      WillReturnError(errors.New("db down"))
  ```
- **When:** `r.GetTaskAuditTrail(context.Background(), "task-id-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `err.Error()` contains `"failed to query audit trail"` (architecture.md §US027 AC wrap text, line 34 of `audit_repo.go`)
  - `mock.ExpectationsWereMet()` returns nil
- **Important:** the `ExpectQuery` must have `WithArgs("task", "task-id-1")` — confirms the entity-type argument is passed correctly.
- **Architecture cite:** architecture.md §4.2 `_QueryError` branch; `audit_repo.go:34`

---

### UT-002 — `TestAuditRepo_GetTaskAuditTrail_ScanError`
- **Service:** `services/agent-board`
- **Function under test:** `AuditRepo.GetTaskAuditTrail`
- **Given:**
  ```go
  // Wrong type in first column (id) forces Scan failure
  auditCols := []string{"id", "entity_type", "entity_id", "from_status", "to_status", "changed_by", "changed_at"}
  mock.ExpectQuery(`SELECT .* FROM status_audit_trail WHERE entity_type`).
      WithArgs("task", "task-id-1").
      WillReturnRows(sqlmock.NewRows(auditCols).AddRow(12345 /* wrong type for string id */, "task", "task-id-1", "pending", "done", "user-1", time.Now()))
  ```
- **When:** `r.GetTaskAuditTrail(context.Background(), "task-id-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `err.Error()` contains `"failed to scan audit trail entry"` (architecture.md §US027 AC wrap text, line 42 of `audit_repo.go`)
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_ScanError` branch; `audit_repo.go:42`

---

### UT-003 — `TestAuditRepo_GetTaskAuditTrail_RowsErr`
- **Service:** `services/agent-board`
- **Function under test:** `AuditRepo.GetTaskAuditTrail`
- **Given:**
  ```go
  auditCols := []string{"id", "entity_type", "entity_id", "from_status", "to_status", "changed_by", "changed_at"}
  mock.ExpectQuery(`SELECT .* FROM status_audit_trail WHERE entity_type`).
      WithArgs("task", "task-id-1").
      WillReturnRows(sqlmock.NewRows(auditCols).AddRow(/* valid row */).RowError(0, errors.New("rows err")))
  ```
- **When:** `r.GetTaskAuditTrail(context.Background(), "task-id-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `err.Error()` contains `"error iterating audit trail"` (architecture.md §US027 AC wrap text, line 47 of `audit_repo.go`)
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_RowsErr` branch; `audit_repo.go:47`

---

### UT-004 — `TestAuditRepo_GetUserStoryAuditTrail_QueryError`
- **Service:** `services/agent-board`
- **Function under test:** `AuditRepo.GetUserStoryAuditTrail`
- **Given:**
  ```go
  mock.ExpectQuery(`SELECT .* FROM status_audit_trail WHERE entity_type`).
      WithArgs("user_story", "us-id-1").
      WillReturnError(errors.New("db down"))
  ```
- **When:** `r.GetUserStoryAuditTrail(context.Background(), "us-id-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `err.Error()` contains `"failed to query audit trail"`
  - `mock.ExpectationsWereMet()` returns nil
- **Important:** `WithArgs("user_story", "us-id-1")` — confirms the entity-type is `"user_story"` not `"task"`.
- **Architecture cite:** architecture.md §4.2 `_QueryError` branch; `audit_repo.go:34`

---

### UT-005 — `TestAuditRepo_GetUserStoryAuditTrail_ScanError`
- **Service:** `services/agent-board`
- **Function under test:** `AuditRepo.GetUserStoryAuditTrail`
- **Given:**
  ```go
  auditCols := []string{"id", "entity_type", "entity_id", "from_status", "to_status", "changed_by", "changed_at"}
  mock.ExpectQuery(`SELECT .* FROM status_audit_trail WHERE entity_type`).
      WithArgs("user_story", "us-id-1").
      WillReturnRows(sqlmock.NewRows(auditCols).AddRow(12345 /* wrong type */, "user_story", "us-id-1", "draft", "done", "user-1", time.Now()))
  ```
- **When:** `r.GetUserStoryAuditTrail(context.Background(), "us-id-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `err.Error()` contains `"failed to scan audit trail entry"`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_ScanError` branch; `audit_repo.go:42`

---

### UT-006 — `TestAuditRepo_GetUserStoryAuditTrail_RowsErr`
- **Service:** `services/agent-board`
- **Function under test:** `AuditRepo.GetUserStoryAuditTrail`
- **Given:**
  ```go
  auditCols := []string{"id", "entity_type", "entity_id", "from_status", "to_status", "changed_by", "changed_at"}
  mock.ExpectQuery(`SELECT .* FROM status_audit_trail WHERE entity_type`).
      WithArgs("user_story", "us-id-1").
      WillReturnRows(sqlmock.NewRows(auditCols).AddRow(/* valid row */).RowError(0, errors.New("rows err")))
  ```
- **When:** `r.GetUserStoryAuditTrail(context.Background(), "us-id-1")`
- **Then:**
  - `result` is `nil`
  - `err` is non-nil
  - `err.Error()` contains `"error iterating audit trail"`
  - `mock.ExpectationsWereMet()` returns nil
- **Architecture cite:** architecture.md §4.2 `_RowsErr` branch; `audit_repo.go:47`

## Integration tests

### IT-001 — per-file coverage ≥95%
- **Service:** `services/agent-board`
- **Command:**
  ```
  cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo.out -run TestAuditRepo
  go tool cover -func=/tmp/repo.out | grep audit_repo.go
  ```
- **Expect:** `audit_repo.go` total statement coverage ≥95%. Baseline was `getAuditTrail` at 78.6%; 6 tests exercising all three error branches through both public callers should lift this well past 95%.

### IT-002 — full suite regression
- **Service:** `services/agent-board`
- **Command:** `cd services/agent-board && go test ./... && golangci-lint run ./...`
- **Expect:** all pre-existing tests still pass; no new lint issues.

## Coverage exemptions

None anticipated for `audit_repo.go`. The three error branches through both callers cover every reachable line in `getAuditTrail`. If any line is found to be unreachable via sqlmock, document under OQ-4 in the test report.
