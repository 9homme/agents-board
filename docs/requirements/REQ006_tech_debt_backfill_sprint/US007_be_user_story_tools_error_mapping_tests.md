# US007/be_user_story_tools_error_mapping_tests

**Requirement:** REQ006
**Story:** US007
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** none
**Worked-by:**
**Implements:** REQ006/US007 AC (all scenarios — 27 verbatim test function names lifting `RegisterUserStoryTools` from 63.5%, including the passthrough error semantics, ≥95% per-file coverage modulo §4.5 exemptions, no production-code change). Architecture §3 US007 touch row + §4.3 cluster-2 mock-repo pattern + §4.5 exemption mechanism + §4.6 local verification command (US007 row).

## Goal
Backfill `user_story_tools.go` IT-* error-mapping tests so per-file statement coverage clears ≥95%, with the 27 verbatim test functions named in US007 AC. The file has **passthrough** error semantics on most paths (no `fmt.Errorf` wrap); assertion idiom is `errors.Is(returnedErr, mockErr)`. Tests-only.

## Scope
- **In:** Edit `services/agent-board/internal/handler/user_story_tools_test.go` to add 27 test functions per US007 AC. Use a hand-written `MockUserStoryRepo` per architecture §4.3 (or extend an existing one if present).
- **Out:** Any change to `user_story_tools.go`. Any change to siblings.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/handler/user_story_tools_test.go` (edit — add `MockUserStoryRepo` if absent + 27 test functions)

## Test contract
Dev makes the 27 verbatim test-function names from US007 AC pass. Coverage:
- `TestRegisterUserStoryTools_RegistersAllFiveTools`.
- Create / Get / Update / Delete / List error matrix per architecture §4.3 (passthrough variant).
- `UpdateUserStoryTool` status-change branches + `_InvalidStatusTransition`.

Tester's `US007_be_unit_tests.md` IT-* IDs map 1:1.

## Implementation notes
- **`user_story_tools.go` PASSTHROUGHS repo errors** on most paths (architecture §4.3 "assertion nuance" — second bullet). Assertion idiom: `assert.ErrorIs(t, returnedErr, mockErr)`. Do NOT assert on `"failed to <op>"` wrap-prefixes — `user_story_tools.go` does not wrap these. **Architecture §3 US007 row explicit note:** "verify via `errors.Is(returnedErr, mockErr)`".
- **`_NotFound` returns a fresh `errors.New("user story not found")`** (sentinel-less). Assert via `assert.Contains(t, err.Error(), "user story not found")` AND `assert.False(t, errors.Is(err, repo.ErrNotFound))`.
- **Status-transition tests:** same pattern as US006 — populate `GetUserStoryFunc` to return an existing record with a known status; invoke with a `status` field that maps onto each valid transition + the invalid transition.
- **Mock-repo shape:** hand-written `MockUserStoryRepo` embedding `repo.UserStoryRepository`. Func fields cover every method including `UpdateUserStoryStatus`.
- **Read the source FIRST** for the exact validation strings (the architecture confirms most repo errors passthrough, but the VALIDATION errors like `"id is required"` or `"projectId and title are required"` are fresh — match production literally).
- **Coverage check command** (architecture §4.6, US007 row):
  ```
  cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out \
      -run "TestRegisterUserStoryTools|Test(Create|Get|Update|Delete|List)UserStor(y|ies)Tool"
  go tool cover -func=/tmp/handler.out | grep user_story_tools.go
  ```
  Must show ≥95% statement coverage on `user_story_tools.go`.

## Definition of done
- All 27 new test functions present with US007 AC's verbatim names; all green via the local verification command above.
- `cd services/agent-board && go vet ./... && go test ./...` clean across the whole module.
- `user_story_tools.go` ≥95% statement coverage (modulo any §4.5 exemptions named in the test report).
- `user_story_tools.go` byte-for-byte unchanged.
- `golangci-lint run ./...` clean.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` + `scripts/review/run-gate.sh cross` both `REVIEW GATE: PASS`.
- **Live e2e NOT required** (tests-only); instead 3 clean runs of `cd services/agent-board && go test -count=3 ./internal/handler -race`.
- Dev set status to `in_review`; tech-lead approved.

## Review log
