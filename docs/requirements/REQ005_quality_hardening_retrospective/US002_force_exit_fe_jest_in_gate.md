# US002 — Add `--forceExit` to FE gate's `npm test` invocation

**Requirement:** REQ005 — quality hardening retrospective
**Status:** draft

## Story
As a **reviewer running `scripts/review/run-gate.sh fe`**, I want the FE Jest step to **terminate deterministically** within seconds of the last test finishing, so that I do not have to manually kill a hung gate run or wonder whether the suite passed.

## Acceptance criteria

- **Scenario: FE gate terminates cleanly with `--forceExit`**
  - Given the FE working tree is clean and tests pass
  - When `scripts/review/run-gate.sh fe` is invoked from the repo root
  - Then the `npm test (--watchAll=false)` line passes
  - And the script returns control to the shell within 60 seconds of the last test finishing (no indefinite hang waiting for MSW handles)
  - And the final line is `REVIEW GATE: PASS`
  - And exit code is 0

- **Scenario: failing FE test still reports correctly**
  - Given a deliberately failing FE test exists in `web/`
  - When `scripts/review/run-gate.sh fe` is invoked
  - Then the `npm test (--watchAll=false)` step shows `FAIL`
  - And the failure output is captured and printed (combined with US001's `printf` fix to ensure visibility)
  - And the script's final line is `REVIEW GATE: FAIL (1 check(s))` listing the failing test
  - And exit code is 1
  - And the script does NOT hang after Jest reports the failure

- **Scenario: invocation matches audit recommendation exactly**
  - Given `scripts/review/run-gate.sh`
  - When the FE-gate `npm test` line is read
  - Then it invokes Jest with both `--watchAll=false` AND `--forceExit` (the exact pair the audit §3.3 item 2 recommends)

- **Scenario: no other Jest invocation regressed**
  - Given any other place where Jest might be invoked from a script in the repo (`make`, CI config, README snippets)
  - When the gate script changes
  - Then no other Jest-invoking script is silently modified by this story (scope is limited to the FE-gate line)

## UI / UX flow expectations

**No UI:** developer / CI tooling change. Flow: reviewer runs `scripts/review/run-gate.sh fe` and gets a deterministic terminal result.

## Out of scope
- **Fixing the actual MSW / mermaid handle leak.** This story is the surgical unblock; the real leak hunt is intentionally left for a follow-up story (see REQ005 README "Open questions" — may become US010).
- Switching to `--detectOpenHandles` in the gate (that flag belongs in the leak-hunt story, not the production gate).
- Changing `web/jest.config.js` to add `forceExit: true` config-side (preference is the CLI flag in the gate script so day-to-day `npm test` invocations are unaffected — if architect prefers the config-side option for any reason, that's a valid `ARCHITECTURE_GAP_FOUND` to raise).

## Dependencies
- Soft: US001 should land first so that if a Jest failure does occur, the captured output is actually printed. Not a hard blocker — the two stories can ship in either order.

## Notes for the team

- **Concrete fix** per audit §3.3 item 2: edit `scripts/review/run-gate.sh` line 116 (current text: `run_check "npm test (--watchAll=false)"             bash -c 'npm test --silent -- --watchAll=false'`) to: `run_check "npm test (--watchAll=false)"             bash -c 'npm test --silent -- --watchAll=false --forceExit'`.
- **Evidence the leak is real** is in the audit: per-task reviews for REQ004 US001/US002/US003 FE work documented the hang; Jest's `A worker process has failed to exit gracefully` warning has been observed even on the audit run that did not hang.
- **Why not fix the leak itself in this story:** the audit explicitly recommends `--forceExit` as the must-fix and `--detectOpenHandles` as eventual-fix-debt. Doing both in one story bundles a fast unblock with an open-ended investigation. Keep this story to ≤1 point.

## Sign-off log
(po-ba appends here on each sign-off pass)
