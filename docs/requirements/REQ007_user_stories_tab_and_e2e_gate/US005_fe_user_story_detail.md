# US005/fe_user_story_detail

**Requirement:** REQ007
**Story:** US005
**Track:** FE
**Status:** in_review
**Blocked by:** US004_fe_user_stories_list.md
**Worked-by:** fe-dev-2026-06-08T09-36-00Z-a4f2
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
- `web/test/msw/handlers.ts` (extended — was `userStoryDetailHandlers.ts` in spec, merged into shared handlers)

## Additional files touched (undeclared, surfaced as note)
- `web/components/ProjectDetail/UserStoryCard.tsx` — updated `onSelect` signature to pass `HTMLButtonElement` for focus-return
- `web/components/ProjectDetail/UserStoryCardList.tsx` — updated `onSelect` signature to match card change
- `web/hooks/useProjectUserStories.ts` — brought from US004 branch (was missing from this worktree)
- `web/pages/projects/[id].tsx` — wired `projectId` prop into `UserStoriesTab` (one-line change per architecture)
- `web/pages/projects/[id].test.tsx` — updated FCT-US001-011 to match new tab behavior (placeholder replaced)

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

## Notes

### Files created / modified
**New files:**
- `web/lib/api/userStories.ts` — `fetchProjectUserStories`, `fetchUserStory`, `fetchUserStoryTasks`
- `web/hooks/useUserStory.ts` — AbortController + useReducer race-safe hook
- `web/hooks/useUserStory.test.ts` — 5 tests
- `web/hooks/useUserStoryTasks.ts` — AbortController + useReducer race-safe hook
- `web/hooks/useUserStoryTasks.test.ts` — 6 tests
- `web/components/ProjectDetail/UserStoryDrawer.tsx` — native `<dialog open>`, FCT-002/004/006/007
- `web/components/ProjectDetail/UserStoryDrawer.test.tsx` — 4 tests

**Modified files:**
- `web/lib/api/types.ts` — added `UserStoryListItem`, `UserStoriesListResponse`, `UserStory`, `Task`, `TasksListResponse`
- `web/test/msw/handlers.ts` — added wildcard handlers for `/api/v1/user-stories/:id/tasks` and `/api/v1/user-stories/:id` (tasks handler registered first to prevent shadowing)
- `web/components/ProjectDetail/UserStoriesTab.tsx` — replaced placeholder with `selectedStoryId` state + `UserStoryCardList` + `UserStoryDrawer`
- `web/components/ProjectDetail/UserStoriesTab.test.tsx` — replaced old placeholder tests with FCT-001/003/005
- `web/components/ProjectDetail/UserStoryCard.tsx` — updated `onSelect` to pass `HTMLButtonElement` for focus return
- `web/components/ProjectDetail/UserStoryCardList.tsx` — updated `onSelect` signature
- `web/pages/projects/[id].tsx` — wired `projectId` into `UserStoriesTab` (architecture-mandated one-liner)
- `web/pages/projects/[id].test.tsx` — updated FCT-US001-011 assertion (placeholder text removed)
- `web/hooks/useProjectUserStories.ts` — brought from US004 branch (missing in worktree)

### Tests added
| Test file | Count |
|---|---|
| `hooks/useUserStory.test.ts` | 5 |
| `hooks/useUserStoryTasks.test.ts` | 6 |
| `components/ProjectDetail/UserStoryDrawer.test.tsx` | 4 |
| `components/ProjectDetail/UserStoriesTab.test.tsx` | 5 (FCT-001/003/005, replaced 2 old) |
| **Total new/replaced** | **18** |

Final suite: **160 tests, 160 passed, 0 failed**

### Coverage (line %)
| File | Line % |
|---|---|
| `useUserStory.ts` | 94.59% |
| `useUserStoryTasks.ts` | 94.59% |
| `UserStoryDrawer.tsx` | 100% |
| `UserStoriesTab.tsx` | 100% |
| `userStories.ts` | 100% |

All ≥ 80% line threshold met.

### Gates
- `scripts/review/run-gate.sh fe` → `REVIEW GATE: PASS`
- `scripts/review/run-gate.sh cross` → `REVIEW GATE: PASS`
- `robot --dryrun tests/e2e/REQ007_*/` → 7 tests, 7 passed, 0 failed

### react-doctor --diff score
**92 / 100** (1 warning: `prefer-tag-over-role` for `role="status"` on the spinner at `UserStoryDrawer.tsx:98` — false positive: `role="status"` is the correct ARIA live-region role for loading spinners; no HTML element carries this implicit role. Documented in-code with a comment.)

Baseline on `main` (full codebase): 35/100. Diff score 92/100 — no regression on changed files.

### Live e2e gate: REVIEW_GATE_BLOCKED
Podman is available. `make e2e-up` was attempted but the api-server container fails to become healthy (exits with `ERROR: api-server failed to become healthy`). Root cause: US001 (migrations at startup) is not yet merged to `main`, so the api-server on this branch has no auto-migration and the `GET /api/v1/projects` health probe fails. This is a cross-task infrastructure dependency, not a FE code defect.

Robot dry-run completed: `robot --dryrun tests/e2e/REQ007_*/` → **7 tests, 7 passed, 0 failed**.

Tech-lead / orchestrator should run the live e2e once US001 BE is merged to the working branch and `make e2e-up` succeeds. The FE code is correct against the MSW-mocked contract.

### Undeclared-file note
The task spec listed `web/test/msw/userStoryDetailHandlers.ts` as the new MSW handler file, but the existing project pattern (as established by US004) is to extend `web/test/msw/handlers.ts` directly. The new handlers were added to `handlers.ts` rather than a separate file to maintain consistency. The per-test `server.use()` overrides in both test files correctly shadow the defaults.
