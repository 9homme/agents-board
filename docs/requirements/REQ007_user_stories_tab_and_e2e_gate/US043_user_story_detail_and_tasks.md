# US043 — User story detail + tasks in a side drawer on card click

**Requirement:** REQ007 — User Stories Tab + E2E Quality Gate + Health-Check Fixes
**Status:** done

## Story
As a user, I want to click a user-story card and see the full story detail together with all of its tasks in a right-side drawer — with the card list still visible behind it — so that I can drill into a story and its breakdown without losing my place in the list.

## Acceptance criteria
- **Scenario: clicking a card opens a right-side detail drawer (happy path)**
  - Given the User Stories tab is showing the card list (US042)
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
- **Entry points:** clicking/activating a card in the User Stories tab list (US042).
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
- **US042** — the card list and selection mechanism this drawer is reached from.
- **Architecture (D-007):** REST endpoints for a single story's detail and its tasks. Proposed `GET /api/v1/user-stories/{id}` (bare `UserStory` object) and `GET /api/v1/user-stories/{id}/tasks` (`{"tasks":[Task,...]}`) — **System Architect finalises exact contract**, including whether tasks are embedded in the story-detail response or fetched separately. Single-resource responses follow the existing bare-object convention; list responses are wrapped.
- New `web/lib/api/` functions + types + MSW handlers mirroring the contract.

## Notes for the team
- **D-006 REVISED — CONFIRMED:** detail is a **right-side drawer/panel**, NOT an in-tab view that replaces the grid and NOT a dedicated route. The card list stays visible behind the panel; the panel has a **close button** (Escape also closes it).
- Selecting a different card while the drawer is open updates the drawer in place (selection state lifted from US042 drives which story the drawer shows).
- Domain shapes exist: `UserStory{id,projectId,title,description,status,createdAt,updatedAt}`, `Task{id,userStoryId,title,description,status,createdAt,updatedAt}`.
- If the architect chooses to embed tasks in the story-detail response, the FE makes one request; if separate, two — the tester's specs and MSW handlers must match the chosen contract exactly.
- Reuse the loading/empty/error patterns established in US042 and `DocumentsTab.tsx`. Consider an accessible drawer/dialog pattern (role, focus trap or sensible focus management, Escape-to-close).

## Sign-off log
(po-ba appends here on each sign-off pass)

### Sign-off pass 1 — 2026-06-09 — verdict: changes_requested
- **Spec review:** AC coverage is solid across layers. Happy path (FCT-001 / IT-001 / IT-003 / E2E-001), empty-state (FCT-002 / IT-004), close (FCT-003), switch (FCT-005 / E2E-002), loading (FCT-006), error (FCT-007 / IT-006 / IT-007 / UT-002 / UT-004), keyboard Escape (FCT-004). Two e2e spec-vs-implementation mismatches found:
  - **E2E-US043-003 (standalone empty-state scenario) is in `US043_e2e_tests.md` but absent from the robot suite** — the suite has only E2E-001 and E2E-002. The empty-state assertion ("No tasks for this story.") was folded into E2E-002 (Story 2 has 0 tasks), so the AC is still proven at e2e (plus FCT-002 + IT-004). But the spec claims 3 e2e cases and the implementation has 2 — the spec and the robot file must agree.
  - **E2E-US043-002 spec step 4 says "Click the close (X) button"; the robot file closes via Escape (line 83), not the X button.** The X-button close path is therefore only proven at FE unit level (FCT-003), never at e2e. Spec and implementation must agree on which control the e2e scenario exercises.
- **Result review:** All 11 BE (UT-001..004, IT-001..007) and all FE spec cases (FCT-001..007, verified present and passing in `UserStoryDrawer.test.tsx` / `UserStoriesTab.test.tsx`) pass; no skips. However the **test report's FCT ID table (US043_test_report.md lines 31-43) does not map to the FE spec's IDs** — it relabels cases (e.g. "FCT-002 = renders story title") and invents FCT-008..011, while showing no entry for the spec's E2E-US043-003. The underlying tests are correct and green; the report's ID mapping is inaccurate and must be regenerated against the spec's actual IDs before the next sign-off so the audit trail is trustworthy.
- **Routed to:**
  - **tester** — reconcile `US043_e2e_tests.md` with `US043_user_story_detail_and_tasks.robot`: (a) either restore E2E-US043-003 as its own case or update the spec to reflect the consolidation into E2E-002 and renumber; (b) align E2E-US043-002 close-control wording (X button) with the implementation (Escape) — or implement the X-button click in the robot case. No code change required; both AC outcomes already pass.
  - **orchestrator (report capture)** — regenerate `US043_test_report.md` FE/E2E tables mapped to the spec's real FCT-* / E2E-* IDs (and reflect E2E-003's status) on the next pass.
  - No dev/task routing — no failing behavior; all behaviors meet their ACs.

### Sign-off pass 2 — 2026-06-09 — verdict: approved
- **Spec review:** All three pass-1 findings resolved and now internally consistent across spec + robot file:
  - **E2E-US043-003 count mismatch — fixed.** `US043_e2e_tests.md` revision 1 documents the removal (lines 37, 45) with a coverage-exemption rationale; the robot suite has exactly two cases (E2E-001, E2E-002) and the suite Documentation string (lines 5-6) explains the consolidation. Spec count now matches implementation.
  - **E2E-US043-002 close-control mismatch — fixed.** Spec step 4 now reads "Press the Escape key" (line 30) with rationale (line 35); robot line 87 uses `Keyboard Key press Escape`. X-button close remains proven at FCT-003. Spec and implementation agree on the exercised control.
  - AC-to-test coverage re-verified, all layers, no skips/`t.Skip`/`[Tags] skip`: happy path (FCT-001 / IT-001 / IT-003 / E2E-001); empty-state (FCT-002 / IT-004 / E2E-002 step 3a); close (FCT-003 X-button + FCT-004 / E2E-002 Escape); switch (FCT-005 / E2E-002); loading (FCT-006); error (FCT-007 / IT-006 / IT-007 / UT-002 / UT-004); 404 paths (IT-002 / IT-005 / UT-001 / UT-003). Pyramid is honest — empty-state and error mapping live at the right layers; only the genuinely user-observable drawer lifecycle is at e2e.
- **Result review:** `US043_test_report.md` regenerated against real spec IDs — FCT-001..007 (lines 33-39), no invented FCT-008..011, E2E-US043-001/002 with explicit E2E-003-removal note (lines 50-52). All 11 BE (UT-001..004, IT-001..007), all 7 FE (FCT-001..007), and both E2E cases report PASS; 0 failed; e2e green ×3 consecutive runs. No skipped or dropped cases. Counts match the specs.
- **Routed to:** none — story approved, `Status: done`.
