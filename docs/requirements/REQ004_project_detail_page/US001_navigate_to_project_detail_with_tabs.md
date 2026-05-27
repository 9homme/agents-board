# US001 — Navigate to project detail page with tabs

**Requirement:** REQ004 — project_detail_page
**Status:** draft

## Story
As a user browsing the dashboard, I want to click a project card and land on a per-project detail page with a Documents tab and a User Stories tab, so that I have a single place to drill into a project's content.

## Acceptance criteria

- **Scenario: Dashboard project cards are clickable (in-scope for this story)**
  - Given I am on the dashboard at `/` and at least one project card is displayed
  - When I look at any project card
  - Then the card is a focusable, interactive element (functions as a link to `/projects/{id}` for that card's project)
  - And the card is reachable via keyboard Tab navigation
  - And activating the card with Enter or Space navigates to `/projects/{id}`
  - And the card's existing visual design from REQ002 is preserved (clickability is additive, not a redesign)

- **Scenario: Click a project card to open the detail page**
  - Given I am on the dashboard at `/` and at least one project card is displayed
  - When I click a project card
  - Then I navigate to `/projects/{id}` where `{id}` is that project's id
  - And the URL in the address bar reflects that route
  - And the dashboard navigation (browser back) returns me to `/`

- **Scenario: Project detail header shows project info**
  - Given I have navigated to `/projects/{id}` for a project that exists
  - When the page finishes loading
  - Then I see the project's `name` rendered prominently as the page heading
  - And I see the project's `description` rendered below the heading (or a muted "No description" placeholder if the description is empty)

- **Scenario: Two tabs visible, Documents active by default**
  - Given I am on `/projects/{id}` for an existing project
  - When the page finishes loading
  - Then I see a tab switcher with exactly two tabs: "Documents" and "User Stories"
  - And the "Documents" tab is the active/selected tab by default
  - And the URL reflects `?tab=documents` (either implicitly — no query param defaults to documents — or explicitly; the architect/tester picks one and the test contract locks it)

- **Scenario: Switching to the User Stories tab**
  - Given I am on `/projects/{id}` with the Documents tab active
  - When I click the "User Stories" tab
  - Then the User Stories tab becomes the active tab
  - And the URL reflects `?tab=user-stories`
  - And the tab content area displays the **exact verbatim** placeholder text: `Coming soon — user stories will appear here in a future release.`
  - And no network call is made to fetch user stories (placeholder only)

- **Scenario: Switching back to the Documents tab**
  - Given the User Stories tab is active
  - When I click the "Documents" tab
  - Then the Documents tab becomes active and the URL updates to `?tab=documents`
  - (The contents of the Documents tab — list + previewer — are specified by US002 and US003; for US001 the Documents tab content area may be an empty container / placeholder.)

- **Scenario: Refresh preserves the active tab**
  - Given I am on `/projects/{id}?tab=user-stories`
  - When I refresh the browser
  - Then the User Stories tab is still active after the page loads

- **Scenario: Loading state for the project header**
  - Given I navigate to `/projects/{id}` and the `GET /api/v1/projects/{id}` call has not yet resolved
  - When the page is mid-fetch
  - Then I see a loading indicator (skeleton or spinner) in the header area
  - And the tab switcher may already render (it does not depend on project fetch)

- **Scenario: Project not found**
  - Given I navigate to `/projects/{id}` where `{id}` does not match any project (backend returns `404`)
  - When the page resolves the error
  - Then I see a friendly "Project not found" message in place of the header
  - And the tab switcher and tab content are hidden (no point showing tabs for a project that doesn't exist)
  - And there is a clear way back to the dashboard (e.g. a "Back to dashboard" link or button)

- **Scenario: Project fetch fails (network/server error)**
  - Given I navigate to `/projects/{id}` and the backend returns a 500 (or the network fails)
  - When the failure resolves
  - Then I see a friendly "Failed to load project" error message in the header area
  - And a "Back to dashboard" link is visible

## UI / UX flow expectations

- **Entry points:**
  - From `/` (the dashboard): each `ProjectCard` is clickable. The card becomes a link / has a click handler that routes to `/projects/{id}`.
  - Direct URL: a user can also paste / bookmark `/projects/{id}` and land on the detail page.

- **Happy-path flow:**
  1. User is on `/` and sees the project grid.
  2. User clicks a card → browser navigates to `/projects/{id}`.
  3. The detail page mounts. While the project fetch is in flight, the header shows a skeleton; the tab switcher renders immediately with "Documents" active.
  4. The header populates with the project's `name` (large, primary heading) and `description` (subdued, secondary text).
  5. Below the header, two tabs: `[Documents]` (active, underlined or otherwise visually marked) and `[User Stories]`.
  6. The Documents tab content area is owned by US002/US003. For US001 it is just an empty content slot (a placeholder div).
  7. Clicking `[User Stories]` swaps the content area to show the placeholder copy. URL becomes `/projects/{id}?tab=user-stories`. Clicking `[Documents]` restores it.

- **Layout (desktop-first, single column above the tab switcher; tab content can be multi-column as defined per-tab):**
  ```
  +----------------------------------------------------+
  | < Back to dashboard                                |   ← link to /
  |                                                    |
  |  {Project name}                                    |   ← h1, large
  |  {Project description}                             |   ← muted subtitle
  |                                                    |
  |  [ Documents ]  [ User Stories ]                   |   ← tab switcher
  |  ----------------                                  |   ← underline on active
  |                                                    |
  |  ( tab content area )                              |
  |                                                    |
  +----------------------------------------------------+
  ```

- **Empty / loading / error states:**
  - **Loading (header):** skeleton block where the project name and description will go. Tab switcher may already be visible.
  - **Empty (project description):** if `description` is empty/null, show a small muted "No description" line — do NOT collapse silently.
  - **Empty (User Stories tab):** the exact verbatim copy `Coming soon — user stories will appear here in a future release.` (locked in REQ README's Confirmed Decisions).
  - **Empty (Documents tab content):** owned by US002 — for US001's tests, the Documents content area is an empty slot.
  - **404 — project not found:** replace the header area with a "Project not found" message + "Back to dashboard" link. Hide the tab switcher entirely.
  - **5xx / network error on project fetch:** replace the header area with "Failed to load project" + "Back to dashboard" link. Hide the tab switcher.

- **Validation rules visible to the user:** none (no forms in this story).

- **Accessibility expectations:**
  - The tab switcher should be keyboard-operable (arrow keys to switch tabs, Enter to activate, focus visible). It should follow the WAI-ARIA Tabs pattern (`role="tablist"`, `role="tab"`, `role="tabpanel"`, `aria-selected`).
  - Project cards on the dashboard, once clickable, should be keyboard-focusable and activated by Enter / Space — they are functionally links.
  - The `Back to dashboard` link should be a real `<a>`/`<Link>` (not a button).

- **Out of UI scope:**
  - Documents tab list/previewer (US002, US003).
  - Real User Stories content (future requirement).
  - Mobile-specific layout polish.
  - Animations / transitions between tabs.

## Out of scope
- Anything inside the Documents tab content (sidebar, previewer, markdown rendering) — that's US002 and US003.
- Real User Stories tab content.
- Editing project fields from the detail page.
- Deep-linking to specific documents (`?doc=<id>`) — that's US002.

## Dependencies
- **REQ002 / US001** — project list page and `ProjectCard` component already exist. This story modifies `ProjectCard` to be clickable.
- **New backend endpoint:** `GET /api/v1/projects/{id}` on `api-server`. System Architect will design the exact JSON; expected shape mirrors the existing list endpoint's per-project object (`id`, `name`, `description`, `createdAt`, `updatedAt`), with a `404` body matching REQ002's error model (`{ "code": "...", "message": "..." }`).

## Notes for the team
- The `ProjectCard` is currently in `web/components/Dashboard/ProjectCard.tsx`. Making it clickable should preserve its existing visual design (REQ002 sign-off cited "minimal beautiful") — clickability is additive.
- Use Next.js Pages Router dynamic routing: `web/pages/projects/[id].tsx`. **CSR only** — no `getServerSideProps`, no `getStaticProps`. The page reads `id` from `useRouter().query.id` and fetches via the existing API client pattern (`web/lib/api/...` + a hook like `useProject(id)`).
- The tab switcher is reusable in spirit — but unless we have another consumer in sight, build it inline / local to the page. The tester and tech-lead will decide whether to extract a `Tabs` component during planning.
- Query-string-as-tab-state is intentional (not React state alone) so refresh + share-links preserve which tab the user was on. Implement via `router.replace` with `shallow: true` when switching tabs to avoid re-running data fetches.
- The "Back to dashboard" link is mandatory in error states because in a 404 / 5xx the user would otherwise be stuck on a detail page with no navigation.

## Sign-off log
(po-ba appends here on each sign-off pass)
