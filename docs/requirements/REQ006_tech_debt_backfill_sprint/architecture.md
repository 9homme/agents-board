# Architecture — REQ006 Tech-Debt Backfill Sprint

**Approval:** approved
**Approved-by:** human
**Approved-at:** 2026-06-04T03:14:00Z

---

## 0. Reading guide

REQ006 is a **mixed-track tech-debt sprint**, NOT a feature delivery. It adds **zero endpoints, zero pages, zero microservices, zero migrations**. The architecture document therefore omits the usual "API contracts (exact JSON)" mega-section — there are no new contracts to lock. Substance lives in:

- **§2 Scope reaffirmation** — what's in, what's out, and why the usual sections are skipped.
- **§3 File-level touch map (US001..US015)** — the authoritative scoping table for tech-lead + devs.
- **§4 Cluster-1+2+3 test pattern (the load-bearing section)** — the sqlmock / mock-repo / httptest patterns each cluster follows, defined ONCE so devs and tester do not re-derive per story.
- **§5 US010 env-var harmonisation contract** — `DATABASE_URL` wins, helper shape, log-line wording.
- **§6 US012 Go toolchain bump** — pinned `go 1.26.4` based on live `govulncheck` run.
- **§7 US013 FE TabSwitcher scope** — confirms component capabilities and pins FCT-* IDs the tester will produce.
- **§8 US008 message.go test harness shape** — httptest + Echo for `HandleMessage`, direct call for the two helpers.
- **§9 US014 ADR-001 — MCP-only-writes (inline ADR text).** This IS the deliverable of US014.
- **§10 Skill / hook usage** — TDG + react-doctor + live-e2e + 3-clean-run flake check, with the US014 documentation-only carve-out.
- **§11 Decision log** — D-001..D-005 echoed from po-ba README, plus D-006..D-014 introduced here.
- **§12 Risks & open questions** — anything the human still needs to confirm at re-approval.

### 0.1 Executive summary (one-screen scan)

1. **Cluster 1+2+3 test pattern is locked** in §4 — sqlmock + `WillReturnError(sql.ErrNoRows)` for `_NotFound`, generic `WillReturnError` for `_GenericError`, `RowError(...)` for `_RowsErr`, `AddRow(<wrong-type>)` for `_ScanError`, hand-written mock-repo (preferred) or `testify/mock` for handler tests. `≥95% modulo enumerated unreachable lines` is the exemption mechanism; documented in test report under OQ-4.
2. **US010 env-var contract: `DATABASE_URL` is the SOLE accepted env var; `DB_URL` is REJECTED at startup with a fatal error** (D-006, revised per human direction at Phase 1 HARD STOP). There is no precedence rule because there is only one variable. Helper still lives in new `services/agent-board/internal/config/dburl.go` (kept for the unit-test surface US010's AC requires; ≥95% per-package coverage). Helper API simplifies to `ResolveDBURL() (url string, err error)` — returns the `DATABASE_URL` value on the happy path; returns a clear, operator-actionable error if `DB_URL` is set (telling the operator to rename), if both are set (telling the operator to remove `DB_URL`), or if neither is set. Startup log line: just `"db config: using DATABASE_URL"` (single happy-path variant — the both-set log variant is dropped because both-set is now a fatal error, not a precedence resolution). `docker-compose.yml` switches mcp-server's env block to `DATABASE_URL` in the same commit; api-server is already on `DATABASE_URL`. Migration impact is explicit and intentional: any external operator (Helm chart, custom compose override, CI script) still passing `DB_URL` to mcp-server MUST rename — there is no quiet upgrade path. This is the deliberate cost of the hard-fail choice (avoids silent dual-source-of-truth bugs during the migration).
3. **US012 Go toolchain target: `go 1.26.4`** (D-007). Live `govulncheck` against the host `go1.26.3` showed TWO real findings (`crypto/x509` GO-2026-5037 + `net/textproto` GO-2026-5039), both fixed in `1.26.4`. `go.mod` updated to `go 1.26.4` with `toolchain go1.26.4`; `Dockerfile:9` builder updated to `golang:1.26-alpine` (the alpine tag tracks the latest 1.26.x at build time, which is `1.26.4` or newer).
4. **US014 ADR location: INLINE in this `architecture.md` as §9** (D-008). Not a new `docs/adr/` convention — REQ006 is the requirement that formalises the decision, so the rationale lives next to the rest of the REQ. The text appears in §9 below, ready for `read`-based verification by tech-lead.
5. **US013 TabSwitcher target: ≥80% stmts/branches/lines/functions, no `Home`/`End`** (D-009). The component does NOT implement `Home`/`End`; the AC already excludes them. FCT-* IDs pinned in §7.
6. **US008 `message.go` harness: httptest + Echo for `HandleMessage`, direct call for `sendError` + `sendToolResultError`** (D-010). Real `mcp.SessionManager` (no interface introduction) + real `mcp.ToolRegistry` with controllable tool registrations — matches existing handler-test convention.
7. **US015 Makefile consolidation (D-013, with US011 absorbed per D-014):** delete `startup.sh`/`shutdown.sh`; add `make dev-up/down/migrate/seed`. Native local Postgres on `:5432` (Q1); existing `e2e-*` family byte-identical (Q2); PID + log files at repo root mirror `startup.sh` byte-for-byte (Q3); BOTH `PG_CONN ?=` (e2e, port 15432) AND new `DEV_PG_CONN ?=` (dev, port 5432) env-overridable in the same change (absorbed from former US011, Option A union). Zero `DB_URL` in any new recipe. Closes tech-debt line 86 (PG_CONN hardcoded) and the original US011 ask. US015 SHOULD ship before or with US010.
8. **No new open questions left for the human.** OQ-1 → D-006. OQ-2 → D-007. OQ-3 → D-008. OQ-4 → carry-forward via §4.5 exemption mechanism. OQ-5 → D-010. US015 introduced under D-013 with Q1–Q4 already answered by the user; D-014 merges former US011 into US015 (Option A union — both `?=` overrides preserved). Nothing further to surface unless the human disagrees with a specific decision above.

---

## 1. Scope reaffirmation

### 1.1 In scope (verbatim from po-ba README §"Scope summary")

Five clusters, fourteen stories:

- **Cluster 1 — `internal/repo/` error-branch backfill (BE-test only).** US001 (`task_repo.go`), US002 (`user_story_repo.go`), US003 (`audit_repo.go`).
- **Cluster 2 — `internal/handler/*_tools.go` error-mapping backfill (BE-test only).** US004 (`project_tools.go`), US005 (`document_tools.go`), US006 (`task_tools.go`), US007 (`user_story_tools.go`), US008 (`message.go`).
- **Cluster 3 — `internal/mcp/server.go` ToolRegistry + Session message methods (BE-test only).** US009.
- **Cluster 4 — Env-var harmonisation (BE-prod).** US010.
- **Cluster 5 — Housekeeping + ADR.** US012 (Go toolchain bump), US013 (FE TabSwitcher coverage), US014 (ADR — MCP-only-writes), US015 (Makefile consolidation, absorbs former US011 — retires `startup.sh`/`shutdown.sh` and flips `PG_CONN := → ?=`).

### 1.2 Out of scope (explicit non-goals — anti-scope)

Repeated from po-ba README §"Anti-scope" so all four downstream agents (tech-lead, tester, be-dev, fe-dev) share one boundary:

- **No new REST endpoints on `api-server`.** Stays at 4 GET routes. Adding write endpoints requires a future REQ. (See §9 ADR-001 for the formal decision.)
- **No new product features.** No new pages, no new business logic, no new MCP tools.
- **No microservice split.** `services/agent-board/` stays one Go module.
- **No FE state-management migration.** No Redux / Zustand / React Query.
- **No test-framework swap.** Go `testing` + `testify` + `sqlmock` (BE), Jest + RTL + MSW (FE), Robot Framework (e2e) continue as-is.
- **No production code changes in cluster-1/2 stories.** US001–US008 are tests-only. If a backfill test reveals a real bug (e.g. a missing `rows.Err()` check), the dev raises `ARCHITECTURE_GAP_FOUND` and the orchestrator routes back to system-architect for a new follow-up story — devs do NOT silently fix in this scope.
- **No retroactive REQ001-005 feature changes.** REQ001–005 shipped; this is debt cleanup.
- **No e2e additions.** Unit/integration backfill on BE, component-test backfill on FE. No new `.robot` files. (US015 and US012 incidentally exercise `make e2e-up/run/down`, but they do not add new e2e tests.)

### 1.3 Sections deliberately omitted

- **§"API contracts (exact)"** — REQ006 adds zero endpoints; the existing REQ001–005 contracts are unchanged byte-for-byte. The architecture-template section would be vacuous; omitted on purpose, per the system-architect agent's REQ006-specific brief.
- **§"Data flow / sequence diagram"** — no new request shapes. Existing REQ001/004 sequence diagrams remain authoritative.
- **§"Data model"** — no schema changes, no new migrations, no new indexes.

---

## 2. Service topology

Unchanged from REQ005. Restated here so reviewers do not need to cross-check.

| Service | New / Modified by REQ006 | Responsibility | Inter-service calls |
|---|---|---|---|
| `services/agent-board/cmd/api-server` | modified (US010 env var, US012 toolchain) | Read-only REST surface (4 GETs on `/api/v1/...`) | Postgres only |
| `services/agent-board/cmd/mcp-server` | modified (US010 env var, US012 toolchain) | MCP write surface — SSE + `/message` JSON-RPC + tool registry | Postgres only |
| `services/agent-board/internal/{repo,handler,mcp,domain}` | test files modified (US001..US009); production code untouched | Domain + persistence + MCP plumbing | — |
| `services/agent-board/internal/config` | **NEW package** for US010 (one file, `dburl.go`) | Resolves which env var supplies `DATABASE_URL`; emits startup log | — |
| `web/components/ProjectDetail/TabSwitcher.tsx` | test file modified (US013); production code untouched | WAI-ARIA Tabs pattern for project detail page | — |

---

## 3. File-level touch map (US001..US015)

This is the authoritative list of files each story touches. Tech-lead uses it for scoping; devs use it to know whether they are in their lane (any path outside this table is `WRONG_TRACK` or `ARCHITECTURE_GAP_FOUND`). **Paths in bold are NEW files; the rest are edits.**

### US001 — `task_repo.go` error-branch tests

| File | New / Edit | Change |
|---|---|---|
| `services/agent-board/internal/repo/task_repo_test.go` | edit | Add the 12 test functions named verbatim in US001 AC. Follow §4 pattern. Test-only. |
| `services/agent-board/internal/repo/task_repo.go` | **no change** | Byte-for-byte unchanged. |

### US002 — `user_story_repo.go` error-branch tests

| File | New / Edit | Change |
|---|---|---|
| `services/agent-board/internal/repo/user_story_repo_test.go` | edit | Add 12 test functions per US002 AC. Same shape as US001. |
| `services/agent-board/internal/repo/user_story_repo.go` | **no change** | Byte-for-byte unchanged. |

### US003 — `audit_repo.go` error-branch tests

| File | New / Edit | Change |
|---|---|---|
| `services/agent-board/internal/repo/audit_repo_test.go` | edit | Add 6 test functions per US003 AC (3 error branches × 2 public callers). |
| `services/agent-board/internal/repo/audit_repo.go` | **no change** | Byte-for-byte unchanged. |

### US004 — `project_tools.go` error-mapping tests

| File | New / Edit | Change |
|---|---|---|
| `services/agent-board/internal/handler/project_tools_test.go` | edit | Add 18 test functions per US004 AC. Includes `TestRegisterProjectTools_RegistersAllFiveTools` (covers 0% on `RegisterProjectTools`). Use mock `repo.ProjectRepository` per §4.3. |
| `services/agent-board/internal/handler/project_tools.go` | **no change** | Byte-for-byte unchanged. |

### US005 — `document_tools.go` error-mapping tests

| File | New / Edit | Change |
|---|---|---|
| `services/agent-board/internal/handler/document_tools_test.go` | edit | Add 20 test functions per US005 AC. Existing `MockDocumentRepo` already at top of file — reuse it. Includes `TestRegisterDocumentTools_RegistersAllFiveTools`. |
| `services/agent-board/internal/handler/document_tools.go` | **no change** | Byte-for-byte unchanged. |

### US006 — `task_tools.go` error-mapping tests

| File | New / Edit | Change |
|---|---|---|
| `services/agent-board/internal/handler/task_tools_test.go` | edit | Add 25 test functions per US006 AC. `UpdateTaskTool` carries the bulk (5 distinct status-change branches + no-status-change branch). |
| `services/agent-board/internal/handler/task_tools.go` | **no change** | Byte-for-byte unchanged. |

### US007 — `user_story_tools.go` error-mapping tests

| File | New / Edit | Change |
|---|---|---|
| `services/agent-board/internal/handler/user_story_tools_test.go` | edit | Add 27 test functions per US007 AC. Note the **passthrough** error semantics (no `fmt.Errorf` wrap on most paths — verify via `errors.Is(returnedErr, mockErr)`). |
| `services/agent-board/internal/handler/user_story_tools.go` | **no change** | Byte-for-byte unchanged. |

### US008 — `message.go` error-routing tests

| File | New / Edit | Change |
|---|---|---|
| **`services/agent-board/internal/handler/message_test.go`** | **NEW file** | Create from scratch. Add 13 test functions per US008 AC, following §8 harness. Use real `mcp.SessionManager` + real `mcp.ToolRegistry` with controllable tool registrations. |
| `services/agent-board/internal/handler/message.go` | **no change** | Byte-for-byte unchanged. |

### US009 — `internal/mcp/server.go` ToolRegistry + session messaging tests

| File | New / Edit | Change |
|---|---|---|
| **`services/agent-board/internal/mcp/server_test.go`** | **NEW file** | Create from scratch. Holds the 15 test functions per US009 AC: `TestNewToolRegistry_*`, `TestToolRegistry_*`, `TestSession_*`, `TestSessionManager_RemoveSession_*`. Tester MAY split into `tool_registry_test.go` + `session_test.go` additions; the existing `session_test.go` is tiny and may be merged into. Test names are authoritative regardless of placement. |
| `services/agent-board/internal/mcp/server.go` | **no change** | Byte-for-byte unchanged. Note: `ListTools` doc-comment claims "lexicographic order" but the implementation does NOT sort. AC for `TestToolRegistry_ListTools_ReturnsAllRegisteredNames` uses unordered membership check; doc/code mismatch is **flagged as tech-debt** in the test report under OQ-4 but **not silently fixed** here. |

### US010 — Standardise on `DATABASE_URL` (reject `DB_URL`)

| File | New / Edit | Change |
|---|---|---|
| **`services/agent-board/internal/config/dburl.go`** | **NEW file** | New package `config`. Exports `func ResolveDBURL() (url string, err error)`. Single accepted env var: `DATABASE_URL`. `DB_URL` set → returns an actionable error (caller `log.Fatal`s). See §5 for the exact contract. |
| **`services/agent-board/internal/config/dburl_test.go`** | **NEW file** | Unit tests for the four env-state combinations listed in §5.6. Uses `t.Setenv`. Covers ≥95% of the new package. |
| `services/agent-board/cmd/api-server/main.go` | edit | Replace lines 44–48 (`os.Getenv("DATABASE_URL")` + log.Fatal block) with `dbURL, err := config.ResolveDBURL()` + `if err != nil { log.Fatal(err) }` + the happy-path log line. Add `agent-board/internal/config` to imports. Runtime behaviour for api-server is unchanged on the happy path (api-server already only reads `DATABASE_URL`). |
| `services/agent-board/cmd/mcp-server/main.go` | edit | Replace lines 30–33 (`os.Getenv("DB_URL")` + log.Fatal block) with `dbURL, err := config.ResolveDBURL()` + `if err != nil { log.Fatal(err) }` + the happy-path log line. **This is the meaningful runtime change in REQ006.** mcp-server now reads `DATABASE_URL`. Any deployment still passing `DB_URL` to mcp-server will HARD-FAIL at startup until the env-var is renamed. The api-server vs mcp-server diff is now identical at this site. |
| `docker-compose.yml` | edit | Change `mcp-server.environment.DB_URL` (line 48) → `DATABASE_URL` (same URL value). `api-server` already sets `DATABASE_URL`. Drop the pre-existing-inconsistency comment (lines 46–47) and replace with a one-line `# Standardised on DATABASE_URL per REQ006/US010 (D-006). DB_URL is rejected at startup.` Both services now set `DATABASE_URL` and ONLY `DATABASE_URL`. |
| `services/agent-board/cmd/api-server/main_test.go` | edit OR new | If startup-log integration test is added (US010 AC scenario "Scenario: startup log line is integration-tested"), place here. Tester's call between subprocess `os/exec` vs `log.SetOutput` capture; either acceptable. Asserts the single happy-path line `"db config: using DATABASE_URL"`. |
| `services/agent-board/cmd/mcp-server/main_test.go` | edit OR new | Same. Plus (optional) an integration test that spawns mcp-server with `DB_URL` set / `DATABASE_URL` unset and asserts it exits non-zero with the rename-instruction error message — guards against regressing the hard-fail. |
| `tests/e2e/README.md` | edit | If it references `DB_URL`, update to: "`DATABASE_URL` only; `DB_URL` is rejected at startup as of REQ006/US010". (Tech-lead to confirm by grep during planning.) |
| `docs/tech_debt.md` | edit | Strike-through line 97 with `→ fixed in REQ006/US010 (standardised on DATABASE_URL; DB_URL rejected at startup)`. |

**Note (US015 sequencing):** `startup.sh` and `shutdown.sh` at repo root are NOT in US010's touch map — they are **deleted by US015**, not edited by US010. The legacy `startup.sh` references `DB_URL` (which is exactly what motivated US015), but US010's be-dev MUST NOT touch those scripts: by the time US010 lands they may already be gone (US015 sequencing — see §11 D-013 and §13.1 R-6). If US010 lands first, the existing `startup.sh` workflow breaks at the next invocation with the hard-fail rename-instruction error from §5.4; that breakage is the trigger to merge US015 next. Tech-lead's call whether to pair both PRs into a single merge or sequence US015 → US010 in two PRs.

### US012 — Go toolchain bump

| File | New / Edit | Change |
|---|---|---|
| `services/agent-board/go.mod` | edit | Line 3: `go 1.25.0` → `go 1.26.4`. Add `toolchain go1.26.4` directive on a new line (`go` directive + `toolchain` directive is the modern shape). |
| `services/agent-board/Dockerfile` | edit | Line 9: `FROM golang:1.25-alpine AS build` → `FROM golang:1.26-alpine AS build`. The minor-tracking alpine tag picks up the latest patch (1.26.4 or newer) at build time. |
| `docs/tech_debt.md` | edit | Strike-through line 28 with `→ fixed in REQ006/US012`. |

### US013 — `TabSwitcher.tsx` coverage backfill

| File | New / Edit | Change |
|---|---|---|
| `web/components/ProjectDetail/TabSwitcher.test.tsx` | edit | Add the FCT-* test cases enumerated in §7. Use `@testing-library/user-event` for keyboard events. Match elements by accessible name + role. Target ≥80% stmts/branches/lines/functions. |
| `web/components/ProjectDetail/TabSwitcher.tsx` | **no change** | Byte-for-byte unchanged. |
| `docs/tech_debt.md` | edit | Strike-through line 75 with `→ fixed in REQ006/US013`. |

### US014 — ADR — MCP-only-writes

| File | New / Edit | Change |
|---|---|---|
| `docs/requirements/REQ006_tech_debt_backfill_sprint/architecture.md` (this file) | edit | The ADR text is in **§9 below**, authored by system-architect in this revision. Tech-lead's task for US014 reduces to "verify §9 satisfies every AC scenario in `US014_*.md`." No code changes. |
| `docs/tech_debt.md` | edit | Strike-through line 98 with `→ won't-fix per REQ006/US014 ADR-001 (MCP-only-writes is permanent)`. |

### US015 — Consolidate dev workflow into the Makefile; retire `startup.sh`/`shutdown.sh`; flip existing `PG_CONN := → ?=` (absorbs former US011 per D-014)

| File | New / Edit | Change |
|---|---|---|
| `startup.sh` | **DELETE** | Repo-root file (~38 lines). Legacy local-dev launcher superseded by `make dev-up`. Deletion is explicit — be-dev removes the file, does not edit it. |
| `shutdown.sh` | **DELETE** | Repo-root file (~30 lines). Legacy teardown script superseded by `make dev-down`. Deletion is explicit. |
| `Makefile` | edit (existing line ~16) | Existing `PG_CONN := postgres://...:15432/...` on line ~16 → change `:=` to `?=` (env-overridable; absorbed from former US011 per D-014). **Default URL byte-identical.** This is the one-character flip that closes tech-debt line 86. |
| `Makefile` | edit (new lines) | Add four new targets: `dev-up`, `dev-down`, `dev-migrate`, `dev-seed`. Add new variable `DEV_PG_CONN ?= postgres://agent_board:agent_board@localhost:5432/agent_board?sslmode=disable` on its own line. **Both variables now env-overridable via `?=` (Option A union per D-014):** `PG_CONN ?=` for the e2e family on port 15432, `DEV_PG_CONN ?=` for the dev family on port 5432. They serve distinct workflows and MUST NOT be collapsed into one variable. Both `api-server` and `mcp-server` receive ONLY `DATABASE_URL` (validates US010 alignment — zero `DB_URL` in any new recipe). Preserve every existing `e2e-*` target body byte-identical (US015 AC scenario "existing `make e2e-*` targets are byte-identical"). PID files (`.mcp.pid`, `.api.pid`, `.web.pid`) and log files (`mcp-server.log`, `api-server.log`, `web.log`) at repo root, mirroring `startup.sh` byte-for-byte (Q3). Update any `make help` convention to list the four new targets. |
| `README.md` (repo root) | edit | Replace any `./startup.sh` / `./shutdown.sh` references with `make dev-up` / `make dev-down`. Add a brief paragraph on the dev-local workflow + native-Postgres prerequisite (Q1). |
| `tests/e2e/README.md` | edit (optional) | Add 1–2 lines distinguishing `dev-*` family (native processes, native local Postgres on `:5432`) from `e2e-*` family (docker-compose stack, dockerised Postgres on `:15432`). Light touch — exact wording is the dev's call. |
| `services/agent-board/README.md` | edit (if exists; conditional) | If the file exists and references `startup.sh`/`shutdown.sh`, replace with `make dev-*` equivalents. If it does not exist, no-op. |
| `.claude/agents/*.md` | edit (sweep) | `git grep -l 'startup\.sh\|shutdown\.sh' .claude/agents/` → for each match, replace references with the appropriate `make dev-*` target (most commonly `make dev-up` / `make dev-down`). Then run `python3 scripts/sync-gemini.py` per project rule (CLAUDE.md §"Editing agent definitions"). Do NOT hand-edit `.gemini/agents/*.md`. |
| `CLAUDE.md` | edit (conditional) | `git grep 'startup\.sh\|shutdown\.sh' CLAUDE.md` → if non-empty, replace with `make dev-*` equivalent. Likely a no-op; verify. |
| `docs/tech_debt.md` | edit | Strike-through **line 86** with `→ fixed in REQ006/US015 (absorbed from former US011 for line 86; PG_CONN now `?=` and env-overridable)`. **Additionally** if any line references `startup.sh` / `shutdown.sh` directly, strike-through with `→ fixed in REQ006/US015`. |

**Note (variable namespace — both `?=` overrides preserved per D-014, Option A union):** `PG_CONN ?=` (e2e family, port 15432, absorbed from former US011) and `DEV_PG_CONN ?=` (dev family, port 5432, new in US015) are deliberately separate variables on separate Makefile lines serving separate workflows. They use the same `?=` idiom but MUST NOT be collapsed into one variable. Option B (drop the e2e override and use only `DEV_PG_CONN`) was rejected at D-014 because it would re-open tech-debt line 86 (the original US011 ask).

**Note (sequencing with US010):** US015 SHOULD ship before or in the same merge as US010 — see §11 D-013 consequences and §13.1 R-6. No hard `Blocked by` link; tech-lead's call whether to pair the PRs.

**Note (no production-Go-code change):** US015 touches the Makefile + repo-root scripts + docs + agent definitions. Zero `services/agent-board/` Go source files are edited. The "no production code change for cluster-1/2/3 tests-only" rule remains accurate; US015 is a separate cluster-5 housekeeping touch, in the same shape as US012 (toolchain + Dockerfile).

**Note (no regression test for the `?=` flip):** The merged US015 AC explicitly excludes the standalone `scripts/review/test/test_makefile_pg_conn.sh` regression test that the original US011 had proposed. The `?=` change is exercised implicitly by `make e2e-up` continuing to work; a dedicated assertion script is out of scope.

---

## 4. Cluster 1 + 2 + 3 test pattern (the load-bearing section)

This section defines the **single shared test pattern** that US001..US009 follow. It is written ONCE here so devs and tester do not re-derive it per story. The pattern is a direct descendant of REQ005/US005's already-proven approach for `document_repo` / `project_repo`.

### 4.1 Reference: REQ005/US005

The canonical example. Re-read `services/agent-board/internal/repo/project_repo_test.go` and `services/agent-board/internal/repo/document_repo_test.go` for the exact shape: `sqlmock.New()` → `mock.ExpectQuery / ExpectExec / ExpectBegin / ExpectCommit` configured per branch → invoke repo method → assert `(returnValue, err)` pair → assert `mock.ExpectationsWereMet()`. Cluster 1 (US001–US003) mirrors this pattern across the remaining repos. Cluster 2 (US004–US008) lifts the same idea up one layer — replace `sqlmock` with a hand-written mock `repo.<Entity>Repository`. Cluster 3 (US009) drops to bare struct testing because the SUT is in-memory plumbing.

### 4.2 Cluster 1 — repo error-branch tests (sqlmock)

**Test file structure (uniform across `task_repo_test.go`, `user_story_repo_test.go`, `audit_repo_test.go`):**

```go
func TestXxxRepo_<Method>_<BranchName>(t *testing.T) {
    // Arrange
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer func() { _ = db.Close() }()

    r := NewXxxRepo(db) // or the concrete struct constructor

    mock.ExpectQuery(`^...`).
        WithArgs(<args>).
        WillReturnError(errors.New("db down")) // OR sql.ErrNoRows, or AddRow(<wrong-type>), or RowError(idx, err)

    // Act
    result, err := r.<Method>(context.Background(), <args>)

    // Assert
    assert.Nil(t, result)
    assert.Error(t, err)
    // For _NotFound branches:
    //   assert.ErrorIs(t, err, ErrNotFound)
    // For _GenericError / _QueryError / wrap branches:
    //   assert.False(t, errors.Is(err, ErrNotFound))
    //   assert.Contains(t, err.Error(), "<wrap prefix from source>") // for wrap branches only
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

**Branch → mock-shape mapping (verbatim, copy-pasteable for each story):**

| Branch suffix | sqlmock idiom | Production-code condition exercised |
|---|---|---|
| `_GenericError` | `mock.ExpectQuery(...).WillReturnError(errors.New("db down"))` | non-`sql.ErrNoRows` from `QueryRowContext` |
| `_NotFound` | `mock.ExpectQuery(...).WillReturnError(sql.ErrNoRows)` | `errors.Is(err, sql.ErrNoRows)` → `ErrNotFound` mapping |
| `_QueryError` | `mock.ExpectQuery(...).WillReturnError(errors.New("db down"))` | `QueryContext` error (the `ListXxx` shape) |
| `_ScanError` | `mock.ExpectQuery(...).WillReturnRows(sqlmock.NewRows(cols).AddRow(<wrong-type>))` | `rows.Scan` type-mismatch failure mid-loop |
| `_RowsErr` | `mock.ExpectQuery(...).WillReturnRows(sqlmock.NewRows(cols).AddRow(...).RowError(0, errors.New("rows err")))` | `rows.Err()` after the loop |
| `_BeginTxError` | `mock.ExpectBegin().WillReturnError(errors.New("begin fail"))` | `BeginTx` returns err (transactional path) |
| `_AuditInsertError` | `mock.ExpectBegin(); mock.ExpectQuery(...).WillReturnRows(<happy>); mock.ExpectExec(<audit-insert>).WillReturnError(errors.New("audit fail")); mock.ExpectRollback()` | `tx.ExecContext` on the audit insert fails |
| `_CommitError` | `mock.ExpectBegin(); mock.ExpectQuery(<happy>); mock.ExpectExec(<happy>); mock.ExpectCommit().WillReturnError(errors.New("commit fail"))` | `tx.Commit()` returns err |
| `_UpdateGenericError` (transactional) | `mock.ExpectBegin(); mock.ExpectQuery(<update>).WillReturnError(errors.New("update fail")); mock.ExpectRollback()` | non-`sql.ErrNoRows` from the transactional `QueryRowContext` |

**Note on `_AuditInsertError` / `_CommitError`:** The current production code's `defer { if err != nil { tx.Rollback() }} ` pattern means sqlmock needs `ExpectRollback()` declared so the rollback call does not trip `ExpectationsWereMet`. Tester / dev confirms by reading `task_repo.go:96-102` and `user_story_repo.go:65-71`.

**Coverage assertion:** Each story's `≥95% per-file` threshold is checked via:

```
cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo.out -run '<RegexpPerStoryAC>'
go tool cover -func=/tmp/repo.out | grep <filename>
```

The `_RowsErr` / `_ScanError` shapes are sufficient to lift coverage past 95% on every cluster-1 file given REQ005/US005's empirical results on the analogous files.

### 4.3 Cluster 2 — handler error-mapping tests (mock-repo)

**Mock-repo shape (preferred — already established by `MockDocumentRepo` in `document_tools_test.go:19-43`):**

```go
type Mock<Entity>Repo struct {
    repo.<Entity>Repository // embed for forward-compat; only override what each test needs
    CreateXxxFunc func(ctx context.Context, x *domain.Xxx) (*domain.Xxx, error)
    GetXxxFunc    func(ctx context.Context, id string) (*domain.Xxx, error)
    UpdateXxxFunc func(ctx context.Context, x *domain.Xxx) (*domain.Xxx, error)
    DeleteXxxFunc func(ctx context.Context, id string) error
    ListXxxFunc   func(ctx context.Context, parentID string) ([]*domain.Xxx, error)
    // Plus status-update method for task/user_story
}
// Implement the interface methods by delegating to Func fields.
```

`testify/mock` is acceptable as an alternative IF tester / dev prefers it (US004, US006 notes both allow). Hand-written wins on readability for a 5-method interface; `testify/mock` wins when many tests share setup. Pick once per file, do not mix.

**Test file structure (uniform across `project_tools_test.go`, `document_tools_test.go`, `task_tools_test.go`, `user_story_tools_test.go`):**

```go
func TestHandle<Action><Entity>_<BranchName>(t *testing.T) {
    // Arrange
    registry := mcp.NewToolRegistry()
    mockRepo := &Mock<Entity>Repo{}
    handler.Register<Entity>Tools(registry, mockRepo)

    mockRepo.GetXxxFunc = func(ctx context.Context, id string) (*domain.Xxx, error) {
        return nil, repo.ErrNotFound // or errors.New("db down") etc per branch
    }

    tool, ok := registry.GetTool("<tool-name>")
    require.True(t, ok)

    // Act
    result, err := tool(context.Background(), json.RawMessage(`{"id":"abc"}`))

    // Assert
    assert.Nil(t, result)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "<exact wrap prefix from source>")
    // OR for passthrough (user_story_tools.go behaviour):
    //   assert.ErrorIs(t, err, mockErr)
}
```

**Branch → setup mapping:**

| Branch suffix | Mock setup | Production-code condition exercised |
|---|---|---|
| `_InvalidArguments` | _no mock needed_; invoke with malformed JSON like `json.RawMessage("not json")` | `json.Unmarshal` failure → `"invalid arguments"` (or `fmt.Errorf("invalid arguments: %w", err)` for document/task/user-story families) |
| `_EmptyID` / `_MissingID` / `_EmptyName` / `_MissingProjectIDOrTitle` / etc. | _no mock needed_; invoke with body that trims to empty | per-source validation strings (e.g. `"id is required"`, `"projectId and title are required"`, `"missing id"`) |
| `_NotFound` | `mockRepo.GetXxxFunc = func(...) { return nil, repo.ErrNotFound }` | `errors.Is(err, repo.ErrNotFound)` → unwrap to `errors.New("<entity> not found")` / `fmt.Errorf("<entity> not found")` |
| `_GenericError` / `_RepoError` | `mockRepo.<Method>Func = func(...) { return nil, errors.New("db down") }` | passthrough OR `fmt.Errorf("failed to <op> <entity>: %w", err)` (document/task wrap; project/user_story passthrough) |
| `_InvalidStatusTransition` (task/user_story only) | `mockRepo.GetXxxFunc` returns existing with status A; invoke with `Status` field = status B where transition is invalid | `existing.IsValidTransition(*req.Status)` false branch |
| `_StatusChange_*Error` (task/user_story only) | Get returns happy; `UpdateXxxStatus` or `UpdateXxx` returns err | corresponding wrap line in the handler |
| `_RegistersAll*Tools` | Register against a fresh `mcp.NewToolRegistry()`; assert each expected name resolves via `registry.GetTool(name)` and an unknown name returns `(nil, false)` | `RegisterXxxTools` registration calls |

**Assertion nuance (READ THE SOURCE FIRST):**

- `document_tools.go` and `task_tools.go` **wrap** repo errors with `fmt.Errorf("failed to <op> <entity>: %w", err)`. Assert via `assert.Contains(t, err.Error(), "failed to <op>")`.
- `project_tools.go` and `user_story_tools.go` **passthrough** most repo errors. Assert via `assert.ErrorIs(t, returnedErr, mockErr)` (the wrap-prefix check fails on these).
- `_NotFound` branches **return a fresh error** (`errors.New("<entity> not found")` for project/user_story, `fmt.Errorf("<entity> not found")` for document/task — both are sentinel-less). Assert via `assert.Contains(t, err.Error(), "<entity> not found")` AND `assert.False(t, errors.Is(err, repo.ErrNotFound))`.
- Every wrap-prefix string in the AC is the exact substring to match.

### 4.4 Cluster 3 — bare-struct unit tests (US009 only)

`internal/mcp/server.go` has no DB / no HTTP layer — the SUT is plain in-memory state with mutex coordination. Construct the struct directly (`mcp.NewToolRegistry()`, `mcp.NewSessionManager()`), invoke methods, assert.

**Test file structure:**

```go
func TestToolRegistry_<Method>_<BranchName>(t *testing.T) {
    registry := mcp.NewToolRegistry()
    // Optionally pre-populate via registry.RegisterTool(name, handler)
    handler, ok := registry.GetTool("some-name")
    assert.True(t, ok)
    assert.NotNil(t, handler)
}
```

**Special cases per US009 AC:**

- `TestToolRegistry_ConcurrentRegisterAndGet` and `TestSessionManager_RemoveSession_ConcurrentSafe` MUST run cleanly under `go test -race ./internal/mcp`. Use `sync.WaitGroup` + 100 goroutines minimum. The current code uses `sync.RWMutex`; properly written tests will not race.
- `TestSession_QueueMessage_FullReturnsError` — pre-fill exactly 100 messages (capacity is hard-coded `make(chan []byte, 100)` at `server.go:59`), assert 101st `QueueMessage` returns `errors.New("message queue full")`.
- `TestSession_ReceiveMessage_ContextCancelled` — `ctx, cancel := context.WithCancel(...)`; `cancel()`; assert returned error matches `context.Canceled` via `errors.Is`. Alternative `context.WithTimeout(ctx, 1*time.Millisecond)` for `DeadlineExceeded` is equally acceptable; tester picks ONE.
- `TestToolRegistry_ListTools_ReturnsAllRegisteredNames` — current code does NOT sort, despite the doc comment. Use unordered membership check (`assert.ElementsMatch`). DO NOT silently fix the doc-comment-vs-code mismatch; flag it in the test report under OQ-4 as a follow-up tech-debt line.

### 4.5 `≥95% modulo enumerated unreachable lines` — exemption mechanism

Some lines genuinely cannot be reached via the chosen test harness:

- **sqlmock-unreachable:** the `log.Printf` inside the `defer { rollback on err }` clauses (e.g. `task_repo.go:99`, `user_story_repo.go:68`) — these only fire when `tx.Rollback()` itself errors AND that error is not `sql.ErrTxDone`. sqlmock's rollback behaviour does not surface this combination realistically. **Acceptable.** Tester names these lines explicitly in the test report.
- **`json.Marshal` on always-marshallable structs:** the two `mcp.InternalError` fallback branches in `message.go:46` and `message.go:64` — the structs only contain `string` / `int` / `interface{}` fields whose contents (other than the tool result) are always marshallable. The tool result COULD be non-marshallable if a stub tool returns a `chan int` or similar; US008 notes this as an optional reach. **Acceptable to leave uncovered IF flagged in test report.**
- **Doc-comment mismatch path (`ListTools` ordering):** not an uncoverable line, but a behaviour discrepancy. Surface as test-report note + a new line in `docs/tech_debt.md`; do not silently fix.

**Test-report shape (each story Phase 3c report appends one block):**

```
### Coverage exemptions (OQ-4)
- services/agent-board/internal/repo/task_repo.go:99 — defer-rollback log.Printf path — unreachable via sqlmock (rollback returning non-ErrTxDone). Acceptable.
- services/agent-board/internal/handler/message.go:46 — json.Marshal fallback — unreachable for marshallable JSONRPCResponse. Acceptable.
```

If the tester / dev finds a line they can reach but did not — that is a real coverage gap, not an exemption. Treat as `changes_requested` at code review.

### 4.6 Local verification command per cluster (devs run before flipping to `in_review`)

| Cluster | Story | Command |
|---|---|---|
| 1 | US001 | `cd services/agent-board && go test ./internal/repo -cover -v -run TestTaskRepo` |
| 1 | US002 | `cd services/agent-board && go test ./internal/repo -cover -v -run TestUserStoryRepo` |
| 1 | US003 | `cd services/agent-board && go test ./internal/repo -cover -v -run TestAuditRepo` |
| 2 | US004 | `cd services/agent-board && go test ./internal/handler -cover -v -run "TestHandle(Create\|Get\|Update\|Delete\|List)Project\|TestRegisterProjectTools"` |
| 2 | US005 | `cd services/agent-board && go test ./internal/handler -cover -v -run "TestRegisterDocumentTools\|Test(Create\|Get\|Update\|Delete\|List)Document(s?)Tool"` |
| 2 | US006 | `cd services/agent-board && go test ./internal/handler -cover -v -run "TestRegisterTaskTools\|Test(Create\|Get\|Update\|Delete\|List)Task(s?)Tool"` |
| 2 | US007 | `cd services/agent-board && go test ./internal/handler -cover -v -run "TestRegisterUserStoryTools\|Test(Create\|Get\|Update\|Delete\|List)UserStor(y\|ies)Tool"` |
| 2 | US008 | `cd services/agent-board && go test ./internal/handler -cover -v -run "TestHandleMessage\|TestSendError\|TestSendToolResultError"` |
| 3 | US009 | `cd services/agent-board && go test ./internal/mcp -cover -race -v` |

Every story also runs the broader `cd services/agent-board && go test ./... && golangci-lint run ./...` before review.

---

## 5. US010 — Env-var harmonisation contract

### 5.1 Decision statement (D-006 — revised at Phase 1 HARD STOP)

**`DATABASE_URL` is the SOLE accepted DB-connection env var across `api-server` and `mcp-server`. `DB_URL` is REJECTED at startup with a fatal error (exit non-zero). There is no precedence rule because there is only one variable.**

Rationale (brief):

- **12-factor / Heroku convention** is `DATABASE_URL`. Matches what `api-server` already uses; mcp-server's `DB_URL` is the outlier and gets renamed.
- **Avoid silent dual-source-of-truth bugs.** A precedence rule (accept both, prefer one) means a forgotten `DB_URL` in a `.env` file or a stale Helm value can silently pick the wrong DB URL during the migration. The hard-fail forces operator awareness.
- **Migration is one-shot, not gradual.** The team owns both deployment surfaces (docker-compose for local + e2e); there is no third-party consumer that needs a quiet upgrade path.

Earlier draft of D-006 proposed a precedence rule (`DATABASE_URL` wins, `DB_URL` still accepted). Rejected by the human during the Phase 1 HARD STOP for the silent-misconfig risk above. The current decision is the second pass.

### 5.2 Helper shape (D-006 — shared `internal/config` package, simplified API)

A small new package `services/agent-board/internal/config` with one file `dburl.go`. The helper remains (despite the simpler semantics) so US010's AC ≥95% per-package coverage requirement has a clean unit-test surface — without it, the four env-state cases would need to be exercised via subprocess-spawning `main_test.go` tests.

Strawman implementation (the exact API is the dev's to lock at TDD time; the requirements after the snippet are authoritative):

```go
// Package config holds startup-time configuration resolution.
// Currently exposes ResolveDBURL — used identically by api-server and mcp-server.
package config

import (
    "errors"
    "os"
)

// ResolveDBURL returns the DATABASE_URL value, or an actionable error.
//
// Behaviour by env-state:
//   - DATABASE_URL set, DB_URL not set  → returns (value, nil).
//   - DATABASE_URL set, DB_URL also set → returns ("", error) telling the operator
//     to REMOVE DB_URL. We refuse to start in a partially-migrated env because the
//     operator's intent is ambiguous (did they mean to override? did they forget to
//     delete the old line?). Hard-fail surfaces the ambiguity loudly.
//   - DATABASE_URL not set, DB_URL set  → returns ("", error) telling the operator
//     to RENAME DB_URL to DATABASE_URL. The deprecated name is no longer supported.
//   - Neither set                       → returns ("", error) "DATABASE_URL environment
//     variable is required".
//
// Returns errors (not log.Fatal) so the caller (main.go) controls process exit and
// the helper stays unit-testable without sub-process spawning.
func ResolveDBURL() (url string, err error) {
    dbURL := os.Getenv("DATABASE_URL")
    legacyURL := os.Getenv("DB_URL")

    switch {
    case dbURL != "" && legacyURL != "":
        return "", errors.New("DB_URL is set but no longer supported; remove DB_URL from your environment to disambiguate (DATABASE_URL is the sole accepted name as of REQ006/US010)")
    case dbURL != "":
        return dbURL, nil
    case legacyURL != "":
        return "", errors.New("DB_URL is no longer supported; rename to DATABASE_URL (REQ006/US010)")
    default:
        return "", errors.New("DATABASE_URL environment variable is required")
    }
}
```

**Required properties of the final implementation (these are the contract — the strawman API may be refined as long as these hold):**

1. **Exported and unit-testable.** Package `config`, exported `ResolveDBURL` symbol, no `log.Fatal` inside the helper.
2. **Covers all four env-state combinations** with **distinct, operator-actionable error messages** (matching §5.4 wording).
3. **Returns `(string, error)` only.** Caller does the `log.Fatal` and the happy-path log line. This keeps the helper free of `*log.Logger` dependency-injection wiring and makes test assertions trivial (`assert.EqualError` / `assert.Contains` on `err.Error()`).
4. **No log emission from inside the helper.** The single happy-path log line `"db config: using DATABASE_URL"` is emitted from `main.go` immediately after a successful `ResolveDBURL()` call, before the DB ping.

Why a shared package rather than duplicating two ~15-line helpers across `cmd/*/main.go` (REQ005 D-008 chose duplication for the lifecycle-context case):

- US010's AC requires unit tests for the helper (≥95% per-package coverage). A package gives a single test surface (`dburl_test.go`) rather than two duplicate test files inside `cmd/api-server/` + `cmd/mcp-server/`.
- The behaviour is genuinely shared, not a 9-line lifecycle copy. Two callers, identical behaviour, four distinct error paths to assert against = the threshold for "factor it out" is met.
- The precedent of REQ005 D-008 (duplicate 9-line context glue) does not apply: that decision was about avoiding a new package for code whose only test surface was the cmd-level test. US010 IS the test surface.

### 5.3 `main.go` call sites (post-US010)

**api-server `main.go` lines 44–48 become:**

```go
dbURL, err := config.ResolveDBURL()
if err != nil {
    log.Fatal(err)
}
log.Print("db config: using DATABASE_URL")
```

**mcp-server `main.go` lines 30–33 become:**

```go
dbURL, err := config.ResolveDBURL()
if err != nil {
    log.Fatal(err)
}
log.Print("db config: using DATABASE_URL")
```

Both binaries are byte-identical at this site. The happy-path log line is emitted from `main.go` (not from the helper) so the helper stays log-free and unit-testable without buffer capture.

### 5.4 Error messages (locked wording — tester + be-dev assert against these)

| Env state | Behaviour | Exact message |
|---|---|---|
| Only `DATABASE_URL` set | Happy path — returns URL, nil err | _(log line `"db config: using DATABASE_URL"` emitted from `main.go`)_ |
| Only `DB_URL` set | Helper returns error; caller `log.Fatal`s | `"DB_URL is no longer supported; rename to DATABASE_URL (REQ006/US010)"` |
| Both set | Helper returns error; caller `log.Fatal`s | `"DB_URL is set but no longer supported; remove DB_URL from your environment to disambiguate (DATABASE_URL is the sole accepted name as of REQ006/US010)"` |
| Neither set | Helper returns error; caller `log.Fatal`s | `"DATABASE_URL environment variable is required"` |

Tester pins these exact strings as `assert.EqualError(t, err, "...")` (or `assert.Contains` if the dev decides to enrich the message at TDD time — but the leading clause must match for grep-friendliness in operator-facing logs).

### 5.5 `docker-compose.yml` change

`mcp-server.environment.DB_URL: postgres://...` (line 48, set by REQ005 D-009) → `mcp-server.environment.DATABASE_URL: postgres://...` (same URL value, renamed key). `api-server.environment.DATABASE_URL` is already correct — no change there. Drop the pre-existing-inconsistency comment block (lines 46–47). Add one new comment: `# Standardised on DATABASE_URL per REQ006/US010 (D-006). DB_URL is rejected at startup.`

**No mention of "still accepting both" — we don't.**

### 5.6 Unit tests for `ResolveDBURL` (per US010 AC — four cases)

In `services/agent-board/internal/config/dburl_test.go`:

1. `TestResolveDBURL_OnlyDatabaseURLSet_Happy` — `t.Setenv("DATABASE_URL", "postgres://x")`; explicitly unset `DB_URL` via `t.Setenv("DB_URL", "")` then `os.Unsetenv("DB_URL")` (or use the `t.Setenv("DB_URL", "")` plus a check that `os.Getenv` returns ""; see Go stdlib semantics). Assert `(url=="postgres://x", err==nil)`.
2. `TestResolveDBURL_OnlyDBURLSet_RejectsWithRenameError` — `DB_URL` set, `DATABASE_URL` unset. Assert `(url=="", err.Error() == "DB_URL is no longer supported; rename to DATABASE_URL (REQ006/US010)")`.
3. `TestResolveDBURL_BothSet_RejectsWithDisambiguateError` — both set. Assert `(url=="", err.Error() == "DB_URL is set but no longer supported; remove DB_URL from your environment to disambiguate (DATABASE_URL is the sole accepted name as of REQ006/US010)")`.
4. `TestResolveDBURL_NeitherSet_RejectsWithRequiredError` — both unset. Assert `(url=="", err.Error() == "DATABASE_URL environment variable is required")`.

**Note vs. the existing US010 story AC:** the previously-listed `BothSet_PreferredWins` test case disappears because both-set is now an error, not a precedence resolution. The four cases above replace it. Reconciliation flag: see §5.8.

Package coverage ≥95% (trivial — four cases cover every line of a four-branch switch).

### 5.7 Integration test for startup log line (per US010 AC)

Place under `cmd/api-server/main_test.go` and `cmd/mcp-server/main_test.go`. Two acceptable shapes; tester picks ONE:

- **(a)** Subprocess via `os/exec`: spawn `go run ./cmd/<binary>` with a fake `DATABASE_URL` (it will fail at the ping stage, that is fine), capture stdout+stderr, assert the log line `"db config: using DATABASE_URL"` precedes the ping-fail line.
- **(b)** Refactor `main.go` to expose a `run()` helper (already exists per current source) and call it directly from the test with `log.SetOutput(&buf)`; assert buf contains the line.

(b) is faster + more deterministic; (a) is more end-to-end. Both satisfy US010 AC.

**Optional but recommended hard-fail regression test (mcp-server only):** subprocess-spawn mcp-server with `DB_URL=postgres://x` and `DATABASE_URL` unset; assert non-zero exit code AND captured stderr contains `"DB_URL is no longer supported; rename to DATABASE_URL"`. Guards against a future refactor silently re-accepting `DB_URL`.

### 5.8 US010 story-AC reconciliation flag (route to po-ba)

**The existing `US010_harmonise_db_url_env_var.md` story has AC that is now stale and must be revised by po-ba before Phase 2 picks it up.** Specifically:

- Any scenario phrased as "both binaries accept both env-var names" → MUST be replaced with "both binaries accept ONLY `DATABASE_URL`; `DB_URL` is rejected at startup."
- Any scenario referencing a "precedence rule" (documented or enforced) → MUST be removed. There is no precedence rule.
- The `TestResolveDBURL_BothSet_PreferredWins` unit-test scenario name → MUST be replaced with the three error scenarios listed in §5.6 (DB_URL-only error, both-set error, neither-set error — plus the happy-path test stays).
- Any "Notes" section text describing precedence / dual-acceptance → MUST be rewritten to describe the hard-fail behaviour and migration impact.

System-architect does NOT edit the story file. The orchestrator must route to po-ba next (story revision) before Phase 2 spawns tester / tech-lead on US010. The other twelve stories (US001..US009, US012..US014) and the merged US015 (which absorbs former US011 per D-014) are unaffected by this revision and can proceed independently.

### 5.9 Migration impact (operator-facing — be explicit)

The post-REQ005 `docker-compose.yml` (D-009) uses `DATABASE_URL` for `api-server` and `DB_URL` for `mcp-server`. The mcp-server line is renamed in the same commit as the US010 code change. Internal infra: zero residual breakage.

**External operators (anyone running this code outside our compose):** any deployment that has been passing `DB_URL` to mcp-server — Helm chart values, custom compose overrides, CI scripts, hand-set shell envs — **MUST rename to `DATABASE_URL`**. There is no quiet upgrade path; mcp-server will refuse to start with the rename-instruction error from §5.4.

This is the deliberate, accepted cost of the hard-fail choice. The architecture is honest about it so the orchestrator can flag it in the test report and so the human can re-evaluate at re-approval if the operator-impact appetite changes.

### 5.10 Closes REQ005 architecture OQ-7

REQ005 architecture (rev 4, still `pending_approval`) carries OQ-7 about env-var harmonisation that was deferred. US010 closes it: single-var contract + helper shape + error wording + log-line wording locked here. REQ005 architecture can reference D-006 / §5 when it lands.

---

## 6. US012 — Go toolchain bump

### 6.1 Verified findings (live `govulncheck` run, 2026-06-03)

Run command: `cd services/agent-board && govulncheck ./...` against the host toolchain (`go version go1.26.3 darwin/arm64`).

**Result — TWO standard-library findings affect this code:**

1. **GO-2026-5039** — `net/textproto` — "Arbitrary inputs are included in errors without any escaping in `net/textproto`." Found in `net/textproto@go1.26.3`. **Fixed in `net/textproto@go1.26.4`.** Reached via `mcp.run` → `echo.Echo.Start` → eventually `textproto.Reader.ReadMIMEHeader`.
2. **GO-2026-5037** — `crypto/x509` — "Inefficient candidate hostname parsing in `crypto/x509`." Found in `crypto/x509@go1.26.3`. **Fixed in `crypto/x509@go1.26.4`.** Reached via `handler.HandleSSE` / `handler.HandleMessage` calling `fmt.Fprintf` which transitively reaches `x509.Certificate.Verify`.

Additional dep-tree findings exist but are NOT called by this code (govulncheck classifies them as "vulnerabilities in modules you require but your code doesn't appear to call these") — these are noise for US012's purposes.

### 6.2 Target version: **`go 1.26.4`** (D-007)

`1.26.4` is the lowest released minor that clears BOTH `crypto/x509` AND `net/textproto` findings. It is the current latest patch in the 1.26 line at the time of architecture authoring.

**Implications:**

- `services/agent-board/go.mod` line 3: `go 1.25.0` → `go 1.26.4`.
- Add `toolchain go1.26.4` directive on a new line (the modern shape; `go` directive is the language version, `toolchain` directive pins the build).
- `services/agent-board/Dockerfile:9`: `FROM golang:1.25-alpine` → `FROM golang:1.26-alpine`. The minor-tracking tag automatically picks the latest 1.26.x patch at image-build time. Locking to `golang:1.26.4-alpine` is acceptable but adds maintenance cost; the minor-tracking tag is fine because Docker layer caching makes the rebuild cheap.

### 6.3 Why not 1.27 or later

- 1.27 ships in mid-2026 per the Go release cadence (Feb + Aug); at REQ006 authoring it may not be released yet. US012 AC requires "lowest released minor that fixes" — 1.26.4 wins.
- 1.27 may introduce new vulnerabilities or new dep-tree compat issues not yet tested. 1.26.4 is the conservative pick.
- If 1.27 is released before US012 lands AND the human prefers it, that is a one-line override at re-approval; D-007 does not preclude.

### 6.4 CI / Docker knock-on

- The existing `Dockerfile:9` is the only `golang:<ver>` reference under `services/agent-board/`. Confirmed via `grep -r "golang:" services/agent-board/`.
- `scripts/review/run-gate.sh` does NOT pin a Go version; it runs whatever `go` is on PATH. No change needed in the gate script.
- `cd services/agent-board && go test ./...` runs against the dev's local toolchain. Per US012 AC, dev installs `go1.26.4` locally (`go install golang.org/dl/go1.26.4@latest && go1.26.4 download`) OR upgrades the system toolchain.
- The compose-stack rebuild (`make e2e-up`) recompiles the binaries against the new builder image automatically.

### 6.5 Govulncheck post-bump assertion

After US012 lands, `cd services/agent-board && govulncheck ./...` MUST exit clean (zero `Your code is affected by N vulnerabilities` lines). Tester adds this as an explicit assertion in the test report. If a new finding emerges from the bump (transitive dep that flares on 1.26.4 stdlib), that becomes a NEW follow-up story — do not silently fix in US012.

---

## 7. US013 — TabSwitcher FE story scope

### 7.1 Component capability confirmation

Read `web/components/ProjectDetail/TabSwitcher.tsx` (82 lines, verified at architecture authoring). Confirmed handlers:

- `onClick={() => onTabChange(tab.id)}` — mouse click activation.
- `onKeyDown` handles **`ArrowRight`, `ArrowLeft`, `Enter`, `' '`** (Space). Each calls `event.preventDefault()`.
- `ArrowRight` / `ArrowLeft` cycle via modulo arithmetic (wraps at boundaries).
- ARIA: `role="tablist"`, `aria-label="Project tabs"`, `role="tab"` on each button, `aria-selected`, `aria-controls={tabpanel-${id}}`, `id={tab-${id}}`, `tabIndex={isSelected ? 0 : -1}` (roving tabindex).
- **NOT IMPLEMENTED:** `Home`, `End`, `Tab` (handled by browser default), `Escape`. Any other key is a no-op.

po-ba's removal of `Home`/`End` from AC is correct. Pinning here so tester does not re-add them.

### 7.2 FCT-* test IDs (pinned for tester)

The tester will produce `US013_fe_unit_tests.md` containing the following FCT-* IDs in order. Each maps to a scenario in `US013_*.md` AC:

| FCT ID | Scenario | Mapping to AC scenario |
|---|---|---|
| FCT-US013-001 | Clicking non-active tab fires `onTabChange` with clicked id; component does not internally mutate `activeTab` | "Scenario: clicking a non-active tab fires `onTabChange` with the clicked tab id" |
| FCT-US013-002 | `ArrowRight` moves focus forward and fires `onTabChange` | "Scenario: `ArrowRight` moves focus and fires `onTabChange` to the next tab" |
| FCT-US013-003 | `ArrowRight` from last tab wraps to first | "Scenario: `ArrowRight` from the last tab wraps to the first" |
| FCT-US013-004 | `ArrowLeft` moves focus backward and fires `onTabChange` | "Scenario: `ArrowLeft` moves focus and fires `onTabChange` to the previous tab" |
| FCT-US013-005 | `ArrowLeft` from first tab wraps to last | "Scenario: `ArrowLeft` from the first tab wraps to the last" |
| FCT-US013-006 | `Enter` activates focused tab; calls `preventDefault` | "Scenario: `Enter` activates the focused tab" |
| FCT-US013-007 | `Space` activates focused tab; calls `preventDefault` | "Scenario: `Space` activates the focused tab" |
| FCT-US013-008 | `aria-selected` per tab reflects `activeTab` prop | "Scenario: `aria-selected` reflects the active tab" |
| FCT-US013-009 | Roving `tabIndex` per tab reflects `activeTab` prop | "Scenario: roving `tabIndex` reflects the active tab" |
| FCT-US013-010 | Prop-driven `activeTab` change re-renders with new active tab AND does NOT fire `onTabChange` | "Scenario: prop-driven `activeTab` override re-renders with the new active tab" |
| FCT-US013-011 | Tablist semantics (`role="tablist"`, `aria-label`, `aria-controls`, `id`) | "Scenario: tablist semantics are present" |
| FCT-US013-012 | Unrelated keys do NOT fire `onTabChange` AND do NOT call `preventDefault` | "Scenario: unrelated keys do not fire `onTabChange`" |

### 7.3 Coverage target (D-009)

**≥80% stmts AND ≥80% branches AND ≥80% lines AND ≥80% functions.** Below the BE-test 95% bar because component tests often hit React-internal branches that require unrealistic setups. The 80% figure is the project's existing FE convention (per `web/jest.config.*` if such thresholds are encoded; otherwise tester sets it).

The component is 82 lines, of which ~13 are JSX wrapper / className computation. 12 FCT-* tests on a 4-handler + 2-button component reaches 80% comfortably without forced gymnastics.

### 7.4 Tooling pin

`@testing-library/user-event` for keyboard interactions (the `user-event` API dispatches the full sequence `keydown → keyup` with focus management; `fireEvent.keyDown` alone skips focus updates). Match elements by `getByRole('tab', { name: /documents/i })` — not `getByText`, not `getByTestId`. This matches the existing project convention.

---

## 8. US008 — `message.go` test harness shape (resolves OQ-5)

### 8.1 Harness decision (D-010)

**`HandleMessage` is tested via `httptest` + Echo.** The two helpers (`sendError`, `sendToolResultError`) are tested by **direct call** on a `*Handler` constructed with a real `mcp.SessionManager` and pre-created session. **Do NOT introduce a `SessionManager` interface** for mocking — the existing handler tests in `handler_test.go` already use the real one; consistency wins over isolation.

### 8.2 Test setup boilerplate (shared across all 13 functions)

```go
package handler_test

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"

    "agent-board/internal/handler"
    "agent-board/internal/mcp"

    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func newTestHandler(t *testing.T) (*handler.Handler, *mcp.Session, *mcp.ToolRegistry) {
    t.Helper()
    sm := mcp.NewSessionManager()
    tr := mcp.NewToolRegistry()
    h := handler.NewHandler(sm, tr)
    sess := sm.CreateSession() // returns a known session ID
    return h, sess, tr
}

func postMessage(t *testing.T, h *handler.Handler, sessionID string, body []byte) (*httptest.ResponseRecorder, error) {
    t.Helper()
    e := echo.New()
    e.POST("/message", h.HandleMessage)
    req := httptest.NewRequest(http.MethodPost, "/message?sessionId="+sessionID, bytes.NewReader(body))
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()
    e.ServeHTTP(rec, req)
    return rec, nil
}
```

### 8.3 The `sendError` / `sendToolResultError` direct-call setup

These two methods take `*mcp.Session` not `echo.Context`, so they can be invoked from a test without an HTTP request:

```go
func TestSendError_QueuesAndReturnsEchoError(t *testing.T) {
    h, sess, _ := newTestHandler(t)
    // h.sendError is unexported — use a reflection-free approach by calling HandleMessage
    // with a request that routes to sendError, then asserting on what was queued.
    // Alternatively, expose a thin wrapper if tester / dev finds reflection ugly.
}
```

**Note on unexported access.** `sendError` and `sendToolResultError` are package-private (lowercase). The test file lives in `package handler_test` (external test package — matches the existing convention in `document_tools_test.go:1`). External tests CANNOT call unexported functions directly. Two acceptable workarounds; tester / dev picks ONE:

- **(a)** Add an `internal-only` test helper file `services/agent-board/internal/handler/handler_internal_test.go` with `package handler` that re-exports `var SendError = (*Handler).sendError` and `var SendToolResultError = (*Handler).sendToolResultError`. This is the idiomatic Go pattern for testing private methods from an external test package. **NOTE:** This is the only place REQ006 introduces a tiny prod-adjacent file — but it is a `_test.go` file, so it is not "production code change" by US008's no-prod-change rule. Confirmed acceptable.
- **(b)** Route `sendError` exclusively through `HandleMessage` (i.e. test it indirectly via the "wrong JSON-RPC version" / "non-tools-call method" paths) — those paths in production already cover `sendError`. This is what `TestHandleMessage_NonToolsCallMethod` and `TestHandleMessage_WrongJSONRPCVersion` already do per US008 AC items 4 and 5.

**Recommended:** (b) for items 4, 5, 6, 7 (which are routed through `HandleMessage` anyway); (a) for items 10, 11, 12, 13 ONLY IF the indirect routing leaves coverage gaps. Tester picks at spec time.

### 8.4 Queue-full setup

For `TestHandleMessage_QueueMessageFails` and `TestSendError_QueueFailure_LogsButReturnsEchoError`:

```go
// Pre-fill the session's message queue (capacity 100) to force the next QueueMessage to return error.
for i := 0; i < 100; i++ {
    err := sess.QueueMessage([]byte("filler"))
    require.NoError(t, err)
}
// Now the next QueueMessage call inside HandleMessage / sendError will return err.
```

### 8.5 The two `json.Marshal` fallbacks (US008 AC note)

`message.go:46` and `message.go:64` are reachable only if a registered tool returns a value that `json.Marshal` cannot serialise (e.g. a `chan int`). Optional: tester registers a stub tool that returns `make(chan int)`, then assertion is that `sendError` is invoked with `mcp.InternalError`. If skipped, the test report names these two lines under OQ-4 (§4.5 exemption mechanism).

---

## 9. ADR-001 — MCP-only-writes is the permanent write API

> This is the inline ADR text for US014. The system-architect authored it during Phase 1 per US014's direction; tech-lead's role is verification, po-ba's role is sign-off. **ADR location decision (D-008): inline here, NOT a separate `docs/adr/` document.** Rationale: REQ006 is the requirement that formalises the decision; the rationale lives next to the rest of the REQ. The downstream verification cost (read this section, check it matches the US014 AC scenarios) is identical either way; the discoverability cost is lower (one fewer place to look). If a future REQ formalises a `docs/adr/` convention with more than one ADR, US014's content gets lifted into `docs/adr/0001-mcp-only-writes.md` as part of that future REQ.

### 9.1 Status

**Accepted.** Effective: 2026-06-03 (REQ006 approval date). Authored by: system-architect. Approved by: human (REQ006 architecture re-approval). Affects: REQ001–REQ005 retroactively (existing code already conforms) and all future REQs.

### 9.2 Context

`api-server` (Go, REST, Echo) was scaffolded at REQ001 with FOUR `GET` endpoints under `/api/v1/`. `mcp-server` (Go, MCP/JSON-RPC over SSE + POST, Echo) was added later in the same `services/agent-board/` module to expose CRUD operations on projects, documents, user stories, and tasks via MCP tools. The two binaries share the same Postgres database; api-server reads, mcp-server reads-and-writes.

Three open questions periodically surface for contributors:

1. Why is `api-server` read-only? Is that an oversight, a temporary state, or intentional?
2. Should we add REST `POST` / `PUT` / `DELETE` endpoints to `api-server` to support browser-direct CRUD without going through MCP?
3. If the answer is "no," under what circumstances would we change our minds?

This ADR answers all three.

### 9.3 Decision

**The `api-server` is intentionally read-only. All create / update / delete operations on projects, documents, user stories, tasks (and any future write surface) are exposed exclusively through MCP tools (registered in `internal/handler/*_tools.go` and served by `cmd/mcp-server`). REST `POST` / `PUT` / `DELETE` endpoints will NOT be added to `api-server` unless a future requirement explicitly overrides this ADR.**

### 9.4 The read-only `api-server` surface

The full public surface of `api-server` today is exactly these four `GET` endpoints (verified against `cmd/api-server/main.go:73-76`):

1. `GET /api/v1/projects`
2. `GET /api/v1/projects/:id`
3. `GET /api/v1/projects/:id/documents`
4. `GET /api/v1/documents/:id`

Future read-only `GET` endpoints (e.g. `/api/v1/user_stories`, `/api/v1/tasks`) may be added by future REQs WITHOUT requiring an override of this ADR — they remain consistent with the read-only intent.

### 9.5 The MCP write surface

All writes flow through MCP tool families registered at mcp-server boot (verified against `cmd/mcp-server/main.go:73-77`):

1. **`RegisterProjectTools`** — 5 tools: `create_project`, `get_project`, `update_project`, `delete_project`, `list_projects`.
2. **`RegisterDocumentTools`** — 5 tools: `create_document`, `get_document`, `update_document`, `delete_document`, `list_documents`.
3. **`RegisterUserStoryTools`** — 5 tools: `create_user_story`, `get_user_story`, `update_user_story`, `delete_user_story`, `list_user_stories`.
4. **`RegisterTaskTools`** — 5 tools: `create_task`, `get_task`, `update_task`, `delete_task`, `list_tasks`.
5. **`RegisterAuditTools`** — read-side audit trail (included for completeness; it is a read tool, not a write tool).

Any future entity requiring CUD operations adds an `*_tools.go` family at mcp-server, NOT REST endpoints at api-server.

### 9.6 Rationale

1. **Single source of truth for writes.** All write paths go through the MCP tool registry, which means audit-trail invariants (`status_audit_trail` rows produced by `UpdateXxxStatus` inside a transaction), domain-level transition guards (`IsValidTransition`), and transactional consistency (`BeginTx` / `Commit` in `*_repo.UpdateXxxStatus`) are uniformly enforced. A second write surface (REST) would require either re-implementing all of this in handler code OR extracting a shared service layer that does not exist today. Both are work; neither delivers a user-visible benefit absent a concrete browser-direct client.
2. **Sub-agent-first design.** This project's primary client is the team of sub-agents — po-ba, system-architect, tech-lead, tester, be-dev, fe-dev, orchestrator. Sub-agents speak MCP natively (the `mcp-server` is reachable via SSE + JSON-RPC POST, which is the protocol Claude Code's MCP client integration is designed for). Adding REST writes for a hypothetical browser client is premature when no such client exists and the existing FE (`web/`) currently consumes only the 4 `GET` endpoints.
3. **Operational surface area.** Fewer endpoints = fewer surfaces to monitor, fewer auth concerns (today `api-server` is unauthenticated by design because everything it returns is read-only public-by-default; adding writes immediately raises the auth question), smaller CORS attack surface, fewer rate-limit decisions.
4. **Composability of tools.** MCP tools are composable building blocks for the sub-agent team (an agent can chain `list_projects → get_project → update_project → list_user_stories` in a single MCP session). A REST API would not improve this; it would duplicate it.

### 9.7 Alternatives considered

1. **REST writes added to `api-server`.** Rejected per the rationale above. Cost: doubles the write surface; requires audit/transition logic duplication or extraction into a shared service layer; introduces auth/CORS/rate-limit decisions that are not currently load-bearing.
2. **MCP-as-write + REST-as-read with a shared `internal/service/` layer.** Considered but **deferred indefinitely**. Cost: introduces a `services/agent-board/internal/service/` layer that does not exist today; adds an abstraction step with no concrete consumer. Revisit only when at least TWO concrete callers (current REST reads + a new browser-direct CRUD client) coexist.
3. **Bidirectional REST + MCP for the same operations.** Rejected. Same problems as #1 plus the operational-surface concern compounded.
4. **GraphQL.** Not considered — orthogonal to the read/write surface question; would replace BOTH today's REST and MCP, which is out of scope for any single REQ.

### 9.8 Conditions for revisiting this decision

This ADR is revisited ONLY if one of the following becomes true. A REST-writes proposal that does NOT cite at least one of these conditions is rejected at architecture review:

1. **A concrete non-sub-agent browser-direct client requirement.** I.e. a user-facing UI that needs to create / update / delete projects, documents, user stories, or tasks without an MCP proxy. The current `web/` frontend does NOT meet this bar — it currently only reads.
2. **An external integration partner who cannot integrate MCP.** A third party who must POST JSON over REST to integrate with this system.
3. **A measured performance or operational benefit to a REST-write path** that outweighs the duplication / service-layer cost. Measured: with numbers, in a documented bench / production-load report. Not "it would feel cleaner."

**This decision is NOT revisited just because adding a REST endpoint would be technically easy in an isolated PR.** Ease of implementation is not a justification for doubling the write surface.

### 9.9 Consequences

- The `web/` dashboard cannot offer "Create Project" / "Edit Project" / "Delete Project" buttons without a future REQ that explicitly overrides this ADR. Any such future REQ must answer §9.8 conditions in its README's "Decision log."
- `e2e` test fixtures cannot rely on `POST /api/v1/projects` for setup — they must go through MCP tools (this is what REQ005/US008 implemented via the `mcp_keywords.resource` Robot keywords, and what `make e2e-up` enables by including mcp-server in compose per REQ005 D-015).
- Any tech-debt item that proposes "add REST writes" is now closed as `won't-fix` with a pointer to this ADR (US014 AC requires the strike-through in `docs/tech_debt.md` line 98).
- New entities being added to the domain (e.g. a future `Comment` entity) must add an MCP tool family, NOT a REST endpoint.

### 9.10 References

- `docs/tech_debt.md` line 98 (the "MCP-as-write-API" finding being formalised).
- `cmd/api-server/main.go:73-76` (the 4 GET endpoints).
- `cmd/mcp-server/main.go:73-77` (the 4 tool-family registrations).
- REQ005 D-015 (mcp-server in compose — the operational consequence of this ADR for e2e).
- REQ005/US008 (e2e suite-setup rewrite to go through MCP — historical evidence the team has already lived with this ADR's consequences).

---

## 10. Skill / hook usage

### 10.1 BE stories (US001..US012)

- **TDG (Test Doctor / TDD-guide skill).** Mandatory per project convention. Devs write failing test first, prove it fails for the right reason, then implement (US010, US011, US012 only — US001..US009 / US013 are tests-only, so "TDG" reduces to "write the test, run it, see it pass against the existing untouched production code; if it passes without changes that means it was already covered or you wrote the wrong assertion — re-check").
- **Live e2e + 3-clean-run flake check** (REQ005/US008 mandate D-014). Any story that touches `services/<>/` production code (US010, US012) requires a `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` clean run three times in a row before code review can pass. US015 (Makefile only, absorbing the former US011 `PG_CONN ?=` flip) requires the same — it changes `make` behaviour and the regression bar is "the e2e pipeline still works."
- **`govulncheck` clean** (US012 specifically). The bumped toolchain MUST produce zero findings reachable from the project's own code. Tester captures `govulncheck ./...` output in the US012 test report.

### 10.2 FE stories (US013)

- **TDG.** Tests-only story; the "test first" rule reduces to: write the FCT-* assertion, run it, watch it pass against the untouched production component, confirm via coverage report that the previously-uncovered line is now hit.
- **React-doctor (skill `.claude/skills/react-doctor/`).** D-014 from REQ005 makes react-doctor mandatory for FE code touches. US013 does NOT touch production code, so react-doctor's rule-set applies only as a sanity sweep on the new test file (no `dangerouslySetInnerHTML` in test code, no large reducer cascades in test code — neither is a realistic risk).
- **Live e2e + 3-clean-run.** US013 is component-level only; it does not require an e2e run. (If it did, it would also need to hit the 3-clean-run bar.)

### 10.3 Meta / ADR story (US014)

- **No TDG, no live e2e, no react-doctor.** US014 is documentation-only; the deliverable is this §9.
- **Verification ≠ tests.** Tech-lead's verification task reads §9 and checks each AC scenario in `US014_*.md`. The "live e2e" for US014 is **a `grep`-based assertion** as part of the tech-lead review:
  ```
  grep -q "^## 9. ADR-001 — MCP-only-writes is the permanent write API" docs/requirements/REQ006_tech_debt_backfill_sprint/architecture.md
  grep -q "^### 9.3 Decision" docs/requirements/REQ006_tech_debt_backfill_sprint/architecture.md
  grep -q "Conditions for revisiting" docs/requirements/REQ006_tech_debt_backfill_sprint/architecture.md
  ```
  Plus a manual read against the US014 AC. Tester writes US014's spec files (`US014_be_unit_tests.md`, `US014_fe_unit_tests.md`, `US014_e2e_tests.md`) as one-line disclaimers per US014 AC.

### 10.4 What counts as "live e2e" for cluster-1/2/3 (US001..US009)

Cluster-1/2/3 stories do NOT touch production code. The "live e2e" rule from REQ005 D-014 was written for production-code touches. For tests-only stories the equivalent is:

- `cd services/agent-board && go test ./... -race -cover` clean across the whole module (not just the touched package).
- `cd services/agent-board && golangci-lint run ./...` clean.
- 3 consecutive runs of `go test -count=3 ./internal/<touched-package>` clean (catches `-race` flakes).
- Make e2e-up / e2e-run does NOT need to be re-exercised for tests-only stories (the production binaries are byte-identical).

---

## 11. Decision log

Echoes po-ba's D-001..D-005 (verbatim where applicable) then adds D-006..D-013 introduced by this architecture revision.

### D-001 — All-in-one REQ006 (user Q1, verbatim) — locked at intake
Source: po-ba README.
> "**All-in-one REQ006.** One big sprint with ~30 stories mixing BE-test, BE-prod, FE."
Architect impact: stories are scoped per-file (cluster-1/2) per D-005; mixed-track parallelism in Phase 3 implies tech-lead spawning be-dev + fe-dev concurrently.

### D-002 — Defer REST-writes indefinitely (user Q2, verbatim) — locked at intake
Source: po-ba README.
> "**Defer indefinitely.** MCP-only-writes is now formalised as a permanent ADR in REQ006."
Architect impact: §9 ADR-001 written.

### D-003 — Include FE TabSwitcher in REQ006 (user Q3, verbatim) — locked at intake
Source: po-ba README. Implemented as US013.

### D-004 — Cluster-5 split-out of would-be REST stories — locked at intake
Source: po-ba README. No architect override; cluster 5 carries housekeeping + ADR only.

### D-005 — Per-file granularity for cluster-1/2 stories — locked at intake
Source: po-ba README. Each cluster-1/2 story owns ONE source file. Test-only.

### D-006 — `DATABASE_URL` is the sole accepted env var; `DB_URL` is rejected at startup; helper in new `internal/config` package
- **Context:** OQ-1 (env-var handling for `DB_URL` / `DATABASE_URL`) needed a single answer to ship US010. The first draft proposed a precedence rule (accept both, prefer `DATABASE_URL`). The human rejected that during the Phase 1 HARD STOP because of silent-misconfiguration risk: a forgotten `DB_URL` in a `.env` file or a stale Helm value could silently bypass the rename during the migration. The revised decision below is the second pass.
- **Decision:** `DATABASE_URL` is the SOLE accepted env var across both binaries. `DB_URL` is REJECTED at startup with a fatal error and a clear, operator-actionable message (telling them to rename, or to remove `DB_URL` if both are set). There is no precedence rule — there is one variable. Both binaries call a shared `config.ResolveDBURL()` defined in NEW package `services/agent-board/internal/config`. The helper returns `(string, error)`; the caller `main.go` does `log.Fatal(err)` and emits the happy-path log line.
- **Alternatives rejected:**
  - **Precedence rule (`DATABASE_URL` wins, `DB_URL` still accepted).** Rejected by the human during the Phase 1 HARD STOP — silent dual-source-of-truth risk during the migration; partially-migrated environments could silently pick the wrong URL.
  - **`DB_URL` wins / `DB_URL` is the canonical name.** Rejected: less common; `DATABASE_URL` is the 12-factor / Heroku convention; api-server already uses `DATABASE_URL` in production; switching mcp-server is the lower-impact change.
  - **Log-warning-and-continue when `DB_URL` is set.** Rejected for the same silent-misconfig reason as the precedence rule.
  - **Duplicate the helper into both `cmd/*/main.go` files (per REQ005 D-008 precedent).** Rejected: US010 AC requires unit tests for the helper (≥95% per-package coverage), which is cleaner with a package than two duplicate test files; the four-branch behaviour is more substantive than REQ005's 9-line lifecycle glue.
- **Consequences:**
  - One new package added (small, ~30 lines source + ~80 lines tests). Test surface is its own file.
  - `docker-compose.yml` standardises on `DATABASE_URL` for both services. `DB_URL` is no longer set by our infra AND no longer accepted by either binary.
  - **Operator migration impact (explicit and accepted):** any external deployment passing `DB_URL` to mcp-server (Helm charts, custom compose overrides, CI scripts, hand-set shell envs) must rename to `DATABASE_URL`. mcp-server will refuse to start otherwise. There is no quiet upgrade path. This is the deliberate cost of the hard-fail choice.
  - Closes REQ005 architecture OQ-7.
  - **Forces a US010 story-AC reconciliation by po-ba** before Phase 2 spawns for US010 — see §5.8. The three pre-existing AC scenarios that describe dual-acceptance / precedence / `BothSet_PreferredWins` are now stale and must be replaced.

### D-007 — Go toolchain pin: `go 1.26.4`
- **Context:** OQ-2 (Go version that clears the `crypto/x509` finding).
- **Decision:** `go 1.26.4` (with matching `toolchain go1.26.4` directive). `Dockerfile` builder image moves to `golang:1.26-alpine` (minor-tracking tag).
- **Alternatives rejected:**
  - Stay on `go 1.25.x` and wait for a backport. Rejected: 1.25 is past its support window; no fixes backported for these findings.
  - Pin `golang:1.26.4-alpine` exactly. Considered; rejected for maintenance cost — minor-tracking picks up future patches automatically.
  - Bump to `go 1.27`. Rejected: 1.27 may not be released at architecture authoring; if it is, the human can override at re-approval.
- **Consequences:**
  - `services/agent-board/go.mod` line 3 changes + new `toolchain` directive.
  - `services/agent-board/Dockerfile:9` updates.
  - Devs install `go1.26.4` locally OR upgrade system toolchain.
  - Live `govulncheck` post-bump MUST exit clean; tester captures evidence in test report.

### D-008 — ADR location: inline in REQ006 `architecture.md` §9
- **Context:** OQ-3 (separate `docs/adr/` document vs. inline).
- **Decision:** Inline as §9 of this document. NOT a new `docs/adr/` convention.
- **Alternatives rejected:**
  - New `docs/adr/0001-mcp-only-writes.md`. Rejected for one-ADR-in-isolation cost (new convention with one entry is noisy); future ADRs can lift §9 into a `docs/adr/` directory then.
- **Consequences:**
  - Tech-lead's US014 verification is a grep+read on §9 of this file.
  - Future ADRs may live anywhere; this decision is REQ006-local.

### D-009 — US013 coverage target: 80% across stmts / branches / lines / functions
- **Context:** US013 AC specifies 80%; po-ba intake confirmed.
- **Decision:** 80% is the target. Pushing higher is welcome; below 80% blocks review.
- **Alternatives rejected:** 95% (BE-test bar) is unrealistic for component tests; 70% leaves too much surface uncovered.
- **Consequences:** Tester pins the 12 FCT-* IDs in §7.2; coverage is measured via `npm test -- --coverage --collectCoverageFrom=...`.

### D-010 — US008 message.go harness: httptest + Echo for `HandleMessage`, direct call for helpers
- **Context:** OQ-5 (test harness shape for `message.go`).
- **Decision:** Use `httptest.NewRequest` + `httptest.NewRecorder` against an `echo.New()` server with `HandleMessage` mounted. Use real `mcp.SessionManager` + real `mcp.ToolRegistry` (with controllable tool registrations) — NOT a `SessionManager` interface + mock.
- **Alternatives rejected:**
  - Introduce a `SessionManager` interface for mocking. Rejected: would require a production-code change (interface declaration) and the existing handler tests already use the real type — consistency wins.
  - Test helpers (`sendError` / `sendToolResultError`) ONLY via `HandleMessage` routing. Considered; recommended as the default (path-b in §8.3) but path-a (an `internal_test.go` re-export) is acceptable if coverage gaps emerge.
- **Consequences:** US008 AC's 13 test functions map cleanly to httptest invocations plus 2-4 indirect helper coverage assertions.

### D-011 — Cluster-1/2 `≥95% modulo enumerated unreachable lines` carries forward from REQ005/US005
- **Context:** OQ-4 (some lines genuinely unreachable via sqlmock).
- **Decision:** Exemption mechanism in §4.5: tester names each unreachable line in the test report with a one-line justification. Threshold becomes `≥95% modulo enumerated unreachable lines`. Same as REQ005/US005's precedent.
- **Alternatives rejected:** Hard 95% with no exemption — too brittle; the doc-comment-vs-code `ListTools` mismatch and the `defer rollback log.Printf` lines are real and unreachable.
- **Consequences:** Per-story test reports include the exemption block when needed.

### D-012 — `internal/config` is the only NEW production file across REQ006
- **Context:** US010 needs a unit-test surface for the env-var helper.
- **Decision:** `services/agent-board/internal/config/dburl.go` + matching `_test.go` is the ONLY new production-code file across all of REQ006. Every other production file is either edited (US010 cmd/*, US012 go.mod + Dockerfile, US015 Makefile — which now also carries the absorbed US011 `PG_CONN := → ?=` flip per D-014) or untouched (everything else). US015's deletions of `startup.sh`/`shutdown.sh` and its Makefile edits do not introduce any new production-Go-code files.
- **Consequences:** Tech-lead's "no production code change" assertion for US001–US009 + US013 remains accurate; US010 carves out ONE new file for the helper, justified by the test surface.

### D-013 — Retire `startup.sh`/`shutdown.sh`; consolidate dev + docker workflows in the Makefile (US015)
- **Context:** During the Phase 1 HARD STOP for REQ006 architecture rev 2, the human identified that `startup.sh` at repo root (line 16) still passes `DB_URL=...` to `mcp-server`, which would hard-fail once US010's single-var contract (D-006) lands. Rather than spot-patching env-var names in the legacy script, the human directed the team to **retire `startup.sh`/`shutdown.sh` entirely** and make the `Makefile` the single entry point for both local-dev and docker e2e workflows. **Verbatim user direction (recorded for future readers):** *"just ignore startup/shutdown env name, delete it, and enhance make file to work for both docker and dev local for all action like start/stop/migration_db/seed data in us015"*.
- **Decision:** A new story **US015** (cluster 5, housekeeping) was authored by po-ba to carry this work. Architecture impact:
  - Delete `startup.sh` and `shutdown.sh` at repo root.
  - Add four new Makefile targets: `dev-up`, `dev-down`, `dev-migrate`, `dev-seed`.
  - Add a new Makefile variable `DEV_PG_CONN ?= postgres://agent_board:agent_board@localhost:5432/agent_board?sslmode=disable` — distinct variable, distinct workflow from US011's `PG_CONN ?=`.
  - Existing `make e2e-*` family stays byte-identical (additive only).
  - Zero `DB_URL` in any new recipe — validates US010 alignment.
- **Sub-decisions locked from po-ba intake (Q1–Q4):**
  - **Q1 — Postgres source for `dev-*`:** native install only. Operator installs `postgres` on `localhost:5432`; `make dev-up` does NOT start a DB container. Recommended macOS path documented in US015 `Notes for the team`; cross-platform install docs are explicitly out of scope.
  - **Q2 — Naming:** keep the existing `e2e-*` namespace untouched; add a new `dev-*` namespace alongside it. No rename of any existing target — REQ005 docs, CI, tester's e2e harness continue to work without modification.
  - **Q3 — Process lifecycle:** PID files (`.mcp.pid`, `.api.pid`, `.web.pid`) and log files (`mcp-server.log`, `api-server.log`, `web.log`) at repo root, mirroring `startup.sh`'s convention byte-for-byte. Backgrounded native processes. `dev-down` is PID-file-driven with a port-kill fallback against `:8080` / `:8081` / `:3000`. `dev-down` is idempotent (second invocation with no PID files MUST exit 0).
  - **Q4 — Relationship with US011:** US011 (`PG_CONN ?=` for the e2e family) stays an independent story. US015 introduces a separate `DEV_PG_CONN ?=` for the dev family. Same `?=` idiom, different variables, different default ports, different workflows. No hard `Blocked by` link in either direction.
    - **Q4 subsequently superseded by D-014 (merge during HARD STOP rev 4):** US011 was absorbed into US015 as part of the Phase 1 HARD STOP rev-4 round. Both `?=` overrides (`PG_CONN ?=` AND `DEV_PG_CONN ?=`) now live together in US015 (Option A union). The "independent" framing was historically correct at intake; the merge during rev 4 made it redundant. See D-014 for full context.
- **Alternatives rejected:**
  - **(a) Spot-patch the env-var name in `startup.sh` as part of US010.** Rejected by the human at the Phase 1 HARD STOP — leaves a divergent local-dev script with no consolidation benefit; the legacy script remains tech-debt indefinitely.
  - **(b) Make the dev DB switchable between native and docker (a `dev-up-docker` companion target).** Rejected — adds operator knobs without clear demand. The existing `make e2e-up` already provides the dockerised-Postgres path for any dev who wants it.
  - **(c) Rename `e2e-*` targets to `docker-*` for symmetry with `dev-*`.** Rejected — breaks REQ005 docs, CI references, and tester's e2e harness invocations for no proportional gain. The two namespaces are clear enough as `dev-*` vs `e2e-*`.
  - **(d) Merge US011 into US015.** Rejected — loses the per-story granularity REQ006 maintains elsewhere; the two changes can land in any order and serve distinct workflows.
  - **(e) Introduce a process supervisor (`tmux`/`foreman`/`overmind`/`concurrently`/`pm2`) instead of PID files.** Rejected — adds a new dependency; backgrounded native processes with PID files already mirror `startup.sh`'s existing developer muscle memory.
- **Consequences:**
  - **Sequencing with US010:** US015 SHOULD ship before or in the same merge as US010. **Rationale:** US010's hard-fail on `DB_URL` would break the existing `startup.sh` workflow until US015 deletes it. No hard `Blocked by` — tech-lead may pair both PRs in one merge, or sequence US015 → US010 in two PRs.
  - **§3 US010 row** carries an explicit note that `startup.sh`/`shutdown.sh` are NOT in US010's lane — they are deleted by US015. US010's be-dev MUST NOT touch them.
  - **README story count** moves from 14 → 15. Cluster 5 grows from 4 stories to 5. **(Subsequently revised by D-014: US011 absorbed into US015 → story count returns to 14, cluster 5 returns to 4.)**
  - **Sweep impact:** US015 includes documentation + agent-definition references to `startup.sh`/`shutdown.sh` — `README.md`, `tests/e2e/README.md`, `services/agent-board/README.md` (if exists), `.claude/agents/*.md`, `CLAUDE.md`. Plus a `python3 scripts/sync-gemini.py` run after agent-file edits per project rule.
  - **No production-Go-code change.** US015 is Makefile + repo-root scripts + docs + agent-defs only. The D-012 "only `internal/config` is the new production file" claim still holds.
  - **No new architecture OQ raised.** Q1–Q4 were answered by the human at po-ba intake before this revision was written.

### D-014 — US011 absorbed into US015 (merge during Phase 1 HARD STOP rev 4)
- **Context:** User identified during the Phase 1 HARD STOP rev-4 round that US011 (1-line `PG_CONN := → ?=` flip) was redundant with US015 (Makefile consolidation). Both are Makefile env-overridability hygiene; same file, same `?=` pattern. A standalone 1-line story for a Makefile flip that lives next door to four new `dev-*` targets and a new `DEV_PG_CONN ?=` variable is conceptual overlap with negligible discrete value.
- **Decision:** Merge US011 INTO US015. Delete the `US011_makefile_pg_conn_overridable.md` story file. US015 AC now covers BOTH `?=` variables. **Option A (union) selected** — BOTH `PG_CONN ?=` (e2e family, port 15432, default URL byte-identical to the pre-flip value) AND `DEV_PG_CONN ?=` (dev family, port 5432, new in US015) are preserved in the merged Makefile. Option B (drop the e2e override and use only `DEV_PG_CONN`) was rejected because it would re-open tech-debt line 86 (the original US011 ask).
- **Alternatives rejected:**
  - **(a) Keep both stories as authored.** Rejected: accepts the conceptual overlap and ships a 1-line standalone story with negligible discrete value. Maintenance and review overhead of two separate PRs / two separate task threads exceeds the per-story granularity benefit at this size.
  - **(b) Rename US015 to "Makefile consolidation + env-overridability" without merging US011.** Rejected: the current US015 title already covers the consolidation work that the merged scope still fits under; renaming is cosmetic and does not address the underlying duplication.
  - **(c) Option B — drop the e2e `PG_CONN ?=` override and consolidate on `DEV_PG_CONN ?=` only.** Rejected: re-opens tech-debt line 86 (the e2e family loses env-overridability that the original US011 was authored specifically to add). Option A (union) honours both workflows.
- **Consequences:**
  - **Story count drops from 15 → 14.** Cluster 5 drops from 5 → 4 stories (now `US012, US013, US014, US015`).
  - **Architecture §3 US011 row removed.** §3 US015 row expanded to include the existing-line `PG_CONN := → ?=` flip and the `docs/tech_debt.md` line 86 strike-through obligation (previously owned by US011).
  - **D-013 Q4 sub-decision annotated as superseded** by D-014, with a forward reference preserving the original "independent stories" framing as historical record (mirrors po-ba's audit-trail-preserving approach in the README).
  - **The `scripts/review/test/test_makefile_pg_conn.sh` regression test** that the original US011 had optionally proposed is **explicitly out of scope** per the merged US015 AC. The `?=` change is exercised implicitly by `make e2e-up` continuing to work.
  - **No re-routing required for tester or tech-lead.** The merge happened during Phase 1 HARD STOP rev 4, before Phase 2 spawns. Tester will produce one set of spec files for US015 covering both the `dev-*` targets and the `PG_CONN ?=` flip; tech-lead will produce US015 task(s) for the merged scope.
  - **README (po-ba's) and the story file** are already updated by po-ba (story count 15 → 14, cluster 5 row 5 → 4, new README D-007 entry documenting the merge, OQ-6 updated, `US011_makefile_pg_conn_overridable.md` deleted, US015 story file expanded with the absorbed scope). This architecture revision (rev 4) reconciles the architecture-side documents.

---

## 12. Cross-cutting

### 12.1 Config / env vars

- `DATABASE_URL` is the SOLE accepted env var post-REQ006. `DB_URL` is REJECTED at startup with a fatal error (rename-instruction message). No precedence rule; one variable only.
- `PORT` (both binaries), `FRONTEND_URL` (api-server CORS), `NEXT_PUBLIC_API_BASE_URL` (web build-time) — unchanged.
- `PG_CONN` (Makefile, e2e family, port 15432) and `DEV_PG_CONN` (Makefile, dev family, port 5432) are both env-overridable via `?=` (US015, absorbing former US011 per D-014) — unchanged default values.

### 12.2 Logging keys

- US010 introduces the `db config: using ...` log line at startup, before the DB ping. Stay consistent with the existing `log.Printf` calls (no log levels — Go stdlib's `log` package has none).
- No new logging keys introduced for cluster-1/2/3 (tests-only).

### 12.3 Metrics

No new metrics. No metrics infrastructure exists today; out of scope for REQ006.

### 12.4 Error model

Unchanged. `repo.ErrNotFound` remains the only sentinel; MCP tools translate to `"<entity> not found"` strings (fresh errors, not wrapped sentinels — confirmed via reading `*_tools.go` source). Cluster-2 tests assert this distinction; do NOT silently convert to wrapped sentinels.

### 12.5 Observability

Unchanged. REQ005/US008 added container logs via `make e2e-logs`; REQ006 does not extend this.

### 12.6 CORS

Unchanged. REQ005's CORS config on api-server (allowlists `FRONTEND_URL`) remains.

---

## 13. Risks & open questions

### 13.1 Risks

1. **R-1: Bumping Go toolchain to 1.26.4 may surface new transitive dep issues.** Mitigation: US012 AC requires `cd services/agent-board && govulncheck ./... && go test ./... && go build ./...` clean post-bump. If a new finding flares, raise a follow-up story; do not silently widen US012's scope.
2. **R-2: `t.Setenv` on `DATABASE_URL` / `DB_URL` in `dburl_test.go` could leak into parallel tests.** Mitigation: `t.Setenv` (Go 1.17+) auto-restores on test cleanup; the `internal/config` package has no other tests sharing env state. Acceptable.
3. **R-3: 14 stories in parallel may exceed tech-lead's review bandwidth.** Mitigation: per po-ba D-001 the user accepted the trade-off; orchestrator's Phase 3a caps at 2 BE + 2 FE per tick which serialises naturally.
4. **R-4: hard-fail on `DB_URL` will break any external deployment (Helm chart, custom compose override, CI script, hand-set shell env) that has been passing `DB_URL` to mcp-server.** Mitigation: this is intentional and accepted per D-006's revised decision — the rename-instruction error message is operator-actionable, the migration cost is one env-var rename per affected deployment, and the upside is no silent-misconfig risk during the migration. Surfaced in §5.9 explicitly so the orchestrator can flag it in the US010 test report and so the human can re-evaluate at re-approval if the appetite for operator-impact changes. Internal infra: zero residual breakage — `docker-compose.yml` is updated in the same commit.
5. **R-5: The doc-comment lie on `ListTools` ("lexicographic order") going unfixed may confuse future contributors.** Mitigation: US009 AC explicitly flags it in the test report under OQ-4; tech-debt entry added; not silently fixed in US009's tests-only scope.
6. **R-6: If US010 lands before US015, the existing `startup.sh` workflow breaks at the next invocation** with the hard-fail rename-instruction error from §5.4 (because `startup.sh:16` passes `DB_URL` to mcp-server). Mitigation: D-013 explicitly directs sequencing — US015 SHOULD ship before or in the same merge as US010. Tech-lead's call whether to pair both PRs into one merge or sequence US015 → US010 in two PRs. The breakage is loud (operator-actionable startup error), not silent; the recovery is "run `make dev-up` instead of `./startup.sh`" once US015 has landed. Acceptable.

### 13.2 Open questions for the human (re-approval gate)

**None left open by this architecture.** Every OQ from po-ba's README has been resolved:

- OQ-1 → D-006 (`DATABASE_URL` wins; helper in `internal/config`).
- OQ-2 → D-007 (`go 1.26.4`).
- OQ-3 → D-008 (ADR inline in §9).
- OQ-4 → §4.5 + D-011 (exemption mechanism carries forward).
- OQ-5 → D-010 (httptest + Echo for HandleMessage; direct call for helpers).
- US015 sub-decisions Q1–Q4 → D-013 (locked at po-ba intake; no architecture-side OQs raised).
- US011-vs-US015 merge (po-ba README OQ-6 in its local sequence) → D-014 (US011 absorbed into US015, Option A union — both `?=` overrides preserved).

If the human disagrees with any of D-006, D-007, D-008, D-009, D-010, D-011, D-012, D-013, or D-014, please call it out at re-approval — the architecture will be revised in place and an Approval-log entry will record the change.

---

## 14. Approval log

### Revision 1 — 2026-06-03 — author: system-architect
Initial draft. Resolves OQ-1..OQ-5 from po-ba README via D-006..D-012. Inline ADR-001 in §9 (US014 deliverable). File-level touch map covers US001..US014. Cluster-1/2/3 test pattern locked in §4 once for all 9 BE-test stories. `Approval: pending_approval` set.

### Revision 2 — 2026-06-03 — driver: human feedback pass 1
US010 env-var contract revised end-to-end. The human rejected the precedence-rule approach from Revision 1 during the Phase 1 HARD STOP because of silent-misconfiguration risk during the migration. The revised contract is **single-var only**: `DATABASE_URL` is the SOLE accepted env var; `DB_URL` is REJECTED at startup with a fatal, operator-actionable error.

Sections touched:
- **§0.1 bullet #2** — rewritten to describe the single-var contract + hard-fail behaviour; dropped "precedence" language; updated helper API signature; updated log-line examples (single happy-path variant only).
- **§3 US010 touch map** — heading renamed ("Standardise on `DATABASE_URL` (reject `DB_URL`)"); per-row edit descriptions updated to reflect the hard-fail; mcp-server row now explicitly notes this is the meaningful runtime change with breaking-deployment implications; docker-compose row drops any "still accepting both" language; tests/e2e/README.md row updated to reflect rejection; tech_debt.md strike-through note revised.
- **§5** — full rewrite. New shape: decision statement + rationale (5.1), simplified helper API `ResolveDBURL() (string, error)` (5.2), call-site snippets (5.3), four locked error-message strings (5.4), docker-compose change (5.5), four-case unit-test plan (5.6), integration test + optional hard-fail regression test (5.7), **US010 story-AC reconciliation flag for po-ba (5.8)**, operator-facing migration impact (5.9), REQ005 OQ-7 closure note (5.10).
- **§11 D-006** — decision text rewritten. Now records the revised "single var, hard-fail" decision, the human's Phase-1-HARD-STOP rejection of the precedence approach, and the explicit operator-migration consequence. Three alternatives now rejected (precedence rule, `DB_URL`-wins, log-warning-and-continue).
- **§12.1** — cross-cutting config bullet updated to "sole accepted" wording.
- **§13.1 R-4** — risk rewritten. Was "compose change could break local-dev that hand-sets `DB_URL`, mitigated by helper still accepting `DB_URL`"; is now "hard-fail will break any external deployment passing `DB_URL`, mitigation is that this is intentional and accepted with explicit operator-facing surfacing in §5.9."

Sections deliberately untouched (per the revision brief — these are independent of the US010 contract): Clusters 1/2/3 (§4, §6, §8), US011/US012/US013/US014 (§6, §7, §9), FE TabSwitcher (§7), cross-cutting §12.2–12.6 other than §12.1, ADR-001 (§9), D-007..D-012.

**Routing note for orchestrator:** before Phase 2 spawns tester / tech-lead on US010, route to po-ba to revise `US010_harmonise_db_url_env_var.md` per §5.8. The other 13 stories (US001..US009, US011..US014) are unaffected and can proceed independently.

`Approval: pending_approval` re-set. Awaiting human re-approval at HARD STOP.

### Revision 3 — 2026-06-03 — driver: human feedback pass 2 — surfaced US015 scope from startup.sh DB_URL reference
At the Phase 1 HARD STOP for revision 2, the human surfaced that `startup.sh` at repo root still references `DB_URL` and directed the team to retire `startup.sh`/`shutdown.sh` entirely rather than spot-patch the env-var name. **Verbatim user direction:** *"just ignore startup/shutdown env name, delete it, and enhance make file to work for both docker and dev local for all action like start/stop/migration_db/seed data in us015"*. po-ba authored a new story `US015_consolidate_dev_and_docker_workflow_in_makefile.md` with locked sub-decisions Q1–Q4. This architecture revision reconciles to the new 15-story scope.

Sections touched (revision 3):
- **§0 reading guide** — story-count phrasing in the §3 / §11 section pointers updated (`US001..US014` → `US001..US015`; `D-006..D-012` → `D-006..D-013`).
- **§0.1 executive summary** — new bullet 7 for US015 summarising D-013 (delete `startup.sh`/`shutdown.sh`; add `dev-up`/`dev-down`/`dev-migrate`/`dev-seed`; Q1–Q4 sub-decisions; `DEV_PG_CONN ?=` distinct from US011's `PG_CONN ?=`; sequencing with US010). Previous bullet 7 (OQ summary) renumbered to bullet 8 and extended with the US015 OQ-closure note.
- **§1.1 in-scope** — "fourteen stories" → "fifteen stories"; Cluster 5 grown from 4 stories to 5 (US011, US012, US013, US014, US015).
- **§3 heading** — `US001..US014` → `US001..US015`.
- **§3 US010 row** — appended a non-table note that `startup.sh`/`shutdown.sh` are NOT in US010's lane (US015 deletes them). Tech-lead routing for the sequencing decision called out (pair PRs or sequence US015 → US010). No table row changes.
- **§3 US015 row (NEW)** — full touch map per po-ba US015 AC: explicit `DELETE` markers on `startup.sh` and `shutdown.sh`; four new Makefile targets; new `DEV_PG_CONN ?=` variable; existing `e2e-*` family byte-identical; sweep across `README.md`, `tests/e2e/README.md`, `services/agent-board/README.md` (conditional), `.claude/agents/*.md` (with `sync-gemini.py` mandate), `CLAUDE.md` (conditional), `docs/tech_debt.md` (conditional strike-through). Three explanatory notes: namespace coordination with US011, sequencing with US010, no-production-Go-code-change.
- **§11 D-012 consequences** — appended US015's Makefile edit to the list of edited production files; explicitly noted no new Go source files from US015.
- **§11 D-013 (NEW)** — full decision entry covering context (verbatim user direction at HARD STOP), decision (delete scripts + four Makefile targets + `DEV_PG_CONN ?=`), Q1–Q4 sub-decisions locked from po-ba intake, five rejected alternatives (spot-patch, switchable docker/native dev DB, rename `e2e-*` to `docker-*`, merge US011 into US015, process supervisor), and consequences (sequencing, README count bump, sweep impact, no new production-Go-code, no new OQ).
- **§13.1 R-6 (NEW)** — risk entry covering US010-lands-before-US015 sequencing breakage. Mitigation: D-013 sequencing direction; loud not silent breakage; recovery is "run `make dev-up` instead of `./startup.sh`".
- **§13.2 OQ summary** — added US015 Q1–Q4 closure bullet; D-013 added to the disagreement-callout list.

Sections deliberately untouched (per the revision brief):
- **§4** test pattern (clusters 1+2+3) — no impact from US015.
- **§5** US010 contract (decision text, helper API, four error messages, compose change) — untouched. Only the §3 US010 row added a sequencing note; §5's body is unchanged.
- **§6** US012 Go toolchain bump — no impact.
- **§7** US013 TabSwitcher — no impact.
- **§8** US008 message.go harness — no impact.
- **§9** ADR-001 — no impact.
- **§10** skill / hook usage — no impact (US015 follows the same Makefile-only convention as US011; no new hook rules).
- **§12** cross-cutting — no impact (US015 introduces `DEV_PG_CONN` as a Makefile variable, not an env var passed to services; the §12.1 "sole accepted env var" wording for the services already stands).
- **D-006 through D-012** in §11 — untouched (D-006 retains its env-var meaning; D-013 is the new entry for US015).

`Approval: pending_approval` re-set (unchanged from rev 2). Orchestrator should re-surface the executive summary to the human via `AskUserQuestion` for the third re-presentation. The 7-bullet form (rev-1 OQ-resolution shape + rev-2 single-var pivot + rev-3 US015 addition) is in §0.1.

**Routing note for orchestrator (carries forward from rev 2 + adds):** before Phase 2 spawns tester / tech-lead on US010, the rev-2 routing to po-ba for `US010_*.md` AC revision per §5.8 still stands. US015 is already authored by po-ba (`US015_consolidate_dev_and_docker_workflow_in_makefile.md`) and can proceed into Phase 2 without further story-side revision. The other 13 stories (US001..US009, US011..US014) remain unaffected and can proceed independently.

### Revision 4 — 2026-06-03 — driver: human feedback pass 3 — US011 absorbed into US015 (Option A union, both `?=` overrides preserved)
At the Phase 1 HARD STOP for revision 3, the human identified that US011 (1-line `PG_CONN := → ?=` flip) was redundant with US015 (Makefile consolidation). Both are Makefile env-overridability hygiene; same file, same `?=` pattern. po-ba merged US011 INTO US015: the `US011_makefile_pg_conn_overridable.md` story file is deleted; the US015 story file absorbs the `PG_CONN ?=` flip and the `docs/tech_debt.md` line 86 strike-through obligation. **Option A (union) selected** — BOTH `PG_CONN ?=` (e2e, port 15432) AND `DEV_PG_CONN ?=` (dev, port 5432) are preserved in the merged Makefile. Option B (drop the e2e override) was rejected because it would re-open tech-debt line 86. This architecture revision (rev 4) reconciles the architecture-side documents.

Sections touched (revision 4):
- **§0 reading guide** — `D-006..D-013` → `D-006..D-014`.
- **§0.1 executive summary** — bullet 7 (US015) rewritten to explicitly include the absorbed `PG_CONN ?=` flip from former US011 (Option A union); bullet 8 (OQ summary) extended with D-014 closure note.
- **§1.1 in-scope** — "fifteen stories" → "fourteen stories"; cluster 5 enumeration drops US011 (now reads `US012, US013, US014, US015`); US015 description updated to flag absorption of former US011 and the `PG_CONN ?=` flip.
- **§3 file-level touch map** — **US011 row deleted in full.** US015 row expanded with the existing-line `PG_CONN := → ?=` flip (default URL byte-identical) and the merged `docs/tech_debt.md` line 86 strike-through obligation. New note documents Option A union (both `?=` overrides preserved) and the explicit out-of-scope status of the standalone `test_makefile_pg_conn.sh` regression test.
- **§10.1 BE skill / hook usage** — US011 reference replaced with US015 (same `make e2e-*` regression bar still applies; now owned by the merged story).
- **§11 D-012 consequences** — edited-production-files enumeration drops the US011 Makefile reference (US015 already covers Makefile edits; the absorbed scope rolls into the same line).
- **§11 D-013 Q4 sub-decision** — appended forward-reference annotation `Q4 subsequently superseded by D-014 (merge during HARD STOP rev 4)`. Original Q4 text preserved as historical record (mirrors po-ba's audit-trail-preserving approach in the README).
- **§11 D-013 consequences** — README story-count line annotated with "Subsequently revised by D-014: US011 absorbed into US015 → story count returns to 14, cluster 5 returns to 4."
- **§11 D-014 (NEW)** — full decision entry: context (HARD STOP rev-4 user observation of redundancy), decision (merge US011 into US015, Option A union — both `?=` overrides preserved), three rejected alternatives (keep both, rename-only, Option B which would re-open line 86), consequences (story count 15 → 14, cluster 5 5 → 4, §3 US011 row removed and US015 row expanded, D-013 Q4 annotated, regression test out of scope, no re-routing required for downstream agents).
- **§12.1 cross-cutting config** — `PG_CONN` env-overridability line now references US015 (with absorbed-US011 callout via D-014) and mentions BOTH variables (`PG_CONN` and `DEV_PG_CONN`).
- **§13.1 R-3** — "15 stories" → "14 stories".
- **§13.2 OQ summary** — added US011-vs-US015 merge closure bullet pointing at D-014; disagreement-callout list extended to include D-014.

Sections deliberately untouched (per the revision brief):
- **§2** service topology — no impact.
- **§4** test pattern (clusters 1+2+3) — no impact.
- **§5** US010 contract (decision text, helper API, four error messages, compose change, §5.8 reconciliation flag) — no impact.
- **§6** US012 Go toolchain bump — no impact.
- **§7** US013 TabSwitcher — no impact.
- **§8** US008 message.go harness — no impact.
- **§9** ADR-001 — no impact.
- **§10.2 / §10.3 / §10.4** — no impact (US015 already exists under §10.1 via the same `make e2e-*` regression rule that previously named US011).
- **§12.2 through §12.6** — no impact (DEV_PG_CONN remains a Makefile variable; no env var passed to services changes).
- **D-001 through D-013** — left in place except for the annotations to D-013 noted above; D-006..D-012 untouched.

**No invalidation of prior-revision work.** Revision 4 is a targeted documentation reconciliation following the po-ba-side story merge. No spec files, task files, or code have been produced yet (still pre-Phase-2). The other 13 stories (US001..US009, US010, US012..US014) are unaffected by this revision; their content carries forward identically from revisions 1–3.

`Approval: pending_approval` re-set (unchanged from revs 2 and 3). Orchestrator should re-surface the executive summary to the human via `AskUserQuestion` for the fourth re-presentation. The 7-bullet form in §0.1 covers the cumulative state (rev-1 OQ-resolution + rev-2 single-var pivot + rev-3 US015 addition + rev-4 US011 absorption into US015).

**Routing note for orchestrator (carries forward from rev 3 + updates):** before Phase 2 spawns tester / tech-lead on US010, the rev-2 routing to po-ba for `US010_*.md` AC revision per §5.8 still stands. US015 is already authored by po-ba (`US015_consolidate_dev_and_docker_workflow_in_makefile.md`) with the absorbed US011 scope; it can proceed into Phase 2 without further story-side revision. The deleted `US011_makefile_pg_conn_overridable.md` is no longer routable. The other 12 stories (US001..US009, US012..US014) remain unaffected and can proceed independently.

### Revision 5 — 2026-06-04 — driver: human approval
- Approved by human at 2026-06-04T03:14:00Z.
