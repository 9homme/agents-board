# US003/be_github_actions_gate

**Requirement:** REQ007
**Story:** US003
**Track:** BE
**Service:** services/agent-board
**Status:** in_review
**Blocked by:** US002_be_makefile_healthcheck.md
**Worked-by:** be-dev-2026-06-08T10:00:00Z-b4e2
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

## Coverage exemption
`.github/workflows/e2e.yml` is a YAML CI configuration file — not a Go source file — and is therefore excluded from `go tool cover` measurement. The only Go file added is `services/agent-board/internal/workflow/workflow_test.go` (a test file with no production counterpart), which exercises the workflow YAML's structure. This is the correct split: the test validates the deliverable, and there is no production `.go` file to measure.

## Notes

### Files touched
- `.github/workflows/e2e.yml` (new) — GitHub Actions workflow: `pull_request` → `main` trigger, `ubuntu-latest`, `make e2e-up` / `make e2e-seed` / `make e2e-run`, upload-artifact with `if: always()`, `make e2e-down` with `if: always()`.
- `services/agent-board/internal/workflow/workflow_test.go` (new) — 5 Go tests (UT-US003-001 through UT-US003-005) validating the workflow YAML structure.

### Tests added
- 5 unit tests in `agent-board/internal/workflow` package: trigger, no-push, make targets, always() conditions, ubuntu-latest runner.
- `go test ./...` in services/agent-board: 312 passed (307 pre-existing + 5 new).

### E2E results (US003-tagged, 3 runs)
Run 1: 1 test, 1 passed, 0 failed
Run 2: 1 test, 1 passed, 0 failed
Run 3: 1 test, 1 passed, 0 failed

Full suite note: `make e2e-run` on this worktree shows 30 tests, 24 passed, 6 failed. The 6 failures are in FE tests for US004/US005 (UserStoriesTab, UserStoryDrawer not yet implemented — parallel FE tasks) and 1 pre-existing missing resource file in `REQ005/US006_rapid_navigation.robot`. None of the 6 failures are caused by US003 changes. The US003-tagged test `E2E-US003-001 GitHub Actions workflow is correctly configured` passes in all 3 runs.

### Review gate
- `scripts/review/run-gate.sh be services/agent-board`: REVIEW GATE: PASS
- `scripts/review/run-gate.sh cross`: REVIEW GATE: PASS
- `robot --dryrun tests/e2e/REQ007_*/`: 7 tests, 7 passed, 0 failed

## Review log
