# Tech Debt

Single backlog. `tech-lead-reviewer` appends one row to the **Open** table before `approved`. Strike through the row when resolved and move it to **Resolved**.

---

## Open

| # | Date | Location | Issue | Suggested fix | Source |
|---|---|---|---|---|---|
| 1 | 2026-06-02 | `scripts/review/test/test_run_gate.sh` UT-001/003 | Assertions only check banner presence; missing body-text and rc-value checks | Tighten per IT-US001-001/002 | REQ005/US001 |
| 2 | 2026-06-02 | `scripts/review/test/test_run_gate.sh` | Test-ID naming uses UT-NNN instead of canonical IT-US001-NNN | Re-map on next touch | REQ005/US001 |
| 3 | 2026-06-02 | `US001_be_unit_tests.md` IT-US001-004 | TTY-mode test is manual | Automate when portable TTY harness is available | REQ005/US001 |
| 4 | 2026-06-05 | TDG skill + `be-dev.md` / `fe-dev.md` | `refactor: chore:` double-prefix drift — 4 recurrences across REQ006/REQ007 | Add explicit `chore:` prefix to TDG skill; update dev agent rules | REQ006/US001 |
| 5 | 2026-06-05 | `scripts/review/run-gate.sh` | No semgrep baseline/ignore; one pre-existing finding FAILs the cross-gate on every new task | Add `.semgrepignore` or `--baseline-commit` | REQ006/US013 |
| 6 | 2026-06-07 | `internal/mcp/server.go:115` `ListTools` | Doc comment claims sorted output; map iteration is unspecified order | Sort before return, or fix the doc comment | REQ006/US009 |
| 7 | 2026-06-07 | `US012_be_unit_tests.md` UT-001 | Spec asserts `toolchain` directive; `go mod tidy` strips it — spec wording wrong | Reword UT-001 on next touch | REQ006/US012 |
| 8 | 2026-06-08 | `web/` Jest suite-wide | `--forceExit` masks an open-handle leak (likely unaborted fetch or unref'd timer) | Run `--detectOpenHandles`, fix handle, drop `--forceExit` | REQ007/US004 |
| 9 | 2026-06-08 | repo root `log.html` `output.xml` `report.html` `playwright-log.txt` | Robot artifacts tracked in git; dirties `git status` after every e2e run | `git rm --cached` the 4 files; add `.gitignore` rules | REQ007/US002 |
| 10 | 2026-06-08 | `scripts/review/run-gate.sh` BE gate | gosec + govulncheck WARN-skipped ("not installed"); security tier silently absent | Pin/install in review env or invoke via `go run` | REQ007/US002 |
| 11 | 2026-06-09 | `services/agent-board/cmd/api-server/main.go` | 47.5% func coverage; no `## Coverage exemption` filed on US004/US005 BE tasks | Add standing exemption boilerplate for `cmd/*/main.go` to the BE task template | REQ007/US004+US005 |
| 12 | 2026-06-09 | `web/package.json` | 2 npm audit advisories (1 high, 1 moderate) in `next`/`postcss` transitive chain | Schedule `next` major bump as hardening task | REQ007 |
| 13 | 2026-06-03 | REQ001–REQ004 robot specs | Specs first validated live in REQ005; similar latent issues may remain | Re-audit REQ001-004 robot specs for live-stack correctness | REQ001-004 |

---

## Resolved

| # | Date | Location | Issue | Resolved in |
|---|---|---|---|---|
| R01 | 2026-06-02 | `scripts/review/run-gate.sh:117` | FE gate in subshell — `fail()` never propagates; lint/test failures silently dropped | REQ007/US004 commit f380ce0 |
| R02 | 2026-06-03 | `internal/repo/task_repo.go` (5 functions) | 62–87% coverage; error-branch tests missing | REQ006/US001 (all ≥95%) |
| R03 | 2026-06-03 | `internal/repo/user_story_repo.go` (5 functions) | 57–87% coverage; same pattern | REQ006/US002 (all ≥95%) |
| R04 | 2026-06-03 | `internal/repo/audit_repo.go:31` | 78.6% coverage; query/scan error paths uncovered | REQ006/US003 (100%) |
| R05 | 2026-06-03 | `internal/handler/project_tools.go` (6 functions) | 0–75% coverage; error-mapping tests missing | REQ006/US004 (avg 99.25%) |
| R06 | 2026-06-03 | `internal/handler/document_tools.go` | 69.2% coverage | REQ006/US005 (100%) |
| R07 | 2026-06-03 | `internal/handler/task_tools.go` | 67.4% coverage | REQ006/US006 (95.3%) |
| R08 | 2026-06-03 | `internal/handler/user_story_tools.go` | 63.5% coverage | REQ006/US007 (97.7%) |
| R09 | 2026-06-03 | `internal/handler/message.go` (3 functions) | 0–70.4% coverage; error helpers untested | REQ006/US008 (96.3%) |
| R10 | 2026-06-03 | `internal/mcp/server.go` (7 functions) | 0–66.7% coverage; ToolRegistry entirely untested | REQ006/US009 (100%) |
| R11 | 2026-06-03 | `web/components/ProjectDetail/TabSwitcher.tsx` | 41.66% stmt coverage; only default-tab render tested | REQ006/US013 |
| R12 | 2026-06-03 | `cmd/mcp-server/main.go` vs `cmd/api-server/main.go` | `DB_URL` vs `DATABASE_URL` inconsistency | REQ006/US010 (standardised on DATABASE_URL) |
| R13 | 2026-06-03 | tests/e2e REQ001-004 suites | 21 tests passing dryrun but never run live | REQ005/US008 (23/23 green ×3 runs) |
| R14 | 2026-06-03 | tests/e2e REQ001-004 suite setups | MCP `/message?sessionId=None` calls; mcp-server not in compose | REQ005/US008 (mcp-server added to compose on :8081) |
| R15 | 2026-06-03 | `Makefile:16 PG_CONN` | Hardcoded `localhost:15432`; env override impossible | REQ006/US015 (PG_CONN now `?=`) |
| R16 | 2026-06-03 | `services/agent-board/Dockerfile` | No `USER` directive; runs as root; cross-gate FAIL | REQ006 (USER nonroot:nonroot added) |
| R17 | 2026-06-03 | `web/Dockerfile` | Same missing `USER` | REQ006 (USER node added) |
| R18 | 2026-06-03 | `cmd/mcp-server/sse.go` | govulncheck finding on stdlib crypto/x509 | REQ006/US012 (Go toolchain bump) |
| R19 | 2026-06-03 | `Makefile e2e-up` | api-server health-check polls `/api/v1/projects` before tables exist; SSE curl hangs without `--max-time` | REQ007/US001+US002 |
| R20 | 2026-06-03 | REST surface | No REST writes; all writes via MCP — perceived gap | REQ006/US014 ADR-001 (won't-fix — MCP-only-writes permanent) |
| R21 | 2026-06-05 | `tests/e2e/REQ005_*/US006_rapid_navigation.robot` | `Create Project Via API` → 405 (api-server has no REST POST for projects) | REQ005/US008 (rewritten to use MCP tools) |
