# US005/be_user_story_detail

**Requirement:** REQ007
**Story:** US005
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** US004_be_user_stories_list.md
**Worked-by:** 
**Implements:** US005, API contract GET /api/v1/user-stories/{id}, GET /api/v1/user-stories/{id}/tasks, D-004

## Goal
Implement the two REST endpoints for retrieving a single user story's detail and its list of tasks.

## Scope
- **In:** `GET /api/v1/user-stories/{id}` and `GET /api/v1/user-stories/{id}/tasks` endpoints.
- **Out:** Modifying the list endpoint. Embedded tasks in the story response.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/handler/user_story_detail_handler.go`
- `services/agent-board/internal/handler/user_story_detail_handler_test.go`
- `services/agent-board/cmd/api-server/main.go`

*(Note: Created separate handler file to prevent merge conflicts with US004 if they ran in parallel, although this is blocked by US004. Still cleaner separation.)*

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US005_be_unit_tests.md`: All applicable UT and IT tests.

## Implementation notes
- `GetUserStory` endpoint returns a bare JSON object (no wrapper, no `taskCount`).
- `GetUserStoryTasks` returns `{"tasks": [...]}`. Re-use existing `ListTasks` from `task_repo.go`.
- Ensure 404 behavior matches API contracts.
- Wire the routes in `cmd/api-server/main.go`.

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- (Track: BE) `go test -coverprofile=/tmp/cov.out ./... && go tool cover -func=/tmp/cov.out` — every production `.go` file in this task's `## Files touched` clears ≥ 80% line coverage, OR the task has a written `## Coverage exemption` section justifying each below-threshold file.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0 AND emits `REVIEW GATE: PASS` on stdout. Also `scripts/review/run-gate.sh cross` exits 0 AND emits `REVIEW GATE: PASS`. If the REQ has Robot e2e suites, `robot --dryrun tests/e2e/REQ007_*/` also passes.
- Dev set status to `in_review` and reported back.

## Review log
