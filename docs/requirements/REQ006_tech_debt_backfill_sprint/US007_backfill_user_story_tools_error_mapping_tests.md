# US007 — Backfill `user_story_tools.go` error-mapping tests

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** draft

## Story
As a **future contributor changing `services/agent-board/internal/handler/user_story_tools.go`**, I want **every repo-error → MCP-error-envelope mapping AND every status-transition guard in `RegisterUserStoryTools` and its 5 underlying tool closures to be covered by integration tests**, so that a regression (e.g. swallowing `repo.ErrNotFound`, dropping the `IsValidTransition` check, or breaking the post-status-change field update) fails CI immediately.

## Acceptance criteria

- **Scenario: `user_story_tools_test.go` gains the following test functions (verbatim names)**
  - Given the existing `services/agent-board/internal/handler/user_story_tools_test.go`
  - When the story is complete
  - Then the following new test functions exist:
    1. `TestRegisterUserStoryTools_RegistersAllFiveTools` — covers `RegisterUserStoryTools` 63.5% by asserting the five tool names resolve.
    2. `TestCreateUserStoryTool_InvalidArguments`
    3. `TestCreateUserStoryTool_MissingProjectIDOrTitle` — empty projectId OR empty title → "missing required fields".
    4. `TestCreateUserStoryTool_DefaultStatusWhenOmitted` — req.Status omitted → defaults to `domain.UserStoryStatusDraft`.
    5. `TestCreateUserStoryTool_InvalidInitialStatus` — req.Status = "in_signoff" (non-draft) → `domain.NewUserStory` returns err → handler wraps as `"invalid initial status: <err>"`.
    6. `TestCreateUserStoryTool_RepoError` — `CreateUserStory` returns generic err → handler returns err directly (no wrap).
    7. `TestGetUserStoryTool_InvalidArguments`
    8. `TestGetUserStoryTool_MissingID` — error: "missing id".
    9. `TestGetUserStoryTool_NotFound` — `repo.ErrNotFound` → `"user story not found"`.
    10. `TestGetUserStoryTool_GenericError` — handler returns err directly (no wrap).
    11. `TestUpdateUserStoryTool_InvalidArguments`
    12. `TestUpdateUserStoryTool_MissingID`
    13. `TestUpdateUserStoryTool_NotFoundOnInitialGet`
    14. `TestUpdateUserStoryTool_GenericErrorOnInitialGet`
    15. `TestUpdateUserStoryTool_InvalidStatusTransition` — `IsValidTransition` returns false → error contains `"invalid transition from <from> to <to>"`.
    16. `TestUpdateUserStoryTool_StatusChange_UpdateUserStoryStatusError` — `UpdateUserStoryStatus` returns err.
    17. `TestUpdateUserStoryTool_StatusChange_PostStatusFieldUpdateError` — status change + title/description provided; `UpdateUserStory` (post-status field-save) returns err.
    18. `TestUpdateUserStoryTool_StatusChange_HappyPath_NoExtraFields` — status change valid, no extra field changes → returns updated story directly.
    19. `TestUpdateUserStoryTool_StatusChange_HappyPath_WithExtraFields` — status change valid + title/description, both repo calls succeed → returns saved story.
    20. `TestUpdateUserStoryTool_NoStatusChange_RepoUpdateError` — no status change, `UpdateUserStory` returns err.
    21. `TestDeleteUserStoryTool_InvalidArguments`
    22. `TestDeleteUserStoryTool_MissingID`
    23. `TestDeleteUserStoryTool_RepoError`
    24. `TestListUserStoriesTool_InvalidArguments`
    25. `TestListUserStoriesTool_MissingProjectID`
    26. `TestListUserStoriesTool_RepoError`
    27. `TestListUserStoriesTool_EmptySliceReturnsEmptyArray` — empty result → `{"userStories": []}` (NOT nil; verify `make([]UserStoryResponse, 0, ...)` behaviour).

- **Scenario: each new test exercises the specific uncovered branch**
  - Given a mock `repo.UserStoryRepository`
  - When the corresponding tool closure is invoked with a `json.RawMessage`
  - Then assertions per the test name:
    - `_InvalidArguments` → error contains `"invalid arguments"`
    - `_MissingProjectIDOrTitle` → error contains `"missing required fields"`
    - `_MissingID` → error contains `"missing id"` OR `"missing projectId"` (per the specific tool)
    - `_NotFound` → error message contains `"user story not found"`
    - `_GenericError` / `_RepoError` → error is the **exact** mock error (passthrough; verify via `errors.Is(returnedErr, mockErr)`)
    - `_*Error` (the `UpdateUserStoryStatus`/`UpdateUserStory` failure cases) → same passthrough assertion
    - `_InvalidStatusTransition` → error matches `"invalid transition from <from> to <to>"`
    - `_HappyPath_*` → returned `interface{}` is a `UserStoryResponse` with expected fields
    - `_EmptySliceReturnsEmptyArray` → returned map has key `"userStories"` whose value is a non-nil slice of length 0

- **Scenario: per-file coverage hits ≥95%**
  - Given `cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out -run "TestRegisterUserStoryTools|Test(Create|Get|Update|Delete|List)UserStor(y|ies)Tool"`
  - When `go tool cover -func=/tmp/handler.out | grep user_story_tools.go` is inspected
  - Then `user_story_tools.go` shows **≥95% statement coverage** (baseline per `docs/tech_debt.md` line 58: `RegisterUserStoryTools` 63.5%)
  - And the only uncovered lines (if any) are documented in the test report under OQ-4

- **Scenario: existing tests still pass and behaviour is unchanged**
  - Given `user_story_tools.go` is **NOT** modified by this story
  - When `cd services/agent-board && go test ./...` runs
  - Then all pre-existing tests pass
  - And all new tests pass
  - And `golangci-lint run ./...` is clean

- **Scenario: no production-code changes**
  - Given `git diff` of the story's commits
  - When inspected
  - Then **only** `services/agent-board/internal/handler/user_story_tools_test.go` (and optionally a shared mock-repo helper) is modified
  - And `services/agent-board/internal/handler/user_story_tools.go` is **byte-for-byte unchanged**

## UI / UX flow expectations
**No UI: BE-test only.**

## Out of scope
- **Modifying handler production code.** Tests-only.
- **`project_tools.go` / `document_tools.go` / `task_tools.go` / `message.go`** — US004, US005, US006, US008.
- **`domain.UserStory.IsValidTransition` unit tests** — already covered in `internal/domain`.

## Dependencies
- None. Independent.

## Notes for the team

- **Behavioural differences from `task_tools.go`.** Unlike `task_tools.go`, `user_story_tools.go` does NOT wrap repo errors with `fmt.Errorf` in most paths — they are returned directly. The test assertions reflect this (passthrough via `errors.Is`, not substring match on a wrap prefix). Read the source before writing the assertions.
- **`UpdateUserStoryTool` has TWO happy-path branches** (status change + no-extra-fields; status change + extra-fields). Both must be tested explicitly.
- **Audit reference.** `docs/tech_debt.md` line 58 for the `RegisterUserStoryTools` baseline.
- **Run locally before pushing:** `cd services/agent-board && go test ./internal/handler -cover -v -run "TestRegisterUserStoryTools|Test(Create|Get|Update|Delete|List)UserStor(y|ies)Tool"`.

## Sign-off log
(po-ba appends here on each sign-off pass)
