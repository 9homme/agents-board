# Tech Debt

Single backlog. `tech-lead-reviewer` appends one row to the **Open** table before `approved`. Strike through the row when resolved and move it to **Resolved**.

---

## Open

| # | Date | Location | Issue | Suggested fix | Source |
|---|---|---|---|---|---|
| 1 | 2026-06-02 | `scripts/review/test/test_run_gate.sh` UT-001/003 | Assertions only check banner presence; missing body-text and rc-value checks | Tighten per IT-US001-001/002 | REQ005/US015 |
| 2 | 2026-06-02 | `scripts/review/test/test_run_gate.sh` | Test-ID naming uses UT-NNN instead of canonical IT-US001-NNN | Re-map on next touch | REQ005/US015 |
| 3 | 2026-06-02 | `US015_be_unit_tests.md` IT-US001-004 | TTY-mode test is manual | Automate when portable TTY harness is available | REQ005/US015 |
| 4 | 2026-06-05 | TDG skill + `be-dev.md` / `fe-dev.md` | `refactor: chore:` double-prefix drift — 4 recurrences across REQ006/REQ007 | Add explicit `chore:` prefix to TDG skill; update dev agent rules | REQ006/US025 |
| 5 | 2026-06-05 | `scripts/review/run-gate.sh` | No semgrep baseline/ignore; one pre-existing finding FAILs the cross-gate on every new task | Add `.semgrepignore` or `--baseline-commit` | REQ006/US036 |
| 6 | 2026-06-07 | `internal/mcp/server.go:115` `ListTools` | Doc comment claims sorted output; map iteration is unspecified order | Sort before return, or fix the doc comment | REQ006/US033 |
| 7 | 2026-06-07 | `US035_be_unit_tests.md` UT-001 | Spec asserts `toolchain` directive; `go mod tidy` strips it — spec wording wrong | Reword UT-001 on next touch | REQ006/US035 |
| 8 | 2026-06-08 | `web/` Jest suite-wide | `--forceExit` masks an open-handle leak (likely unaborted fetch or unref'd timer) | Run `--detectOpenHandles`, fix handle, drop `--forceExit` | REQ007/US042 |
| 9 | 2026-06-08 | repo root `log.html` `output.xml` `report.html` `playwright-log.txt` | Robot artifacts tracked in git; dirties `git status` after every e2e run | `git rm --cached` the 4 files; add `.gitignore` rules | REQ007/US040 |
| 10 | 2026-06-08 | `scripts/review/run-gate.sh` BE gate | gosec + govulncheck WARN-skipped ("not installed"); security tier silently absent | Pin/install in review env or invoke via `go run` | REQ007/US040 |
| 11 | 2026-06-09 | `services/agent-board/cmd/api-server/main.go` | 47.5% func coverage; no `## Coverage exemption` filed on US042/US043 BE tasks | Add standing exemption boilerplate for `cmd/*/main.go` to the BE task template | REQ007/US042+US043 |
| 12 | 2026-06-09 | `web/package.json` | 2 npm audit advisories (1 high, 1 moderate) in `next`/`postcss` transitive chain | Schedule `next` major bump as hardening task | REQ007 |
| 13 | 2026-06-03 | REQ001–REQ004 robot specs | Specs first validated live in REQ005; similar latent issues may remain | Re-audit REQ001-004 robot specs for live-stack correctness | REQ001-004 |
| 14 | 2026-06-10 | US046 git history (handoff commit `feat(REQ008/US046): ... [in_review]`) | TDG drift — handoff/status-flip commit uses `feat(...)` prefix + `[in_review]` suffix instead of `red:/green:/refactor:` + `(US046)`; also a `refactor: chore:` double-prefix (recurrence of #4). Substantive TDD commits (red→green→refactor) all conform and are correctly ordered. | Squash handoff commits on merge, or adopt `refactor: chore: ... (US<NNN>)` handoff convention per #4 | REQ008/US046 |
| 15 | 2026-06-10 | `services/agent-board/internal/repo/requirement_repo.go:187` `isFKViolation` | FK detection uses `strings.Contains(err.Error(), "23503")` substring match instead of typed code inspection (`pq.Error.Code`/`pgconn.PgError`); brittle against driver error-string changes. Acceptable here since `pq` is not imported in this package, but fragile. | On next touch, type-assert the driver error and compare `.Code`, or centralise a shared FK-violation helper in the repo package. | REQ008/US045 |
| 16 | 2026-06-10 | US045 git history (handoff commit `feat(REQ008/US045): ... [in_review]`) | TDG drift — handoff/status-flip commit uses `feat(...)` prefix + `[in_review]` suffix instead of TDG convention (recurrence of #4/#14). The substantive red→green→refactor→refactor commits all conform and are correctly ordered. | Same fix as #14 — squash handoff commits on merge or adopt the documented handoff prefix | REQ008/US045 |
| 17 | 2026-06-10 | `web/lib/api/types.ts:55,74,89,115` (`requirementId?`) | Spec FCT-047-025/026 + arch extract call for `requirementId: string` (required); dev made it optional for backward-compat with the legacy project-scoped fetchers that don't carry it. Sound interim choice (tests green, field carried through tabs) but diverges from the contract. | When US048 removes the flat BE routes + legacy fetchers, make `requirementId: string` required on `UserStoryListItem`/`UserStory`/`DocumentListItem`/`Document`. | REQ008/US047 |
| 18 | 2026-06-10 | US047 handoff commit `refactor: chore: hand off US047 ... (US047)` (0238ed5) | `refactor: chore:` double-prefix on the non-substantive handoff commit (recurrence of #4/#14/#16). Substantive red→green→refactor×3 all conform and are correctly ordered. | Same fix as #14 — squash handoff commits on merge or adopt the documented handoff prefix | REQ008/US047 |
| 19 | 2026-06-10 | `web/hooks/useProjectRequirements.ts:84-87`, `useRequirementUserStories.ts:79-82`, `useRequirementDocuments.ts:78-81` | Branch coverage 57-64% on the three new hooks; the `else if (err instanceof Error) / else` error-classification fallback is uncovered. Stmt/line coverage ≥90% (meets gate). Identical untested pattern to pre-existing `useProjectDocuments`/`useProjectUserStories` siblings — not regressed. | Add a non-Error rejection test (or collapse the redundant `instanceof Error`/`else` arms into one) when next touching these hooks. | REQ008/US047 |
| 20 | 2026-06-11 | `services/agent-board/internal/handler/hierarchy_handler_test.go:1721` `buildRouterWithoutFlatRoutes` | Removed-route tests IT-048-024..031 assert against a local `echo.Echo` that mirrors `main.go` rather than the production route registration; if a flat route ever reappears in `main.go` only, these tests would still pass green. Removal is currently verified by the green-commit diff + Mode 2 live e2e. | Extract route registration into a shared `RegisterRoutes(e, handlers)` func in a non-`main` package and have both `main.go` and the test import it, so the test exercises the real registration. | REQ008/US048 |

---

## Resolved

| # | Date | Location | Issue | Resolved in |
|---|---|---|---|---|
| R01 | 2026-06-02 | `scripts/review/run-gate.sh:117` | FE gate in subshell — `fail()` never propagates; lint/test failures silently dropped | REQ007/US042 commit f380ce0 |
| R02 | 2026-06-03 | `internal/repo/task_repo.go` (5 functions) | 62–87% coverage; error-branch tests missing | REQ006/US025 (all ≥95%) |
| R03 | 2026-06-03 | `internal/repo/user_story_repo.go` (5 functions) | 57–87% coverage; same pattern | REQ006/US026 (all ≥95%) |
| R04 | 2026-06-03 | `internal/repo/audit_repo.go:31` | 78.6% coverage; query/scan error paths uncovered | REQ006/US027 (100%) |
| R05 | 2026-06-03 | `internal/handler/project_tools.go` (6 functions) | 0–75% coverage; error-mapping tests missing | REQ006/US028 (avg 99.25%) |
| R06 | 2026-06-03 | `internal/handler/document_tools.go` | 69.2% coverage | REQ006/US029 (100%) |
| R07 | 2026-06-03 | `internal/handler/task_tools.go` | 67.4% coverage | REQ006/US030 (95.3%) |
| R08 | 2026-06-03 | `internal/handler/user_story_tools.go` | 63.5% coverage | REQ006/US031 (97.7%) |
| R09 | 2026-06-03 | `internal/handler/message.go` (3 functions) | 0–70.4% coverage; error helpers untested | REQ006/US032 (96.3%) |
| R10 | 2026-06-03 | `internal/mcp/server.go` (7 functions) | 0–66.7% coverage; ToolRegistry entirely untested | REQ006/US033 (100%) |
| R11 | 2026-06-03 | `web/components/ProjectDetail/TabSwitcher.tsx` | 41.66% stmt coverage; only default-tab render tested | REQ006/US036 |
| R12 | 2026-06-03 | `cmd/mcp-server/main.go` vs `cmd/api-server/main.go` | `DB_URL` vs `DATABASE_URL` inconsistency | REQ006/US034 (standardised on DATABASE_URL) |
| R13 | 2026-06-03 | tests/e2e REQ001-004 suites | 21 tests passing dryrun but never run live | REQ005/US022 (23/23 green ×3 runs) |
| R14 | 2026-06-03 | tests/e2e REQ001-004 suite setups | MCP `/message?sessionId=None` calls; mcp-server not in compose | REQ005/US022 (mcp-server added to compose on :8081) |
| R15 | 2026-06-03 | `Makefile:16 PG_CONN` | Hardcoded `localhost:15432`; env override impossible | REQ006/US038 (PG_CONN now `?=`) |
| R16 | 2026-06-03 | `services/agent-board/Dockerfile` | No `USER` directive; runs as root; cross-gate FAIL | REQ006 (USER nonroot:nonroot added) |
| R17 | 2026-06-03 | `web/Dockerfile` | Same missing `USER` | REQ006 (USER node added) |
| R18 | 2026-06-03 | `cmd/mcp-server/sse.go` | govulncheck finding on stdlib crypto/x509 | REQ006/US035 (Go toolchain bump) |
| R19 | 2026-06-03 | `Makefile e2e-up` | api-server health-check polls `/api/v1/projects` before tables exist; SSE curl hangs without `--max-time` | REQ007/US039+US040 |
| R20 | 2026-06-03 | REST surface | No REST writes; all writes via MCP — perceived gap | REQ006/US037 ADR-001 (won't-fix — MCP-only-writes permanent) |
| R21 | 2026-06-05 | `tests/e2e/REQ005_*/US020_rapid_navigation.robot` | `Create Project Via API` → 405 (api-server has no REST POST for projects) | REQ005/US022 (rewritten to use MCP tools) |
