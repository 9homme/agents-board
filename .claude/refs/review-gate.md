# Shared ref — quality gate runbook

How to run the mandatory review gate and what its verdict means. Consumed by **be-dev** / **fe-dev** (who run it once before hand-off) and **tech-lead-reviewer** (Mode 1 verifies the dev's pasted output; Mode 2 re-runs the whole gate on the integrated working branch).

The gate is `scripts/review/run-gate.sh`. It bundles linters, security scanners, dependency-vuln checks, and CSR-only enforcement on top of the unit tests.

## Commands

| Track | Command |
|---|---|
| BE | `scripts/review/run-gate.sh be services/<service-name>` |
| FE | `scripts/review/run-gate.sh fe` |
| cross (always) | `scripts/review/run-gate.sh cross` |

- **BE** runs: `gofmt -s` (no diff), `go vet ./...`, `golangci-lint` (staticcheck/errcheck/unused/ineffassign/gocritic/revive/errorlint/bodyclose/sqlclosecheck), `go test ./...`, plus soft-warn `gosec` + `govulncheck`.
- **FE** runs (inside `web/`): `npm run typecheck`, `npm run lint --max-warnings=0` (with `eslint-plugin-security`), `npm test`, soft-warn `npm audit`, plus CSR-only scan (no `getServerSideProps`/`getStaticProps`/`getInitialProps`, no `web/pages/api/`) and the `fetch()`-boundary scan (all backend calls via `web/lib/api/`).
- **cross** runs: `semgrep` (OWASP top-10 + golang + typescript + react packs) and `gitleaks` (no secrets).

Exit codes: `0` all pass · `1` one or more checks failed · `2` bad invocation / missing required tool.

## The PASS rule

**The gate MUST emit `REVIEW GATE: PASS` on stdout (exit 0).** Both the per-track gate AND `... cross` must PASS.

**NO SUBSTITUTIONS.** Pasting `npm test`, `go test`, `npm run lint`, `golangci-lint`, etc. individually does NOT replace the gate's `REVIEW GATE: PASS` line. The gate exists precisely to bundle the right checks with the right enforcement; running them piecemeal lets the runner cherry-pick which checks to honor — exactly the drift this rule prevents. If the gate is broken, fix the gate via the gate-fix track — never approve / hand off around it.

## Coverage gate (per track)

Also run:
- BE: `cd services/<svc> && go test -coverprofile=/tmp/cov.out ./... && go tool cover -func=/tmp/cov.out`
- FE: `cd web && npm test -- --coverage --watchAll=false --forceExit`

Every production file in the task's `## Files touched` (not `*_test.go` / `*.test.tsx`) MUST be ≥ 80% line coverage, UNLESS the task has a `## Coverage exemption` section justifying each below-threshold file. Quote the per-file numbers verbatim.

## Robot e2e parse check (cross addendum)

If any `tests/e2e/REQ[ID]_*/` directory exists for the REQ, also run `robot --dryrun tests/e2e/REQ[ID]_*/` from the repo root. A `--dryrun` failure is a **spec defect** → `SPEC_GAP_FOUND` routed to tester, NOT a dev `changes_requested`.

## Live e2e + flake verification (REQ Quality Gate only — Mode 2)

**Live e2e does NOT run per-task.** It runs ONCE at the REQ Quality Gate (tech-lead-reviewer Mode 2), when all tasks across all stories are `completed` and the full stack is integrated. Running it per-task is impossible — a BE task cannot satisfy FE-dependent e2e tests and vice versa.

Bring the stack up with `make e2e-up && make e2e-seed`, then run `make e2e-run` **THREE consecutive times — all three 100% green, no failures, no flakes.** A single failure in any of the 3 runs is a FLAKE and is disqualifying. Paste all three `N tests, N passed, 0 failed` summary lines verbatim. Container runtime is **Podman** — always go through the Makefile targets, never call `docker`. If the e2e stack itself can't run, that's `blocked_review_gate`, not a pass.

## blocked_review_gate semantics

If the gate, coverage tooling, `robot --dryrun`, or the e2e stack **could not run cleanly through to a clear PASS/FAIL** — exit 2, hang, missing tool, missing binary, script defect, stack won't come up — the artifact is the gate/tooling, not the code:
- A **dev** sets the task `Status: blocked_review_gate` and reports the exact failure mode (never `in_review`).
- A **reviewer** sets `Status: blocked_review_gate` and reports `REVIEW_GATE_BLOCKED` (never `approved`, never `changes_requested` — the code is not at fault when the gate itself is broken).
- The orchestrator routes `blocked_review_gate` to the gate-fix track, never to a dev.

A gate `FAIL` line where a **check legitimately failed and the code is at fault** is `changes_requested`, not `blocked_review_gate`. Decide which by reading the failed check's actual output, not by rationalising.
