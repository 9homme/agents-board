# US043 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ007_user_stories_tab_and_e2e_gate/US043_user_story_detail_and_tasks.robot`.

## Why e2e
US043 implements the drill-down view (the right-side drawer). Proving that clicking a card triggers the parallel fetch of detail and tasks, displays them correctly without losing the list context, and manages selection/focus requires a real browser DOM and network lifecycle.

## Scenarios
### E2E-US043-001 — Clicking a card opens detail drawer with tasks
- **Tag:** US043, regression
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

### E2E-US043-002 — Switching stories and closing the drawer
- **Tag:** US043
- **Preconditions:** Seeded DB via MCP: a project with two user stories (Story 1 has 2 tasks; Story 2 has 0 tasks).
- **Steps:** 
  1. Click the first story card. Drawer opens showing Story 1.
  2. Click the second story card (without closing the drawer).
  3. Verify the drawer updates to show Story 2 details.
  3a. (Sub-assertion) Story 2 has 0 tasks — verify the empty-state text "No tasks for this story." is visible. This covers the E2E surface of the empty-tasks AC; the scenario does not require a separate test case because the seed data already provides the 0-task story in this context.
  4. Press the Escape key to close the drawer.
- **Expected:**
  - The drawer unmounts and is no longer visible.
- **Cleanup:** None.

**Note — close control:** The e2e scenario uses the Escape key rather than the X button. Both controls are required by the AC and proven by the implementation (FCT-003 covers the X-button click at FE component level; FCT-004 covers Escape). Escape is preferred at e2e because it is more robust in headless Playwright runs: no risk of the button being obscured or mislocated by layout. The X-button path is fully verified at the FE component layer.

**Note — empty-state coverage:** A standalone E2E-US043-003 scenario for the empty-state is not carried as a separate test case. The empty-state assertion ("No tasks for this story.") is exercised as step 3a of this scenario. Combined with FCT-002 (component-level empty state) and IT-004 (backend returns empty task list), the AC is fully covered without an additional e2e case that would duplicate the same seed data and drawer-open steps.

## Spec change log

### Revision 1 — 2026-06-09 — driver: po-ba sign-off pass 1

- changed E2E-US043-002 — updated close step from "Click the close (X) button" to "Press the Escape key"; added rationale note (Escape is more reliable in headless Playwright; X-button path proven at FCT-003). Updated precondition wording to make explicit that Story 2 has 0 tasks.
- changed E2E-US043-002 — added step 3a as an explicit sub-assertion for the empty-state ("No tasks for this story.") that was previously implicit in the robot file but undocumented in the spec.
- removed E2E-US043-003 — standalone empty-state scenario removed from the spec. The AC is fully covered by step 3a of E2E-US043-002 (e2e layer) + FCT-002 (FE component layer) + IT-004 (BE integration layer). A separate test case would duplicate seed data and drawer-open steps without adding coverage. Coverage exemption note added inline.
- No dev tasks affected — no test contract change to BE or FE specs; robot file already matched the revised spec intent (Escape key, 0-task story in E2E-002). Robot file documentation string updated to reflect the removal of the standalone E2E-003 case.
