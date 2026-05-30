# US002 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ004_project_detail_page/US002_documents_tab.robot`.

## Why e2e

The following scenarios cannot be verified at a lower layer alone:

1. **Auto-select first document + previewer loads:** the interaction requires the real browser router mounting the page, the real MSW-free backend serving the list, the auto-select `router.replace` being executed in the actual Next.js runtime, and the subsequent content fetch resolving against the real `api-server`. Component tests mock the backend and the router; only an e2e test exercises the full round-trip.
2. **Click sidebar → previewer updates + URL reflects `?doc=`:** requires a real browser navigating shallow-route changes and persisting them in the actual URL bar. JSDOM cannot verify this.
3. **Refresh with `?doc=` rehydrates the same document:** requires an actual browser page reload (not a React re-render) to confirm the URL-sourced state survives a full lifecycle.

The AbortController race condition, error envelopes, ordering correctness, and individual component states are all verified at the unit layer and are not promoted to e2e.

## Scenarios

### E2E-US002-001 — Documents tab auto-selects first document; click another doc updates previewer and URL
- **Tag:** US002, smoke, regression
- **Preconditions:**
  - Next.js frontend at `${WEB_BASE_URL}`, `api-server` at `${API_BASE_URL}`, MCP server running.
- **Setup (data):**
  1. Via MCP `create_project`: create project `"REQ004 US002 E2E <random>"`. Record `projectId`.
  2. Via MCP `create_document`: create two documents for this project:
     - Doc1: `title="First Document"`, `content="# First\n\nHello from doc 1."` — created first.
     - Doc2: `title="Second Document"`, `content="# Second\n\nHello from doc 2."` — created second (will have a later `updated_at`).
  - Record both `documentId` values.
- **Steps:**
  1. Navigate to `${WEB_BASE_URL}/projects/{projectId}`.
  2. Wait for the Documents tab to be active (default).
  3. Wait for the sidebar to show two document titles.
  4. Assert "Second Document" is listed first in the sidebar (most recently updated).
  5. Assert "Second Document" sidebar item is visually active (auto-selected as the first in the list).
  6. Assert the previewer shows "Second Document" content or title.
  7. Assert the URL contains `?doc={doc2Id}`.
  8. Click "First Document" in the sidebar.
  9. Wait for the previewer to update.
  10. Assert the previewer shows "First Document" content.
  11. Assert the URL contains `?doc={doc1Id}`.
  12. Assert "First Document" sidebar item is now active.
- **Expected:**
  - First item in list is the most recently updated document.
  - Auto-selection works (first item pre-selected and in URL).
  - Manual selection updates the previewer and the URL.
- **Cleanup:** none required (data is randomised).

### E2E-US002-002 — Refresh with `?doc=` rehydrates the same document
- **Tag:** US002, regression
- **Preconditions:** same running stack; reuse project and documents from E2E-US002-001 suite setup.
- **Steps:**
  1. Navigate to `${WEB_BASE_URL}/projects/{projectId}?tab=documents&doc={doc1Id}`.
  2. Wait for the previewer to show "First Document" content.
  3. Reload the browser page.
  4. Wait for the page to reload and data to re-fetch.
  5. Assert the previewer still shows "First Document" content.
  6. Assert "First Document" is marked active in the sidebar.
  7. Assert the URL still contains `?doc={doc1Id}`.
- **Expected:**
  - Deep-linked document selection survives a full browser page reload.
- **Cleanup:** none.
