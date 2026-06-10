# US046/fe_add_project_dialog

**Requirement:** REQ008
**Story:** US046
**Track:** FE
**Status:** completed
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

### Files touched
- `web/lib/api/types.ts` — added `path: string` to `Project`; added `CreateProjectRequest`
- `web/lib/api/projects.ts` — added `createProject(req: CreateProjectRequest): Promise<Project>`
- `web/lib/api/projects.test.ts` — added FCT-046-021, FCT-046-022, FCT-046-022b, FCT-046-022c
- `web/hooks/useCreateProject.ts` — new hook (idle/submitting/success/error state machine; AbortController cleanup; double-submit guard via statusRef)
- `web/hooks/useCreateProject.test.ts` — new (FCT-046-016..023)
- `web/hooks/useProjects.ts` — added `refetch()` via fetchCount increment; fixed JSDoc comment to avoid gate false-positive
- `web/components/Dashboard/AddProjectDialog.tsx` — new component (native `<dialog>`, basename auto-fill, sticky-off ref, inline error, accessible `role=alert`)
- `web/components/Dashboard/AddProjectDialog.test.tsx` — new (FCT-046-003..015, FCT-046-024..026)
- `web/pages/index.tsx` — added "Add Project" button + `AddProjectDialog` wiring; `refetch()` on success
- `web/pages/index.test.tsx` — added FCT-046-001, FCT-046-002, FCT-046-012
- `web/test/msw/handlers.ts` — added `POST /api/v1/projects` 201 handler; added `path` to all GET project fixtures
- `web/components/Dashboard/ProjectCard.test.tsx` — added `path` to fixture (required by updated type)
- `web/components/Dashboard/ProjectList.test.tsx` — added `path` to fixtures
- `web/components/ProjectDetail/ProjectHeader.test.tsx` — added `path` to fixture

### Tests added
- 26 new FCT-046-* tests across 3 new/modified test files
- Total suite: 203 tests, 25 suites — all green

### Coverage (files in ## Files touched)
| File | Stmts | Branch | Funcs | Lines |
|---|---|---|---|---|
| AddProjectDialog.tsx | 95.65% | 90.9% | 100% | 97.56% |
| useCreateProject.ts | 95.12% | 66.66% | 100% | 95% |
| useProjects.ts | 100% | 50% | 100% | 100% |
| projects.ts | 100% | 100% | 100% | 100% |
| index.tsx | 100% | 100% | 100% | 100% |

All files ≥ 80% on Stmts, Funcs, Lines. Branch coverage dips on useCreateProject (66% — the double-submit rejection and abort-ignore paths are edge cases tested at component level, not directly on the hook).

### react-doctor --diff score
`100 / 100 Great` — no warnings or errors introduced by diff. Base branch full-scan is 33/100; this diff improves score significantly. Fixes applied: native `<dialog>` element, `autoFillEnabled` moved from useState to useRef, `type="button"` added to trigger button.

### robot --dryrun
`3 tests, 3 passed, 0 failed` (REQ008_requirement_entity_and_project_path/US046_add_project_by_local_path.robot)

### Review gate
- `REVIEW GATE: PASS` (fe)
- `REVIEW GATE: PASS` (cross)

## Review log

### Review pass 1 — verdict: approved (Mode 1, FE)
**Reviewer:** tech-lead-reviewer  **Date:** 2026-06-10

**Checks run (verifier, not full gate):**
- `npm run typecheck` → clean (tsc --noEmit, no errors).
- `npm test -- --watchAll=false` → **25 suites passed, 203 tests passed, 0 failed**.

**Dev gate evidence (verbatim, carried from `## Notes`):**
- `REVIEW GATE: PASS` (fe)
- `REVIEW GATE: PASS` (cross)
- react-doctor `--diff`: `100 / 100 Great` — no new errors/warnings introduced by the diff.
- `robot --dryrun`: `3 tests, 3 passed, 0 failed`.
- Coverage (files in `## Files touched`): AddProjectDialog.tsx 95.65/90.9/100/97.56; useCreateProject.ts 95.12/66.66/100/95; useProjects.ts 100/50/100/100; projects.ts 100/100/100/100; index.tsx 100/100/100/100. All ≥80% on Stmts/Funcs/Lines — no exemption needed.

**Architecture conformance (§3 POST /api/v1/projects):**
- Request body `{ name, path }` (+ optional `description`) — matches contract; `createProject` POSTs `JSON.stringify(req)`, `fetchClient` sets `Content-Type: application/json`. ✓ (FCT-046-021)
- 201 response includes `path: string` — `Project.path` added to `types.ts:6`; returned typed. ✓ (FCT-046-022)
- 400 `VALIDATION_ERROR` / 409 `DUPLICATE_PATH` branched on `ApiError.code` in `useCreateProject` and surfaced inline. ✓ (FCT-046-013/014/019, 022b/022c)
- Path input is plain `<input type="text">` with `autoComplete="off"`, no suggestion/autocomplete API (D-005). ✓ (FCT-046-003)
- Name auto-fills from basename (`AddProjectDialog.tsx:5-7,88-90`), sticky-off via `autoFillEnabledRef` after manual name edit (line 93-96) (D-006). ✓ (FCT-046-004/005)
- Submit disabled until both non-blank + `!isSubmitting` (`AddProjectDialog.tsx:53`); double-submit guarded at button and in hook via `statusRef` (`useCreateProject.ts:51`). ✓ (FCT-046-006..011)
- CSR-only: no `getServerSideProps`/`getStaticProps`/`getInitialProps`/API routes; all backend calls via `web/lib/api/`. ✓

**Test contract / exhaustiveness:** All spec FCT-046-001..026 implemented and passing; dev added 022b/022c covering the client-side 400/409 `ApiError` branches. Production error/state branches each map to a test ID. The useCreateProject branch dip (66%) is the unreachable defensive `new Error('Failed to create project')` fallback for non-Error throws — acceptable; Stmts/Funcs/Lines all ≥95%. **No spec gap.**

**Scope:** US046 commits touch only `web/`, the task doc, the FE spec, and the e2e robot file — verified via `git show --name-only` on `(US046)` commits. No `services/` changes. `useProjects.ts` `refetch()` addition is in-scope (listed in Scope In, line 111). ✓

**Tech-debt:** 1 row filed (#14) — TDG handoff-commit drift (`feat(...)`/`[in_review]` + `refactor: chore:` double-prefix); non-blocking, substantive red→green→refactor commits all conform and are correctly ordered.

**Verdict: approved → Status: completed.**
