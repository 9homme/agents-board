# End-to-End Test Runbook

This runbook covers running the Robot Framework e2e test suites against the local
Docker/Podman stack. Architecture reference: REQ005 §6.

## `dev-*` vs `e2e-*` — which family to use?

| Family | Targets | Postgres | Services | Use case |
|---|---|---|---|---|
| `dev-*` | `make dev-up` / `make dev-down` / `make dev-migrate` / `make dev-seed` | Native `:5432` | Go binaries + Next.js run as background processes (no Docker) | Daily local development; fast iteration |
| `e2e-*` | `make e2e-up` / `make e2e-down` / `make e2e-seed` / `make e2e-run` | Docker compose `:15432` | All services in containers (healthcheck-gated) | Robot Framework end-to-end test runs; CI |

Both families accept env-var overrides: `DEV_PG_CONN=...` for `dev-*`; `PG_CONN=...` for `e2e-*`.

---

## Prerequisites

### Container runtime

Either `docker compose` (Docker Desktop / Docker Engine ≥ 20) or `podman-compose` must
be on your PATH. The `Makefile` auto-detects which one is available:

```bash
# Docker (recommended)
docker --version  # must be present

# OR Podman
podman-compose --version
```

### Postgres client

`psql` (PostgreSQL client tools) must be on your PATH for `make e2e-seed`:

```bash
# macOS
brew install libpq && brew link --force libpq

# Ubuntu/Debian
apt-get install -y postgresql-client
```

### Robot Framework and libraries

Install on the host (Robot runs on the host, NOT in a container):

```bash
pip install robotframework robotframework-requests
# For Browser library (Playwright-based UI tests):
pip install robotframework-browser
rfbrowser init
```

---

## Makefile target reference

All targets are run from the repo root.

| Target | What it does |
|---|---|
| `make e2e-up` | Starts `postgres` + `api-server` + `web` via compose. Waits for healthchecks. |
| `make e2e-down` | Stops and removes containers **and** the `postgres-data` volume. |
| `make e2e-seed` | Applies `services/agent-board/migrations/*.up.sql` (lex order), then `tests/e2e/data/seeds/*.sql` (lex order). Idempotent. |
| `make e2e-run` | Runs all Robot suites under `tests/e2e/REQ*/`. Results in `tests/e2e/results/`. |
| `make e2e-run REQ=REQ001` | Narrows run to `tests/e2e/REQ001_*/`. |
| `make e2e-run REQ=REQ001 US=US001` | Further narrows to suites tagged `US001` inside REQ001 directories. |
| `make e2e` | Full pipeline: `e2e-up` → `e2e-seed` → `e2e-run` → `e2e-down` (always, even on failure). |
| `make e2e-logs` | Streams container logs (tail=100) for debugging. |

### Compose tool override

The Makefile auto-detects `docker compose` vs `podman-compose`. To override explicitly:

```bash
make COMPOSE="podman-compose" e2e-up
```

---

## Host-Postgres path (skip the postgres container)

If you already have Postgres running locally on port 5432, you can point the seed
and migration commands at it directly:

```bash
E2E_DATABASE_URL=postgres://myuser:mypassword@localhost:5432/agentboard_e2e
# Apply migrations manually:
for f in services/agent-board/migrations/*.up.sql; do
  psql "$E2E_DATABASE_URL" -v ON_ERROR_STOP=1 -f "$f"
done
# Apply seeds:
for f in tests/e2e/data/seeds/*.sql; do
  psql "$E2E_DATABASE_URL" -v ON_ERROR_STOP=1 -f "$f"
done
```

Then start `api-server` and `web` as usual (e.g. via `make dev-up`).

---

## Adding a new seed fixture

1. Create `tests/e2e/data/seeds/REQ###_<short_name>.sql` (zero-padded REQ number, snake_case).
2. Use `INSERT ... ON CONFLICT DO NOTHING` for every statement — seeds **must** be idempotent.
3. Use deterministic (hard-coded) UUIDs for primary keys so Robot suites can reference them by ID.
4. Run `make e2e-seed` twice in a row and verify the second run exits 0 with no errors.

**Example:**

```sql
-- REQ001_checkout.sql
INSERT INTO projects (id, name, ...)
VALUES ('11111111-0000-0000-0000-000000000001', 'Checkout Project', ...)
ON CONFLICT DO NOTHING;
```

Files are loaded in alphabetical order (`REQ000_*` before `REQ001_*`), which matches
REQ dependency order.

---

## Debugging a failing Robot run

1. **Stream live container logs:**
   ```bash
   make e2e-logs
   ```

2. **Inspect Robot artefacts** — after any run, results are in `tests/e2e/results/`:
   - `tests/e2e/results/log.html` — full step-by-step execution log with screenshots (Browser tests)
   - `tests/e2e/results/report.html` — summary pass/fail report
   - `tests/e2e/results/output.xml` — machine-readable output (used by orchestrator)

3. **Run a single suite manually:**
   ```bash
   robot --outputdir tests/e2e/results tests/e2e/REQ001_agent_board_mcp/
   ```

4. **Re-seed without restart** (if data is stale):
   ```bash
   make e2e-seed
   ```
   Seeds are idempotent — safe to re-run on a running stack.

5. **Full reset:**
   ```bash
   make e2e-down && make e2e-up && make e2e-seed
   ```

---

## Orchestrator responsibility — Phase 3c

After all BE and FE tasks for a story are `completed`, the orchestrator runs:

```bash
make e2e-up && make e2e-seed
make e2e-run REQ=REQ001   # replace REQ001 with the target REQ
make e2e-down
```

Or the single-command equivalent:

```bash
make e2e REQ=REQ001
```

The orchestrator captures `tests/e2e/results/output.xml` (and `log.html`/`report.html`),
maps Robot test outcomes to E2E-* IDs from `US[ID]_e2e_tests.md`, and includes the
summary table in the per-story `US[ID]_test_report.md`.

Skipped tests must be called out explicitly with a one-line justification in the test
report.
