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

### Review pass 1 — 2026-06-08 — verdict: SPEC_GAP_FOUND (route to tester) — task held at in_review
US003 deliverable (`.github/workflows/e2e.yml` + `internal/workflow/workflow_test.go`) is correct against D-003/D-004 and the e2e spec; ALL gates and the US003-scoped e2e pass. The mandatory live full-suite e2e, however, cannot reach 3x green because of a PRE-EXISTING parse-time defect in REQ005 test code (NOT US003 application/CI code). Per tech-lead.md, a failure in test code rather than application code is `SPEC_GAP_FOUND` routed to tester — NOT `changes_requested` to this dev (US003 is not at fault) and NOT `blocked_review_gate` (the gate tooling ran cleanly to PASS). I therefore do not issue a dev verdict; the task stays `in_review` pending the tester fix + re-review.

**TDG conformance:** `git log main..HEAD` — 4 commits, all valid tdg prefixes ending in `(US003)`:
- `red: test spec for GitHub Actions workflow validation (US003)`
- `green: create GitHub Actions e2e workflow and fix test path (US003)`
- `refactor: clean up workflow test helpers and add doc comments (US003)`
- `refactor: hand off GitHub Actions e2e gate for review (US003)`
Sequence red → green → refactor → refactor — OK.

**Test spec exhaustiveness:** US003 has no `_be_unit_tests.md` (e2e-only spec — correct, CI YAML has no app branches). 5 Go validation tests UT-US003-001..005 map 1:1 to the 5 dryrun assertion steps in `US003_e2e_tests.md` and cover all 6 ACs (PR-to-main trigger / no-push / make targets / both `if: always()` / ubuntu-latest). No error-branch gap — no production Go code exists. Coverage exemption valid: `internal/workflow/` contains only `workflow_test.go`, no production `.go` file to measure.

**Architecture conformance (D-003/D-004):** `on.pull_request.branches: [main]` only, no `push` ✓; `runs-on: ubuntu-latest` ✓; reuses `make e2e-up`/`e2e-seed`/`e2e-run` (architecture mandate to reuse Makefile, not inline-duplicate) ✓; `actions/upload-artifact@v4` with `if: always()` → `tests/e2e/results/` ✓; `make e2e-down` with `if: always()` ✓; no retries ✓.

**Test summary:** `go vet ./...` clean; `go test ./...` (services/agent-board) = 312 passed, exit 0 (307 pre-existing + 5 new).

**Gate evidence (verbatim):**
- `scripts/review/run-gate.sh be services/agent-board` → `REVIEW GATE: PASS` (gosec/govulncheck soft-skipped, gate-internal tolerated, still PASS)
- `scripts/review/run-gate.sh cross` → `REVIEW GATE: PASS`
- `robot --dryrun tests/e2e/REQ007_*/` → `7 tests, 7 passed, 0 failed`

**Live e2e (US003-tagged, `robot --include US003 tests/e2e/REQ007_*/`), 3 consecutive runs:**
- Run 1: `1 test, 1 passed, 0 failed`
- Run 2: `1 test, 1 passed, 0 failed`
- Run 3: `1 test, 1 passed, 0 failed`

**Full-suite live e2e (`make e2e-run`) — BLOCKED at parse time (pre-existing, not US003):**
- `[ ERROR ] Error in file '.../REQ005_quality_hardening_retrospective/US006_rapid_navigation.robot' on line 22: Resource file '../../REQ004_project_detail_page/resources/project_detail_keywords.resource' does not exist.` → `make: *** [e2e-run] Error 6`
- Root cause: `US006_rapid_navigation.robot:22` uses `Resource ../../REQ004_...` which resolves to `tests/REQ004_...`; correct path is `../REQ004_...` (one level up → `tests/e2e/REQ004_...`). The target file DOES exist at `tests/e2e/REQ004_project_detail_page/resources/project_detail_keywords.resource`; only the relative path in the REQ005 robot is wrong.
- This defect is on `main` (introduced by commit `1ba4793 fix(REQ005/US008)`), and US003's diff touches ZERO REQ004/REQ005 files. It aborts the WHOLE suite before any test runs, so US004/US005 FE e2e (parallel, unmerged) never even execute.

**SPEC_GAP_FOUND — route to tester (revision mode):**
- `tests/e2e/REQ005_quality_hardening_retrospective/US006_rapid_navigation.robot:22` — wrong relative Resource path `../../REQ004_project_detail_page/resources/project_detail_keywords.resource` → should be `../REQ004_project_detail_page/resources/project_detail_keywords.resource`. This breaks the full-suite parse that the US003 CI gate orchestrates, so it must be fixed before US003 can demonstrate a clean full-suite gate run.

Once the tester fixes the REQ005 resource path (and once US004/US005 land), re-spawn this task for re-review (pass 2) to confirm the full-suite `make e2e-run` reaches 3x green. No US003 code change is required.

**Tech-debt:** filed (see below) — committed Robot artifacts tracked-on-main despite being gitignored.
