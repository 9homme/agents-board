# US001 — Frontend component test specification

**For FE Dev:** these are the tests you write FIRST (TDD red). Implement in TypeScript using **Jest + React Testing Library**. Mock the backend at the API client layer (`web/lib/api/`) using **MSW** with handlers that match the architecture's exact JSON request/response shapes.

## Coverage matrix

| AC / UI flow | Test ID | Component / hook under test | What it asserts |
|---|---|---|---|
| ProjectCard renders as a link to `/projects/{id}` | FCT-US001-001 | `web/components/Dashboard/ProjectCard.tsx` | wraps content in `<a>` pointing to `/projects/{id}` |
| ProjectCard keyboard: Tab-focusable, Enter/Space navigates | FCT-US001-002 | `web/components/Dashboard/ProjectCard.tsx` | role="link", keyboard activation routes correctly |
| ProjectCard middle-click / right-click not hijacked | FCT-US001-003 | `web/components/Dashboard/ProjectCard.tsx` | uses Next `<Link>` (not a plain click handler) so native browser behavior is preserved |
| ProjectCard preserves existing visual classes | FCT-US001-004 | `web/components/Dashboard/ProjectCard.tsx` | existing CSS class names still present on the inner `<article>` |
| Detail page: header renders project name and description | FCT-US001-005 | `web/pages/projects/[id].tsx` + `web/components/ProjectDetail/ProjectHeader.tsx` | `<h1>` contains project name; description text visible |
| Detail page: empty description shows "No description" | FCT-US001-006 | `web/components/ProjectDetail/ProjectHeader.tsx` | renders "No description" placeholder when `description === ""` |
| Detail page: loading skeleton shown during project fetch | FCT-US001-007 | `web/pages/projects/[id].tsx` | skeleton/loading indicator visible while MSW is pending |
| Tab switcher: two tabs, Documents active by default | FCT-US001-008 | `web/components/ProjectDetail/TabSwitcher.tsx` | two tab elements; "Documents" has `aria-selected="true"` when no `?tab=` param |
| Tab switcher: click "User Stories" activates it and updates URL | FCT-US001-009 | `web/components/ProjectDetail/TabSwitcher.tsx` | after click, "User Stories" `aria-selected="true"`; `router.replace` called with `?tab=user-stories` |
| Tab switcher: click "Documents" re-activates it | FCT-US001-010 | `web/components/ProjectDetail/TabSwitcher.tsx` | clicking "Documents" from User Stories state sets it active; URL updated to `?tab=documents` |
| Tab switcher: `?tab=user-stories` query param on mount activates User Stories | FCT-US001-011 | `web/pages/projects/[id].tsx` | rendering with `query.tab = "user-stories"` shows User Stories tab as active |
| User Stories tab: verbatim placeholder text | FCT-US001-012 | `web/components/ProjectDetail/UserStoriesTab.tsx` | renders exact string "Coming soon — user stories will appear here in a future release." |
| User Stories tab: no network calls made | FCT-US001-013 | `web/components/ProjectDetail/UserStoriesTab.tsx` | MSW receives zero requests when only UserStoriesTab is rendered |
| Detail page: 404 renders "Project not found" + "Back to dashboard" | FCT-US001-014 | `web/pages/projects/[id].tsx` | "Project not found" text visible; tab switcher hidden; "Back to dashboard" link to "/" present |
| Detail page: 500 renders "Failed to load project" + "Back to dashboard" | FCT-US001-015 | `web/pages/projects/[id].tsx` | "Failed to load project" text visible; "Back to dashboard" link present |

## Component tests

### FCT-US001-001 — ProjectCard renders as a Next.js Link to `/projects/{id}`
- **Component / hook under test:** `web/components/Dashboard/ProjectCard.tsx`
- **Render with:** a project fixture `{ id: "proj-001", name: "Test Project", description: "desc", createdAt: "2026-05-20T10:00:00Z", updatedAt: "2026-05-20T10:00:00Z" }`; wrap in a Next.js Router mock (e.g. `jest-mock-next-router` or `createMockRouter` with `RouterContext`).
- **MSW handlers:** none required for this test.
- **User interactions (RTL):** none.
- **Expect:**
  - `screen.getByRole('link', { name: /Test Project/i })` exists.
  - The link element's `href` attribute equals `/projects/proj-001`.
- **Architecture cite:** FE surface `web/components/Dashboard/ProjectCard.tsx`; §"State strategy" — "ProjectCard becomes a Next `<Link href={/projects/${project.id}}>`".

### FCT-US001-002 — ProjectCard: keyboard-focusable and activatable
- **Component / hook under test:** `web/components/Dashboard/ProjectCard.tsx`
- **Render with:** same fixture as FCT-US001-001; mock router.
- **MSW handlers:** none.
- **User interactions (RTL):**
  1. `userEvent.tab()` to move focus to the card link.
  2. Assert `document.activeElement` is the link element (or contains the link).
- **Expect:**
  - The link is reachable via Tab.
  - The link has a visible focus indicator (CSS class or `outline` — assert that the element has a `focus` style; this can be checked structurally via the presence of a focus-ring class or by asserting the element is the tab stop).
- **Notes:** The "Enter activates navigation" behavior is native to `<a>` elements and does not need a synthetic `keyDown` simulation — asserting the `href` is correct (FCT-US001-001) is sufficient for keyboard navigation correctness. If the team adds a `onKeyDown` handler, add a test for it.
- **Architecture cite:** §"State strategy" — "The Link supplies focusability, Enter activation … for free"; FE surface `ProjectCard.tsx`.

### FCT-US001-003 — ProjectCard: uses Next Link (not a plain onClick) so middle-click / right-click behavior is native
- **Component / hook under test:** `web/components/Dashboard/ProjectCard.tsx`
- **Render with:** same fixture.
- **MSW handlers:** none.
- **Expect:**
  - The rendered output contains an `<a>` element (not a `<div>` with an `onClick` handler).
  - The `<a>` element has no `onClick` prop that calls `router.push` directly (i.e. the navigation is delegated to the Link's native anchor, not to a manual click handler).
- **Notes:** This is a structural assertion. Middle-click and right-click opening a new tab are native browser behaviors for `<a>` elements; we cannot simulate them in JSDOM, but confirming the structure is `<a href=...>` (not `<div onClick=...>`) is sufficient to guarantee the behavior.
- **Architecture cite:** §"State strategy" — "right-click/middle-click correctness for free".

### FCT-US001-004 — ProjectCard: existing visual classes preserved on the inner `<article>`
- **Component / hook under test:** `web/components/Dashboard/ProjectCard.tsx`
- **Render with:** same fixture.
- **MSW handlers:** none.
- **Expect:**
  - The inner `<article>` element retains the CSS class names it had before this story's changes (snapshot the class list from the REQ002 implementation).
  - The new `<Link>` wrapper does NOT strip or replace any classes that were on the card in REQ002.
- **Notes:** Take a class-list snapshot in the test or assert specific class names known from REQ002. The goal is a non-regression guard — clickability is additive.
- **Architecture cite:** §"State strategy" — "The article retains its visual classes".

### FCT-US001-005 — Detail page: project header renders name and description
- **Component / hook under test:** `web/pages/projects/[id].tsx` + `web/components/ProjectDetail/ProjectHeader.tsx`
- **Render with:**
  - `query = { id: "proj-001" }` (no `tab`, no `doc`)
  - MSW handler for `GET /api/v1/projects/proj-001` → 200
- **MSW handlers:**
  ```json
  GET /api/v1/projects/proj-001 → 200
  {
    "id": "proj-001",
    "name": "E-commerce Website",
    "description": "A new online store for electronics",
    "createdAt": "2026-05-20T10:00:00Z",
    "updatedAt": "2026-05-20T10:00:00Z"
  }
  ```
- **User interactions (RTL):** wait for loading to resolve.
- **Expect:**
  - `screen.findByRole('heading', { level: 1, name: /E-commerce Website/i })` is visible.
  - `screen.findByText(/A new online store for electronics/i)` is visible.
- **Architecture cite:** API contract §1; FE surface `web/components/ProjectDetail/ProjectHeader.tsx`.

### FCT-US001-006 — ProjectHeader: empty description shows "No description" placeholder
- **Component / hook under test:** `web/components/ProjectDetail/ProjectHeader.tsx`
- **Render with:** `project = { id: "p1", name: "Project Alpha", description: "", createdAt: "2026-05-20T10:00:00Z", updatedAt: "2026-05-20T10:00:00Z" }`.
- **MSW handlers:** none (render component directly with props, not via page).
- **User interactions (RTL):** none.
- **Expect:**
  - `screen.getByText(/No description/i)` is visible.
  - The empty string `""` is NOT rendered as a blank line (i.e. the placeholder is shown, not an empty `<p></p>`).
- **Architecture cite:** API contract §1 — "`description` MAY be `""` — never `null`; the FE shows 'No description' placeholder when empty"; §"State strategy".

### FCT-US001-007 — Detail page: loading skeleton visible during project fetch
- **Component / hook under test:** `web/pages/projects/[id].tsx`
- **Render with:**
  - `query = { id: "proj-001" }`
  - MSW handler for `GET /api/v1/projects/proj-001` configured with an indefinite delay.
- **MSW handlers:**
  - `GET /api/v1/projects/proj-001` → delayed (use `http.delay('infinite')` or equivalent).
- **User interactions (RTL):** render and immediately assert (do not await).
- **Expect:**
  - A loading indicator (skeleton element, spinner, or element with a loading-related `aria-*` attribute) is present in the header area before the fetch resolves.
  - The project name `"E-commerce Website"` is NOT yet present in the DOM.
- **Architecture cite:** US001 AC "Loading state for the project header"; §"Empty / loading / error states".

### FCT-US001-008 — TabSwitcher: two tabs, Documents active by default (no `?tab=` param)
- **Component / hook under test:** `web/components/ProjectDetail/TabSwitcher.tsx`
- **Render with:**
  - Props: `activeTab="documents"`, `onTabChange={jest.fn()}` (or however the component receives its active-tab state — adjust to the implementation).
  - Alternatively, render `web/pages/projects/[id].tsx` with `query = { id: "p1" }` (no `tab` param) and a successful project MSW response.
- **MSW handlers:** `GET /api/v1/projects/p1` → 200 (any valid project JSON per architecture §1).
- **User interactions (RTL):** none.
- **Expect:**
  - Two elements with `role="tab"` are rendered.
  - The tab with accessible name matching `/Documents/i` has `aria-selected="true"`.
  - The tab with accessible name matching `/User Stories/i` has `aria-selected="false"`.
  - The tab container has `role="tablist"`.
- **Architecture cite:** §"Components" `TabSwitcher.tsx`; §"Accessibility cross-cutting" — WAI-ARIA Tabs pattern.

### FCT-US001-009 — TabSwitcher: clicking "User Stories" activates it and calls `router.replace` with `?tab=user-stories`
- **Component / hook under test:** `web/components/ProjectDetail/TabSwitcher.tsx` (or via the page)
- **Render with:** active tab = `"documents"`; mock `router.replace`.
- **MSW handlers:** `GET /api/v1/projects/p1` → 200.
- **User interactions (RTL):**
  1. `userEvent.click(screen.getByRole('tab', { name: /User Stories/i }))`
- **Expect:**
  - `router.replace` was called with a query object containing `tab: 'user-stories'` and `shallow: true`.
  - OR: if the component is rendered in a full-page context with a Next.js router mock that tracks pushed routes, assert that `query.tab === 'user-stories'` after the click.
  - The tab with name `/User Stories/i` now has `aria-selected="true"`.
- **Architecture cite:** §"State strategy" — "Tab changes are written with `router.replace({ query: {...query, tab: 'user-stories'} }, undefined, { shallow: true })`".

### FCT-US001-010 — TabSwitcher: clicking "Documents" from User Stories state re-activates it
- **Component / hook under test:** `web/components/ProjectDetail/TabSwitcher.tsx`
- **Render with:** active tab = `"user-stories"`.
- **MSW handlers:** minimal project response.
- **User interactions (RTL):**
  1. `userEvent.click(screen.getByRole('tab', { name: /Documents/i }))`
- **Expect:**
  - `router.replace` called with `tab: 'documents'`, `shallow: true`.
  - "Documents" tab has `aria-selected="true"`.
- **Architecture cite:** US001 AC "Switching back to the Documents tab".

### FCT-US001-011 — Detail page: `?tab=user-stories` in query activates User Stories tab on mount
- **Component / hook under test:** `web/pages/projects/[id].tsx`
- **Render with:** `query = { id: "p1", tab: "user-stories" }`; project MSW response 200.
- **MSW handlers:** `GET /api/v1/projects/p1` → 200 (any valid project per §1).
- **User interactions (RTL):** await page load.
- **Expect:**
  - The tab with name `/User Stories/i` has `aria-selected="true"`.
  - The User Stories tab content area contains the verbatim string `"Coming soon — user stories will appear here in a future release."`.
- **Architecture cite:** US001 AC "Refresh preserves the active tab"; §"State strategy" — URL is source of truth for navigation state.

### FCT-US001-012 — UserStoriesTab: renders exact verbatim placeholder text
- **Component / hook under test:** `web/components/ProjectDetail/UserStoriesTab.tsx`
- **Render with:** no props required.
- **MSW handlers:** none.
- **User interactions (RTL):** none.
- **Expect:**
  - `screen.getByText('Coming soon — user stories will appear here in a future release.')` is present and visible.
  - The text matches character-for-character (em dash `—`, not a hyphen `-`).
- **Architecture cite:** US001 AC "Switching to the User Stories tab"; §"Components" `UserStoriesTab.tsx` — "Renders the exact verbatim string".

### FCT-US001-013 — UserStoriesTab: no network requests made
- **Component / hook under test:** `web/components/ProjectDetail/UserStoriesTab.tsx`
- **Render with:** no props; install MSW in request-recording mode.
- **User interactions (RTL):** render and wait one tick.
- **Expect:**
  - MSW captures zero HTTP requests.
- **Architecture cite:** US001 AC "no network call is made to fetch user stories (placeholder only)"; §"Components" `UserStoriesTab.tsx` — "No network calls".

### FCT-US001-014 — Detail page: 404 response renders "Project not found" + "Back to dashboard" link; tabs hidden
- **Component / hook under test:** `web/pages/projects/[id].tsx`
- **Render with:** `query = { id: "no-such-project" }`.
- **MSW handlers:**
  ```json
  GET /api/v1/projects/no-such-project → 404
  { "code": "NOT_FOUND", "message": "Project not found" }
  ```
- **User interactions (RTL):** await page load.
- **Expect:**
  - `screen.findByText(/Project not found/i)` is visible.
  - `screen.queryByRole('tablist')` returns `null` (tab switcher is hidden).
  - `screen.findByRole('link', { name: /Back to dashboard/i })` has `href="/"`.
- **Architecture cite:** US001 AC "Project not found"; §"State strategy" — `useProject` discriminates `error.code === 'NOT_FOUND'`; API contract §1, 404 envelope `{"code":"NOT_FOUND","message":"Project not found"}`.

### FCT-US001-015 — Detail page: 500 response renders "Failed to load project" + "Back to dashboard" link
- **Component / hook under test:** `web/pages/projects/[id].tsx`
- **Render with:** `query = { id: "broken-project" }`.
- **MSW handlers:**
  ```json
  GET /api/v1/projects/broken-project → 500
  { "code": "INTERNAL_ERROR", "message": "Failed to fetch project" }
  ```
- **User interactions (RTL):** await page load.
- **Expect:**
  - `screen.findByText(/Failed to load project/i)` is visible.
  - `screen.findByRole('link', { name: /Back to dashboard/i })` present.
- **Architecture cite:** US001 AC "Project fetch fails (network/server error)"; API contract §1, 500 envelope `{"code":"INTERNAL_ERROR","message":"Failed to fetch project"}`.
