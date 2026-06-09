# US005 — User story detail + tasks in a side drawer on card click

**Requirement:** REQ007 — User Stories Tab + E2E Quality Gate + Health-Check Fixes
**Status:** in_signoff

## Story
As a user, I want to click a user-story card and see the full story detail together with all of its tasks in a right-side drawer — with the card list still visible behind it — so that I can drill into a story and its breakdown without losing my place in the list.

## Acceptance criteria
- **Scenario: clicking a card opens a right-side detail drawer (happy path)**
  - Given the User Stories tab is showing the card list (US004)
  - When the user clicks a story card
  - Then a drawer/panel slides in from the right showing the detail for that story
  - And the card list remains visible behind the drawer (the list is NOT replaced or unmounted)
  - And the drawer shows the story's **title**, **status**, and full **description**
  - And the drawer shows a list of **all tasks** for that story
  - And each task shows its **title**, **status**, and **description**
- **Scenario: story with no tasks**
  - Given the selected story has zero tasks
  - When the drawer renders
  - Then the story detail (title/status/description) is shown
  - And a tasks empty-state is displayed (e.g. "No tasks for this story.")
- **Scenario: closing the drawer returns to the card list**
  - Given the detail drawer is open
  - When the user activates the drawer's **close** control (e.g. an "X" / close button)
  - Then the drawer closes
  - And the card list is fully interactive again with no full reload
  - And no story is selected
- **Scenario: switching stories without closing**
  - Given the detail drawer is open for one story
  - When the user clicks a different card behind/beside the drawer
  - Then the drawer updates to show the newly selected story's detail and tasks
  - (the drawer stays open; selection moves to the new story)
- **Scenario: detail loading state**
  - Given the user clicked a card
  - When the story-detail and/or tasks requests are in flight
  - Then a loading indicator is shown inside the drawer
- **Scenario: detail error state**
  - Given the story-detail or tasks request fails
  - When the request settles
  - Then an error message is shown inside the drawer (e.g. "Couldn't load this user story.")
  - And the drawer's close control is still available
- **Scenario: keyboard accessibility**
  - Given the detail drawer is open
  - When the user navigates by keyboard
  - Then the close control is focusable and activatable
  - And focus is managed sensibly on open (focus moves into the drawer, e.g. to the heading or close control)
  - And the drawer is dismissible via the Escape key (in addition to the close button)

## UI / UX flow expectations
- **Entry points:** clicking/activating a card in the User Stories tab list (US004).
- **Happy-path flow:**
  1. From the card list, user clicks a story card.
  2. A drawer/panel slides in from the right (D-006 REVISED — side panel, NOT an in-tab replacement, NOT a separate route). The card list stays visible behind it.
  3. Drawer shows title, status, and full description at the top.
  4. Below, a list of all tasks, each showing title, status badge, and description.
  5. User clicks the close button (or presses Escape) to dismiss the drawer and return focus to the list.
- **Empty / loading / error states:**
  - Loading: spinner/skeleton inside the drawer while fetching story + tasks.
  - Tasks empty: "No tasks for this story." under the story detail.
  - Error: "Couldn't load this user story." inside the drawer, with the close control still present.
- **Validation rules visible to the user:** none (read-only).
- **Out of UI scope:** editing story/task; status changes; reordering; deep-linking to a story via URL (in-tab selection state only for v1); styling/animation polish of the drawer transition.

## Out of scope
- Editing, creating, deleting stories or tasks (read-only).
- URL/deep-link to a specific story (in-tab selection state only — D-006).
- Filtering/sorting tasks (D-008).
- Real-time updates / SSE refresh of the drawer.

## Dependencies
- **US004** — the card list and selection mechanism this drawer is reached from.
- **Architecture (D-007):** REST endpoints for a single story's detail and its tasks. Proposed `GET /api/v1/user-stories/{id}` (bare `UserStory` object) and `GET /api/v1/user-stories/{id}/tasks` (`{"tasks":[Task,...]}`) — **System Architect finalises exact contract**, including whether tasks are embedded in the story-detail response or fetched separately. Single-resource responses follow the existing bare-object convention; list responses are wrapped.
- New `web/lib/api/` functions + types + MSW handlers mirroring the contract.

## Notes for the team
- **D-006 REVISED — CONFIRMED:** detail is a **right-side drawer/panel**, NOT an in-tab view that replaces the grid and NOT a dedicated route. The card list stays visible behind the panel; the panel has a **close button** (Escape also closes it).
- Selecting a different card while the drawer is open updates the drawer in place (selection state lifted from US004 drives which story the drawer shows).
- Domain shapes exist: `UserStory{id,projectId,title,description,status,createdAt,updatedAt}`, `Task{id,userStoryId,title,description,status,createdAt,updatedAt}`.
- If the architect chooses to embed tasks in the story-detail response, the FE makes one request; if separate, two — the tester's specs and MSW handlers must match the chosen contract exactly.
- Reuse the loading/empty/error patterns established in US004 and `DocumentsTab.tsx`. Consider an accessible drawer/dialog pattern (role, focus trap or sensible focus management, Escape-to-close).

## Sign-off log
(po-ba appends here on each sign-off pass)
