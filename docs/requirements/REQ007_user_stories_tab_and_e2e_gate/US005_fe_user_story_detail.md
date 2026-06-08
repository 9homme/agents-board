# US005/fe_user_story_detail

**Requirement:** REQ007
**Story:** US005
**Track:** FE
**Status:** pending
**Blocked by:** US004_fe_user_stories_list.md
**Worked-by:** 
**Implements:** US005, D-006, Frontend surface (UserStoryDrawer)

## Goal
Build the right-side detail drawer that displays a selected story's full details and its tasks, accessible by clicking a card.

## Scope
- **In:** `UserStoryDrawer` component. `useUserStory` and `useUserStoryTasks` hooks. State management in `UserStoriesTab`.
- **Out:** List view (done in US004). Routing/deep-linking to detail view.

## Files touched (estimated, exclusive)
- `web/lib/api/types.ts`
- `web/lib/api/userStories.ts`
- `web/hooks/useUserStory.ts`
- `web/hooks/useUserStory.test.ts`
- `web/hooks/useUserStoryTasks.ts`
- `web/hooks/useUserStoryTasks.test.ts`
- `web/components/ProjectDetail/UserStoryDrawer.tsx`
- `web/components/ProjectDetail/UserStoryDrawer.test.tsx`
- `web/components/ProjectDetail/UserStoriesTab.tsx`
- `web/test/msw/userStoryDetailHandlers.ts`

## Test contract
The dev must make these tests pass:
- (Track: FE) from `US005_fe_unit_tests.md`: All FCT-* tests.

## Implementation notes
- Add `UserStory`, `Task`, `TasksListResponse` to `types.ts`.
- Expand `UserStoriesTab` to manage `selectedStoryId` state and conditionally render `UserStoryDrawer`.
- `UserStoryDrawer` issues two parallel hook calls: `useUserStory` and `useUserStoryTasks`.
- Implement focus management, `role=dialog`, and escape-to-close for accessibility.

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
