# US017 — Soft-warn `gosec` / `govulncheck` when missing instead of hard-exit

**Requirement:** REQ005 — quality hardening retrospective
**Status:** draft

## Story
As a **reviewer on any developer machine that does not have the standalone `gosec` and `govulncheck` binaries installed**, I want `scripts/review/run-gate.sh be <service>` to **complete with a clear WARN line and a `PASS` / `FAIL` verdict** instead of `exit 2` MISSING TOOL, so that BE gate runs are not blocked by an environmental gap that is already covered by `golangci-lint`'s `gosec` linter.

## Acceptance criteria

- **Scenario: BE gate completes when `gosec` is absent**
  - Given a dev environment where `command -v gosec` returns non-zero (binary not installed)
  - And `golangci-lint` IS installed (precondition unchanged)
  - When `scripts/review/run-gate.sh be services/agent-board` is invoked
  - Then the script does NOT `exit 2` on the missing-tool check
  - And the gate output contains a line of the form `WARN  gosec (skipped — not installed; coverage via golangci-lint gosec linter)` (or close equivalent)
  - And the gate continues running the remaining checks (`gofmt`, `go vet`, `golangci-lint`, `go test`, `govulncheck`)
  - And the final line is either `REVIEW GATE: PASS` or `REVIEW GATE: FAIL (N check(s))` based on the other checks' outcomes — never `exit 2`

- **Scenario: BE gate completes when `govulncheck` is absent**
  - Given a dev environment where `command -v govulncheck` returns non-zero
  - When `scripts/review/run-gate.sh be services/agent-board` is invoked
  - Then the script does NOT `exit 2`
  - And the gate output contains a line of the form `WARN  govulncheck (skipped — not installed; install with: go install golang.org/x/vuln/cmd/govulncheck@latest)`
  - And the gate continues with all other checks
  - And the final exit is 0 or 1, not 2

- **Scenario: when both binaries ARE installed, behaviour unchanged**
  - Given both `gosec` and `govulncheck` ARE on PATH
  - When `scripts/review/run-gate.sh be services/agent-board` is invoked
  - Then the script runs `gosec -quiet -severity=medium ./...` and `govulncheck ./...` exactly as today
  - And reports PASS or FAIL based on their actual output
  - And no skipped/WARN line appears for those two checks

- **Scenario: `go` and `golangci-lint` remain hard-required**
  - Given `go` OR `golangci-lint` is not installed
  - When `scripts/review/run-gate.sh be services/agent-board` is invoked
  - Then the script exits 2 with the existing MISSING TOOL message (these two are not soft-warnable — the BE gate is meaningless without them)

- **Scenario: cross-cutting and FE gates unaffected**
  - Given the soft-warn change applies only to the BE gate's `gosec` and `govulncheck` `require_tool` calls
  - When `scripts/review/run-gate.sh cross` or `scripts/review/run-gate.sh fe` is invoked
  - Then their `require_tool` invocations (semgrep, gitleaks, npm) behave exactly as today
  - And exit-2 semantics are preserved for those tracks

- **Scenario: README documents the substitution**
  - Given `scripts/review/README.md` exists
  - When a new developer reads it after this story ships
  - Then the README explains: (a) which tools are hard-required vs soft-warn, (b) for soft-warn tools, what is the substitute coverage (specifically that `golangci-lint` v2 enables `gosec` as a linter so the standalone is additive, not load-bearing), (c) one-liner install commands for both `gosec` and `govulncheck` for anyone who wants to opt in.

## UI / UX flow expectations

**No UI:** developer tooling change. Flow: reviewer runs `scripts/review/run-gate.sh be services/<svc>`, gets a useful verdict on every machine without per-host install ceremony.

## Out of scope
- Installing `gosec` or `govulncheck` in CI or as a `Makefile` target. (The audit explicitly says "pick one" — install OR soft-warn. This story picks soft-warn per D-002.)
- Changing `.golangci.yml` to add or remove linters (gosec linter remains enabled — that's the substitute coverage).
- The `printf "--"` and `--forceExit` fixes (separate stories US015, US016).
- Audit's own `REVIEW GATE: PASS` printout formatting.

## Dependencies
- None. Self-contained edit to `scripts/review/run-gate.sh` + `scripts/review/README.md`.

## Notes for the team

- **Concrete implementation hint** per audit §3.3 item 3: in `gate_be()`, replace `require_tool gosec ...` and `require_tool govulncheck ...` with a `command -v` test that either (a) skips the corresponding `run_check` call and prints a WARN line, or (b) routes to a new helper `run_check_warn_if_missing` that does both. Either shape is fine; AC pins the observable behaviour, not the helper name.
- **Why not install:** the `gosec` ruleset is already running inside `golangci-lint` v2 (see `services/agent-board/.golangci.yml` — `gosec` is enabled). The standalone run is therefore additive coverage, not unique coverage. `govulncheck` IS unique coverage (vuln-DB lookup), but blocking the entire BE gate on its absence on a dev laptop is hostile. Soft-warn lets the user opt in.
- **Watch out for:** the existing `require_tool` helper exits 2; the new path must NOT call `require_tool` for these two binaries. Either inline the check or add a sibling helper.

## Sign-off log
(po-ba appends here on each sign-off pass)
