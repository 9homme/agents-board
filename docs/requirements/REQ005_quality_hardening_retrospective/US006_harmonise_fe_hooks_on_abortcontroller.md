# US006 — Harmonise `useProject`, `useProjectDocuments`, `useDocument` on AbortController + signal-threaded `lib/api/`

**Requirement:** REQ005 — quality hardening retrospective
**Status:** draft

## Story
As a **future FE developer adding a "switch projects without unmounting the page" flow** (or any flow where a hook's input changes rapidly), I want **all three project/document fetch hooks to use the same race-safe AbortController pattern** that `useDocument` already exemplifies, so that a fast dep change cancels the in-flight network request and there is no risk of a stale earlier-started-but-later-resolved response clobbering newer state.

## Acceptance criteria

- **Scenario: `fetchProject` accepts `signal?: AbortSignal`**
  - Given `web/lib/api/projects.ts`
  - When `fetchProject(id)` is called today (no signal)
  - Then after this story it MUST accept an optional second parameter `signal?: AbortSignal`
  - And forward it to `fetchClient<Project>(..., { signal })` (same shape as `fetchDocument` and `fetchProjectDocuments` already do)
  - And the existing call sites that pass no signal continue to compile and work (backwards-compatible signature)

- **Scenario: all three `lib/api/` fetch functions accept `signal`**
  - Given `web/lib/api/projects.ts`, `web/lib/api/documents.ts`
  - When the story is complete
  - Then `fetchProject`, `fetchProjects`, `fetchProjectDocuments`, `fetchDocument` all accept `signal?: AbortSignal` as the last parameter
  - And `fetchProjects` (currently signal-less) also gains the signal parameter for consistency (callers like `useProjects` can ignore it for now, but the contract is uniform)

- **Scenario: `useProject` aborts on unmount and on id change**
  - Given the current `useProject(id)` uses the `mounted` flag pattern
  - When the hook is refactored
  - Then the hook creates a new `AbortController` per effect run
  - And passes `controller.signal` to `fetchProject(id, controller.signal)`
  - And the effect cleanup calls `controller.abort()`
  - And on rapid id change (sequence `id=A` → `id=B` before A resolves), the A request is aborted and only B's response commits to state
  - And a stale-id `latestIdRef` belt-and-braces guard prevents committing aborted responses to state (parity with `useDocument`)

- **Scenario: `useProjectDocuments` aborts on unmount and on projectId change**
  - Given the current `useProjectDocuments(projectId)` uses the `mounted` flag and `fetchCount` for `refetch()`
  - When the hook is refactored
  - Then the hook creates a new `AbortController` per effect run
  - And passes `controller.signal` to `fetchProjectDocuments(projectId, controller.signal)`
  - And the effect cleanup calls `controller.abort()`
  - And `refetch()` continues to work (incrementing `fetchCount` triggers a new effect run, which creates a new controller and aborts the previous one)
  - And rapid `projectId` change cancels the in-flight earlier request

- **Scenario: `useDocument` pattern unchanged**
  - Given `useDocument` is already the reference implementation
  - When the story is complete
  - Then `useDocument`'s implementation is not regressed
  - And its existing tests still pass
  - And any extraction of shared logic (NOT required by this story — see "Out of scope") preserves `useDocument`'s `latestIdRef` + abort semantics exactly

- **Scenario: aborted requests do NOT update state**
  - Given any of the three hooks where the input dep changes before the in-flight request resolves
  - When the hook is unit-tested with MSW + a deliberate delay
  - Then the test asserts that the resolved-but-aborted response does NOT cause `data` to update
  - And does NOT cause `error` to update
  - And does NOT cause `isLoading` to flicker

- **Scenario: aborted-request errors are silently ignored (not surfaced as `error`)**
  - Given `controller.abort()` is called while `fetch` is pending
  - When the resulting `AbortError` (or DOMException with name `'AbortError'`) is thrown
  - Then the hook's `.catch` handler discriminates it and does NOT set `error` state
  - And does NOT log to console (no noise)
  - And the abort is treated as expected control flow

- **Scenario: existing tests still pass**
  - Given `web/hooks/useProject.test.ts`, `useProjectDocuments.test.ts`, `useDocument.test.ts`
  - When the refactor is complete
  - Then all existing tests pass without modification beyond what the new abort semantics require (e.g. tests may need to assert NO state update after a deliberate abort — additions, not deletions, of assertions)
  - And `npm run typecheck && npm run lint -- --max-warnings=0 && npm test --watchAll=false` is clean

- **Scenario: no end-user-visible behaviour change in the happy path**
  - Given a user navigates to a project detail page and the hooks fetch as today
  - When everything resolves normally (no rapid nav, no unmount mid-flight)
  - Then the user sees identical loading / data / error states to before the refactor
  - And there is no visible regression in the project detail page, dashboard, or document tab

## UI / UX flow expectations

**No end-user-visible flow change.** The benefit is developer-observable (correctness when dep changes rapidly) and a small network-efficiency win (cancelled requests don't keep occupying connection pool slots). For completeness:

- **Entry points:** `pages/projects/[id].tsx` (uses `useProject`, `useProjectDocuments`, `useDocument` indirectly through the documents tab); dashboard's project list (uses `useProjects`).
- **Happy-path flow:** unchanged. User clicks a project card, page mounts, hooks fetch, data renders.
- **Edge case made safer:** if the user clicks a different project card before the first project's data resolves, the first project's request is now cancelled at the network layer and definitely cannot overwrite the second project's data on resolve. Today this is technically possible only via a contrived flow (the route change unmounts the page anyway), so the user-visible win is theoretical-but-real.
- **Loading / error states:** unchanged. Aborted requests do not emit error states (per AC).
- **Validation rules visible to the user:** none — no form changes.
- **Out of UI scope:** any visual refactor of project / document UI.

## Out of scope
- **Extracting a shared `useFetch<T>(key, fetcher)` hook.** That's a refactor with its own design surface — leave for a future story. The three hooks may end up looking similar after this story; that's acceptable.
- **Adding signal support to `useProjects`** beyond the `fetchProjects` signature change. The hook itself can stay `mounted`-flag-based if tech-lead prefers — dashboard list usage is not at risk of rapid dep change. AC says `fetchProjects` accepts signal (uniform API surface); hook refactor of `useProjects` is optional and not required by AC.
- **Removing `mounted` flag from anywhere else in the codebase** unless directly relevant to the three named hooks.
- **Server-side / SSR concerns.** Pages Router CSR-only is locked; this story does not touch SSR.

## Dependencies
- None. Independent of every other US in REQ005.

## Notes for the team

- **Reference implementation:** `web/hooks/useDocument.ts` (lines 50-94). The pattern is: `controllerRef.current?.abort()`, create new controller, `latestIdRef.current = id`, fetch with `controller.signal`, in `.then` check `latestIdRef.current === id` before committing, in `.catch` check `controller.signal.aborted` and bail.
- **Existing `fetchClient` already plumbs `signal`** (confirmed in `web/lib/api/client.ts` per `useDocument` working). No client-side changes expected; if the architect / tech-lead finds otherwise, surface it.
- **MSW test helpers:** `web/test/msw/` already has the project + document handlers. Add MSW handlers that delay (`await delay(100)`) for the abort tests so the dep change can land before the response.
- **Be careful with React 18 strict-mode double-invocation** in tests. The existing `useDocument` tests handle it; copy the pattern.
- **Lint note:** `useCallback` for `refetch` is fine and already in place. No new hooks-deps disable comments should be needed — the dep arrays should be `[id]`, `[projectId, fetchCount]`, `[documentId, fetchCount]` as today.

## Sign-off log
(po-ba appends here on each sign-off pass)
