# US042 — Frontend component test specification

**For FE Dev:** these are the tests you write FIRST (TDD red). Implement in TypeScript using **Jest + React Testing Library**. Mock the backend at the API client layer using **MSW** with handlers that match the architecture's exact JSON request/response shapes.

## Coverage matrix
| AC / UI flow | Test ID | Component / hook under test | What it asserts |
|---|---|---|---|
| Renders user story cards with correct details | FCT-001 | `web/components/ProjectDetail/UserStoryCardList.tsx` | Cards display title, status badge, task count, and truncated description. |
| Empty state when no stories | FCT-002 | `web/components/ProjectDetail/UserStoryCardList.tsx` | Shows "No user stories yet for this project." |
| Loading state | FCT-003 | `web/components/ProjectDetail/UserStoryCardList.tsx` | Shows a loading indicator when fetching. |
| Error state | FCT-004 | `web/components/ProjectDetail/UserStoryCardList.tsx` | Shows "Couldn't load user stories." on API failure. |
| Card is clickable and accessible | FCT-005 | `web/components/ProjectDetail/UserStoryCard.tsx` | Keyboard/click triggers selection handler. Uses role=button. |
| Hook aborts pending fetch on remount/id change | FCT-006 | `web/hooks/useProjectUserStories.ts` | Dispatches AbortController signal correctly. |

## Component tests
### FCT-001 — Renders user story cards with correct details
- **Component under test:** `web/components/ProjectDetail/UserStoryCardList.tsx`
- **Render with:** `projectId` prop. MSW handlers intercepting the GET request.
- **MSW handlers:**
  - `GET /api/v1/projects/:id/user-stories` → 200 with an array containing one story with a long description (>80 chars) and `taskCount=3`, and one with a short description.
- **Expect:**
  - Renders 2 cards.
  - The first card displays "3 tasks", the status, title, and the full description (jsdom does not process CSS truncation).
  - The second card displays full description.
- **Architecture cite:** API contract row `GET /api/v1/projects/{id}/user-stories`.

### FCT-002 — Empty state when no stories
- **Component under test:** `web/components/ProjectDetail/UserStoryCardList.tsx`
- **Render with:** `projectId` prop.
- **MSW handlers:**
  - `GET /api/v1/projects/:id/user-stories` → 200 with `{"userStories":[]}`
- **Expect:**
  - `screen.findByText(/No user stories yet for this project/i)` is visible.
  - No cards are rendered.

### FCT-003 — Loading state
- **Component under test:** `web/components/ProjectDetail/UserStoryCardList.tsx`
- **MSW handlers:**
  - Delay the response for `GET /api/v1/projects/:id/user-stories`.
- **Expect:**
  - A loading spinner/indicator is visible initially.
  - No stale cards or error text are visible during loading.

### FCT-004 — Error state
- **Component under test:** `web/components/ProjectDetail/UserStoryCardList.tsx`
- **MSW handlers:**
  - `GET /api/v1/projects/:id/user-stories` → 500 Internal Server Error
- **Expect:**
  - `screen.findByText(/Couldn't load user stories/i)` is visible.
  - No cards are rendered.

### FCT-005 — Card is clickable and accessible
- **Component under test:** `web/components/ProjectDetail/UserStoryCard.tsx`
- **Render with:** `story` object prop, and a mocked `onSelect` function.
- **User interactions (RTL):**
  - Verify card element has `role="button"` (if a structural button is not used).
  - Tab focus to the card, hit Enter or Space.
  - Click the card.
- **Expect:**
  - `onSelect` is called with the story ID.
  - Accessible name (aria-label or inner text) includes the title.

### FCT-006 — Hook aborts pending fetch on remount/id change
- **Hook under test:** `web/hooks/useProjectUserStories.ts`
- **Expect:**
  - When the hook is unmounted or `projectId` changes mid-fetch, the `AbortController.abort()` is called and the fetch is cancelled, avoiding state updates on unmounted components. (Mirrors `useProjectDocuments`).

## Coverage exemption
- FCT-001: Skipped asserting visual text truncation. CSS-based truncation (`line-clamp` or `text-overflow: ellipsis`) is not computed in jsdom/RTL. Test asserts the full text is in the DOM.

## Spec change log
### Revision 1 — 2024-03-XX — driver: po-ba sign-off pass
- FCT-001: Modified to expect full text and added a coverage exemption for CSS visual truncation, as it is not testable in jsdom.
