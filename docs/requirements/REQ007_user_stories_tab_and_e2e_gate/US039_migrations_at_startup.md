# US039 — api-server runs DB migrations at startup

**Requirement:** REQ007 — User Stories Tab + E2E Quality Gate + Health-Check Fixes
**Status:** draft

## Story
As an operator (and CI), I want the api-server binary to automatically apply all pending database migrations on boot — before it accepts traffic — so that application tables exist as soon as the server is up, breaking the circular dependency that currently makes `make e2e-up` time out (the health-check polls a DB-backed endpoint that needs tables created only by a separate seed step).

## Acceptance criteria
- **Scenario: migrations run on a fresh database at startup**
  - Given a fresh, empty database with no application tables
  - When the api-server binary starts
  - Then it applies every `.up.sql` file in `services/agent-board/migrations/` (in deterministic order — e.g. lexical by filename)
  - And it only begins accepting HTTP traffic after all migrations have completed successfully
  - And the application tables (projects, user_stories, tasks, documents, etc.) exist once the server is listening
- **Scenario: migrations are idempotent across restarts**
  - Given an api-server that has already started once and applied all migrations
  - When the api-server is restarted against the same database
  - Then it does not re-apply or error on already-applied migrations
  - And it starts and accepts traffic normally (already-applied migrations are detected and skipped)
- **Scenario: a failing migration aborts startup**
  - Given a migration file that fails to apply (e.g. a SQL error)
  - When the api-server starts
  - Then the server does NOT begin accepting traffic
  - And the process exits non-zero (or logs a fatal error and stops) so the failure is visible to the operator/CI rather than serving on a half-migrated schema
- **Scenario: DB-backed endpoint works immediately after startup, before any seed**
  - Given the api-server has started against a fresh (un-seeded) database
  - When a client sends `GET /api/v1/projects`
  - Then the response status is `200` (an empty list is fine — `{"projects":[]}`)
  - And the request does NOT error with a missing-table / undefined-relation error
  - (this is what makes the existing `make e2e-up` health-check poll on `GET /api/v1/projects` succeed without a prior seed)

## UI / UX flow expectations
No UI: startup migration is an internal server-boot behavior consumed by operators, the Makefile health-check, and CI. It has no frontend surface.

## Out of scope
- A `/healthz` or readiness endpoint — explicitly NOT part of this requirement (the original US039 healthz approach is superseded by migrations-at-startup; the existing `GET /api/v1/projects` poll is sufficient once tables exist).
- Down-migrations / rollback at startup — startup applies `.up.sql` only.
- Seeding application/test data — that remains the (now data-only) `e2e-seed` step (US040).
- A standalone migration CLI/command — migrations run as part of the normal api-server boot for this requirement (a separate migrate command is not required, though the implementation may share a package).

## Dependencies
- None blocking. Migrations live in `services/agent-board/migrations/`; api-server boot is in `services/agent-board/cmd/api-server/main.go`.

## Notes for the team
- **D-001 REVISED — CONFIRMED:** no `/healthz` endpoint. The api-server auto-runs `.up.sql` migrations at startup before accepting traffic; the `make e2e-up` health-check can keep polling `GET /api/v1/projects`, which now works because tables exist after startup migrations.
- api-server uses Echo; boot/wiring is in `cmd/api-server/main.go`. Run migrations in the boot sequence after establishing the DB connection and before `e.Start(...)` begins listening.
- Migrations must be deterministically ordered and tracked (e.g. a schema-migrations bookkeeping table or an embedded migration library) so re-runs are idempotent. The tester will need integration coverage that a fresh DB ends up fully migrated and that a restart is a no-op.
- **Testability note for the tester:** prefer integration coverage against a real/throwaway Postgres (fresh DB → start → tables present + `GET /api/v1/projects` returns `200` empty list; restart → no error). A failing-migration case proves startup aborts. Keep full container bring-up in e2e.
- Knock-on: US040 makes `e2e-seed` data-only (the migration step moves here) and drops the api-server circular-dependency fix (now resolved by this story).

## Sign-off log
(po-ba appends here on each sign-off pass)
