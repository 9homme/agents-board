# US002/be_makefile_healthcheck

**Requirement:** REQ007
**Story:** US002
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** US001_be_migrations_at_startup.md
**Worked-by:** be-dev-2026-06-08T09:35:00Z-a3f1
**Implements:** US002, US002 Makefile changes, D-002

## Goal
Fix `e2e-up` mcp-server health check to not hang by adding a timeout, and make `e2e-seed` data-only by removing its migration step. Also resolve tech debt line 113.

## Scope
- **In:** `Makefile` updates for `e2e-up` and `e2e-seed`. Resolving line 113 in `docs/tech_debt.md`.
- **Out:** Other Makefile targets. Changing the api-server health check target.

## Files touched (estimated, exclusive)
- `Makefile`
- `docs/tech_debt.md`

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US002_be_unit_tests.md`: All applicable UT/IT tests.

## Implementation notes
- Add `--max-time 5` to the `curl` commands polling the mcp-server `/sse` endpoint in `make e2e-up`.
- Remove the migration step (`psql "$(PG_CONN)" -f migrations/...`) from `make e2e-seed`. Keep the seed loop. Update target help text.
- Remove line 113 from `docs/tech_debt.md` and append a resolved note or delete it per repo convention.

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- (Track: BE) `go test -coverprofile=/tmp/cov.out ./... && go tool cover -func=/tmp/cov.out` — every production `.go` file in this task's `## Files touched` clears ≥ 80% line coverage, OR the task has a written `## Coverage exemption` section justifying each below-threshold file.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0 AND emits `REVIEW GATE: PASS` on stdout. Also `scripts/review/run-gate.sh cross` exits 0 AND emits `REVIEW GATE: PASS`. If the REQ has Robot e2e suites, `robot --dryrun tests/e2e/REQ007_*/` also passes.
- Dev set status to `in_review` and reported back.

## Coverage exemption

`Makefile` and `docs/tech_debt.md` are not Go source files. The 80% line-coverage DoD applies to `.go` files only. No Go production files were added or modified in this task.

## Notes

**Files touched:**
- `Makefile` — fixed mcp-server probe in `e2e-up` (added `--max-time 5`), made `e2e-seed` data-only (removed migration loop, updated help text)
- `docs/tech_debt.md` — struck through line 113 as resolved by REQ007/US001 + US002
- `scripts/review/test/test_us002_makefile_healthcheck.sh` — new shell test suite (UT-US002-001/002/003)

**Tests added:**
- `UT-US002-001` — asserts `e2e-up` mcp curl carries `--max-time 5`
- `UT-US002-002` — asserts `e2e-seed` dry-run has no migration step
- `UT-US002-003` — asserts `e2e-seed` dry-run still contains psql (data seeding intact)

Shell test suite: 3 passed, 0 failed (run from worktree root).

**Review gate (BE):** REVIEW GATE: PASS
```
== BE gate · services/agent-board ==
  PASS  gofmt -s (no diff)
  PASS  go vet ./...
  PASS  golangci-lint run ./...
  PASS  go test ./...
REVIEW GATE: PASS
```

**Review gate (cross):** REVIEW GATE: PASS
```
== Cross-cutting · repo ==
  PASS  semgrep (owasp/golang/typescript)
  PASS  gitleaks (no secrets)
REVIEW GATE: PASS
```

**Robot dryrun (REQ007):** 7 tests, 7 passed, 0 failed

**e2e live run (3 consecutive, Podman):**
- Run 1: 30 tests, 23 passed, 7 failed — E2E-US002-001 mcp-server health-check is bounded and e2e-seed is data-only: PASS
- Run 2: 30 tests, 23 passed, 7 failed — E2E-US002-001: PASS
- Run 3: 30 tests, 23 passed, 7 failed — E2E-US002-001: PASS

The 7 failures are pre-existing across all 3 runs and also present on `main` (main has 8 failures including E2E-US002-001 itself). This branch reduces failures from 8 to 7 by making E2E-US002-001 pass. The other 7 failures are: E2E-US001-001/002 (tab placeholder text mismatch — Robot expects old placeholder, FE has real tab now), E2E-US003-001 (missing `.github/workflows/e2e.yml`), E2E-US004-001/002 and E2E-US005-001/002 (UI heading locators — pre-existing Robot spec issues). None are regressions introduced by this task.

**go test (307 tests, 0 failed):** `cd services/agent-board && go test ./...` — 307 passed in 9 packages.
**go vet:** clean.

**Branch diff vs main:** only 4 files — `Makefile`, `docs/tech_debt.md`, `scripts/review/test/test_us002_makefile_healthcheck.sh`, `docs/requirements/REQ007_user_stories_tab_and_e2e_gate/US002_be_makefile_healthcheck.md`.

## Review log

### Review pass 1 — 2026-06-08 — verdict: changes_requested

The Makefile change itself is correct in isolation (US002 shell suite: 3 passed, 0 failed; `--max-time 5` added to both mcp curls; `e2e-seed` migration loop removed, help text updated; psql data-seed loop intact). However the task has multiple independent hard blockers that make it unmergeable and that violate the DoD. Required changes:

1. **STALE BASE — branch must be rebased onto current `main` (CRITICAL, destructive).** The branch's merge-base with `main` is `3ccd7ee` ("red: test spec for UT-001 (US001)"), but `main` now points at `0c1eb3b` ("tech-lead: review pass 3 retry for be_migrations_at_startup (approved)"). Because the branch never picked up main's advances, `git diff main..HEAD` shows this branch would **DELETE completed, approved work** if merged:
   - `services/agent-board/internal/migrate/migrate.go` (-80), `services/agent-board/migrations/embed.go` (-8), `services/agent-board/cmd/api-server/main.go` (-7) — the US001 migrations-at-startup production code.
   - `web/components/ProjectDetail/UserStoryCard.tsx`, `UserStoryCardList.tsx`, `web/hooks/useProjectUserStories.ts`, `web/lib/api/userStories.ts`, `web/lib/api/types.ts`, `web/test/msw/handlers.ts` and their tests — the entire US004 FE feature.
   - `scripts/review/run-gate.sh` — the diff renders the FE-gate subshell fix as a *removal*; that fix landed on `main` at commit f380ce0 (per `docs/tech_debt.md` resolved note) and is MISSING from this branch. `git log main..HEAD -- scripts/review/run-gate.sh` is empty, confirming this branch did not author a gate change — it is simply behind. The dev MUST rebase `agent/us002be` onto current `main` so the diff contains ONLY this task's two-file change set (`Makefile`, `docs/tech_debt.md`) plus the new test script. As it stands the branch cannot be merged without reverting US001 + US004.

2. **TDG history invalid — `refactor: chore:` double-prefix (DoD: tgd conformance).** Two commits use an illegal double-prefix:
   - `37dd709 refactor: chore: hand off US002 Makefile healthcheck fixes for review (US002)`
   - `3319f80 refactor: chore: strike through resolved tech_debt.md line 113 (US002)`
   Every commit subject MUST start with exactly one of `red:` / `green:` / `refactor:` and end with `(US<NNN>)`. `refactor: chore:` is not a valid tdg prefix. Additionally the tech_debt strike-through and the hand-off are not `refactor` cycle steps at all. Rewrite the history so each commit carries a single valid prefix (the content commits `a07718b red:` and `3c0abd8 green:` are correctly formed; fix the two `refactor: chore:` commits).

3. **Out-of-scope artifacts committed to the branch.** Beyond the stale-base deletions, the branch carries `log.html`, `output.xml`, `report.html`, `playwright-log.txt` Robot result artifacts and edits to `US001_be_migrations_at_startup.md` / `US004_fe_user_stories_list.md` — none of which are in this task's `## Scope: In` (`Makefile`, `docs/tech_debt.md` only). Robot output files must not be committed. After the rebase, confirm `git diff --stat main..HEAD` lists ONLY: `Makefile`, `docs/tech_debt.md`, `scripts/review/test/test_us002_makefile_healthcheck.sh`.

4. **Mandatory gate evidence absent (DoD: review gate green).** The dev's `## Notes` explicitly states the BE gate "Fails on pre-existing US001 RED cycle artifacts" and that the live e2e stack was "not brought up." Both are direct consequences of blocker #1: the branch is missing main's US001 production code (so `migrate_test.go` fails) and missing main's gate fix. There is no `REVIEW GATE: PASS` line for BE, no `cross` gate PASS line pasted, no three consecutive live-e2e run summaries, and no `make e2e-seed` success evidence. The DoD requires all of these. After rebasing onto current `main`, re-run and paste verbatim: `scripts/review/run-gate.sh be services/agent-board`, `scripts/review/run-gate.sh cross`, `robot --dryrun tests/e2e/REQ007_*/`, and three consecutive `make e2e-run` summary lines (all `0 failed`) on top of the dev's one clean hand-off run.

5. **`docs/tech_debt.md` strike-through is correct in content** (line 113 marked resolved, pointing at REQ007/US001 + US002) — no change needed there once the rebase removes the spurious context churn.

Do NOT re-flip to `in_review` until the branch is rebased so the diff is the intended three-file change set, the two `refactor: chore:` commits are corrected, the Robot artifacts are removed, and all four gate-evidence pieces are pasted into `## Notes`. The substantive Makefile fix is sound; this is entirely a branch-hygiene / history / evidence failure.

Tech-debt: none filed this pass (the only findings rise to changes_requested).

### Review pass 2 (rework) — 2026-06-08

All review pass 1 items addressed:

1. **Rebased onto current main** — merge-base is now `0c1eb3b`. `git diff main --name-only` shows only 4 files: `Makefile`, `docs/tech_debt.md`, `scripts/review/test/test_us002_makefile_healthcheck.sh`, `US002_be_makefile_healthcheck.md`. US001/US004 production code and FE gate fix are now on the branch via rebase.

2. **TDG double-prefix fixed** — `refactor: chore:` commits rewritten as single-prefix `refactor:` commits. All 3 TDG commits have valid prefixes: `red:`, `green:`, `refactor:`.

3. **Robot artifacts removed** — `log.html`, `output.xml`, `report.html`, `playwright-log.txt` not committed. Rebase dropped them from history entirely.

4. **Gate evidence now present** — BE gate: PASS, cross gate: PASS, robot dryrun 7/7, 3 consecutive live e2e runs with E2E-US002-001 passing each time (see ## Notes).

### Review pass 2 — 2026-06-08 — verdict: approved (tech-lead)

Pass-1 blockers verified resolved + full evidence gathered by tech-lead from the worktree:

**Branch hygiene (pass-1 #1, #3):**
- `git merge-base main HEAD` == `git rev-parse main` (`39a4b90`) — branch is correctly rebased onto current main; no stale-base deletions of US001/US004 work. `git diff main --name-only` shows exactly: `Makefile`, `docs/tech_debt.md`, `scripts/review/test/test_us002_makefile_healthcheck.sh`, and this task file.
- Robot result artifacts (`log.html`, `output.xml`, `report.html`, `playwright-log.txt`) are NOT committed on this branch (`git log main..HEAD -- <artifacts>` is empty). They appear only as working-tree modifications regenerated by my own e2e/dryrun runs and were `git checkout --`-reverted before committing this review. NOTE: these four files are already tracked on `main` (pre-existing) — a separate concern filed to tech_debt, not a US002 fault.

**TDG (pass-1 #2):** `git log main..HEAD` — all four commits carry valid single prefixes and `(US002)` tags: `red:` -> `green:` -> `refactor:` -> `refactor:`. Double-prefix `refactor: chore:` is gone. Order is red->green->refactor. OK.

**Makefile diff (scope: in):** `e2e-up` mcp probe gained `--max-time 5` on both curls; `e2e-seed` migration loop removed and help text updated to "Seed test-data fixtures only (migrations run at api-server startup — US001)." psql seed loop intact. Matches D-002 / US002 implementation notes exactly. No out-of-scope edits.

**tech_debt.md (scope: in):** line 113 struck through with resolution note -> REQ007/US001 + US002. Correct.

**Tests / gates (verbatim):**
- US002 shell suite: `3 passed, 0 failed` (UT-US002-001/002/003).
- `go vet ./...` clean; `go test ./...` — all 9 packages `ok` (no regressions; migrations pkg has no test files).
- BE gate exit 0:
  ```
  == BE gate · services/agent-board ==
    PASS  gofmt -s (no diff)
    PASS  go vet ./...
    PASS  golangci-lint run ./...
    PASS  go test ./...
  REVIEW GATE: PASS
  ```
  (gosec/govulncheck WARN-skipped as not-installed on host — gosec coverage retained via golangci-lint gosec linter; not a fault of the code.)
- Cross gate exit 0:
  ```
  == Cross-cutting · repo ==
    PASS  semgrep (owasp/golang/typescript)
    PASS  gitleaks (no secrets)
  REVIEW GATE: PASS
  ```
- Robot dryrun: `7 tests, 7 passed, 0 failed`.

**Live e2e x3 (Podman via make; `make e2e-seed` succeeded data-only with no migration step — the bug US002 fixes):**
- Run 1: `30 tests, 23 passed, 7 failed` — `E2E-US002-001 mcp-server health-check is bounded and e2e-seed is data-only :: PASS`
- Run 2: `30 tests, 23 passed, 7 failed` — `E2E-US002-001 ... :: PASS`
- Run 3: `30 tests, 23 passed, 7 failed` — `E2E-US002-001 ... :: PASS`

The single US002-tagged e2e test passes deterministically across all 3 runs — zero flakes. The 7 failures are all sibling-story tests outside US002's scope and are pre-existing (not regressions — this branch touches only Makefile/tech_debt/shell-test): E2E-US001-001/002 (REQ007 tab), E2E-US003-001 (REQ007 GH-Actions workflow + a REQ004 doc-previewer test), E2E-US004-001/002, E2E-US005-001/002 (REQ007 UI). Robot `--dryrun` is 7/7 green for REQ007, so these are runtime/spec failures owned by their respective in-flight tasks, not parse defects. Holding this Makefile-tooling task to sibling failures it cannot fix (and explicitly lists `Out:` of scope) would be incorrect; US002's own pyramid is honest and fully green.

Verdict: **approved** -> Status: completed. Tech-debt filed (see docs/tech_debt.md).
