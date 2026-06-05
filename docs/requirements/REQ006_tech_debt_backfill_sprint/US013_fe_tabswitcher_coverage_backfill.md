# US013/fe_tabswitcher_coverage_backfill

**Requirement:** REQ006
**Story:** US013
**Track:** FE
**Status:** in_progress
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

## Review log
