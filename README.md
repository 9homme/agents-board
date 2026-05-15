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
- **Frontend (Component):** `cd web && npm test`
- **E2E (Robot Framework):** `python3 -m robot tests/e2e/`
