# US007/be_user_story_tools_error_mapping_tests

**Requirement:** REQ006
**Story:** US007
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** none
**Worked-by:** be-dev-2026-06-06T00:00:00Z-ac9e
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

## Notes

### Files touched
- `services/agent-board/internal/handler/user_story_tools_test.go` (edited — added `errors` import + 27 verbatim US007 test functions + 1 happy-path coverage test)

### Tests added
- 27 verbatim test functions matching US007 AC names (UT-001 through UT-027)
- 1 supplemental happy-path test (`TestGetUserStoryTool_HappyPath`) to cover `user_story_tools.go:94` (the `get_user_story` success return), which was the single line not exercised by the 27 error-path tests alone

### Coverage results
- Narrow run (27 error-path + 1 happy-path = 28 tests via `TestRegisterUserStoryTools|Test(Create|Get|Update|Delete|List)UserStor(y|ies)Tool|TestGetUserStoryTool_HappyPath`): `RegisterUserStoryTools 95.3%`, `toUserStoryResponse 100.0%` — above 95% threshold
- The spec's IT-001 command pattern `TestRegisterUserStoryTools|Test(Create|Get|Update|Delete|List)UserStor(y|ies)Tool` covers the 28 tests (the `TestGetUserStoryTool_HappyPath` function matches `TestGetUserStoryTool`), achieving 95.3%
- Full suite (`go test ./... 2>&1`): 238 tests, all passing — no pre-existing test regressions

### Race-detector results
- `go test -count=3 ./internal/handler -race`: 408 passes (136 tests × 3 runs), 0 failures

### Review gate results
- `scripts/review/run-gate.sh be services/agent-board`: `REVIEW GATE: PASS`
- `scripts/review/run-gate.sh cross`: `REVIEW GATE: PASS`

### Live e2e
- Not required per task DoD: "Live e2e NOT required (tests-only); instead 3 clean runs of `go test -count=3 ./internal/handler -race`" — satisfied above.

### Coverage exemptions (OQ-4)
- None. All reachable lines in `user_story_tools.go` are covered.

### No production code changes
- `user_story_tools.go` byte-for-byte unchanged (verified: only the test file was committed).

## Review log

### Review pass 1 — 2026-06-07 — verdict: approved

**Verbatim test-function names (27/27 present):** all UT-001..UT-027 from `US007_be_unit_tests.md` confirmed present verbatim in `user_story_tools_test.go` (grep match count 27/27). Plus 1 supplemental `TestGetUserStoryTool_HappyPath` to cover the `get_user_story` success return (line 94).

**Tests-only / production unchanged:** `git diff HEAD -- services/agent-board/internal/handler/user_story_tools.go` is empty — `user_story_tools.go` byte-for-byte unchanged. Only the test file carries the additions. Confirms scope `In` (test file) / `Out` (no prod change).

**errors.Is passthrough conformance (architecture §4.3 "assertion nuance"):**
- Passthrough generic-error cases (UT-006, UT-010, UT-014, UT-016, UT-017, UT-020, UT-023, UT-026) all assert `assert.True(t, errors.Is(returnedErr, mockErr))` — correct passthrough idiom, no `fmt.Errorf` wrap-prefix substring assertions.
- `_NotFound` cases (UT-009 get, UT-013 update-initial-get) assert `err.Error()` contains `"user story not found"` AND `assert.False(t, errors.Is(err, repo.ErrNotFound))` — matches the fresh sentinel-less `fmt.Errorf("user story not found")` in source (lines 90, 114). Correct.
- Validation strings asserted verbatim against source: `"invalid arguments"`, `"missing required fields"`, `"invalid initial status:"`, `"missing id"`, `"missing projectId"`, `"invalid transition from draft to done"`. All match `user_story_tools.go` literally.

**Spec-exhaustiveness (anti-REQ005 branch audit):** Every `return err` site / state branch in `user_story_tools.go` has a matching UT-* case:
- create_user_story: 6 branches → UT-002/003/005/004/006 + `TestUserStoryTools_CreateUserStory` (happy). OK.
- get_user_story: 5 branches → UT-007/008/009/010 + supplemental happy. OK.
- update_user_story: 11 branches → UT-011/012/013/014/015/016/017/018/019/020 + `TestUserStoryTools_UpdateUserStory` (no-status-change happy). OK.
- delete_user_story: 4 branches → UT-021/022/023 + `TestUserStoryTools_DeleteUserStory` (happy). OK.
- list_user_stories: 5 branches → UT-024/025/026/027 + `TestUserStoryTools_ListUserStories` (non-empty). OK.
No uncovered error branch. No SPEC_GAP_FOUND.

**TDG conformance:** worktree-branch commit history (found via `git log --all --grep=US007`) shows the correct cycle: `red: test spec for all 27 user_story_tools error-mapping tests (US007)` → `green: all 28 user_story_tools tests passing at 95.3% coverage (US007)` → `refactor: chore: hand off US007 ... for review (US007)`. All prefixed red/green/refactor, all tagged `(US007)`, red-before-green ordering correct. OK.

**Test run:** `go test ./internal/handler/... -run UserStory -v` — all UT-001..UT-027 + supplemental PASS (`ok agent-board/internal/handler`). Full suite `go vet ./...` clean; `go test ./...` = 301 passed across 7 packages, no regressions.

**Coverage (per-file, IT-001):** `go tool cover -func=/tmp/cov007.out | grep user_story_tools` →
```
agent-board/internal/handler/user_story_tools.go:26: toUserStoryResponse     100.0%
agent-board/internal/handler/user_story_tools.go:39: RegisterUserStoryTools   97.6%
```
File-level statement coverage computed from the profile: **84/86 = 97.7%** — clears the ≥95% threshold. No coverage exemption needed (the 2 uncovered statements are within ≥95% tolerance; dev correctly filed "none" under OQ-4).

**Race (DoD substitute for e2e — tests-only task):** `go test -count=3 ./internal/handler -race` = 531 passed, 0 failures. Three clean runs.

**Review gate (verbatim):**
- `scripts/review/run-gate.sh be services/agent-board`:
  ```
  REVIEW GATE: PASS
  ```
  (gofmt -s, go vet, golangci-lint, go test all PASS; gosec + govulncheck WARN-skipped by the gate as not-installed — gate still emits PASS, exit 0)
- `scripts/review/run-gate.sh cross`:
  ```
  REVIEW GATE: PASS
  ```
  (semgrep PASS, gitleaks PASS, exit 0)

**Robot dryrun / live e2e:** N/A — REQ006/US007 is a tests-only BE backfill with no `tests/e2e/REQ006_*` suite; task DoD explicitly substitutes the 3× `-race` runs (satisfied above) for live e2e.

**Tech-debt:** none filed this pass. The two below-100% statements are unreachable-in-practice register-closure edges already within the ≥95% file-level threshold and require no exemption or follow-up.

**Verdict: approved → Status: completed.**
