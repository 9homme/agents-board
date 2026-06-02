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
