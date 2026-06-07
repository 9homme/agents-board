# US005 — Backfill `document_tools.go` error-mapping tests

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** done

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

### Sign-off pass 1 — 2026-06-07 — verdict: approved
- **Spec review:** All 5 AC scenarios are covered. The 20 verbatim test-function names in AC scenario 1 map 1:1 onto UT-001..UT-020 in `US005_be_unit_tests.md`; coverage matrix is complete. Branch/error-path scenario 2 is faithfully specced (invalid args, missing fields, empty id, NotFound-unwrap, wrap-prefix assertions, empty-slice → non-nil `{"documents":[]}`). NotFound cases correctly assert both the `"document not found"` substring AND `errors.Is(err, repo.ErrNotFound) == false`. Tool name resolved to `list_documents` (confirmed against `document_tools.go:147`), matching UT-001's "confirm exact name" note. No e2e/unit mis-leveling — tests-only story, correctly BE-unit-only.
- **Result review:** Verified independently, not trusted from the report.
  - All 20 AC test functions present in `internal/handler/document_tools_test.go` (lines 205–556) plus 5 pre-existing `TestDocumentTools_*` happy-path tests. Scoped run: 20 passed. Full module: 301 tests passed, 0 failed (matches report). No skips, no `t.Skip`.
  - Production code unchanged: `git diff` on `document_tools.go` is empty; working tree clean; only the test file was touched. AC scenario 5 satisfied.
  - Per-file coverage on `document_tools.go` measured at **100.0%** (`mapDocumentToResponse` 100%, `RegisterDocumentTools` 100%) when the package's full test set runs — clears the ≥95% target. AC scenario 3 met in substance.
- **Finding (non-blocking, accepted):** The IT-001 coverage command as written in this story (scenario 3) and in `US005_be_unit_tests.md` uses `-run "TestRegisterDocumentTools|Test(Create|Get|Update|Delete|List)Document(s?)Tool"`. That regex EXCLUDES the pre-existing `TestDocumentTools_*` happy-path tests, so running IT-001 verbatim yields `mapDocumentToResponse 0.0%`, `RegisterDocumentTools 89.2%` — below 95%. The ≥95% target is only demonstrably met by the full-package run. The defect is in the AC/spec measurement command, not in the code or behavior: the real per-file coverage is 100% and all required branches are exercised. Approving because the substantive intent (≥95% per-file coverage, no prod-code change, 20 named tests all green) is fully and verifiably satisfied. Recommend the tester correct IT-001's `-run` regex to append `|TestDocumentTools` (or drop the `-run` filter) so a future contributor running the documented command sees the true figure; the test report's coverage line should also cite the exact command used. This is a documentation/measurement fix, not a behavioral one, and does not warrant blocking or re-running the pipeline.
- **Routed to:** none (approved). Advisory follow-up for tester to tighten the IT-001 command wording (non-blocking).
