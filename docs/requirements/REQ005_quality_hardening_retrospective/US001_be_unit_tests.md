# US001 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). This story touches only `scripts/review/run-gate.sh` — a shell script, not a Go package. Tests are therefore shell-harness tests (bash scripts) that invoke the gate script in a controlled environment and assert on stdout/exit-code. They live in `scripts/review/test/` (create the directory if missing). No Go test files are produced for this story.

## Coverage matrix

| AC scenario | Layer | Test ID | Package / script | Function or behaviour under test |
|---|---|---|---|---|
| Non-TTY failing `run_check` prints failure body, no `invalid option` | integration | IT-US001-001 | `scripts/review/run-gate.sh` | `run_check` failure output in non-TTY mode |
| Non-TTY failing `run_check_warn` prints warning body, no `invalid option` | integration | IT-US001-002 | `scripts/review/run-gate.sh` | `run_check_warn` failure output in non-TTY mode |
| PASS path: no failure banner, exit 0 | integration | IT-US001-003 | `scripts/review/run-gate.sh` | full gate PASS path smoke |
| TTY run prints YELLOW banner (no regression) | integration | IT-US001-004 | `scripts/review/run-gate.sh` | `run_check` failure output in TTY-like mode |

## Integration tests

These are shell-harness integration tests, not Go integration tests. The "red" phase reproduces the `printf: --: invalid option` error on an unpatched script; the "green" phase passes on the patched script.

### IT-US001-001 — Non-TTY failing `run_check` shows failure body without `invalid option`

- **Script under test:** `scripts/review/run-gate.sh`
- **Boundary:** shell invocation, non-TTY stdout
- **Setup:**
  - Create a minimal harness file `scripts/review/test/test_printf_fix.sh`.
  - Source or invoke the gate script with a deliberately failing `run_check` step (e.g. inject a `run_check "always-fails" false` call into the `gate_cross` section, or invoke a minimal wrapper script that defines a fake `gate_cross` calling `run_check "test-step" bash -c 'echo failing output; exit 1'`).
  - Force non-TTY mode by invoking via `bash -c '... 2>&1 | cat'` so `[ -t 1 ]` evaluates to false and `YELLOW=""`.
- **When:** harness invokes `bash scripts/review/run-gate.sh cross 2>&1 | cat` (or the minimal wrapper equivalent)
- **Then:**
  - Captured output contains the substring `--- output (rc=1) ---` (the failure banner).
  - Captured output does NOT contain the substring `printf: --: invalid option`.
  - Captured output contains the text `failing output` (the command's captured stdout is visible beneath the banner).
- **Edge cases:** also assert that the format string parsing produces the rc value correctly (e.g. `rc=1` not `rc=--`).
- **Architecture cite:** architecture §2 US001 row — `printf` at lines 58 and 71 of `run-gate.sh` must be prefixed with `--`.

### IT-US001-002 — Non-TTY failing `run_check_warn` shows warning body without `invalid option`

- **Script under test:** `scripts/review/run-gate.sh`
- **Boundary:** shell invocation, non-TTY stdout
- **Setup:**
  - Same harness file `scripts/review/test/test_printf_fix.sh`.
  - Inject a failing `run_check_warn "warn-step" bash -c 'echo warn output; exit 1'` call and force non-TTY via `| cat`.
- **When:** harness invokes the gate script invoking the `run_check_warn` path.
- **Then:**
  - Captured output contains `--- output (rc=1) ---` (the warn-output banner).
  - Captured output does NOT contain `printf: --: invalid option`.
  - Captured output contains `warn output`.
  - Exit code of the gate script overall is still 0 (warn does not fail the gate).
- **Architecture cite:** architecture §2 US001 row — line 71 (`run_check_warn` output print block) receives the same `--` fix.

### IT-US001-003 — PASS path smoke: no failure banner, exit 0

- **Script under test:** `scripts/review/run-gate.sh`
- **Boundary:** full gate invocation in non-TTY mode with all steps passing
- **Setup:**
  - Invoke `bash scripts/review/run-gate.sh cross 2>&1 | cat` from the repo root (or `be services/agent-board` if the BE environment is available; prefer `cross` for lighter dependencies).
  - Alternatively, inject all-pass stubs for any step that requires external tools not guaranteed in the test environment.
- **When:** gate completes with all steps returning exit 0
- **Then:**
  - Final line of output is `REVIEW GATE: PASS`.
  - Exit code is 0.
  - No `--- output (rc=` substring appears anywhere in the output (no failure banner).
  - No `printf: --: invalid option` line appears.
- **Architecture cite:** US001 AC "Scenario: no regression on the PASS path".

### IT-US001-004 — TTY-like mode: YELLOW banner still rendered (no regression)

- **Script under test:** `scripts/review/run-gate.sh`
- **Boundary:** shell invocation simulating TTY detection
- **Setup:**
  - Invoke the script with `YELLOW="\033[0;33m"` and `RESET="\033[0m"` pre-set via environment override (or use `script -q /dev/null bash -c '...'` to attach a pseudo-TTY if the test environment supports it).
  - Inject a failing step as in IT-US001-001.
- **When:** the failure output block executes with `YELLOW` non-empty
- **Then:**
  - Output contains the ANSI colour-escape-prefixed banner (or at minimum `--- output (rc=` is present without the `invalid option` error — the exact colour rendering depends on the terminal emulator, so substring match on the non-escape portion is sufficient).
  - No `printf: --: invalid option` line appears.
- **Notes:** If the test environment cannot produce a pseudo-TTY, this test may be marked as a manual verification step in the test report; that is acceptable provided IT-US001-001 and IT-US001-002 are automated.
- **Architecture cite:** US001 AC "Scenario: TTY run unchanged".
