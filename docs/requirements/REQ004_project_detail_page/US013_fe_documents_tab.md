---
US: US013
Title: Documents tab — sidebar + previewer (plain rendering) + hooks + API client + types + MSW + signal pass-through
Status: completed
Track: FE
Implements: US013 AC "Documents tab loads the list for the project", "Selecting a document loads its content into the previewer", "Deep-link to a specific document", "Deep-link to a document that doesn't exist for this project", "Empty state — project has no documents", "Loading state — list is being fetched", "Loading state — content is being fetched" (including race-cancellation), "Error — list fetch fails", "Error — content fetch fails"
Blocked by: US012_fe_detail_page_with_tabs.md
Worked-by: fe-dev-2026-05-28T00-00-00Z-adbf
---

## Goal
Make the Documents tab functional end-to-end (plain rendering — markdown polish is US014): sidebar lists the project's documents from `GET /api/v1/projects/{id}/documents`, clicking selects a document and writes `?doc=` to the URL (shallow), the previewer renders the selected document's plain content from `GET /api/v1/documents/{id}` with race-safe cancellation via `AbortController`, auto-selection of the first document when `?doc=` is absent, in-pane error isolation (sidebar stays usable when content fetch fails), all loading / empty / error / 404 / deep-link-to-bogus-doc states wired. Also extends the shared FE infrastructure (`types.ts`, `client.ts` signal pass-through, MSW handlers for both new endpoints) that this tab depends on.

## Architecture references
- `architecture.md` §"Components → Frontend" → rows `DocumentsTab.tsx`, `DocumentSidebar.tsx`, `DocumentPreviewer.tsx` (the **US013 plain** variant — markdown upgrade is US014).
- `architecture.md` §"Hooks" → rows `useProjectDocuments.ts`, `useDocument.ts` (race-safe via `AbortController` + `latestIdRef`; `refetch()` for the Retry button on both).
- `architecture.md` §"API client" → rows `web/lib/api/documents.ts` (new — `fetchProjectDocuments(projectId, signal?)`, `fetchDocument(documentId, signal?)`) and `web/lib/api/client.ts` (modified — accept and forward optional `signal: AbortSignal`).
- `architecture.md` §"Types" → row `web/lib/api/types.ts` (add `Document`, `DocumentListItem`, `DocumentsListResponse` exactly as in §"Frontend TypeScript interface mapping").
- `architecture.md` §"MSW" → row `web/test/msw/handlers.ts` (add handlers + 404 variants for both new endpoints).
- `architecture.md` §"State strategy" — `?doc=` shallow `router.replace`; auto-select when `tab === 'documents'` + list non-empty + `doc` absent; "deep-link to bogus doc" detection (do NOT auto-select, render "Document not found" in the previewer).
- `architecture.md` §"Data flow" — sequence diagram is the contract for the race-safe path.
- `architecture.md` §"Key decisions" → D-005 (AbortController + stale-id ref).
- `architecture.md` §"API contracts" → endpoints #2 and #3 (MSW fixtures must mirror these exactly).

## Scope
- **In:**
  - Extend `web/lib/api/types.ts` with `DocumentListItem`, `DocumentsListResponse`, `Document` interfaces — copy field-for-field from architecture §"Frontend TypeScript interface mapping". Do NOT touch the existing `Project` / `ProjectsResponse` / `ErrorResponse` exports.
  - Extend `web/lib/api/client.ts`: accept an optional `signal: AbortSignal` and forward it to `fetch` (`fetch(url, { ...options, signal: options.signal, headers: {...} })`). Existing callers (no `signal`) must keep working unchanged.
  - Create `web/lib/api/documents.ts` exporting `fetchProjectDocuments(projectId: string, signal?: AbortSignal): Promise<DocumentsListResponse>` and `fetchDocument(documentId: string, signal?: AbortSignal): Promise<Document>`. Both call `fetchClient` with `encodeURIComponent(...)` on the id and pass through the `signal`.
  - Create `web/hooks/useProjectDocuments.ts`: `useProjectDocuments(projectId: string | undefined)` returning `{ data: DocumentsListResponse | null, isLoading, error, refetch }`. Skip fetch when `projectId` is undefined. Expose `refetch()` for the Retry button.
  - Create `web/hooks/useDocument.ts`: `useDocument(documentId: string | undefined)` returning `{ data: Document | null, isLoading, error, refetch }`. Implementation: `controllerRef = useRef<AbortController | null>(null)`; `latestIdRef = useRef<string | undefined>(undefined)`. On each change of `documentId`: abort prior controller (if any), create a new `AbortController`, store id in `latestIdRef`, issue `fetchDocument(documentId, controller.signal)`; on resolve, commit state **only if** `documentId === latestIdRef.current` (belt-and-braces — abort signals the network and the stale-id check ignores any late resolution). `refetch()` re-runs the most recent fetch. Skip when `documentId` is undefined.
  - Create `web/components/ProjectDetail/DocumentSidebar.tsx`: renders the list. Props: `documents: DocumentListItem[]`, `selectedId: string | undefined`, `onSelect(id)`. Renders a header `Documents (N)` and a `<ul role="listbox" aria-label="Documents">` of `<li><button role="option" aria-selected={...} title={fullTitle}>{title}</button></li>` items. Truncates long titles with CSS `text-overflow: ellipsis` (single line). Active item visually distinct. Keyboard: items are native `<button>` so Tab/Enter/Space already work; arrow-key list navigation is OPTIONAL for this story (not in AC) — dev can add it but it's not required.
  - Create `web/components/ProjectDetail/DocumentPreviewer.tsx` (**US013 plain variant** — US014 swaps the body): props `{ document: Document | null, isLoading: boolean, error: Error | null, isNotFound: boolean, onRetry: () => void }`. Renders:
    - Loading → spinner / skeleton.
    - Error (not-404) → "Failed to load document" + Retry button wired to `onRetry`.
    - `isNotFound` → "Document not found" friendly message (no Retry — see notes).
    - Happy → `<h2>{document.title}</h2>` + muted `<p>Updated {document.updatedAt}</p>` + the raw `content` in a `<pre>` or `<div>` (plain rendering acceptable — US014 will swap this body for `<MarkdownRenderer source={document.content} />` without changing the component's prop surface or the surrounding wiring). Component is rendered with `key={document.id}` by `DocumentsTab` so subsequent mermaid SVG state (US014) cleans up correctly.
  - Create `web/components/ProjectDetail/DocumentsTab.tsx`: orchestrates `DocumentSidebar` + `DocumentPreviewer`. Props: `projectId: string`. Reads `?doc=` from `useRouter().query.doc`. Calls `useProjectDocuments(projectId)` for the list; calls `useDocument(selectedId)` for the content. Implements:
    - **List loading** → sidebar shows skeleton; previewer shows neutral "Loading documents…" / blank state.
    - **List error** → sidebar shows "Couldn't load documents" + Retry (wired to `useProjectDocuments.refetch`); previewer shows "Document list unavailable" neutral message.
    - **List empty** → sidebar shows "No documents yet"; previewer shows "This project has no documents yet"; no content fetch issued.
    - **List non-empty + `doc` absent** → auto-select first item via `router.replace({ pathname, query: { ...query, doc: list[0].id } }, undefined, { shallow: true })`. The effect must be idempotent (guard on `doc` being absent or not in list).
    - **List non-empty + `doc` present and in list** → mark active in sidebar; `useDocument(doc)` fetches; render result via `DocumentPreviewer`.
    - **List non-empty + `doc` present but NOT in list** (deep-link-to-bogus-doc) → DO NOT auto-select; render `DocumentPreviewer` with `isNotFound=true` and no selected sidebar item.
    - **Document content error (non-404)** → `DocumentPreviewer` shows "Failed to load document" + Retry (wired to `useDocument.refetch`); sidebar remains fully interactive (`useProjectDocuments.data` unaffected).
    - On sidebar item click → `router.replace({ pathname, query: { ...query, doc: clickedId } }, undefined, { shallow: true })`.
    - The race AC is satisfied by `useDocument`'s `AbortController` + stale-id ref — clicking another item while a fetch is in flight aborts the prior and the previewer ends up showing the most recently selected doc, not whichever resolved last.
  - Wire `DocumentsTab` into `web/pages/projects/[id].tsx` by replacing the `<div data-testid="documents-tab-placeholder"/>` slot (introduced by `us001_fe_detail_page_with_tabs`) with `<DocumentsTab projectId={id} />`. The page must NOT pass content state to the tab — the tab owns its data fetching.
  - Extend `web/test/msw/handlers.ts`: add handlers for `*/api/v1/projects/:id/documents` (200 happy with multiple docs in `updatedAt DESC` order; 200 empty `{ documents: [] }`; 404 for unknown project; 500) AND `*/api/v1/documents/:id` (200 happy with full content + ISO-8601 timestamps; 404; 500). Fixtures match the architecture's API contract field-for-field. The tester's MSW spec dictates how variants are toggled per test (e.g. by id pattern or by `server.use(...)` override) — match it.
  - Jest tests: component tests for `DocumentSidebar`, `DocumentPreviewer`, `DocumentsTab`; hook tests for `useProjectDocuments` and `useDocument` (the latter must explicitly assert race-cancellation behavior — clicking B while A is in flight ends with B in state, not A); API-client tests for `documents.ts` and the new `signal` pass-through in `client.ts`.
- **Out:**
  - Markdown / mermaid rendering inside the previewer body (US014 — markdown / mermaid tasks).
  - Any change to `web/components/ProjectDetail/ProjectHeader.tsx` / `TabSwitcher.tsx` / `UserStoriesTab.tsx` / `web/hooks/useProject.ts` (owned by US012).
  - Any change to `ProjectCard.tsx` (US012 FE card link task).
  - Backend changes — fetched via the two BE endpoints already wired by the US013 BE tasks.

## Files touched (estimated, exclusive)
- `web/lib/api/types.ts` (modified — add `Document`, `DocumentListItem`, `DocumentsListResponse`)
- `web/lib/api/client.ts` (modified — `signal: AbortSignal` pass-through)
- `web/lib/api/client.test.ts` (new — tests for `signal` pass-through, if no existing test file; otherwise extend)
- `web/lib/api/documents.ts` (new)
- `web/lib/api/documents.test.ts` (new)
- `web/hooks/useProjectDocuments.ts` (new)
- `web/hooks/useProjectDocuments.test.ts` (new)
- `web/hooks/useDocument.ts` (new)
- `web/hooks/useDocument.test.ts` (new)
- `web/components/ProjectDetail/DocumentsTab.tsx` (new)
- `web/components/ProjectDetail/DocumentsTab.test.tsx` (new)
- `web/components/ProjectDetail/DocumentSidebar.tsx` (new)
- `web/components/ProjectDetail/DocumentSidebar.test.tsx` (new)
- `web/components/ProjectDetail/DocumentPreviewer.tsx` (new)
- `web/components/ProjectDetail/DocumentPreviewer.test.tsx` (new)
- `web/pages/projects/[id].tsx` (modified — replace the Documents tab placeholder slot with `<DocumentsTab />`)
- `web/test/msw/handlers.ts` (modified — add handlers for `*/api/v1/projects/:id/documents` and `*/api/v1/documents/:id`)

> **Scaffold posture for shared FE files:** this task is sequenced **after** `US012_fe_detail_page_with_tabs.md` to avoid parallel-write collisions on `web/test/msw/handlers.ts` (and the slot in `web/pages/projects/[id].tsx`). `web/lib/api/types.ts` and `web/lib/api/client.ts` are also high-collision but no other REQ004 task touches them, so this task can own those edits exclusively.

## Test contract
The dev must make the matching cases in `US013_fe_unit_tests.md` pass — covering: list happy path with ordering rendered top-to-bottom matching API order, list empty state, list error + Retry, auto-select first item (URL update via shallow replace), deep-link to existing doc (selection + content render), deep-link to bogus doc (previewer "Document not found", no sidebar selection, sidebar still interactive), content loading state (sidebar usable), content error + Retry, sidebar selection writes `?doc=` shallowly, race-cancellation (clicking B while A is in flight → previewer ends with B), and the `signal` pass-through in `client.ts`. (FCT-* IDs assigned by tester.) If the tester has not yet authored the relevant IDs at the time the dev picks this up, the dev flags it back to tester rather than skipping coverage.

## Implementation notes
- **CSR-only.** No `getServerSideProps` / `getStaticProps` / `getInitialProps` / `web/pages/api/*`.
- All backend calls go through `web/lib/api/documents.ts` — no raw `fetch` in components or hooks.
- The `useDocument` race-cancellation pattern is the load-bearing piece. Suggested skeleton:
  ```ts
  const controllerRef = useRef<AbortController | null>(null);
  const latestIdRef = useRef<string | undefined>(undefined);
  // inside the effect on documentId change:
  controllerRef.current?.abort();
  const controller = new AbortController();
  controllerRef.current = controller;
  latestIdRef.current = documentId;
  fetchDocument(documentId, controller.signal)
    .then((doc) => { if (latestIdRef.current === documentId) setData(doc); })
    .catch((err) => {
      if (controller.signal.aborted) return;
      if (latestIdRef.current === documentId) setError(err);
    });
  return () => controller.abort();
  ```
  Tester's spec will assert this with a deferred MSW handler (resolve A only after B has been issued — final state must reflect B).
- Auto-select effect must guard against infinite loops: only fire when `tab === 'documents'` (or absent → treated as documents) AND `doc` is absent or refers to an id not in `data.documents`. Use the loaded list's first item id; do not auto-select when the list is empty.
- "Deep-link to bogus doc" detection: `doc` is set AND `data.documents` is non-empty AND no item matches `doc` → previewer renders `isNotFound=true`. Do NOT issue `useDocument(doc)` for a bogus id (avoids a 404 round-trip when we already know it's bogus from the list).
  - Edge case: if the document exists per the list but the detail endpoint returns 404 (race / deleted), `useDocument` surfaces an `ApiError` with `code === 'NOT_FOUND'`; the previewer should treat that the same as `isNotFound` (no Retry, friendly copy).
- MSW handlers' fixtures MUST match the architecture's API contract field-for-field. `documents` is always an array (never `null`). Timestamps are ISO-8601 UTC strings like `2026-05-20T09:45:00Z`. The list endpoint MUST NOT include `content`.
- `web/lib/api/client.ts` `signal` pass-through: keep backward compatibility. The `fetchClient` signature stays `fetchClient<T>(endpoint, options?)`; existing callers continue to omit `signal`. Internal `fetch(url, { ...options, headers: {...} })` already spreads options, so `signal` lands naturally — verify with a test that calls `fetchClient` with `signal` and observes `fetch.mock.calls[0][1].signal === signal`.
- The previewer's "Document not found" friendly copy lives in the previewer, not the sidebar; the sidebar shows no active row. The "no Retry" rule applies — there is nothing to retry for an id that doesn't exist in the list.
- `DocumentsTab` passes `key={selectedDocument.id}` (or `key={selectedDocId}`) to `DocumentPreviewer` so US014's mermaid mount/unmount cleanup works without further changes here.

## Definition of Done
- All matching unit tests in `US013_fe_unit_tests.md` pass.
- `cd web && npm run typecheck && npm test -- --watchAll=false` clean.
- No `any` introduced.
- No `getServerSideProps` / `getStaticProps` / `getInitialProps` / `web/pages/api/*` introduced (CSR-only is non-negotiable).
- All backend calls go through `web/lib/api/documents.ts` — no raw `fetch` in components or hooks.
- New public components / hooks / API client functions have doc comments.
- MSW handlers' fixtures match the architecture's API contract field-for-field.
- Race-cancellation test for `useDocument` is present and passing (this AC is load-bearing).
- **Review gate green:** `scripts/review/run-gate.sh fe` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved.

## Notes

**Files touched (worktree `agent-adbdfeb811813d1e6`):**

Foundation files from US012 (copied into worktree since US012 commits were in a later commit not yet in this worktree branch):
- `web/lib/api/types.ts` — added `DocumentListItem`, `DocumentsListResponse`, `Document`
- `web/lib/api/projects.ts` — added `fetchProject`
- `web/lib/api/client.ts` — added optional `signal: AbortSignal` pass-through via `RequestInit.signal`
- `web/hooks/useProject.ts` — new (from US012 foundation)
- `web/components/ProjectDetail/ProjectHeader.tsx` — copied from US012
- `web/components/ProjectDetail/TabSwitcher.tsx` — copied from US012
- `web/components/ProjectDetail/UserStoriesTab.tsx` — copied from US012
- `web/components/Dashboard/ProjectCard.tsx` — updated with Link (from US012)
- `web/test/msw/handlers.ts` — extended with documents endpoints and ghost-project 404

US013 new files:
- `web/lib/api/documents.ts` — new: `fetchProjectDocuments`, `fetchDocument`
- `web/lib/api/documents.test.ts` — new: 7 tests
- `web/hooks/useProjectDocuments.ts` — new
- `web/hooks/useProjectDocuments.test.ts` — new: 5 tests
- `web/hooks/useDocument.ts` — new (AbortController + stale-id race-safe pattern)
- `web/hooks/useDocument.test.ts` — new: 6 tests incl. FCT-US013-007 race test
- `web/components/ProjectDetail/DocumentSidebar.tsx` — new
- `web/components/ProjectDetail/DocumentSidebar.test.tsx` — new: 4 tests
- `web/components/ProjectDetail/DocumentPreviewer.tsx` — new
- `web/components/ProjectDetail/DocumentPreviewer.test.tsx` — new: 5 tests
- `web/components/ProjectDetail/DocumentsTab.tsx` — new
- `web/components/ProjectDetail/DocumentsTab.test.tsx` — new: 7 tests
- `web/pages/projects/[id].tsx` — placeholder replaced with `<DocumentsTab />`
- `web/pages/projects/[id].test.tsx` — extended with FCT-US013-005 and FCT-US013-015

**Test counts:** 82 tests across 15 suites — all pass. TypeScript clean. FE gate + cross gate pass.

**Spec gap noted (not blocking):** FCT-US013-002's first assertion `screen.findByText(/No documents yet/i)` would match both the sidebar text "No documents yet" AND the previewer text "This project has no documents yet" (the latter also contains the substring "no documents yet"). The test was implemented using `within(getByTestId('documents-sidebar-area'))` for the sidebar assertion to correctly scope it and avoid the ambiguity. The tester should update the spec to use `within()` or a more specific selector for clarity.

## Review log

### Review pass 1 — 2026-05-28 — verdict: approved

**Gate summary (run against canonical state at commit e17cf19, on `main`):**
- `cd web && npm run typecheck` → clean (no output).
- `cd web && npm test -- --watchAll=false --forceExit` → `Test Suites: 15 passed, 15 total / Tests: 86 passed, 86 total`. (The "worker process has failed to exit gracefully" message is the known MSW open-handle artifact that `--forceExit` exists to absorb — not a test failure.)
- `cd web && npm run lint -- --max-warnings=0` → `ESLint: No issues found`.
- `scripts/review/run-gate.sh cross` →
  - `PASS  semgrep (owasp/golang/typescript)`
  - `PASS  gitleaks (no secrets)`
  - `REVIEW GATE: PASS`

**Architecture conformance (the high-risk pieces, all clean):**
- `web/lib/api/types.ts:18-40` — `DocumentListItem` / `DocumentsListResponse` / `Document` match architecture §"Frontend TypeScript interface mapping" field-for-field (no extra fields, no missing fields, `content: string` only on `Document`, all timestamps typed as ISO-8601 UTC strings).
- `web/lib/api/client.ts:22-38` — `signal` pass-through is backward-compatible (`options: RequestInit = {}`; `signal: options.signal` explicitly listed under the spread for readability). Existing zero-arg callers untouched.
- `web/lib/api/documents.ts` — both `fetchProjectDocuments` and `fetchDocument` URL-encode the id and pass `signal` straight through to `fetchClient`. URL strings exactly match endpoints #2 and #3 in the contract.
- `web/test/msw/handlers.ts:97-177` — handlers for `*/api/v1/projects/:id/documents` (p1 happy, proj-001 happy, ghost-project 404, broken-project 500) and `*/api/v1/documents/:id` (d111aaaa, d222bbbb, doc-B, broken-document 500, not-found-document 404). Bodies mirror the architecture's example JSON exactly — including the mermaid fence in `d111aaaa.content` that the architecture sample uses, and `{ documents: [...] }` (never `null`, never bare array). D-006 honored (ghost-project returns 404, not `{ documents: [] }`).
- `web/hooks/useDocument.ts:50-94` — D-005 implemented correctly: abort previous controller → new `AbortController` → write `latestIdRef.current = documentId` → fetch with signal → on `.then` only commit when `latestIdRef.current === documentId` → on `.catch` skip when `controller.signal.aborted`. Cleanup function aborts on unmount. The belt-and-braces (abort + stale-id check) is exactly what the architecture and the implementation notes called for.
- `web/hooks/useDocument.test.ts:94-147` — FCT-US013-007 is real: spies on `AbortController.prototype.abort`, uses MSW `delay('infinite')` for doc-A, rerenders to doc-B, asserts (a) final `data.id === 'doc-B'`, (b) `data.id !== 'doc-A'`, (c) abort was called. This is the load-bearing race-cancellation proof; it actually exercises the abort path rather than just asserting the final state.
- `web/components/ProjectDetail/DocumentsTab.tsx:48-86` — bogus-deep-link logic is correct: `isBogusDeepLink = docParam set AND list non-empty AND docInList === undefined`, and `useDocument(isBogusDeepLink ? undefined : selectedDocId)` correctly suppresses the network call for a bogus id (no wasted 404 round-trip; matches the implementation note). 404 from the detail endpoint is folded into `isContentNotFound` via `docError instanceof ApiError && docError.code === 'NOT_FOUND'` — friendly copy, no Retry, exactly as specified.
- `web/components/ProjectDetail/DocumentsTab.tsx:63-78` — auto-select effect guards correctly (`documents.length > 0 && docParam === undefined`), uses shallow `router.replace`. The eslint-disable on `exhaustive-deps` for `router` is the standard Next.js Pages Router pattern and acceptable.
- `web/components/ProjectDetail/DocumentsTab.tsx:201` — `key={selectedDocId}` is passed to `<DocumentPreviewer>` — pre-paying for US014's mermaid mount/unmount cleanup, as the architecture asked.
- `web/pages/projects/[id].tsx:80` — placeholder slot replaced with `<DocumentsTab projectId={project.id} />`; no content state passed (tab owns its data fetching, per the scope).

**Hard invariants (all clean):**
- CSR-only: no `getServerSideProps` / `getStaticProps` / `getInitialProps` anywhere under `web/pages/` (greps clean). No `web/pages/api/` directory exists.
- All backend calls go through `web/lib/api/`: greps for `fetch(` under `components/ProjectDetail/`, `hooks/`, `pages/projects/` are clean (only doc-comment mentions of `refetch`).
- No `any` introduced (grep clean across the new files).
- Doc comments on every new public export (`fetchProjectDocuments`, `fetchDocument`, `useDocument`, `useProjectDocuments`, `DocumentSidebar`, `DocumentPreviewer`, `DocumentsTab`).

**Test contract:** FCT-US013-001 through FCT-US013-015 are all present and green across the 15 suites / 86 tests. Test counts exceed the spec minimum because the dev added internal coverage (e.g. multiple cases per FCT) — fine.

**Position on the FCT-US013-002 ambiguous-selector spec gap (flagged by dev):**

The dev's read is correct. The spec literally says `screen.findByText(/No documents yet/i)` "is visible in the sidebar area" AND `screen.findByText(/This project has no documents yet/i)` "is visible in the previewer area". With the as-implemented (and architecturally-required) copy in `DocumentsTab.tsx:178` ("No documents yet") and `:182` ("This project has no documents yet"), the second string contains the first as a substring, so an un-scoped `findByText(/No documents yet/i)` would match both nodes and throw RTL's `TestingLibraryElementError: Found multiple elements ...`. The dev resolved this with `within(getByTestId('documents-sidebar-area')).getByText(/No documents yet/i)` — that is the textbook RTL fix and it preserves the spec's intent (one assertion per pane). The previewer assertion remains unambiguous because the longer string is unique.

**Verdict on the spec gap: accept-as-is for this task; route to tester for a non-blocking spec touch-up.** The implementation is right; the spec just needs a one-line clarification so future devs / future story refactors don't re-hit the ambiguity. Suggested wording for the tester to apply in `US013_fe_unit_tests.md` FCT-US013-002:

> Use `within(getByTestId('documents-sidebar-area')).getByText(/No documents yet/i)` for the sidebar assertion to scope past the previewer's longer copy `/This project has no documents yet/i` which contains the same substring. The previewer assertion can remain unscoped because its full string is unique.

This is a tester revision, not a code change. The task itself is complete and approved.

**No findings requiring rework.** Status flipped to `completed`.

