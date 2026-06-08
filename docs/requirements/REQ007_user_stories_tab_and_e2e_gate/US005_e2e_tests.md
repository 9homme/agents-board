# US005 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ007_user_stories_tab_and_e2e_gate/US005_user_story_detail_and_tasks.robot`.

## Why e2e
US005 implements the drill-down view (the right-side drawer). Proving that clicking a card triggers the parallel fetch of detail and tasks, displays them correctly without losing the list context, and manages selection/focus requires a real browser DOM and network lifecycle.

## Scenarios
### E2E-US005-001 — Clicking a card opens detail drawer with tasks
- **Tag:** US005, regression
- **Preconditions:** Seeded DB via MCP: a project with a user story that has 2 tasks.
- **Steps:** 
  1. Navigate to the project detail page, User Stories tab.
  2. Click the user story card.
- **Expected:**
  - A drawer slides in (role=dialog).
  - The drawer displays the story's full description.
  - The drawer lists the 2 tasks with their titles and descriptions.
  - The card list is still visible behind the drawer.
- **Cleanup:** None.

### E2E-US005-002 — Switching stories and closing the drawer
- **Tag:** US005
- **Preconditions:** Seeded DB via MCP: a project with two user stories.
- **Steps:** 
  1. Click the first story card. Drawer opens showing Story 1.
  2. Click the second story card (without closing the drawer).
  3. Verify the drawer updates to show Story 2 details.
  4. Click the close ("X") button in the drawer.
- **Expected:**
  - The drawer unmounts and is no longer visible.
- **Cleanup:** None.

### E2E-US005-003 — Empty state for tasks in drawer
- **Tag:** US005
- **Preconditions:** Seeded DB via MCP: a project with a user story having 0 tasks.
- **Steps:** 
  1. Click the user story card.
- **Expected:**
  - Drawer shows "No tasks for this story."
- **Cleanup:** None.
