# US001/be_migrations_at_startup

**Requirement:** REQ007
**Story:** US001
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** 
**Worked-by:** 
**Implements:** D-001, D-002, D-003, Migrations data model, A. Migration at startup flow, US001

## Goal
Implement auto-running of database migrations at api-server startup using `//go:embed` and a `schema_migrations` tracking table to ensure idempotency and break the `e2e-up` circular dependency.

## Scope
- **In:** A new `internal/migrate` package that applies embedded `.up.sql` files on boot. Wiring in `cmd/api-server/main.go` to call it before listening.
- **Out:** No `/healthz` endpoint. No standalone CLI. No down-migrations. No application table changes.

## Files touched (estimated, exclusive)
- `services/agent-board/migrations.go`
- `services/agent-board/internal/migrate/migrate.go`
- `services/agent-board/internal/migrate/migrate_test.go`
- `services/agent-board/cmd/api-server/main.go`

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US001_be_unit_tests.md`: All applicable UT and IT tests.

## Implementation notes
- Create a top-level Go file (e.g. `migrations.go` in package `agentboard` or `main`) to `//go:embed migrations/*.up.sql` so it can see the directory.
- `internal/migrate/migrate.go` exposes `Run(ctx, db *sql.DB, fs fs.FS)` which:
  1. Creates `schema_migrations` table if not exists.
  2. Queries applied migrations.
  3. Sorts and applies pending `*.up.sql` files in transactions.
  4. Records the filename in `schema_migrations`.
- Update `main.go` to call `migrate.Run` after `db` is established. If it fails, return the error.

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- (Track: BE) `go test -coverprofile=/tmp/cov.out ./... && go tool cover -func=/tmp/cov.out` — every production `.go` file in this task's `## Files touched` clears ≥ 80% line coverage, OR the task has a written `## Coverage exemption` section justifying each below-threshold file.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0 AND emits `REVIEW GATE: PASS` on stdout. Also `scripts/review/run-gate.sh cross` exits 0 AND emits `REVIEW GATE: PASS`. If the REQ has Robot e2e suites, `robot --dryrun tests/e2e/REQ007_*/` also passes.
- Dev set status to `in_review` and reported back.

## Review log
