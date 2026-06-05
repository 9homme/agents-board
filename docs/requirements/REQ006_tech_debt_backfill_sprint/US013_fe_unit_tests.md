# US013 — Frontend component test specification
# `TabSwitcher.tsx` coverage backfill

**For FE Dev:** these are the tests you write FIRST (TDD red). Implement in TypeScript using **Jest + React Testing Library + `@testing-library/user-event`**. Tests live in `web/components/ProjectDetail/TabSwitcher.test.tsx` (edit existing file). Production code in `TabSwitcher.tsx` is **byte-for-byte unchanged** — if a test surfaces a real bug, raise `ARCHITECTURE_GAP_FOUND`.

**Key tooling note (architecture.md §7.4):** use `@testing-library/user-event` for keyboard events (dispatches the full `keydown → keyup` sequence with focus management). `fireEvent.keyDown` alone skips focus updates and will not correctly exercise the `ArrowRight`/`ArrowLeft` roving-focus logic. Match elements by `getByRole('tab', { name: /documents/i })` — not `getByText`, not `getByTestId`.

**Coverage target:** ≥80% stmts AND ≥80% branches AND ≥80% lines AND ≥80% functions (architecture.md §7.3, D-009).

**NOT implemented (do not write tests for):** `Home`, `End`, `Tab` (browser default), `Escape`. The component does not handle them (architecture.md §7.1).

**FCT IDs are pinned by architecture.md §7.2 — do NOT renumber.**

## Coverage matrix

| AC scenario | Test ID | Component under test | What it asserts |
|---|---|---|---|
| Click non-active tab fires `onTabChange` | FCT-US013-001 | `TabSwitcher` | callback called once with clicked id; controlled component does not mutate `activeTab` |
| `ArrowRight` moves focus forward and fires `onTabChange` | FCT-US013-002 | `TabSwitcher` | focus on next tab, callback called with next tab id |
| `ArrowRight` from last tab wraps to first | FCT-US013-003 | `TabSwitcher` | modulo wrap; callback called with first tab id |
| `ArrowLeft` moves focus backward and fires `onTabChange` | FCT-US013-004 | `TabSwitcher` | focus on previous tab, callback called with previous tab id |
| `ArrowLeft` from first tab wraps to last | FCT-US013-005 | `TabSwitcher` | modulo wrap; callback called with last tab id |
| `Enter` activates focused tab; calls `preventDefault` | FCT-US013-006 | `TabSwitcher` | callback called for currently focused tab |
| `Space` activates focused tab; calls `preventDefault` | FCT-US013-007 | `TabSwitcher` | callback called for currently focused tab |
| `aria-selected` reflects `activeTab` prop | FCT-US013-008 | `TabSwitcher` | correct `aria-selected` on each tab button |
| Roving `tabIndex` reflects `activeTab` prop | FCT-US013-009 | `TabSwitcher` | `tabIndex={0}` on active, `tabIndex={-1}` on inactive |
| Prop-driven `activeTab` change re-renders; no `onTabChange` fired | FCT-US013-010 | `TabSwitcher` | new active tab rendered; callback NOT called by re-render |
| Tablist semantics present | FCT-US013-011 | `TabSwitcher` | `role="tablist"`, `aria-label`, `aria-controls`, `id` attributes |
| Unrelated keys do not fire `onTabChange` or `preventDefault` | FCT-US013-012 | `TabSwitcher` | `onTabChange` not called; `preventDefault` not called |

## Component tests

### FCT-US013-001 — Clicking non-active tab fires `onTabChange` with clicked id
- **Component:** `web/components/ProjectDetail/TabSwitcher.tsx`
- **Render with:**
  ```tsx
  const onTabChange = jest.fn()
  render(<TabSwitcher activeTab="documents" onTabChange={onTabChange} />)
  ```
- **User interaction:**
  ```tsx
  const user = userEvent.setup()
  await user.click(screen.getByRole('tab', { name: /user stories/i }))
  ```
- **Expect:**
  - `onTabChange` called exactly once
  - `onTabChange` called with `"user-stories"` (the tab id for the User Stories tab — confirm exact id from component source)
  - The Documents tab still has `aria-selected="true"` (component is controlled; it does NOT update `activeTab` internally)
- **Architecture cite:** architecture.md §7.2 FCT-US013-001

---

### FCT-US013-002 — `ArrowRight` moves focus forward and fires `onTabChange`
- **Component:** `TabSwitcher`
- **Render with:** `activeTab="documents"`
- **User interaction:**
  ```tsx
  const user = userEvent.setup()
  const docsTab = screen.getByRole('tab', { name: /documents/i })
  await user.click(docsTab) // focus the Documents tab
  await user.keyboard('{ArrowRight}')
  ```
- **Expect:**
  - `onTabChange` called with `"user-stories"`
  - Focus is now on the User Stories tab button (`document.activeElement === screen.getByRole('tab', { name: /user stories/i })`)
- **Architecture cite:** architecture.md §7.2 FCT-US013-002

---

### FCT-US013-003 — `ArrowRight` from last tab wraps to first
- **Component:** `TabSwitcher`
- **Render with:** `activeTab="user-stories"`
- **User interaction:**
  ```tsx
  await user.click(screen.getByRole('tab', { name: /user stories/i }))
  await user.keyboard('{ArrowRight}')
  ```
- **Expect:**
  - `onTabChange` called with `"documents"` (wraps back to first)
  - Focus is on the Documents tab
- **Architecture cite:** architecture.md §7.2 FCT-US013-003; architecture.md §7.1 modulo arithmetic

---

### FCT-US013-004 — `ArrowLeft` moves focus backward and fires `onTabChange`
- **Component:** `TabSwitcher`
- **Render with:** `activeTab="user-stories"`
- **User interaction:**
  ```tsx
  await user.click(screen.getByRole('tab', { name: /user stories/i }))
  await user.keyboard('{ArrowLeft}')
  ```
- **Expect:**
  - `onTabChange` called with `"documents"`
  - Focus is on the Documents tab
- **Architecture cite:** architecture.md §7.2 FCT-US013-004

---

### FCT-US013-005 — `ArrowLeft` from first tab wraps to last
- **Component:** `TabSwitcher`
- **Render with:** `activeTab="documents"`
- **User interaction:**
  ```tsx
  await user.click(screen.getByRole('tab', { name: /documents/i }))
  await user.keyboard('{ArrowLeft}')
  ```
- **Expect:**
  - `onTabChange` called with `"user-stories"` (wraps to last)
  - Focus is on the User Stories tab
- **Architecture cite:** architecture.md §7.2 FCT-US013-005

---

### FCT-US013-006 — `Enter` activates focused tab; calls `preventDefault`
- **Component:** `TabSwitcher`
- **Render with:** `activeTab="documents"`, User Stories tab focused via prior `ArrowRight`
- **User interaction:**
  ```tsx
  await user.click(screen.getByRole('tab', { name: /documents/i }))
  await user.keyboard('{ArrowRight}') // focus moves to User Stories
  await user.keyboard('{Enter}')
  ```
- **Expect:**
  - `onTabChange` called with `"user-stories"` (the currently focused tab)
  - `onTabChange` was called (the test can verify this by count — once for ArrowRight, once for Enter; total 2 calls on `onTabChange` with appropriate args, OR reset mock after ArrowRight and assert 1 call for Enter only)
- **Note on `preventDefault` assertion:** verifying `event.preventDefault()` via `@testing-library/user-event` is implicit — user-event fires the full event sequence. The assertion that `onTabChange` was called AND focus did not blur the tab confirms the `preventDefault` contract per architecture.md §7.4.
- **Architecture cite:** architecture.md §7.2 FCT-US013-006

---

### FCT-US013-007 — `Space` activates focused tab; calls `preventDefault`
- **Component:** `TabSwitcher`
- **Render with:** same as FCT-US013-006
- **User interaction:**
  ```tsx
  await user.click(screen.getByRole('tab', { name: /documents/i }))
  await user.keyboard('{ArrowRight}') // User Stories focused
  await user.keyboard('{ }') // Space key
  ```
- **Expect:**
  - `onTabChange` called for the focused tab's id
  - Scroll behaviour suppressed (implicit via user-event's `preventDefault` dispatch)
- **Architecture cite:** architecture.md §7.2 FCT-US013-007

---

### FCT-US013-008 — `aria-selected` reflects `activeTab` prop
- **Component:** `TabSwitcher`
- **Render with:** `activeTab="documents"`
- **Expect (DOM assertions, no user interaction needed):**
  - `screen.getByRole('tab', { name: /documents/i })` has `aria-selected="true"`
  - `screen.getByRole('tab', { name: /user stories/i })` has `aria-selected="false"`
- **Architecture cite:** architecture.md §7.2 FCT-US013-008; architecture.md §7.1 ARIA attributes

---

### FCT-US013-009 — Roving `tabIndex` reflects `activeTab` prop
- **Component:** `TabSwitcher`
- **Render with:** `activeTab="documents"`
- **Expect:**
  - Documents tab has `tabIndex={0}`
  - User Stories tab has `tabIndex={-1}`
- **Architecture cite:** architecture.md §7.2 FCT-US013-009; architecture.md §7.1 roving tabindex

---

### FCT-US013-010 — Prop-driven `activeTab` change re-renders; no `onTabChange` fired
- **Component:** `TabSwitcher`
- **Render with initial props:** `activeTab="documents"`
- **Re-render with new props:** `rerender(<TabSwitcher activeTab="user-stories" onTabChange={onTabChange} />)`
- **Expect:**
  - User Stories tab now has `aria-selected="true"` and `tabIndex={0}`
  - Documents tab now has `aria-selected="false"` and `tabIndex={-1}`
  - `onTabChange` was NOT called at all (prop-driven re-render does not fire the callback)
- **Architecture cite:** architecture.md §7.2 FCT-US013-010

---

### FCT-US013-011 — Tablist semantics are present
- **Component:** `TabSwitcher`
- **Render with:** any valid props (use `activeTab="documents"`)
- **Expect:**
  - A single element with `role="tablist"` AND `aria-label="Project tabs"` exists in the DOM
  - Exactly 2 elements with `role="tab"` exist
  - The Documents tab has `aria-controls="tabpanel-documents"` AND `id="tab-documents"`
  - The User Stories tab has `aria-controls="tabpanel-user-stories"` AND `id="tab-user-stories"`
  - (confirm exact id strings by reading `TabSwitcher.tsx` source — architecture.md §7.1 confirms these attribute patterns)
- **Architecture cite:** architecture.md §7.2 FCT-US013-011; architecture.md §7.1 ARIA attributes

---

### FCT-US013-012 — Unrelated keys do not fire `onTabChange` or `preventDefault`
- **Component:** `TabSwitcher`
- **Render with:** `activeTab="documents"`, Documents tab focused
- **User interaction:** press `Escape`, `a`, `Tab` (in sequence via user-event)
  ```tsx
  await user.click(screen.getByRole('tab', { name: /documents/i }))
  await user.keyboard('{Escape}a{Tab}')
  ```
- **Expect:**
  - `onTabChange` NOT called at all
- **Note:** `Tab` is handled by the browser's native focus management; user-event will move focus away from the tablist but `onTabChange` must not fire.
- **Architecture cite:** architecture.md §7.2 FCT-US013-012; architecture.md §7.1 NOT IMPLEMENTED list

## Coverage verification

```bash
cd web && npm test -- --coverage --collectCoverageFrom="components/ProjectDetail/TabSwitcher.tsx" --watchAll=false
```

Target: ≥80% stmts AND ≥80% branches AND ≥80% lines AND ≥80% functions.

## Coverage exemptions

None anticipated. 12 tests across all 4 keyboard handlers, click handler, and prop-driven re-render should comfortably exceed 80% on an 82-line component. If any JSX wrapper or `className` computation line remains uncovered, document under OQ-4 in the test report.
