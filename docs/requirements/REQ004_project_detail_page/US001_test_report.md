# US001 — Navigate to project detail page with tabs · Test Report

**Generated:** 2026-05-28 (Asia/Bangkok)
**Commit at capture:** `c9f5233` (HEAD of `main` after all US001 tech-lead approvals)
**Story status at capture:** all 3 tasks `Status: completed`
**Capture driver:** orchestrator (Phase 3c)

---

## Backend (Go — `services/agent-board`)

Command: `go test -v ./internal/handler/... -run 'TestProjectHandler_GetProject|TestProjectHandler_RouteRegistration'`
Full suite: `go test ./...` — **90 passed in 6 packages** (no failures, no skips).

| Spec ID | Test (Go func) | Outcome |
|---|---|---|
| UT-US001-001 (happy 200) | `TestProjectHandler_GetProject_200` | PASS |
| UT-US001-001 (edge: empty description = `""`) | `TestProjectHandler_GetProject_EmptyDescription` | PASS |
| UT-US001-002 (404 on `repo.ErrNotFound`) | `TestProjectHandler_GetProject_404` | PASS |
| UT-US001-003 (500 on generic error) | `TestProjectHandler_GetProject_500` | PASS |
| IT-US001-001 (sqlmock round-trip — found) | `TestProjectHandler_GetProject_Integration_Found` | PASS |
| IT-US001-002 (sqlmock round-trip — not found) | `TestProjectHandler_GetProject_Integration_NotFound` | PASS |
| IT-US001-003 (route registration on echo) | `TestProjectHandler_RouteRegistration` | PASS |

**Aux gates (per tech-lead review pass 1):** `go vet ./...` clean; `golangci-lint run ./...` clean; `gofmt -s -d .` no diff.

**Substitution note:** IT-US001-001/002 use `sqlmock` rather than testcontainers. This matches the existing `GetProjects` integration pattern in the same file and was explicitly accepted by tech-lead at the same boundary in review pass 1.

---

## Frontend (Jest + React Testing Library — `web`)

Command: `cd web && npm test -- --watchAll=false --forceExit`
Result: **`Test Suites: 9 passed, 9 total; Tests: 44 passed, 44 total`** (3.3 s).

| Spec ID | Test file | Outcome |
|---|---|---|
| FCT-US001-001 | `web/components/Dashboard/ProjectCard.test.tsx` (role=link / href) | PASS |
| FCT-US001-002 | `web/components/Dashboard/ProjectCard.test.tsx` (keyboard reachable + focus-visible) | PASS |
| FCT-US001-003 | `web/components/Dashboard/ProjectCard.test.tsx` (anchor, not div+onClick) | PASS |
| FCT-US001-004 | `web/components/Dashboard/ProjectCard.test.tsx` (preserves REQ002 article styling) | PASS |
| FCT-US001-005 | `web/pages/projects/[id].test.tsx` (renders project header on 200) | PASS |
| FCT-US001-006 | `web/components/ProjectDetail/ProjectHeader.test.tsx` (header content + empty-description placeholder) | PASS |
| FCT-US001-007 | `web/pages/projects/[id].test.tsx` (loading skeleton via MSW `delay('infinite')`) | PASS |
| FCT-US001-008 | `web/components/ProjectDetail/TabSwitcher.test.tsx` + `web/pages/projects/[id].test.tsx` (tab list rendered) | PASS |
| FCT-US001-009 | `web/components/ProjectDetail/TabSwitcher.test.tsx` + `web/pages/projects/[id].test.tsx` (router.replace `shallow:true`) | PASS |
| FCT-US001-010 | `web/components/ProjectDetail/TabSwitcher.test.tsx` (keyboard ArrowLeft/Right + Enter/Space) | PASS |
| FCT-US001-011 | `web/pages/projects/[id].test.tsx` (URL `?tab=…` restored on mount) | PASS |
| FCT-US001-012 | `web/components/ProjectDetail/UserStoriesTab.test.tsx` (exact em-dash placeholder string) | PASS |
| FCT-US001-013 | `web/components/ProjectDetail/UserStoriesTab.test.tsx` (renders independently) | PASS |
| FCT-US001-014 | `web/pages/projects/[id].test.tsx` (404 hides tab list, shows back link) | PASS |
| FCT-US001-015 | `web/pages/projects/[id].test.tsx` (500 generic error state) | PASS |

**Aux gates (per tech-lead reviews):** `cd web && npm run typecheck` clean; `cd web && npm run lint -- --max-warnings=0` clean (`ESLint: No issues found`); `bash scripts/review/run-gate.sh cross` → `REVIEW GATE: PASS` (semgrep + gitleaks).

**Pre-existing tooling note:** the FE half of `scripts/review/run-gate.sh` hangs at `npm test --watchAll=false` because `web/jest.setup.ts` keeps an MSW open handle and the gate does not pass `--forceExit`. Constituent checks were verified individually by both FE-track tech-lead reviews. Tech-debt follow-up suggested in those review logs (add `--forceExit` to the gate, or `forceExit: true` to `jest.config.js`).

---

## E2E (Robot Framework — `tests/e2e/REQ004_project_detail_page`)

Command attempted: `robot --include US001 tests/e2e/REQ004_project_detail_page/`
Pre-flight: `robot --dryrun --include US001 …` → **FAIL at parse / suite setup**.

| Spec ID | Test (Robot) | Outcome |
|---|---|---|
| E2E-US001-001 (dashboard → detail navigation) | `tests/e2e/REQ004_project_detail_page/US001_navigate.robot` | BLOCKED (suite parse) |
| E2E-US001-002 (tab switching + URL persistence + refresh) | `tests/e2e/REQ004_project_detail_page/US001_navigate.robot` | BLOCKED (suite parse) |

**Root cause:** `US001_navigate.robot` line 7 imports `../../REQ001_agent_board_mcp/mcp_keywords.resource`. The actual path from `tests/e2e/REQ004_project_detail_page/` is one `..` away (`../REQ001_agent_board_mcp/mcp_keywords.resource`), not two. Robot can't find the resource, so the suite setup fails with:

```
1) No keyword with name 'Connect To MCP SSE' found.
2) No keyword with name 'Create Project Tool' found.
```

This is a **test-spec defect** (tester wrote the wrong relative path), not a defect in the application code. It is identical in `US002_documents_tab.robot` and `US003_markdown_mermaid.robot` (same import line). Routing recommendation for po-ba sign-off: **spec issue (tester revision mode)**, not application behaviour — US001 application code itself is fully exercised by passing BE + FE unit/integration/component tests above. Once tester fixes the import path, the orchestrator can re-run robot against a live `web` + `api-server` stack and append the actual E2E outcomes.

---

## Skipped tests — called out

- **No BE tests skipped.**
- **No FE tests skipped.**
- **E2E (E2E-US001-001, E2E-US001-002):** not executed due to the spec import-path defect above. Once fixed, e2e additionally requires a running `cd web && npm run dev` (CSR) + `cd services/agent-board && go run ./cmd/api-server` (with a seeded DB) — typical e2e env-up that the orchestrator currently does not stand up automatically.

---

## Addendum (2026-05-30) — e2e import-path fix verified

Per po-ba sign-off pass 1 routing to tester revision:

- Tester committed fix as `31f162d tester: fix wrong relative import path in REQ004 robot suites (US001/US002/US003)`.
- Re-validation: `robot --dryrun --include US001 tests/e2e/REQ004_project_detail_page/` → **PASS** (2/2 tests parsed cleanly, all keywords resolved).
- Full live e2e execution still requires standing up `web` + `api-server` + a seeded DB — orchestrator currently does not automate that. The dry-run validates parse + keyword resolution, which is the failure mode po-ba flagged.
