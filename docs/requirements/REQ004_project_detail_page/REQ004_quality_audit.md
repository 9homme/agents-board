# REQ004 Quality + Coverage Audit

**Audit date:** 2026-05-30
**HEAD commit:** `9a77ed0`
**Audit branch:** `worktree-agent-ab9da34ff39564aa2`
**Audited range:** `a5aa9ce..HEAD` (all REQ004 Phase 3 commits, including tech-lead, tester, and po-ba revisions)
**Auditor:** tech-lead (cross-cutting review mode, NOT per-task)
**Scope:** boundary-respected — no source / config / spec / task / gate-script changes; this report is the only file written.

---

## Executive summary (5 bullets)

1. **BE handler coverage is excellent — the user's #1 worry is overblown for the lines REQ004 actually touched.** The two new handlers (`GetProject`, `ListProjectDocuments`, `GetDocument`) are at **100% function coverage**, with every branch (200 / 404 / 500, both project-lookup and document-lookup failure paths) exercised by `httptest`. 107 / 107 Go tests pass; `internal/handler` package = 80.3% lines, `internal/repo` = 81.5%.
2. **There IS genuine but pre-existing repo-layer coverage thinness — and REQ004 inherited rather than added it.** `document_repo.go` `UpdateDocument` (62.5%), `ListDocuments` (80.0%), and the symmetric branches in `project_repo.go` have uncovered "generic DB error wrap" and `rows.Scan`/`rows.Err()` paths. The single line REQ004 changed in `document_repo.go` (the ORDER BY) IS covered by `TestDocumentRepo_ListDocuments_OrderByUpdatedAtDescIDDesc`. Recommendation: catch up this debt in a focused fix story, not a REQ004 hot-patch.
3. **Code quality across the diff is clean.** `go vet` clean. `gofmt -s -d .` no diff. `golangci-lint run ./...` (which includes `gosec`, `staticcheck`, `errcheck`, `bodyclose`, `rowserrcheck`, `sqlclosecheck`, `errorlint`, `revive`) reports zero issues. Cross-cutting gate (`semgrep p/owasp-top-ten + p/golang + p/typescript + p/react`, `gitleaks`) `REVIEW GATE: PASS`. FE `npm run typecheck` / `npm run lint --max-warnings=0` / 107 Jest tests all pass.
4. **One real script bug + two environmental tool gaps in the review gate.** `scripts/review/run-gate.sh:58` and `:71` both call `printf "${YELLOW}--- output ..."` — when stdout is not a TTY the colour vars become empty, the format string starts with `--`, and `printf` aborts with `printf: --: invalid option` so the failing-check output is hidden. The standalone `gosec` and `govulncheck` binaries are missing on this machine (gosec ruleset is exercised inside golangci-lint instead — accepted by prior per-task reviews). The FE gate's `npm test --watchAll=false` step did **not** hang in this audit (rc=0 in ~5 s) but Jest does emit `A worker process has failed to exit gracefully` — the leak is real, just not currently blocking. The fix is one line: pass `--forceExit` in the gate.
5. **Verdict: ship-ready with a short tech-debt list, not "thin coverage".** No must-fix-before-next-REQ defects in the production code. The three must-fix items are (a) the `printf "--"` bug in the gate script, (b) add `--forceExit` to the FE gate's `npm test` line, and (c) make `gosec` / `govulncheck` either installed-or-skipped-with-warning so the gate can complete end-to-end in this dev environment without `exit 2`. Eventually-fix list focuses on harmonising the three hooks' lifecycle patterns and backfilling repo error-branch tests.

---

## 1. Backend test coverage

Test summary: **107 passed, 0 failed, 0 skipped, 6 packages** (`agent-board/cmd/api-server`, `agent-board/cmd/mcp-server`, `agent-board/internal/domain`, `agent-board/internal/handler`, `agent-board/internal/mcp`, `agent-board/internal/repo`).

### 1.1 Per-function coverage — REQ004-touched files

| File | Function | Coverage | Status |
|---|---|---|---|
| `internal/handler/project_handler.go` | `NewProjectHandler` | 100.0% | OK |
| `internal/handler/project_handler.go` | `GetProjects` | 100.0% | OK (touched incidentally) |
| `internal/handler/project_handler.go` | `GetProject` (US012) | **100.0%** | OK |
| `internal/handler/document_handler.go` | `NewDocumentHandler` | 100.0% | OK |
| `internal/handler/document_handler.go` | `ListProjectDocuments` (US013) | **100.0%** | OK |
| `internal/handler/document_handler.go` | `GetDocument` (US013) | **100.0%** | OK |
| `internal/repo/project_repo.go` | `NewProjectRepo` | 100.0% | OK |
| `internal/repo/project_repo.go` | `CreateProject` | 87.5% | Pre-existing gap |
| `internal/repo/project_repo.go` | `GetProject` | 87.5% | Pre-existing gap (gen-err branch) |
| `internal/repo/project_repo.go` | `UpdateProject` | 62.5% | Pre-existing gap |
| `internal/repo/project_repo.go` | `DeleteProject` | 80.0% | Pre-existing gap |
| `internal/repo/project_repo.go` | `ListProjects` | 80.0% | Pre-existing gap |
| `internal/repo/document_repo.go` | `NewDocumentRepo` | 100.0% | OK |
| `internal/repo/document_repo.go` | `CreateDocument` | 88.9% | Pre-existing gap |
| `internal/repo/document_repo.go` | `GetDocument` | 87.5% | Pre-existing gap (gen-err branch) |
| `internal/repo/document_repo.go` | `UpdateDocument` | 62.5% | Pre-existing gap |
| `internal/repo/document_repo.go` | `DeleteDocument` | 80.0% | Pre-existing gap |
| `internal/repo/document_repo.go` | `ListDocuments` (US013 modified) | **80.0%** | Pre-existing gap; the ORDER-BY line REQ004 added IS covered |

### 1.2 Per-package totals

- `agent-board/internal/handler`: **80.3%** statement coverage.
- `agent-board/internal/repo`: **81.5%** statement coverage.
- `total` (all 6 packages): 65.5% (dragged down by `cmd/api-server`, `cmd/mcp-server`, `internal/mcp` server registry — none of which REQ004 touched).

### 1.3 Branch-level audit of REQ004 happy + error paths

For each REQ004 endpoint, I cross-referenced uncovered statement ranges from `/tmp/req004_cover.out` against the source. **All handler-side happy paths AND all handler-side error paths are covered.** Specifically:

**`GetProject` (project_handler.go:64-90):**
- 200 happy path: `TestProjectHandler_GetProject_200`, `TestProjectHandler_GetProject_EmptyDescription`, `TestProjectHandler_GetProject_Integration_Found` ✓
- 404 not-found branch: `TestProjectHandler_GetProject_404`, `TestProjectHandler_GetProject_Integration_NotFound` ✓
- 500 error branch: `TestProjectHandler_GetProject_500` ✓
- Field-shape lock: `assert.Len(t, res, 5)` + key-absence on `"project"` wrapper ✓

**`ListProjectDocuments` (document_handler.go:41-86):**
- 200 multi-doc: `TestDocumentHandler_ListProjectDocuments_200_MultipleDocuments` ✓
- 200 empty list (`{"documents":[]}` not `null`): `TestDocumentHandler_ListProjectDocuments_200_EmptyList`, integration variant ✓
- 404 project-not-found (D-006): `TestDocumentHandler_ListProjectDocuments_404_ProjectNotFound`, integration variant ✓
- 500 on project-lookup failure (non-ErrNotFound from `projectRepo.GetProject`): `TestDocumentHandler_ListProjectDocuments_500_ProjectLookupFailure` ✓
- 500 on document-list failure: `TestDocumentHandler_ListProjectDocuments_500_DocumentListFailure` ✓
- `content` field absent in list items: `TestDocumentHandler_ListProjectDocuments_ContentFieldAbsent` ✓
- ListDocuments NOT called when project 404 (D-006): asserted via `listCallCount == 0` in the 404 test ✓

**`GetDocument` (document_handler.go:91-111):**
- 200 with content + 6-field shape: `TestDocumentHandler_GetDocument_200_HappyPath`, `TestDocumentHandler_IT_GetDocument_Found` ✓
- 200 with empty content serialised as `""` not `null`: `TestDocumentHandler_GetDocument_200_EmptyContent` ✓
- 404: `TestDocumentHandler_GetDocument_404_NotFound`, `TestDocumentHandler_IT_GetDocument_NotFound` ✓
- 500: `TestDocumentHandler_GetDocument_500_InternalError` ✓
- Route registration smoke test: `TestDocumentHandler_IT_RouteRegistration_BothDocumentRoutes` ✓

**`ListDocuments` repo change (document_repo.go:103, `ORDER BY updated_at DESC, id DESC`):**
- The exact ORDER BY clause regex is asserted in `TestDocumentRepo_ListDocuments` (updated) and `TestDocumentRepo_ListDocuments_OrderByUpdatedAtDescIDDesc` (US-US013-010) ✓
- Three-row tiebreaker scenario (same `updated_at`, lexicographic id DESC) verified ✓

### 1.4 Is the user's worry justified?

**Largely no, with a fair caveat.** The handlers REQ004 added are at 100% — every happy and error JSON-envelope branch is exercised against the contract. The integration tests (`IT-US012-*`, `IT-US013-*`) also pin the wire format via `sqlmock` plus `httptest`. That is exactly the BE coverage discipline this project asked for, and the user's "BE unit-test coverage too low" hypothesis does not apply to the new code.

**The legitimate caveat: the repo layer has pre-existing thin spots.** They are not REQ004 regressions — `git diff a5aa9ce..HEAD -- services/agent-board/internal/repo` shows REQ004 changed exactly **two lines of production code**: the `ORDER BY` clause in `ListDocuments` and a test variable rename. So flagging this as "REQ004 broke coverage" is not accurate. The uncovered branches in `project_repo.go` / `document_repo.go` are:

- `Create*` — `db.QueryRowContext(...).Scan(...) → err != nil` wrap branch (3-line "generic DB error" path; pre-existing).
- `Get*` — non-`ErrNoRows` error branch (line `return nil, fmt.Errorf("failed to get …: %w", err)`; pre-existing).
- `Update*` — both `ErrNoRows` mapping and generic-error branches (pre-existing — UpdateDocument is not even called by any REQ004 endpoint).
- `Delete*` — `ExecContext` error branch (pre-existing).
- `List*` — `QueryContext` initial error, `rows.Scan` error inside loop, `rows.Err()` after loop (pre-existing — ListProjects is on the dashboard path, ListDocuments is the one US013 modified).

If the team wants to close this for real (not REQ004-specific), see §4.3 for the exact test-function shopping list.

---

## 2. Code quality

### 2.1 Backend (`services/agent-board`)

| Tool | Result | Notes |
|---|---|---|
| `gofmt -s -d .` | clean (no diff) | All Go files canonical |
| `go vet ./...` | clean | No issues |
| `golangci-lint run ./...` | clean (No issues found) | v2 schema, `bodyclose`, `errcheck`, `errorlint`, `gocritic`, `gosec`, `govet`, `ineffassign`, `misspell`, `noctx`, `revive`, `rowserrcheck`, `sqlclosecheck`, `staticcheck`, `unused` |
| `staticcheck ./...` | not installed standalone | Folded into `golangci-lint` (v2 schema) — coverage equivalent |
| `gosec` (standalone) | not installed | Coverage equivalent via golangci-lint's `gosec` linter — same ruleset |
| `govulncheck` | not installed | Tool gap — see §3 |

**Manual sweep of REQ004 BE files for issues the linters might miss:**

- **Context propagation:** all repo methods take `ctx` and forward it to `QueryContext`/`QueryRowContext`/`ExecContext` ✓.
- **DB resource lifecycle:** `ListDocuments` (line 108) uses `defer func() { _ = rows.Close() }()` ✓; the explicit `_ =` is intentional to satisfy `errcheck`. Same pattern in `ListProjects` (line 108).
- **Cyclomatic complexity outliers:** none — handlers max out at 4 branches (200 / 404 / 500 / nil-project guard).
- **Error wrapping:** consistent `fmt.Errorf("failed to %s: %w", op, err)` style; `errors.Is(err, sql.ErrNoRows)` and `errors.Is(err, repo.ErrNotFound)` are the only sentinel checks — correct.
- **Doc comments on public exports:** every new exported symbol (`NewDocumentHandler`, `NewProjectHandler`, `GetProject`, `ListProjectDocuments`, `GetDocument`) has a doc comment that names the architecture decision it implements (D-006, D-002, etc.) ✓.
- **Logging:** `log.Printf("Failed to …: %v", err)` is used at the 500 sites — minimum-viable; not structured logging, but no spam either. Acceptable for this team's baseline; not a blocker.
- **No leaked DB connections, no missing `rows.Err()` check, no missing `defer rows.Close()`** — `rowserrcheck` and `sqlclosecheck` would have caught those and report clean.

### 2.2 Frontend (`web/`)

| Tool | Result |
|---|---|
| `npm run typecheck` (`tsc --noEmit`) | clean |
| `npm run lint -- --max-warnings=0` | `No ESLint warnings or errors` |
| `npm test -- --watchAll=false --forceExit` | 17 suites, 107 tests, 0 failures |
| `npm audit --omit=dev --audit-level=high` | warns but does not block (gate `run_check_warn`) |

**Manual sweep of REQ004 FE files:**

- **CSR-only enforcement:** `grep -rEn 'getServerSideProps|getStaticProps|getInitialProps' web/components web/hooks web/lib web/pages` returns only doc-comment mentions (the literal absence is asserted in `pages/projects/[id].tsx:20`). No `web/pages/api/` directory. PASS.
- **`fetch()` boundary:** no raw `fetch()` calls outside `web/lib/api/` (gate's regex run repeats the check — clean).
- **`dangerouslySetInnerHTML`:** exactly one usage in `MermaidDiagram.tsx:139`, with a comment naming D-004 as the sanctioned exception. The fed string is the output of `mermaid.render()` with `securityLevel:'strict'`, not raw user input. Acceptable.
- **API state lives in `web/lib/api/`:** confirmed — `ApiError`, `fetchClient`, `fetchProject`, `fetchProjectDocuments`, `fetchDocument` are all imported from `lib/api/*` by hooks; no parallel state machine.

**Inconsistency worth flagging (NOT a defect — design smell):**

- `useDocument.ts` uses the **AbortController + latestIdRef** pattern (correct race-safe pattern per D-005).
- `useProject.ts` and `useProjectDocuments.ts` use only the **`mounted` flag** pattern — they do NOT abort the in-flight network request when the dep changes or the component unmounts. Currently safe because (a) the project id never changes within a page mount, (b) `useProjectDocuments` only changes when the route changes, which unmounts the consumer. But the three hooks now use two different async-safety patterns, and `fetchProject` doesn't even accept a `signal` (vs. `fetchDocument`/`fetchProjectDocuments` which do). If a future task adds a "switch projects without route change" flow, `useProject` will race. **Tech-debt, not REQ004-blocking.**

**One eslint-disable in production code:**

- `web/components/ProjectDetail/DocumentsTab.tsx:78` — `// eslint-disable-line react-hooks/exhaustive-deps` on the auto-select effect (deps: `documents, docParam` — intentionally omits `router`). This is justified (omitting `router` is the standard idiom to avoid re-running on every router-internals change), but a brief inline comment explaining why would be cleaner. **Tech-debt, not blocking.**

**One unguarded `console.error`:**

- `web/components/ProjectDetail/MarkdownErrorBoundary.tsx:38` — `console.error('[MarkdownErrorBoundary] caught render error:', error, info);`. The comment says "in production a real logger would go here". Acceptable for now; flag for the eventual frontend logging story.

### 2.3 Cross-cutting

```
== Cross-cutting · repo ==
  PASS  semgrep (owasp/golang/typescript)
  PASS  gitleaks (no secrets)

REVIEW GATE: PASS
```

Independent targeted scan: `semgrep --config=auto services/agent-board/internal services/agent-board/cmd web/components web/hooks web/lib web/pages` → **0 findings, 0 blocking, 293 rules across 67 files**. No regressions.

---

## 3. Quality-gate audit

### 3.1 Per-gate-command actual behaviour

| Gate invocation | Runs end-to-end? | If no, why |
|---|---|---|
| `scripts/review/run-gate.sh cross` | **YES** — `REVIEW GATE: PASS` in ~7 s | — |
| `scripts/review/run-gate.sh be services/agent-board` | **NO** — `exit 2 MISSING TOOL: gosec` immediately | Tool presence: standalone `gosec` not installed; same ruleset is in `.golangci.yml` so the script's expectation is redundant. Would also fail on `govulncheck` (also missing). |
| `scripts/review/run-gate.sh fe` | **YES** in this audit run — completes in ~30 s with `REVIEW GATE: PASS` | Did NOT hang here; however Jest prints `A worker process has failed to exit gracefully and has been force exited` — the user's earlier reports of a hang are plausible and have been worked around in per-task reviews by adding `--forceExit`. Without `--forceExit`, the run is fragile (works today, may stick another day depending on MSW handle GC). |

### 3.2 Defects vs. environmental gaps

- **Script defect (must fix):** `scripts/review/run-gate.sh` lines 58 and 71 both use `printf "${YELLOW}--- output (rc=%d) ---${RESET}\n%s\n…" "$rc" "$out"`. When stdout is non-TTY (CI, pipe, `bash -c`), the colour vars are empty, the format string becomes `--- output (rc=%d) ---\n%s\n…`, and `printf` parses the leading `--` as an option terminator, prints `printf: --: invalid option`, and **drops the failing-check output**. Minimum patch: `printf -- "${YELLOW}--- output ...${RESET}\n..." ...` (add `--` before the format), or prepend a space inside the format. The same bug pattern should be looked for in any other `printf` whose format starts with a colour-variable (none other I could find).
- **Script defect (must fix):** FE gate runs `npm test --silent -- --watchAll=false` without `--forceExit`. The MSW server's open handle has caused intermittent hangs (documented in the per-task review logs for US012 / US013 / US014 FE tasks). Minimum patch: add `--forceExit` to that one line. Long-term, fix the actual leak (see §4.3 item 5).
- **Tool gap (must address, not strictly a bug):** the script `require_tool gosec` will `exit 2` even though the repo's `.golangci.yml` enables `gosec` as a linter. The substitution has been accepted by every per-task BE review in REQ004 (US012-be, US013-list-be, US013-get-be). The script should EITHER (a) require `gosec` and treat the standalone run as additive, OR (b) skip the standalone `gosec` step with a `WARN` line when golangci-lint already has gosec enabled. Same call applies to `govulncheck`.

### 3.3 Minimum patch to make every gate terminate with a clean PASS/FAIL line

Three small edits to `scripts/review/run-gate.sh`:

1. **Fix the `printf` `--` bug** (lines 58 and 71): prefix the format with `--` or insert a literal space. Example: `printf -- "${YELLOW}--- output (rc=%d) ---${RESET}\n%s\n${YELLOW}-----------------------${RESET}\n" "$rc" "$out"`. (Doesn't change PASS behaviour; **does** restore visibility of failure output.)
2. **Add `--forceExit` to the FE test command** (line 105 area): `bash -c 'npm test --silent -- --watchAll=false --forceExit'`. Stops the intermittent hang.
3. **Make `gosec` and `govulncheck` soft-required**: replace `require_tool gosec ...` / `require_tool govulncheck ...` with a `command -v` check + `run_check_warn` when missing, or auto-skip those two `run_check` calls if the tool isn't on PATH and emit a `WARN  gosec (skipped — not installed; coverage via golangci-lint)` line. The gate then prints `REVIEW GATE: PASS` (or FAIL on real issues) instead of `exit 2 MISSING TOOL`.

After those three edits, all three gate invocations on this machine should terminate with `REVIEW GATE: PASS/FAIL` and a non-2 exit code.

---

## 4. Verdict + ranked recommendations

### 4.1 Verdict

**REQ004 is ship-ready. The user's "BE coverage too low" concern is not supported by the data for code REQ004 actually added.** Handler coverage on the three new endpoints is 100%. The repo-layer gaps are pre-existing tech debt inherited from REQ001-003, and REQ004 left them as-is (correct — that's tester's job to spec, not a tech-lead-imposed scope creep). Code quality, security scans, and the cross-cutting gate are all clean. The story-level reviews and po-ba sign-offs are consistent with this audit.

### 4.2 Top 3 BEFORE next REQ ships

1. **Fix `scripts/review/run-gate.sh` lines 58 + 71 `printf "--"` bug.** Currently silently swallows failure-output in CI / non-TTY runs. One-line fix per call site; high signal-to-noise.
2. **Add `--forceExit` to the FE gate's `npm test` invocation.** The MSW leak has caused real hangs in two prior tech-lead reviews on this REQ. Until the actual leak is fixed (4.3 item 5), pass `--forceExit` so the gate completes deterministically. One-token edit.
3. **Make `gosec` and `govulncheck` soft-required in the gate (skip-with-warning when missing) OR document them as install-required in `scripts/review/README.md` AND wire `make install-tools` to grab them.** Pick one. Today, every BE per-task review on this machine has to manually document "the standalone gosec was missing, accepting golangci-lint's gosec linter as substitute" — that's a smell. Either install the tools or stop pretending the gate requires them.

### 4.3 Top 3 EVENTUALLY (tech debt)

1. **Backfill the repo error-branch tests** so `internal/repo` exceeds 90% per-file coverage. See §4.4 for the exact function-and-scenario shopping list. Pre-existing, but worth closing while the team is fresh on the file shapes.
2. **Harmonise the three React hooks (`useProject`, `useProjectDocuments`, `useDocument`) on a single async-safety pattern.** Today, `useDocument` uses AbortController + `latestIdRef`; the other two use a `mounted` flag and a `fetchProject` that doesn't even accept `signal`. Either everyone gets AbortController + signal forwarding, or extract a shared `useFetch<T>(key, fetcher)` hook. Reduces the chance that a future "switch projects without route change" feature reintroduces a race.
3. **Address the actual MSW handle leak that triggers `worker process has failed to exit gracefully`.** Candidate fixes: (a) ensure `afterAll(() => server.close())` is being reached on every test file (it is, per `jest.setup.ts`), (b) check whether react-markdown / mermaid's lazy-load schedules any pending timer at unmount, (c) add `jest.useFakeTimers()` in the few component tests that use mermaid. Optional eventual win: replace the FE gate's `--forceExit` with `--detectOpenHandles` in a focused-fix story to actually identify the culprit.

### 4.4 BE coverage shopping list (if anyone wants to close the §4.3 item 1 debt)

Each entry below is **one additional `sqlmock`-driven test function** that closes a specific uncovered branch. None are blocking REQ004; collectively they would push `internal/repo` from 81.5% to ~95%.

**`internal/repo/document_repo_test.go`:**
- `TestDocumentRepo_CreateDocument_GenericError` — `mock.ExpectQuery(...).WillReturnError(errors.New("db down"))`; assert `errors.Is(err, ErrNotFound) == false` and the returned error wraps `"failed to create document"`. Closes line 42-44.
- `TestDocumentRepo_GetDocument_GenericError` — same shape; closes line 65.
- `TestDocumentRepo_UpdateDocument_NotFound` and `_GenericError` — closes lines 83-87. (Not on REQ004's hot path but the function exists.)
- `TestDocumentRepo_DeleteDocument_GenericError` — `mock.ExpectExec(...).WillReturnError(...)`. Closes line 96-98.
- `TestDocumentRepo_ListDocuments_QueryError` — `mock.ExpectQuery(...).WillReturnError(...)`. Closes line 105-107.
- `TestDocumentRepo_ListDocuments_ScanError` — return a row with a column-type mismatch (`sqlmock` `AddRow("not-a-time", ...)`). Closes line 113-115.
- `TestDocumentRepo_ListDocuments_RowsErr` — use `sqlmock`'s `RowError(rowIdx, err)`. Closes line 119-121.

**`internal/repo/project_repo_test.go`:** the symmetric six (`Create_GenericError`, `Get_GenericError`, `Update_NotFound`, `Update_GenericError`, `Delete_GenericError`, `ListProjects_QueryError`, `ListProjects_ScanError`, `ListProjects_RowsErr`).

**`internal/handler/document_handler.go`:** already 100% — no additions needed.
**`internal/handler/project_handler.go`:** already 100% — no additions needed.

---

## Appendix A — Raw evidence

### Go test summary
```
107 passed in 6 packages (cmd/api-server, cmd/mcp-server, internal/domain,
internal/handler, internal/mcp, internal/repo)
total: (statements) 65.5%
internal/handler: 80.3%
internal/repo:    81.5%
```

### Cross-cutting gate
```
== Cross-cutting · repo ==
  PASS  semgrep (owasp/golang/typescript)
  PASS  gitleaks (no secrets)
REVIEW GATE: PASS
```

### Semgrep targeted scan
```
Ran 293 rules on 67 files: 0 findings.
```

### FE gate (this run)
```
== FE gate · web/ ==
  PASS  npm run typecheck
  PASS  npm run lint (--max-warnings=0)
  PASS  npm test (--watchAll=false)
  WARN  npm audit (omit=dev, high+)
== FE anti-pattern scan · web/ ==
  PASS  no getServerSideProps / getStaticProps / getInitialProps in web/pages/
  PASS  no web/pages/api/ directory
  PASS  no raw fetch() outside web/lib/api/
REVIEW GATE: PASS
```
(Side note: the `printf -- invalid option` line appeared in the warn output, confirming the line 71 bug.)

### BE gate (this run)
```
== BE gate · services/agent-board ==
MISSING TOOL: gosec
  install: go install github.com/securego/gosec/v2/cmd/gosec@latest
(exit 2)
```

### REQ004 production-code diff scope vs. test-code diff scope (sanity check)
```
services/agent-board/internal/repo/document_repo.go     | 2 +-     (production)
services/agent-board/internal/repo/document_repo_test.go| 50 ++++++/--   (tests)
services/agent-board/internal/handler/project_handler.go  | + (US012 handler — NEW function)
services/agent-board/internal/handler/project_handler_test.go | + (US012/US013 tests)
services/agent-board/internal/handler/document_handler.go | + (US013 handlers — NEW functions)
services/agent-board/internal/handler/document_handler_test.go | + (US013 tests, 23 KB)
web/... (16 new components + hooks + tests, package-lock churn for markdown stack)
```
