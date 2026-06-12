# US034 — Standardise on `DATABASE_URL`; reject `DB_URL` at startup

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** done

## Story
As an **operator deploying both `api-server` and `mcp-server`**, I want **both binaries to read a single `DATABASE_URL` env var and refuse to start with a clear, actionable error if the deprecated `DB_URL` env var is set**, so that there is one obvious source of truth for the DB connection string and partially-migrated environments fail loudly at startup instead of silently using the wrong source.

## Acceptance criteria

- **Scenario: both binaries accept only `DATABASE_URL`**
  - Given `services/agent-board/cmd/api-server/main.go` currently reads `DATABASE_URL` only (line 45)
  - And `services/agent-board/cmd/mcp-server/main.go` currently reads `DB_URL` only (line 30)
  - When this story is complete
  - Then **both** binaries (`api-server` and `mcp-server`) start successfully when **only** `DATABASE_URL` is set to a valid Postgres URL (with `DB_URL` unset)
  - And **both** binaries refuse to start (log a fatal error and exit non-zero) when only `DB_URL` is set, with the exact message `"DB_URL is no longer supported; rename to DATABASE_URL (REQ006/US034)"` (locked in `architecture.md` §5.4)
  - And **both** binaries refuse to start (log a fatal error and exit non-zero) when **neither** env var is set, with the exact message `"DATABASE_URL environment variable is required"` (locked in `architecture.md` §5.4)
  - And the api-server vs mcp-server diff at the env-resolution call site is byte-identical (`config.ResolveDBURL()` + `log.Fatal(err)` + happy-path log line)

- **Scenario: deprecated `DB_URL` is rejected at startup with a clear actionable error**
  - Given an operator deploys a binary with only `DB_URL` set (the legacy shape, common in unmigrated `.env` files or Helm charts)
  - When the binary starts
  - Then it MUST exit non-zero **before any DB ping attempt**
  - And the fatal log line MUST contain the substring `"DB_URL is no longer supported"` AND the substring `"rename to DATABASE_URL"` AND the substring `"REQ006/US034"` (so a grep finds the migration ticket)
  - Given the same deployment but with BOTH `DB_URL` AND `DATABASE_URL` set
  - When the binary starts
  - Then it MUST exit non-zero **before any DB ping attempt**
  - And the fatal log line MUST contain the substring `"remove DB_URL"` AND the substring `"DATABASE_URL is the sole accepted name"` AND the substring `"REQ006/US034"`
  - **Exact wording (per `architecture.md` §5.4):**
    - DB_URL-only case: `"DB_URL is no longer supported; rename to DATABASE_URL (REQ006/US034)"`
    - Both-set case: `"DB_URL is set but no longer supported; remove DB_URL from your environment to disambiguate (DATABASE_URL is the sole accepted name as of REQ006/US034)"`

- **Scenario: startup log line names the resolved env var (single-var case)**
  - Given only `DATABASE_URL` is set (the happy path)
  - When the binary starts
  - Then a startup log line is emitted matching the literal shape `"db config: using DATABASE_URL"` (locked in `architecture.md` §5.4 / §5.3; emitted from `main.go`, not from the helper)
  - And the log line is emitted **before** the DB ping attempt (so a failed ping logs the chosen var first, easing debug)
  - And there is exactly ONE happy-path log-line shape (the previous "using DB_URL" variant is dropped because `DB_URL` is no longer accepted)

- **Scenario: existing happy-path behaviour is preserved (api-server) and intentionally broken (mcp-server passing `DB_URL`)**
  - Given a deployment that sets `DATABASE_URL` (the new uniform shape across both binaries)
  - When either binary starts
  - Then it behaves identically to today's `api-server` behaviour (same connection, same ping, same startup sequence)
  - Given a deployment that sets only `DB_URL` (the legacy `mcp-server` shape from REQ005 D-009)
  - When the mcp-server binary starts
  - Then it now hard-fails at startup with the §5.4 rename-instruction error — this is the deliberate breaking change introduced by REQ006/US034 per `architecture.md` §5.1 / D-006
  - And the required compose/Helm/env-file updates that accompany this story are listed in `architecture.md` §5.5 + §5.9

- **Scenario: env-var resolution is unit-tested**
  - Given the new shared helper `config.ResolveDBURL() (url string, err error)` in package `services/agent-board/internal/config` (file `dburl.go` — locked in `architecture.md` §5.2)
  - When the helper is invoked with various env states
  - Then unit tests in `services/agent-board/internal/config/dburl_test.go` cover all four env-state combinations:
    1. `TestResolveDBURL_OnlyDatabaseURLSet` — happy path: `DATABASE_URL` set, `DB_URL` unset → returns the value, nil error
    2. `TestResolveDBURL_OnlyDBURLSet_Errors` — `DB_URL` set, `DATABASE_URL` unset → returns `("", err)` where `err.Error()` matches the §5.4 wording `"DB_URL is no longer supported; rename to DATABASE_URL (REQ006/US034)"`
    3. `TestResolveDBURL_BothSet_Errors` — both set → returns `("", err)` where `err.Error()` matches the §5.4 wording `"DB_URL is set but no longer supported; remove DB_URL from your environment to disambiguate (DATABASE_URL is the sole accepted name as of REQ006/US034)"`
    4. `TestResolveDBURL_NeitherSet_Errors` — both unset → returns `("", err)` where `err.Error()` matches the §5.4 wording `"DATABASE_URL environment variable is required"`
  - And the tests use `t.Setenv` (Go 1.17+) to isolate env state per sub-test (no global env mutation that leaks across tests)
  - And package coverage on the new `internal/config` package is ≥95% (trivial — four cases cover every line of the four-branch switch)

- **Scenario: startup log line is integration-tested**
  - Given the binary is invoked with `DATABASE_URL` set (happy path)
  - When stdout/stderr is captured
  - Then the captured output contains the literal substring `"db config: using DATABASE_URL"` emitted **before** any DB ping attempt
  - **Implementation note:** test can be a Go integration test that spawns the binary in a sub-process via `os/exec`, OR captures `log.SetOutput` to a buffer via a `run()` helper extracted from `main`. Architect / tester picks the shape per `architecture.md` §5.7 — but the assertion MUST be on actual logged output, not a mocked logger interface. No "precedence" assertions exist; only the single happy-path log line is in scope.

- **Scenario: Docker/compose configuration updated consistently**
  - Given `docker-compose.yml` at repo root currently sets `DB_URL` for the `mcp-server` service (line 48, set by REQ005 D-009) and `DATABASE_URL` for the `api-server` service
  - When this story is complete
  - Then `docker-compose.yml` sets `DATABASE_URL` (and ONLY `DATABASE_URL`) for BOTH services — the `mcp-server.environment.DB_URL` key is renamed to `DATABASE_URL` with the same URL value
  - And the pre-existing-inconsistency comment block (REQ005 lines 46–47) is replaced with the one-liner `# Standardised on DATABASE_URL per REQ006/US034 (D-006). DB_URL is rejected at startup.` (per `architecture.md` §5.5)
  - And `make e2e-up && make e2e-run` continues to pass green (no regression in the e2e stack)

- **Scenario: documentation updated**
  - Given documentation referencing `DB_URL` may exist in `tests/e2e/README.md`, `services/agent-board/README.md`, or elsewhere (tech-lead grep-confirms during planning)
  - When this story is complete
  - Then any such documentation is updated to say `"DATABASE_URL only; DB_URL is rejected at startup as of REQ006/US034"`
  - And `docs/tech_debt.md` line 97 (the env-var-harmonisation finding) is struck through with `→ fixed in REQ006/US034 (single-var contract; DB_URL now rejected at startup)`

- **Scenario: closes REQ005 architecture OQ-7**
  - Given the architecture OQ-7 about env-var harmonisation was deferred from REQ005
  - When REQ006/US034 is `done`
  - Then `architecture.md` (REQ006) explicitly notes "Closes REQ005 arch Rev-4 OQ-7"

## UI / UX flow expectations
**No UI: BE-prod story (operator-facing).** Operational expectations:

- **Startup log line — ONE happy-path variant.** Both binaries MUST emit, BEFORE the DB ping, the literal log line `"db config: using DATABASE_URL"` (emitted from `main.go`, not from the helper). This is what makes "did the binary read the env var I expected?" a 1-grep operation in production. There is no "both-set" log variant — both-set is now a fatal error, not a precedence resolution.
- **Startup error lines — THREE distinct, operator-actionable messages** (the three §5.4 cases):
  - Only `DB_URL` set: `"DB_URL is no longer supported; rename to DATABASE_URL (REQ006/US034)"` — names the env var to remove, the env var to set, and the migration ticket so an operator can grep.
  - Both `DB_URL` and `DATABASE_URL` set: `"DB_URL is set but no longer supported; remove DB_URL from your environment to disambiguate (DATABASE_URL is the sole accepted name as of REQ006/US034)"` — explicitly tells the operator to remove the legacy var so intent is unambiguous.
  - Neither set: `"DATABASE_URL environment variable is required"` — names the (sole) accepted env var so the operator does not have to read source.
- **No new env vars introduced.** The acceptable surface is now exactly `{DATABASE_URL}` (down from `{DB_URL, DATABASE_URL}` in REQ005). No third name (`POSTGRES_URL`, `PG_URL`, etc.) is added.
- **No runtime behaviour change beyond the env-var contract.** Same ping timeout, same connection-string format, same downstream behaviour once the binary is past startup. The only operator-visible changes are the log line + the three error paths above.

## Out of scope
- **Refactoring the rest of the env-var surface.** `PORT`, `FRONTEND_URL`, `NEXT_PUBLIC_API_BASE_URL`, `MCP_PORT` are left alone.
- **Adding a config-file mechanism** (TOML / YAML / env-file loader). Env vars only.
- **Adding a backward-compatibility shim that silently maps `DB_URL` to `DATABASE_URL`.** The hard-fail is intentional per `architecture.md` §5.1 / D-006 — silent shimming would re-introduce the dual-source-of-truth bug the hard-fail is designed to prevent.
- **Adding a third name like `POSTGRES_URL`.**
- **Adding a `--db-url` CLI flag.**

## Dependencies
- None directly. The architecture decision (now D-006) is locked in `architecture.md` rev 2 as single-var `DATABASE_URL` with hard-fail on `DB_URL`. OQ-1 is closed. No outstanding architecture prerequisites.

## Notes for the team

- **Shared helper, locked.** `architecture.md` §5.2 locks the shared helper in new package `services/agent-board/internal/config` (file `dburl.go`). The "duplicate into both `main.go`" option is dropped — US034's AC requires unit tests for the helper at ≥95% coverage, and a package gives a single clean test surface rather than two duplicate test files.
- **Helper API: `ResolveDBURL() (url string, err error)`.** Returns errors (no `log.Fatal` inside the helper); the caller `main.go` does the fatal and emits the happy-path log line. This keeps the helper unit-testable without subprocess spawning or `*log.Logger` injection.
- **Why both names historically.** `api-server` followed Heroku/12-factor convention (`DATABASE_URL`); `mcp-server` was written later by a contributor who used `DB_URL`. REQ006/US034 ends this divergence by deprecating `DB_URL` outright.
- **Operator-breaking change.** Any deployment passing `DB_URL` to mcp-server (Helm chart values, CI script, custom compose override, hand-set shell envs) MUST be updated to `DATABASE_URL` before pulling this build. mcp-server will refuse to start otherwise. `architecture.md` §5.9 documents this explicitly; the test report (Phase 3c) should call it out so the orchestrator can surface it at sign-off.
- **No "warning" log level needed.** Go's stdlib `log` package does not have levels; the existing code uses `log.Printf` for everything. Stay consistent.
- **Test coverage.** The unit tests on `ResolveDBURL` MUST hit ≥95% on the new `internal/config` package. Four cases over a four-branch switch makes this trivial.
- **Closes tech-debt.** Strike through `docs/tech_debt.md` line 97 in the same commit that ships US034 (or in the sign-off commit) with the wording `→ fixed in REQ006/US034 (single-var contract; DB_URL now rejected at startup)` — see "Scenario: documentation updated" above.
- **Run locally before pushing:**
  - `cd services/agent-board && go test ./... -cover`
  - `cd services/agent-board && DATABASE_URL=postgres://... go run ./cmd/api-server` (verify happy-path log line; should start)
  - `cd services/agent-board && DATABASE_URL=postgres://... go run ./cmd/mcp-server` (verify happy-path log line; should start)
  - `cd services/agent-board && DB_URL=postgres://... go run ./cmd/mcp-server` (verify hard-fail with §5.4 rename-instruction error message)
  - `cd services/agent-board && DATABASE_URL=A DB_URL=B go run ./cmd/api-server` (verify hard-fail with §5.4 remove-DB_URL disambiguate error message)

## Sign-off log
(po-ba appends here on each sign-off pass)

### Sign-off pass 1 — 2026-06-07 — verdict: approved
- **Spec review:** All 9 AC scenarios map to test IDs in `US034_be_unit_tests.md` and are genuinely proven, not just adjacent:
  - "Both binaries accept only DATABASE_URL" → UT-001 (helper happy path) + IT-001/IT-002 (both binaries emit log line and proceed); byte-identical call site confirmed at tech-lead review pass 2 (§5.3).
  - "Deprecated DB_URL rejected with actionable error" → UT-002 (DB_URL-only, exact rename wording), UT-003 (both-set, exact disambiguate wording), and IT-003 (mcp-server subprocess: non-zero exit + rename message before any DB ping). All three §5.4 error strings asserted verbatim via `assert.EqualError`.
  - "Neither set" → UT-004 (exact required-error wording).
  - "Startup log line names resolved var (single-var)" → IT-001/IT-002 assert literal `"db config: using DATABASE_URL"`; single happy-path variant confirmed (no "using DB_URL" variant survives).
  - "Happy path preserved / mcp-server DB_URL deliberately broken" → IT-002 (mcp happy path) + IT-003 (mcp hard-fail). Deliberate breaking change exercised.
  - "Env-var resolution unit-tested, ≥95% cov, t.Setenv" → UT-001..UT-004 + IT-004 (100% statement coverage on dburl.go).
  - "Startup log integration-tested on real logged output" → IT-001/IT-002 use approach (b) `run()` helper asserting captured buffer output (not a mocked logger), log line before ping — honest e2e/unit split.
  - "Docker/compose updated" + "docs updated" + "closes OQ-7" → doc/config concerns, verified by tech-lead (DB_URL→DATABASE_URL rename, §5.5 comment present, tech_debt.md line 97 struck through, e2e README grep-clean) and exercised live by the e2e stack.
  - Edge/error paths the AC implies are all present; no skipped scenarios. E2E justification honest — the helper logic stays at unit level; only the live-stack integration is exercised via the existing Robot suite.
- **Result review:** `US034_test_report.md` reports 9/9 test IDs PASS (UT-001..UT-004, IT-001..IT-005), counts match the spec coverage matrix exactly (no silent dropping). `Skipped Tests: None`; no `t.Skip` / `[Tags] skip`. Full module `go test ./...` 301 passed / 0 failed across 7 packages. Live e2e + 3-clean-run flake check satisfied: 3 consecutive `make e2e-run` invocations each `23 tests, 23 passed, 0 failed` (architecture §10.1 gate met). Tech-lead reached task `Status: completed` at review pass 2 (live e2e gate completed once Podman was available). Report carries timestamp + commit SHA `6fa0726`. The E2E table's "(workaround applied)" annotation is benign — it denotes `DATABASE_URL` supplied via the compose stack (the intended config), corroborated by the tech-lead's pass-2 fresh-rebuild run; not a test bypass or masked failure.
- **Routed to:** none — story approved, `Status: done`.
