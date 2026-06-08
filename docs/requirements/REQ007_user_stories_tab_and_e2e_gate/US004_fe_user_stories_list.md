# US004/fe_user_stories_list

**Requirement:** REQ007
**Story:** US004
**Track:** FE
**Status:** pending
**Blocked by:** 
**Worked-by:** 
**Implements:** US004, D-005, Frontend surface (UserStoriesTab, UserStoryCardList, UserStoryCard)

## Goal
Build the User Stories tab UI that fetches and renders a grid of user-story cards with truncated descriptions and task counts.

## Scope
- **In:** Replacing the "Coming soon" placeholder in `UserStoriesTab`. API client and types for user stories list. `UserStoryCardList` and `UserStoryCard` components.
- **Out:** Detail drawer (US005). Real-time updates. Filtering/sorting.

## Files touched (estimated, exclusive)
- `web/lib/api/types.ts`
- `web/lib/api/userStories.ts`
- `web/hooks/useProjectUserStories.ts`
- `web/hooks/useProjectUserStories.test.ts`
- `web/components/ProjectDetail/UserStoriesTab.tsx`
- `web/components/ProjectDetail/UserStoryCardList.tsx`
- `web/components/ProjectDetail/UserStoryCardList.test.tsx`
- `web/components/ProjectDetail/UserStoryCard.tsx`
- `web/components/ProjectDetail/UserStoryCard.test.tsx`
- `web/test/msw/userStoriesHandlers.ts`
- `web/pages/projects/[id].tsx`

## Test contract
The dev must make these tests pass:
- (Track: FE) from `US004_fe_unit_tests.md`: All FCT-* tests.

## Implementation notes
- Extend `types.ts` with `UserStoryListItem` (with `taskCount`) and `UserStoriesListResponse`.
- Create `fetchProjectUserStories` in `userStories.ts`.
- `useProjectUserStories` should mirror `useProjectDocuments` abort/race-safe pattern.
- The `UserStoryCard` should truncate the `description` to ~80 chars.
- Wire `projectId` into `UserStoriesTab` in `pages/projects/[id].tsx`.
- Include MSW setup mirroring the backend JSON contract exactly.

## Definition of done
- All listed tests green.
- (Track: FE) `npm run typecheck` and `npm test` clean in `web/`. No `any` types added without justification.
- (Track: FE) `cd web && npm test -- --coverage --watchAll=false --forceExit` — every non-test `.ts` / `.tsx` file in this task's `## Files touched` clears ≥ 80% line coverage, OR the task has a written `## Coverage exemption` section.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- **Review gate green:** `scripts/review/run-gate.sh fe` exits 0 AND emits `REVIEW GATE: PASS` on stdout. Also `scripts/review/run-gate.sh cross` exits 0 AND emits `REVIEW GATE: PASS`. If the REQ has Robot e2e suites, `robot --dryrun tests/e2e/REQ007_*/` also passes.
- react-doctor evidence in Notes.
- Dev set status to `in_review` and reported back.

## Review log
