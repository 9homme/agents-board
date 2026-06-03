# Makefile — repo root
# e2e harness targets for the local Docker/Podman stack.
# Architecture ref: REQ005 §6.3.
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
PG_CONN        := postgres://agent_board:agent_board@localhost:15432/agent_board?sslmode=disable

.PHONY: e2e-up e2e-down e2e-seed e2e-run e2e e2e-logs _check-compose

# Internal guard — fails at recipe time when neither docker nor podman is installed.
# Does NOT fire on dry-run (`make -n`) so static assertions still work.
_check-compose:
	@command -v docker >/dev/null 2>&1 || command -v podman-compose >/dev/null 2>&1 || \
	  { echo "ERROR: Neither 'docker compose' nor 'podman-compose' is available on PATH. Install one first." >&2; exit 1; }

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
	@echo "-> stack is healthy: api-server :8080, web :3000"

e2e-down: _check-compose       ## Stop and remove containers + volumes.
	$(COMPOSE) down -v

e2e-seed:                      ## Apply migrations then seed fixtures (idempotent).
	@for f in $$(ls $(MIGRATIONS_DIR)/*.up.sql | sort); do \
	  echo "-> applying migration $$f"; \
	  psql "$(PG_CONN)" -v ON_ERROR_STOP=1 -f $$f; \
	done
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
