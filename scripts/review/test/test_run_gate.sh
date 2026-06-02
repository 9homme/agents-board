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
