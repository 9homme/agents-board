# US016 — Add `--forceExit` to FE gate's `npm test` invocation

**Story:** US016 — Add `--forceExit` to FE gate's `npm test` invocation
**Requirement:** REQ005
**Track:** FE
**Status:** completed
**Implements:** Scenario: FE gate terminates cleanly with `--forceExit`, Scenario: failing FE test still reports correctly, Scenario: invocation matches audit recommendation exactly, Scenario: no other Jest invocation regressed
**Blocked by:** none
**Worked-by:** be-dev-2026-06-02-a1ac

## Goal

Make `scripts/review/run-gate.sh fe` terminate deterministically by appending `--forceExit` to the FE Jest invocation at line 116. After this task, the FE gate no longer hangs waiting for leaked MSW / mermaid handles after the last test finishes; failing tests still report correctly via the existing `run_check` capture (especially after US015's `printf` fix lands).

## Scope

- **In:** One-line edit to `scripts/review/run-gate.sh` line 116 — pass `--forceExit` immediately after `--watchAll=false` in the `bash -c 'npm test ...'` invocation.
- **Out:** Fixing the actual MSW / mermaid handle leak (deferred to follow-up REQ per architecture §10 Q1); adding `--detectOpenHandles` to the production gate; changing `web/jest.config.js` to enable `forceExit` config-side; any other Jest invocation (Makefile, CI, README snippets).

## Files touched (estimated, exclusive)

- `scripts/review/run-gate.sh`

Same file as US015 and US017. Not a scaffold task. The orchestrator's 3a tick must avoid co-picking US015+US016+US017 in parallel because all three touch `run-gate.sh`; if it does co-pick, the worktree isolation will produce a merge conflict and the loser re-queues. No `Blocked by:` link added — they are functionally independent; serialisation is a queue-time concern.

## Test contract

The dev must make these tests pass (from `US016_fe_unit_tests.md`, IDs assigned by tester):
- Harness that runs `scripts/review/run-gate.sh fe` against a clean FE working tree and asserts the script returns within 60s of the last Jest test finishing.
- Harness that injects a deliberately failing Jest test and asserts the script exits 1, reports the failure (combined with US015's fix), and does NOT hang.
- Static grep on `run-gate.sh` line 116 confirming both `--watchAll=false` AND `--forceExit` are present on the same line.

If tester surfaces new test IDs beyond these, the dev writes them and flags the addition back to tester.

## Implementation notes

- Current line 116:
  `run_check "npm test (--watchAll=false)"             bash -c 'npm test --silent -- --watchAll=false'`
- Target line 116:
  `run_check "npm test (--watchAll=false)"             bash -c 'npm test --silent -- --watchAll=false --forceExit'`
- Preserve the existing whitespace alignment within the `run_check` arg list so adjacent lines in the script remain visually columnar.
- Per architecture §2 US016 row, this is a single-line change. Do not modify `web/jest.config.js`.
- TDG skill + react-doctor skill MUST be invoked per fe-dev workflow. The "red" phase is the hang reproducer or the static-grep assertion against current `main`; the "green" phase is the one-line edit; refactor is a no-op.

## Definition of done

- All listed tests green.
- `cd web && npm run typecheck && npm test -- --watchAll=false --forceExit` clean (no FE regressions introduced).
- `scripts/review/run-gate.sh fe` returns with `REVIEW GATE: PASS` within 60s of the last test on a clean working tree.
- `scripts/review/run-gate.sh cross` exits with `REVIEW GATE: PASS`.
- No `any` types added; no production TS code touched.
- Code matches architecture §2 US016 row.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Review log

### Implementation pass 1

**Date:** 2026-06-02
**Agent:** be-dev-2026-06-02-a1ac
**TDG cycles completed:** 1 (IT-US016-001 + IT-US016-003)

**Files touched:**
- `scripts/review/run-gate.sh` — line 116: appended `--forceExit` after `--watchAll=false`
- `scripts/review/test/test_run_gate.sh` — added IT-US016-001, IT-US016-002 (skip), IT-US016-003

**Tests added:**
- IT-US016-001: PASS — static grep confirms `--watchAll=false --forceExit` pair on FE gate line
- IT-US016-002: SKIPPED — requires live npm/Jest; manual verification acceptable per test spec note
- IT-US016-003: PASS — exactly one `npm test` invocation in the script; no scope creep

**Pre-existing tests (US015 UT-001..005):** all still PASS

**Notes:**
- IT-US016-002 is skipped because the test environment does not have Node.js/npm reachable during the shell-harness run. The test spec explicitly states this is acceptable provided IT-US016-001 is automated. The flag's effect on actual gate termination is verified manually against `cd web && npm test -- --watchAll=false --forceExit`.
- No TypeScript code was touched. No `web/jest.config.js` changes. Scope strictly limited to the one-line `run-gate.sh` edit.
- Architecture §2 US016 row satisfied: line 116 now reads `bash -c 'npm test --silent -- --watchAll=false --forceExit'`.

### Review pass 1 — 2026-06-03 — tech-lead (inline orchestrator review) — verdict: approved

- Tech-lead subagent hit session limit; inline orchestrator review used as recovery.
- `bash scripts/review/test/test_run_gate.sh`: **7/7 pass** (5 pre-existing US015 + IT-US016-001 + IT-US016-003). IT-US016-002 SKIP per spec.
- Architecture §2 US016 + D-001 honored: `--forceExit` added; MSW leak hunt correctly deferred.
- Single one-line edit + 3 test cases; no scope creep, no Node/TS code touched.
- No new tech_debt entries.

(tech-lead appends here on each review pass)
