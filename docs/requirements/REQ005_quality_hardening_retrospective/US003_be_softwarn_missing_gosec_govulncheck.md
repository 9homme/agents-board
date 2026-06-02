# US003 — Soft-warn `gosec` / `govulncheck` when missing instead of hard-exit

**Story:** US003 — Soft-warn `gosec` / `govulncheck` when missing instead of hard-exit
**Requirement:** REQ005
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Implements:** Scenario: BE gate completes when `gosec` is absent, Scenario: BE gate completes when `govulncheck` is absent, Scenario: when both binaries ARE installed, behaviour unchanged, Scenario: `go` and `golangci-lint` remain hard-required, Scenario: cross-cutting and FE gates unaffected, Scenario: README documents the substitution
**Blocked by:** none
**Worked-by:** _(none)_

## Goal

Make the BE gate degrade gracefully when `gosec` or `govulncheck` is not installed locally — print a WARN line and skip that check rather than `exit 2` MISSING TOOL. `go` and `golangci-lint` remain hard-required. README is updated to explain the soft-warn vs hard-required split and provide install one-liners. After this task, BE gate runs are not blocked by an environmental gap that is already covered by `golangci-lint`'s built-in `gosec` linter.

## Scope

- **In:** Edits to `scripts/review/run-gate.sh` `gate_be()` — replace the `require_tool gosec ...` and `require_tool govulncheck ...` calls with inline `command -v` checks per architecture §2 US003 row and D-011. Add a "Soft-warn vs. hard-required" subsection to `scripts/review/README.md` listing hard tools (go, golangci-lint, npm, semgrep, gitleaks) vs soft-warn tools (gosec, govulncheck) plus install one-liners.
- **Out:** Installing `gosec` or `govulncheck` in CI or a Makefile target; changing `.golangci.yml`; the `printf` fix (US001); the `--forceExit` fix (US002); refactoring `require_tool` for cross or FE tracks.

## Files touched (estimated, exclusive)

- `scripts/review/run-gate.sh`
- `scripts/review/README.md`

Same `run-gate.sh` as US001 and US002 — see note in US002 about queue-time serialisation. README.md is unique to this task.

## Test contract

The dev must make these tests pass (from `US003_be_unit_tests.md`, IDs assigned by tester):
- Harness that hides `gosec` from PATH (e.g. `PATH=/usr/bin` or a tmpdir-only PATH) and asserts the script does NOT `exit 2`, prints a WARN line for `gosec`, and continues to run remaining checks.
- Symmetric harness for `govulncheck`.
- Harness that hides `go` (or `golangci-lint`) and asserts the script DOES `exit 2` with the existing MISSING TOOL message (hard-required preserved).
- Harness that confirms cross and FE gates are unaffected (still exit 2 if their hard-required tools are missing).
- README check: a grep for the new "Soft-warn vs. hard-required" subsection that names both soft tools and provides the install one-liners.

If tester surfaces new test IDs beyond these, the dev writes them and flags the addition back to tester.

## Implementation notes

- Per architecture §9 D-011: inline the `command -v` check in `gate_be()` before each affected `run_check`. Do NOT add a new helper `run_check_warn_if_missing`. Two call sites only.
- Pseudocode shape:
  ```bash
  if ! command -v gosec >/dev/null 2>&1; then
    printf -- "${YELLOW}WARN${RESET}  gosec (skipped — not installed; coverage via golangci-lint gosec linter)\n"
  else
    run_check "gosec" gosec -quiet -severity=medium ./...
  fi
  ```
  (Note the `printf --` form — US001's fix.)
- Use the same WARN prefix shape that `run_check_warn` already emits so output looks consistent.
- README addition lives under the existing "What it runs" section. Install one-liners: `go install github.com/securego/gosec/v2/cmd/gosec@latest` and `go install golang.org/x/vuln/cmd/govulncheck@latest`.
- TDG skill (.claude/skills/tdg/SKILL.md) MUST be invoked at each TDD phase per be-dev workflow. The "red" phase is the missing-tool harness reproducing `exit 2` on current `main`; the "green" phase is the inline `command -v` edits + README subsection; refactor is minimal.

## Definition of done

- All listed tests green.
- `scripts/review/run-gate.sh be services/agent-board` exits with `REVIEW GATE: PASS` when both binaries ARE installed; exits with 0 or 1 (never 2) and prints the WARN line when either is missing.
- `scripts/review/run-gate.sh cross` and `fe` are unaffected — verify by running both on a machine missing `gosec` / `govulncheck` and confirming the cross / fe verdicts are unchanged.
- README section is present and accurate.
- Code matches architecture §2 US003 row and §9 D-011.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Review log

(tech-lead appends here on each review pass)
