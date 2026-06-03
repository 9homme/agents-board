# Agents Board

A minimal Jira/Confluence platform designed specifically for AI agents, using the **Model Context Protocol (MCP)**.

## Project Overview

Agents Board allows AI agents to interact with project management entities (Projects, Documents, User Stories, and Tasks) as native tools. This enables seamless automation and integration into AI-driven workflows.

### Core Features
- **MCP Integration:** Fully compliant MCP server over SSE/HTTP.
- **Hierarchical Data Model:**
  - **Project:** The root container.
  - **Documents:** Knowledge base articles and docs (under Project).
  - **User Stories:** Agile requirements (under Project).
  - **Tasks:** Actionable work items (under User Story).
- **Persistence:** Robust PostgreSQL backend.
- **Tech Stack:** Go (Echo Framework), PostgreSQL (pgx/v5), Robot Framework (E2E).

---

## Multi-Agent Engineering Workflow

This project is developed using a unique **Multi-Agent Engineering Team** approach. A team of specialized AI subagents handles the entire software development lifecycle, from requirement analysis to final sign-off.

### The Team

| Agent | Role | Responsibility |
| :--- | :--- | :--- |
| **po-ba** | Product Owner / BA | Decomposes requirements into INVEST user stories and performs final sign-off. |
| **system-architect** | Architect | Designs the system topology and locks the API/JSON contracts. |
| **tech-lead** | Tech Lead | Decomposes stories into technical tasks and performs strict code reviews. |
| **tester** | QA Engineer | Designs the test pyramid and implements E2E Robot Framework tests. |
| **be-dev** | Backend Dev | Implements Go microservices using TDD (Test-Driven Development). |
| **fe-dev** | Frontend Dev | Implements Next.js CSR frontends using TDD. |

### Development Phases

The project progresses through three distinct, gated phases:

1.  **Phase 1: Discovery & Design**
    - `po-ba` clarifies requirements.
    - `system-architect` drafts `architecture.md`.
    - **Human Gate:** The user must formally approve the architecture to proceed.
2.  **Phase 2: Planning & Testing**
    - `tech-lead` breaks stories into parallelizable BE and FE tasks.
    - `tester` generates unit specs and E2E test scripts.
3.  **Phase 3: Implementation & TDD**
    - `be-dev` and `fe-dev` work in parallel tracks.
    - `tech-lead` reviews every task; implementation must meet the "Definition of Done".
    - `tester` runs E2E validation.
    - `po-ba` signs off on completed stories.

---

## Getting Started

### Prerequisites
- **Go:** 1.22+
- **Postgres:** 15+
- **Python 3:** (for Robot Framework tests)

### Running the Backend
The backend is a unified Go module in `services/agent-board` that produces two separate binaries.

**1. MCP Server (AI Interface)**
```bash
cd services/agent-board
export DB_URL=postgres://localhost/agent_board?sslmode=disable
go run cmd/mcp-server/main.go
```

**2. API Server (Web Dashboard Interface)**
```bash
cd services/agent-board
export DATABASE_URL=postgres://localhost/agent_board?sslmode=disable
export PORT=8080
export FRONTEND_URL=http://localhost:3000
go run cmd/api-server/main.go
```

### Running the Frontend
The dashboard is a Next.js (CSR-only) application.
```bash
cd web
export NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
npm install
npm run dev
```

### Running Tests
- **Backend (Unit/Integration):** `cd services/agent-board && go test ./...`
- **Frontend (Component):** `cd web && npm test -- --watchAll=false`
- **E2E (Robot Framework):** see [Running the live E2E stack](#running-the-live-e2e-stack) below — Robot suites require a running stack and a seeded database.

---

## Running the live E2E stack

The end-to-end Robot Framework suites under `tests/e2e/` exercise the **real stack** (Postgres + api-server + mcp-server + web) over HTTP/SSE. Use the `Makefile` at the repo root — it handles container build, healthcheck-gated startup, migration + seed application, and Robot invocation in one consistent command surface.

### Prerequisites
- **Docker Compose v2** (`docker compose`) **OR** **podman-compose** (`brew install podman-compose`). The Makefile auto-detects which is installed.
- **psql** (PostgreSQL client) for migration + seed application against the compose Postgres on `127.0.0.1:15432`.
- **Python 3** with Robot Framework + Browser library installed.

### Stack topology
| Service | Container port | Host port | Notes |
|---|---|---|---|
| `postgres` | 5432 | 15432 | Healthcheck-gated; volume `postgres-data`. |
| `api-server` | 8080 | 8080 | Read API (4 `GET` endpoints). Distroless runtime (no shell). |
| `mcp-server` | 8080 | 8081 | MCP write API over SSE/`/message?sessionId=...`. **Required for all writes** — the api-server is read-only. |
| `web` | 3000 | 3000 | Next.js production build (`next start`). |

### Environment variables baked into the compose stack
| Variable | Used by | Value |
|---|---|---|
| `DATABASE_URL` | api-server | `postgres://agent_board:agent_board@postgres:5432/agent_board?sslmode=disable` |
| `DB_URL` | mcp-server | same DSN (note: different name — pre-existing inconsistency between the two binaries) |
| `PORT` | api-server / mcp-server / web | `8080` (containers); host-mapped per the table above |
| `FRONTEND_URL` | api-server (CORS allowlist) | `http://localhost:3000` |
| `NEXT_PUBLIC_API_BASE_URL` | web | `http://localhost:8080` — **set as a build ARG** in `web/Dockerfile`, not at runtime. Next.js bakes `NEXT_PUBLIC_*` into the client JS bundle at `next build` time, so the compose `environment:` block has no effect on it. Change requires `podman rmi localhost/agents-board_web:latest && make e2e-up`. |

### One-command pipeline
```bash
make e2e          # up → seed → run → ALWAYS down (trap on failure, even on test fail)
```

### Step-by-step targets
```bash
make e2e-up       # docker compose / podman-compose up -d + curl-poll for :8080, :3000, :8081/sse healthy
make e2e-seed     # psql -f migrations + tests/e2e/data/seeds/*.sql (alphabetical, idempotent)
make e2e-run      # robot tests/e2e/REQ*/ → tests/e2e/results/{output.xml,log.html,report.html}
make e2e-run REQ=REQ005 US=US008    # narrow scope
make e2e-logs     # stream container logs
make e2e-down     # docker compose / podman-compose down -v (removes volumes)
```

### Host-Postgres alternative (if you already have Postgres running locally)
Override `PG_CONN` to point at your local DB and skip the compose Postgres container. You still need api-server + mcp-server + web running somewhere reachable (locally or in containers).

```bash
# Example: your local Postgres on :5432, dedicated DB named agentboard_e2e
createdb agentboard_e2e
PG_CONN='postgres://yourname@localhost:5432/agentboard_e2e?sslmode=disable' make e2e-seed
```

### Validation discipline (REQ005 / US008 mandate)
- **be-dev / fe-dev** must run `make e2e-run` (and paste the `N tests, N passed, 0 failed` line into the task's `## Notes`) **before** flipping the task to `Status: in_review`. Unit / component / react-doctor green is not a substitute.
- **tech-lead** must run `make e2e-run` **three consecutive times** during review; all three must be `0 failed`. Any flake is `changes_requested`. Three summary lines must be pasted verbatim into the `### Review pass N` log entry.

See `.claude/agents/{be-dev,fe-dev,tech-lead}.md` and `tests/e2e/README.md` for the full discipline.
