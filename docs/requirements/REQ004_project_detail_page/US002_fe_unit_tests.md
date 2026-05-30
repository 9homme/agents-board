# US002 — Frontend component test specification

**For FE Dev:** these are the tests you write FIRST (TDD red). Implement in TypeScript using **Jest + React Testing Library**. Mock the backend at the API client layer (`web/lib/api/`) using **MSW** with handlers that match the architecture's exact JSON request/response shapes.

## Coverage matrix

| AC / UI flow | Test ID | Component / hook under test | What it asserts |
|---|---|---|---|
| Sidebar lists documents in `updatedAt DESC, id DESC` order | FCT-US002-001 | `web/components/ProjectDetail/DocumentSidebar.tsx` | fixture with a timestamp tie asserts correct DOM order |
| Empty state: no documents | FCT-US002-002 | `web/components/ProjectDetail/DocumentsTab.tsx` | "No documents yet" message; no document fetch attempted |
| Auto-select first doc when `?doc=` absent | FCT-US002-003 | `web/components/ProjectDetail/DocumentsTab.tsx` | first item highlighted; `router.replace` called with first doc id |
| Click sidebar item updates `?doc=` via shallow routing | FCT-US002-004 | `web/components/ProjectDetail/DocumentSidebar.tsx` | click item B → `router.replace` called with `doc=B`; item B marked active |
| Deep-link to valid `?doc=` pre-selects that document | FCT-US002-005 | `web/pages/projects/[id].tsx` | doc fixture shows selected; correct item `aria-selected="true"` |
| Deep-link to bogus `?doc=` shows "Document not found"; sidebar usable | FCT-US002-006 | `web/components/ProjectDetail/DocumentsTab.tsx` | "Document not found" in previewer; no item active; other items clickable |
| Rapid click cancels in-flight prior fetch (AbortController) | FCT-US002-007 | `web/hooks/useDocument.ts` | abort called on prior controller before new fetch; only latest doc shown |
| Previewer: loading indicator during content fetch | FCT-US002-008 | `web/components/ProjectDetail/DocumentPreviewer.tsx` | loading indicator present while MSW is delayed |
| Previewer: shows doc title + updatedAt + content when loaded | FCT-US002-009 | `web/components/ProjectDetail/DocumentPreviewer.tsx` | heading, timestamp, content text all visible |
| Previewer: in-pane error + Retry on 5xx content fetch; sidebar stays usable | FCT-US002-010 | `web/components/ProjectDetail/DocumentPreviewer.tsx` | error message + Retry button; sidebar item still clickable |
| Previewer: Retry button re-issues content fetch | FCT-US002-011 | `web/components/ProjectDetail/DocumentPreviewer.tsx` | clicking Retry triggers a second MSW request for the same doc |
| List loading state: skeleton in sidebar | FCT-US002-012 | `web/components/ProjectDetail/DocumentsTab.tsx` | loading indicator present while list MSW is pending |
| List error: sidebar error + Retry; previewer shows neutral state | FCT-US002-013 | `web/components/ProjectDetail/DocumentsTab.tsx` | error message; Retry present; previewer not crashed |
| List Retry re-issues the list fetch | FCT-US002-014 | `web/components/ProjectDetail/DocumentsTab.tsx` | clicking Retry triggers a second MSW list request |
| Project-not-found 404 from list endpoint surfaces page-level error state | FCT-US002-015 | `web/pages/projects/[id].tsx` | 404 from list triggers "Project not found" via `useProject` hook (coordinated with US001) |

## Component tests

### FCT-US002-001 — Sidebar: documents listed in `updatedAt DESC, id DESC` order
- **Component / hook under test:** `web/components/ProjectDetail/DocumentSidebar.tsx`
- **Render with:**
  - Props: `documents` array (pre-ordered by the server — sidebar renders in the order received); `selectedDocId=null`; `onSelect={jest.fn()}`.
  - Use a fixture that exercises the tiebreaker: three items where two share the same `updatedAt`:
    ```json
    [
      { "id": "cccc0003-0000-0000-0000-000000000003", "projectId": "p1", "title": "Doc C", "createdAt": "2026-05-15T08:00:00Z", "updatedAt": "2026-05-20T10:00:00Z" },
      { "id": "aaaa0001-0000-0000-0000-000000000001", "projectId": "p1", "title": "Doc A", "createdAt": "2026-05-14T08:00:00Z", "updatedAt": "2026-05-20T10:00:00Z" },
      { "id": "bbbb0002-0000-0000-0000-000000000002", "projectId": "p1", "title": "Doc B", "createdAt": "2026-05-13T08:00:00Z", "updatedAt": "2026-05-19T10:00:00Z" }
    ]
    ```
    (The fixture is already sorted as the server would return it; the component must render them in the order it receives them — no client-side re-sort.)
- **MSW handlers:** none (component receives pre-fetched data via props).
- **User interactions (RTL):** none.
- **Expect:**
  - `screen.getAllByRole('option')` (or equivalent list items) returns three elements.
  - First item text matches `/Doc C/`.
  - Second item text matches `/Doc A/`.
  - Third item text matches `/Doc B/`.
- **Notes:** The ordering is enforced by the BE (UT-US002-010 / IT-US002-003). This FE test confirms the component renders in received order without client-side re-sorting.
- **Architecture cite:** API contract §2 — "Order: updatedAt desc, then id desc".

### FCT-US002-002 — DocumentsTab: empty state when project has no documents
- **Component / hook under test:** `web/components/ProjectDetail/DocumentsTab.tsx`
- **Render with:**
  - `projectId = "p1"`, `docQueryParam = null`.
- **MSW handlers:**
  - `GET /api/v1/projects/p1/documents` → 200 `{ "documents": [] }`
  - (Assert that `GET /api/v1/documents/*` is NOT called.)
- **User interactions (RTL):** await list load.
- **Expect:**
  - `within(screen.getByTestId('documents-sidebar-area')).findByText(/No documents yet/i)` is visible. Use the scoped query to avoid a false-positive match against the previewer's longer string `/This project has no documents yet/i`, which is a superset substring match.
  - `screen.findByText(/This project has no documents yet/i)` is visible in the previewer area (full string is unique; no scoping needed).
  - MSW receives zero calls to `GET /api/v1/documents/*`.
- **Architecture cite:** US002 AC "Empty state — project has no documents"; §"Empty / loading / error states".

### FCT-US002-003 — DocumentsTab: auto-selects first document when `?doc=` absent
- **Component / hook under test:** `web/components/ProjectDetail/DocumentsTab.tsx`
- **Render with:**
  - `projectId = "p1"`, `docQueryParam = null` (absent).
  - Mock `router.replace`.
- **MSW handlers:**
  - `GET /api/v1/projects/p1/documents` → 200 with two-document fixture (first doc `id = "d111aaaa-..."`):
    ```json
    {
      "documents": [
        { "id": "d111aaaa-1111-1111-1111-111111111111", "projectId": "p1", "title": "Architecture overview", "createdAt": "2026-05-18T08:30:00Z", "updatedAt": "2026-05-20T09:45:00Z" },
        { "id": "d222bbbb-2222-2222-2222-222222222222", "projectId": "p1", "title": "Onboarding guide", "createdAt": "2026-05-15T11:00:00Z", "updatedAt": "2026-05-19T16:20:00Z" }
      ]
    }
    ```
  - `GET /api/v1/documents/d111aaaa-1111-1111-1111-111111111111` → 200 full document (see architecture §3).
- **User interactions (RTL):** await list load.
- **Expect:**
  - `router.replace` was called with `doc: "d111aaaa-1111-1111-1111-111111111111"` (and `shallow: true`).
  - The first sidebar item has `aria-selected="true"`.
  - The previewer shows the first document's title or content.
- **Architecture cite:** §"State strategy" — "auto-selects the first list item by issuing a shallow `router.replace` with `doc=<first.id>`"; US002 AC "Documents tab loads the list for the project".

### FCT-US002-004 — DocumentSidebar: clicking a different item updates `?doc=` via shallow routing
- **Component / hook under test:** `web/components/ProjectDetail/DocumentSidebar.tsx` (or `DocumentsTab.tsx` integrated)
- **Render with:**
  - Two-document fixture; `selectedDocId = "d111aaaa-..."` (first selected).
  - Mock `onSelect` callback (or `router.replace`).
- **MSW handlers:**
  - `GET /api/v1/documents/d222bbbb-2222-2222-2222-222222222222` → 200:
    ```json
    {
      "id": "d222bbbb-2222-2222-2222-222222222222",
      "projectId": "p1",
      "title": "Onboarding guide",
      "content": "# Onboarding\n\nWelcome.",
      "createdAt": "2026-05-15T11:00:00Z",
      "updatedAt": "2026-05-19T16:20:00Z"
    }
    ```
- **User interactions (RTL):**
  1. `userEvent.click(screen.getByRole('option', { name: /Onboarding guide/i }))` (or `getByText`)
- **Expect:**
  - `router.replace` (or `onSelect`) called with `doc: "d222bbbb-2222-2222-2222-222222222222"` and `shallow: true`.
  - Second sidebar item has `aria-selected="true"`.
  - First sidebar item has `aria-selected="false"`.
  - Previewer updates to show "Onboarding guide" content.
- **Architecture cite:** §"State strategy"; US002 AC "Selecting a document loads its content".

### FCT-US002-005 — Detail page: deep-link `?doc=` pre-selects and loads that document
- **Component / hook under test:** `web/pages/projects/[id].tsx`
- **Render with:**
  - `query = { id: "p1", tab: "documents", doc: "d222bbbb-2222-2222-2222-222222222222" }`
- **MSW handlers:**
  - `GET /api/v1/projects/p1` → 200 project fixture.
  - `GET /api/v1/projects/p1/documents` → 200 two-document fixture (same as FCT-US002-003).
  - `GET /api/v1/documents/d222bbbb-2222-2222-2222-222222222222` → 200 full document fixture.
- **User interactions (RTL):** await page load.
- **Expect:**
  - Sidebar item "Onboarding guide" has `aria-selected="true"`.
  - Sidebar item "Architecture overview" has `aria-selected="false"`.
  - Previewer renders "Onboarding guide" title.
  - `router.replace` is NOT called for auto-selection (the `?doc=` param is already set and valid).
- **Architecture cite:** US002 AC "Deep-link to a specific document"; §"State strategy".

### FCT-US002-006 — DocumentsTab: deep-link to bogus `?doc=` shows "Document not found"; sidebar remains usable
- **Component / hook under test:** `web/components/ProjectDetail/DocumentsTab.tsx`
- **Render with:**
  - `projectId = "p1"`, `docQueryParam = "bogus-id-not-in-list"`.
- **MSW handlers:**
  - `GET /api/v1/projects/p1/documents` → 200 two-document fixture.
  - (Assert `GET /api/v1/documents/bogus-id-not-in-list` is NOT called — we do not fetch a doc that's not in the list.)
- **User interactions (RTL):** await list load.
- **Expect:**
  - `screen.findByText(/Document not found/i)` is visible in the previewer.
  - No sidebar item has `aria-selected="true"`.
  - `router.replace` is NOT called for auto-selection (do not auto-select when the `?doc=` param is set but bogus).
  - Clicking "Architecture overview" in the sidebar still works (fires `onSelect` or triggers content fetch).
- **Architecture cite:** §"State strategy" — "detect that `doc` is set but is not in the list — do NOT auto-select; show 'Document not found'"; US002 AC "Deep-link to a document that doesn't exist".

### FCT-US002-007 — useDocument hook: rapid clicks cancel in-flight fetch (AbortController)
- **Component / hook under test:** `web/hooks/useDocument.ts`
- **Render with:** `renderHook(() => useDocument(documentId))` from `@testing-library/react-hooks` (or equivalent).
- **MSW handlers:**
  - `GET /api/v1/documents/doc-A` → delayed indefinitely (first request, will be aborted).
  - `GET /api/v1/documents/doc-B` → 200 immediately with:
    ```json
    {
      "id": "doc-B",
      "projectId": "p1",
      "title": "Doc B",
      "content": "Doc B content",
      "createdAt": "2026-05-20T10:00:00Z",
      "updatedAt": "2026-05-20T10:00:00Z"
    }
    ```
- **User interactions (renderHook):**
  1. Initial render with `documentId = "doc-A"` — first request goes in-flight.
  2. `rerender` (or `act`) with `documentId = "doc-B"` before doc-A resolves.
- **Expect:**
  - The MSW request for `doc-A` has `signal.aborted === true` when the second render occurs (verify via a spy on the `AbortController` the hook creates; or verify that MSW receives a request with a signal that becomes aborted).
  - After doc-B resolves, `result.current.data.id === "doc-B"` and `result.current.isLoading === false`.
  - `result.current.data` does NOT momentarily contain doc-A data (stale-id check prevents it).
- **Architecture cite:** §"State strategy" — "AbortController + stale-id check"; D-005 — "abort prior controller, create new one, store id in ref, issue fetch; on resolve only commit state if the resolved id matches `latestIdRef.current`".

### FCT-US002-008 — DocumentPreviewer: loading indicator during content fetch
- **Component / hook under test:** `web/components/ProjectDetail/DocumentPreviewer.tsx`
- **Render with:**
  - Props: `documentId = "doc-A"`, `isLoading = true` (or render with a delayed MSW response and let the hook feed in state).
- **MSW handlers:**
  - `GET /api/v1/documents/doc-A` → delayed indefinitely.
- **User interactions (RTL):** render and assert immediately.
- **Expect:**
  - A loading indicator (spinner or skeleton) is visible.
  - The sidebar is NOT rendered by this component (previewer is isolated — sidebar stays usable; assert this via component boundary).
- **Architecture cite:** US002 AC "Loading state — content is being fetched"; §"Empty / loading / error states".

### FCT-US002-009 — DocumentPreviewer: renders title, `updatedAt`, and content when loaded
- **Component / hook under test:** `web/components/ProjectDetail/DocumentPreviewer.tsx`
- **Render with:**
  - Props: `document = { id: "d111", projectId: "p1", title: "Architecture overview", content: "# Architecture\n\nThis project uses…", createdAt: "2026-05-18T08:30:00Z", updatedAt: "2026-05-20T09:45:00Z" }`, `isLoading = false`, `error = null`.
- **MSW handlers:** none (document fed via props).
- **User interactions (RTL):** none.
- **Expect:**
  - `screen.getByRole('heading', { name: /Architecture overview/i })` visible (title as heading in the previewer, e.g. `<h2>`).
  - `screen.getByText(/2026-05-20/i)` visible (muted updatedAt display — exact format per implementation, but must include the date).
  - `screen.getByText(/This project uses/i)` visible (content rendered — as plain text or `<pre>` acceptable in US002).
- **Architecture cite:** US002 §"Previewer specifics for US002"; API contract §3 response shape.

### FCT-US002-010 — DocumentPreviewer: in-pane error + Retry when content fetch fails; sidebar unaffected
- **Component / hook under test:** `web/components/ProjectDetail/DocumentPreviewer.tsx`
- **Render with:**
  - Props: `error = { code: "INTERNAL_ERROR", message: "Failed to fetch document" }`, `isLoading = false`, `document = null`, `onRetry = jest.fn()`.
- **MSW handlers:** none (error fed via props).
- **User interactions (RTL):** render.
- **Expect:**
  - `screen.getByText(/Failed to load document/i)` visible.
  - `screen.getByRole('button', { name: /Retry/i })` visible.
  - The previewer does NOT render sidebar DOM (previewer is isolated — sidebar is a sibling, not a child).
- **Architecture cite:** US002 AC "Error — content fetch fails"; §"Empty / loading / error states".

### FCT-US002-011 — DocumentPreviewer: Retry button re-issues content fetch
- **Component / hook under test:** `web/components/ProjectDetail/DocumentPreviewer.tsx`
- **Render with:** error state as in FCT-US002-010. `onRetry = jest.fn()`.
- **MSW handlers:** none.
- **User interactions (RTL):**
  1. `userEvent.click(screen.getByRole('button', { name: /Retry/i }))`
- **Expect:**
  - `onRetry` mock was called exactly once.
- **Notes:** The `onRetry` callback wires to `useDocument.refetch()` in the parent component. This test verifies the button calls the callback; the `refetch()` behavior is implicitly covered by FCT-US002-007 and the hook's own behavior.
- **Architecture cite:** US002 AC "clicking Retry re-issues the content fetch".

### FCT-US002-012 — DocumentsTab: list loading state shows skeleton in sidebar
- **Component / hook under test:** `web/components/ProjectDetail/DocumentsTab.tsx`
- **Render with:**
  - `projectId = "p1"`, `docQueryParam = null`.
- **MSW handlers:**
  - `GET /api/v1/projects/p1/documents` → delayed indefinitely.
- **User interactions (RTL):** render and assert before resolving.
- **Expect:**
  - A loading indicator (skeleton rows or spinner) is visible in the sidebar area.
  - No document titles are rendered.
- **Architecture cite:** US002 AC "Loading state — list is being fetched"; §"Empty / loading / error states".

### FCT-US002-013 — DocumentsTab: list fetch error shows sidebar error + Retry; previewer neutral
- **Component / hook under test:** `web/components/ProjectDetail/DocumentsTab.tsx`
- **Render with:**
  - `projectId = "p1"`, `docQueryParam = null`.
- **MSW handlers:**
  - `GET /api/v1/projects/p1/documents` → 500 `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch documents" }`.
- **User interactions (RTL):** await error.
- **Expect:**
  - `screen.findByText(/Couldn't load documents/i)` (or similar error message) in the sidebar.
  - `screen.findByRole('button', { name: /Retry/i })` in the sidebar.
  - Previewer shows a neutral state (e.g. "Document list unavailable") — not an unhandled crash.
- **Architecture cite:** US002 AC "Error — list fetch fails"; §"Empty / loading / error states"; API contract §2, 500 envelope.

### FCT-US002-014 — DocumentsTab: list Retry re-issues the list fetch
- **Component / hook under test:** `web/components/ProjectDetail/DocumentsTab.tsx`
- **Render with:** configure MSW to fail the first request then succeed on the second.
  - First `GET /api/v1/projects/p1/documents` → 500.
  - Second `GET /api/v1/projects/p1/documents` → 200 two-document fixture.
- **User interactions (RTL):**
  1. Await error state.
  2. `userEvent.click(screen.getByRole('button', { name: /Retry/i }))`.
  3. Await success state.
- **Expect:**
  - MSW received two requests to `GET /api/v1/projects/p1/documents`.
  - After Retry, document titles appear in the sidebar.
- **Architecture cite:** US002 AC "clicking Retry re-issues the list fetch".

### FCT-US002-015 — Detail page: 404 from list endpoint surfaces page-level missing state
- **Component / hook under test:** `web/pages/projects/[id].tsx`
- **Render with:** `query = { id: "ghost-project", tab: "documents" }`.
- **MSW handlers:**
  - `GET /api/v1/projects/ghost-project` → 404 `{ "code": "NOT_FOUND", "message": "Project not found" }`.
  - `GET /api/v1/projects/ghost-project/documents` → 404 `{ "code": "NOT_FOUND", "message": "Project not found" }`.
- **User interactions (RTL):** await page load.
- **Expect:**
  - `screen.findByText(/Project not found/i)` visible (the page-level error from `useProject` hook).
  - Tab switcher hidden (same behavior as FCT-US001-014).
- **Notes:** The primary guard for "project not found" is the `useProject` hook (US001). This test verifies the coordination: even if the documents list endpoint also returns 404, the page surfaces the project-not-found state (not a documents-specific error message). The `useProject` hook fires in parallel with the documents list fetch; its 404 takes precedence.
- **Architecture cite:** §"State strategy" — `useProject` discriminates `error.code === 'NOT_FOUND'`; D-006.

## Spec change log

### Revision 1 — 2026-05-30 — driver: po-ba sign-off pass 1
- changed FCT-US002-002 — scoped the sidebar "No documents yet" assertion to `within(screen.getByTestId('documents-sidebar-area'))` to prevent a false-positive substring match against the previewer's longer string "This project has no documents yet". The previewer assertion remains unscoped because its full string is unique. Test ID unchanged; AC mapping unchanged.
