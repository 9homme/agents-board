# US001/fe_scaffold_api_client

**Requirement:** REQ002
**Story:** US001
**Track:** FE
**Status:** completed
**Blocked by:** 
**Worked-by:** fe-dev
**Implements:** GET /api/v1/projects API contract, API client layer

## Goal
Scaffold the Next.js API client, shared types, and MSW handlers corresponding to the architecture's API contracts, enabling decoupled frontend UI work.

## Scope
- **In:** Creating `web/lib/api/types.ts` for the Project data types. Creating `web/lib/api/client.ts` as a base API fetch wrapper. Creating `web/lib/api/projects.ts` containing the `fetchProjects()` method. Updating or creating `web/test/msw/handlers.ts` to mock the REST endpoint.
- **Out:** Building the UI components or hooks (handled in subsequent task).

## Files touched (estimated, exclusive)
- `web/lib/api/types.ts`
- `web/lib/api/client.ts`
- `web/lib/api/projects.ts`
- `web/test/msw/handlers.ts`

This is a **scaffold task**. Other FE tasks for this story are blocked by this task.

## Test contract
The dev must make these tests pass:
- (Track: FE) from `US001_fe_unit_tests.md`: FCT-001 (refer to the FE test spec for the API client tests).

## Implementation notes
- Ensure `web/lib/api/types.ts` exactly matches the JSON properties outlined in `architecture.md` (e.g., `id`, `name`, `description`, `createdAt`, `updatedAt`).
- `client.ts` should read from `NEXT_PUBLIC_API_BASE_URL`.
- `msw/handlers.ts` must provide the exact mock structures so the subsequent UI task can be implemented in total isolation.

## Definition of done
- All listed tests green.
- `npm run typecheck` and `npm test` clean in `web/`.
- No `any` types added without justification.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- **Review gate green:** `scripts/review/run-gate.sh fe` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved.

## Review log
### Review pass 1 — 2026-05-15 — tech-lead
- **Verdict:** approved
- **Findings:** FE scaffold verified. MSW handlers, types, and client implementation match the architecture. Gate is green.
