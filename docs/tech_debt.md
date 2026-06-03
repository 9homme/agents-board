# Tech Debt Backlog

Durable backlog of non-blocking findings raised at code review or sign-off. Tech-lead appends one line per finding BEFORE issuing `approved` on a task (see `.claude/agents/tech-lead.md` §Verdict). Burying findings inside per-task review logs is explicitly disallowed — review logs scatter, this file does not.

**Format (one line per finding):**

```
- YYYY-MM-DD — <file:line> — <what's wrong> — <suggested fix> — REQ[ID]/US[ID]/<task-name>
```

Re-file fixed items by striking through, not by deleting:

```
- ~~2026-05-30 — web/package.json:14 — @testing-library/dom in dependencies instead of devDependencies — move it — REQ004/US003/fe_mermaid_diagram~~ → fixed in REQ005/US007
```

A retrospective at the end of each REQ scans this file and decides which items become a new tech-debt REQ.

---

## Findings

- 2026-06-02 — scripts/review/test/test_run_gate.sh:UT-001 — UT-001 only asserts banner presence, missing body-text + rc value visibility — tighten assertion to also check captured body string and `rc=1` per IT-US001-001 — REQ005/US001/be_fix_printf_double_dash
- 2026-06-02 — scripts/review/test/test_run_gate.sh:UT-003 — UT-003 symmetric tightening required for IT-US001-002 (same as above for `run_check_warn`) — REQ005/US001/be_fix_printf_double_dash
- 2026-06-02 — scripts/review/test/test_run_gate.sh:* — Test-ID naming uses UT-NNN instead of canonical IT-US001-NNN — re-map naming on next touch — REQ005/US001/be_fix_printf_double_dash
- 2026-06-02 — docs/requirements/REQ005_*/US001_be_unit_tests.md:IT-US001-004 — TTY-mode test remains manual (spec allows it; promote to automated when a portable TTY harness is added) — REQ005/US001/be_fix_printf_double_dash
- 2026-06-02 — docs/requirements/REQ005_*/US001_be_fix_printf_double_dash.md:DoD — BE-gate precondition implicitly couples to US003 (gosec/govulncheck soft-warn); add explicit `Depends-on: US003` or rephrase DoD — REQ005/US001/be_fix_printf_double_dash
- 2026-06-03 — services/agent-board/cmd/mcp-server/sse.go,message.go — pre-existing govulncheck finding on stdlib crypto/x509 (transitive via Go stdlib) — bump Go toolchain or pin a fix when a runtime upgrade is available — REQ005/US005/be_project_repo_error_tests

### 2026-06-03 — Pre-existing happy-path-bias coverage gaps across REQ001–REQ004 (same root cause as REQ005/US005, not in REQ005 scope)

These were uncovered by the orchestrator's project-wide test-coverage audit run after REQ005's US005 closed. They are the SAME pattern US005 fixed for `document_repo` and `project_repo`. The REQ004 quality audit named only those two files; the rest of `internal/repo/`, the `internal/handler/*_tools.go` family, and `internal/mcp/server.go` are sitting on identical debt. Backfill via a future REQ (suggested: REQ006 — "repo + handler exhaustiveness backfill"). The new tester exhaustiveness mandate (committed in this REQ) means new code won't reopen this; these are the historic items.

**`internal/repo/` sub-threshold functions (target ≥95% per US005's DoD baseline):**

- 2026-06-03 — services/agent-board/internal/repo/task_repo.go:70 UpdateTask — 62.5% coverage; missing error-branch tests (Query/Scan/rows.Err/NotFound vs GenericError splits) — backfill UT-* per US005 pattern — REQ001-004/tech-debt/repo-error-branches
- 2026-06-03 — services/agent-board/internal/repo/task_repo.go:91 UpdateTaskStatus — 71.4% coverage; same pattern — backfill UT-* — REQ001-004/tech-debt/repo-error-branches
- 2026-06-03 — services/agent-board/internal/repo/task_repo.go:36 CreateTask — 80.0% coverage; GenericError path missing — backfill UT-* — REQ001-004/tech-debt/repo-error-branches
- 2026-06-03 — services/agent-board/internal/repo/task_repo.go:51 GetTask — 87.5% coverage; one error path uncovered — backfill UT-* — REQ001-004/tech-debt/repo-error-branches
- 2026-06-03 — services/agent-board/internal/repo/task_repo.go:140 ListTasks — 81.2% coverage; Scan/rows.Err paths uncovered — backfill UT-* — REQ001-004/tech-debt/repo-error-branches
- 2026-06-03 — services/agent-board/internal/repo/user_story_repo.go:97 UpdateUserStory — 57.1% coverage; same pattern — backfill UT-* — REQ001-004/tech-debt/repo-error-branches
- 2026-06-03 — services/agent-board/internal/repo/user_story_repo.go:60 UpdateUserStoryStatus — 76.2% coverage; same pattern — backfill UT-* — REQ001-004/tech-debt/repo-error-branches
- 2026-06-03 — services/agent-board/internal/repo/user_story_repo.go:117 ListUserStories — 76.5% coverage; same pattern — backfill UT-* — REQ001-004/tech-debt/repo-error-branches
- 2026-06-03 — services/agent-board/internal/repo/user_story_repo.go:35 CreateUserStory — 80.0% coverage; same pattern — backfill UT-* — REQ001-004/tech-debt/repo-error-branches
- 2026-06-03 — services/agent-board/internal/repo/user_story_repo.go:45 GetUserStory — 87.5% coverage; same pattern — backfill UT-* — REQ001-004/tech-debt/repo-error-branches
- 2026-06-03 — services/agent-board/internal/repo/audit_repo.go:31 getAuditTrail — 78.6% coverage; query/scan error paths uncovered — backfill UT-* — REQ001-004/tech-debt/repo-error-branches

**`internal/handler/*_tools.go` sub-threshold (MCP tool handlers — happy-path bias on JSON marshalling + repo-error mapping):**

- 2026-06-03 — services/agent-board/internal/handler/project_tools.go:52 handleGetProject — 58.3% coverage; missing error-mapping tests (repo error → MCP error envelope) — backfill IT-* — REQ001-004/tech-debt/handler-tools
- 2026-06-03 — services/agent-board/internal/handler/project_tools.go:76 handleUpdateProject — 68.2% coverage; same — backfill IT-* — REQ001-004/tech-debt/handler-tools
- 2026-06-03 — services/agent-board/internal/handler/project_tools.go:118 handleDeleteProject — 70.0% coverage; same — backfill IT-* — REQ001-004/tech-debt/handler-tools
- 2026-06-03 — services/agent-board/internal/handler/project_tools.go:139 handleListProjects — 71.4% coverage; same — backfill IT-* — REQ001-004/tech-debt/handler-tools
- 2026-06-03 — services/agent-board/internal/handler/project_tools.go:23 handleCreateProject — 75.0% coverage; same — backfill IT-* — REQ001-004/tech-debt/handler-tools
- 2026-06-03 — services/agent-board/internal/handler/project_tools.go:15 RegisterProjectTools — 0.0% coverage; registration path entirely untested — backfill IT-* — REQ001-004/tech-debt/handler-tools
- 2026-06-03 — services/agent-board/internal/handler/document_tools.go:37 RegisterDocumentTools — 69.2% coverage; registration partial — backfill IT-* — REQ001-004/tech-debt/handler-tools
- 2026-06-03 — services/agent-board/internal/handler/task_tools.go:39 RegisterTaskTools — 67.4% coverage; same — backfill IT-* — REQ001-004/tech-debt/handler-tools
- 2026-06-03 — services/agent-board/internal/handler/user_story_tools.go:39 RegisterUserStoryTools — 63.5% coverage; same — backfill IT-* — REQ001-004/tech-debt/handler-tools
- 2026-06-03 — services/agent-board/internal/handler/message.go:14 HandleMessage — 70.4% coverage; error-routing paths uncovered — backfill IT-* — REQ001-004/tech-debt/handler-tools
- 2026-06-03 — services/agent-board/internal/handler/message.go:74 sendError — 0.0% coverage; error helper never exercised — backfill IT-* — REQ001-004/tech-debt/handler-tools
- 2026-06-03 — services/agent-board/internal/handler/message.go:95 sendToolResultError — 0.0% coverage; error helper never exercised — backfill IT-* — REQ001-004/tech-debt/handler-tools

**`internal/mcp/server.go` ToolRegistry — entire family untested:**

- 2026-06-03 — services/agent-board/internal/mcp/server.go:75 RemoveSession — 0.0% coverage — backfill UT-* — REQ001-004/tech-debt/mcp-server-toolregistry
- 2026-06-03 — services/agent-board/internal/mcp/server.go:92 NewToolRegistry — 0.0% coverage — backfill UT-* — REQ001-004/tech-debt/mcp-server-toolregistry
- 2026-06-03 — services/agent-board/internal/mcp/server.go:99 RegisterTool — 0.0% coverage — backfill UT-* — REQ001-004/tech-debt/mcp-server-toolregistry
- 2026-06-03 — services/agent-board/internal/mcp/server.go:107 GetTool — 0.0% coverage — backfill UT-* — REQ001-004/tech-debt/mcp-server-toolregistry
- 2026-06-03 — services/agent-board/internal/mcp/server.go:116 ListTools — 0.0% coverage — backfill UT-* — REQ001-004/tech-debt/mcp-server-toolregistry
- 2026-06-03 — services/agent-board/internal/mcp/server.go:19 QueueMessage — 66.7% coverage — backfill UT-* for the uncovered branch — REQ001-004/tech-debt/mcp-server-toolregistry
- 2026-06-03 — services/agent-board/internal/mcp/server.go:29 ReceiveMessage — 66.7% coverage — same — REQ001-004/tech-debt/mcp-server-toolregistry

**Frontend (`TabSwitcher.tsx`) — pre-existing REQ004 gap, not in REQ005 scope:**

- 2026-06-03 — web/components/ProjectDetail/TabSwitcher.tsx:35-47,67 — 41.66% stmts / 33.33% branches / 39.13% lines coverage; only the default-tab render path is tested. Missing FCT-* for: tab change handler, keyboard navigation, aria-selected state on each tab, prop-driven tab override — REQ004/tech-debt/fe-tab-switcher

**E2E (all REQs) — parse-clean but unverified live:**

- 2026-06-03 — tests/e2e/REQ001..REQ004/*.robot — 21 e2e tests pass `robot --dryrun` but have never executed against a live stack (since REQ001 shipped). REQ005/US008 will provide the stack-up that makes live execution possible; first live run will likely surface seed-data / assertion drift to be fixed per-REQ — REQ001-004/tech-debt/e2e-unverified-live
- 2026-06-03 — tests/e2e/REQ005_quality_hardening_retrospective/US008_stack_smoke.robot — US008 Robot smoke deferred to first live `make e2e` run (no Docker/podman runtime in dev environment at hand-off). Verify on first `make e2e` execution; if smoke fails, file per-finding tech-debt then — REQ005/US008/e2e-smoke-deferred-live

### 2026-06-03 — First live `make e2e` against podman stack — findings

**Result:** 23 tests, 2 passed, 21 failed. US008's own smoke (E2E-US008-001/002) PASS. Pre-existing REQ001-004 + REQ005-US006 suites FAIL — confirms the "unverified live" tech-debt above. Four US008 inline fixes applied during this verification (see commit log).

- 2026-06-03 — Makefile:16 PG_CONN — hardcoded `localhost:15432` makes the host-Postgres path the runbook promises (point at local `:5432`) non-functional. Change to `?=` so env can override — REQ005/US008/makefile-host-postgres-path
- 2026-06-03 — tests/e2e/REQ001_agent_board_mcp/*.robot (5 tests) — suites assume MCP-server reachable on `:8080` and POST to `/message?sessionId=...`. Architecture §6.2 explicitly excludes mcp-server from compose. Either add mcp-server to compose with a separate port + dependent healthcheck, or rewrite REQ001 suites to drive the api-server REST surface — REQ001/tech-debt/e2e-mcp-not-in-compose
- 2026-06-03 — tests/e2e/REQ002_dashboard/US001_view_project_dashboard.robot, tests/e2e/REQ003_status_state_machine/US001-003*.robot, tests/e2e/REQ004_project_detail_page/US001-003*.robot — all 10 suites have a Suite Setup that POSTs `/message?sessionId=None` (MCP-style) for fixture creation. Same root cause as above. Convert suite setup to REST API calls against api-server OR add the mcp-server to compose — REQ002-004/tech-debt/e2e-suite-setup-mcp-coupling
- 2026-06-03 — tests/e2e/REQ005_quality_hardening_retrospective/US006_rapid_navigation.robot — `Create Project Via API` keyword POSTs `/api/v1/projects`, api-server returns HTTP 405 (Method Not Allowed). api-server doesn't expose a REST POST for projects (creation goes through MCP-tool). Either add a REST `POST /api/v1/projects` to api-server (would need new arch contract), use the host-Postgres path to seed directly, or refactor US006 keyword to use a different setup mechanism — REQ005/US006/tech-debt/e2e-no-rest-project-create

**Suggested fold-into-REQ006:** "e2e suite portability + REST surface" — adds mcp-server to compose AND adds the missing REST endpoints AND rewrites suite-setups to use REST. ~1 week of work, mostly mechanical.
