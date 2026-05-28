# REQ004 — project_detail_page

## Summary
Extend the dashboard so a user can click a project card and land on a per-project detail page. The detail page is organized into two tabs: **Documents** (functional this requirement) and **User Stories** (placeholder only — content deferred to a later requirement). The Documents tab shows a sidebar list of the project's documents; selecting one renders its markdown content (including mermaid diagrams) in a previewer pane.

## Business Goal
Give human users a read-only knowledge-base view into a project's documents that AI agents have been authoring via MCP since REQ001. Until now, those documents have been invisible to humans — they were write-only data feeding agent context. This requirement turns the documents into a viewable, navigable artifact in the UI.

## Confirmed Decisions
> All items below are locked. Items marked **[confirmed 2026-05-20]** were explicitly confirmed by the user on 2026-05-20 during Phase 1 intake.

- **Document source = existing `documents` table from REQ001.** Documents are not user-uploaded in this requirement; they are the rows agents create through MCP `create_document` / `update_document`. **[confirmed 2026-05-20]**
- **New REST endpoints needed on `services/agent-board` (`api-server`):** **[confirmed 2026-05-20]**
  - `GET /api/v1/projects/{projectId}` — fetch single project (powers the detail page header).
  - `GET /api/v1/projects/{projectId}/documents` — list documents (id, title, updatedAt) for the sidebar.
  - `GET /api/v1/documents/{documentId}` — fetch full document (including `content`) for the previewer.
  - System Architect will lock the exact JSON shapes in Phase 1.
- **Markdown dialect = GitHub-Flavored Markdown (GFM)** with mermaid diagrams (rendered as SVG) and fenced-code-block syntax highlighting. Tables, task lists, fenced code blocks, strikethrough all in scope. **[confirmed 2026-05-20]**
- **Mermaid:** fenced code blocks with language `mermaid` are rendered as SVG diagrams (not as raw code).
- **Code syntax highlighting:** fenced code blocks with a recognized language are syntax-highlighted. Library choice (e.g. `rehype-highlight`, `shiki`) is the System Architect's call. **[confirmed 2026-05-20]**
- **Math (KaTeX):** out of scope. **[confirmed 2026-05-20]**
- **Raw / rendered toggle:** out of scope — always rendered.
- **Document list ordering:** flat (no grouping), ordered by `updatedAt` desc.
- **URL design:** project detail at `/projects/[id]`. Active tab and selected document persisted in the query string as `?tab=documents&doc=<documentId>` so refresh and shared links restore state.
- **Click-from-dashboard entry point:** the existing `ProjectCard` in REQ002 is NOT currently clickable. **This requirement (in US001) makes it clickable** so it navigates to `/projects/[id]`. **[confirmed 2026-05-20]**
- **Previewer error isolation:** when a single document's content fetch fails, the error is shown **in the previewer pane only** with a Retry affordance. The sidebar stays fully usable so the user can pick another document. The whole page does NOT fail. **[confirmed 2026-05-20]**
- **User Stories tab placeholder copy:** the tab content displays the exact string `"Coming soon — user stories will appear here in a future release."` No fetch, no rendering. Real content is a separate future requirement. **[confirmed 2026-05-20]**
- **Auth:** none — consistent with REQ001 + REQ002 which both excluded auth.

## Out of scope (whole requirement)
- Creating / editing / deleting documents from the UI (still MCP-only).
- Content of the User Stories tab (header + empty state only).
- Document search, filter, pagination.
- Document version history.
- Math rendering (KaTeX).
- Document export / download.
- Sharing / permissions / auth.
- Mobile-specific responsive design beyond "doesn't break" — desktop-first is acceptable.

## Dependencies
- **REQ001 (`agent_board_mcp`)** — provides the `documents` table and the data; the new REST endpoints read from it.
- **REQ002 (`dashboard`)** — provides the project list page, the `ProjectCard` component (to be made clickable), the `api-server` executable, the `web/lib/api/` client layer, and the FE routing host.

## Glossary
- **Project detail page** — the new page at `/projects/[id]` that this requirement introduces.
- **Documents tab** — one of two tabs on the project detail page; this requirement makes it functional.
- **User Stories tab** — the other tab; placeholder only in this requirement.
- **Document** — a row in the existing `documents` table (REQ001), with `title` and `content` (treated as markdown for display purposes).
- **Previewer** — the right-hand pane that renders the selected document's markdown.
- **Sidebar** — the left-hand list of documents within the Documents tab.

## User Stories
- `US001_navigate_to_project_detail_with_tabs.md` — clickable project card, `/projects/[id]` page, header with project info, tab switcher (Documents default + User Stories placeholder).
- `US002_documents_tab_list_and_select.md` — sidebar list of project documents, select to load and display content in the previewer (plain text rendering — rich markdown layered in US003).
- `US003_markdown_and_mermaid_rendering.md` — upgrade the previewer to render GFM markdown + mermaid diagrams + syntax-highlighted code fences.

## Tasks
| Task File | US | Title | Track | Service | Blocked By | Status |
|---|---|---|---|---|---|---|
| `US001_be_get_project_endpoint.md` | US001 | GET /api/v1/projects/{id} single-project endpoint | BE | services/agent-board | None | pending |
| `US001_fe_project_card_link.md` | US001 | Make ProjectCard clickable as a link to /projects/{id} | FE | — | None | pending |
| `US001_fe_detail_page_with_tabs.md` | US001 | /projects/[id] page shell — header + tab switcher + useProject + MSW | FE | — | None | completed |
| `US002_be_list_documents_endpoint.md` | US002 | GET /api/v1/projects/{id}/documents endpoint + ListDocuments SQL ordering change | BE | services/agent-board | `US001_be_get_project_endpoint.md` | pending |
| `US002_be_get_document_endpoint.md` | US002 | GET /api/v1/documents/{id} single-document endpoint | BE | services/agent-board | `US002_be_list_documents_endpoint.md` | pending |
| `US002_fe_documents_tab.md` | US002 | Documents tab — sidebar + previewer (plain) + hooks + API client + types + MSW + signal pass-through | FE | — | `US001_fe_detail_page_with_tabs.md` | pending |
| `US003_fe_markdown_renderer.md` | US003 | MarkdownRenderer + sanitization + syntax highlighting + DocumentPreviewer body swap | FE | — | `US002_fe_documents_tab.md` | pending |
| `US003_fe_mermaid_diagram.md` | US003 | MermaidDiagram lazy-loaded component + wire into MarkdownRenderer code override | FE | — | `US003_fe_markdown_renderer.md` | pending |

### Task slicing rationale (tech-lead)

- **BE/FE parallelism:** BE and FE tasks for the same story have **no cross-track `Blocked by`** — they meet only at the architecture's locked API contract. FE tasks consume the contract via MSW; real integration is proven in Phase 3c by e2e.
- **Intra-track sequencing (BE):** US001 BE is the scaffold task for `cmd/api-server/main.go`. US002 BE list-documents is sequenced after it (shares `main.go` for route registration). US002 BE get-document is sequenced after list-documents because both write to the same new `internal/handler/document_handler.go` and to `main.go`. This serialises three small BE PRs but keeps each one collision-free.
- **Intra-track sequencing (FE):**
  - `US001_fe_project_card_link.md` is independent (only touches `ProjectCard.tsx` + its test) — parallelisable with any other FE task in the queue.
  - `US001_fe_detail_page_with_tabs.md` is the scaffold task for `web/test/msw/handlers.ts` and the `/projects/[id].tsx` placeholder slot.
  - `US002_fe_documents_tab.md` is sequenced after `US001_fe_detail_page_with_tabs.md` because both write to `web/test/msw/handlers.ts` AND `US002` replaces the Documents-tab placeholder slot in `web/pages/projects/[id].tsx` that `US001` creates.
  - `US003_fe_markdown_renderer.md` is sequenced after `US002_fe_documents_tab.md` because it modifies `DocumentPreviewer.tsx` (created by US002 FE), and is also the scaffold task for the markdown-stack additions to `web/package.json`.
  - `US003_fe_mermaid_diagram.md` is sequenced after `US003_fe_markdown_renderer.md` because both write to `web/package.json` (mermaid dep) and `MarkdownRenderer.tsx` (`code` override).
- **No BE for US003:** US003 is FE-only per architecture (FE-only sanitization is the trust boundary per the Option D risk note).
- **Cohesion choice on `US002_fe_documents_tab.md`:** the user-spec slicing guidance noted this could optionally split into "sidebar+list" vs "previewer+content fetch". Kept as one task because the race-cancellation AC ties the hook + previewer + sidebar wiring tightly together; splitting them would force one half to ship behind feature flags and create a synchronization burden across two PRs. Tradeoff: this is the largest FE task in the REQ (~1.5–2 days). If the dev finds it ballooning, the natural split point is `useDocument + DocumentPreviewer + content MSW` versus `useProjectDocuments + DocumentSidebar + list MSW + DocumentsTab orchestration`, with the latter blocked by the former.


## Open questions for the human
None — all five clarifying questions were resolved on 2026-05-20 and folded into Confirmed Decisions above.
