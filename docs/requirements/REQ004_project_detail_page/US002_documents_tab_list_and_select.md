# US002 — Documents tab: list documents and view content

**Requirement:** REQ004 — project_detail_page
**Status:** draft

## Story
As a user on a project detail page, I want to see a sidebar list of the project's documents and click one to view its content in a previewer pane, so that I can navigate the project's knowledge base.

> Scope note: this story delivers the **structural** Documents tab — list + selection + previewer container + content fetch + display the content. The previewer in this story may render content as plain text / `<pre>` / minimal styling. **Rich markdown + mermaid rendering is US003.** Splitting this way lets US002 ship a working, useful slice (you can already browse what the document says) while US003 layers visual fidelity on top.

## Acceptance criteria

- **Scenario: Documents tab loads the list for the project**
  - Given I am on `/projects/{id}?tab=documents` (or with the Documents tab active by default) for a project that has at least one document
  - When the page finishes loading
  - Then I see a sidebar listing each document's title
  - And the list is ordered by `updatedAt` descending (most recently updated first)
  - And exactly one document is auto-selected by default (the first item in the list)
  - And the previewer pane shows the auto-selected document's content

- **Scenario: Selecting a document loads its content into the previewer**
  - Given the document list has multiple items and one is currently selected
  - When I click a different document in the sidebar
  - Then the previewer pane updates to show the newly selected document's content
  - And the sidebar visually marks the newly selected document as active (e.g. highlighted row, different background)
  - And the URL query string updates to `?tab=documents&doc={selectedDocumentId}` (using shallow routing so the browser back button can return me to the previously selected document)

- **Scenario: Deep-link to a specific document**
  - Given I open `/projects/{id}?tab=documents&doc={documentId}` directly (via bookmark or shared link)
  - When the page finishes loading
  - Then the sidebar marks `{documentId}` as the selected document
  - And the previewer shows that document's content
  - And no other document is shown as selected

- **Scenario: Deep-link to a document that doesn't exist for this project**
  - Given I open `/projects/{id}?tab=documents&doc=<bogusId>`
  - When the page resolves (list loads but the deep-linked doc isn't in it)
  - Then the previewer shows a friendly "Document not found" message
  - And no sidebar item is marked active
  - And the sidebar remains usable so I can click any real document to recover

- **Scenario: Empty state — project has no documents**
  - Given I am on the Documents tab of a project that has zero documents
  - When the list resolves
  - Then the sidebar shows a friendly "No documents yet" empty state
  - And the previewer shows a corresponding empty state (e.g. "Select a document to view it" — but with no items to select, copy adjusts to "This project has no documents yet")
  - And no document fetch is attempted

- **Scenario: Loading state — list is being fetched**
  - Given the Documents tab has just become active and the list fetch is in flight
  - When the response has not yet arrived
  - Then the sidebar shows a loading indicator (skeleton rows or spinner)
  - And the previewer shows a neutral loading or empty state (no content yet)

- **Scenario: Loading state — content is being fetched**
  - Given the list has loaded and a document has been selected (auto or by click)
  - When the content fetch (`GET /api/v1/documents/{id}`) is in flight
  - Then the previewer shows a loading indicator while keeping the sidebar interactive
  - And clicking a different sidebar item while loading cancels/supersedes the in-flight fetch (the previewer ends up showing the most recently selected document, not whichever resolved last)

- **Scenario: Error — list fetch fails**
  - Given the `GET /api/v1/projects/{id}/documents` call fails (network or 5xx)
  - When the failure resolves
  - Then the sidebar shows a friendly error message with a "Retry" affordance
  - And the previewer shows a neutral state ("Document list unavailable")
  - And clicking Retry re-issues the list fetch

- **Scenario: Error — content fetch fails**
  - Given the list has loaded but a `GET /api/v1/documents/{id}` call fails for the selected document
  - When the failure resolves
  - Then the previewer shows a friendly error message ("Failed to load document") with a "Retry" affordance
  - And the sidebar remains fully usable (user can click other documents)
  - And clicking Retry re-issues the content fetch for the currently selected document

## UI / UX flow expectations

- **Entry points:** The Documents tab is the default tab on `/projects/{id}` (see US001). Users land here from the dashboard click-through or by direct URL.

- **Layout (within the Documents tab content area):**
  ```
  +-----------------------+---------------------------------------+
  |  Documents (N)        |                                       |
  |  -------------------  |                                       |
  |  > Architecture       |     {selected document title}         |   ← h2
  |    Onboarding guide   |     {muted: updated {timestamp}}      |
  |    API conventions    |     -----------------------------     |
  |    Deployment runbook |                                       |
  |    ...                |     {document content}                |   ← previewer area
  |                       |                                       |
  |  (sidebar — ~280px)   |     (previewer — fluid width)         |
  +-----------------------+---------------------------------------+
  ```

- **Happy-path flow:**
  1. User lands on the Documents tab.
  2. Sidebar lists every document for the project, titles only, most-recent-first.
  3. The first document is auto-selected; its content streams into the previewer.
  4. User clicks any other title → sidebar highlight moves; previewer swaps to that doc's content; URL `?doc=<id>` updates.
  5. User can navigate by repeatedly clicking titles. Browser back/forward steps through the selection history.

- **Sidebar specifics:**
  - Fixed width on desktop (~240–320px); scrollable if the list overflows the viewport.
  - Each item shows the document `title`. Long titles truncate with ellipsis (single line) — full title visible on hover via `title` attribute.
  - The selected item is visually distinct (background fill + slightly bolder weight).
  - Items are keyboard-focusable; Enter / Space activate them.
  - Header above the list shows the document count, e.g. "Documents (4)".

- **Previewer specifics for US002:**
  - The previewer is just a container that displays the document's `content` field.
  - For this story, rendering may be plain text or `<pre>`-formatted — it's allowed to look unstyled. **US003 upgrades this to GFM markdown + mermaid + code highlighting.** Tests in US002 only assert that `content` is reachable to the user (e.g. as text), not that markdown is rendered as HTML.
  - The selected document's `title` is shown as a heading above the content; `updatedAt` is shown muted beside or below it.

- **Empty / loading / error states:**
  - **List loading:** skeleton rows in sidebar; previewer shows "Loading documents…" or stays blank.
  - **List empty:** sidebar shows "No documents yet" message; previewer shows "This project has no documents yet".
  - **List error:** sidebar shows "Couldn't load documents" + Retry button.
  - **Content loading:** previewer shows a spinner / skeleton; sidebar remains usable.
  - **Content error:** previewer shows "Failed to load document" + Retry button; sidebar remains usable.
  - **Deep-link to non-existent doc:** previewer shows "Document not found"; sidebar selection is empty.

- **Validation rules visible to the user:** none.

- **Accessibility expectations:**
  - The sidebar is a `nav` landmark with `role="listbox"` or `<ul>` + items as `<button>`s. Selected item exposes `aria-selected="true"`.
  - The previewer is a `region` with an accessible name (the document title).
  - Keyboard: Tab into the sidebar, arrow keys move selection, Enter / Space activate.

- **Out of UI scope:**
  - GFM markdown / mermaid rendering (US003).
  - Editing / deleting / creating documents.
  - Search within documents.
  - Document grouping (by type, by US, etc.).

## Out of scope
- Rendering content as anything more than reachable text/`<pre>` in this story (rendering polish = US003).
- Hierarchical document organization.
- Search, filter, multi-select.
- Edit / delete / create.

## Dependencies
- **US001** (this requirement) — provides the page shell + tab switcher within which this story lives.
- **New backend endpoints:**
  - `GET /api/v1/projects/{projectId}/documents` — returns the list. Expected response shape (System Architect to confirm/lock) mirrors REQ001's MCP `list_documents` result and uses REQ002's REST envelope conventions:
    ```json
    { "documents": [{ "id": "...", "projectId": "...", "title": "...", "updatedAt": "..." }] }
    ```
    `content` deliberately excluded from the list response — fetched on demand.
  - `GET /api/v1/documents/{documentId}` — returns the full document including `content`:
    ```json
    { "id": "...", "projectId": "...", "title": "...", "content": "...", "createdAt": "...", "updatedAt": "..." }
    ```
  - `404` body matches REQ002's error envelope (`{ "code": "NOT_FOUND", "message": "..." }`).

## Notes for the team
- **Why two endpoints (list + detail) rather than one fat list with content?** Documents are markdown — `content` can be large. Loading every doc's full content for the sidebar is wasteful. List returns metadata only; content is lazy-loaded per selection. This also makes the loading state of selection meaningful.
- **Race condition on rapid clicks:** if the user clicks doc A then doc B before A's content fetch resolves, the previewer must show B, not whichever request happened to finish last. The dev should use a request-id / abort pattern (e.g. AbortController, or react-query's automatic cancellation). The AC "clicking a different sidebar item while loading cancels/supersedes" makes this testable.
- **Shallow routing:** use `router.replace(..., { shallow: true })` when changing `?doc=<id>` so we don't re-trigger the page-level fetches.
- **Where the rendering goes:** keep the previewer as a standalone component (e.g. `web/components/ProjectDetail/DocumentPreviewer.tsx`) whose only input is the loaded document. US003 will replace its internals with the markdown renderer without changing how US002 wires it up.

## Sign-off log
(po-ba appends here on each sign-off pass)
