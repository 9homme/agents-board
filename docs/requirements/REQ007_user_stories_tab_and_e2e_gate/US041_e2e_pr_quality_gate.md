# US041 — GitHub Actions e2e PR quality gate

**Requirement:** REQ007 — User Stories Tab + E2E Quality Gate + Health-Check Fixes
**Status:** draft

## Story
As a maintainer, I want every pull request to automatically run the Robot Framework e2e suite and block merge on failure, so that broken end-to-end behavior cannot land on `main` and the manual workaround is no longer the only evidence.

## Acceptance criteria
- **Scenario: workflow triggers on pull requests to main only (D-003 CONFIRMED)**
  - Given a GitHub Actions workflow file exists under `.github/workflows/`
  - And its trigger is `on: pull_request` with `branches: [main]`
  - When a pull request targeting the `main` branch is opened, synchronized, or reopened
  - Then the e2e workflow is triggered
  - And the workflow does NOT trigger on `push` (to `main` or any branch) or on PRs targeting non-`main` branches
- **Scenario: workflow brings up the stack and runs the full e2e suite**
  - Given the workflow is running on a GitHub-hosted Ubuntu runner
  - When the e2e job executes
  - Then it brings up the stack via `docker compose` (using the repo's compose definition / `make e2e-up` family)
  - And it applies migrations and seeds (`make e2e-seed`)
  - And it runs the Robot suites (`make e2e-run`, covering all `tests/e2e/REQ*/`)
- **Scenario: any e2e failure fails the gate**
  - Given at least one Robot test fails (or the stack fails to come up)
  - When the e2e job completes
  - Then the job exits non-zero and the check is reported as failed (no retry masks the failure — D-004)
- **Scenario: all e2e passing passes the gate**
  - Given every Robot test passes
  - When the e2e job completes
  - Then the job exits zero and the check is reported as passed
- **Scenario: artifacts always uploaded**
  - Given the e2e job has run (whether passed or failed)
  - When the job finishes
  - Then the Robot outputs (`output.xml`, `log.html`, `report.html` from `tests/e2e/results/`) are uploaded as workflow artifacts (`if: always()`)
- **Scenario: the stack is always torn down**
  - Given the e2e job has started the stack
  - When the job finishes (pass or fail)
  - Then the stack is torn down (`make e2e-down` / compose down) so runners are not left dirty

## UI / UX flow expectations
No UI: this is a CI/CD configuration change. Its only surface is the GitHub PR checks UI, which renders from the workflow's job status automatically.

## Out of scope
- Configuring branch-protection to make the check **required** — that is a repo-admin action in GitHub settings, not a file in the repo (noted as a dependency below).
- Running on `push` to main or on other branches (D-003: PR-to-main only).
- Caching/optimization of build times beyond a straightforward setup.
- Unit-test or lint gates (this story is e2e-only; existing/other CI is untouched).
- Using `podman-compose` in CI (D-003: `docker compose`).

## Dependencies
- **US040** — CI relies on a working `make e2e-up`/`e2e-seed`/`e2e-run` chain; the health-check bugs must be fixed first or the workflow will time out/hang exactly as the local stack does.
- **External (human/repo-admin):** enabling branch protection to mark this check required — the workflow alone does not block merges until that toggle is set.
- Existing Robot suites under `tests/e2e/` and the `Makefile` e2e-* targets.

## Notes for the team
- **D-003 — CONFIRMED:** trigger is `pull_request` to `main` only (no push-to-main, no other branches). **D-004 — CONFIRMED:** `docker compose` runtime, all-must-pass with no retry, always upload artifacts.
- GitHub-hosted Ubuntu runners have Docker preinstalled, so `docker compose` works out of the box. The workflow needs Go, Node, and Python/Robot toolchains available (set up via standard actions) for migrations/seeds/Robot, OR run everything through the compose stack + a Robot container — architect/tech-lead to pick the simplest path consistent with the Makefile.
- Reuse the `Makefile` targets where possible rather than duplicating commands, so CI and local stay in lockstep.
- **Testability note for the tester:** assert on the workflow YAML (trigger is `pull_request` with `branches: [main]`; steps invoke the e2e make targets; an upload-artifact step has `if: always()`; a teardown step has `if: always()`). The workflow's real green/red run is itself the ultimate evidence but is observed in the PR, not in the local test pyramid.

## Sign-off log
(po-ba appends here on each sign-off pass)
