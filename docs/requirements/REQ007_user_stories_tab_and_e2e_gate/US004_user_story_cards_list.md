# US004 — User Stories tab — story cards list

**Requirement:** REQ007 — User Stories Tab + E2E Quality Gate + Health-Check Fixes
**Status:** done

## Story
As a user viewing a project, I want the User Stories tab to show a grid of cards — one per user story — each summarising the story, so that I can scan all stories in a project at a glance and pick one to drill into.

## Acceptance criteria
- **Scenario: cards render for a project's user stories (happy path)**
  - Given a project that has one or more user stories
  - When the user opens the User Stories tab
  - Then the tab fetches the project's user stories from the backend
  - And renders one card per user story
  - And each card shows the story **title**, a **status** badge, a **task count** ("N tasks"), and a **description preview** (the first ~80 characters of the description)
- **Scenario: description preview is truncated**
  - Given a story whose description is longer than ~80 characters
  - When its card renders
  - Then the card shows only the first ~80 characters of the description, with a truncation indicator (e.g. an ellipsis)
  - And a story with a short or empty description renders its full description (or nothing) without a truncation indicator
- **Scenario: empty state**
  - Given a project that has zero user stories
  - When the user opens the User Stories tab
  - Then no cards are shown
  - And an empty-state message is displayed (e.g. "No user stories yet for this project.")
- **Scenario: loading state**
  - Given the user opens the User Stories tab
  - When the user-stories request is in flight
  - Then a loading indicator is shown in the tab panel
  - And no stale/placeholder card content is shown
- **Scenario: error state**
  - Given the user-stories request fails (network or non-2xx)
  - When the tab finishes attempting the fetch
  - Then an error message is shown (e.g. "Couldn't load user stories.")
  - And no cards are rendered
- **Scenario: cards are keyboard- and click-activatable (prepares US005)**
  - Given cards are rendered
  - When the user clicks a card (or activates it via keyboard)
  - Then the card invokes the selection handler with that story's id (the detail view itself is US005)
  - And each card is a focusable, accessible control (role/button semantics, accessible name = story title)

## UI / UX flow expectations
- **Entry points:** the existing "User Stories" tab on the project detail page (`web/components/ProjectDetail/`), currently a "Coming soon" placeholder. This story replaces the placeholder.
- **Happy-path flow:**
  1. User is on a project detail page and clicks the "User Stories" tab.
  2. Tab panel shows a loading indicator while fetching.
  3. Cards appear in a responsive grid, each showing title, status badge, "N tasks", and a ~80-char description preview.
  4. Hovering/focusing a card shows it is interactive; clicking selects it (drill-in handled by US005).
- **Empty / loading / error states:**
  - Loading: spinner / skeleton inside the tab panel.
  - Empty: friendly "No user stories yet for this project." message, no cards.
  - Error: "Couldn't load user stories." with no cards (a retry affordance is welcome but not required for v1).
- **Validation rules visible to the user:** none (read-only view; no inputs).
- **Out of UI scope:** filtering, sorting, search (D-008); pagination; the detail/drill-in view (US005); visual styling polish beyond clear card/badge legibility.

## Out of scope
- The detail panel and tasks list (US005).
- Any create/edit/delete of stories (read-only).
- Filtering, sorting, searching, pagination (D-008).

## Dependencies
- **Architecture (D-007):** a new REST endpoint returning a project's user stories with enough data to render the card, including the **task count** and the story **description** (the card shows a ~80-char preview). Proposed `GET /api/v1/projects/{id}/user-stories` returning `{"userStories":[{id,projectId,title,description,status,taskCount,...}]}` — **the System Architect must finalise the exact contract**, including how `taskCount` is provided (embedded on the list item vs. derived) and whether the full `description` is returned (preferred — truncation to ~80 chars is an FE concern). po-ba flags `taskCount` and `description` as card requirements so the architect designs for them rather than forcing an N+1 or a missing field from the FE.
- New `web/lib/api/` module + types + MSW handlers mirroring the contract.

## Notes for the team
- **D-005 — CONFIRMED:** card fields = title + status badge + task count + description preview (first ~80 chars). The ~80-char truncation is an FE rendering concern; the backend should return the full `description`.
- Domain shape exists server-side: `UserStory{id,projectId,title,description,status,createdAt,updatedAt}`. `taskCount` is a derived/aggregate field the architect must add to the list contract; `description` (full) must also be on the list item so the card can render its preview.
- Follow existing FE conventions: all backend calls through `web/lib/api/`; types in `web/lib/api/types.ts` match the contract field-for-field; component tests use MSW with the architect's exact JSON. See `DocumentsTab.tsx`/`documents.ts` as the closest precedent (list-in-tab with loading/empty/error states).
- Keep selection state lifted so US005 can plug in the detail view without restructuring.

## Sign-off log
(po-ba appends here on each sign-off pass)

### Sign-off pass 1 — 2026-06-09 — verdict: approved
- **Spec review:** All six AC scenarios map to at least one UT-* / IT-* / FCT-* / E2E-* case.
  - Happy path (title/status/task count/desc preview): FCT-001, FCT-005, IT-001, E2E-US004-001.
  - Description truncation: FCT-001 (>80-char description) with an honest coverage exemption — CSS `line-clamp`/`text-overflow` is not computed in jsdom, so the test asserts the full text is in the DOM; visual truncation is verified at the e2e layer (E2E-US004-001 "truncated descriptions are visible"). Acceptable.
  - Empty state: FCT-002, UT-002, E2E-US004-002. Loading: FCT-003. Error: FCT-004 + IT-002/IT-003/IT-004 (404/500 paths).
  - Keyboard/click activation: FCT-005 (role=button, accessible name = title, click + Enter/Space → onSelect with story id).
  - E2E justification is honest — the 2 e2e cases (cross-stack task-count aggregation, empty state) genuinely require full-stack integration; everything else is correctly pushed down the pyramid.
- **Result review:** BE 4/4 PASS (IT-001..IT-004), FE all green (174 passed, 0 failed, 23 suites), E2E 2/2 PASS, full suite green × 3 consecutive runs. REQ Quality Gate approved. No tests skipped, no `t.Skip`, no `[Tags] skip`.
  - Note: the report renumbers FCT IDs (FCT-001..FCT-011) more granularly than the spec (FCT-001..FCT-006) — it is a superset (splits click/Enter and fetch/error into separate cases, adds a tab-panel integration test). Every spec case maps to a passing report case; no case was dropped.
- **Routed to:** none — story approved, Status set to done.
