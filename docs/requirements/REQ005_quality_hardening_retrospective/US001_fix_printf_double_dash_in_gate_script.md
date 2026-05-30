# US001 — Fix `printf "--"` bug in `run-gate.sh` (lines 58 + 71)

**Requirement:** REQ005 — quality hardening retrospective
**Status:** draft

## Story
As a **reviewer running `scripts/review/run-gate.sh` in a non-TTY context** (CI pipe, `bash -c`, orchestrator capture), I want the script to **print the failing-check output verbatim** instead of swallowing it under `printf: --: invalid option`, so that I can actually diagnose a failing gate run without re-running the script under a TTY.

## Acceptance criteria

- **Scenario: TTY run unchanged**
  - Given `scripts/review/run-gate.sh` is invoked with stdout attached to a TTY
  - And a `run_check` step has been forced to fail (e.g. inject a no-op failing command in a test harness)
  - When the script prints the failure block at line 58 (or line 71 for `run_check_warn`)
  - Then the output shows the YELLOW-coloured `--- output (rc=N) ---` banner
  - And the captured `out` string is printed in full beneath it
  - And no `printf: --: invalid option` line appears anywhere in the output

- **Scenario: non-TTY run shows the failure body**
  - Given `scripts/review/run-gate.sh fe` (or any track) is invoked via `bash -c '... | cat'` so that `[ -t 1 ]` is false and `YELLOW=""`
  - And one of the gate's `run_check` steps fails (rc != 0)
  - When the script reaches the failure-output print block
  - Then the line `--- output (rc=N) ---` is printed without colour escapes
  - And the captured stdout/stderr of the failed command is printed in full beneath it
  - And no `printf: --: invalid option` line appears anywhere in the output

- **Scenario: same fix applied to `run_check_warn` at line 71**
  - Given a `run_check_warn` step is non-fatally failing (rc != 0) in non-TTY mode
  - When the warning-output block prints
  - Then the same protection applies — full output visible, no `invalid option` line

- **Scenario: no regression on the PASS path**
  - Given every gate step passes
  - When `scripts/review/run-gate.sh cross` (or `fe` / `be`) is run in either TTY or non-TTY mode
  - Then the final line is `REVIEW GATE: PASS`
  - And exit code is 0
  - And no `--- output` banner is printed (no failures to report)

## UI / UX flow expectations

**No UI:** developer / CI tooling change. The "flow" is: reviewer runs `scripts/review/run-gate.sh <track>` in any shell context (TTY or piped), reads the output to diagnose failures. The fix restores the diagnostic value of that output when piped.

## Out of scope
- Refactoring the colour-detection logic (`[ -t 1 ]`).
- Restructuring the gate script's section / pass / fail helpers.
- Fixing any other gate-script bug not specifically the `printf "--"` issue at lines 58 and 71.
- The FE gate `--forceExit` fix (separate story US002).
- The `gosec` / `govulncheck` missing-tool exit (separate story US003).

## Dependencies
- None. This is a self-contained one-file edit to `scripts/review/run-gate.sh`.

## Notes for the team

- **Concrete fix** per audit §3.3 item 1: change `printf "${YELLOW}--- output (rc=%d) ---${RESET}\n%s\n..." ...` to `printf -- "${YELLOW}--- output (rc=%d) ---${RESET}\n%s\n..." ...` at both line 58 and line 71. The `--` tells `printf` that the next argument is the format string, not an option. Alternative valid fix: prepend a literal space inside the format string. Either is acceptable per AC.
- **Test approach** (tester will detail): a small harness shell script in `scripts/review/test/` (or inline in the spec) that:
  1. Sources `run-gate.sh` functions, OR runs the script with a manufactured failing command,
  2. Captures output via `bash -c '... 2>&1 | cat'` to force non-TTY,
  3. Greps for `--- output (rc=` (must be present) and `printf: --: invalid option` (must NOT be present).
- **Evidence the bug is real** is in the audit appendix (FE gate output line: "Side note: the `printf -- invalid option` line appeared in the warn output").
- This story is **the smallest** in REQ005 (≤1 story point). It's listed first because subsequent stories' test reports will benefit from working failure output.

## Sign-off log
(po-ba appends here on each sign-off pass)
