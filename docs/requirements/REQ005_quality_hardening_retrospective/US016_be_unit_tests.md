# US016 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). This story edits a single line of `scripts/review/run-gate.sh` — no Go packages are touched. Tests are shell-harness tests (bash scripts) in `scripts/review/test/` asserting the flag is present and the gate terminates.

## Coverage matrix

| AC scenario | Layer | Test ID | Script | Behaviour under test |
|---|---|---|---|---|
| Gate invocation string contains `--forceExit` after `--watchAll=false` | integration | IT-US016-001 | `scripts/review/run-gate.sh` | FE Jest invocation includes `--forceExit` |
| Failing FE test reports FAIL and does not hang | integration | IT-US016-002 | `scripts/review/run-gate.sh` | gate exits non-zero with Jest failure; no timeout hang |
| No other Jest-invoking script is silently modified | integration | IT-US016-003 | `scripts/review/run-gate.sh` | scope is limited to the FE-gate line |

## Integration tests

### IT-US016-001 — Gate invocation contains `--forceExit`

- **Script under test:** `scripts/review/run-gate.sh`
- **Boundary:** static content of the script file
- **Setup:** Read the file; no invocation needed for this assertion.
- **When:** `grep` for the FE gate's `npm test` invocation line
- **Then:**
  - The line matching `run_check "npm test` contains both `--watchAll=false` AND `--forceExit` as arguments to the `npm test` command (order: `--watchAll=false --forceExit`).
  - Specifically: `grep -n 'run_check.*npm test' scripts/review/run-gate.sh` produces a line containing `--forceExit`.
  - The `--forceExit` flag appears AFTER `--watchAll=false` in the same invocation (audit §3.3 item 2 exact pair).
- **Architecture cite:** architecture §2 US016 row — edit line 116 of `run-gate.sh` to append `--forceExit`.

### IT-US016-002 — Gate terminates within timeout on FE Jest completion

- **Script under test:** `scripts/review/run-gate.sh`
- **Boundary:** full `fe` gate invocation (requires `cd web && npm install` to be runnable in the test environment)
- **Setup:**
  - Invoke `timeout 120 bash scripts/review/run-gate.sh fe 2>&1 | cat` from the repo root.
  - The `web/` directory must have a valid `package.json` and `node_modules/` (or `npm ci` must succeed first).
- **When:** the FE gate runs Jest with `--watchAll=false --forceExit`
- **Then:**
  - The command returns within 120 seconds (the `timeout` wrapper exits 0 or 1, never 124 which would indicate timeout expiry).
  - The final output line is either `REVIEW GATE: PASS` (all tests green) or `REVIEW GATE: FAIL (N check(s))` (some test fails) — not a hung process.
  - Exit code is 0 (pass) or 1 (test failure) — never 124 (timeout/hang).
- **Notes:** This test is the definitive proof that the flag fixes the MSW/mermaid handle-leak hang. If the test environment does not have Node.js, mark as a manual verification step in the test report; that is acceptable provided IT-US016-001 is automated.
- **Architecture cite:** US016 AC "Scenario: FE gate terminates cleanly with `--forceExit`".

### IT-US016-003 — Only the FE gate line is modified; no other Jest invocations changed

- **Script under test:** `scripts/review/run-gate.sh`
- **Boundary:** static content of the script file
- **Setup:** Read the file.
- **When:** count all `npm test` invocations in the script
- **Then:**
  - Exactly one line in the script invokes `npm test` (the FE gate line at the original line 116).
  - No other line that previously did NOT contain `--forceExit` now contains it.
- **Architecture cite:** US016 AC "Scenario: no other Jest invocation regressed".
