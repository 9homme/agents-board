# US004 — Backfill `project_tools.go` error-mapping tests

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** draft

## Story
As a **future contributor changing `services/agent-board/internal/handler/project_tools.go`**, I want **every repo-error → MCP-error-envelope mapping in `RegisterProjectTools` and its 5 underlying tool handlers to be covered by integration tests**, so that a regression in error mapping (e.g. swallowing a `repo.ErrNotFound` or dropping a `fmt.Errorf` wrap) fails CI immediately instead of silently shipping a tool that returns the wrong error string to MCP clients.

## Acceptance criteria

- **Scenario: `project_tools_test.go` gains the following test functions (verbatim names)**
  - Given the existing `services/agent-board/internal/handler/project_tools_test.go`
  - When the story is complete
  - Then the following new test functions exist:
    1. `TestRegisterProjectTools_RegistersAllFiveTools` — covers `RegisterProjectTools` at 0% coverage by registering against a fresh `mcp.ToolRegistry` and asserting `registry.GetTool("create_project")`, `"get_project"`, `"update_project"`, `"delete_project"`, `"list_projects"` all return `(handler, true)` and an unknown name returns `(nil, false)`.
    2. `TestHandleCreateProject_InvalidArguments` — invalid JSON → error "invalid arguments".
    3. `TestHandleCreateProject_EmptyName` — name="" or whitespace-only → error "name is required and cannot be empty".
    4. `TestHandleCreateProject_RepoError` — projectRepo.CreateProject returns generic err → handler returns same err (passthrough).
    5. `TestHandleGetProject_InvalidArguments`
    6. `TestHandleGetProject_EmptyID`
    7. `TestHandleGetProject_NotFound` — projectRepo.GetProject returns `repo.ErrNotFound` → handler returns `errors.New("project not found")` (distinct from passthrough).
    8. `TestHandleGetProject_GenericError`
    9. `TestHandleUpdateProject_InvalidArguments`
    10. `TestHandleUpdateProject_EmptyID`
    11. `TestHandleUpdateProject_NotFoundOnInitialGet` — initial `GetProject` returns `ErrNotFound` → "project not found".
    12. `TestHandleUpdateProject_GenericErrorOnInitialGet`
    13. `TestHandleUpdateProject_EmptyNameWhenProvided` — `req.Name` provided but trims to empty → "name cannot be empty if provided".
    14. `TestHandleUpdateProject_RepoUpdateError` — initial Get succeeds, UpdateProject returns generic err.
    15. `TestHandleDeleteProject_InvalidArguments`
    16. `TestHandleDeleteProject_EmptyID`
    17. `TestHandleDeleteProject_RepoError`
    18. `TestHandleListProjects_RepoError`
  - **Note on `RegisterProjectTools`.** Per `docs/tech_debt.md` line 55, this function is at 0% coverage. UT-1 above is the load-bearing test that fixes that.
  - **Note on the test harness.** Tester may use a mock `repo.ProjectRepository` (handwritten mock or `testify/mock`) — see OQ-5 in the README for architect direction. Handler functions are invoked directly (no `httptest`) since they are pure `func(ctx, json.RawMessage) (interface{}, error)`.

- **Scenario: each new test exercises the specific uncovered branch**
  - Given a mock `repo.ProjectRepository`
  - And the mock's relevant method is stubbed to either return a value or an error (per the test name)
  - When the corresponding handler closure is invoked with a `json.RawMessage` payload
  - Then the test asserts the returned `(interface{}, error)` pair:
    - For `_InvalidArguments` cases: error is non-nil and contains `"invalid arguments"` (substring match)
    - For `_EmptyID` / `_EmptyName` cases: error is non-nil and contains the exact wording from the source (`"id is required"`, `"name is required and cannot be empty"`, `"name cannot be empty if provided"`)
    - For `_NotFound` cases: error is non-nil and equals (or `errors.Is` / message-contains) `"project not found"` — confirming the `repo.ErrNotFound` was mapped, not passed through
    - For `_GenericError` / `_RepoError` cases: error is non-nil; for handlers that wrap (none in `project_tools.go` today — confirm passthrough), the returned error is the exact mock error (use `errors.Is`)
    - For `RegistersAllFiveTools`: each tool name resolves to a non-nil handler; unknown names return `false`

- **Scenario: per-file coverage hits ≥95%**
  - Given `cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out -run "TestHandle(Create|Get|Update|Delete|List)Project|TestRegisterProjectTools"`
  - When `go tool cover -func=/tmp/handler.out | grep project_tools.go` is inspected
  - Then `project_tools.go` shows **≥95% statement coverage** (today's baseline per `docs/tech_debt.md` lines 50–55: `handleGetProject` 58.3%, `handleUpdateProject` 68.2%, `handleDeleteProject` 70.0%, `handleListProjects` 71.4%, `handleCreateProject` 75.0%, `RegisterProjectTools` 0.0%)
  - And the only uncovered lines (if any) are documented in the test report under OQ-4

- **Scenario: existing tests still pass and behaviour is unchanged**
  - Given `project_tools.go` is **NOT** modified by this story
  - When `cd services/agent-board && go test ./...` runs
  - Then all pre-existing tests pass
  - And all new tests pass
  - And `golangci-lint run ./...` is clean

- **Scenario: no production-code changes**
  - Given `git diff` of the story's commits
  - When inspected
  - Then **only** `services/agent-board/internal/handler/project_tools_test.go` (and optionally a shared mock-repo helper) is modified
  - And `services/agent-board/internal/handler/project_tools.go` is **byte-for-byte unchanged**

## UI / UX flow expectations
**No UI: BE-test only.**

## Out of scope
- **Modifying handler production code.** Tests-only.
- **`document_tools.go` / `task_tools.go` / `user_story_tools.go` / `message.go`** — US005, US006, US007, US008 respectively.
- **`audit_tools.go`** — already covered, not in the tech-debt list.
- **REST API handler tests** (`project_handler.go`, `document_handler.go`) — not in REQ006 scope.

## Dependencies
- None. Independent.

## Notes for the team

- **`RegisterProjectTools` is the load-bearing test.** Hitting it requires constructing a real `*mcp.ToolRegistry` (or whatever the architect blesses in OQ-5) and asserting each of the 5 tool names was registered with a non-nil handler. Tester should plan one focused test that walks `RegisterProjectTools` and pokes the registry.
- **Audit reference.** `docs/tech_debt.md` lines 50–55 for baselines.
- **Mock repo strategy.** po-ba leaves the choice between handwritten mock and `testify/mock` to the tester. Hand-written mocks are usually fewer lines for a 5-method interface; `testify/mock` is friendlier if many of the new tests share setup. Either is acceptable.
- **No status-transition logic to test here.** Project does not have a status state machine (unlike `task` / `user_story`) — so the `_InvalidTransition` test family is NOT applicable.
- **Run locally before pushing:** `cd services/agent-board && go test ./internal/handler -cover -v -run "TestHandle(Create|Get|Update|Delete|List)Project|TestRegisterProjectTools"`.

## Sign-off log
(po-ba appends here on each sign-off pass)
