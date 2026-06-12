# US036 — Backfill `TabSwitcher.tsx` coverage

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** done

## Story
As a **user navigating the project detail page**, I want the **tab switcher (Documents / User Stories) to work correctly when I click a tab, when I use the arrow keys / Home / End to move between tabs, and when the page receives a different `activeTab` prop** (e.g. from a URL change), so that I can reach either tab pane reliably via mouse OR keyboard, with the WAI-ARIA `tablist` pattern's accessibility guarantees intact (correct `aria-selected`, roving `tabIndex`, focus management).

The covered behaviours are user-observable in the existing component; this story raises test coverage to ≥80% statements/branches so a regression in any of those behaviours fails CI immediately.

## Acceptance criteria

- **Scenario: clicking a non-active tab fires `onTabChange` with the clicked tab id**
  - Given `TabSwitcher` is rendered with `activeTab="documents"` and a `jest.fn()` for `onTabChange`
  - When the user clicks the **User Stories** tab button (matched by `role="tab"` + accessible name)
  - Then `onTabChange` is called exactly once with the string `"user-stories"`
  - And the component does NOT internally update `activeTab` (it is a controlled component — parent owns state)

- **Scenario: `ArrowRight` moves focus and fires `onTabChange` to the next tab**
  - Given `TabSwitcher` is rendered with `activeTab="documents"`
  - And the Documents tab is focused
  - When the user presses `ArrowRight`
  - Then `event.preventDefault()` was called
  - And focus is now on the User Stories tab
  - And `onTabChange` is called with `"user-stories"`

- **Scenario: `ArrowRight` from the last tab wraps to the first**
  - Given `TabSwitcher` is rendered with `activeTab="user-stories"`
  - And the User Stories tab is focused
  - When the user presses `ArrowRight`
  - Then focus moves to the Documents tab
  - And `onTabChange` is called with `"documents"`

- **Scenario: `ArrowLeft` moves focus and fires `onTabChange` to the previous tab**
  - Given `TabSwitcher` is rendered with `activeTab="user-stories"`
  - And the User Stories tab is focused
  - When the user presses `ArrowLeft`
  - Then focus moves to the Documents tab
  - And `onTabChange` is called with `"documents"`

- **Scenario: `ArrowLeft` from the first tab wraps to the last**
  - Given `TabSwitcher` is rendered with `activeTab="documents"`
  - And the Documents tab is focused
  - When the user presses `ArrowLeft`
  - Then focus moves to the User Stories tab
  - And `onTabChange` is called with `"user-stories"`

- **Scenario: `Enter` activates the focused tab**
  - Given `TabSwitcher` is rendered with `activeTab="documents"`
  - And the User Stories tab is focused (e.g. via prior arrow-key navigation)
  - When the user presses `Enter`
  - Then `onTabChange` is called with `"user-stories"`
  - And `event.preventDefault()` was called

- **Scenario: `Space` activates the focused tab**
  - Given the same setup as above
  - When the user presses `Space` (the `" "` key)
  - Then `onTabChange` is called with the focused tab's id
  - And `event.preventDefault()` was called

- **Scenario: `aria-selected` reflects the active tab**
  - Given `TabSwitcher` is rendered with `activeTab="documents"`
  - When the rendered DOM is inspected
  - Then the Documents button has `aria-selected="true"`
  - And the User Stories button has `aria-selected="false"`

- **Scenario: roving `tabIndex` reflects the active tab**
  - Given `TabSwitcher` is rendered with `activeTab="documents"`
  - When the rendered DOM is inspected
  - Then the Documents button has `tabIndex={0}`
  - And the User Stories button has `tabIndex={-1}`

- **Scenario: prop-driven `activeTab` override re-renders with the new active tab**
  - Given `TabSwitcher` was initially rendered with `activeTab="documents"`
  - When the parent re-renders with `activeTab="user-stories"`
  - Then the User Stories button now has `aria-selected="true"` and `tabIndex={0}`
  - And the Documents button now has `aria-selected="false"` and `tabIndex={-1}`
  - And `onTabChange` was NOT called by the re-render alone

- **Scenario: tablist semantics are present**
  - Given `TabSwitcher` is rendered
  - When the rendered DOM is inspected
  - Then a single element has `role="tablist"` AND `aria-label="Project tabs"` (matches current source)
  - And exactly two children have `role="tab"`
  - And each tab button has the `aria-controls` attribute matching `tabpanel-<tabId>` (`tabpanel-documents`, `tabpanel-user-stories`)
  - And each tab button has the `id` attribute matching `tab-<tabId>` (`tab-documents`, `tab-user-stories`)

- **Scenario: unrelated keys do not fire `onTabChange`**
  - Given `TabSwitcher` is rendered with `activeTab="documents"`
  - And the Documents tab is focused
  - When the user presses `Tab`, `Escape`, `a`, or any other key NOT in `{ArrowLeft, ArrowRight, Enter, " "}`
  - Then `onTabChange` is NOT called
  - And `event.preventDefault()` is NOT called

- **Scenario: per-file coverage hits ≥80% statements / ≥80% branches / ≥80% lines**
  - Given `cd web && npm test -- --coverage --collectCoverageFrom="components/ProjectDetail/TabSwitcher.tsx"`
  - When the coverage report is inspected
  - Then `TabSwitcher.tsx` shows **≥80% stmts AND ≥80% branches AND ≥80% lines** (today's baseline per `docs/tech_debt.md` line 75: 41.66% stmts / 33.33% branches / 39.13% lines)
  - And ≥80% functions (currently the only uncovered function is the `useRef` callback ref — that count's small surface area means this threshold is easy to hit)

- **Scenario: existing tests still pass and component behaviour is unchanged**
  - Given the production code in `web/components/ProjectDetail/TabSwitcher.tsx` is **NOT** modified by this story
  - When `cd web && npm test` runs
  - Then all pre-existing tests pass
  - And all new tests pass
  - And `cd web && npm run typecheck` is clean
  - And `cd web && npm run lint` is clean

- **Scenario: no production-code changes**
  - Given `git diff` of the story's commits
  - When inspected
  - Then **only** `web/components/ProjectDetail/TabSwitcher.test.tsx` (and optionally a shared test helper under `web/test/`) is modified
  - And `web/components/ProjectDetail/TabSwitcher.tsx` is **byte-for-byte unchanged**

- **Scenario: closes tech-debt finding**
  - Given `docs/tech_debt.md` line 75 contains the finding about `TabSwitcher.tsx` 41.66% coverage
  - When this story is `done`
  - Then `docs/tech_debt.md` line 75 is struck through with `→ fixed in REQ006/US036`

## UI / UX flow expectations

This story does not change UI/UX behaviour — it adds test coverage for behaviour that already exists. Recapped for the tester / fe-dev:

- **Entry points:** the `TabSwitcher` component is rendered inside the project detail page. It receives `activeTab` (string) and `onTabChange` (callback) from the parent — fully controlled.
- **Happy-path flow:** user lands on the page → sees two tabs (Documents, User Stories) → clicks one → tab pane changes → parent re-renders the switcher with the new `activeTab`.
- **Keyboard flow:** user `Tab`s into the tablist → arrow keys move between tabs (with focus AND `onTabChange` firing per arrow press, per the current implementation's "activate-on-focus" semantics) → `Enter`/`Space` re-fires the change for the currently focused tab.
- **Empty / loading / error states:** N/A — the component is purely presentational; there is no loading state inside the switcher itself. The parent's tab panes handle their own loading.
- **Validation rules visible to the user:** N/A — no inputs.
- **Out of UI scope:**
  - The URL-update behaviour. The component's job is to fire `onTabChange`; the parent owns URL updates (per the existing source comment "The caller (the page) is responsible for writing the tab to the URL via a shallow router.replace call").
  - Styling — Tailwind classes are present in the source; this story does not test them beyond verifying ARIA-related class behaviour stays intact.
  - `Home`/`End` keys — **NOT implemented in the current source.** The story header originally mentioned them; po-ba removed them from AC after re-reading the source. If the team wants `Home`/`End` support, raise as a NEW behaviour-adding story (not a test-backfill).

## Out of scope
- **Modifying `TabSwitcher.tsx` production code.** Tests-only.
- **Adding `Home`/`End` keyboard support** — current source does not implement; out of scope for a test-backfill story. New behaviour requires a new story.
- **Adding a third tab** — out of scope.
- **Refactoring the component to use `useId` or a different ARIA pattern.**
- **Coverage backfill for sibling components** (`DocumentSidebar`, `MarkdownRenderer`, etc.) — not in REQ006's debt list.

## Dependencies
- None.

## Notes for the team

- **Use React Testing Library + `@testing-library/user-event`.** Use `user-event` for keyboard interactions (it dispatches the full sequence of events — `keydown` + `keyup` + focus management) rather than `fireEvent.keyDown`. This is the existing convention in the `web/` codebase.
- **Match by accessible name.** `screen.getByRole('tab', { name: /documents/i })` and `screen.getByRole('tab', { name: /user stories/i })` — do not use `getByText` or `getByTestId`.
- **`event.preventDefault()` assertion.** Use `user.keyboard` and verify default scroll behaviour is suppressed, OR (more practically) verify that `onTabChange` was called AND focus is correctly managed — the `preventDefault` is implementation detail. po-ba accepts either approach as long as the scenarios above pass.
- **Audit reference.** `docs/tech_debt.md` line 75 for the baseline coverage numbers.
- **Coverage threshold is 80% (not 95%).** REQ005/US029's BE-test 95% was set because `sqlmock` makes every `repo` branch reachable. FE component branches sometimes hit React internals that are harder to exercise without unrealistic setups. 80% is the project's existing FE-coverage convention for component tests; tester / fe-dev may push higher if achievable.
- **Closes tech-debt.** Strike `docs/tech_debt.md` line 75 in the same commit (or sign-off commit) as the test backfill.
- **Run locally before pushing:**
  - `cd web && npm test -- --watchAll=false TabSwitcher`
  - `cd web && npm test -- --coverage --collectCoverageFrom="components/ProjectDetail/TabSwitcher.tsx"`
  - `cd web && npm run typecheck`
  - `cd web && npm run lint`

## Sign-off log
(po-ba appends here on each sign-off pass)

### Sign-off pass 1 — 2026-06-07 — verdict: approved
- **Spec review:** All 13 AC scenarios are covered. FCT-US036-001…012 map 1:1 onto the behaviour scenarios (click + controlled-component invariant; ArrowRight/ArrowLeft with both wrap edges; Enter; Space; aria-selected; roving tabIndex; prop-driven override with no-callback assertion; tablist semantics incl. `aria-controls`/`id`; unrelated-key no-op). The coverage, no-prod-change, and tech-debt-closure scenarios are proven by the report and the task review log. `Home`/`End` correctly excluded (not in source; explicitly out of scope). E2E legitimately waived per architecture §10.2 — pure component-level behaviour, no API contract. No inappropriate promotion to e2e; pyramid is honest.
- **Result review:** All 12 FCT-* report PASS, 0 FAIL, 0 skipped. Counts match the spec (12 FCT IDs ↔ 12 tested). Full Jest suite 142/142 green, no regressions. Per-file coverage on `TabSwitcher.tsx` = 100% stmts / 100% branches / 100% lines / 100% functions, well above the ≥80% target. `TabSwitcher.tsx` byte-for-byte unchanged (`git diff` empty). `docs/tech_debt.md` line 75 struck through with `→ fixed in REQ006/US036`. The cross-gate semgrep Dockerfile finding is pre-existing and out of scope (verified identical on `main`, filed as tech-debt) — not a US036 defect.
- **Routed to:** none (story done).
