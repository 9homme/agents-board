# US047 — Frontend component test specification

**For FE Dev:** these are the tests you write FIRST (TDD red). Implement in TypeScript using **Jest + React Testing Library**. Mock the backend at the API client layer (`web/lib/api/`) using **MSW** with handlers that match the architecture's exact JSON request/response shapes.

## Coverage matrix

| AC / UI flow | Test ID | Component / hook under test | What it asserts |
|---|---|---|---|
| Project header shows linked path (read-only) | FCT-047-001 | `web/components/ProjectDetail/ProjectHeader.tsx` | path text visible |
| Project header shows path from project object | FCT-047-002 | `ProjectHeader` | correct path string rendered |
| Requirements area: loading state | FCT-047-003 | `web/components/ProjectDetail/RequirementSelector.tsx` | loading indicator visible |
| Requirements area: populated list | FCT-047-004 | `RequirementSelector` | requirement names rendered |
| Requirements area: empty state | FCT-047-005 | `RequirementSelector` | "No requirements yet" text |
| Requirements area: error state | FCT-047-006 | `RequirementSelector` | inline error visible |
| Selecting a requirement updates URL query param | FCT-047-007 | `web/pages/projects/[id].tsx` | router.push/replace with ?requirement= |
| Deep-link: ?requirement= loads correct requirement | FCT-047-008 | `web/pages/projects/[id].tsx` | requirement pre-selected from URL |
| UserStoriesTab fetches by requirementId | FCT-047-009 | `web/components/ProjectDetail/UserStoriesTab.tsx` | API call uses /requirements/:rid/user-stories |
| DocumentsTab fetches by requirementId | FCT-047-010 | `web/components/ProjectDetail/DocumentsTab.tsx` | API call uses /requirements/:rid/documents |
| useProjectRequirements — idle/loading | FCT-047-011 | `web/hooks/useProjectRequirements.ts` | loading=true initially |
| useProjectRequirements — success | FCT-047-012 | `useProjectRequirements` | requirements array populated |
| useProjectRequirements — empty | FCT-047-013 | `useProjectRequirements` | empty array, no error |
| useProjectRequirements — 404 project | FCT-047-014 | `useProjectRequirements` | error state from 404 |
| useProjectRequirements — 500 error | FCT-047-015 | `useProjectRequirements` | error state from 500 |
| useProjectRequirements — abort on unmount | FCT-047-016 | `useProjectRequirements` | AbortController cleanup |
| useRequirementUserStories — fetches by rid | FCT-047-017 | `web/hooks/useRequirementUserStories.ts` (or re-parameterised hook) | URL uses /requirements/:rid/user-stories |
| useRequirementDocuments — fetches by rid | FCT-047-018 | `web/hooks/useRequirementDocuments.ts` | URL uses /requirements/:rid/documents |
| Requirement item shows name and status | FCT-047-019 | `RequirementSelector` | name + status rendered per item |
| Default requirement renders for migrated project | FCT-047-020 | `RequirementSelector` | "Default" requirement item visible |
| No requirement selected: tabs show empty/placeholder | FCT-047-021 | `web/pages/projects/[id].tsx` | tab bodies prompt to select a requirement |
| fetchProjectRequirements sends correct URL | FCT-047-022 | `web/lib/api/requirements.ts` | GET /api/v1/projects/:pid/requirements |
| fetchRequirementUserStories sends correct URL | FCT-047-023 | `web/lib/api/userStories.ts` | GET /api/v1/projects/:pid/requirements/:rid/user-stories |
| fetchRequirementDocuments sends correct URL | FCT-047-024 | `web/lib/api/documents.ts` | GET /api/v1/projects/:pid/requirements/:rid/documents |
| userStory list item includes requirementId field | FCT-047-025 | `UserStoriesTab` | requirementId on each rendered story |
| document list item includes requirementId field | FCT-047-026 | `DocumentsTab` | requirementId on each rendered document |
| Accessibility: requirements list is navigable by keyboard | FCT-047-027 | `RequirementSelector` | focus/keyboard selection |
| Accessibility: error state announced | FCT-047-028 | `RequirementSelector` | role=alert on error |

---

## MSW handler definitions

Add to `web/test/msw/handlers.ts`:

```typescript
const PROJECT_ID = '11111111-1111-1111-1111-111111111111'
const REQ_ID     = 'b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f'

// GET /api/v1/projects/:pid — with path field
http.get('/api/v1/projects/:pid', ({ params }) => {
  return HttpResponse.json({
    id: params.pid,
    name: 'agents-board',
    description: '',
    path: '/Users/me/workspace/agents-board',
    createdAt: '2026-06-01T09:00:00Z',
    updatedAt: '2026-06-01T09:00:00Z',
  })
})

// GET /api/v1/projects/:pid/requirements — populated
http.get('/api/v1/projects/:pid/requirements', ({ params }) => {
  return HttpResponse.json({
    requirements: [
      {
        id: REQ_ID,
        projectId: params.pid,
        name: 'Default',
        description: '',
        status: 'draft',
        createdAt: '2026-06-09T10:00:00Z',
        updatedAt: '2026-06-09T10:00:00Z',
      }
    ]
  })
})

// GET /api/v1/projects/:pid/requirements — empty
// (override per-test via server.use)

// GET /api/v1/projects/:pid/requirements/:rid/user-stories
http.get('/api/v1/projects/:pid/requirements/:rid/user-stories', () => {
  return HttpResponse.json({
    userStories: [
      {
        id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        projectId: PROJECT_ID,
        requirementId: REQ_ID,
        title: 'Add item to basket',
        description: '',
        status: 'in_progress',
        taskCount: 3,
        createdAt: '2026-06-02T09:00:00Z',
        updatedAt: '2026-06-02T09:00:00Z',
      }
    ]
  })
})

// GET /api/v1/projects/:pid/requirements/:rid/documents
http.get('/api/v1/projects/:pid/requirements/:rid/documents', () => {
  return HttpResponse.json({
    documents: [
      {
        id: 'cccccccc-cccc-cccc-cccc-cccccccccccc',
        projectId: PROJECT_ID,
        requirementId: REQ_ID,
        title: 'README',
        createdAt: '2026-06-02T09:00:00Z',
        updatedAt: '2026-06-02T09:00:00Z',
      }
    ]
  })
})
```

---

## Component tests

### FCT-047-001 — ProjectHeader renders linked path as read-only text
- **Component / hook under test:** `web/components/ProjectDetail/ProjectHeader.tsx`
- **Render with:** `<ProjectHeader project={{ id: '...', name: 'agents-board', path: '/Users/me/workspace/agents-board', ... }} />`
- **Expect:**
  - `screen.getByText('/Users/me/workspace/agents-board')` is in the document.
  - No input field contains the path (it is read-only text, not editable).
- **Architecture cite:** US047 AC "Linked path visible"; FE surface `ProjectHeader`

### FCT-047-002 — ProjectHeader renders the exact path string from the project object
- **Component / hook under test:** `web/components/ProjectDetail/ProjectHeader.tsx`
- **Render with:** Various path values including paths with spaces, unicode characters, and a long path (>80 chars).
- **Expect:**
  - Rendered text matches the supplied `path` string exactly (no truncation in the DOM that would break the assertion, though CSS truncation is allowed).
- **Architecture cite:** §2 `path` string — "always a non-empty directory path; key is always present"

### FCT-047-003 — RequirementSelector shows loading state while fetching
- **Component / hook under test:** `web/components/ProjectDetail/RequirementSelector.tsx`
- **Render with:** MSW handler delayed (as in FCT-046-010 pattern).
- **Expect (before response resolves):**
  - A loading indicator (spinner, skeleton, `aria-busy="true"`, or text like "Loading...") is visible.
  - No requirement items are rendered yet.
- **Architecture cite:** US047 AC "Loading state"

### FCT-047-004 — RequirementSelector renders requirement list on success
- **Component / hook under test:** `web/components/ProjectDetail/RequirementSelector.tsx`
- **MSW handler:** `GET /api/v1/projects/:pid/requirements` → 200 with one requirement item.
- **Render with:** `projectId = PROJECT_ID`.
- **Expect (after response):**
  - `screen.getByText('Default')` visible.
  - Requirement item shows status `"draft"` (or a badge/text representing it).
- **Architecture cite:** US047 AC "Project shows its requirements"

### FCT-047-005 — RequirementSelector renders empty state for project with no requirements
- **Component / hook under test:** `web/components/ProjectDetail/RequirementSelector.tsx`
- **MSW handler:** Override to return `{"requirements": []}`.
- **Expect:**
  - `screen.getByText(/no requirements yet/i)` or equivalent empty-state text is visible.
  - No requirement items rendered.
- **Architecture cite:** US047 AC "Empty requirements"

### FCT-047-006 — RequirementSelector renders error state on fetch failure
- **Component / hook under test:** `web/components/ProjectDetail/RequirementSelector.tsx`
- **MSW handler:** Override `GET /api/v1/projects/:pid/requirements` → 500 `{"code":"INTERNAL_ERROR","message":"Failed to fetch requirements"}`.
- **Expect:**
  - An inline error message is visible in the requirements area.
  - The rest of the page (project header) still renders (error is scoped to the requirements area).
- **Architecture cite:** US047 AC "Error state"; §4 500 response

### FCT-047-007 — Selecting a requirement updates the URL query param
- **Component / hook under test:** `web/pages/projects/[id].tsx`
- **MSW handlers:** Project + requirements list populated.
- **Setup:** Mock `useRouter` from `next/router`; capture `router.push` or `router.replace` calls.
- **User interactions:**
  1. Render the page; wait for requirements list.
  2. Click on the "Default" requirement item.
- **Expect:**
  - `router.push` (or `router.replace` with `{ shallow: true }`) is called with a URL containing `?requirement=b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f` (the requirement ID).
- **Architecture cite:** US047 AC "Drill into a requirement"; US047 notes — "shallow routing" mirroring existing `tab` pattern

### FCT-047-008 — Deep-link: page pre-selects requirement from ?requirement= query param
- **Component / hook under test:** `web/pages/projects/[id].tsx`
- **MSW handlers:** Project, requirements list (one item), user-stories for that requirement.
- **Setup:** Mock `useRouter` with `query = { id: PROJECT_ID, requirement: REQ_ID }`.
- **Expect:**
  - The "Default" requirement is visually selected (has active/selected styling or aria-selected).
  - The user stories tab fetches from `/api/v1/projects/${PROJECT_ID}/requirements/${REQ_ID}/user-stories` (verify via MSW request capture).
- **Architecture cite:** US047 AC "Deep-link to a requirement"

### FCT-047-009 — UserStoriesTab fetches user stories via /requirements/:rid/user-stories
- **Component / hook under test:** `web/components/ProjectDetail/UserStoriesTab.tsx`
- **MSW handlers:** Capture requests to the new canonical path.
- **Render with:** `projectId = PROJECT_ID`, `requirementId = REQ_ID`.
- **Expect:**
  - MSW receives a request to `GET /api/v1/projects/${PROJECT_ID}/requirements/${REQ_ID}/user-stories`.
  - The OLD flat route `GET /api/v1/projects/:id/user-stories` is NOT called.
  - User story items are rendered with `requirementId` present in their data.
- **Architecture cite:** §6 new canonical path; Breaking changes — flat route removed

### FCT-047-010 — DocumentsTab fetches documents via /requirements/:rid/documents
- **Component / hook under test:** `web/components/ProjectDetail/DocumentsTab.tsx`
- **MSW handlers:** Capture requests to the canonical path.
- **Render with:** `projectId = PROJECT_ID`, `requirementId = REQ_ID`.
- **Expect:**
  - MSW receives `GET /api/v1/projects/${PROJECT_ID}/requirements/${REQ_ID}/documents`.
  - Old flat route `GET /api/v1/projects/:id/documents` NOT called.
  - Document items rendered with `requirementId` in their data.
- **Architecture cite:** §10 new canonical path; Breaking changes

### FCT-047-011 — useProjectRequirements: loading=true initially
- **Component / hook under test:** `web/hooks/useProjectRequirements.ts`
- **Render with:** `renderHook(() => useProjectRequirements(PROJECT_ID))`. MSW handler delayed.
- **Expect (before response):**
  - `result.current.loading === true`
  - `result.current.requirements` is empty / undefined.
  - `result.current.error` is null.

### FCT-047-012 — useProjectRequirements: success — requirements array populated
- **Component / hook under test:** `web/hooks/useProjectRequirements.ts`
- **MSW handler:** 200 with one requirement.
- **Expect (after response):**
  - `result.current.loading === false`
  - `result.current.requirements` has length 1.
  - `result.current.requirements[0]` has `id`, `projectId`, `name`, `description`, `status`, `createdAt`, `updatedAt`.
- **Architecture cite:** §4 200 response shape

### FCT-047-013 — useProjectRequirements: empty — empty array, no error
- **Component / hook under test:** `web/hooks/useProjectRequirements.ts`
- **MSW handler:** 200 `{"requirements": []}`.
- **Expect:**
  - `result.current.requirements` is an empty array (`[]`).
  - `result.current.error` is null.
  - `result.current.loading === false`.

### FCT-047-014 — useProjectRequirements: 404 project not found → error state
- **Component / hook under test:** `web/hooks/useProjectRequirements.ts`
- **MSW handler:** Override to 404 `{"code":"NOT_FOUND","message":"Project not found"}`.
- **Expect:**
  - `result.current.error` is non-null (contains the error code or message).
  - `result.current.loading === false`.
  - `result.current.requirements` is empty.

### FCT-047-015 — useProjectRequirements: 500 → error state
- **Component / hook under test:** `web/hooks/useProjectRequirements.ts`
- **MSW handler:** 500.
- **Expect:** Same as FCT-047-014 (error state populated).

### FCT-047-016 — useProjectRequirements: AbortController cleanup on unmount
- **Component / hook under test:** `web/hooks/useProjectRequirements.ts`
- **MSW handler:** Delayed response.
- **When:** Call `unmount()` before response resolves.
- **Expect:**
  - No state-update-after-unmount warning.
  - MSW receives the abort signal (request is cancelled).
- **Architecture cite:** Tester policy — "one FCT-* per `useEffect` cleanup / abort path"

### FCT-047-017 — useRequirementUserStories fetches from /projects/:pid/requirements/:rid/user-stories
- **Component / hook under test:** `web/hooks/useRequirementUserStories.ts` (or `useProjectUserStories` re-parameterised)
- **MSW handler:** Capture the request URL.
- **Render with:** `renderHook(() => useRequirementUserStories(PROJECT_ID, REQ_ID))`.
- **Expect:**
  - Fetches `GET /api/v1/projects/${PROJECT_ID}/requirements/${REQ_ID}/user-stories`.
  - Does NOT fetch the old flat URL `/api/v1/projects/:id/user-stories`.
- **Architecture cite:** §6

### FCT-047-018 — useRequirementDocuments fetches from /projects/:pid/requirements/:rid/documents
- **Component / hook under test:** `web/hooks/useRequirementDocuments.ts`
- **MSW handler:** Capture the request URL.
- **Render with:** `renderHook(() => useRequirementDocuments(PROJECT_ID, REQ_ID))`.
- **Expect:**
  - Fetches `GET /api/v1/projects/${PROJECT_ID}/requirements/${REQ_ID}/documents`.
  - Does NOT fetch the old flat URL.
- **Architecture cite:** §10

### FCT-047-019 — RequirementSelector renders name and status for each requirement
- **Component / hook under test:** `web/components/ProjectDetail/RequirementSelector.tsx`
- **MSW handler:** Two requirements, one with status `"draft"`, one with status `"in_progress"`.
- **Expect:**
  - Both requirement names are visible.
  - Both status labels (e.g. `"draft"`, `"in_progress"`) are visible or have accessible text.
- **Architecture cite:** US047 AC "I see a list of the project's requirements (name, plus its status)"

### FCT-047-020 — RequirementSelector renders "Default" requirement for migrated projects
- **Component / hook under test:** `web/components/ProjectDetail/RequirementSelector.tsx`
- **MSW handler:** One requirement named `"Default"` with status `"draft"`.
- **Expect:**
  - `screen.getByText('Default')` visible.
  - The selector is usable with only one item (no crash, no infinite loop).
- **Architecture cite:** US047 notes — "Migrated projects each have one Default requirement"

### FCT-047-021 — No requirement selected: tab bodies show placeholder to select
- **Component / hook under test:** `web/pages/projects/[id].tsx`
- **MSW handlers:** Project loaded, requirements loaded, no `requirement` query param set.
- **Expect:**
  - User Stories and Documents tab bodies show a message prompting the user to select a requirement (e.g. "Select a requirement to view user stories").
  - No fetch to `/api/v1/projects/:pid/requirements/:rid/user-stories` is made.
- **Architecture cite:** US047 UX — requirement selection drives the tab content

### FCT-047-022 — fetchProjectRequirements API client sends correct URL
- **Component / hook under test:** `web/lib/api/requirements.ts` → `fetchProjectRequirements`
- **MSW handler:** Capture and inspect the request URL.
- **When:** `fetchProjectRequirements(PROJECT_ID)` called.
- **Expect:**
  - URL is `GET /api/v1/projects/${PROJECT_ID}/requirements`.
  - No auth headers.
- **Architecture cite:** §4

### FCT-047-023 — fetchRequirementUserStories API client sends correct URL
- **Component / hook under test:** `web/lib/api/userStories.ts` → `fetchRequirementUserStories`
- **When:** `fetchRequirementUserStories(PROJECT_ID, REQ_ID)` called.
- **Expect:**
  - URL is `GET /api/v1/projects/${PROJECT_ID}/requirements/${REQ_ID}/user-stories`.
- **Architecture cite:** §6

### FCT-047-024 — fetchRequirementDocuments API client sends correct URL
- **Component / hook under test:** `web/lib/api/documents.ts` → `fetchRequirementDocuments`
- **When:** `fetchRequirementDocuments(PROJECT_ID, REQ_ID)` called.
- **Expect:**
  - URL is `GET /api/v1/projects/${PROJECT_ID}/requirements/${REQ_ID}/documents`.
- **Architecture cite:** §10

### FCT-047-025 — User story list items include requirementId field
- **Component / hook under test:** `web/components/ProjectDetail/UserStoriesTab.tsx`
- **MSW handler:** User stories response includes `requirementId` on each item (as per §6).
- **Expect:**
  - The rendered story cards/items expose `requirementId` accessible to the component (e.g. `data-requirement-id` attribute or visible text). At minimum, the data is passed correctly without being dropped.
  - TypeScript type `UserStory` has `requirementId: string`.
- **Architecture cite:** §6 user-story list item shape — `requirementId` always present

### FCT-047-026 — Document list items include requirementId field
- **Component / hook under test:** `web/components/ProjectDetail/DocumentsTab.tsx`
- **MSW handler:** Documents response includes `requirementId` (as per §10).
- **Expect:**
  - `requirementId` is present on the document data type and passed to/from the component.
  - TypeScript type `Document` has `requirementId: string`.
- **Architecture cite:** §10 document list item shape

### FCT-047-027 — RequirementSelector is navigable by keyboard
- **Component / hook under test:** `web/components/ProjectDetail/RequirementSelector.tsx`
- **MSW handler:** Two requirements.
- **Render with:** Selector rendered and focused.
- **User interactions:**
  1. `userEvent.tab()` to move focus to the first requirement item.
  2. `userEvent.keyboard('{Enter}')` or `{Space}` to select it.
- **Expect:**
  - `router.push` (or equivalent) is called with the first requirement's id in the query param.
  - The selected item has `aria-selected="true"` or equivalent.
- **Architecture cite:** Tester policy — keyboard nav

### FCT-047-028 — RequirementSelector error state has accessible announcement
- **Component / hook under test:** `web/components/ProjectDetail/RequirementSelector.tsx`
- **MSW handler:** 500.
- **Expect:**
  - The error container has `role="alert"` or a live region.
- **Architecture cite:** Tester policy — `aria-*` on error
