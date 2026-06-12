# US013 — Documents tab list + select · Test Report

**Generated:** 2026-05-30 (Asia/Bangkok)
**Commit at capture:** `HEAD` of `main` after all US013 tech-lead approvals (latest: `tech-lead: review pass 1 for US013_be_get_document_endpoint (approved)`)
**Story status at capture:** all 3 tasks `Status: completed`
**Capture driver:** orchestrator (Phase 3c)

---

## Backend (Go — `services/agent-board`)

Command: `go test ./...` — **107 passed, 0 failed, 0 skipped** across 6 packages.

| Spec ID | Test (Go func) | Outcome |
|---|---|---|
| UT-US013-001 (list 200 multi-doc) | `TestDocumentHandler_ListProjectDocuments_200_MultipleDocuments` | PASS |
| UT-US013-002 (list 200 empty `{"documents":[]}`) | `TestDocumentHandler_ListProjectDocuments_200_EmptyList` | PASS |
| UT-US013-003 (list 404 project missing — D-006) | `TestDocumentHandler_ListProjectDocuments_404_ProjectNotFound` | PASS |
| UT-US013-004 (list 500 on project lookup) | `TestDocumentHandler_ListProjectDocuments_500_ProjectLookupFailure` | PASS |
| UT-US013-005 (list 500 on document fetch) | `TestDocumentHandler_ListProjectDocuments_500_DocumentListFailure` | PASS |
| UT-US013-006 (list items omit `content` field) | `TestDocumentHandler_ListProjectDocuments_ContentFieldAbsent` | PASS |
| UT-US013-007 (get 200 happy + empty-content edge) | `TestDocumentHandler_GetDocument_200_HappyPath` + `TestDocumentHandler_GetDocument_200_EmptyContent` | PASS |
| UT-US013-008 (get 404) | `TestDocumentHandler_GetDocument_404_NotFound` | PASS |
| UT-US013-009 (get 500) | `TestDocumentHandler_GetDocument_500_InternalError` | PASS |
| UT-US013-010 (repo `ListDocuments` ORDER BY `updated_at DESC, id DESC`) | `TestDocumentRepo_ListDocuments_OrderByUpdatedAtDescIDDesc` | PASS |
| IT-US013-001 (list integration — missing project → 404) | `TestDocumentHandler_IT_ListProjectDocuments_MissingProject_404` | PASS |
| IT-US013-002 (list integration — empty project → 200) | `TestDocumentHandler_IT_ListProjectDocuments_EmptyProject_200` | PASS |
| IT-US013-003 (list integration — ordering A2 > A1 > B) | `TestDocumentHandler_IT_ListProjectDocuments_OrderingVerified` | PASS |
| IT-US013-004 (get integration — found) | `TestDocumentHandler_IT_GetDocument_Found` | PASS |
| IT-US013-005 (get integration — 404) | `TestDocumentHandler_IT_GetDocument_NotFound` | PASS |
| IT-US013-006 (both routes registered on echo) | `TestDocumentHandler_IT_RouteRegistration_BothDocumentRoutes` | PASS |

**Aux gates (per tech-lead reviews on both BE tasks):** `go vet ./...` clean; `golangci-lint run ./...` clean (gosec rules exercised via golangci-lint config); `gofmt -s -d .` no diff; `bash scripts/review/run-gate.sh cross` → `REVIEW GATE: PASS` (semgrep + gitleaks).

**Tooling note (reported by both BE tech-lead reviews):** `bash scripts/review/run-gate.sh be` exits 2 because the standalone `gosec` binary isn't installed on the runner. Coverage of the same ruleset is exercised inside `golangci-lint`'s gosec linter — reviewers explicitly accepted this substitution.

---

## Frontend (Jest + React Testing Library — `web`)

Command (US013 scope): `cd web && npm test -- --watchAll=false --forceExit --testPathPatterns='Document|useDocument'`
Result: **`Test Suites: 6 passed, 6 total; Tests: 40 passed, 40 total`** (2.8 s).

Full FE suite: **15 suites / 86 tests pass** (US012 + US013 combined; no regressions).

| Spec ID | Test file(s) | Outcome |
|---|---|---|
| FCT-US013-001 | `web/components/ProjectDetail/DocumentsTab.test.tsx` (renders list once `useProjectDocuments` resolves) | PASS |
| FCT-US013-002 | `web/components/ProjectDetail/DocumentSidebar.test.tsx` + `DocumentPreviewer.test.tsx` (empty-state copy — sidebar/previewer scoped via `within()`; see spec gap below) | PASS |
| FCT-US013-003 | `web/components/ProjectDetail/DocumentSidebar.test.tsx` (loading skeleton while list in-flight) | PASS |
| FCT-US013-004 | `web/components/ProjectDetail/DocumentSidebar.test.tsx` (error + Retry exercising `refetch`) | PASS |
| FCT-US013-005 | `web/pages/projects/[id].test.tsx` (URL `?doc=<id>` selects + persists) | PASS |
| FCT-US013-006 | `web/components/ProjectDetail/DocumentPreviewer.test.tsx` (empty-state when no selection) | PASS |
| FCT-US013-007 | `web/hooks/useDocument.test.ts` (race-cancellation via `AbortController` spy + stale-id guard — D-005) | PASS |
| FCT-US013-008 | `web/components/ProjectDetail/DocumentPreviewer.test.tsx` (loading state while detail in-flight) | PASS |
| FCT-US013-009 | `web/components/ProjectDetail/DocumentPreviewer.test.tsx` (200 + content rendered) | PASS |
| FCT-US013-010 | `web/components/ProjectDetail/DocumentPreviewer.test.tsx` (404 — friendly message, no Retry) | PASS |
| FCT-US013-011 | `web/components/ProjectDetail/DocumentPreviewer.test.tsx` (500 — friendly message + Retry hook) | PASS |
| FCT-US013-012 | `web/lib/api/documents.test.ts` + `client.test.ts` (typed client + AbortSignal pass-through) | PASS |
| FCT-US013-013 | `web/hooks/useProjectDocuments.test.ts` (list hook + `refetch`) | PASS |
| FCT-US013-014 | `web/hooks/useDocument.test.ts` (hook lifecycle — undefined id suppresses fetch) | PASS |
| FCT-US013-015 | `web/pages/projects/[id].test.tsx` (ghost-project deep-link → 404 cascade) | PASS |

**Aux gates (per FE tech-lead review):** `npm run typecheck` clean; `npm run lint -- --max-warnings=0` clean (`ESLint: No issues found`); `bash scripts/review/run-gate.sh cross` → `REVIEW GATE: PASS`.

**Pre-existing FE gate-script hang:** unchanged from US012 — `scripts/review/run-gate.sh fe` hangs at `npm test --watchAll=false` because MSW keeps an open handle and the gate script doesn't pass `--forceExit`. Both FE tech-lead reviews documented this and verified constituent checks individually.

**Open spec touch-up (non-blocking) — FCT-US013-002 selector ambiguity:** `/No documents yet/i` is a substring of the previewer's `/This project has no documents yet/i`, so an unscoped `findByText` would error on a multi-match. The dev correctly used `within(getByTestId('documents-sidebar-area'))` to scope; the test passes. Suggested tester touch-up: update FCT-US013-002 to mandate either `within()` scoping or a `data-testid`-narrowed query. Not blocking sign-off.

---

## E2E (Robot Framework — `tests/e2e/REQ004_project_detail_page`)

Command attempted: `robot --dryrun --include US013 tests/e2e/REQ004_project_detail_page/`
Outcome: **FAIL at suite setup — keyword arity mismatch (separate defect from the US012 import-path bug which is fixed).**

| Spec ID | Test (Robot) | Outcome |
|---|---|---|
| E2E-US013-001 (list documents → select doc → previewer renders) | `tests/e2e/REQ004_project_detail_page/US013_documents_tab.robot` | BLOCKED (suite setup) |
| E2E-US013-002 (deep-link `/projects/:id?doc=<id>`) | `tests/e2e/REQ004_project_detail_page/US013_documents_tab.robot` | BLOCKED (suite setup) |

**Root cause (raised by tester during the US012 import-path fix):** `US013_documents_tab.robot` calls `Create Document Tool` with 3 positional arguments; the keyword definition in `tests/e2e/REQ001_agent_board_mcp/mcp_keywords.resource` expects 4. Robot dry-run output:

```
Keyword 'mcp_keywords.Create Document Tool' expected 4 arguments, got 3.
```

This is a **test-spec defect** (tester-owned), not a defect in application code. Application behaviour is fully exercised by the passing BE + FE unit/integration/component tests above (10 BE handler tests, 6 BE integration tests, 1 BE route-registration test, 1 BE repo ordering test, 40 FE Jest tests targeting US013 components/hooks/API).

**Routing recommendation for po-ba sign-off:** **spec issue (tester revision)** — tester needs to either supply the missing 4th argument in the `Create Document Tool` call site(s) inside `US013_documents_tab.robot`, or align the keyword signature in `mcp_keywords.resource` to match what the REQ004 suite is actually calling. Once aligned, e2e can be re-attempted against a live stack.

---

## Skipped tests — called out

- **No BE tests skipped.**
- **No FE tests skipped.**
- **E2E (E2E-US013-001, E2E-US013-002):** not executed due to the spec keyword-arity defect above. Even after the fix, full e2e execution additionally requires standing up `cd web && npm run dev` (CSR) + `cd services/agent-board && go run ./cmd/api-server` (with a seeded DB) — typical e2e env-up that the orchestrator does not currently automate.

---

## Addendum (2026-05-30) — e2e keyword-arity fix verified

Per po-ba sign-off pass 1 routing to tester revision:

- Tester committed fix as `9f9bae0 tester: fix US013 e2e arity + clarify FCT-US013-002 selector scoping`.
- Root cause was Robot Framework parsing the leading `#` of the content arg as an inline comment, eating the 4th positional argument. Tester quoted the content via `Set Variable` with `\#` escape.
- Tester also applied the FCT-US013-002 selector clarification — implementation already had `data-testid="documents-sidebar-area"` (verified at `web/components/ProjectDetail/DocumentsTab.tsx:175`), so no FE rework needed.
- Re-validation: `robot --dryrun --include US013 tests/e2e/REQ004_project_detail_page/` → **PASS** (2/2 tests parsed cleanly, all keywords resolved with correct arity).
- Full live e2e execution still requires standing up `web` + `api-server` + a seeded DB — orchestrator does not currently automate that. Dry-run validates the parse + keyword-resolution layer, which is the failure mode po-ba flagged.
