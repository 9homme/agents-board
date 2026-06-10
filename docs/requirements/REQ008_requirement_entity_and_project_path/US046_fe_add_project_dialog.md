# US046/fe_add_project_dialog

**Requirement:** REQ008
**Story:** US046
**Track:** FE
**Status:** in_progress
**Blocked by:** none
**Worked-by:** fe-dev-2026-06-10-a9a6
**Implements:** US046, D-005 (plain text path input, no autocomplete), D-006 (name + path required), FE-surface rows `web/pages/index.tsx`, `AddProjectDialog`, `useCreateProject`, `web/lib/api/projects.ts createProject`, API contract §3 `POST /api/v1/projects`

## Goal
Add an "Add Project" entry point + dialog on the dashboard where the user types a name and a plain-text local path (name auto-fills from basename, sticky-off after manual edit), submits via `createProject`, and sees inline 400/409 errors — wired against MSW mocks of `POST /api/v1/projects`.

## Scope
- **In:**
  - `web/lib/api/types.ts` — add `CreateProjectRequest`; add `path: string` to `Project`.
  - `web/lib/api/projects.ts` — add `createProject(req): Promise<Project>` (POST §3).
  - `web/hooks/useCreateProject.ts` — submit state machine (idle/submitting/error), distinguishing 400 vs 409.
  - `web/components/Dashboard/AddProjectDialog.tsx` — the form (plain text path + name, basename auto-fill, client-side non-blank validation, in-flight + inline error states).
  - `web/pages/index.tsx` — "Add Project" button that opens the dialog; refresh the projects list on success.
  - `web/test/msw/handlers.ts` — handler for `POST /api/v1/projects` (201 / 400 / 409) reflecting §3.
- **Out:**
  - Requirement navigation / project detail page (US047).
  - Any filesystem autocomplete / `web/lib/api/fs.ts` / suggestions (removed this REQ — does not exist).
  - Editing a project's path after creation.

## Files touched (estimated, exclusive)
- `web/lib/api/types.ts` (modify — add `CreateProjectRequest`, add `path` to `Project`)
- `web/lib/api/projects.ts` (modify — add `createProject`)
- `web/lib/api/projects.test.ts` (modify)
- `web/hooks/useCreateProject.ts` (new)
- `web/hooks/useCreateProject.test.ts` (new)
- `web/components/Dashboard/AddProjectDialog.tsx` (new)
- `web/components/Dashboard/AddProjectDialog.test.tsx` (new)
- `web/pages/index.tsx` (modify — Add Project button + dialog wiring)
- `web/pages/index.test.tsx` (modify)
- `web/test/msw/handlers.ts` (modify — POST /api/v1/projects handler)

**Shared scaffold collision (`web/lib/api/types.ts`, `web/test/msw/handlers.ts`):** Both US046 and US047 modify `types.ts` and `handlers.ts`. US046 adds `CreateProjectRequest` + `path` on `Project`; US047 adds `Requirement`/`RequirementsResponse`/`requirementId` and the GET handlers. The additions are disjoint (different declarations), but the orchestrator should expect a small merge if both run in parallel, OR sequence them. Neither blocks the other logically (no shared API contract dependency). Keep edits additive and localized.

## Architecture extract

### Decision D-005 — No filesystem autocomplete; plain text path input
Path input is a plain `<input type="text">`. User types the full path manually; name auto-fills from basename. No `FSHandler`, no suggestion logic, no `web/lib/api/fs.ts`, no `PathAutocomplete`, no `usePathSuggestions`.

### Decision D-006 — `path` required everywhere
The create form enforces non-blank `name` + `path` client-side, backed by API validation. Submit disabled until both non-blank.

### FE surface / data flow (Add Project, US046)
The user opens the dashboard, clicks "Add Project", types a name and a local path in plain text inputs. The name auto-fills from the path basename (editable). On submit, `createProject` POSTs name + path; the server validates existence/`IsDir` and uniqueness, returns the created project (or 400/409 surfaced inline). On success the dialog closes and the projects list refreshes.
- Entry point: dashboard (`web/pages/index.tsx`). Backend endpoint used: `POST /api/v1/projects`.
- API client layer: every backend call lives in `web/lib/api/`; components never call `fetch` directly. `web/lib/api/projects.ts` gains `createProject`. New shapes in `web/lib/api/types.ts`. Mock at this boundary via MSW (`web/test/msw/handlers.ts`).
- Name auto-fill must be "sticky-off" once the user manually edits the name, so path changes don't clobber a deliberate name.
- The FE only enforces non-blank client-side; existence/is-directory/uniqueness are server-authoritative. The FE distinguishes 400 (code `VALIDATION_ERROR`) vs 409 (code `DUPLICATE_PATH`) to render the right inline message. On error the form stays open with input preserved.
- Open question assumption (locked): on success → close dialog + refresh list (not navigate).

### API contract §3 — POST /api/v1/projects
- **Request body:**
```json
{ "name": "agents-board", "description": "", "path": "/Users/me/workspace/agents-board" }
```
`name` required non-blank; `description` optional (default `""`); `path` required non-blank.
- **201 Created** — bare project object including `path`:
```json
{
  "id": "33333333-3333-3333-3333-333333333333",
  "name": "agents-board",
  "description": "",
  "path": "/Users/me/workspace/agents-board",
  "createdAt": "2026-06-09T11:00:00Z",
  "updatedAt": "2026-06-09T11:00:00Z"
}
```
- **400 Bad Request** (code `VALIDATION_ERROR`) — blank name/path, or path missing-on-disk / not-a-directory:
```json
{ "code": "VALIDATION_ERROR", "message": "path is required" }
{ "code": "VALIDATION_ERROR", "message": "path does not exist or is not a directory" }
{ "code": "VALIDATION_ERROR", "message": "name is required" }
```
- **409 Conflict** (code `DUPLICATE_PATH`):
```json
{ "code": "DUPLICATE_PATH", "message": "path already linked to another project" }
```
- **500**: `{ "code": "INTERNAL_ERROR", "message": "Failed to create project" }`

### Error envelope + client behaviour
Shared envelope `{ "code", "message" }`. The existing `fetchClient` (`web/lib/api/client.ts`) throws `ApiError(message, code)` on non-2xx. `useCreateProject` should catch `ApiError` and branch on `error.code` (`VALIDATION_ERROR` → 400 inline message; `DUPLICATE_PATH` → "Path already linked to another project"). Base URL via `process.env.NEXT_PUBLIC_API_BASE_URL`.

### Types to add (match contract field-for-field)
```ts
export interface Project {            // EXISTING — add path
  id: string; name: string; description: string;
  path: string;                       // NEW
  createdAt: string; updatedAt: string;
}
export interface CreateProjectRequest {
  name: string;
  description?: string;
  path: string;
}
```

### MSW handler outline (`web/test/msw/handlers.ts`)
`POST ${base}/api/v1/projects` returning 201 with a `Project` (echo name/path), or 400 `{code:'VALIDATION_ERROR',...}` / 409 `{code:'DUPLICATE_PATH',...}` per the test case being exercised. Mirror the exact JSON above.

### CSR-only constraint
No `getServerSideProps` / `getStaticProps` / `getInitialProps` / API routes. All data flows through `web/lib/api/`.

### Existing patterns to mirror
- `web/lib/api/projects.ts` `fetchProject` uses `fetchClient<T>(path, { signal })`. For POST: `fetchClient<Project>('/api/v1/projects', { method: 'POST', body: JSON.stringify(req) })` — `fetchClient` already sets `Content-Type: application/json`.
- Dashboard list refresh: `web/hooks/useProjects.ts` drives `web/pages/index.tsx`; re-trigger its fetch after a successful create (expose a refetch or re-mount key).
- Basename: `path.split('/').filter(Boolean).pop() ?? ''` (handle trailing slash; pure string op, no Node `path` module needed in CSR).

## Test contract
The dev must make these tests pass:
- (Track: FE) from `US046_fe_unit_tests.md`: the FCT-* IDs covering — open dialog from dashboard; path field is plain text (no suggestions); name auto-fills from basename and is editable; auto-fill sticky-off after manual name edit; submit disabled until both non-blank; successful create closes dialog + refreshes list; 400 path-invalid inline error with input preserved; 409 duplicate-path inline error; in-flight disabled/loading state prevents double-submit. Plus `createProject` client test (201 returns `Project` with `path`; ApiError on 400/409 carrying `code`).
- Flag any spec gaps back to tester.

## Implementation notes
- `useCreateProject` shape suggestion: `{ submit(req): Promise<Project>, status: 'idle'|'submitting'|'error', error: ApiError|null, reset() }`. Guard against double-submit by ignoring calls while `status === 'submitting'`.
- Keep `AddProjectDialog` controlled; preserve inputs on error (do not reset on failure).
- No `any` without justification.

## Definition of done
- All listed tests green.
- `npm run typecheck` and `npm test` clean in `web/`.
- Coverage ≥80% on each new/modified non-test `.ts`/`.tsx` in `## Files touched`, or a written `## Coverage exemption`.
- No new public component/export without a doc comment.
- Code matches the `## Architecture extract` (§3 JSON + 400/409 codes; plain text input; sticky-off auto-fill).
- react-doctor evidence in `## Notes` (verbatim `--diff` score line, no regression).
- Review gate green (FE + cross; paste `REVIEW GATE: PASS` into `## Notes`).
- `robot --dryrun tests/e2e/REQ008_*/` parses (paste output into `## Notes`).
- Dev set status to `in_review` and reported back.

## Notes

## Review log
