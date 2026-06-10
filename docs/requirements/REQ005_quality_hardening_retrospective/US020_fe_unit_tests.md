# US020 — Frontend component test specification

**For FE Dev:** these are the tests you write FIRST (TDD red). Implement in TypeScript using **Jest + React Testing Library**. Mock the backend at the API client layer (`web/lib/api/`) using **MSW** with handlers that match the architecture's exact JSON request/response shapes. Hook tests use `renderHook` from `@testing-library/react`.

MSW handlers must use the exact API shapes from architecture §8 (N/A for new endpoints — the existing shapes from REQ001–REQ004 are unchanged). For delay-based abort tests, use `http.delay()` or `delay()` from MSW.

Architecture reference: §4 (AbortController hook contract), §4.2 (frozen `lib/api/` signatures), §4.3 (`useProject` contract), §4.4 (`useProjectDocuments` contract), §4.5 (AbortError handling rules).

## Coverage matrix

| AC / UI flow | Test ID | Component / hook under test | What it asserts |
|---|---|---|---|
| `fetchProject` signature accepts optional `signal` | FCT-US020-001 | `web/lib/api/projects.ts` | compiles with and without signal; signal forwarded to fetchClient |
| `fetchProjects` signature accepts optional `signal` | FCT-US020-002 | `web/lib/api/projects.ts` | compiles with and without signal |
| `useProject` aborts on unmount | FCT-US020-003 | `web/hooks/useProject.ts` | controller.abort called on cleanup; state not updated after unmount |
| `useProject` aborts on id change (rapid navigation) | FCT-US020-004 | `web/hooks/useProject.ts` | stale A response does not commit when B is pending |
| `useProject` AbortError silently swallowed | FCT-US020-005 | `web/hooks/useProject.ts` | error state not set; isLoading not left true on abort |
| `useProject` latestIdRef prevents stale commit | FCT-US020-006 | `web/hooks/useProject.ts` | only the latest id's response commits to state |
| `useProjectDocuments` aborts on unmount | FCT-US020-007 | `web/hooks/useProjectDocuments.ts` | cleanup aborts in-flight; no state update after unmount |
| `useProjectDocuments` aborts on projectId change | FCT-US020-008 | `web/hooks/useProjectDocuments.ts` | stale response does not overwrite newer |
| `useProjectDocuments` AbortError silently swallowed | FCT-US020-009 | `web/hooks/useProjectDocuments.ts` | no error state on abort |
| `useProjectDocuments` refetch() creates new controller | FCT-US020-010 | `web/hooks/useProjectDocuments.ts` | refetch() aborts previous in-flight; new request begins |
| `useDocument` pattern unchanged (regression guard) | FCT-US020-011 | `web/hooks/useDocument.ts` | existing tests still pass; no new failures |
| Happy path: data and loading state correct on success | FCT-US020-012 | `web/hooks/useProject.ts` | data populated; isLoading false; error null |

## Component tests

### FCT-US020-001 — `fetchProject` accepts optional `signal` and forwards it

- **Component / hook under test:** `web/lib/api/projects.ts` — `fetchProject` function
- **Render with:** no React component; invoke `fetchProject` directly in a test.
- **MSW handlers:**
  - `GET /api/v1/projects/:id` → 200 with:
    ```json
    {
      "id": "proj-001",
      "name": "Test Project",
      "description": "desc",
      "createdAt": "2026-05-20T10:00:00Z",
      "updatedAt": "2026-05-20T10:00:00Z"
    }
    ```
- **Test steps:**
  1. Create `const controller = new AbortController()`.
  2. Call `fetchProject("proj-001", controller.signal)` — must compile without TypeScript error (signature check).
  3. Call `fetchProject("proj-001")` — must also compile (optional parameter).
  4. Await the result.
- **Expect:**
  - TypeScript compiles cleanly for both call forms.
  - The resolved value has `id === "proj-001"`.
  - MSW receives exactly one request to `/api/v1/projects/proj-001`.
- **Architecture cite:** architecture §4.2 frozen signature `fetchProject(id: string, signal?: AbortSignal): Promise<Project>`.

### FCT-US020-002 — `fetchProjects` accepts optional `signal`

- **Component / hook under test:** `web/lib/api/projects.ts` — `fetchProjects` function
- **MSW handlers:**
  - `GET /api/v1/projects` → 200 with `{ "projects": [] }`
- **Test steps:**
  1. `fetchProjects()` — no signal, must compile.
  2. `fetchProjects(new AbortController().signal)` — with signal, must compile.
- **Expect:** TypeScript compiles for both forms; result resolves without error.
- **Architecture cite:** architecture §4.2 — `fetchProjects(signal?: AbortSignal): Promise<ProjectsResponse>`.

### FCT-US020-003 — `useProject` aborts on unmount; state not updated after unmount

- **Component / hook under test:** `web/hooks/useProject.ts`
- **Render with:** `renderHook(() => useProject("proj-001"))`
- **MSW handlers:**
  - `GET /api/v1/projects/proj-001` → delayed (`http.delay('infinite')`) — never resolves during the test.
- **Test steps:**
  1. `renderHook(() => useProject("proj-001"))`.
  2. Confirm `result.current.isLoading === true`.
  3. `unmount()` the hook.
  4. Wait one tick (`await new Promise(resolve => setTimeout(resolve, 50))`).
- **Expect:**
  - No React state-update warning ("Can't perform a React state update on an unmounted component") is logged to console.
  - `result.current.data` is still null (no state update after unmount).
  - `result.current.error` is still null.
- **Architecture cite:** architecture §4.3 — cleanup `return () => { controller.abort(); }`.

### FCT-US020-004 — `useProject` aborts stale request on rapid id change

- **Component / hook under test:** `web/hooks/useProject.ts`
- **Render with:** `renderHook(({ id }) => useProject(id), { initialProps: { id: "proj-A" } })`
- **MSW handlers:**
  - `GET /api/v1/projects/proj-A` → delayed 200 ms then resolves with `{ id: "proj-A", name: "Project A", ... }`.
  - `GET /api/v1/projects/proj-B` → immediate 200 with `{ id: "proj-B", name: "Project B", ... }`.
- **Test steps:**
  1. Initial render with `id = "proj-A"` — verify `isLoading === true`.
  2. Before proj-A resolves, `rerender({ id: "proj-B" })`.
  3. `await waitFor(() => expect(result.current.data?.id).toBe("proj-B"))`.
- **Expect:**
  - `result.current.data.id` is `"proj-B"` (B's response committed).
  - `result.current.data.name` is NOT `"Project A"` (A's stale response did not commit).
  - `result.current.isLoading` is false.
  - `result.current.error` is null.
- **Architecture cite:** architecture §4.3 — `latestIdRef` belt-and-braces guard; `controllerRef.current?.abort()` on dep change.

### FCT-US020-005 — `useProject` AbortError silently swallowed (no error state)

- **Component / hook under test:** `web/hooks/useProject.ts`
- **Render with:** `renderHook(() => useProject("proj-001"))`
- **MSW handlers:**
  - `GET /api/v1/projects/proj-001` → delayed 200 ms (so the fetch is in-flight when abort is called).
- **Test steps:**
  1. Render hook.
  2. Immediately `unmount()` (triggers `controller.abort()`).
  3. Wait 300 ms.
- **Expect:**
  - `result.current.error` is null — AbortError was NOT surfaced as error state.
  - `result.current.isLoading` is false (hook is unmounted, no state updates occurred).
  - Console has no error/warning related to the abort.
- **Architecture cite:** architecture §4.5 rule 1 — "If `controller.signal.aborted === true` when `.catch` fires, swallow the error."

### FCT-US020-006 — `useProject` `latestIdRef` prevents stale `.then` commit

- **Component / hook under test:** `web/hooks/useProject.ts`
- **Render with:** `renderHook(({ id }) => useProject(id), { initialProps: { id: "proj-A" } })`
- **MSW handlers:**
  - `GET /api/v1/projects/proj-A` → delayed 150 ms, then 200 with `{ id: "proj-A", name: "Project A", ... }`.
  - `GET /api/v1/projects/proj-B` → immediate 200 with `{ id: "proj-B", name: "Project B", ... }`.
- **Test steps:**
  1. Initial render `id = "proj-A"`.
  2. After 50 ms (proj-A still in-flight), `rerender({ id: "proj-B" })`.
  3. Wait 300 ms for both requests to settle.
- **Expect:**
  - `result.current.data?.name` is `"Project B"` — proj-A's late-resolving `.then` was blocked by the `latestIdRef.current !== id` guard.
- **Notes:** This case is subtly different from FCT-US020-004: here the abort may NOT fully cancel proj-A (the mock may resolve anyway after the delay), so the `latestIdRef` guard is the last line of defence.
- **Architecture cite:** architecture §4.3 — `.then` guard `if (latestIdRef.current === id)`.

### FCT-US020-007 — `useProjectDocuments` aborts on unmount

- **Component / hook under test:** `web/hooks/useProjectDocuments.ts`
- **MSW handlers:**
  - `GET /api/v1/projects/proj-001/documents` → delayed (`http.delay('infinite')`).
- **Test steps:** same shape as FCT-US020-003 but with `useProjectDocuments("proj-001")`.
- **Expect:** no state-update-on-unmounted-component warning; `data` and `error` remain null.
- **Architecture cite:** architecture §4.4 — cleanup `return () => { controller.abort(); }`.

### FCT-US020-008 — `useProjectDocuments` aborts stale request on projectId change

- **Component / hook under test:** `web/hooks/useProjectDocuments.ts`
- **MSW handlers:**
  - `GET /api/v1/projects/proj-A/documents` → delayed 200 ms, then 200 with `{ documents: [{ id: "doc-A", ... }], total: 1 }`.
  - `GET /api/v1/projects/proj-B/documents` → immediate 200 with `{ documents: [{ id: "doc-B", ... }], total: 1 }`.
- **Test steps:**
  1. Render `useProjectDocuments("proj-A")`.
  2. Before proj-A resolves, rerender with `"proj-B"`.
  3. `await waitFor(() => expect(result.current.data?.documents[0]?.id).toBe("doc-B"))`.
- **Expect:** `data.documents[0].id` is `"doc-B"`; proj-A's stale response did not commit.
- **Architecture cite:** architecture §4.4 — `latestKeyRef` guard; `controllerRef.current?.abort()` on dep change.

### FCT-US020-009 — `useProjectDocuments` AbortError silently swallowed

- **Component / hook under test:** `web/hooks/useProjectDocuments.ts`
- **MSW handlers:**
  - `GET /api/v1/projects/proj-001/documents` → delayed 200 ms.
- **Test steps:** render; immediately unmount; wait 300 ms.
- **Expect:** `result.current.error` null; no console error/warning about abort.
- **Architecture cite:** architecture §4.5 rule 1.

### FCT-US020-010 — `useProjectDocuments` `refetch()` aborts previous in-flight and creates new request

- **Component / hook under test:** `web/hooks/useProjectDocuments.ts`
- **MSW handlers:**
  - First request: `GET /api/v1/projects/proj-001/documents` → delayed 300 ms.
  - Subsequent (same path, MSW will serve again): immediate 200 with `{ documents: [{ id: "doc-1", ... }], total: 1 }`.
- **Test steps:**
  1. Render `useProjectDocuments("proj-001")`.
  2. Before first request resolves (at ~100 ms), call `result.current.refetch()`.
  3. `await waitFor(() => expect(result.current.data?.documents).toBeDefined())`.
- **Expect:**
  - `result.current.data` is populated from the refetch response.
  - `result.current.isLoading` is false.
  - `result.current.error` is null.
  - The first in-flight request was aborted (MSW received two requests; the first request was aborted mid-flight — verifiable via MSW handler counting or by asserting the final state matches the second response).
- **Architecture cite:** architecture §4.4 — `refetch()` increments `fetchCount`; new effect run aborts previous controller.

### FCT-US020-011 — `useDocument` pattern unchanged (regression guard)

- **Component / hook under test:** `web/hooks/useDocument.ts`
- **MSW handlers:**
  - `GET /api/v1/documents/doc-001` → 200 with `{ id: "doc-001", projectId: "proj-001", name: "Doc 1", content: "# Hello", createdAt: "2026-05-20T10:00:00Z", updatedAt: "2026-05-20T10:00:00Z" }`.
- **Test steps:**
  1. Run all existing `useDocument` tests (this test case confirms ZERO existing tests were broken).
  2. Additionally, render `useProjectDocuments` and confirm existing return shape `{ data, isLoading, error, refetch }` is intact.
- **Expect:** all pre-existing `useDocument.test.ts` assertions pass without modification.
- **Notes:** This test case is a meta-assertion — it documents the non-regression contract. The actual assertion is that `npm test` exits 0 with no failures in `useDocument.test.ts`.
- **Architecture cite:** US020 AC "Scenario: `useDocument` pattern unchanged"; architecture §4.6.

### FCT-US020-012 — `useProject` happy path: data populated, isLoading false, error null

- **Component / hook under test:** `web/hooks/useProject.ts`
- **MSW handlers:**
  - `GET /api/v1/projects/proj-001` → immediate 200 with:
    ```json
    {
      "id": "proj-001",
      "name": "Happy Project",
      "description": "All good",
      "createdAt": "2026-05-20T10:00:00Z",
      "updatedAt": "2026-05-20T10:00:00Z"
    }
    ```
- **Test steps:**
  1. `renderHook(() => useProject("proj-001"))`.
  2. `await waitFor(() => expect(result.current.isLoading).toBe(false))`.
- **Expect:**
  - `result.current.data.id` === `"proj-001"`.
  - `result.current.data.name` === `"Happy Project"`.
  - `result.current.isLoading` === false.
  - `result.current.error` === null.
  - `result.current.isNotFound` === false.
- **Architecture cite:** architecture §4.3 — `UseProjectResult` public shape `{ data, isLoading, error, isNotFound }`.
