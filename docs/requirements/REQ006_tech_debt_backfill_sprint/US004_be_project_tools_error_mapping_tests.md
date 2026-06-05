# US004/be_project_tools_error_mapping_tests

**Requirement:** REQ006
**Story:** US004
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** none
**Worked-by:**
**Implements:** REQ006/US004 AC (all scenarios — 18 verbatim test function names including `TestRegisterProjectTools_RegistersAllFiveTools` to lift `RegisterProjectTools` from 0%, ≥95% per-file coverage modulo §4.5 exemptions, no production-code change). Architecture §3 US004 touch row + §4.3 cluster-2 mock-repo pattern + §4.5 exemption mechanism + §4.6 local verification command (US004 row).

## Goal
Backfill `project_tools.go` IT-* error-mapping tests so per-file statement coverage clears ≥95% — including `RegisterProjectTools` (currently 0%) — with the 18 verbatim test functions named in US004 AC, using a mock `repo.ProjectRepository` per architecture §4.3. Tests-only.

## Scope
- **In:** Edit `services/agent-board/internal/handler/project_tools_test.go` to add 18 test functions per US004 AC. Use the hand-written mock-repo pattern from architecture §4.3 (preferred) — define `MockProjectRepo` at the top of the file embedding `repo.ProjectRepository` with `<Op>Func func(...)` overrides. `testify/mock` is acceptable if the dev prefers but the file MUST NOT mix the two styles.
- **Out:** Any change to `project_tools.go`. Any change to `document_tools*`, `task_tools*`, `user_story_tools*`, `message*`. Do NOT introduce shared mock infra across the cluster-2 files (each file keeps its own mock per architecture §4.3 lead).

## Files touched (estimated, exclusive)
- `services/agent-board/internal/handler/project_tools_test.go` (edit — add `MockProjectRepo` if absent + 18 test functions)

## Test contract
Dev makes the 18 verbatim test-function names from US004 AC pass. Per architecture §4.3 the cross-product is roughly {`_InvalidArguments`, `_EmptyID` / `_MissingProjectIDOrTitle`, `_NotFound`, `_GenericError`} × {`Create`, `Get`, `Update`, `Delete`, `List`} plus `TestRegisterProjectTools_RegistersAllFiveTools`. Tester's `US004_be_unit_tests.md` IT-* IDs map 1:1 onto these names.

## Implementation notes
- **`project_tools.go` uses PASSTHROUGH error semantics** for most repo errors (architecture §4.3 "assertion nuance" — last bullet). Assertion idiom: `assert.ErrorIs(t, returnedErr, mockErr)`. Do NOT assert on `"failed to <op>"` wrap-prefixes here — `project_tools.go` does not wrap.
- **`_NotFound` branches return a fresh error string** (`errors.New("project not found")`) — assert via `assert.Contains(t, err.Error(), "project not found")` AND `assert.False(t, errors.Is(err, repo.ErrNotFound))`. The fresh-error convention is intentional; do NOT silently wrap.
- **`TestRegisterProjectTools_RegistersAllFiveTools`** — register against `mcp.NewToolRegistry()`, assert each of `create_project` / `get_project` / `update_project` / `delete_project` / `list_projects` resolves via `registry.GetTool(name)`, and an unknown name returns `(nil, false)`. This single test lifts `RegisterProjectTools` from 0% on its own.
- **Mock-repo shape (architecture §4.3 verbatim):** hand-written `MockProjectRepo` that embeds `repo.ProjectRepository` for forward-compat and overrides only the per-test func fields (`CreateProjectFunc`, `GetProjectFunc`, `UpdateProjectFunc`, `DeleteProjectFunc`, `ListProjectsFunc`). Implement the interface methods by delegating to the func fields with a nil-guard so unset funcs no-op.
- **Read the source FIRST** for exact validation strings (e.g. `"projectId and title are required"`) — assertion strings must match `project_tools.go` literally.
- **Coverage check command** (architecture §4.6, US004 row):
  ```
  cd services/agent-board && go test ./internal/handler -coverprofile=/tmp/handler.out \
      -run "TestHandle(Create|Get|Update|Delete|List)Project|TestRegisterProjectTools"
  go tool cover -func=/tmp/handler.out | grep project_tools.go
  ```
  Must show ≥95% statement coverage on `project_tools.go`.

## Definition of done
- All 18 new test functions present with US004 AC's verbatim names; all green via the local verification command above.
- `cd services/agent-board && go vet ./... && go test ./...` clean across the whole module.
- `project_tools.go` ≥95% statement coverage (modulo any §4.5 exemptions named in the test report).
- `project_tools.go` byte-for-byte unchanged.
- `golangci-lint run ./...` clean.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` + `scripts/review/run-gate.sh cross` both `REVIEW GATE: PASS`.
- **Live e2e NOT required** (tests-only, prod unchanged — architecture §10.4); instead 3 clean runs of `cd services/agent-board && go test -count=3 ./internal/handler -race`.
- Dev set status to `in_review`; tech-lead approved.

## Review log
