# US005 — Backfill `document_tools.go` error-mapping tests

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** in_signoff

## Story
As a **future contributor changing `services/agent-board/internal/handler/document_tools.go`**, I want **every repo-error → MCP-error-envelope mapping in `RegisterDocumentTools` and its 5 underlying tool closures to be covered by integration tests**, so that a regression (e.g. swallowing a `repo.ErrNotFound`, dropping a `fmt.Errorf("failed to ...: %w", err)` wrap, or breaking the `mapDocumentToResponse` time-formatting) fails CI immediately.

## Acceptance criteria

- **Scenario: `document_tools_test.go` gains the following test functions (verbatim names)**
  - Given the existing `services/agent-board/internal/handler/document_tools_test.go`
  - When the story is complete
  - Then the following new test functions exist:
    1. `TestRegisterDocumentTools_RegistersAllFiveTools` — covers the partial 69.2% on `RegisterDocumentTools` by asserting `registry.GetTool("create_document"|"get_document"|"update_document"|"delete_document"|"list_document")` all resolve.
    2. `TestCreateDocumentTool_InvalidArguments`
    3. `TestCreateDocumentTool_MissingProjectIDOrTitle` — empty projectId OR empty title → "projectId and title are required".
    4. `TestCreateDocumentTool_RepoError` — `CreateDocument` returns generic err → handler wraps as `"failed to create document: <err>"` (verify wrap text).
    5. `TestGetDocumentTool_InvalidArguments`
    6. `TestGetDocumentTool_EmptyID`
    7. `TestGetDocumentTool_NotFound` — `repo.ErrNotFound` → handler returns `"document not found"`.
    8. `TestGetDocumentTool_GenericError` — handler wraps as `"failed to get document: <err>"`.
    9. `TestUpdateDocumentTool_InvalidArguments`
    10. `TestUpdateDocumentTool_EmptyID`
    11. `TestUpdateDocumentTool_NotFoundOnInitialGet` — `GetDocument` returns `ErrNotFound` → `"document not found"`.
    12. `TestUpdateDocumentTool_GenericErrorOnInitialGet` — wraps as `"failed to get document: <err>"`.
    13. `TestUpdateDocumentTool_UpdateRepoError` — initial Get succeeds, `UpdateDocument` returns err → wraps as `"failed to update document: <err>"`.
    14. `TestDeleteDocumentTool_InvalidArguments`
    15. `TestDeleteDocumentTool_EmptyID`
    16. `TestDeleteDocumentTool_RepoError` — wraps as `"failed to delete document: <err>"`.
    17. `TestListDocumentsTool_InvalidArguments`
    18. `TestListDocumentsTool_MissingProjectID`
    19. `TestListDocumentsTool_RepoError` — wraps as `"failed to list documents: <err>"`.
    20. `TestListDocumentsTool_EmptySliceReturnsEmptyDocumentsArray` — `ListDocuments` returns `nil` or `[]` → response is `{"documents": []}` (NOT `nil`).
  - **Note on response shape.** Several success-path tests may already exist; do not re-add them. Tester checks the existing file and only adds the gaps above.

- **Scenario: each new test exercises the specific uncovered branch**
  - Given a mock `repo.DocumentRepository`
  - When the corresponding tool closure (obtained via `registry.GetTool(name)`) is invoked with a `json.RawMessage`
  - Then for `_InvalidArguments`: error contains `"invalid arguments"`
  - And for `_MissingProjectIDOrTitle`: error contains `"projectId and title are required"`
  - And for `_EmptyID`: error contains `"id is required"`
  - And for `_NotFound`: error message equals or contains `"document not found"` AND `errors.Is(err, repo.ErrNotFound)` is **false** (the handler unwraps to a fresh error)
  - And for `_GenericError` / `_RepoError`: error message contains the exact wrap prefix (`"failed to <op> document: "`)
  - And for `_EmptySliceReturnsEmptyDocumentsArray`: the returned `interface{}` is a `map[string]interface{}` with key `"documents"` whose value is a non-nil slice of length 0

- **Scenario: per-file coverage hits ≥95%**
  - Given `cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out -run "TestRegisterDocumentTools|Test(Create|Get|Update|Delete|List)Document(s?)Tool"`
  - When `go tool cover -func=/tmp/handler.out | grep document_tools.go` is inspected
  - Then `document_tools.go` shows **≥95% statement coverage** (baseline per `docs/tech_debt.md` line 56: `RegisterDocumentTools` 69.2%)
  - And the only uncovered lines (if any) are documented in the test report under OQ-4

- **Scenario: existing tests still pass and behaviour is unchanged**
  - Given `document_tools.go` is **NOT** modified by this story
  - When `cd services/agent-board && go test ./...` runs
  - Then all pre-existing tests pass
  - And all new tests pass
  - And `golangci-lint run ./...` is clean

- **Scenario: no production-code changes**
  - Given `git diff` of the story's commits
  - When inspected
  - Then **only** `services/agent-board/internal/handler/document_tools_test.go` (and optionally a shared mock-repo helper) is modified
  - And `services/agent-board/internal/handler/document_tools.go` is **byte-for-byte unchanged**

## UI / UX flow expectations
**No UI: BE-test only.**

## Out of scope
- **Modifying handler production code.** Tests-only.
- **`project_tools.go` / `task_tools.go` / `user_story_tools.go` / `message.go`** — US004, US006, US007, US008.
- **REST `document_handler.go`** — not in scope.

## Dependencies
- None. Independent.

## Notes for the team

- **Document handler wraps errors more aggressively than `project_tools.go`.** Almost every branch uses `fmt.Errorf("failed to <op> document: %w", err)`. Asserting the wrap text is the test's load-bearing job — if a future contributor removes the wrap, these tests must fail.
- **`mapDocumentToResponse` time formatting** is implicitly exercised by every success-path test that already exists; we are NOT adding a dedicated test for it (unless coverage profiling shows otherwise).
- **Audit reference.** `docs/tech_debt.md` line 56 for the `RegisterDocumentTools` baseline.
- **Run locally before pushing:** `cd services/agent-board && go test ./internal/handler -cover -v -run "TestRegisterDocumentTools|Test(Create|Get|Update|Delete|List)Document(s?)Tool"`.

## Sign-off log
(po-ba appends here on each sign-off pass)
