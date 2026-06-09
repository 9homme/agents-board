# Agents Board

A lightweight project-management platform built for AI agents. Exposes all entities — Projects, Documents, User Stories, and Tasks — as native **Model Context Protocol (MCP)** tools so agents can read and write project state without screen-scraping or custom glue code.

## Features

- **MCP server** — fully compliant SSE/HTTP MCP endpoint; every entity is a tool call
- **REST API** — read endpoints for dashboard and external consumers
- **Hierarchical data model** — Project → Documents / User Stories → Tasks
- **Next.js dashboard** — CSR-only web UI for humans to inspect the same data
- **PostgreSQL backend** — pgx/v5, schema-migrated, fully typed

## Tech stack

| Layer | Technology |
|---|---|
| Backend | Go (Echo), two binaries: `mcp-server` + `api-server` |
| Database | PostgreSQL 15+, pgx/v5 |
| Frontend | Next.js (Pages Router, CSR-only), TypeScript |
| E2E tests | Robot Framework + Browser (Playwright) |

---

## Quick start

### Prerequisites

- Go 1.22+
- Node.js 18+
- PostgreSQL 15+ on `localhost:5432`
- `psql` client tools
- Python 3 (for E2E tests only)

### Local dev (native Postgres)

```bash
createdb agent_board
make dev-migrate        # apply all migrations
make dev-seed           # optional fixtures
make dev-up             # mcp-server :8081 · api-server :8080 · web :3000
make dev-down           # stop all processes
```

Override the connection string:

```bash
DEV_PG_CONN=postgres://user:pass@localhost:5432/mydb make dev-migrate
```

Logs: `mcp-server.log`, `api-server.log`, `web.log` at the repo root.

### Manual startup

**MCP server**
```bash
cd services/agent-board
DATABASE_URL=postgres://localhost/agent_board?sslmode=disable go run cmd/mcp-server/main.go
```

**API server**
```bash
cd services/agent-board
DATABASE_URL=postgres://localhost/agent_board?sslmode=disable \
PORT=8080 \
FRONTEND_URL=http://localhost:3000 \
go run cmd/api-server/main.go
```

**Frontend**
```bash
cd web
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080 npm install && npm run dev
```

---

## Tests

```bash
# Unit / integration
cd services/agent-board && go test ./...

# Frontend components
cd web && npm test -- --watchAll=false

# E2E — requires the full stack (see below)
make e2e
```

---

## E2E stack

Robot Framework suites under `tests/e2e/` run against the real containerised stack. The `Makefile` handles build, health-gated startup, migrations, seed, and test invocation.

### Prerequisites

- **Podman + podman-compose** (`brew install podman podman-compose`) **or** Docker Compose v2
- `psql` client (used by `make e2e-seed` against the compose Postgres on `127.0.0.1:15432`)
- Python 3 with `robotframework` and `robotframework-browser` installed

### Container topology

| Service | Container port | Host port |
|---|---|---|
| `postgres` | 5432 | 15432 |
| `api-server` | 8080 | 8080 |
| `mcp-server` | 8080 | 8081 |
| `web` | 3000 | 3000 |

### Makefile targets

```bash
make e2e                            # full pipeline: up → seed → run → down
make e2e-up                         # start containers (health-gated)
make e2e-seed                       # apply migrations + seed fixtures
make e2e-run                        # run all Robot suites
make e2e-run REQ=REQ001 US=US001    # narrow to one story
make e2e-logs                       # stream container logs
make e2e-down                       # tear down containers + volumes
```

### Host-Postgres alternative

Skip the compose Postgres and point at a local instance:

```bash
createdb agentboard_e2e
PG_CONN='postgres://yourname@localhost:5432/agentboard_e2e?sslmode=disable' make e2e-seed
```

> **Note on `NEXT_PUBLIC_API_BASE_URL`:** Next.js bakes `NEXT_PUBLIC_*` into the client bundle at build time. Changing it requires rebuilding the image: `podman rmi localhost/agents-board_web:latest && make e2e-up`.

---

## Project structure

```
services/agent-board/   ← Go module; cmd/mcp-server + cmd/api-server
web/                    ← Next.js CSR dashboard
tests/e2e/              ← Robot Framework suites
docs/requirements/      ← Per-requirement specs and architecture docs
Makefile                ← All dev + e2e targets
```
