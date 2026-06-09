# Makefile — repo root
# Two target families:
#   dev-*   — native local-dev stack (no Docker; native Postgres on :5432)  [REQ006/US015]
#   e2e-*   — Docker/Podman compose stack for Robot Framework e2e tests      [REQ005/US008]
# Architecture refs: REQ005 §6.3 (e2e family); REQ006/US015 D-013/D-014 (dev family).
#
# Podman compatibility: prefers 'docker compose' when docker is present;
# falls back to 'podman-compose' when only Podman is available.
# Override with: make COMPOSE="podman-compose" e2e-up
#
# If neither is installed, the COMPOSE variable defaults to 'docker compose'
# for dry-run legibility; the _check-compose guard errors at runtime before
# any actual container invocation.
COMPOSE ?= $(shell command -v docker >/dev/null 2>&1 && echo "docker compose" || (command -v podman-compose >/dev/null 2>&1 && echo "podman-compose" || echo "docker compose"))

SEEDS_DIR      := tests/e2e/data/seeds
MIGRATIONS_DIR := services/agent-board/migrations
PG_CONN ?= postgres://agent_board:agent_board@localhost:15432/agent_board?sslmode=disable
DEV_PG_CONN ?= postgres://agent_board:agent_board@localhost:5432/agent_board?sslmode=disable

# Dev env-var defaults (native processes; all use ?= so the operator can override).
API_PORT              ?= 8080
MCP_PORT              ?= 8081
FRONTEND_URL          ?= http://localhost:3000
NEXT_PUBLIC_API_BASE_URL ?= http://localhost:$(API_PORT)

.PHONY: e2e-up e2e-build e2e-down e2e-seed e2e-run e2e e2e-logs _check-compose \
        dev-up dev-down dev-migrate dev-seed

# Internal guard — fails at recipe time when neither docker nor podman is installed.
# Does NOT fire on dry-run (`make -n`) so static assertions still work.
_check-compose:
	@command -v docker >/dev/null 2>&1 || command -v podman-compose >/dev/null 2>&1 || \
	  { echo "ERROR: Neither 'docker compose' nor 'podman-compose' is available on PATH. Install one first." >&2; exit 1; }

e2e-build: _check-compose      ## Rebuild container images from source (needed after code changes).
	$(COMPOSE) build

e2e-up: _check-compose         ## Start postgres + api-server + web (compose, healthcheck-gated).
	$(COMPOSE) up -d
	@echo "-> waiting for stack to become healthy (timeout 120s)..."
	@i=0; until curl -sf http://localhost:8080/api/v1/projects >/dev/null 2>&1; do \
	  i=$$((i+1)); \
	  if [ $$i -ge 60 ]; then echo "ERROR: api-server failed to become healthy" >&2; exit 1; fi; \
	  sleep 2; \
	done
	@i=0; until curl -sf http://localhost:3000/ >/dev/null 2>&1; do \
	  i=$$((i+1)); \
	  if [ $$i -ge 60 ]; then echo "ERROR: web failed to become healthy" >&2; exit 1; fi; \
	  sleep 2; \
	done
	@i=0; until curl -sf --max-time 5 http://localhost:8081/sse >/dev/null 2>&1 || curl -s --max-time 5 -o /dev/null -w "%{http_code}" http://localhost:8081/sse 2>&1 | grep -qE "^(200|405|404)$$"; do \
	  i=$$((i+1)); \
	  if [ $$i -ge 60 ]; then echo "ERROR: mcp-server failed to become healthy" >&2; exit 1; fi; \
	  sleep 2; \
	done
	@echo "-> stack is healthy: api-server :8080, web :3000, mcp-server :8081"

e2e-down: _check-compose       ## Stop and remove containers + volumes.
	$(COMPOSE) down -v

e2e-seed:                      ## Seed test-data fixtures only (migrations run at api-server startup — US001).
	@for f in $$(ls $(SEEDS_DIR)/*.sql 2>/dev/null | sort); do \
	  echo "-> applying seed $$f"; \
	  psql "$(PG_CONN)" -v ON_ERROR_STOP=1 -f $$f; \
	done

e2e-run:                       ## Run Robot suites. REQ=REQ001 US=US001 narrows scope.
	@mkdir -p tests/e2e/results
	@INCLUDE_FLAGS=""; \
	if [ -n "$(US)" ]; then INCLUDE_FLAGS="--include $(US)"; fi; \
	if [ -n "$(REQ)" ]; then \
	  robot --outputdir tests/e2e/results $$INCLUDE_FLAGS tests/e2e/$(REQ)_*/; \
	else \
	  robot --outputdir tests/e2e/results $$INCLUDE_FLAGS tests/e2e/REQ*/; \
	fi

e2e: e2e-up e2e-seed           ## Full pipeline: up -> seed -> run -> ALWAYS down (trap on failure).
	@set -e; \
	trap '$(MAKE) e2e-down' EXIT; \
	$(MAKE) e2e-run

e2e-logs: _check-compose       ## Stream container logs (tail=100).
	$(COMPOSE) logs -f --tail=100

# ---------------------------------------------------------------------------
# dev-* family — native local-dev stack (no Docker; native Postgres on :5432)
# ---------------------------------------------------------------------------

dev-up:                        ## Start mcp-server, api-server, web as background native processes.
	@command -v go >/dev/null 2>&1 || { echo "ERROR: go not in PATH — install Go and retry" >&2; exit 1; }
	@command -v node >/dev/null 2>&1 || { echo "ERROR: node not in PATH — install Node.js and retry" >&2; exit 1; }
	@command -v npm >/dev/null 2>&1 || { echo "ERROR: npm not in PATH — install npm and retry" >&2; exit 1; }
	@echo "[dev-up] starting mcp-server on port $(MCP_PORT)..."
	@cd services/agent-board && \
	  DATABASE_URL=$(DEV_PG_CONN) PORT=$(MCP_PORT) go run cmd/mcp-server/main.go \
	    > ../../mcp-server.log 2>&1 & \
	  echo $$! > ../../.mcp.pid
	@echo "[dev-up] mcp-server started (pid $$(cat .mcp.pid), log mcp-server.log)"
	@echo "[dev-up] starting api-server on port $(API_PORT)..."
	@cd services/agent-board && \
	  DATABASE_URL=$(DEV_PG_CONN) PORT=$(API_PORT) FRONTEND_URL=$(FRONTEND_URL) \
	    go run cmd/api-server/main.go \
	    > ../../api-server.log 2>&1 & \
	  echo $$! > ../../.api.pid
	@echo "[dev-up] api-server started (pid $$(cat .api.pid), log api-server.log)"
	@echo "[dev-up] starting web on port 3000..."
	@cd web && \
	  PORT=3000 NEXT_PUBLIC_API_BASE_URL=$(NEXT_PUBLIC_API_BASE_URL) \
	    npm run dev \
	    > ../web.log 2>&1 & \
	  echo $$! > ../.web.pid
	@echo "[dev-up] web started (pid $$(cat .web.pid), log web.log)"
	@echo "[dev-up] all services initiated."
	@echo "   - mcp-server:  http://localhost:$(MCP_PORT)/sse"
	@echo "   - api-server:  http://localhost:$(API_PORT)/api/v1/projects" # nosemgrep: generic.secrets.security.detected-username-and-password-in-uri
	@echo "   - web:         http://localhost:3000"
	@echo "Logs: mcp-server.log, api-server.log, web.log"
	@echo "Stop with: make dev-down"

dev-down:                      ## Stop background native processes (idempotent).
	@for pid_file in .mcp.pid .api.pid .web.pid; do \
	  if [ -f "$$pid_file" ]; then \
	    PID=$$(cat "$$pid_file"); \
	    if ps -p $$PID > /dev/null 2>&1; then \
	      pkill -P $$PID 2>/dev/null || true; \
	      kill $$PID 2>/dev/null || true; \
	      sleep 1; \
	      kill -0 $$PID 2>/dev/null && kill -9 $$PID 2>/dev/null || true; \
	      echo "[dev-down] terminated process $$PID (from $$pid_file)"; \
	    fi; \
	    rm -f "$$pid_file"; \
	  fi; \
	done
	@for port in 8080 8081 3000; do \
	  PIDS=$$(lsof -ti :$$port 2>/dev/null || true); \
	  if [ -n "$$PIDS" ]; then \
	    for pid in $$PIDS; do \
	      kill -9 $$pid 2>/dev/null || true; \
	      echo "[dev-down] force killed PID $$pid on port $$port"; \
	    done; \
	  fi; \
	done
	@echo "[dev-down] done."

dev-migrate:                   ## Apply all migrations against DEV_PG_CONN (native :5432).
	@command -v psql >/dev/null 2>&1 || { echo "ERROR: psql not in PATH — install postgres client and retry" >&2; exit 1; }
	@for f in $$(ls $(MIGRATIONS_DIR)/*.up.sql | sort); do \
	  echo "[dev-migrate] applying $$f"; \
	  psql "$(DEV_PG_CONN)" -v ON_ERROR_STOP=1 -f $$f; \
	done
	@echo "[dev-migrate] done."

dev-seed:                      ## Apply seed fixtures against DEV_PG_CONN (native :5432).
	@command -v psql >/dev/null 2>&1 || { echo "ERROR: psql not in PATH — install postgres client and retry" >&2; exit 1; }
	@for f in $$(ls $(SEEDS_DIR)/*.sql 2>/dev/null | sort); do \
	  echo "[dev-seed] applying $$f"; \
	  psql "$(DEV_PG_CONN)" -v ON_ERROR_STOP=1 -f $$f; \
	done
	@echo "[dev-seed] done."
