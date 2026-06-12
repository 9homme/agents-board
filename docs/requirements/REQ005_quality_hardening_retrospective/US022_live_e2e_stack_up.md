# US022 — Live e2e stack-up: docker-compose + Makefile + seeded Postgres for end-to-end Robot runs

**Requirement:** REQ005 — quality hardening retrospective
**Status:** draft

## Story
As an **orchestrator (or any developer) finishing a story's Phase 3c test-report capture**, I want a **single command to bring up `web` + `api-server` + a seeded Postgres** so that `robot --include US### tests/e2e/REQ###_*/` runs end-to-end against a live stack, instead of being skipped at po-ba sign-off with "dry-run + comprehensive unit/component coverage."

## Acceptance criteria

- **Scenario: `make e2e-up` brings up the stack**
  - Given a clean checkout with Docker Desktop (or compatible) running
  - When `make e2e-up` is invoked from the repo root
  - Then a Postgres container starts and reaches `ready to accept connections` (healthcheck-gated)
  - And the `api-server` container starts (or local binary; see "Notes for the team") with `DATABASE_URL` pointing at the compose Postgres and responds to `GET /api/v1/projects` with 200 (after migrations + seeds applied)
  - And the `web` container (or local `next dev`) is reachable at `http://localhost:3000`
  - And `make e2e-up` exits 0 once all three are healthy
  - And the total stand-up time is under 60 seconds on a typical dev laptop after first-pull (`docker-compose pull`)

- **Scenario: `make e2e-seed` applies migrations and seed data**
  - Given `make e2e-up` has run (or the seed target is invoked standalone after `e2e-up`)
  - When `make e2e-seed` is invoked
  - Then `services/agent-board/migrations/*.up.sql` are applied in order (using whatever migration mechanism the team picks: `migrate -path migrations -database $DATABASE_URL up`, a Go-based runner, or `psql -f` for the simplest path)
  - And SQL fixtures from `tests/e2e/data/seeds/*.sql` are applied (alphabetical order)
  - And `make e2e-seed` is idempotent (re-running it on an already-seeded DB does not error — uses `INSERT ... ON CONFLICT DO NOTHING` or `TRUNCATE` + insert, tech-lead's call)

- **Scenario: `make e2e-run` executes Robot suites against the live stack**
  - Given `make e2e-up && make e2e-seed` has succeeded
  - When `make e2e-run` is invoked (optionally `make e2e-run REQ=REQ001 US=US015` for narrowing)
  - Then Robot Framework runs the matching `.robot` files under `tests/e2e/REQ*/`
  - And tests pass (assuming seed data matches what the suites expect)
  - And the Robot output (`output.xml`, `log.html`, `report.html`) is written to `tests/e2e/results/` (gitignored)

- **Scenario: `make e2e` runs the whole pipeline**
  - Given a clean checkout
  - When `make e2e` is invoked
  - Then it runs `e2e-up` → `e2e-seed` → `e2e-run` → captures exit code → `e2e-down` (always, even on failure — trap)
  - And exits with the Robot run's exit code

- **Scenario: `make e2e-down` tears down cleanly**
  - Given the stack is up
  - When `make e2e-down` is invoked
  - Then `docker-compose down -v` (or equivalent) runs
  - And all containers stop and volumes are removed
  - And exit code is 0

- **Scenario: `make e2e-logs` exposes container logs**
  - Given the stack is up (or has crashed)
  - When `make e2e-logs` is invoked
  - Then logs from `api-server`, `web`, and `postgres` containers are streamed (or printed)

- **Scenario: `docker-compose.yml` lives at repo root**
  - Given the repo structure after the story
  - When the root is listed
  - Then `docker-compose.yml` (or `compose.yaml`) exists at `/Users/a667282/workspace/agents-board/docker-compose.yml`
  - And it defines services: `postgres` (with healthcheck), `api-server`, `web` (web optional if the team prefers running `next dev` locally — see "Notes for the team")
  - And the compose file uses bind mounts or explicit images so the running stack reflects the checkout (no stale containers)
  - And ports are bound to localhost only (no `0.0.0.0` exposure beyond what's needed)

- **Scenario: seed fixtures are under version control and per-REQ**
  - Given a new REQ wants distinct seed data
  - When a developer adds `tests/e2e/data/seeds/REQ###_<name>.sql`
  - Then `make e2e-seed` picks it up (alphabetical glob — files are sorted REQ001 → REQ002 → ...)
  - And there is an example seed file (`tests/e2e/data/seeds/REQ001_baseline.sql` or similar) committed to demonstrate the pattern

- **Scenario: runbook section exists**
  - Given `tests/e2e/README.md` (created if missing)
  - When a new developer reads it
  - Then it explains: (a) Docker prerequisites, (b) the `make e2e-*` target list and what each does, (c) how to add a new seed fixture, (d) how to debug a failing Robot run (`make e2e-logs`, results dir location), (e) the orchestrator's responsibility — run `make e2e` after Phase 3c go/jest captures and include the Robot summary in the per-story test report.

- **Scenario: existing per-REQ tests/e2e suites still run**
  - Given existing `tests/e2e/REQ###_*/US###_*.robot` files
  - When `make e2e-run REQ=REQ001 US=US015` is invoked
  - Then those suites execute against the live stack
  - And tests that previously could not run (because the stack was unreachable) now produce real PASS / FAIL outcomes
  - Note: this story does NOT require that previously-deferred e2e suites pass — only that they CAN run. If they fail because of stale fixtures or stale assumptions, that's a per-REQ follow-up (the orchestrator will surface them at sign-off).

## UI / UX flow expectations

**No end-user UI:** developer / orchestrator harness. For completeness:

- **Entry points:** developer in repo root runs `make e2e-up && make e2e-seed && make e2e-run` (or the all-in-one `make e2e`).
- **Operator flow:** orchestrator in Phase 3c spawns `make e2e-run REQ=REQ005 US=US022` (or whatever) after BE / FE test captures, parses the Robot output for E2E-* IDs, writes them into the per-story `US###_test_report.md`.
- **Failure flow:** Robot test fails → developer reads `tests/e2e/results/log.html` → fixes seed or test → reruns `make e2e-run`.
- **Out of UI scope:** any visual reporting dashboard. CLI + Robot's HTML output is enough.

## Out of scope
- **Hosting the stack on a cloud CI runner.** Local docker-compose is the deliverable. A CI hookup (GitHub Actions, etc.) is a follow-up story.
- **A dedicated test-data fixture loader (Go binary or Python script).** SQL files + psql / migrate is sufficient and simpler.
- **Containerising the Robot Framework runner itself.** Robot runs on the host (uses the local Python install) and hits the containerised stack over HTTP.
- **Replacing or rewriting any existing `.robot` files.** This story enables them to run; updating them is per-REQ work.
- **Rewriting `Dockerfile`s for `api-server` or `web`** beyond what's needed to start them in compose. If a `Dockerfile` doesn't exist yet, this story may add a minimal one per service. If they exist, reuse.
- **Test-data seeding via API calls** (the alternative to SQL fixtures). SQL is faster and simpler; if the team later wants API-driven seeding, that's a new story.

## Dependencies
- Soft: US015 + US016 + US017 should land first so the per-track gates can run cleanly while this story is in development. Not a hard blocker.

## Notes for the team

- **Web container vs local `next dev`.** Architect / tech-lead's call. Two valid shapes:
  1. **Containerise `web` too** — full reproducibility, one `docker-compose up`. Cons: slower iteration loop for FE devs.
  2. **`web` runs as `next dev` on the host, only Postgres + `api-server` containerised** — faster FE iteration, but the Makefile target needs to start `next dev` in a background process or instruct the user to start it.
  - po-ba's mild preference: option 1 (everything containerised) so the orchestrator has a single command. Architect can override.
- **api-server container.** Easiest: build from local source (`build: ./services/agent-board` with a `Dockerfile` that does `go build` + minimal runtime). If a Dockerfile doesn't exist, add a multi-stage one.
- **Postgres image.** Pin to `postgres:16-alpine` (or whatever matches the migrations' assumed Postgres version — check `migrations/000001_init_schema.up.sql` for any version-specific syntax).
- **Migrations runner.** Three options ranked by simplicity:
  1. `psql -f migrations/*.up.sql` — works for now; brittle if migrations have semicolon-in-string edge cases.
  2. `migrate -path services/agent-board/migrations -database $DATABASE_URL up` (using `golang-migrate/migrate`) — robust, requires installing the tool.
  3. A small Go program in `services/agent-board/cmd/migrate/` that uses `embed.FS`.
  - po-ba's mild preference: option 1 for now (no new tool dep), upgrade later.
- **Volume mounts for seeds.** Mount `tests/e2e/data/seeds/` into the postgres container's `/docker-entrypoint-initdb.d/` IF and only if we accept that initdb scripts run once at container init. Alternative: run seeds from the Makefile after the container is up (more flexible).
- **`make e2e` exit semantics.** Use `set -e` + a `trap` to ensure `e2e-down` runs even on failure of `e2e-run`. Pseudocode: `make e2e-up && make e2e-seed && (make e2e-run; rc=$?; make e2e-down; exit $rc)`.
- **Ports.** Don't collide with whatever the developer might run locally — Postgres on `15432:5432`, api-server on `8080:8080` (already its default), web on `3000:3000`.
- **Gitignore.** Add `tests/e2e/results/` and any `*.pid` / Docker volume artefacts to `.gitignore`.
- **What "live" means for sign-off.** After US022 ships, po-ba's sign-off REJECTS "e2e dry-run" as a substitute. Every story with an e2e spec must have a real Robot run captured in `US###_test_report.md`. This is a behavioural change to po-ba's sign-off discipline that the team should be aware of; the change to po-ba's agent definition is OUT of scope here (po-ba's agent prompt does not need a code change — sign-off can simply hold the line once the harness exists).

## Sign-off log
(po-ba appends here on each sign-off pass)
