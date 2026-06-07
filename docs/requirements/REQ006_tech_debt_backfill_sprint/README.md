# REQ006 — Tech-Debt Backfill Sprint

**Status:** draft (intake complete; awaiting `system-architect` → `architecture.md`)
**Source:** `docs/tech_debt.md` (post-REQ005 retrospective additions, 2026-06-03)
**Audience:** internal — developers, reviewers, orchestrator. **No new end-user-facing features.**
**Track mix:** BE-test + BE-prod + FE + meta/ADR (all-in-one mixed-track sprint, per D-001).

## Business goal

Pay down the project-wide coverage and consistency debt surfaced by REQ005's project-wide audit (`docs/tech_debt.md` §"Pre-existing happy-path-bias coverage gaps across REQ001–REQ004") in a single coordinated sprint. The debt is the same root cause REQ005/US005 fixed for `document_repo` and `project_repo` — error-branch under-testing — but spread across `internal/repo/`, `internal/handler/*_tools.go`, `internal/mcp/server.go`, plus a handful of small consistency / hygiene items (env-var name, Makefile `PG_CONN`, Go toolchain bump, FE `TabSwitcher` coverage). One small ADR formalises a decision the team has already made implicitly (MCP-only-writes).

This is **infrastructure and quality work**. There is no new user-observable behaviour. Stories are scoped one-file or one-concern each so they can be picked off in parallel and reviewed independently.

## Scope summary

Five clusters, fourteen stories.

### Cluster 1 — `internal/repo/` error-branch backfill (BE-test only)

| US ID | Title | Track | Service |
|---|---|---|---|
| US001 | `task_repo.go` — backfill error-branch tests for 5 functions | BE-test | services/agent-board |
| US002 | `user_story_repo.go` — backfill error-branch tests for 5 functions | BE-test | services/agent-board |
| US003 | `audit_repo.go` — backfill error-branch tests for `getAuditTrail` | BE-test | services/agent-board |

### Cluster 2 — `internal/handler/*_tools.go` error-mapping backfill (BE-test only)

| US ID | Title | Track | Service |
|---|---|---|---|
| US004 | `project_tools.go` — IT-* error-mapping backfill (6 sub-95% functions incl. `RegisterProjectTools` at 0%) | BE-test | services/agent-board |
| US005 | `document_tools.go` — IT-* error-mapping backfill (`RegisterDocumentTools` + sub-95% siblings) | BE-test | services/agent-board |
| US006 | `task_tools.go` — IT-* error-mapping backfill (`RegisterTaskTools` 67.4%) | BE-test | services/agent-board |
| US007 | `user_story_tools.go` — IT-* error-mapping backfill (`RegisterUserStoryTools` 63.5%) | BE-test | services/agent-board |
| US008 | `message.go` — IT-* backfill for `HandleMessage`/`sendError`/`sendToolResultError` | BE-test | services/agent-board |

### Cluster 3 — `internal/mcp/server.go` ToolRegistry + Session message methods (BE-test only)

| US ID | Title | Track | Service |
|---|---|---|---|
| US009 | `server.go` — UT-* backfill for `ToolRegistry` family + `QueueMessage`/`ReceiveMessage` + `RemoveSession` | BE-test | services/agent-board |

### Cluster 4 — Env-var harmonisation (BE-prod)

| US ID | Title | Track | Service |
|---|---|---|---|
| US010 | Harmonise `DB_URL` / `DATABASE_URL` across `mcp-server` and `api-server` (both binaries accept both names; one preferred; explicit startup log line) | BE-prod | services/agent-board |

### Cluster 5 — Housekeeping + ADR (small)

| US ID | Title | Track | Service |
|---|---|---|---|
| US012 | Go toolchain bump to fix transitive `crypto/x509` govulncheck finding | BE-prod (toolchain) | services/agent-board |
| US013 | FE `TabSwitcher.tsx` coverage backfill (tab change handler, keyboard nav, aria-selected, default-tab fallback, prop-driven override) | FE | web/ |
| US014 | ADR — formalise MCP-only-writes as the permanent write API; close related tech-debt items as `won't-fix` | BE-meta (docs) | architecture.md |
| US015 | Consolidate dev workflow into Makefile (`make dev-up/down/migrate/seed`) — absorbs former US011's `PG_CONN ?=` change; retire `startup.sh`/`shutdown.sh` | BE-prod (tooling) | repo root |

**Totals:** 14 stories — 9 BE-test (US001–US009), 3 BE-prod (US010, US012, US015), 1 FE (US013), 1 meta/ADR (US014).

## Decision log (locked at intake by po-ba — verbatim from user answers to clarifying questions)

The user was asked three clarifying questions before this README was written. Answers are recorded verbatim and locked. Architect / tech-lead may push back via `ARCHITECTURE_GAP_FOUND` if anything below is wrong, but **D-001..D-003 are user-direct decisions** and require a fresh user round-trip to overturn.

- **D-001 (sprint size and track mix — user Q1, verbatim).** "**All-in-one REQ006.** One big sprint with ~30 stories mixing BE-test, BE-prod, FE. User accepts the trade-off that cluster 4/5 architect input may slow the bulk work." po-ba ended at 14 stories rather than 30 by giving each cluster-1/2 file a single story (one per source file) instead of splitting per function. Rationale: per-file stories are still INVEST-small, easier to review as a unit, and keep BE/FE devs from thrashing across micro-stories. If the architect / tech-lead wants finer splits later, easy follow-up.
- **D-002 (REST-writes decision — user Q2, verbatim).** "**Defer indefinitely.** MCP-only-writes is now formalised as a permanent ADR in REQ006. The api-server stays read-only by design; document the architectural intent and close the relevant tech_debt items as `won't-fix`. The architect will write the ADR section during Phase 1 architecture step." This becomes **US014** (po-ba writes the AC; architect authors the ADR text in `architecture.md`).
- **D-003 (FE TabSwitcher inclusion — user Q3, verbatim).** "**Include in REQ006 as one FE story.** REQ006 is mixed-track (BE + FE)." This becomes **US013** (single FE story).
- **D-004 (cluster-5 split-out of the would-be REST decision).** Cluster-5 was originally framed as "housekeeping + REST decision." Per D-002 the REST decision is closed (deferred / `won't-fix`). Cluster 5 now carries (a) three small housekeeping items (Makefile / Go toolchain / FE TabSwitcher) and (b) one ADR-writing story (US014). The "would-be REST endpoint" stories are NOT in REQ006.
- **D-005 (per-file granularity for cluster-1/2 backfills).** Each cluster-1 / cluster-2 story owns ONE source file (`task_repo.go`, `user_story_repo.go`, `audit_repo.go`, `project_tools.go`, `document_tools.go`, `task_tools.go`, `user_story_tools.go`, `message.go`). Per-file stories follow REQ005/US005's pattern: AC enumerates the verbatim test-function names that must land, target ≥95% per-file statement coverage, tests-only (no production code changes). Rationale: matches REQ005/US005's proven cadence; tester / be-dev know the shape; review is mechanical.
- **D-006 (US015 added at Phase 1 HARD STOP — Makefile consolidation, retire `startup.sh`/`shutdown.sh`).** Surfaced by the user during the REQ006 Phase 1 HARD STOP: the existing `startup.sh` at repo root still references the deprecated `DB_URL` env var. Rather than spot-fix the env-var name, the user directed the team to retire `startup.sh`/`shutdown.sh` entirely and consolidate the local-dev workflow into the Makefile as a new `dev-*` target family — additive only (existing `e2e-*` family unchanged). Locked sub-decisions: (Q1) native Postgres install only — no dev-Postgres container; (Q2) keep `e2e-*` + add `dev-*` — two namespaces in the same Makefile, no breaking rename; (Q3) PID-file + log-file lifecycle mirroring `startup.sh` byte-for-byte (`.mcp.pid`/`.api.pid`/`.web.pid`, `mcp-server.log`/`api-server.log`/`web.log` at repo root); (Q4) US015 was originally independent of US011 (different Makefile variables: `DEV_PG_CONN` vs `PG_CONN`) — **subsequently overridden by D-007 below: US011 was absorbed into US015 at HARD STOP rev 4.** US015 SHOULD ship before or in the same merge as US010 to avoid stranding `startup.sh` mid-workflow once US010's `DB_URL` hard-fail lands.
- **D-007 (US011 absorbed into US015 at HARD STOP rev 4).** Surfaced during Phase 1 HARD STOP revision 4. User identified US011 (`PG_CONN ?=`, 1-line) as redundant with US015 (Makefile consolidation): both stories operate on the same root `Makefile`, both introduce a `?=` env override, and shipping them as two PRs would mean two touches of the same file with no review benefit. Merge direction: **US015 absorbs US011; US011 file deleted.** **Option A (union)** chosen — BOTH `PG_CONN ?=` (e2e stack, port 15432) and `DEV_PG_CONN ?=` (dev stack, port 5432) env overrides preserved; the e2e stack's host-Postgres-path runbook (tech-debt line 86) remains fixed via the absorbed AC scenario. The merged US015 now owns both: the dev-workflow consolidation AND the `PG_CONN := → PG_CONN ?=` one-line flip on Makefile line ~16. The Q4 clause in D-006 ("US015 is independent of US011") is **superseded by this decision**.

## Open questions (for the system-architect to resolve in `architecture.md`)

The architect should answer each of these in the `architecture.md` they author for REQ006. po-ba flags them here so they are not missed.

1. **OQ-1 (US010 env-var precedence).** Which name wins when **both** `DB_URL` and `DATABASE_URL` are set in the environment? po-ba's intuition: `DATABASE_URL` (the more-widely-used Heroku/12-factor name) wins, with a startup `log.Printf("warning: both DB_URL and DATABASE_URL set — using DATABASE_URL")`. But the architect owns the precedence call. The chosen rule must be documented in `architecture.md` AND in the startup log line the AC requires.
2. **OQ-2 (US012 Go toolchain target version).** `services/agent-board/go.mod` is currently on `go 1.25.0`. The govulncheck finding is on stdlib `crypto/x509` (transitive). Which Go minor version contains the fix? Architect picks the lowest version that is (a) released, (b) supported, (c) `govulncheck`-clean. Story AC says "the next minor that fixes it"; architect picks the exact number.
3. **OQ-3 (US014 ADR scope and location).** ADR will be written by the architect in `architecture.md` as a dedicated `## ADR — MCP-only-writes` section (or whatever the architect's chosen ADR convention is for REQ006). Architect decides: separate ADR document under `docs/adr/`, or inline in `REQ006/architecture.md`. po-ba's preference: inline in `REQ006/architecture.md` since REQ006 is the requirement that formalises the decision; architect can override.
4. **OQ-4 (cluster-1/2 `≥95%` coverage threshold — same nuance as REQ005/US005).** Some lines genuinely cannot be reached by `sqlmock` (rare). If the architect / tech-lead spots one during planning, the AC's `≥95%` becomes `≥95% modulo enumerated unreachable lines documented in the test report`. po-ba is fine with this; flag it here so the tester does not get blindsided.
5. **OQ-5 (US008 `message.go` test harness shape).** `HandleMessage` runs inside Echo; testing the error-routing paths (`sendError` / `sendToolResultError`) needs either Echo's `httptest` recorder or a mock `mcp.Session`. Architect should call out the preferred test shape in `architecture.md`'s "FE/BE test contract" section so the tester does not re-invent it.
6. **OQ-6 (US015 Makefile-consolidation §3 touch map).** US015 was added at Phase 1 HARD STOP (see D-006) and subsequently absorbed US011 at HARD STOP rev 4 (see D-007). The architect must (a) add a new §3 touch-map entry for US015 enumerating the four new `dev-*` targets, the new `DEV_PG_CONN ?=` variable AND the absorbed `PG_CONN := → PG_CONN ?=` change on line ~16, and the two file deletions (`startup.sh`, `shutdown.sh`); (b) flag the `startup.sh` / `shutdown.sh` removal from US010's §3 row so it does not get re-edited by US010's be-dev; (c) remove the former US011 row from §3 entirely (it no longer exists as a story); and (d) cross-check the §0.1 executive summary still reflects reality with the **14-story** scope. po-ba flags this so the architect does not miss it during the architecture revision pass that follows this intake update.

## Anti-scope (NOT in REQ006)

- **No new REST endpoints on `api-server`.** Per D-002, api-server stays read-only (4 GET endpoints). Adding write endpoints requires a new REQ.
- **No new product features.** No new pages, no new business logic, no new MCP tools.
- **No microservice split.** `services/agent-board/` stays one Go module.
- **No FE state-management migration.** No introduction of Redux / Zustand / React Query.
- **No test-framework swap.** Continue with Go `testing` + `testify` + `sqlmock` (BE), Jest + RTL + MSW (FE), Robot Framework (e2e).
- **No production code changes in cluster-1/2 stories.** Every cluster-1 (US001–US003) and cluster-2 (US004–US008) story is **tests-only**. If a backfill test surfaces an actual bug (e.g. a missing `rows.Err()` check), raise it as a new follow-up story — do not silently fix in the test-only story.
- **No retroactive REQ001-005 feature changes.** REQ001-005 shipped; this is debt cleanup, not REQ001-005 revision.
- **No e2e additions.** REQ006 is unit/integration-test-pyramid backfill on the BE side and component-test backfill on the FE side. No new `.robot` files.

## Audit reference

- `/Users/a667282/workspace/agents-board/docs/tech_debt.md` §"Pre-existing happy-path-bias coverage gaps across REQ001–REQ004" (the 23 sub-threshold items mapped to clusters 1–3).
- `/Users/a667282/workspace/agents-board/docs/tech_debt.md` §"US008 follow-up" (the 2 cluster-4 items — env-var harmonisation and the MCP-only-writes architecture note).
- `/Users/a667282/workspace/agents-board/docs/tech_debt.md` line 28 (the cluster-5 Go-toolchain govulncheck finding).
- `/Users/a667282/workspace/agents-board/docs/tech_debt.md` line 86 (the cluster-5 Makefile `PG_CONN` finding).
- `/Users/a667282/workspace/agents-board/docs/tech_debt.md` line 75 (the cluster-5 `TabSwitcher.tsx` finding).
- REQ005/US005 reference pattern: `/Users/a667282/workspace/agents-board/docs/requirements/REQ005_quality_hardening_retrospective/US005_backfill_repo_error_branch_tests.md` — cluster-1/2 stories explicitly mirror this story's shape.

## Phase 2 task table

Populated by tech-lead 2026-06-04 after `architecture.md` rev 5 was approved (human, 2026-06-04T03:14:00Z). One task per story (the 14-story scope after D-014 absorbed US011 into US015). No `Blocked by` links between tasks — every cluster-1/2/3 story is independent (the architecture's API contract is the shared frozen surface), and cluster-4/5 stories have only a soft sequencing recommendation (US015 SHOULD ship before US010) which is documented in the relevant task `## Notes` rather than encoded as a hard `Blocked by`. This maximises Phase 3a parallelism.

| Task file | Story | Track | Service | Title (short) | Blocked by | Status |
|---|---|---|---|---|---|---|
| `US001_be_task_repo_error_tests.md` | US001 | BE | services/agent-board | Backfill `task_repo.go` error-branch tests (12 cases, ≥95% per-file) | none | pending |
| `US002_be_user_story_repo_error_tests.md` | US002 | BE | services/agent-board | Backfill `user_story_repo.go` error-branch tests (12 cases, ≥95%) | none | pending |
| `US003_be_audit_repo_error_tests.md` | US003 | BE | services/agent-board | Backfill `audit_repo.go` error-branch tests (6 cases, ≥95%) | none | pending |
| `US004_be_project_tools_error_mapping_tests.md` | US004 | BE | services/agent-board | Backfill `project_tools.go` IT-* error-mapping tests (18 cases, ≥95%) | none | pending |
| `US005_be_document_tools_error_mapping_tests.md` | US005 | BE | services/agent-board | Backfill `document_tools.go` IT-* error-mapping tests (20 cases, ≥95%) | none | pending |
| `US006_be_task_tools_error_mapping_tests.md` | US006 | BE | services/agent-board | Backfill `task_tools.go` IT-* error-mapping tests (25 cases, ≥95%) | none | pending |
| `US007_be_user_story_tools_error_mapping_tests.md` | US007 | BE | services/agent-board | Backfill `user_story_tools.go` IT-* error-mapping tests (27 cases, ≥95%) | none | pending |
| `US008_be_message_handler_error_tests.md` | US008 | BE | services/agent-board | Create `message_test.go` with httptest+Echo harness (13 cases, ≥95%) | none | pending |
| `US009_be_mcp_server_toolregistry_tests.md` | US009 | BE | services/agent-board | Create `server_test.go` for ToolRegistry + Session messaging (15 cases, race-clean) | none | pending |
| `US010_be_resolve_dburl_helper_and_main_wiring.md` | US010 | BE | services/agent-board | New `internal/config.ResolveDBURL` + main wiring + docker-compose rename | none (soft: SHOULD follow US015) | pending |
| `US012_be_go_toolchain_bump_1_26_4.md` | US012 | BE | services/agent-board | Bump Go toolchain to 1.26.4; govulncheck clean; Dockerfile builder bump | none | pending |
| `US013_fe_tabswitcher_coverage_backfill.md` | US013 | FE | (web) | Backfill `TabSwitcher.test.tsx` to ≥80% (12 FCT-* cases) | none | pending |
| `US014_be_adr_verification_and_tech_debt_strikethrough.md` | US014 | BE (meta/docs) | services/agent-board (nominal) | Verify `architecture.md` §9 ADR-001; strike `docs/tech_debt.md` line 98 | none | pending |
| `US015_be_makefile_consolidation_and_retire_startup_scripts.md` | US015 | BE | services/agent-board (nominal — repo root + Makefile + docs) | Retire `startup.sh`/`shutdown.sh`; add four `make dev-*`; flip `PG_CONN := → ?=`; add `DEV_PG_CONN ?=`; doc + agent-def sweep | none (soft: SHOULD precede US010) | completed |

**Track summary:** 9 BE-test (US001..US009), 3 BE-prod (US010, US012, US015), 1 FE (US013), 1 BE-meta/docs (US014). **Total: 14 tasks** — one per story. No cross-track pairs (REQ006 adds no API endpoints; FE and BE do not meet at any contract).

**Dependency graph:** None inside the graph itself — every task is `Blocked by: none`. Two soft sequencing recommendations live in the task `## Notes`:
- **US015 SHOULD ship before US010** (architecture D-013 + D-014 + R-6). Once US010's hard-fail on `DB_URL` lands, the legacy `startup.sh` workflow breaks loudly. If US015 lands first, this is a non-event. Tech-lead's call whether to pair both PRs in one merge or sequence in two PRs.
- **US010 story-AC reconciliation flag** (architecture §5.8). The current `US010_harmonise_db_url_env_var.md` AC is STALE per the rev-2 D-006 revision: scenarios phrased as "both binaries accept both env-var names", references to a "precedence rule", and the `TestResolveDBURL_BothSet_PreferredWins` unit-test scenario must be replaced with the hard-fail semantics (§5.4 / §5.6). Architecture is authoritative — the dev follows architecture, not the stale story AC — but po-ba SHOULD revise the story file before sign-off (Phase 3d) so the AC the orchestrator checks against is current. **Tech-lead surfaces this to the orchestrator as a route-to-po-ba flag** (see end-of-Phase-2 report).

**Parallelism budget (Phase 3a):** every task front is independent. Orchestrator can fan out aggressively — up to its 2 BE + 2 FE per-tick cap — with no inter-task collisions. The only single-writer files in REQ006 are `docs/tech_debt.md` (touched by US010, US012, US013, US014, US015) and `Makefile` (US015 only). `docs/tech_debt.md` is a low-collision target because each task touches a different line, but if two tasks land simultaneously and merge-conflict on `docs/tech_debt.md`, the loser is re-queued and re-applies its strike-through trivially (worktree isolation policy).

