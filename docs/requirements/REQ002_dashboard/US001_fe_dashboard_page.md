# US001/fe_dashboard_page

**Requirement:** REQ002
**Story:** US001
**Track:** FE
**Status:** pending
**Blocked by:** US001_fe_scaffold_api_client.md
**Worked-by:** 
**Implements:** view list of projects UI

## Goal
Implement the main `/` route displaying the minimal and beautiful dashboard of projects via a card UI, integrating with the Next.js API client.

## Scope
- **In:** Creating `web/hooks/useProjects.ts` to manage loading, data, and error state. Building `web/components/Dashboard/ProjectList.tsx` and `web/components/Dashboard/ProjectCard.tsx`. Integrating everything in `web/pages/index.tsx`.
- **Out:** Any project management capabilities (create, edit, delete, etc.), filtering, or pagination.

## Files touched (estimated, exclusive)
- `web/hooks/useProjects.ts`
- `web/components/Dashboard/ProjectList.tsx`
- `web/components/Dashboard/ProjectCard.tsx`
- `web/pages/index.tsx`
- `web/pages/index.test.tsx`
- `web/components/Dashboard/ProjectList.test.tsx`

## Test contract
The dev must make these tests pass:
- (Track: FE) from `US001_fe_unit_tests.md`: FCT-002, FCT-003, FCT-004 (refer to the FE test spec for the UI rendering tests, loading state, error state, empty state).

## Implementation notes
- The design should be "minimal beautiful" focusing on clean layouts, typography, and whitespace. Use basic, accessible HTML semantic tags.
- The `useProjects` hook should call the `fetchProjects()` method scaffolded in the API client and manage `isLoading`, `isError`, `error`, and `data` states.
- Handle all edge cases exactly as per US001:
  - Empty state: Show a "no projects" message.
  - Loading state: Show a visual indicator (spinner or text).
  - Error state: Show a friendly error message.
- Ensure strict Next.js Pages Router CSR-only patterns (no `getServerSideProps` / `getStaticProps`). All data fetching happens via the `useEffect` within the custom hook or component.

## Definition of done
- All listed tests green.
- `npm run typecheck` and `npm test` clean in `web/`.
- No `any` types added without justification.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- **Review gate green:** `scripts/review/run-gate.sh fe` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved.

## Review log
