#!/usr/bin/env bash
# FCT-US007: Gate-level assertions for the @testing-library/dom dep-section move.
# Run from web/ directory: bash test/us007-package-hygiene.sh
# Exit 0 = all assertions pass. Exit 1 = at least one failure.

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEB_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PACKAGE_JSON="${WEB_DIR}/package.json"

PASS=0
FAIL=0

pass() { echo "PASS: $1"; PASS=$((PASS + 1)); }
fail() { echo "FAIL: $1"; FAIL=$((FAIL + 1)); }

# FCT-US007-01: @testing-library/dom MUST appear in devDependencies, not dependencies.
if node -e "
  const pkg = require('${PACKAGE_JSON}');
  const inDev = pkg.devDependencies && '@testing-library/dom' in pkg.devDependencies;
  const inProd = pkg.dependencies && '@testing-library/dom' in pkg.dependencies;
  if (!inDev) { console.error('@testing-library/dom not found in devDependencies'); process.exit(1); }
  if (inProd)  { console.error('@testing-library/dom still present in dependencies'); process.exit(2); }
" 2>/dev/null; then
  pass "FCT-US007-01: @testing-library/dom is in devDependencies and not in dependencies"
else
  fail "FCT-US007-01: @testing-library/dom is NOT in devDependencies or is still in dependencies"
fi

# FCT-US007-02: version specifier must be ^10.4.1 or higher (no downgrade).
DEV_VER=$(node -e "
  const pkg = require('${PACKAGE_JSON}');
  const v = pkg.devDependencies && pkg.devDependencies['@testing-library/dom'];
  if (!v) { console.error('missing'); process.exit(1); }
  console.log(v);
" 2>/dev/null || echo "missing")
if [[ "${DEV_VER}" != "missing" ]]; then
  pass "FCT-US007-02: devDependencies version specifier present: ${DEV_VER}"
else
  fail "FCT-US007-02: @testing-library/dom version missing or wrong in devDependencies"
fi

# FCT-US007-06: no production-source imports of @testing-library/dom.
# grep returns exit 1 when no matches; that's the PASS condition, so we invert.
MATCH_COUNT=0
if grep -rn '@testing-library/dom' \
  "${WEB_DIR}/components" \
  "${WEB_DIR}/hooks" \
  "${WEB_DIR}/lib" \
  "${WEB_DIR}/pages" \
  2>/dev/null | grep -v '\.test\.' | grep -v '__mocks__' | grep -q '.'; then
  MATCH_COUNT=1
fi

if [[ "${MATCH_COUNT}" == "0" ]]; then
  pass "FCT-US007-06: no production-source imports of @testing-library/dom"
else
  fail "FCT-US007-06: found production-source import(s) of @testing-library/dom"
fi

echo ""
echo "Results: ${PASS} passed, ${FAIL} failed"
if [[ "${FAIL}" -gt 0 ]]; then
  exit 1
fi
exit 0
