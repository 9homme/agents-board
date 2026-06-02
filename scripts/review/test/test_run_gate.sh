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
