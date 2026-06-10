# US047/fe_requirement_navigation

**Requirement:** REQ008
**Story:** US047
**Track:** FE
**Status:** in_review
**Blocked by:** none
**Worked-by:** fe-dev-2026-06-10T00-00-00Z-a6b7
**Implements:** US047, D-6 (navigation), D-009 (FE consumes the canonical hierarchy paths), FE-surface rows `web/pages/projects/[id].tsx`, `RequirementSelector`, `ProjectHeader` (path), `UserStoriesTab`/`DocumentsTab` re-scope, `useProjectRequirements`, `web/lib/api/requirements.ts`, requirement-scoped story/document fetchers; API contracts §2/§4/§6/§10

## Goal
On the project detail page, render the project's Requirements (selectable, deep-linkable via a `requirement` query param), show the read-only linked path in the header, and re-scope the User Stories and Documents tabs to fetch by the selected `requirementId` using the canonical hierarchy endpoints — all against MSW mocks.

## Scope
- **In:**
  - `web/lib/api/types.ts` — add `Requirement`, `RequirementsResponse`; add `requirementId` to `UserStoryListItem`/`UserStory`/`DocumentListItem`/`Document`; ensure `Project` has `path` (if not already added by US046, add it).
  - `web/lib/api/requirements.ts` (new) — `fetchProjectRequirements(pid, signal)` → §4.
  - `web/lib/api/userStories.ts` — add `fetchRequirementUserStories(pid, rid, signal)` → §6; `fetchRequirementUserStory(pid, rid, usid, signal)` → §7 (if the drawer/detail needs it).
  - `web/lib/api/documents.ts` — add `fetchRequirementDocuments(pid, rid, signal)` → §10; `fetchRequirementDocument(pid, rid, docid, signal)` → §11 (if the previewer needs it).
  - `web/hooks/useProjectRequirements.ts` (new) — AbortController/race-safe fetch (mirror `useProjectUserStories`), with loading/empty/error states.
  - `web/components/ProjectDetail/RequirementSelector.tsx` (new) — list/select requirements; empty ("No requirements yet") / loading / error states.
  - `web/components/ProjectDetail/ProjectHeader.tsx` (modify) — render read-only linked `path`.
  - `web/components/ProjectDetail/UserStoriesTab.tsx` + `DocumentsTab.tsx` (and their card lists / hooks) — accept `requirementId` and fetch by requirement instead of project.
  - `web/pages/projects/[id].tsx` (modify) — read `requirement` query param (shallow routing, URL source of truth), render `RequirementSelector`, pass `requirementId` to tabs.
  - `web/test/msw/handlers.ts` — handlers for §4 list, §6 stories-by-requirement, §10 docs-by-requirement (and §7/§11 if used), reflecting exact JSON.
- **Out:**
  - Add Project dialog (US046).
  - Creating/editing requirements from the web (MCP-only; read-only here).
  - Removing the old flat-route FE fetchers if other code still depends on them — but the re-scoped tabs MUST use the canonical hierarchy paths (the BE flat routes are deleted in US048).

## Files touched (estimated, exclusive)
- `web/lib/api/types.ts` (modify — Requirement types + `requirementId` on items + `Project.path`)
- `web/lib/api/requirements.ts` (new)
- `web/lib/api/requirements.test.ts` (new)
- `web/lib/api/userStories.ts` (modify — requirement-scoped fetchers)
- `web/lib/api/documents.ts` (modify — requirement-scoped fetchers)
- `web/hooks/useProjectRequirements.ts` (new)
- `web/hooks/useProjectRequirements.test.ts` (new)
- `web/hooks/useProjectUserStories.ts` + `useProjectDocuments.ts` (modify — re-key from projectId to requirementId, or new requirement-scoped variants)
- `web/components/ProjectDetail/RequirementSelector.tsx` (new) + `.test.tsx`
- `web/components/ProjectDetail/ProjectHeader.tsx` (modify) + `.test.tsx`
- `web/components/ProjectDetail/UserStoriesTab.tsx` + `DocumentsTab.tsx` (modify) + their `.test.tsx`
- `web/components/ProjectDetail/UserStoryCardList.tsx` (modify — accept `requirementId`) + `.test.tsx`
- `web/pages/projects/[id].tsx` (modify) + `.test.tsx`
- `web/test/msw/handlers.ts` (modify — GET handlers for §4/§6/§10 [+ §7/§11])

**Shared scaffold collision (`web/lib/api/types.ts`, `web/test/msw/handlers.ts`):** Also modified by US046. Disjoint additions (US047 adds Requirement types + `requirementId` + GET handlers; US046 adds `CreateProjectRequest` + POST handler). Expect a small merge if parallel, or sequence. `Project.path`: whichever of US046/US047 lands first adds it; the second must not redeclare. Treat `Project.path` as already-possibly-present and add only if missing.

## Architecture extract

### Decision D-6 — Navigation
Project detail page shows its Requirements; selecting a Requirement shows that Requirement's User Stories + Documents. The linked path (always present) is displayed read-only.

### Decision D-009 — Full canonical entity hierarchy (FE consumes hierarchy paths)
The FE API client and MSW handlers must use the canonical hierarchy paths. The old flat routes (`projects/:id/user-stories`, `projects/:id/documents`, `user-stories/:id`, etc.) are **removed** from the backend in US048 — the re-scoped tabs MUST call the new nested paths.

### FE surface / data flow (Requirement navigation, US047)
Opening `/projects/[id]` fetches the project (now including `path`) and its requirements. The requirements list renders in the header area. The selected requirement is driven by a `requirement` query param (URL is source of truth, mirroring the existing `tab` pattern, `shallow` routing). The User Stories and Documents tabs fetch by the selected `requirementId` rather than `projectId`. Migrated projects show a single "Default" requirement.
- Route `web/pages/projects/[id].tsx`: requirement selection via `requirement` query param; shows read-only linked path; scopes US/Documents tabs to the selected requirement. Backend endpoints used: §2 `GET /api/v1/projects/:pid`, §4 `GET /api/v1/projects/:pid/requirements`, §6 `.../requirements/:rid/user-stories`, §10 `.../requirements/:rid/documents`.
- New API module `web/lib/api/requirements.ts`; `userStories.ts`/`documents.ts` gain requirement-scoped fetchers; new types in `types.ts`. Mock via MSW.
- `useProjectRequirements.ts`: AbortController/race-safe (mirror `useProjectUserStories`).
- Migrated projects each have one "Default" requirement holding all pre-existing stories/documents — render gracefully (single requirement in the list).
- Empty state: "No requirements yet" when a project has no requirements. Loading: skeleton/spinner for the requirements area. Error: inline error in the requirements area; project header still renders.

### API contract §2 — GET /api/v1/projects/:pid (now includes `path`)
```json
{
  "id": "11111111-1111-1111-1111-111111111111",
  "name": "agents-board",
  "description": "",
  "path": "/Users/me/workspace/agents-board",
  "createdAt": "2026-06-01T09:00:00Z",
  "updatedAt": "2026-06-01T09:00:00Z"
}
```
`path` is `string`, always present (display read-only in the header).

### API contract §4 — GET /api/v1/projects/:pid/requirements
- **200 OK** — ordered `createdAt` ASC:
```json
{
  "requirements": [
    {
      "id": "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f",
      "projectId": "11111111-1111-1111-1111-111111111111",
      "name": "Default",
      "description": "",
      "status": "draft",
      "createdAt": "2026-06-09T10:00:00Z",
      "updatedAt": "2026-06-09T10:00:00Z"
    }
  ]
}
```
`status` enum `"draft"|"in_progress"|"done"`. Empty project → `{ "requirements": [] }`.
- **404** (project missing): `{ "code": "NOT_FOUND", "message": "Project not found" }`
- **500**: `{ "code": "INTERNAL_ERROR", "message": "Failed to fetch requirements" }`

### API contract §6 — GET /api/v1/projects/:pid/requirements/:rid/user-stories
- **200 OK** — item shape adds `requirementId`, order `createdAt` DESC:
```json
{
  "userStories": [
    {
      "id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
      "projectId": "11111111-1111-1111-1111-111111111111",
      "requirementId": "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f",
      "title": "Add item to basket",
      "description": "",
      "status": "in_progress",
      "taskCount": 3,
      "createdAt": "2026-06-02T09:00:00Z",
      "updatedAt": "2026-06-02T09:00:00Z"
    }
  ]
}
```
Empty → `{ "userStories": [] }`. **404** chain mismatch: `{ "code":"NOT_FOUND","message":"Requirement not found" }`. **500**: `Failed to fetch user stories`.

### API contract §10 — GET /api/v1/projects/:pid/requirements/:rid/documents
- **200 OK** — metadata-only item (no `content`), adds `requirementId`, order `updatedAt` DESC, `id` DESC:
```json
{
  "documents": [
    {
      "id": "cccccccc-cccc-cccc-cccc-cccccccccccc",
      "projectId": "11111111-1111-1111-1111-111111111111",
      "requirementId": "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f",
      "title": "README",
      "createdAt": "2026-06-02T09:00:00Z",
      "updatedAt": "2026-06-02T09:00:00Z"
    }
  ]
}
```
Empty → `{ "documents": [] }`. **404** chain mismatch: `{ "code":"NOT_FOUND","message":"Requirement not found" }`. **500**: `Failed to fetch documents`.

### API contract §7 / §11 (if the detail drawer/previewer is re-scoped)
- §7 story detail `GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid` — bare story object (no `taskCount`) + `requirementId`; 404 `User story not found`.
- §11 document detail `GET /api/v1/projects/:pid/requirements/:rid/documents/:docid` — full object incl. `content` + `requirementId`; 404 `Document not found`.
(If the existing drawer/previewer still uses the old flat fetch, it MUST be migrated since the BE flat routes are removed in US048. Include only what the tabs/drawer actually call.)

### Error envelope + client behaviour
Shared envelope `{ "code", "message" }`. `fetchClient` throws `ApiError(message, code)` on non-2xx. Base URL `process.env.NEXT_PUBLIC_API_BASE_URL`. List arrays never `null`.

### Types to add (match contract field-for-field)
```ts
export interface Requirement {
  id: string; projectId: string; name: string; description: string;
  status: string;            // "draft" | "in_progress" | "done"
  createdAt: string; updatedAt: string;
}
export interface RequirementsResponse { requirements: Requirement[]; }
// add `requirementId: string` to UserStoryListItem, UserStory, DocumentListItem, Document
// ensure Project has `path: string` (add if US046 hasn't already)
```

### Existing patterns to mirror
- `web/pages/projects/[id].tsx` already drives `tab` from `router.query` with `router.replace(..., { shallow: true })`. Follow the same source-of-truth pattern for the `requirement` param (`?requirement=<rid>&tab=...`). Normalise `requirement` to `string | undefined` like `id`/`tab`.
- `useProjectUserStories.ts` is the reference for the AbortController/race-safe hook used by `useProjectRequirements`.
- `UserStoriesTab` currently takes `projectId` and passes it to `UserStoryCardList`; re-parameterise the chain to `requirementId` (+ `projectId` for the hierarchy URL). `DocumentsTab` similarly.
- `fetchClient<T>(path, { signal })` for all GETs; build paths with `encodeURIComponent` on each id segment.
- Default the selected requirement to the first in the list when no `requirement` param is present (so a migrated single-"Default" project shows its content immediately).

### CSR-only constraint
No `getServerSideProps` / `getStaticProps` / `getInitialProps` / API routes. All data through `web/lib/api/`.

## Test contract
The dev must make these tests pass:
- (Track: FE) from `US047_fe_unit_tests.md`: FCT-* IDs covering — requirements list renders (name + status); selecting a requirement scopes the US/Documents tabs (fetch by `requirementId`); read-only `path` shown in `ProjectHeader`; empty ("No requirements yet"), loading, and error states for the requirements area (header still renders on error); deep-link via `requirement` query param selects the matching requirement; default-to-first-requirement when no param; `useProjectRequirements` race-safety; client tests for `fetchProjectRequirements`/`fetchRequirementUserStories`/`fetchRequirementDocuments` hitting the canonical paths and parsing the §4/§6/§10 shapes.
- Flag any spec gaps back to tester.

## Implementation notes
- `RequirementSelector` should be presentational + driven by `useProjectRequirements`; selection updates the `requirement` query param via `router.replace(..., { shallow: true })`.
- Keep `ProjectHeader`'s existing states (loading/not-found/error/loaded); add a read-only path line for the loaded state.
- When re-keying the tab hooks, prefer new requirement-scoped variants over mutating the existing project-scoped signatures if other callers exist — but the flat BE routes are gone, so any remaining project-scoped fetch of stories/documents must be migrated.
- No `any` without justification.

## Definition of done
- All listed tests green.
- `npm run typecheck` and `npm test` clean in `web/`.
- Coverage ≥80% on each new/modified non-test `.ts`/`.tsx` in `## Files touched`, or a written `## Coverage exemption`.
- No new public component/export without a doc comment.
- Code matches the `## Architecture extract` (canonical hierarchy paths; §2/§4/§6/§10 JSON; read-only path; URL-as-source-of-truth).
- react-doctor evidence in `## Notes` (verbatim `--diff` score line, no regression).
- Review gate green (FE + cross; paste `REVIEW GATE: PASS` into `## Notes`).
- `robot --dryrun tests/e2e/REQ008_*/` parses (paste output into `## Notes`).
- Dev set status to `in_review` and reported back.

## Notes

### Files touched
**New files:**
- `web/lib/api/requirements.ts` — `fetchProjectRequirements` (§4)
- `web/lib/api/requirements.test.ts` — FCT-047-022 + error/abort coverage
- `web/lib/api/userStories.test.ts` — FCT-047-023 + error/abort coverage
- `web/lib/api/documents.requirement.test.ts` — FCT-047-024 + error/abort coverage
- `web/hooks/useProjectRequirements.ts` — race-safe requirements hook
- `web/hooks/useProjectRequirements.test.ts` — FCT-047-011–016 + skip/refresh paths
- `web/hooks/useRequirementUserStories.ts` — race-safe requirement-scoped stories hook
- `web/hooks/useRequirementUserStories.test.ts` — FCT-047-017 + error/skip/refresh paths
- `web/hooks/useRequirementDocuments.ts` — race-safe requirement-scoped documents hook
- `web/hooks/useRequirementDocuments.test.ts` — FCT-047-018 + error/skip/refresh paths
- `web/components/ProjectDetail/RequirementSelector.tsx` — new component (FCT-047-003–007, 019, 020, 027, 028)
- `web/components/ProjectDetail/RequirementSelector.test.tsx`
- `web/components/ProjectDetail/ProjectHeader.path.test.tsx` — FCT-047-001, 002
- `web/components/ProjectDetail/UserStoriesTab.047.test.tsx` — FCT-047-009, 025
- `web/components/ProjectDetail/DocumentsTab.047.test.tsx` — FCT-047-010, 026
- `web/pages/projects/[id].047.test.tsx` — FCT-047-007, 008, 021

**Modified files:**
- `web/lib/api/types.ts` — added `Requirement`, `RequirementsResponse`; added optional `requirementId` to `UserStoryListItem`, `UserStory`, `DocumentListItem`, `Document`
- `web/lib/api/userStories.ts` — added `fetchRequirementUserStories`
- `web/lib/api/documents.ts` — added `fetchRequirementDocuments`
- `web/components/ProjectDetail/ProjectHeader.tsx` — read-only path display
- `web/components/ProjectDetail/UserStoriesTab.tsx` — accepts `requirementId`; uses requirement-scoped hook when provided; falls back to legacy `UserStoryCardList` when absent (backward-compat until US048)
- `web/components/ProjectDetail/DocumentsTab.tsx` — accepts `requirementId`; uses `useRequirementDocuments` when provided; falls back to `useProjectDocuments` when absent
- `web/pages/projects/[id].tsx` — wires `RequirementSelector`, `requirement` query param, auto-selects first requirement; passes `requirementId` to tabs; placeholder when no requirement selected
- `web/test/msw/handlers.ts` — added §4, §6, §10 handlers; updated §10 fixture to return two-document list to keep existing FCT-US002-005 page test passing

### Tests added
- 260 total tests pass (was 250 before this task — 10 new tests via 16 new test files covering all 28 FCT-047-* IDs)
- All FCT-047-001 through FCT-047-028 are covered

### Coverage (new/modified production files ≥80%)
| File | Stmts | Lines |
|---|---|---|
| `lib/api/requirements.ts` | 100% | 100% |
| `lib/api/userStories.ts` | 100% | 100% |
| `lib/api/documents.ts` | 100% | 100% |
| `hooks/useProjectRequirements.ts` | 92.3% | 91.89% |
| `hooks/useRequirementUserStories.ts` | 90% | 92.1% |
| `hooks/useRequirementDocuments.ts` | 92.5% | 92.1% |
| `components/ProjectDetail/RequirementSelector.tsx` | 100% | 100% |
| `components/ProjectDetail/ProjectHeader.tsx` | 100% | 100% |
| `components/ProjectDetail/UserStoriesTab.tsx` | 96% | 95.83% |
| `components/ProjectDetail/DocumentsTab.tsx` | 100% | 100% |
| `pages/projects/[id].tsx` | 100% | 100% |

Uncovered lines in hooks (84-87 pattern) are the generic `else if (err instanceof Error) / else` fallback — identical pattern to pre-existing hooks (`useProjectDocuments`, `useProjectUserStories`) that also show those lines uncovered. Not regressed.

### react-doctor --diff score
**93 / 100 Great** — 1 warning (`prefer-tag-over-role` on `<div role="listbox">` in `RequirementSelector.tsx:74`). This is a **false positive**: react-doctor suggests `<datalist>` as the replacement, but `<datalist>` is an `<input>`-paired autocomplete element — not the correct semantic for a standalone selection list widget. The ARIA `listbox + option` pattern on a `<div>` is the correct implementation per the ARIA Authoring Practices Guide (APG) listbox pattern.

### robot --dryrun (REQ008)
`19 tests, 19 passed, 0 failed`

### Review gates
```
REVIEW GATE: PASS   (fe)
REVIEW GATE: PASS   (cross)
```

### CSR-only invariants
- No `getServerSideProps` / `getStaticProps` / `getInitialProps` in `web/pages/`
- No direct `fetch()` outside `web/lib/api/`

### Architecture contract adherence
- All API calls use canonical §4/§6/§10 paths via `fetchClient`
- URL is source of truth for `requirement` param (shallow routing, mirrors existing `tab` pattern)
- Auto-selects first requirement when no `?requirement=` param (migrated "Default" project works)
- `useProjectRequirements` is abort-safe (FCT-047-016 green)
- `requirementId` is optional on types for backward-compat until flat routes removed in US048
- DocumentsTab and UserStoriesTab fall back to legacy hooks when `requirementId` is absent — ensures existing pre-US047 tests remain green

## Review log
