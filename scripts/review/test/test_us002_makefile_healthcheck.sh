#!/usr/bin/env bash
# test_us002_makefile_healthcheck.sh
# Shell-level test suite for US002 Makefile changes.
#
# Assertions:
#   UT-US002-001  e2e-up mcp-server curl carries --max-time 5
#   UT-US002-002  e2e-seed dry-run contains no migration step
#
# Usage: bash scripts/review/test/test_us002_makefile_healthcheck.sh
# Must be run from the repository root.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
MAKEFILE="${REPO_ROOT}/Makefile"

pass=0
fail=0

run_test() {
    local id="$1"
    local description="$2"
    local result="$3"   # "pass" or "fail"
    if [ "$result" = "pass" ]; then
        echo "PASS  ${id}: ${description}"
        pass=$((pass + 1))
    else
        echo "FAIL  ${id}: ${description}"
        fail=$((fail + 1))
    fi
}

# UT-US002-001: e2e-up mcp-server curl must carry --max-time 5
mcp_curl_bounded=""
if grep -qE 'curl -sf --max-time 5 http://localhost:8081/sse' "${MAKEFILE}"; then
    mcp_curl_bounded="pass"
else
    mcp_curl_bounded="fail"
fi
run_test "UT-US002-001" "e2e-up mcp-server curl has --max-time 5" "${mcp_curl_bounded}"

# UT-US002-002: e2e-seed dry-run must NOT contain a migration step
seed_dry_run="$(make -n e2e-seed 2>/dev/null)"
if echo "${seed_dry_run}" | grep -q "applying migration"; then
    seed_no_migration="fail"
else
    seed_no_migration="pass"
fi
run_test "UT-US002-002" "e2e-seed dry-run does not contain migration step" "${seed_no_migration}"

# UT-US002-003: e2e-seed dry-run still contains psql (data seeding intact)
if echo "${seed_dry_run}" | grep -q "psql"; then
    seed_has_psql="pass"
else
    seed_has_psql="fail"
fi
run_test "UT-US002-003" "e2e-seed dry-run still contains psql for data seeding" "${seed_has_psql}"

echo ""
echo "Results: ${pass} passed, ${fail} failed"

if [ "${fail}" -gt 0 ]; then
    exit 1
fi
