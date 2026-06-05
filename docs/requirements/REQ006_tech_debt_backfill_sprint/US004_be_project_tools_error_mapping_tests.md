# US004/be_project_tools_error_mapping_tests

**Requirement:** REQ006
**Story:** US004
**Track:** BE
**Service:** services/agent-board
**Status:** in_review
**Blocked by:** none
**Worked-by:** be-dev-2026-06-05T00-00-00Z-a6c9
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

## Notes

**Files touched:**
- `services/agent-board/internal/handler/project_tools_test.go` (edited — only file)
- `project_tools.go` — byte-for-byte unchanged (confirmed via `git log`)

**Tests added:**
- 18 verbatim UT-* test functions from US004 AC (UT-001 through UT-018)
- 6 additional `TestHandle*Project_Success` / `TestHandleListProjects_Nil*` tests to push the IT-001 filter coverage above the ≥95% threshold
- 5 pre-existing happy-path tests migrated to new mock style (no style mixing)

**Mock conversion:** Replaced `testify/mock`-based `MockProjectRepo` with hand-written func-field style per architecture §4.3. The two styles MUST NOT coexist in one file; all existing tests were migrated.

**Coverage (IT-001 filter — architecture §4.6 command):**
```
RegisterProjectTools   100.0%
handleCreateProject    100.0%
handleGetProject       100.0%
handleUpdateProject     95.5%  (≥95% — one branch: description-only update without name)
handleDeleteProject    100.0%
handleListProjects     100.0%
```
`handleUpdateProject` 95.5% is acceptable: the one uncovered branch (`req.Description != nil` when no name change is requested) is reached by the `TestProjectTools_UpdateProject` legacy test but falls outside the IT-001 filter. No §4.5 exemption needed — threshold is met.

**Live e2e:** Not required per task DoD (architecture §10.4, tests-only task). 3 clean race-detected runs instead:
`go test -count=3 ./internal/handler -race` → 264 passed, 0 failed, 0 races.

**Review gates:**
- `scripts/review/run-gate.sh be services/agent-board` → REVIEW GATE: PASS
- `scripts/review/run-gate.sh cross` → REVIEW GATE: PASS

**Full module:** `go test ./...` → 184 passed; `go vet ./...` clean; `golangci-lint run ./...` clean.

## Review log

### Pass 1
