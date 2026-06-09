# US005 — Frontend component test specification

**For FE Dev:** these are the tests you write FIRST (TDD red). Implement in TypeScript using **Jest + React Testing Library**. Mock the backend at the API client layer using **MSW** with handlers that match the architecture's exact JSON request/response shapes.

## Coverage matrix
| AC / UI flow | Test ID | Component / hook under test | What it asserts |
|---|---|---|---|
| Renders story detail and tasks in right-side drawer | FCT-001 | `web/components/ProjectDetail/UserStoriesTab.tsx` + Drawer | Selecting a card opens the drawer and fetches detail and tasks. |
| Drawer shows empty state when no tasks | FCT-002 | `web/components/ProjectDetail/UserStoryDrawer.tsx` | Shows "No tasks for this story." when tasks array is empty. |
| Close button closes drawer and returns focus | FCT-003 | `web/components/ProjectDetail/UserStoriesTab.tsx` | Clicking close unmounts drawer and clears selection. |
| Escape key closes drawer | FCT-004 | `web/components/ProjectDetail/UserStoryDrawer.tsx` | Pressing Escape closes drawer. |
| Switching stories updates drawer in place | FCT-005 | `web/components/ProjectDetail/UserStoriesTab.tsx` | Clicking a second card while drawer is open triggers new fetches. |
| Loading state in drawer | FCT-006 | `web/components/ProjectDetail/UserStoryDrawer.tsx` | Shows spinner while fetching. |
| Error state in drawer | FCT-007 | `web/components/ProjectDetail/UserStoryDrawer.tsx` | Shows "Couldn't load this user story." on API failure. |

## Component tests
### FCT-001 — Renders story detail and tasks in right-side drawer
- **Component under test:** `web/components/ProjectDetail/UserStoriesTab.tsx` (Integration)
- **Render with:** `projectId` prop.
- **MSW handlers:**
  - `GET /api/v1/projects/:id/user-stories` → returns 1 story.
  - `GET /api/v1/user-stories/:id` → returns story detail (no taskCount).
  - `GET /api/v1/user-stories/:id/tasks` → returns 2 tasks.
- **User interactions (RTL):**
  - Click the user story card.
- **Expect:**
  - Drawer (`role=dialog`) becomes visible.
  - Drawer displays full story description, status, and title.
  - Drawer displays the two tasks with their titles, statuses, and descriptions.

### FCT-002 — Drawer shows empty state when no tasks
- **Component under test:** `web/components/ProjectDetail/UserStoryDrawer.tsx`
- **Render with:** `storyId` prop.
- **MSW handlers:**
  - `GET /api/v1/user-stories/:id` → returns story detail.
  - `GET /api/v1/user-stories/:id/tasks` → returns `{"tasks":[]}`.
- **Expect:**
  - `screen.findByText(/No tasks for this story/i)` is visible.

### FCT-003 — Close button closes drawer and returns focus
- **Component under test:** `web/components/ProjectDetail/UserStoriesTab.tsx`
- **Render with:** `projectId` prop.
- **MSW handlers:**
  - `GET /api/v1/projects/:id/user-stories` → returns 1 story.
  - `GET /api/v1/user-stories/:id` → returns story detail.
  - `GET /api/v1/user-stories/:id/tasks` → returns task list.
- **User interactions (RTL):**
  - Click card to open drawer.
  - Wait for drawer to render (e.g., await the fetch to complete).
  - Click the close ("X") button in the drawer.
- **Expect:**
  - Drawer is no longer visible (unmounted).
  - Focus is returned to the card that was clicked (or selection is reset).

### FCT-004 — Escape key closes drawer
- **Component under test:** `web/components/ProjectDetail/UserStoryDrawer.tsx`
- **User interactions (RTL):**
  - Render drawer open.
  - Trigger Escape keydown event on the document/dialog.
- **Expect:**
  - `onClose` callback is triggered.

### FCT-005 — Switching stories updates drawer in place
- **Component under test:** `web/components/ProjectDetail/UserStoriesTab.tsx`
- **User interactions (RTL):**
  - Click Card 1. Drawer opens and loads Story 1 detail.
  - Click Card 2.
- **Expect:**
  - Drawer remains visible but initiates fetch for Story 2 and displays its details.

### FCT-006 — Loading state in drawer
- **Component under test:** `web/components/ProjectDetail/UserStoryDrawer.tsx`
- **MSW handlers:**
  - Delay the response for `GET /api/v1/user-stories/:id/tasks` and detail.
- **Expect:**
  - Loading spinner is visible in the drawer while requests are pending.

### FCT-007 — Error state in drawer
- **Component under test:** `web/components/ProjectDetail/UserStoryDrawer.tsx`
- **MSW handlers:**
  - `GET /api/v1/user-stories/:id` → 500 error.
- **Expect:**
  - `screen.findByText(/Couldn't load this user story/i)` is visible.
  - Close button is still available.

## Spec change log
### Revision 1 — 2024-03-XX — driver: po-ba sign-off pass
- FCT-003: Added explicit MSW handlers to satisfy preconditions for testing the close button interactions.
