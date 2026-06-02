#!/usr/bin/env bash
# scripts/review/test/test_run_gate.sh
# Shell-level test harness for scripts/review/run-gate.sh.
# Tests are named UT-001 through UT-005 (matches US001 task contract).
#
# Run: bash scripts/review/test/test_run_gate.sh
# Exit 0 = all pass, exit 1 = one or more failures.

set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GATE_SCRIPT="$SCRIPT_DIR/../run-gate.sh"
PASS_COUNT=0
FAIL_COUNT=0

pass_test() { echo "  PASS  $1"; PASS_COUNT=$((PASS_COUNT + 1)); }
fail_test() { echo "  FAIL  $1"; echo "        reason: $2"; FAIL_COUNT=$((FAIL_COUNT + 1)); }

# ---------------------------------------------------------------------------
# build_helper
# Writes a temporary script that:
#   1. Extracts the helper-function block from run-gate.sh (up to the first
#      "# ---- BE gate ----" sentinel, which marks the end of the shared
#      output helpers).
#   2. Appends a call to the function named by $1 (the rest are its args).
# The script is run piped through `cat` (non-TTY) so YELLOW/RESET == "".
# ---------------------------------------------------------------------------
build_helper() {
  local tmpfile
  tmpfile=$(mktemp /tmp/gate_helper_XXXXXX.sh)
  # Extract lines before the "# ---- BE gate ----" sentinel (exclusive).
  # Use quit-on-match so the sentinel line itself is not included.
  sed -n '/^# ---- BE gate ----/q; p' "$GATE_SCRIPT" > "$tmpfile"
  cat >> "$tmpfile" <<'TAIL'

# Invoke the requested helper ($1) with remaining args.
fn="$1"; shift
"$fn" "$@"
TAIL
  echo "$tmpfile"
}

# run_capture_nontty <helper-script> <fn> [args...]
# Runs the helper script piped through cat (non-TTY). Output in $OUTPUT.
run_capture_nontty() {
  local helper="$1"; shift
  OUTPUT=$(bash "$helper" "$@" 2>&1 | cat)
}

HELPER_SCRIPT=$(build_helper)
trap 'rm -f "$HELPER_SCRIPT"' EXIT

# ---------------------------------------------------------------------------
# UT-001: run_check failure prints the failure-output banner in non-TTY mode
# ---------------------------------------------------------------------------
echo "UT-001: run_check failure prints '--- output (rc=N) ---' in non-TTY mode"
run_capture_nontty "$HELPER_SCRIPT" run_check failing-step bash -c 'echo some-output; exit 1'
if echo "$OUTPUT" | grep -qe "output (rc="; then
  pass_test "UT-001"
else
  fail_test "UT-001" "banner 'output (rc=' not found. Got: $OUTPUT"
fi

# ---------------------------------------------------------------------------
# UT-002: run_check failure does NOT produce 'printf: --: invalid option'
# ---------------------------------------------------------------------------
echo "UT-002: run_check failure does NOT produce 'invalid option' in non-TTY mode"
run_capture_nontty "$HELPER_SCRIPT" run_check failing-step bash -c 'echo some-output; exit 1'
if echo "$OUTPUT" | grep -qe "invalid option"; then
  fail_test "UT-002" "Found 'invalid option' in output: $OUTPUT"
else
  pass_test "UT-002"
fi

# ---------------------------------------------------------------------------
# UT-003: run_check_warn failure prints the banner in non-TTY mode
# ---------------------------------------------------------------------------
echo "UT-003: run_check_warn failure prints '--- output (rc=N) ---' in non-TTY mode"
run_capture_nontty "$HELPER_SCRIPT" run_check_warn warn-step bash -c 'echo warn-output; exit 1'
if echo "$OUTPUT" | grep -qe "output (rc="; then
  pass_test "UT-003"
else
  fail_test "UT-003" "banner 'output (rc=' not found in warn output. Got: $OUTPUT"
fi

# ---------------------------------------------------------------------------
# UT-004: run_check_warn failure does NOT produce 'printf: --: invalid option'
# ---------------------------------------------------------------------------
echo "UT-004: run_check_warn failure does NOT produce 'invalid option' in non-TTY mode"
run_capture_nontty "$HELPER_SCRIPT" run_check_warn warn-step bash -c 'echo warn-output; exit 1'
if echo "$OUTPUT" | grep -qe "invalid option"; then
  fail_test "UT-004" "Found 'invalid option' in output: $OUTPUT"
else
  pass_test "UT-004"
fi

# ---------------------------------------------------------------------------
# UT-005: PASS path — successful run_check has no banner and no error
# ---------------------------------------------------------------------------
echo "UT-005: run_check pass path produces no '--- output' banner and no error"
run_capture_nontty "$HELPER_SCRIPT" run_check passing-step bash -c 'exit 0'
if echo "$OUTPUT" | grep -qe "output (rc="; then
  fail_test "UT-005" "Unexpected banner on passing step. Got: $OUTPUT"
elif echo "$OUTPUT" | grep -qe "invalid option"; then
  fail_test "UT-005" "Unexpected 'invalid option' on passing step. Got: $OUTPUT"
else
  pass_test "UT-005"
fi

# ---------------------------------------------------------------------------
# IT-US002-001: FE gate npm test invocation contains --watchAll=false AND --forceExit
# (static grep — no invocation required)
# ---------------------------------------------------------------------------
echo "IT-US002-001: run-gate.sh FE npm test line contains --watchAll=false AND --forceExit in that order"
IT_US002_001_LINE=$(grep -n 'run_check.*npm test' "$GATE_SCRIPT" | head -1)
if echo "$IT_US002_001_LINE" | grep -q '\-\-watchAll=false.*\-\-forceExit'; then
  pass_test "IT-US002-001"
else
  fail_test "IT-US002-001" "Expected '--watchAll=false --forceExit' in FE npm test line. Got: $IT_US002_001_LINE"
fi

# ---------------------------------------------------------------------------
# IT-US002-002: skip — requires Node.js/npm invocation; marked as manual verification
# ---------------------------------------------------------------------------
echo "IT-US002-002: SKIP (manual verification — requires live npm/Jest; see test report notes)"

# ---------------------------------------------------------------------------
# IT-US002-003: Only the FE gate line is modified; no other npm test invocations added with --forceExit
# ---------------------------------------------------------------------------
echo "IT-US002-003: Exactly one npm test invocation in run-gate.sh"
NPM_TEST_COUNT=$(grep -c 'npm test' "$GATE_SCRIPT" || true)
if [ "$NPM_TEST_COUNT" -eq 1 ]; then
  pass_test "IT-US002-003"
else
  fail_test "IT-US002-003" "Expected exactly 1 npm test line, found $NPM_TEST_COUNT"
fi

# ---------------------------------------------------------------------------
# IT-US003-001 — BE gate soft-warns when gosec is absent (exit code NOT 2)
# ---------------------------------------------------------------------------
echo "IT-US003-001: BE gate does not exit 2 when gosec is absent; prints WARN gosec skipped"
# Build a PATH that excludes any gosec binary but keeps everything else.
GOSEC_ABSENT_PATH=$(printf '%s' "$PATH" | tr ':' '\n' | grep -v 'gosec' | tr '\n' ':' | sed 's/:$//')
# Create a stub govulncheck so only gosec is missing.
STUB_DIR_001=$(mktemp -d /tmp/gate_stub_001_XXXXXX)
cat > "$STUB_DIR_001/govulncheck" <<'STUB'
#!/usr/bin/env bash
exit 0
STUB
chmod +x "$STUB_DIR_001/govulncheck"
IT_US003_001_OUTPUT=$(PATH="$STUB_DIR_001:$GOSEC_ABSENT_PATH" bash "$GATE_SCRIPT" be services/agent-board 2>&1 | cat)
IT_US003_001_RC=$?
rm -rf "$STUB_DIR_001"
if [ "$IT_US003_001_RC" -eq 2 ]; then
  fail_test "IT-US003-001" "Gate exited 2 (MISSING TOOL) when gosec absent — should soft-warn. Output: $IT_US003_001_OUTPUT"
elif echo "$IT_US003_001_OUTPUT" | grep -qiE 'WARN.*gosec.*(skipped|not installed)'; then
  pass_test "IT-US003-001"
else
  fail_test "IT-US003-001" "Expected WARN gosec skipped line but got: $IT_US003_001_OUTPUT"
fi

# ---------------------------------------------------------------------------
# IT-US003-002 — BE gate soft-warns when govulncheck is absent (exit code NOT 2)
# ---------------------------------------------------------------------------
echo "IT-US003-002: BE gate does not exit 2 when govulncheck is absent; prints WARN govulncheck skipped"
GOVULNCHECK_ABSENT_PATH=$(printf '%s' "$PATH" | tr ':' '\n' | grep -v 'govulncheck' | tr '\n' ':' | sed 's/:$//')
STUB_DIR_002=$(mktemp -d /tmp/gate_stub_002_XXXXXX)
cat > "$STUB_DIR_002/gosec" <<'STUB'
#!/usr/bin/env bash
exit 0
STUB
chmod +x "$STUB_DIR_002/gosec"
IT_US003_002_OUTPUT=$(PATH="$STUB_DIR_002:$GOVULNCHECK_ABSENT_PATH" bash "$GATE_SCRIPT" be services/agent-board 2>&1 | cat)
IT_US003_002_RC=$?
rm -rf "$STUB_DIR_002"
if [ "$IT_US003_002_RC" -eq 2 ]; then
  fail_test "IT-US003-002" "Gate exited 2 (MISSING TOOL) when govulncheck absent — should soft-warn. Output: $IT_US003_002_OUTPUT"
elif echo "$IT_US003_002_OUTPUT" | grep -qiE 'WARN.*govulncheck.*(skipped|not installed)'; then
  if echo "$IT_US003_002_OUTPUT" | grep -q 'golang.org/x/vuln'; then
    pass_test "IT-US003-002"
  else
    fail_test "IT-US003-002" "WARN line found but missing install one-liner (golang.org/x/vuln). Output: $IT_US003_002_OUTPUT"
  fi
else
  fail_test "IT-US003-002" "Expected WARN govulncheck skipped line but got: $IT_US003_002_OUTPUT"
fi

# ---------------------------------------------------------------------------
# IT-US003-003 — When both gosec and govulncheck ARE installed, no WARN line appears
# ---------------------------------------------------------------------------
echo "IT-US003-003: No WARN gosec/govulncheck line appears when both tools are present (stubs)"
STUB_DIR_003=$(mktemp -d /tmp/gate_stub_003_XXXXXX)
for bin in gosec govulncheck; do
  cat > "$STUB_DIR_003/$bin" <<'STUB'
#!/usr/bin/env bash
exit 0
STUB
  chmod +x "$STUB_DIR_003/$bin"
done
IT_US003_003_OUTPUT=$(PATH="$STUB_DIR_003:$PATH" bash "$GATE_SCRIPT" be services/agent-board 2>&1 | cat)
rm -rf "$STUB_DIR_003"
if echo "$IT_US003_003_OUTPUT" | grep -qiE 'WARN.*(gosec|govulncheck)'; then
  fail_test "IT-US003-003" "Unexpected WARN line when both tools present. Output: $IT_US003_003_OUTPUT"
else
  pass_test "IT-US003-003"
fi

# ---------------------------------------------------------------------------
# IT-US003-004 — go and golangci-lint remain hard-required (exit 2)
# ---------------------------------------------------------------------------
echo "IT-US003-004a: BE gate exits 2 (MISSING TOOL) when go is absent"
# Build a PATH that has no 'go' binary — use a tmpdir with only a golangci-lint stub
STUB_DIR_004=$(mktemp -d /tmp/gate_stub_004_XXXXXX)
cat > "$STUB_DIR_004/golangci-lint" <<'STUB'
#!/usr/bin/env bash
exit 0
STUB
chmod +x "$STUB_DIR_004/golangci-lint"
NOBIN_PATH="$STUB_DIR_004:/usr/bin:/bin"
IT_US003_004A_TMPOUT=$(mktemp /tmp/gate_004a_XXXXXX)
PATH="$NOBIN_PATH" bash "$GATE_SCRIPT" be services/agent-board >"$IT_US003_004A_TMPOUT" 2>&1
IT_US003_004A_RC=$?
IT_US003_004A_OUTPUT=$(cat "$IT_US003_004A_TMPOUT")
rm -f "$IT_US003_004A_TMPOUT"
rm -rf "$STUB_DIR_004"
if [ "$IT_US003_004A_RC" -eq 2 ] && echo "$IT_US003_004A_OUTPUT" | grep -q 'MISSING TOOL'; then
  pass_test "IT-US003-004a"
else
  fail_test "IT-US003-004a" "Expected exit 2 + MISSING TOOL when go absent (rc=$IT_US003_004A_RC). Output: $IT_US003_004A_OUTPUT"
fi

echo "IT-US003-004b: BE gate exits 2 (MISSING TOOL) when golangci-lint is absent"
STUB_DIR_004B=$(mktemp -d /tmp/gate_stub_004b_XXXXXX)
cat > "$STUB_DIR_004B/go" <<'STUB'
#!/usr/bin/env bash
exit 0
STUB
chmod +x "$STUB_DIR_004B/go"
NOBIN_PATH_B="$STUB_DIR_004B:/usr/bin:/bin"
IT_US003_004B_TMPOUT=$(mktemp /tmp/gate_004b_XXXXXX)
PATH="$NOBIN_PATH_B" bash "$GATE_SCRIPT" be services/agent-board >"$IT_US003_004B_TMPOUT" 2>&1
IT_US003_004B_RC=$?
IT_US003_004B_OUTPUT=$(cat "$IT_US003_004B_TMPOUT")
rm -f "$IT_US003_004B_TMPOUT"
rm -rf "$STUB_DIR_004B"
if [ "$IT_US003_004B_RC" -eq 2 ] && echo "$IT_US003_004B_OUTPUT" | grep -q 'MISSING TOOL'; then
  pass_test "IT-US003-004b"
else
  fail_test "IT-US003-004b" "Expected exit 2 + MISSING TOOL when golangci-lint absent (rc=$IT_US003_004B_RC). Output: $IT_US003_004B_OUTPUT"
fi

# ---------------------------------------------------------------------------
# IT-US003-005 — Cross and FE gates unaffected; soft-warn logic only in gate_be()
# ---------------------------------------------------------------------------
echo "IT-US003-005: require_tool for semgrep/gitleaks/npm present; no gosec/govulncheck soft-warn outside gate_be"
# Extract gate_cross() and gate_fe() bodies: from "gate_cross()" / "gate_fe()" line to "^}" line
# Strategy: exclude lines inside gate_be() — extract lines in gate_cross and gate_fe
GATE_CROSS_FE_BLOCK=$(awk '/^gate_cross\(\)/,/^}/' "$GATE_SCRIPT"; awk '/^gate_fe\(\)/,/^}/' "$GATE_SCRIPT")
CROSS_FE_REQUIRE_SEMGREP=$(echo "$GATE_CROSS_FE_BLOCK" | grep -c 'require_tool semgrep' || true)
CROSS_FE_REQUIRE_GITLEAKS=$(echo "$GATE_CROSS_FE_BLOCK" | grep -c 'require_tool gitleaks' || true)
CROSS_FE_REQUIRE_NPM=$(echo "$GATE_CROSS_FE_BLOCK" | grep -c 'require_tool npm' || true)
CROSS_FE_GOSEC_WARN=$(echo "$GATE_CROSS_FE_BLOCK" | grep -c 'command -v gosec' || true)
CROSS_FE_GOVULNCHECK_WARN=$(echo "$GATE_CROSS_FE_BLOCK" | grep -c 'command -v govulncheck' || true)

IT_US003_005_FAIL=0
[ "$CROSS_FE_REQUIRE_SEMGREP" -ge 1 ] || { fail_test "IT-US003-005" "require_tool semgrep not found in gate_cross"; IT_US003_005_FAIL=1; }
[ "$CROSS_FE_REQUIRE_GITLEAKS" -ge 1 ] || { fail_test "IT-US003-005" "require_tool gitleaks not found in gate_cross"; IT_US003_005_FAIL=1; }
[ "$CROSS_FE_REQUIRE_NPM" -ge 1 ] || { fail_test "IT-US003-005" "require_tool npm not found in gate_fe"; IT_US003_005_FAIL=1; }
[ "$CROSS_FE_GOSEC_WARN" -eq 0 ] || { fail_test "IT-US003-005" "gosec soft-warn (command -v gosec) found outside gate_be — cross/fe contaminated"; IT_US003_005_FAIL=1; }
[ "$CROSS_FE_GOVULNCHECK_WARN" -eq 0 ] || { fail_test "IT-US003-005" "govulncheck soft-warn (command -v govulncheck) found outside gate_be — cross/fe contaminated"; IT_US003_005_FAIL=1; }
[ "$IT_US003_005_FAIL" -eq 0 ] && pass_test "IT-US003-005"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "Results: $PASS_COUNT passed, $FAIL_COUNT failed"
if [ $FAIL_COUNT -eq 0 ]; then
  echo "ALL TESTS PASSED"
  exit 0
else
  echo "TESTS FAILED"
  exit 1
fi
