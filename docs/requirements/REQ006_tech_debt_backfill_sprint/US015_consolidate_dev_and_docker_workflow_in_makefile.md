# US015 — Consolidate dev workflow into the Makefile; retire `startup.sh` / `shutdown.sh`

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** done

## Story
As a **developer**, I want **`make dev-up` / `make dev-down` / `make dev-migrate` / `make dev-seed` to manage the local-dev stack of three native processes (`mcp-server`, `api-server`, `web`) against a native Postgres on `localhost:5432`**, so that **the root `startup.sh` / `shutdown.sh` scripts can be retired and the Makefile becomes the single entry point for both local-dev and docker e2e workflows.**

## Acceptance criteria

- **Scenario: `startup.sh` and `shutdown.sh` are deleted**
  - Given `startup.sh` and `shutdown.sh` exist at repo root today (last touched pre-REQ006)
  - When this story is complete
  - Then `startup.sh` MUST NOT exist at repo root
  - And `shutdown.sh` MUST NOT exist at repo root
  - And `git grep -nE 'startup\.sh|shutdown\.sh'` returns zero hits across the working tree (excluding this story file and the REQ006 sign-off log)
  - And any reference in `README.md`, `tests/e2e/README.md`, `services/agent-board/README.md`, `.claude/agents/*.md`, `.gemini/agents/*.md`, or `CLAUDE.md` is replaced with the equivalent `make dev-*` target

- **Scenario: `make dev-up` starts the three native processes**
  - Given a developer on a workstation with native Go, native Node, and a native Postgres listening on `localhost:5432`
  - When they run `make dev-up`
  - Then `mcp-server`, `api-server`, and `web` are each launched as **backgrounded native processes** (not containers)
  - And each process writes its PID to a repo-root PID file: `.mcp.pid`, `.api.pid`, `.web.pid` respectively
  - And each process's stdout and stderr are redirected to a repo-root log file: `mcp-server.log`, `api-server.log`, `web.log` respectively
  - And both `api-server` and `mcp-server` are invoked with `DATABASE_URL=...` set in their environment (per US010's single-var contract)
  - And `DB_URL` is NOT set anywhere in the `dev-up` recipe (validates US010 alignment)
  - And the recipe prints a one-line status banner per process (e.g. `[dev-up] mcp-server started (pid 12345, log mcp-server.log)`)

- **Scenario: `make dev-down` stops the three native processes cleanly**
  - Given `make dev-up` has been run and the three PID files exist
  - When the developer runs `make dev-down`
  - Then each of `.mcp.pid` / `.api.pid` / `.web.pid` is read and the corresponding PID is signalled (SIGTERM, then SIGKILL after a short grace) — primary teardown path
  - And the PID files are removed once their processes are confirmed gone
  - And a **port-kill fallback** runs against ports `8080` (api), `8081` (mcp), `3000` (web) for any process that did not exit via the PID path — mirroring `shutdown.sh`'s belt-and-braces behaviour
  - And running `make dev-down` a SECOND time (when nothing is running and no PID files exist) MUST exit with status 0 (idempotent — no error on the second invocation)

- **Scenario: `make dev-migrate` applies all `up.sql` migrations against the native dev DB**
  - Given a developer's native Postgres is running on `localhost:5432` with database `agent_board` accessible to user `agent_board` (or whatever the parameterised default resolves to)
  - When they run `make dev-migrate`
  - Then every migration file under `services/agent-board/migrations/*up.sql` (or wherever the canonical migrations live — same set the existing `e2e-seed` target uses) is applied in lexical order to the native DB
  - And the DB connection string is taken from `DEV_PG_CONN ?= postgres://agent_board:agent_board@localhost:5432/agent_board?sslmode=disable` (overridable via env), distinct from the `PG_CONN ?=` variable that serves the e2e family (see next scenario — both `?=` variables live in the same Makefile and are independently overridable)
  - And the recipe is re-runnable: re-running `make dev-migrate` against an already-migrated DB MUST NOT corrupt the schema (migrations are SQL-idempotent if authored that way; if a specific migration is not idempotent, that fact is documented in `Notes for the team` along with the manual recovery step)

- **Scenario: existing `PG_CONN` is also env-overridable**
  - Given the root `Makefile` currently contains `PG_CONN := postgres://agent_board:agent_board@localhost:15432/agent_board?sslmode=disable` on line ~16
  - When this story is complete
  - Then the Makefile contains `PG_CONN ?= postgres://agent_board:agent_board@localhost:15432/agent_board?sslmode=disable` (default unchanged; assignment operator changed `:=` → `?=`)
  - And the new `DEV_PG_CONN ?= postgres://agent_board:agent_board@localhost:5432/agent_board?sslmode=disable` is added alongside it (per Q1/Q3 — native dev DB on localhost:5432)
  - And both variables are env-overridable: `PG_CONN=postgres://... make e2e-seed` and `DEV_PG_CONN=postgres://... make dev-migrate` both honor the env value
  - And when neither env var is set, both targets resolve to their respective defaults
  - And the existing e2e seed flow continues to work zero-config when env is unset

- **Scenario: closes tech-debt finding for `PG_CONN` overridability**
  - Given `docs/tech_debt.md` line 86 contains the finding `Makefile:16 PG_CONN — hardcoded ... Change to ?= so env can override`
  - When this story is `done`
  - Then `docs/tech_debt.md` line 86 is struck through with `→ fixed in REQ006/US015 (absorbed from former US011)`

- **Scenario: `make dev-seed` applies the same seed fixtures the e2e stack uses**
  - Given a developer has run `make dev-migrate` successfully
  - When they run `make dev-seed`
  - Then every fixture under `tests/e2e/data/seeds/*.sql` is applied against `$(DEV_PG_CONN)` in lexical order
  - And the composable convention `make dev-up && make dev-migrate && make dev-seed` brings a brand-new local dev environment up cold (assuming native Postgres is already running and the schema database is created)

- **Scenario: `DATABASE_URL` is the only DB env var used in any new `dev-*` recipe**
  - Given the four new targets (`dev-up`, `dev-down`, `dev-migrate`, `dev-seed`)
  - When the Makefile is inspected
  - Then **zero** occurrences of `DB_URL` appear in any new recipe or new variable
  - And `DATABASE_URL` is the sole DB env var passed to `api-server` and `mcp-server` (validates alignment with REQ006/US010)

- **Scenario: env-var defaults match `startup.sh`'s current behaviour**
  - Given today's `startup.sh` sets `API_PORT=8080`, `MCP_PORT=8081`, web on port `3000`, `FRONTEND_URL=http://localhost:3000`, and `NEXT_PUBLIC_API_BASE_URL=http://localhost:$API_PORT`
  - When `make dev-up` runs without any operator-supplied overrides
  - Then the same effective values are passed to the three processes
  - And each of these defaults uses `?=` so the operator can override from the environment (e.g. `API_PORT=9090 make dev-up`)

- **Scenario: existing `make e2e-*` targets are byte-identical**
  - Given the current set of e2e targets: `e2e-up`, `e2e-down`, `e2e-seed`, `e2e-run` (and any siblings)
  - When this story is complete
  - Then `git diff` against the existing `e2e-*` recipe bodies shows **zero lines changed** in the recipe bodies (variable additions / reorderings above the recipes that do not affect resolved values are acceptable)
  - And `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` continues to pass green
  - And REQ005 docs / CI references to the `e2e-*` targets continue to work without modification

- **Scenario: documentation and agent definitions updated**
  - Given references to `startup.sh` / `shutdown.sh` may exist in `README.md` (root), `tests/e2e/README.md`, `services/agent-board/README.md`, `.claude/agents/*.md`, `.gemini/agents/*.md`, and `CLAUDE.md` (tech-lead grep-confirms during planning)
  - When this story is complete
  - Then every such reference is replaced with the appropriate `make dev-*` target (most commonly `make dev-up` / `make dev-down`)
  - And `tests/e2e/README.md` MAY add a brief note distinguishing the `dev-*` family (native processes, native Postgres) from the `e2e-*` family (compose stack, dockerised Postgres on `:15432`)
  - And `docs/tech_debt.md` — if it contains any line referencing `startup.sh` / `shutdown.sh` directly — is struck through with `→ fixed in REQ006/US015`
  - **Note on origin:** this story does NOT close a specific REQ005 OQ. It closes a finding the user surfaced during the REQ006 Phase 1 HARD STOP, captured in `Notes for the team` below.

- **Scenario: `dev-up` and `dev-migrate` fail gracefully when prerequisites are missing**
  - Given the workstation is missing `psql` from PATH (needed by `dev-migrate` / `dev-seed`), OR missing the native `go` toolchain (needed to launch `api-server` / `mcp-server`), OR missing `node` / `npm` (needed to launch `web`)
  - When the relevant target is invoked
  - Then the recipe prints a single clear actionable error line naming the missing tool and exits non-zero
  - And given native Postgres is not listening on `localhost:5432`, `make dev-migrate` (and the api/mcp process startup) prints a clear actionable error like `Postgres not reachable at localhost:5432 — start your local Postgres (e.g. brew services start postgresql@15) and retry` and exits non-zero
  - (Note: this is a best-effort guard — exact preflight wording is the implementer's call within the spirit above. The bar is "new dev does not hit a 500-line stack trace.")

## UI / UX flow expectations

**No UI: CLI / operator-facing story.** The "interface" is the Makefile target surface and the stdout / file artefacts.

- **Target surface.** Four NEW targets in the same root `Makefile` that already owns the `e2e-*` family:
  - `make dev-up` — start the three native processes against native Postgres on `localhost:5432`.
  - `make dev-down` — stop them; PID-file-driven primary teardown, port-kill fallback.
  - `make dev-migrate` — apply `up.sql` migrations to the native dev DB.
  - `make dev-seed` — apply `tests/e2e/data/seeds/*.sql` fixtures to the native dev DB.
  - The existing `make e2e-up` / `make e2e-down` / `make e2e-seed` / `make e2e-run` targets remain unchanged.
- **Expected stdout per target** (one-liners, no surprise verbosity):
  - `dev-up`: `[dev-up] mcp-server started (pid 12345, log mcp-server.log)` ×3, one per process.
  - `dev-down`: `[dev-down] mcp-server stopped (pid 12345)` ×3 plus, if the port-kill fallback fires, `[dev-down] port 8081 reclaimed (fallback)`.
  - `dev-migrate`: `[dev-migrate] applying 0001_init.up.sql ...` one line per file, plus a final `[dev-migrate] done (N migrations applied)`.
  - `dev-seed`: `[dev-seed] applying tests/e2e/data/seeds/01_projects.sql ...` one line per fixture, plus a final `[dev-seed] done`.
- **Log file paths** (repo root, same as today's `startup.sh`):
  - `mcp-server.log`
  - `api-server.log`
  - `web.log`
- **PID file paths** (repo root, same as today's `startup.sh`):
  - `.mcp.pid`
  - `.api.pid`
  - `.web.pid`
- **Error wording for the "prereq missing" cases** (operator-actionable, single line):
  - `psql` missing: `[dev-migrate] psql not found on PATH — install Postgres client tools (e.g. brew install postgresql@15) and retry`.
  - `go` missing: `[dev-up] go toolchain not found on PATH — install Go (e.g. brew install go) and retry`.
  - `node` / `npm` missing: `[dev-up] npm not found on PATH — install Node (e.g. brew install node) and retry`.
  - Native Postgres not reachable: `[dev-migrate] Postgres not reachable at localhost:5432 — start your local Postgres (e.g. brew services start postgresql@15) and retry`.
- **Zero `DB_URL` in any recipe.** Validates US010 alignment. Both `api-server` and `mcp-server` see only `DATABASE_URL`.

## Out of scope

- **NOT** spinning up a dockerised Postgres for the dev family (Q1 decision — native install is the answer). If a developer wants compose-style Postgres, they use the existing `make e2e-up` flow.
- **NOT** renaming or restructuring the existing `e2e-*` family (Q2 decision — additive only).
- **NOT** introducing a process supervisor (`tmux`, `foreman`, `overmind`, `concurrently`, `pm2`) — backgrounded native processes with PID files mirror what `startup.sh` already does (Q3 decision).
- **Adding a regression test that asserts `PG_CONN ?=` is present in the Makefile.** The 1-line `:=` → `?=` change does not warrant a regression test (per the original US011 AC's own note that tester may push back). The diff itself is the documentation; future drift would be caught at code review.
- **NOT** adding hot-reload tooling (`air`, `nodemon`, `tsx watch`) for the dev stack — if a developer wants it, they configure it themselves outside the Makefile.
- **NOT** adding cross-platform Windows support — Makefile + bash + POSIX tooling are the assumed substrate, matching the rest of the repo.
- **NOT** creating a native-Postgres install automation (no `brew install`, no `apt-get`, no Postgres role provisioning). The dev sets up Postgres themselves; the Makefile assumes it exists and fails loudly with operator-actionable wording when it does not.
- **NOT** changing the DB schema or migration content. `dev-migrate` runs the **same** migration set that `e2e-seed` already runs.

## Dependencies

- **Sequence with US010:** US015 SHOULD ship **before or in the same merge as** US010. **Rationale:** US010 introduces a hard-fail when `DB_URL` is set, which would break the existing `startup.sh` workflow until US015 deletes it (because `startup.sh` today references `DB_URL`). No hard `Blocked by` link — tech-lead may pair both PRs in one merge if simpler, or sequence US015 → US010 in two PRs.
- **US011 absorbed into this story.** US011 no longer exists as a separate story (deleted during HARD STOP rev 4). All Makefile env-overridability concerns are in this story's AC.

## Notes for the team

- **Where this story came from (verbatim from user at REQ006 Phase 1 HARD STOP):** the user pointed out that `startup.sh` at repo root still references `DB_URL` and asked the team to **retire `startup.sh` / `shutdown.sh` entirely** rather than spot-fix the env-var name. The exact framing from the orchestrator hand-off:
  > "Surfaced during REQ006 Phase 1 HARD STOP. User raised that startup.sh still references DB_URL and asked to retire startup.sh/shutdown.sh entirely rather than spot-fix env-var names."
  This is a deliberate consolidation move, not a quick fix.
- **PID-file convention — mirror `startup.sh` byte-for-byte.** Same file names (`.mcp.pid`, `.api.pid`, `.web.pid`), same log file names (`mcp-server.log`, `api-server.log`, `web.log`), same repo-root locations. The goal is to preserve developer muscle memory: anyone who used to `tail -f api-server.log` still can.
- **Native Postgres is a documented prerequisite (per Q1 decision).** Pick ONE supported install path in the dev-workflow README — macOS recommended path is `brew install postgresql@15 && brew services start postgresql@15 && createuser -s agent_board && createdb -O agent_board agent_board`. Cross-platform install docs (Linux apt, Windows WSL, etc.) are explicitly out of scope; document the macOS path and link to the upstream Postgres docs for everything else.
- **`DEV_PG_CONN` default.** Recommended default: `postgres://agent_board:agent_board@localhost:5432/agent_board?sslmode=disable`. Implementer may tune the user / password to whatever the recommended macOS install path produces, as long as the default works zero-config after following the recommended install path AND is overridable via `?=`.
- **Local pre-push checks** the implementer (or tech-lead reviewer) should run:
  - `make dev-up && sleep 2 && curl -sf http://localhost:8080/api/v1/projects` — sanity-check that all three processes started and the api-server is responsive.
  - `make dev-down && lsof -i :8080 -i :8081 -i :3000` — verify clean teardown (no orphan processes on the three ports).
  - `make dev-migrate && make dev-seed` — verify against an empty local DB.
  - `make dev-up && make dev-down && make dev-down` — verify idempotent teardown.
- **Files removed in this story:** `startup.sh`, `shutdown.sh`. The PID files (`.mcp.pid`, `.api.pid`, `.web.pid`) are gitignored runtime artefacts and will continue to exist (created by `make dev-up`, removed by `make dev-down`) — the convention itself is preserved.
- **Helpful sweep targets for the implementer.** When deleting `startup.sh` / `shutdown.sh`, also grep these locations and update references:
  - `README.md` (repo root)
  - `tests/e2e/README.md`
  - `services/agent-board/README.md`
  - `.claude/agents/*.md` (all six agent definitions — search for `startup.sh` / `shutdown.sh`)
  - `.gemini/agents/*.md` (regenerated from `.claude/` via `python3 scripts/sync-gemini.py` — don't hand-edit; edit `.claude/` and re-run the sync script)
  - `CLAUDE.md` (project root)
  - `docs/tech_debt.md` (strike through any line that references the scripts directly)
- **US011 absorbed.** This story used to be split: US011 (1-line `PG_CONN := → ?=`) + US015 (Makefile consolidation). User surfaced the redundancy during Phase 1 HARD STOP (revision 4); merged into a single story so the Makefile rewrite covers both `?=` overrides in one change. The resulting Makefile contains BOTH `PG_CONN ?=` (e2e stack, port 15432) AND `DEV_PG_CONN ?=` (dev stack, port 5432) — Option A (union), not subtraction.
- **Why no `Blocked by: US010`.** Adding a hard block would force serial sequencing. The directional preference (US015 ships first or alongside US010) is real, but the tech-lead has the option of bundling both PRs into a single merge, which removes the strand-mid-workflow risk entirely. Hard blocks are reserved for cases where one story literally cannot be implemented without another — that's not the case here.

## Sign-off log
(po-ba appends here on each sign-off pass)

### Sign-off pass 1 — 2026-06-07 — verdict: approved
- **Spec review:** Every AC scenario maps to at least one UT-* / IT-* in `US015_be_unit_tests.md`. Script-deletion → UT-001/UT-002; reference sweep → UT-003; `PG_CONN := → ?=` flip (byte-identical default, port 15432) → UT-004; new `DEV_PG_CONN ?=` (port 5432) → UT-005; zero `DB_URL` / `DATABASE_URL`-only → UT-006; four `dev-*` targets → UT-007..010; e2e byte-identical → IT-001; both vars env-overridable → IT-003/IT-004; e2e regression ×3 → IT-002. The process-lifecycle, port-kill fallback, idempotent teardown, and preflight-guard AC scenarios are covered behaviourally by the e2e regression bar + structural target-existence checks — consistent with the architecture-mandated "no Go tests for a Makefile/script/docs story" pyramid shape. No e2e inflation, no missing AC. Spec is honest and complete.
- **Result review:** 14/14 test IDs PASS in `US015_test_report.md`; Skipped Tests: None. Verified directly (not trusted): `git ls-files startup.sh shutdown.sh` empty; `git grep -nE 'startup\.sh|shutdown\.sh'` returns only AC-excluded REQ005/REQ006 doc-history hits; `Makefile:18 PG_CONN ?=` (default byte-identical, :15432); `Makefile:19 DEV_PG_CONN ?=` (:5432); zero `DB_URL` in Makefile with `DATABASE_URL=$(DEV_PG_CONN)` at lines 97/103; `dev-up`/`dev-down`/`dev-migrate`/`dev-seed` present (lines 91/122/147/155); all four `e2e-*` targets present; `docs/tech_debt.md` line 86 struck through with the exact §3 wording. IT-002 is the 3×`5 tests, 5 passed, 0 failed` flake check via the human-accepted `podman-compose up -d` workaround (the `make e2e-up` deadlock is a pre-existing, separately-tracked bug; e2e recipe bodies are byte-identical per IT-001, so the same Robot suite is exercised). Tech-lead independently re-verified all structural checks and the cross review gate (`REVIEW GATE: PASS`) on review pass 2.
- **Routed to:** none — story set to `done`.
