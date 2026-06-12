# US042 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ007_user_stories_tab_and_e2e_gate/US042_user_story_cards_list.robot`.

## Why e2e
US042 implements the User Stories tab listing. Proving that it displays correct task count aggregates and handles empty states requires full-stack integration spanning UI rendering, frontend API client, backend query grouping, and database seeded state.

## Scenarios
### E2E-US042-001 — User stories render with accurate details
- **Tag:** US042, regression
- **Preconditions:** Seeded DB via MCP: a project with one user story having 2 tasks, and one having 0 tasks.
- **Steps:** 
  1. Navigate to the project detail page.
  2. Click the "User Stories" tab.
  3. Wait for the cards to load.
- **Expected:**
  - Two cards are displayed.
  - The card for the first story shows "2 tasks".
  - The card for the second story shows "0 tasks".
  - Status badges and truncated descriptions are visible.
- **Cleanup:** None (test DB isolation).

### E2E-US042-002 — Empty state when no stories
- **Tag:** US042
- **Preconditions:** Seeded DB via MCP: a project with no user stories.
- **Steps:** 
  1. Navigate to the project detail page.
  2. Click the "User Stories" tab.
- **Expected:**
  - Displays "No user stories yet for this project."
  - No cards are rendered.
- **Cleanup:** None.
