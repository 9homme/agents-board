# US005/be_document_tools_error_mapping_tests

**Requirement:** REQ006
**Story:** US005
**Track:** BE
**Service:** services/agent-board
**Status:** in_review
**Blocked by:** none
**Worked-by:** be-dev-2026-06-06T00:00:00Z-a5b3
**Implements:** REQ006/US005 AC (all scenarios — 20 verbatim test function names including `TestRegisterDocumentTools_RegistersAllFiveTools` + sub-95% sibling backfill, ≥95% per-file coverage modulo §4.5 exemptions, no production-code change). Architecture §3 US005 touch row + §4.3 cluster-2 mock-repo pattern + §4.5 exemption mechanism + §4.6 local verification command (US005 row).

## Goal
Backfill `document_tools.go` IT-* error-mapping tests so per-file statement coverage clears ≥95%, with the 20 verbatim test functions named in US005 AC, reusing the existing `MockDocumentRepo` at the top of `document_tools_test.go:19-43`. Tests-only.

## Scope
- **In:** Edit `services/agent-board/internal/handler/document_tools_test.go` to add 20 test functions per US005 AC. **Reuse the existing `MockDocumentRepo`** already at `document_tools_test.go:19-43` (architecture §3 US005 row explicit note); extend its `<Op>Func` fields only if a new method needs to be mocked.
- **Out:** Any change to `document_tools.go`. Any change to `project_tools*`, `task_tools*`, `user_story_tools*`, `message*`. Do NOT replace `MockDocumentRepo` with `testify/mock` — keep the established style for this file.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/handler/document_tools_test.go` (edit — add 20 test functions)

## Test contract
Dev makes the 20 verbatim test-function names from US005 AC pass. Includes `TestRegisterDocumentTools_RegistersAllFiveTools`. Tester's `US005_be_unit_tests.md` IT-* IDs map 1:1 onto these names.

## Implementation notes
- **`document_tools.go` WRAPS repo errors** with `fmt.Errorf("failed to <op> document: %w", err)` (architecture §4.3 "assertion nuance" — first bullet). Assertion idiom: `assert.Contains(t, err.Error(), "failed to <op>")`. Do NOT use `assert.ErrorIs(t, err, mockErr)` for wrapped paths (it still passes due to `%w` but the substring check is the documented convention).
- **`_NotFound` branches return `fmt.Errorf("document not found")`** (fresh — sentinel-less). Assert via `assert.Contains(t, err.Error(), "document not found")` AND `assert.False(t, errors.Is(err, repo.ErrNotFound))`.
- **`TestRegisterDocumentTools_RegistersAllFiveTools`** — same shape as US004's analog: register against `mcp.NewToolRegistry()`, assert `create_document` / `get_document` / `update_document` / `delete_document` / `list_documents` each resolve; unknown name returns `(nil, false)`.
- **Read the source FIRST** for exact `fmt.Errorf("failed to <op> document: %w", err)` strings and exact validation strings — assertions must match `document_tools.go` literally.
- **Mock-repo reuse:** `MockDocumentRepo` already exists at `document_tools_test.go:19-43` with `<Op>Func` fields and a delegated implementation of the `DocumentRepository` interface. Use it as-is; add new func-field-driven cases for the 20 new tests without redefining the struct.
- **Coverage check command** (architecture §4.6, US005 row):
  ```
  cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out \
      -run "TestRegisterDocumentTools|Test(Create|Get|Update|Delete|List)Document(s?)Tool"
  go tool cover -func=/tmp/handler.out | grep document_tools.go
  ```
  Must show ≥95% statement coverage on `document_tools.go`.

## Definition of done
- All 20 new test functions present with US005 AC's verbatim names; all green via the local verification command above.
- `cd services/agent-board && go vet ./... && go test ./...` clean across the whole module.
- `document_tools.go` ≥95% statement coverage (modulo any §4.5 exemptions named in the test report).
- `document_tools.go` byte-for-byte unchanged.
- `golangci-lint run ./...` clean.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` + `scripts/review/run-gate.sh cross` both `REVIEW GATE: PASS`.
- **Live e2e NOT required** (tests-only); instead 3 clean runs of `cd services/agent-board && go test -count=3 ./internal/handler -race`.
- Dev set status to `in_review`; tech-lead approved.

## Review log
