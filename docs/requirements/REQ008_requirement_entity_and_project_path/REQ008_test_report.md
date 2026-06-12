# REQ008 Test Report — Requirement Entity and Project Path

**Timestamp:** 2026-06-12T09:40:00+07:00  
**Commit SHA:** a9fb4184a93b1c05e982138b51051b4be6cbd740  
**Mode 2 gate:** 3× `make e2e-seed && make e2e-run` — all 49 tests green across all 3 runs

---

## BE Unit / Integration Tests

**Total: 529 passed, 0 failed** across 11 packages (`go test ./...` in `services/agent-board/`)

| Test ID | Go Test Name | Package | Result |
|---|---|---|---|
| UT-044-001 | TestRequirementStatusDraft_IsZeroValueDefault | internal/domain | PASS |
| UT-044-002 | TestRequirementStatusConstants | internal/domain | PASS |
| UT-044-003 | TestRequirementStruct_HasAllRequiredFields | internal/domain | PASS |
| UT-044-004 | TestUserStory_HasRequirementIDField | internal/domain | PASS |
| UT-044-005 | TestDocument_HasRequirementIDField | internal/domain | PASS |
| UT-044-006 | TestProject_HasPathField | internal/domain | PASS |
| IT-044-001–010 | Migration suite (000003_requirement_entity) | internal/migrate | PASS |
| UT-045-* | TestValidatePath_ExistsAndIsDir / DoesNotExist / ExistsButIsFile / EmptyPath | internal/fsutil | PASS |
| UT-045-* | TestRequirementHandler_ListProjectRequirements_* (200/404/500) | internal/handler | PASS |
| UT-045-* | TestRequirementRepo_* | internal/repo | PASS |
| UT-045-* | TestCreateProjectTool_WithPath / WithDuplicatePath / WithNonExistentPath / WithMissingPath | internal/mcp | PASS |
| UT-045-* | TestCreateUserStoryTool_WithRequirementID / WithoutRequirementID | internal/mcp | PASS |
| UT-045-* | TestCreateDocumentTool_WithRequirementID / MissingRequirementID / RequirementNotInProject | internal/mcp | PASS |
| UT-048-* | TestHierarchy_UT048_011_UserStoryListItemIncludesRequirementId | internal/handler | PASS |
| UT-048-* | TestHierarchy_UT048_012_UserStoryDetailIncludesRequirementIdNoTaskCount | internal/handler | PASS |
| UT-048-* | TestHierarchy nested route handler tests (US/Document/Task endpoints) | internal/handler | PASS |

All 529 tests passed. No failures. No skips.

---

## FE Component Tests

**Total: 260 passed, 0 failed** across 36 suites (`npm test -- --watchAll=false` in `web/`)

### US046 — Add Project by Local Path (FCT-046-*)

| Test ID | Description | Result |
|---|---|---|
| FCT-046-001 through FCT-046-026 | AddProjectDialog: form rendering, path validation, name auto-fill, duplicate path error, submit states, dialog lifecycle | PASS (all 26) |

### US047 — Requirement Navigation (FCT-047-*)

| Test ID | Description | Result |
|---|---|---|
| FCT-047-001 through FCT-047-028 | RequirementSelector: list rendering, selection state, empty state, URL sync, UserStoriesTab scoping, DocumentsTab scoping | PASS (all 28) |

### Cross-story FE tests also verified green

| Suite | Tests | Result |
|---|---|---|
| useDocument (FCT-US010-005–009, FCT-US002-007) | hierarchical endpoint, race cancel, reducer behaviour | PASS (all 9) |
| useUserStory | hierarchical endpoint, abort, error states | PASS (all 5) |
| useUserStoryTasks | hierarchical endpoint, abort, empty tasks | PASS (all 5) |
| UserStoryDrawer (FCT-002, FCT-004, FCT-006, FCT-007) | drawer open/close/error/loading via hierarchical routes | PASS (all 4) |
| UserStoriesTab (FCT-001, FCT-003, FCT-005) | card→drawer flow, close, story switching | PASS (all 3) |

---

## E2E Tests

**Total (REQ008 only): 19 passed, 0 failed**  
**Total (full suite): 49 passed, 0 failed** — all suites including pre-existing (REQ001–REQ007) passed all 3 runs.

| Test ID | Description | Result |
|---|---|---|
| E2E-044-001 | Projects list includes path field after migration | PASS |
| E2E-044-002 | Seeded project has Default requirement after migration backfill | PASS |
| E2E-045-001 | POST /api/v1/projects with real directory returns 201 | PASS |
| E2E-045-002 | POST /api/v1/projects with missing path returns 400 | PASS |
| E2E-045-003 | POST /api/v1/projects with non-existent path returns 400 | PASS |
| E2E-045-004 | POST /api/v1/projects with duplicate path returns 409 | PASS |
| E2E-045-005 | GET /api/v1/projects/:pid/requirements returns requirements list | PASS |
| E2E-045-006 | GET /api/v1/projects/unknown/requirements returns 404 | PASS |
| E2E-045-007 | MCP create_user_story with requirement_id succeeds (BREAKING CHANGE) | PASS |
| E2E-045-008 | MCP create_user_story without requirement_id returns tool error | PASS |
| E2E-046-001 | Add Project golden path: form opens, submits, project appears in list | PASS |
| E2E-046-002 | Add Project: server rejects non-existent path shows inline error | PASS |
| E2E-046-003 | Add Project: duplicate path shows DUPLICATE_PATH inline error | PASS |
| E2E-047-001 | Project detail page shows linked path in header and requirements list | PASS |
| E2E-047-002 | Clicking a requirement scopes the user stories tab | PASS |
| E2E-047-003 | New project with no requirements shows empty-state message | PASS |
| E2E-048-001 | Full hierarchy leaf endpoints return 200 (golden path) | PASS |
| E2E-048-002 | Chain mismatch returns 404 (no cross-resource leakage) | PASS |
| E2E-048-003 | All 8 removed flat routes return non-200 | PASS |

---

## Skipped Tests

None. All tests in all three pyramids ran to completion.

---

## Notes

- US049 (add_blocked_review_gate_task_status) remains `draft` — no tasks were created for it during Phase 3 and it has no test spec. Not covered by this test report.
- Pre-existing suite failures in US012, US013, US014, US042, US043 were fixed as part of REQ008 quality gate work (FE callers migrated from removed flat routes to hierarchical endpoints; e2e selectors tightened). All 49 tests now green.
- Seed SQL updated with `TRUNCATE ... RESTART IDENTITY CASCADE` to ensure clean state on each `make e2e-seed` call, enabling reliable consecutive runs without stack restart.
