#!/usr/bin/env bash
# scripts/review/test/test_us008_e2e_harness.sh
#
# Shell-level integration test harness for US008 — Live e2e stack-up.
# Asserts structural and dry-run properties of the files introduced by US008.
# Tests are static (no containers required) unless otherwise noted.
#
# Test IDs: IT-US008-001 through IT-US008-007
#   IT-US008-001: `make -n e2e-up`  dry-run contains compose up command
#   IT-US008-002: `make -n e2e-down` dry-run contains compose down command
#   IT-US008-003: `make -n e2e-seed` dry-run lists psql invocations
#   IT-US008-004: seed SQL file exists and contains ON CONFLICT DO NOTHING
#   IT-US008-005: migration SQL files exist and are non-empty
#   IT-US008-006: docker-compose.yml exists and has correct port bindings
#   IT-US008-007: .gitignore contains tests/e2e/results/ entry
#
# Additional static checks (no IT-US008-* ID but required by architecture §6):
#   - services/agent-board/Dockerfile exists
#   - web/Dockerfile exists
#   - .dockerignore exists
#   - tests/e2e/README.md exists with required sections
#
# Run: bash scripts/review/test/test_us008_e2e_harness.sh [--repo-root <path>]
# Exit 0 = all pass, exit 1 = one or more failures.

set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$SCRIPT_DIR/../../.."

# Allow override for CI or manual testing
while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo-root) REPO_ROOT="$2"; shift 2 ;;
    *) echo "Unknown argument: $1"; exit 1 ;;
  esac
done

REPO_ROOT="$(cd "$REPO_ROOT" && pwd)"

PASS_COUNT=0
FAIL_COUNT=0

pass_test() { echo "  PASS  $1"; PASS_COUNT=$((PASS_COUNT + 1)); }
fail_test() { echo "  FAIL  $1"; echo "        reason: $2"; FAIL_COUNT=$((FAIL_COUNT + 1)); }

MAKEFILE="$REPO_ROOT/Makefile"
COMPOSE_FILE="$REPO_ROOT/docker-compose.yml"
DOCKERIGNORE="$REPO_ROOT/.dockerignore"
GITIGNORE="$REPO_ROOT/.gitignore"
SEED_FILE="$REPO_ROOT/tests/e2e/data/seeds/REQ000_baseline.sql"
SEEDS_DIR="$REPO_ROOT/tests/e2e/data/seeds"
MIGRATIONS_DIR="$REPO_ROOT/services/agent-board/migrations"
BE_DOCKERFILE="$REPO_ROOT/services/agent-board/Dockerfile"
WEB_DOCKERFILE="$REPO_ROOT/web/Dockerfile"
E2E_README="$REPO_ROOT/tests/e2e/README.md"

# ---------------------------------------------------------------------------
# IT-US008-001: `make -n e2e-up` dry-run outputs compose up command
# ---------------------------------------------------------------------------
echo "IT-US008-001: make -n e2e-up dry-run contains docker compose up -d --wait"
if [ ! -f "$MAKEFILE" ]; then
  fail_test "IT-US008-001" "Makefile does not exist at $MAKEFILE"
else
  DRY_RUN_UP=$(make -C "$REPO_ROOT" -n e2e-up 2>&1) || true
  if echo "$DRY_RUN_UP" | grep -qE '(docker compose|docker-compose|podman-compose)\s+up\s+(-d\s+--wait|--wait\s+-d)'; then
    pass_test "IT-US008-001"
  else
    fail_test "IT-US008-001" "Expected 'docker compose up -d --wait' (or equivalent) in dry-run. Got: $DRY_RUN_UP"
  fi
fi

# ---------------------------------------------------------------------------
# IT-US008-002: `make -n e2e-down` dry-run outputs compose down command
# ---------------------------------------------------------------------------
echo "IT-US008-002: make -n e2e-down dry-run contains docker compose down -v"
if [ ! -f "$MAKEFILE" ]; then
  fail_test "IT-US008-002" "Makefile does not exist at $MAKEFILE"
else
  DRY_RUN_DOWN=$(make -C "$REPO_ROOT" -n e2e-down 2>&1) || true
  if echo "$DRY_RUN_DOWN" | grep -qE '(docker compose|docker-compose|podman-compose)\s+down\s+-v'; then
    pass_test "IT-US008-002"
  else
    fail_test "IT-US008-002" "Expected 'docker compose down -v' (or equivalent) in dry-run. Got: $DRY_RUN_DOWN"
  fi
fi

# ---------------------------------------------------------------------------
# IT-US008-003: `make -n e2e-seed` dry-run shows psql invocations for migrations
# ---------------------------------------------------------------------------
echo "IT-US008-003: make -n e2e-seed dry-run references .up.sql migration files via psql"
if [ ! -f "$MAKEFILE" ]; then
  fail_test "IT-US008-003" "Makefile does not exist at $MAKEFILE"
else
  DRY_RUN_SEED=$(make -C "$REPO_ROOT" -n e2e-seed 2>&1) || true
  # Must reference psql AND migration directory (and seeds dir if .sql files exist there)
  IT_003_FAIL=0
  echo "$DRY_RUN_SEED" | grep -qE 'psql' || { fail_test "IT-US008-003" "psql not found in dry-run. Got: $DRY_RUN_SEED"; IT_003_FAIL=1; }
  echo "$DRY_RUN_SEED" | grep -qE '(migrations|\.up\.sql)' || { fail_test "IT-US008-003" "migration path not found in dry-run. Got: $DRY_RUN_SEED"; IT_003_FAIL=1; }
  # If seeds dir exists and has .sql files, it must also appear in the output
  if [ -d "$SEEDS_DIR" ] && find "$SEEDS_DIR" -name '*.sql' -maxdepth 1 | grep -q .; then
    echo "$DRY_RUN_SEED" | grep -qE '(seeds|\.sql)' || { fail_test "IT-US008-003" "seed dir not found in dry-run even though seeds exist. Got: $DRY_RUN_SEED"; IT_003_FAIL=1; }
  fi
  [ "$IT_003_FAIL" -eq 0 ] && pass_test "IT-US008-003"
fi

# ---------------------------------------------------------------------------
# IT-US008-004: Seed SQL file exists and contains idempotency guard
# ---------------------------------------------------------------------------
echo "IT-US008-004: REQ000_baseline.sql exists and contains ON CONFLICT DO NOTHING"
if [ ! -f "$SEED_FILE" ]; then
  fail_test "IT-US008-004" "Seed file does not exist: $SEED_FILE"
else
  if grep -qi 'ON CONFLICT DO NOTHING' "$SEED_FILE"; then
    pass_test "IT-US008-004"
  else
    fail_test "IT-US008-004" "Seed file missing 'ON CONFLICT DO NOTHING' idempotency guard. File: $SEED_FILE"
  fi
fi

# ---------------------------------------------------------------------------
# IT-US008-005: Migration SQL files exist and are non-empty
# ---------------------------------------------------------------------------
echo "IT-US008-005: services/agent-board/migrations/*.up.sql files exist and are non-empty"
if [ ! -d "$MIGRATIONS_DIR" ]; then
  fail_test "IT-US008-005" "Migrations directory does not exist: $MIGRATIONS_DIR"
else
  UP_SQL_COUNT=$(find "$MIGRATIONS_DIR" -name '*.up.sql' -size +0c | wc -l | tr -d ' ')
  if [ "$UP_SQL_COUNT" -ge 1 ]; then
    pass_test "IT-US008-005"
  else
    fail_test "IT-US008-005" "No non-empty .up.sql files found in $MIGRATIONS_DIR"
  fi
fi

# ---------------------------------------------------------------------------
# IT-US008-006: docker-compose.yml exists and has correct port bindings per architecture §6.2
# ---------------------------------------------------------------------------
echo "IT-US008-006: docker-compose.yml exists and binds ports on 127.0.0.1 as per architecture §6.2"
if [ ! -f "$COMPOSE_FILE" ]; then
  fail_test "IT-US008-006" "docker-compose.yml does not exist at $COMPOSE_FILE"
else
  IT_US008_006_FAIL=0
  # Postgres must bind on 127.0.0.1:15432
  grep -q '127.0.0.1:15432' "$COMPOSE_FILE" || {
    fail_test "IT-US008-006" "docker-compose.yml missing '127.0.0.1:15432' (postgres port binding)"
    IT_US008_006_FAIL=1
  }
  # api-server must bind on 127.0.0.1:8080
  grep -q '127.0.0.1:8080' "$COMPOSE_FILE" || {
    fail_test "IT-US008-006" "docker-compose.yml missing '127.0.0.1:8080' (api-server port binding)"
    IT_US008_006_FAIL=1
  }
  # web must bind on 127.0.0.1:3000
  grep -q '127.0.0.1:3000' "$COMPOSE_FILE" || {
    fail_test "IT-US008-006" "docker-compose.yml missing '127.0.0.1:3000' (web port binding)"
    IT_US008_006_FAIL=1
  }
  [ "$IT_US008_006_FAIL" -eq 0 ] && pass_test "IT-US008-006"
fi

# ---------------------------------------------------------------------------
# IT-US008-007: .gitignore contains tests/e2e/results/ entry
# ---------------------------------------------------------------------------
echo "IT-US008-007: .gitignore contains tests/e2e/results/ entry"
if [ ! -f "$GITIGNORE" ]; then
  fail_test "IT-US008-007" ".gitignore does not exist at $GITIGNORE"
else
  if grep -qF 'tests/e2e/results/' "$GITIGNORE"; then
    pass_test "IT-US008-007"
  else
    fail_test "IT-US008-007" ".gitignore does not contain 'tests/e2e/results/' entry"
  fi
fi

# ---------------------------------------------------------------------------
# Additional static checks (required by architecture §6 but not assigned IT IDs in spec)
# ---------------------------------------------------------------------------

echo "STATIC-001: services/agent-board/Dockerfile exists"
if [ -f "$BE_DOCKERFILE" ]; then
  pass_test "STATIC-001 (be Dockerfile)"
else
  fail_test "STATIC-001 (be Dockerfile)" "Missing $BE_DOCKERFILE"
fi

echo "STATIC-002: web/Dockerfile exists"
if [ -f "$WEB_DOCKERFILE" ]; then
  pass_test "STATIC-002 (web Dockerfile)"
else
  fail_test "STATIC-002 (web Dockerfile)" "Missing $WEB_DOCKERFILE"
fi

echo "STATIC-003: .dockerignore exists"
if [ -f "$DOCKERIGNORE" ]; then
  pass_test "STATIC-003 (.dockerignore)"
else
  fail_test "STATIC-003 (.dockerignore)" "Missing $DOCKERIGNORE"
fi

echo "STATIC-004: tests/e2e/README.md exists with required runbook sections"
if [ ! -f "$E2E_README" ]; then
  fail_test "STATIC-004 (e2e README)" "Missing $E2E_README"
else
  IT_README_FAIL=0
  # Section (a): prerequisites
  grep -qi 'prerequisite\|pip install\|robotframework' "$E2E_README" || {
    fail_test "STATIC-004 (e2e README)" "README missing prerequisites / pip install section"
    IT_README_FAIL=1
  }
  # Section (b): make targets
  grep -qE 'e2e-up|e2e-down|e2e-seed|e2e-run|e2e-logs' "$E2E_README" || {
    fail_test "STATIC-004 (e2e README)" "README missing Makefile target descriptions"
    IT_README_FAIL=1
  }
  # Section (c): how to add seeds
  grep -qi 'seed\|REQ' "$E2E_README" || {
    fail_test "STATIC-004 (e2e README)" "README missing seed-adding guide"
    IT_README_FAIL=1
  }
  # Section (d): debug guide
  grep -qi 'debug\|e2e-logs\|results' "$E2E_README" || {
    fail_test "STATIC-004 (e2e README)" "README missing debug guide section"
    IT_README_FAIL=1
  }
  # Section (e): orchestrator responsibility / Phase 3c
  grep -qi 'orchestrator\|Phase 3c\|test report' "$E2E_README" || {
    fail_test "STATIC-004 (e2e README)" "README missing orchestrator / Phase 3c responsibility section"
    IT_README_FAIL=1
  }
  # Podman compatibility mention
  grep -qi 'podman\|docker compose' "$E2E_README" || {
    fail_test "STATIC-004 (e2e README)" "README missing Docker/Podman compatibility note"
    IT_README_FAIL=1
  }
  [ "$IT_README_FAIL" -eq 0 ] && pass_test "STATIC-004 (e2e README)"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "Results: $PASS_COUNT passed, $FAIL_COUNT failed"
if [ "$FAIL_COUNT" -eq 0 ]; then
  echo "ALL TESTS PASSED"
  exit 0
else
  echo "TESTS FAILED"
  exit 1
fi
