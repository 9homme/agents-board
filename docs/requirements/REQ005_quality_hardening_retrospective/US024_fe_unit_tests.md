# US024 — Frontend component test specification

**For FE Dev:** these are the tests you write FIRST (TDD red). Implement in TypeScript using **Jest + React Testing Library**. All tests require `renderHook` (for hooks) or `render` (for components) from `@testing-library/react`. US024 depends on US020 being completed first (architecture §11.4) — do NOT start implementation until the US020 FE task is `completed`.

Architecture reference: §11 (US024 contract — all four subsections), §11.1 (MermaidDiagram ref-attach), §11.2 (useDocument reducer), §11.3 (DocumentsTab redirect contract), §11.5 (what US024 does NOT change).

**Important — this spec contains one existing test assertion that MUST be relaxed (architecture §2 US024 notes R8).** See FCT-US024-009 below.

## Coverage matrix

| AC / UI flow | Test ID | Component / hook under test | What it asserts |
|---|---|---|---|
| MermaidDiagram ref-attach: `<svg>` child rendered | FCT-US024-001 | `web/components/ProjectDetail/MermaidDiagram.tsx` | `container.querySelector('svg')` exists after success state |
| MermaidDiagram: no `dangerouslySetInnerHTML` in DOM | FCT-US024-002 | `web/components/ProjectDetail/MermaidDiagram.tsx` | no element has a `dangerouslySetInnerHTML` prop |
| MermaidDiagram: strict-mode double-mount yields exactly one `<svg>` child | FCT-US024-003 | `web/components/ProjectDetail/MermaidDiagram.tsx` | exactly one SVG child after React.StrictMode mount |
| MermaidDiagram: malformed SVG string does not throw | FCT-US024-004 | `web/components/ProjectDetail/MermaidDiagram.tsx` | component renders error fallback, not an uncaught exception |
| useDocument reducer: FETCH_STARTED clears state | FCT-US024-005 | `web/hooks/useDocument.ts` | dispatch FETCH_STARTED → data null, isLoading true, error null |
| useDocument reducer: FETCH_SUCCEEDED commits document | FCT-US024-006 | `web/hooks/useDocument.ts` | dispatch FETCH_SUCCEEDED → data populated, isLoading false, error null |
| useDocument reducer: FETCH_FAILED records error | FCT-US024-007 | `web/hooks/useDocument.ts` | dispatch FETCH_FAILED → data null, isLoading false, error set |
| useDocument reducer: ABORTED is a no-op on state | FCT-US024-008 | `web/hooks/useDocument.ts` | dispatch ABORTED after FETCH_STARTED → state unchanged from FETCH_STARTED |
| useDocument public return shape unchanged | FCT-US024-009 | `web/hooks/useDocument.ts` | `{ data, isLoading, error, refetch }` keys all present; types unchanged |
| DocumentsTab: first document selected at render time WITHOUT router.replace on mount | FCT-US024-010 | `web/components/ProjectDetail/DocumentsTab.tsx` | `selectedDocId` equals `documents[0].id`; `router.replace` NOT called on mount |
| DocumentsTab: user click writes URL via router.replace | FCT-US024-011 | `web/components/ProjectDetail/DocumentsTab.tsx` | clicking a sidebar item calls `router.replace` with `?doc=<id>` |
| DocumentsTab: bogus deep-link yields undefined selectedDocId | FCT-US024-012 | `web/components/ProjectDetail/DocumentsTab.tsx` | `isBogusDeepLink === true` → `selectedDocId` undefined; previewer not-found state |
| react-doctor `--diff` shows 7 rule IDs cleared | FCT-US024-013 | entire `web/` scan | `react-doctor --diff` output lists only removals; 7 named rule IDs gone |

## Component tests

### FCT-US024-001 — MermaidDiagram ref-attach: `<svg>` child present after success

- **Component / hook under test:** `web/components/ProjectDetail/MermaidDiagram.tsx`
- **Render with:**
  - The component needs to be driven into `renderState.status === 'success'`. Since the mermaid import is dynamic, either (a) mock the mermaid module to return a resolved SVG string, or (b) use the component's internal `setRenderState` via a test-only prop/callback if the dev exposes one. Architecture §11.1.2 documents that the ref-attach effect runs when `renderState.status === 'success'` and `renderState.svg` is a valid SVG string — the test must supply that state.
  - Recommended: mock the lazy `import('mermaid')` at the module level (`jest.mock('mermaid', () => ({ ... render: jest.fn().mockResolvedValue({ svg: '<svg><g id="a"/></svg>' }) }))`) and trigger the effect by waiting for the async mermaid render inside the existing lazy-load effect.
- **Expected:**
  - `container.querySelector('svg')` is not null after `await waitFor(...)`.
  - The `<svg>` element is a child of the wrapper `<div>` with `role="img"`.
- **Architecture cite:** architecture §11.1.2 — `<div ref={containerRef} role="img">` with appended `<svg>` node.

### FCT-US024-002 — MermaidDiagram: no element uses `dangerouslySetInnerHTML`

- **Component / hook under test:** `web/components/ProjectDetail/MermaidDiagram.tsx`
- **Render with:** same setup as FCT-US024-001 (drive to success state).
- **Expected:**
  - Inspect the rendered DOM: `container.querySelector('[dangerouslySetInnerHTML]')` is null.
  - Alternatively (since `dangerouslySetInnerHTML` is a React prop, not a DOM attribute): assert that the component's rendered output does NOT contain a `div` with `innerHTML` set to the SVG string. Use `container.innerHTML` and confirm it does NOT contain `<div dangerouslysetinnerhtml=...>` or any raw `innerHTML`-injected content pattern.
- **Notes:** In JSDOM, `dangerouslySetInnerHTML` becomes the `innerHTML` property of the element, not a DOM attribute. The cleanest assertion is: the `<svg>` node's parent is NOT a div whose `innerHTML` was set via React's `dangerouslySetInnerHTML` prop — i.e., `svgNode.parentElement.__reactFiber` has no `dangerouslySetInnerHTML` key. A simpler proxy: assert that no `div` in the container has its `innerHTML` equal to a raw SVG string (the ref-attach path appends a DOM node via `appendChild`, which leaves no innerHTML footprint).
- **Architecture cite:** architecture §11.1.1 — "`react-doctor/no-danger` matches the JSX prop name syntactically... the rule will fire because the prop is still present"; §11.1.2 — no `dangerouslySetInnerHTML` prop in the ref-attach JSX.

### FCT-US024-003 — MermaidDiagram: strict-mode double-mount yields exactly one `<svg>` child

- **Component / hook under test:** `web/components/ProjectDetail/MermaidDiagram.tsx`
- **Render with:**
  - Wrap in `<React.StrictMode>` for this test.
  - Same mermaid mock as FCT-US024-001.
- **Expected:**
  - After `await waitFor(...)`, `containerRef`'s host element has exactly one child element (`host.children.length === 1`).
  - The single child is an `<svg>` element.
- **Notes:** React 18 strict mode double-invokes effects in dev. The cleanup function in the new `useEffect` removes the appended node before the second run re-appends, so the final count must be exactly 1. Architecture R7 identifies this as a risk; this test is the mitigation.
- **Architecture cite:** architecture §11.1.2 — cleanup `return () => { if (host.firstChild === svgNode) host.removeChild(svgNode); }`; §14.1 R7.

### FCT-US024-004 — MermaidDiagram: malformed SVG string does not throw

- **Component / hook under test:** `web/components/ProjectDetail/MermaidDiagram.tsx`
- **Render with:**
  - Mock mermaid to resolve with a malformed SVG: `{ svg: 'not-an-svg' }` or `{ svg: '<div>oops</div>' }`.
  - Wrap in an `ErrorBoundary` to catch uncaught React errors.
- **Expected:**
  - Component does NOT throw (ErrorBoundary's fallback is NOT rendered).
  - The component either renders the error fallback (`<div role="alert">Could not render diagram</div>`) if the ref-attach logic wraps DOMParser failure in a try/catch, OR renders a blank wrapper `<div role="img"/>` with no child.
  - No uncaught exception propagates to the test.
- **Notes:** Architecture §11.1.2 — "Error handling unchanged... parsing failure inside the new effect would throw, but mermaid v11 emits well-formed SVG strings. Tester adds a defensive FCT-* test... and asserts the component does NOT throw (we wrap in try/catch if so)." The dev MUST wrap the DOMParser + appendChild block in try/catch; this test is the enforcement.
- **Architecture cite:** architecture §11.1.2 last paragraph.

### FCT-US024-005 — `useDocument` reducer: FETCH_STARTED clears prior state

- **Component / hook under test:** `web/hooks/useDocument.ts` — internal `reducer` function
- **Test approach:** the reducer is a pure function; test it directly if it is exported (even as an unexported module-level function, the test can access it via module internals if the dev exports it for testing — recommend a named export from a co-located `useDocument.reducer.ts` if the dev prefers clean separation, but that is the dev's call; the spec binds to the observable behaviour via `renderHook` if the reducer is not separately exported).
- **Render with (renderHook approach):**
  - Start hook with `documentId = undefined` (idle state — `data = null, isLoading = false, error = null`).
  - Provide MSW handler for `GET /api/v1/documents/doc-001` → delayed 300 ms.
  - Rerender with `documentId = "doc-001"`.
- **Observe mid-flight state:**
  - `await waitFor(() => expect(result.current.isLoading).toBe(true))` immediately after rerender.
- **Expect at FETCH_STARTED dispatch:**
  - `result.current.data` is null.
  - `result.current.isLoading` is true.
  - `result.current.error` is null.
- **Architecture cite:** architecture §11.2.3 — `case 'FETCH_STARTED': return { data: null, isLoading: true, error: null }`.

### FCT-US024-006 — `useDocument` reducer: FETCH_SUCCEEDED commits document

- **Component / hook under test:** `web/hooks/useDocument.ts`
- **MSW handlers:**
  - `GET /api/v1/documents/doc-001` → immediate 200 with:
    ```json
    {
      "id": "doc-001",
      "projectId": "proj-001",
      "name": "Architecture",
      "content": "# Arch",
      "createdAt": "2026-05-20T10:00:00Z",
      "updatedAt": "2026-05-20T10:00:00Z"
    }
    ```
- **Render with:** `renderHook(() => useDocument("doc-001"))`.
- **When:** `await waitFor(() => expect(result.current.isLoading).toBe(false))`
- **Expect (FETCH_SUCCEEDED state):**
  - `result.current.data.id` === `"doc-001"`.
  - `result.current.isLoading` === false.
  - `result.current.error` === null.
- **Architecture cite:** architecture §11.2.3 — `case 'FETCH_SUCCEEDED': return { data: action.document, isLoading: false, error: null }`.

### FCT-US024-007 — `useDocument` reducer: FETCH_FAILED records error

- **Component / hook under test:** `web/hooks/useDocument.ts`
- **MSW handlers:**
  - `GET /api/v1/documents/doc-bad` → 500 with `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch document" }`.
- **Render with:** `renderHook(() => useDocument("doc-bad"))`.
- **When:** `await waitFor(() => expect(result.current.error).not.toBeNull())`
- **Expect (FETCH_FAILED state):**
  - `result.current.data` === null.
  - `result.current.isLoading` === false.
  - `result.current.error` is non-null (an `ApiError` or `Error` instance).
- **Architecture cite:** architecture §11.2.3 — `case 'FETCH_FAILED': return { data: null, isLoading: false, error: action.error }`.

### FCT-US024-008 — `useDocument` reducer: ABORTED is a no-op on state

- **Component / hook under test:** `web/hooks/useDocument.ts`
- **MSW handlers:**
  - `GET /api/v1/documents/doc-001` → delayed 200 ms.
- **Render with:** `renderHook(() => useDocument("doc-001"))`.
- **Test steps:**
  1. Render hook — `isLoading` goes true (FETCH_STARTED dispatched).
  2. Immediately unmount hook (triggers `controller.abort()`, which will cause ABORTED to be dispatched when catch fires).
  3. Wait 300 ms.
- **Expect:**
  - The state at the time of unmount (FETCH_STARTED state: `{ data: null, isLoading: true, error: null }`) was NOT mutated by the ABORTED dispatch — i.e., no console error, no "state update on unmounted component" warning.
  - After unmount, no further state transitions occur (the hook is gone; this test is about confirming the ABORTED dispatch does not cause side effects).
- **Notes:** ABORTED is a no-op (`return state`) per architecture §11.2.3. The test verifies that the abort path completes silently, consistent with FCT-US020-005.
- **Architecture cite:** architecture §11.2.3 — `case 'ABORTED': return state;`; §11.2.2 — "ABORTED is dispatched but is a no-op on state."

### FCT-US024-009 — `useDocument` public return shape unchanged; existing tests pass

- **Component / hook under test:** `web/hooks/useDocument.ts`
- **Test approach:** structural assertion on the hook's return value.
- **Render with:**
  - MSW handler: `GET /api/v1/documents/doc-001` → 200 with valid document JSON.
  - `renderHook(() => useDocument("doc-001"))`.
- **When:** `await waitFor(() => expect(result.current.isLoading).toBe(false))`
- **Expect:**
  - `result.current` has exactly the keys: `data`, `isLoading`, `error`, `refetch`.
  - `typeof result.current.refetch` === `"function"`.
  - `result.current.data` type-checks as `Document | null` (TypeScript inference via `typeof result.current.data`).
- **Notes on relaxed assertion (architecture R8):** This is the test that MUST be updated from existing test files that may have asserted `router.replace` is called during `DocumentsTab` auto-selection on mount. Specifically: existing tests in `web/components/ProjectDetail/DocumentsTab.test.tsx` that assert `router.replace` is called when the documents list loads WITHOUT a `?doc=` parameter MUST be relaxed to `router.replace` is called ONLY when the user clicks a sidebar item. This is the one external-surface change US024 introduces. Document it in the test report.
- **Architecture cite:** architecture §11.2.4 — `UseDocumentResult` frozen shape `{ data, isLoading, error, refetch }`; §2 US024 notes — "The existing assertion that `router.replace` was called on auto-select must be relaxed."

### FCT-US024-010 — `DocumentsTab`: first document selected at render time; `router.replace` NOT called on initial mount

- **Component / hook under test:** `web/components/ProjectDetail/DocumentsTab.tsx`
- **Render with:**
  - Mock router: `query = { id: "proj-001" }` (no `doc` param).
  - Mock `router.replace` as a jest spy: `const replaceSpy = jest.fn()`.
  - Inject mock documents list: `[{ id: "doc-001", name: "Doc 1", ... }, { id: "doc-002", name: "Doc 2", ... }]`.
  - Mock `useDocument` to return `{ data: <doc-001 content>, isLoading: false, error: null, refetch: jest.fn() }` when called with `"doc-001"`.
- **When:** render `<DocumentsTab ... />` and wait for the initial render to settle.
- **Expect:**
  - The component passes `selectedId="doc-001"` to `DocumentSidebar` (first document is selected).
  - The component passes `document={<doc-001 content>}` to `DocumentPreviewer`.
  - `replaceSpy` was called ZERO times (no auto-select URL write on mount).
- **Notes:** This test is the authoritative assertion for architecture §11.3 — render-time fallback replaces the auto-select `useEffect`. The `selectedDocId = docParam ?? documents?.[0]?.id` computation happens at render time without any `router.replace` call.
- **Architecture cite:** architecture §11.3.2 — "selectedDocId = isBogusDeepLink ? undefined : (docParam ?? documents?.[0]?.id)"; §11.3.3 — "does NOT write to the URL"; R8.

### FCT-US024-011 — `DocumentsTab`: user click on sidebar item calls `router.replace` with `?doc=<id>`

- **Component / hook under test:** `web/components/ProjectDetail/DocumentsTab.tsx`
- **Render with:** same setup as FCT-US024-010.
- **User interactions (RTL):**
  1. `userEvent.click(screen.getByRole('button', { name: /Doc 2/i }))` (or whatever accessible role the sidebar item uses — adapt to the implementation).
- **Expect:**
  - `replaceSpy` was called exactly once.
  - The call included `{ doc: "doc-002" }` in the query (or `?doc=doc-002` in the URL shape — match what `handleSelectDoc` actually produces per existing implementation).
- **Architecture cite:** architecture §11.3.2 — "`handleSelectDoc` click handler... continues to own the URL write on user interaction."

### FCT-US024-012 — `DocumentsTab`: bogus deep-link yields `selectedDocId = undefined`

- **Component / hook under test:** `web/components/ProjectDetail/DocumentsTab.tsx`
- **Render with:**
  - `query = { id: "proj-001", doc: "doc-nonexistent" }` — a `?doc=` value not present in the documents list.
  - Documents list: `[{ id: "doc-001", ... }, { id: "doc-002", ... }]` (no `doc-nonexistent`).
  - `isBogusDeepLink` will be `true` inside the component.
- **Expect:**
  - `DocumentSidebar` receives `selectedId={undefined}` (no item highlighted).
  - `DocumentPreviewer` either renders a not-found state or is not mounted.
  - `router.replace` is NOT called (no auto-redirect).
- **Architecture cite:** architecture §11.3.2 — "`selectedDocId = isBogusDeepLink ? undefined : ...`".

### FCT-US024-013 — react-doctor `--diff` shows 7 named rule IDs cleared (meta-assertion)

- **Test type:** meta-assertion (not a Jest test; executed at code-review gate)
- **Notes:** This is a tech-lead-level assertion at review time. The dev runs `npx react-doctor scan web/` after implementing all four US024 changes and provides the output in the PR. Tech-lead verifies using `react-doctor --diff` against the recorded baseline that:
  1. `react-doctor/no-danger` — zero findings on `MermaidDiagram.tsx`.
  2. `react-doctor/no-cascading-set-state` — zero findings on `useDocument.ts`.
  3. `react-doctor/no-adjust-state-on-prop-change` — zero findings on `useDocument.ts`.
  4. `react-doctor/rendering-usetransition-loading` — zero findings on `useDocument.ts`.
  5. `react-doctor/no-event-handler` — zero findings on `DocumentsTab.tsx`.
  6. `react-doctor/nextjs-no-client-side-redirect` — zero findings on `DocumentsTab.tsx`.
  7. `react-doctor/exhaustive-deps` — zero findings on `DocumentsTab.tsx`.
  - Score reported is at least 96/100.
  - The `--diff` output shows ONLY removals (no new rule fires that was not in the original baseline).
- **Architecture cite:** US024 AC "Scenario: react-doctor score recovers"; §12 — "react-doctor skill is the primary review gate for US024"; §11.2.6 why rules clear.
