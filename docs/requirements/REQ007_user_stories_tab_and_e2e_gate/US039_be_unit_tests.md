# US039 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`.

## Coverage matrix
| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| Startup migration applies on fresh DB | integration | IT-001 | services/agent-board / internal/migrate | `migrate.Run(...)` |
| Idempotent restart skips applied versions | integration | IT-002 | services/agent-board / internal/migrate | `migrate.Run(...)` |
| Startup fails if schema_migrations table creation fails | unit | UT-001 | services/agent-board / internal/migrate | `migrate.Run(...)` |
| Startup fails on fetching applied versions | unit | UT-002 | services/agent-board / internal/migrate | `migrate.Run(...)` |
| Startup fails on transaction begin | unit | UT-003 | services/agent-board / internal/migrate | `migrate.Run(...)` |
| Startup fails on invalid migration SQL | unit | UT-004 | services/agent-board / internal/migrate | `migrate.Run(...)` |
| Startup fails on recording applied version | unit | UT-005 | services/agent-board / internal/migrate | `migrate.Run(...)` |
| Startup fails on transaction commit | unit | UT-006 | services/agent-board / internal/migrate | `migrate.Run(...)` |

## Unit tests

### UT-001 — Startup fails if schema_migrations table creation fails
- **Service:** `services/agent-board`
- **Function under test:** `internal/migrate.Run`
- **Given:** A mocked `sql.DB` that returns an error when executing the `CREATE TABLE IF NOT EXISTS schema_migrations` statement.
- **When:** call `Run(...)`
- **Then:** returns the error and does not proceed.

### UT-002 — Startup fails on fetching applied versions
- **Service:** `services/agent-board`
- **Function under test:** `internal/migrate.Run`
- **Given:** A mocked `sql.DB` that returns an error when executing `SELECT id FROM schema_migrations`.
- **When:** call `Run(...)`
- **Then:** returns the error.

### UT-003 — Startup fails on transaction begin
- **Service:** `services/agent-board`
- **Function under test:** `internal/migrate.Run`
- **Given:** A mocked `sql.DB` that returns an error on `BeginTx()`.
- **When:** call `Run(...)`
- **Then:** returns the error, proving the migration aborts.

### UT-004 — Startup fails on invalid migration SQL
- **Service:** `services/agent-board`
- **Function under test:** `internal/migrate.Run`
- **Given:** A mocked `sql.DB` (via sqlmock) that returns an error when executing the embedded migration SQL.
- **When:** call `Run(...)`
- **Then:** returns the error, proving the transaction rolls back.

### UT-005 — Startup fails on recording applied version
- **Service:** `services/agent-board`
- **Function under test:** `internal/migrate.Run`
- **Given:** A mocked `sql.DB` that succeeds executing migration SQL but fails on `INSERT INTO schema_migrations`.
- **When:** call `Run(...)`
- **Then:** returns the error, proving the transaction rolls back.

### UT-006 — Startup fails on transaction commit
- **Service:** `services/agent-board`
- **Function under test:** `internal/migrate.Run`
- **Given:** A mocked `sql.DB` that returns an error on `tx.Commit()`.
- **When:** call `Run(...)`
- **Then:** returns the error.

## Integration tests
### IT-001 — Startup migration applies on fresh DB
- **Service:** `services/agent-board`
- **Boundary:** `migrate` package ↔ real test Postgres
- **Setup:** A fresh Postgres database (e.g., testcontainers).
- **Endpoint exercised:** `migrate.Run(ctx, db)`
- **Expect:** Returns no error. `schema_migrations` table contains entries for the embedded files. A follow-up `SELECT count(*) FROM projects` succeeds, proving tables were created.
- **Teardown:** Close DB connection, drop container.

### IT-002 — Idempotent restart skips applied versions
- **Service:** `services/agent-board`
- **Boundary:** `migrate` package ↔ real test Postgres
- **Setup:** A Postgres database that has already been migrated once (e.g., call `migrate.Run` once).
- **Endpoint exercised:** `migrate.Run(ctx, db)` (second time)
- **Expect:** Returns no error. No tables are dropped or recreated (idempotency). The `schema_migrations` table count remains the same.
- **Teardown:** Close DB connection, drop container.

## Spec change log
### Revision 1 — 2024-03-XX — driver: po-ba sign-off pass
- expanded `UT-*` cases to explicitly cover distinct error sites in `Run()`: table creation, select versions, BeginTx, file-SQL exec, INSERT schema_migrations, and tx.Commit.