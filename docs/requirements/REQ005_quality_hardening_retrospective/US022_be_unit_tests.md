# US022 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). US022 introduces infrastructure files (docker-compose.yml, Makefile, Dockerfiles, seed SQL). The "tests" here are structural and dry-run assertions — not Go unit tests. They may be implemented as shell scripts in `tests/e2e/` or as part of the Makefile's own self-test, as appropriate.

No new Go packages are introduced. No existing Go packages are changed. The BE dev implementing US022 is responsible for the Makefile targets, docker-compose, and seed SQL; verification is through Makefile dry-runs and seed SQL validity checks.

## Coverage matrix

| AC scenario | Layer | Test ID | File / target | Behaviour under test |
|---|---|---|---|---|
| `make -n e2e-up` outputs expected compose command | integration | IT-US022-001 | `/Makefile` | dry-run produces `docker compose up -d --wait` |
| `make -n e2e-down` outputs expected compose command | integration | IT-US022-002 | `/Makefile` | dry-run produces `docker compose down -v` |
| `make -n e2e-seed` outputs migration and seed psql commands | integration | IT-US022-003 | `/Makefile` | dry-run lists `psql` invocations for each `.up.sql` and `.sql` seed |
| Seed SQL parses without syntax errors | integration | IT-US022-004 | `tests/e2e/data/seeds/REQ000_baseline.sql` | `psql --dry-run` or `psql -c '\i...'` with `ON_ERROR_STOP=1` against a test DB |
| Migration SQL files pass syntax check | integration | IT-US022-005 | `services/agent-board/migrations/*.up.sql` | same psql parse check against test DB |
| `docker-compose.yml` is valid YAML and passes compose parse | integration | IT-US022-006 | `/docker-compose.yml` | `docker compose config` exits 0 |
| `make e2e-seed` is idempotent | integration | IT-US022-007 | `/Makefile` + seed SQL | run seed twice against the same DB; second run exits 0 |

## Integration tests

These are shell-harness integration tests. The "red" phase is running the assertions against the repo before US022 files exist (they fail because the files don't exist); the "green" phase is after the files land.

### IT-US022-001 — `make -n e2e-up` dry-run outputs compose up command

- **File under test:** `/Makefile` (repo root)
- **Setup:** the Makefile must exist.
- **When:** `make -n e2e-up` (dry-run flag — prints commands without executing them)
- **Then:**
  - Output contains `docker compose up -d --wait` (or `docker-compose up -d --wait` — both are acceptable per architecture §6.3 `DOCKER_COMPOSE ?= docker compose`).
  - Exit code is 0.
- **Architecture cite:** architecture §6.3 — `e2e-up` target: `$(DOCKER_COMPOSE) up -d --wait`.

### IT-US022-002 — `make -n e2e-down` dry-run outputs compose down command

- **File under test:** `/Makefile`
- **When:** `make -n e2e-down`
- **Then:**
  - Output contains `docker compose down -v`.
  - Exit code is 0.
- **Architecture cite:** architecture §6.3 — `e2e-down` target: `$(DOCKER_COMPOSE) down -v`.

### IT-US022-003 — `make -n e2e-seed` dry-run shows psql invocations

- **File under test:** `/Makefile`
- **When:** `make -n e2e-seed`
- **Then:**
  - Output contains at least one line referencing a `.up.sql` file from `services/agent-board/migrations/` with `psql`.
  - If the seeds directory exists and has `.sql` files, output also contains a psql invocation referencing `tests/e2e/data/seeds/`.
  - Exit code is 0.
- **Architecture cite:** architecture §6.3 — `e2e-seed` target; §6.4 migrations runner (`psql -v ON_ERROR_STOP=1 -f`).

### IT-US022-004 — Seed SQL file parses without syntax errors

- **File under test:** `tests/e2e/data/seeds/REQ000_baseline.sql`
- **Boundary:** parse-only check against a test Postgres instance (or a disposable Docker container)
- **Setup:** requires a reachable Postgres. Can use the compose stack's Postgres (`make e2e-up` then test) OR a throwaway `docker run --rm postgres:16-alpine` for a parse-only check.
- **When:** `psql "${PG_CONN}" -v ON_ERROR_STOP=1 -f tests/e2e/data/seeds/REQ000_baseline.sql`
- **Then:** exit code is 0; no SQL error is reported.
- **Notes:** The seed must use `INSERT ... ON CONFLICT DO NOTHING` per architecture §6.5 (idempotency contract). Assert the file contains `ON CONFLICT DO NOTHING` or `TRUNCATE` at the top.
- **Architecture cite:** architecture §6.5 seed fixture contract.

### IT-US022-005 — Migration SQL files parse without syntax errors

- **File under test:** `services/agent-board/migrations/*.up.sql`
- **Boundary:** same as IT-US022-004
- **When:** for each `.up.sql` in lex order, `psql "${PG_CONN}" -v ON_ERROR_STOP=1 -f <file>`
- **Then:** all exit 0; no SQL error is reported.
- **Architecture cite:** architecture §6.4 — `psql -v ON_ERROR_STOP=1 -f`.

### IT-US022-006 — `docker-compose.yml` passes compose validation

- **File under test:** `/docker-compose.yml`
- **When:** `docker compose config` (validates and prints the merged compose config)
- **Then:** exit code is 0; no parse or validation errors.
- **Architecture cite:** architecture §6.2 — compose service definitions (postgres, api-server, web).

### IT-US022-007 — `make e2e-seed` is idempotent

- **File under test:** `/Makefile` + seed SQL
- **Boundary:** requires a running compose stack (postgres container)
- **Setup:** `make e2e-up && make e2e-seed`
- **When:** `make e2e-seed` is invoked a second time
- **Then:** exit code is 0; no `ERROR` or `duplicate key` messages in output.
- **Architecture cite:** architecture §6.5 — "Re-running `make e2e-seed` on an already-seeded DB MUST NOT error."
