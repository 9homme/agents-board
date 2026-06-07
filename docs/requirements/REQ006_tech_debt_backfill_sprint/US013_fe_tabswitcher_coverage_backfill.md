# US013/fe_tabswitcher_coverage_backfill

**Requirement:** REQ006
**Story:** US013
**Track:** FE
**Status:** completed
**Blocked by:** none
**Worked-by:** fe-dev-2026-06-05T00:00:00Z-a121
**Implements:** REQ006/US013 AC (all 12 scenarios — click, ArrowRight/Left + wrap, Enter, Space, aria-selected, roving tabIndex, prop-driven override, tablist semantics, unrelated-key no-op), architecture §3 US013 touch row, architecture §7 (component capability confirmation + 12 FCT-* IDs + 80% coverage target + tooling pin), architecture D-009 (≥80% across stmts/branches/lines/functions).

## Goal
Backfill `web/components/ProjectDetail/TabSwitcher.tsx` test coverage to **≥80% stmts/branches/lines/functions** (architecture D-009) by adding the 12 FCT-* test cases listed in architecture §7.2. Tests-only — `TabSwitcher.tsx` is byte-for-byte unchanged. Strike-through `docs/tech_debt.md` line 75.

## Scope
- **In:** Edit `web/components/ProjectDetail/TabSwitcher.test.tsx` to add the 12 FCT-* test cases enumerated in architecture §7.2. Use `@testing-library/user-event` for keyboard interactions (architecture §7.4 — `fireEvent.keyDown` alone skips focus management; user-event dispatches the full `keydown` + `keyup` with focus updates). Match elements by `getByRole('tab', { name: /documents/i })` and similar role+name queries — NOT `getByText`, NOT `getByTestId` (architecture §7.4). Strike-through `docs/tech_debt.md` line 75.
- **Out:** Any change to `TabSwitcher.tsx` itself. Any test for `Home` / `End` keys — the component does NOT implement them and the AC excludes them (architecture §7.1 — confirmed). Any change to other components. Any introduction of new test utilities under `web/test/` unless ≥3 of the new FCT-* cases share genuine setup (no premature abstraction).

## Files touched (estimated, exclusive)
- `web/components/ProjectDetail/TabSwitcher.test.tsx` (edit — add 12 FCT-* cases)
- `docs/tech_debt.md` (edit — strike-through line 75)

## Test contract
Dev makes the following FCT-* IDs pass (architecture §7.2 — verbatim, 1:1 with US013 AC scenarios):
- **FCT-US013-001** — clicking non-active tab fires `onTabChange` with clicked id; component does NOT internally mutate `activeTab`.
- **FCT-US013-002** — `ArrowRight` moves focus forward + fires `onTabChange`.
- **FCT-US013-003** — `ArrowRight` from last tab wraps to first.
- **FCT-US013-004** — `ArrowLeft` moves focus backward + fires `onTabChange`.
- **FCT-US013-005** — `ArrowLeft` from first tab wraps to last.
- **FCT-US013-006** — `Enter` activates focused tab; calls `preventDefault`.
- **FCT-US013-007** — `Space` activates focused tab; calls `preventDefault`.
- **FCT-US013-008** — `aria-selected` per tab reflects `activeTab` prop.
- **FCT-US013-009** — Roving `tabIndex` per tab reflects `activeTab` prop.
- **FCT-US013-010** — Prop-driven `activeTab` change re-renders with new active tab AND does NOT fire `onTabChange`.
- **FCT-US013-011** — Tablist semantics (`role="tablist"`, `aria-label`, `aria-controls`, `id`).
- **FCT-US013-012** — Unrelated keys do NOT fire `onTabChange` AND do NOT call `preventDefault`.

Tester's `US013_fe_unit_tests.md` FCT-* IDs map 1:1 onto the architecture §7.2 list.

## Implementation notes
- **Component capabilities confirmed (architecture §7.1):** `onClick={() => onTabChange(tab.id)}`; `onKeyDown` handles `ArrowRight`, `ArrowLeft`, `Enter`, `' '` (Space) — each calls `event.preventDefault()`. ARIA: `role="tablist"`, `aria-label="Project tabs"`, `role="tab"` per button, `aria-selected`, `aria-controls`, `id`, `tabIndex={isSelected ? 0 : -1}` (roving tabindex). NOT IMPLEMENTED: `Home`, `End`, `Tab`, `Escape`. Any other key is a no-op (verified in code — line ~50 returns early for unrelated keys).
- **Tooling pin (architecture §7.4):**
  - `@testing-library/user-event` for keyboard interactions.
  - `getByRole('tab', { name: /documents/i })` / `{ name: /user stories/i }` for element queries.
  - Spy on `preventDefault` via `jest.spyOn(KeyboardEvent.prototype, 'preventDefault')` OR by passing a `keyDown` handler at component-test level — both acceptable. Tester picks one for consistency in the spec.
- **Controlled-component invariant (FCT-001 + FCT-010):** the component does NOT own `activeTab` state. Verify by rendering with `activeTab="documents"`, clicking the User Stories tab, then asserting `getByRole('tab', { name: /documents/i })` STILL has `aria-selected="true"` (parent hasn't updated yet). The `onTabChange` callback is the only thing that fires.
- **Wrap-around modulo arithmetic (FCT-003 + FCT-005):** the component cycles via modulo — `(activeIndex + 1) % tabs.length` for right, `(activeIndex - 1 + tabs.length) % tabs.length` for left. Tests assert that pressing the wrap-arrow at the boundary fires `onTabChange` with the opposite end's id.
- **Prop-driven override (FCT-010):** use `rerender` from `render({...})` with the new `activeTab` prop; assert the visible `aria-selected="true"` shifts to the new tab AND that `onTabChange` was NOT called by the re-render (the prop change is parent-driven, not internal).
- **Unrelated-keys no-op (FCT-012):** press e.g. `'a'`, `'Tab'`, `'Escape'` — assert `onTabChange` was NOT called AND `preventDefault` was NOT called.
- **Coverage target (architecture §7.3 / D-009):** ≥80% stmts/branches/lines/functions. The 82-line component with ~13 lines of JSX/className-computation reaches 80% comfortably with the 12 FCT-* tests (architecture §7.3 confirms).
- **Coverage check command:**
  ```
  cd web && npm test -- --coverage --watchAll=false --forceExit \
      --collectCoverageFrom='components/ProjectDetail/TabSwitcher.tsx'
  ```
  Read the per-file stmts/branches/lines/functions row for `TabSwitcher.tsx` from the table — all four must be ≥80%.
- **react-doctor evidence (mandatory per `.claude/agents/fe-dev.md`):** even though this is a tests-only story (architecture §10.2 reduces react-doctor to a sanity sweep), the dev MUST still run `npx react-doctor@latest --verbose --diff` and paste the final score line into `## Notes` before hand-off. The diff is small (a single test file) so the report should show no regression and no new errors/warnings. Tech-lead verifies the line is present and the score has not regressed.

## Definition of done
- All 12 FCT-* test cases present (architecture §7.2 verbatim names); all green via `cd web && npm test -- --watchAll=false`.
- `TabSwitcher.tsx` coverage row shows ≥80% stmts / ≥80% branches / ≥80% lines / ≥80% functions in the Jest `--coverage` table.
- `TabSwitcher.tsx` byte-for-byte unchanged (`git diff web/components/ProjectDetail/TabSwitcher.tsx` empty).
- `cd web && npm run typecheck` clean. `cd web && npm run lint --max-warnings=0` clean.
- `docs/tech_debt.md` line 75 strike-through applied with `→ fixed in REQ006/US013`.
- **react-doctor line pasted into `## Notes`** (mandatory — the verbatim final score line from `npx react-doctor@latest --verbose --diff`).
- **Review gate green:** `scripts/review/run-gate.sh fe` + `scripts/review/run-gate.sh cross` both `REVIEW GATE: PASS`.
- **Live e2e NOT required** (architecture §10.2 — component-test only); no e2e re-run.
- Dev set status to `in_review`; tech-lead approved.

## Notes
- **No FE/BE pairing.** US013 stands alone; no API contract is touched. The BE-side state of REQ006 is irrelevant to this story's correctness.

### Hand-off notes (fe-dev-2026-06-05T00:00:00Z-a121)

**Files touched:**
- `web/components/ProjectDetail/TabSwitcher.test.tsx` — edited: added 12 FCT-US013-* test cases; `@testing-library/user-event` imported
- `docs/tech_debt.md` — line 75 struck through with `→ fixed in REQ006/US013`
- `docs/requirements/REQ006_tech_debt_backfill_sprint/US013_fe_tabswitcher_coverage_backfill.md` — status updated
- `docs/requirements/REQ006_tech_debt_backfill_sprint/US013_fe_unit_tests.md` — added to worktree (read-only reference)

**Tests added:** 12 FCT-US013-* cases (FCT-US013-001 through FCT-US013-012). Total suite: 16 TabSwitcher tests, 142 total.

**TabSwitcher.tsx:** byte-for-byte unchanged — `git diff web/components/ProjectDetail/TabSwitcher.tsx` is empty.

**Coverage achieved:**
- `TabSwitcher.tsx` stmts: **100%** (target ≥80%)
- `TabSwitcher.tsx` branches: **100%** (target ≥80%)
- `TabSwitcher.tsx` lines: **100%** (target ≥80%)
- `TabSwitcher.tsx` functions: **100%** (target ≥80%)

**react-doctor --diff score:** `100 / 100 — No issues found` (scanned `worktree-agent-a121284da93c79bfa → main`; no regression, no new errors, no new warnings)

**Review gates:**
- `scripts/review/run-gate.sh fe`: `REVIEW GATE: PASS`
- `scripts/review/run-gate.sh cross`: `REVIEW GATE: FAIL` — pre-existing Dockerfile missing-USER semgrep finding; verified failing identically on base branch before this task's changes (not introduced by this diff)

**Live e2e:** Architecture §10.2 explicitly waives e2e for this story (component-test only; no API contract touched). No e2e run required or performed.

**TDG cycle commits:**
- `red:` — 12 FCT-US013-* test cases written
- `green:` — all 16 TabSwitcher tests pass; TabSwitcher.tsx unchanged
- `refactor:` — tech_debt.md strike-through; REQ006 task files added to worktree

## Review log

### Review pass 1 — 2026-06-05 — verdict: approved

**Verdict:** approved (Status → `completed`).

**Architecture conformance (D-009 / §7.1 / §7.2 / §10.2):**
- All 12 FCT-US013-* test cases present at `web/components/ProjectDetail/TabSwitcher.test.tsx:62-233`, naming matches §7.2 verbatim (FCT-US013-001…012).
- `web/components/ProjectDetail/TabSwitcher.tsx` is **byte-for-byte unchanged** vs `main` (`git diff main -- web/components/ProjectDetail/TabSwitcher.tsx` → 0 lines).
- Architecture-excluded keys (`Home`, `End`, `Escape`, `Tab`) are NOT tested as behaviour-fires — only `FCT-US013-012` asserts the no-op on unrelated keys (`Escape`, `a`, `Tab`). Conforms to §7.1 NOT IMPLEMENTED list.
- Architecture §10.2 explicitly waives live e2e for this component-only story: "US013 is component-level only; it does not require an e2e run." Confirmed — no e2e run performed.

**Test contract:**
- All 12 FCT-* IDs in `US013_fe_unit_tests.md` are implemented; the existing 4 FCT-US001-* tests are preserved.
- TabSwitcher suite: 16/16 passing. Full FE suite: 142/142 passing.

**TDD honesty (TDG):**
- Commit sequence on dev branch `worktree-agent-a121284da93c79bfa`: `2af8d73 red:` → `eb9396b green:` → `0d2c5fe refactor: chore:` → `dd998f4 refactor: chore: hand off`. All commits end with `(US013)` traceability tag and follow red → green → refactor ordering. Conforms.

**Scope:**
- Files touched per actual diff (matches `## Files touched`): `web/components/ProjectDetail/TabSwitcher.test.tsx` (+176 lines), `docs/tech_debt.md` (1 strike-through), plus the REQ006 task/spec doc files (in-scope per `## Files touched` allowance). Zero drive-by edits to other components or to `TabSwitcher.tsx`.

**Coverage (D-009 — ≥80% across all four metrics):**
```
File             | % Stmts | % Branch | % Funcs | % Lines | Uncovered Line #s
TabSwitcher.tsx  |     100 |      100 |     100 |     100 |
```
All four metrics: 100% (well above the 80% architecture target).

**react-doctor evidence (mandatory per `.claude/agents/fe-dev.md`):**
- Present in `## Notes` line 90: `react-doctor --diff score: 100 / 100 — No issues found (scanned worktree-agent-a121284da93c79bfa → main; no regression, no new errors, no new warnings)`. Verified present, no regression, no new errors, no new warnings.

**tech_debt.md line 75 strike-through:**
- Verified at `docs/tech_debt.md:75` — strike-through applied with `→ fixed in REQ006/US013` suffix. Conforms.

**Spec exhaustiveness (anti-REQ005/US005 branch-count check):**
- `TabSwitcher.tsx` branch surface = 4 keyboard branches (`ArrowRight`, `ArrowLeft`, `Enter`/`Space`, unrelated-key no-op) + 1 click + 2 prop-driven render branches (`isSelected` → tabIndex/aria-selected/className) + ARIA tablist + 2 wrap-around modulo edges. 12 FCT-* cases hit all of these (FCT-001 click + controlled invariant; FCT-002..005 four arrow paths with wrap; FCT-006/007 Enter/Space; FCT-008/009 aria + tabIndex; FCT-010 prop-driven; FCT-011 tablist semantics; FCT-012 unrelated-key no-op). 100% coverage corroborates. No spec gap.

**Gate evidence:**
- `cd web && npm run typecheck`: clean (no errors).
- `cd web && npx jest --watchAll=false --forceExit`: `Test Suites: 17 passed, 17 total / Tests: 142 passed, 142 total`.
- `cd web && npx jest --watchAll=false --testPathPatterns=TabSwitcher --coverage --collectCoverageFrom=components/ProjectDetail/TabSwitcher.tsx --forceExit`: `Tests: 16 passed, 16 total`, coverage table as quoted above.
- `scripts/review/run-gate.sh fe`: **`REVIEW GATE: PASS`** (typecheck PASS, lint PASS, jest PASS, npm-audit WARN but non-fatal, CSR-only scan PASS, no-raw-fetch scan PASS).
- `scripts/review/run-gate.sh cross`: emits `REVIEW GATE: FAIL (1 check(s))` — semgrep `dockerfile.security.missing-user.missing-user` finding on `services/agent-board/Dockerfile:31` and `web/Dockerfile:48`. **Verified pre-existing on `main`:** `git diff main -- services/agent-board/Dockerfile web/Dockerfile` → 0 lines. Same finding flagged identically on sibling tasks `US001_be_task_repo_error_tests.md` and `US002_be_user_story_repo_error_tests.md` and approved-around there (cross-gate FAIL on out-of-scope, pre-existing Dockerfile files; same repo state). gitleaks: PASS.
- **Live e2e:** waived per architecture §10.2. Not run.

**Gate verdict reconciliation (cross gate FAIL, approved nonetheless):**
- The semgrep finding is on `services/agent-board/Dockerfile:31` and `web/Dockerfile:48` — both **byte-for-byte unchanged** vs `main` in this diff. The dev cannot remediate within their `## Files touched` scope.
- Sibling-task precedent in this same REQ (US001, US002, both already merged to `main`) documents the identical pre-existing FAIL and approved-around it.
- Approving here follows established REQ006 norm; the dockerfile-USER finding is a **repo-wide tech-debt item not tied to US013** and is being filed below for a dedicated follow-up (so it stops blocking every review).

**Tech-debt filed:**
- 2026-06-05 — services/agent-board/Dockerfile:31 + web/Dockerfile:48 — semgrep `dockerfile.security.missing-user.missing-user` finding (containers run as root) blocks every `scripts/review/run-gate.sh cross` invocation; pre-existing across all REQ006 reviews; add `USER non-root` line OR exempt in a semgrep baseline so cross-gate is meaningful again — REQ006/US013 (filed during review pass 1).
- 2026-06-05 — scripts/review/run-gate.sh — no baseline/ignore mechanism for known pre-existing semgrep findings; once any blocking finding lands on main, every subsequent task's cross-gate is FAIL until the underlying code is fixed; consider a `.semgrepignore` or `--baseline-commit` plumb-through — REQ006/US013 (filed during review pass 1).

