# US001 — Fix `printf "--"` bug in `run-gate.sh` (lines 58 + 71)

**Story:** US001 — Fix `printf "--"` bug in `run-gate.sh` (lines 58 + 71)
**Requirement:** REQ005
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Implements:** Scenario: TTY run unchanged, Scenario: non-TTY run shows the failure body, Scenario: same fix applied to `run_check_warn` at line 71, Scenario: no regression on the PASS path
**Blocked by:** none
**Worked-by:** _(none)_

## Goal

Restore diagnostic output of the review gate script when stdout is not a TTY by prefixing the failing-check `printf` format strings at lines 58 and 71 of `scripts/review/run-gate.sh` with `--` so `printf` does not parse the leading `--` colour-reset sequence as an option. After this task, a piped gate run that hits a `run_check` / `run_check_warn` failure prints the full captured output verbatim instead of `printf: --: invalid option`.

## Scope

- **In:** Two-line edit to `scripts/review/run-gate.sh` (lines 58 and 71); add `--` separator argument to the `printf` invocation.
- **Out:** Refactoring the colour-detection logic (`[ -t 1 ]`); restructuring section/pass/fail helpers; any other gate-script bug; FE `--forceExit` (US002); `gosec` / `govulncheck` soft-warn (US003).

## Files touched (estimated, exclusive)

- `scripts/review/run-gate.sh`

Not a scaffold task — no other task in REQ005 touches the same two lines. (US002 and US003 edit different lines of the same file. The orchestrator must serialise the three Group-A tasks anyway since they all touch this one file; declare that via `Blocked by:` only if the orchestrator's 3a tick is co-picking them — but per spec we keep `Blocked by: none` and rely on file-level overlap detection at queue time.)

## Test contract

The dev must make these tests pass (from `US001_be_unit_tests.md`, IDs assigned by tester):
- The non-TTY failing-`run_check` shell harness that greps for `--- output (rc=` (must appear) and `printf: --: invalid option` (must NOT appear).
- The non-TTY failing-`run_check_warn` symmetric harness.
- The PASS-path smoke that asserts `REVIEW GATE: PASS` and exit 0 with no failure banner.

If tester surfaces new test IDs beyond these, the dev writes them and flags the addition back to tester.

## Implementation notes

- The fix per audit §3.3 item 1 and architecture §2 US001 row: at lines 58 and 71, change
  `printf "${YELLOW}--- output (rc=%d) ---${RESET}\n%s\n..." ...`
  to
  `printf -- "${YELLOW}--- output (rc=%d) ---${RESET}\n%s\n..." ...`
- The `--` argument tells `printf` "no more option flags follow." When `YELLOW=""` (non-TTY), the format string begins with literal `--- output` and is otherwise interpreted as an option.
- Alternative valid fix (prepending a literal space inside the format string) is acceptable per AC but the `--` separator is the canonical POSIX form and is recommended.
- No other lines of `run-gate.sh` are touched in this task.
- TDG skill (.claude/skills/tdg/SKILL.md) MUST be invoked at each TDD phase per be-dev workflow. The "red" phase here is the non-TTY harness that reproduces the `invalid option` line on current `main`; the "green" phase is the two `--` insertions; refactor is a no-op for a two-character change.

## Definition of done

- All listed tests green.
- `scripts/review/run-gate.sh be services/agent-board` exits with `REVIEW GATE: PASS` (no new failures introduced by the edit).
- `scripts/review/run-gate.sh cross` exits with `REVIEW GATE: PASS`.
- No new public exports / no Go code touched.
- Code matches architecture §2 US001 row.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Review log

(tech-lead appends here on each review pass)
