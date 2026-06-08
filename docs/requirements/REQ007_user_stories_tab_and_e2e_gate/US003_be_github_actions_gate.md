# US003/be_github_actions_gate

**Requirement:** REQ007
**Story:** US003
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** US002_be_makefile_healthcheck.md
**Worked-by:** 
**Implements:** US003, D-003, D-004

## Goal
Add a GitHub Actions workflow that runs the e2e Robot test suite on pull requests to the `main` branch to block merges on failure.

## Scope
- **In:** Creating `.github/workflows/e2e.yml`.
- **Out:** Branch protection settings (manual human action). Push to main triggers.

## Files touched (estimated, exclusive)
- `.github/workflows/e2e.yml`

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US003_be_unit_tests.md`: All applicable UT/IT tests.

## Implementation notes
- Trigger: `pull_request` on `branches: [main]`.
- Uses `ubuntu-latest`.
- Needs `actions/checkout@v4`, python setup (for Robot), running `make e2e-up`, `make e2e-seed`, and `make e2e-run`.
- Includes `actions/upload-artifact@v4` with `if: always()` for results (`tests/e2e/results/*`).
- Teardown: `make e2e-down` with `if: always()`.

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- (Track: BE) `go test -coverprofile=/tmp/cov.out ./... && go tool cover -func=/tmp/cov.out` — every production `.go` file in this task's `## Files touched` clears ≥ 80% line coverage, OR the task has a written `## Coverage exemption` section justifying each below-threshold file.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0 AND emits `REVIEW GATE: PASS` on stdout. Also `scripts/review/run-gate.sh cross` exits 0 AND emits `REVIEW GATE: PASS`. If the REQ has Robot e2e suites, `robot --dryrun tests/e2e/REQ007_*/` also passes.
- Dev set status to `in_review` and reported back.

## Review log
