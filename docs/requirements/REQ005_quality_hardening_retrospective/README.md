# REQ005 — Quality Hardening Retrospective

**Status:** intake_complete (awaiting architecture revision for US010 inclusion)
**Source:** REQ004 retrospective + `docs/requirements/REQ004_project_detail_page/REQ004_quality_audit.md` (2026-05-30, commit `9a77ed0`)
**Audience:** internal — developers, reviewers, orchestrator. **No end-user-facing UI changes.**

## Business goal

Close the cross-cutting tech debt and tooling gaps that emerged during REQ004's review cycle so that future requirements ship through a quality gate that (a) actually runs end-to-end on every developer's machine, (b) executes e2e tests live instead of dry-run-with-component-coverage substitution, and (c) does not bleed reviewer time into manual workarounds (worktree path edits, hung Jest runs, missing-tool exits).

This requirement is **infrastructure and developer-experience work**. There is no new user-observable behaviour. Stories are scoped to one defect or one debt item each so they can be picked off in parallel and reviewed independently.

## Scope summary

Three groups, nine stories:

### Group A — quality-gate must-fix (audit §3.3 minimum patch)

| US ID | Title | Track hint |
|---|---|---|
| US001 | Fix `printf "--"` bug in `run-gate.sh` (lines 58 + 71) | Tooling (script) |
| US002 | Add `--forceExit` to FE gate's `npm test` invocation | Tooling (script) |
| US003 | Make `gosec` / `govulncheck` soft-warn when missing | Tooling (script) |

### Group B — code-level tech debt

| US ID | Title | Track hint |
|---|---|---|
| US004 | Replace unbounded `context.Background()` at DB-ping sites with `context.WithTimeout` + signal-cancellable lifecycle context | BE |
| US005 | Backfill 14 repo error-branch tests in `internal/repo` (per audit §4.4) | BE (tests) |
| US006 | Harmonise `useProject`, `useProjectDocuments`, `useDocument` on AbortController + signal-threaded `lib/api/` | FE |
| US007 | Move `@testing-library/dom` from `dependencies` to `devDependencies` in `web/package.json` | FE (package hygiene) |
| US010 | Fix React-Doctor baseline regressions (top-3 state/effect + 1 security) | FE |

### Group C — workflow / harness

| US ID | Title | Track hint |
|---|---|---|
| US008 | Live e2e stack-up: `docker-compose.yml` + `Makefile` targets + seeded Postgres + runbook so `robot` runs end-to-end | Tooling (infra) |
| US009 | Fix sub-agent worktrees to branch off local `main` (harness preferred; agent-definition fallback) | Tooling (harness / docs) |

Numbering aligns with the retrospective groups so that anyone reading the audit can map straight to a story file.

## Decision log (locked at intake by po-ba)

The retrospective listed several options. po-ba locked the following defaults to keep the architect's job constrained and to keep total story count down. Architect / tech-lead may push back via `ARCHITECTURE_GAP_FOUND` if anything below is wrong.

- **D-001 (US002 strategy).** FE gate gets `--forceExit` immediately (one-line `run-gate.sh` edit). The actual MSW leak hunt is split into a **separate** story (NOT in this REQ — added as a follow-up tech-debt note below) so US002 stays a fast unblock and is not held hostage to a real handle-leak investigation. **Rationale:** audit §4.3 item 3 already names `--detectOpenHandles` as the eventual-fix path; that work deserves its own scope, its own time-box, and likely an architect look at whether mermaid's lazy-load or react-markdown is the culprit. If the user wants the leak hunt rolled into REQ005, see "Open questions" — easy to add a US010.
- **D-002 (US003 strategy).** Soft-warn `run_check_warn` is preferred over hard-install for `gosec` / `govulncheck`. The `gosec` ruleset is already exercised inside `golangci-lint` v2 per `.golangci.yml`, and the audit explicitly recommends soft-warn. The README will be updated alongside to call out the substitution. **Rationale:** zero loss of coverage, zero new install-time toil for newcomers, one log line documents the skip.
- **D-003 (US004 scope).** Production `context.Background()` call sites in `services/` are exactly two (`cmd/api-server/main.go:52` and `cmd/mcp-server/main.go:37`). All other matches are in `*_test.go` files and are correct (tests legitimately use `context.Background()` for sqlmock / httptest setup). US004 therefore covers ONLY the two production sites plus wiring a single signal-cancellable lifecycle context through `run()` for SIGTERM handling. **Rationale:** keeps the story INVEST-small. A broader "context discipline" sweep can come later if the team wants.
- **D-004 (US005 scope).** The 14 backfill tests from audit §4.4 are split as: 7 `document_repo_test.go` functions + 7 `project_repo_test.go` functions, each named exactly as the audit specifies. Target: push `internal/repo` from 81.5% to ≥95% per-file. **Rationale:** audit gives the exact shopping list; no design work needed.
- **D-005 (US006 scope).** All three FE hooks (`useProject`, `useProjectDocuments`, `useDocument`) must use the AbortController + signal-thread pattern that `useDocument` already exemplifies. `fetchProject` in `web/lib/api/projects.ts` MUST accept a `signal?: AbortSignal` parameter (parity with `fetchDocument` / `fetchProjectDocuments`). No new shared `useFetch<T>` extraction in this story — that's a refactor with its own risk surface. **Rationale:** the audit names the pattern; copy it three times consistently rather than invent a new abstraction in the same story.
- **D-006 (US008 stack-up shape).** Deliverable is `docker-compose.yml` at repo root + a `Makefile` with `e2e-up`, `e2e-down`, `e2e-seed`, `e2e-run` (one Robot invocation), `e2e` (compose of the above), `e2e-logs` targets + a runbook section appended to `tests/e2e/README.md` (or created if missing). Seeded Postgres uses SQL fixtures under `tests/e2e/data/seeds/` (one `*.sql` per REQ that needs distinct seed data). **Rationale:** docker-compose is the lowest-friction reproducible env; Makefile gives both humans and the orchestrator a stable command surface to invoke; SQL fixtures keep seed data version-controlled and reviewable alongside the e2e specs.
- **D-007 (US009 strategy).** Story has TWO acceptance paths: (a) **preferred** — harness change so sub-agent worktrees branch off local `main` HEAD; (b) **fallback** — formalise the "edit canonical paths under `/Users/.../agents-board/docs/requirements/`" workaround in `.claude/agents/{po-ba,system-architect,tech-lead,tester,be-dev,fe-dev}.md`. Either path satisfies the AC. **Rationale:** the architect / tech-lead is in a better position to know whether the harness layer is touchable from this repo; if not, the agent-definition documentation path closes the orchestrator's recurring `git checkout --theirs` pain.
- **D-008 (US010 inclusion).** React-doctor baseline regressions from REQ004 are folded into REQ005 as US010 — top-3 state/effect + 1 security finding only. Remaining 15 lower-severity findings stay as the recorded baseline. Rationale: tech-lead ran a one-off scan, REQ005 is still pre-Phase-2 so scope is cheap, and a separate REQ would re-litigate scope REQ005 already owns.

## Open questions (for the architect / orchestrator to consider)

1. **MSW leak root-cause story.** Should the actual `--detectOpenHandles` investigation be added to REQ005 as US010? po-ba left it OUT to keep US002 fast. If yes, this is a small (5-pt) BE-style investigation story even though the artefact is in `web/`.
2. **US008 docker-compose location.** `docker-compose.yml` at repo root vs under `tests/e2e/docker-compose.yml`? po-ba prefers repo root (single entry point, also useful for local dev) — architect can override.
3. **US009 harness reachability.** Is the worktree-creation harness in scope to edit from this repo, or is it managed elsewhere? If outside-this-repo, US009 collapses to the documentation fallback (path b).

## Anti-scope (NOT in REQ005)

- **No new product features.** No new endpoints, no new pages, no new business logic.
- **No architectural rewrites.** No microservice split, no FE state-management migration, no test-framework swap.
- **No retroactive REQ001-004 feature changes.** REQ004 shipped; this is debt cleanup, not REQ004 revision.
- **No production observability/logging story.** Audit §2.2 flagged `console.error` in `MarkdownErrorBoundary.tsx` and `log.Printf` in handlers — both are acceptable per the audit and out of scope here.

## Audit reference

- Full audit: `/Users/a667282/workspace/agents-board/docs/requirements/REQ004_project_detail_page/REQ004_quality_audit.md`
- Specifically: §3.2 (gate defects), §3.3 (minimum patches), §4.3 (eventual debt list), §4.4 (the 14-test repo backfill shopping list).

## Phase 2 task table

Authored 2026-06-02 by tech-lead from approved architecture (Rev 3). Sort: US ID ascending, BE-first within a story.

| Story | Task file | Track | Service | Blocked by |
|---|---|---|---|---|
| US001 | `US001_be_fix_printf_double_dash.md` | BE | services/agent-board | none |
| US002 | `US002_fe_force_exit_jest_in_gate.md` | FE | — | none |
| US003 | `US003_be_softwarn_missing_gosec_govulncheck.md` | BE | services/agent-board | none |
| US004 | `US004_be_api_server_lifecycle_context.md` | BE | services/agent-board | none |
| US004 | `US004_be_mcp_server_lifecycle_context.md` | BE | services/agent-board | none |
| US005 | `US005_be_document_repo_error_tests.md` | BE | services/agent-board | none |
| US005 | `US005_be_project_repo_error_tests.md` | BE | services/agent-board | none |
| US006 | `US006_fe_harmonise_hooks_on_abortcontroller.md` | FE | — | none |
| US007 | `US007_fe_move_testing_library_dom_to_devdeps.md` | FE | — | none |
| US008 | `US008_be_live_e2e_stack_up.md` | BE | services/agent-board | none |
| US009 | `US009_be_canonical_path_policy_in_agent_defs.md` | BE | (meta — agent definitions) | none |
| US010 | `US010_fe_mermaid_diagram_ref_attach.md` | FE | — | `US006_fe_harmonise_hooks_on_abortcontroller.md` |
| US010 | `US010_fe_use_document_reducer_and_documents_tab.md` | FE | — | `US006_fe_harmonise_hooks_on_abortcontroller.md` |

**Totals:** 13 tasks (9 BE, 4 FE). Per-story: US001×1, US002×1, US003×1, US004×2, US005×2, US006×1, US007×1, US008×1, US009×1, US010×2.

**Dependency graph (small):**
- `US010_fe_mermaid_diagram_ref_attach.md` ⇐ `US006_fe_harmonise_hooks_on_abortcontroller.md`
- `US010_fe_use_document_reducer_and_documents_tab.md` ⇐ `US006_fe_harmonise_hooks_on_abortcontroller.md`
- All other tasks have `Blocked by: none` and can be picked at any time their track has capacity.

**File-overlap warning for the orchestrator's 3a tick:**
- `scripts/review/run-gate.sh` is touched by US001, US002, US003 (no `Blocked by:` links between them, but co-picking will produce a merge conflict at worktree-integration time and the loser will be re-queued). The orchestrator should serialise these three at queue-build time even though the task files declare no dependency.
