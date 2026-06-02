# US008 — Live e2e stack-up: `docker-compose.yml` + `Makefile` + seeds + runbook

**Story:** US008 — Live e2e stack-up: docker-compose + Makefile + seeded Postgres for end-to-end Robot runs
**Requirement:** REQ005
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Implements:** Scenario: `make e2e-up` brings up the stack, Scenario: `make e2e-seed` applies migrations and seed data, Scenario: `make e2e-run` executes Robot suites against the live stack, Scenario: `make e2e` runs the whole pipeline, Scenario: `make e2e-down` tears down cleanly, Scenario: `make e2e-logs` exposes container logs, Scenario: `docker-compose.yml` lives at repo root, Scenario: seed fixtures are under version control and per-REQ, Scenario: runbook section exists, Scenario: existing per-REQ tests/e2e suites still run
**Blocked by:** none
**Worked-by:** _(none)_

## Goal

Deliver the local e2e stack-up harness — a repo-root `docker-compose.yml` (Postgres + `api-server` + `web`, all healthcheck-gated), `services/agent-board/Dockerfile`, `web/Dockerfile`, `.dockerignore`, repo-root `Makefile` with the six locked targets, `tests/e2e/data/seeds/REQ000_baseline.sql` example fixture, `tests/e2e/README.md` runbook, and `tests/e2e/results/` gitignore entry. After this task, `make e2e` is the single command that brings the stack up, applies migrations + seeds, runs Robot suites, captures results, and tears the stack down on exit even on failure.

## Scope

- **In:** New files at the paths listed in architecture §2 US008 row and §6.1 layout: `/docker-compose.yml` (three services per §6.2); `/Makefile` (six targets verbatim per §6.3); `/services/agent-board/Dockerfile` (multi-stage Go build per §6, golang:1.x-alpine builder → distroless/alpine runtime, same binary for both api-server and mcp-server via CMD override); `/web/Dockerfile` (multi-stage Node 20 build, `npm ci && npm run build && npm start` per D-012); `/.dockerignore` (`node_modules/`, `.next/`, `services/agent-board/bin/`, `*.log`, `tests/e2e/results/`); `/tests/e2e/data/seeds/REQ000_baseline.sql` (one project + two documents with deterministic UUIDs per §6.5); `/tests/e2e/README.md` (runbook per §6.6 / story AC); append `tests/e2e/results/` to `/.gitignore`.
- **Out:** CI integration (GitHub Actions, etc.); containerised Robot Framework — Robot runs on host per §6.7; replacing `psql -f` with `golang-migrate` (D-010); test-data seeding via REST/MCP API (D-010); `mcp-server` in compose (§6 note — UI e2e doesn't need it); rewriting existing `.robot` files; touching `startup.sh`.

## Files touched (estimated, exclusive)

- `/docker-compose.yml` (new)
- `/Makefile` (new)
- `/.dockerignore` (new)
- `/.gitignore` (append one line for `tests/e2e/results/`)
- `/services/agent-board/Dockerfile` (new)
- `/web/Dockerfile` (new)
- `/tests/e2e/data/seeds/REQ000_baseline.sql` (new)
- `/tests/e2e/README.md` (new)

This is effectively a **scaffold task** in spirit — it creates the root-level harness files that no other REQ005 task touches. The orchestrator should run it solo within Group C; no `Blocked by:` arrow is needed because no other REQ005 task depends on this task's deliverables (the Group A gate fixes don't depend on it; US004/US005/US006/US007/US010 don't depend on it; US009 doesn't depend on it). Phase 3c can use the stack once it lands.

## Test contract

The dev must make these tests pass (from `US008_be_unit_tests.md` and any e2e harness checks tester defines):
- Harness check: `make e2e-up` exits 0 within the healthcheck-gated time budget; `docker compose ps` shows postgres + api-server healthy + web running.
- Harness check: `make e2e-seed` is idempotent — second invocation does not error and produces the same seeded state.
- Harness check: `make e2e-run REQ=REQ001 US=US001` executes the existing REQ001 US001 Robot suite against the live stack and writes `tests/e2e/results/output.xml`.
- Harness check: `make e2e` runs up → seed → run → down (trap), propagating Robot's exit code.
- Harness check: `make e2e-down` removes containers and volumes; `docker volume ls | grep postgres-data` returns empty.
- Static check: `docker-compose.yml` binds Postgres on `127.0.0.1:15432:5432`, api-server on `127.0.0.1:8080:8080`, web on `127.0.0.1:3000:3000`.
- Static check: `tests/e2e/README.md` contains the runbook sections enumerated in §6.6 / story AC (prerequisites, target list, seed-adding guide, debug guide, orchestrator responsibility).
- Static check: `.gitignore` contains `tests/e2e/results/`.

If tester surfaces new test IDs beyond these, the dev writes them and flags the addition back to tester.

## Implementation notes

- Architecture §6.2 is the authoritative compose-file shape — copy the service definitions verbatim (refine YAML formatting as needed but keep ports, healthchecks, env vars, dependencies, and the named `postgres-data` volume exactly as specified).
- Architecture §6.3 is the authoritative Makefile — copy the six targets and the supporting variables (`DOCKER_COMPOSE`, `SEEDS_DIR`, `MIGRATIONS_DIR`, `PG_CONN`) verbatim. The `e2e` target uses `set -e` + `trap '$(MAKE) e2e-down' EXIT` to guarantee teardown.
- §6.4: migrations applied via raw `psql -v ON_ERROR_STOP=1 -f` in lex order (D-010). No `golang-migrate` dep.
- §6.5: `REQ000_baseline.sql` uses `INSERT ... ON CONFLICT DO NOTHING` for idempotency; deterministic UUIDs so Robot suites can reference them.
- §6 note (post-§6.2): `mcp-server` NOT in compose — UI e2e tests use `web` + `api-server` only.
- D-012: `web` Dockerfile runs `npm ci && npm run build && npm start` (production build, not `next dev`).
- Dockerfile multi-stage shape for `services/agent-board`: `FROM golang:1.x-alpine AS build` → `COPY go.mod go.sum && go mod download` → `COPY . && go build ./cmd/api-server && go build ./cmd/mcp-server` → `FROM gcr.io/distroless/static-debian12 AS runtime` (or alpine — pick distroless per architect's note for smaller surface) → `COPY --from=build /work/api-server /work/mcp-server /`. CMD defaults to `["/api-server"]`; compose overrides with `command: ["/api-server"]` explicitly.
- `tests/e2e/README.md` runbook content: (a) Docker prerequisites + `pip install robotframework robotframework-browser robotframework-requests`; (b) the six `make e2e-*` targets and what each does; (c) how to add a new `tests/e2e/data/seeds/REQ###_<name>.sql` (naming, idempotency requirement); (d) how to debug a failing Robot run (`make e2e-logs`, `tests/e2e/results/log.html` path); (e) orchestrator's Phase 3c responsibility — run `make e2e` after go/jest captures and include the Robot summary in the per-story test report.
- TDG skill (.claude/skills/tdg/SKILL.md) MUST be invoked at each TDD phase per be-dev workflow. The "red" phase here is the harness check that fails because `make e2e-up` doesn't exist yet on current `main`; the "green" phase is creating the compose + Makefile + Dockerfiles + seed + README. Refactor at the end (e.g. consolidating duplicated YAML anchors) is optional.
- This task does NOT modify Go source. It produces a build artefact (Dockerfile) for the existing Go binaries.

## Definition of done

- All listed harness/static checks green.
- `cd services/agent-board && go vet ./... && go test ./...` clean (no Go source touched).
- `cd web && npm run typecheck && npm test -- --watchAll=false --forceExit` clean (no FE source touched).
- `make e2e-up && make e2e-seed && make e2e-run REQ=REQ001 && make e2e-down` runs end-to-end on the dev's machine producing a real Robot output.
- `scripts/review/run-gate.sh be services/agent-board` exits with `REVIEW GATE: PASS`.
- `scripts/review/run-gate.sh fe` exits with `REVIEW GATE: PASS`.
- `scripts/review/run-gate.sh cross` exits with `REVIEW GATE: PASS`.
- `robot --dryrun tests/e2e/REQ005_*/` (if any REQ005 e2e suites land — currently none expected) passes.
- Code matches architecture §6 contract end-to-end.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Review log

(tech-lead appends here on each review pass)
