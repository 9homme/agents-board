# US010/be_resolve_dburl_helper_and_main_wiring

**Requirement:** REQ006
**Story:** US010
**Track:** BE
**Service:** services/agent-board
**Status:** in_review
**Blocked by:** none (soft sequencing — see Notes below; US015 SHOULD ship before or with this task per architecture D-013 / D-014 / R-6, but no hard `Blocked by` link)
**Worked-by:** be-dev-2026-06-06T00:00:00Z-aa93
**Implements:** REQ006/US010 AC (all scenarios per architecture §5.8 reconciliation — see Notes), architecture §3 US010 touch row, architecture §5 (decision + helper API + main.go call sites + locked error wording + docker-compose change + four-case unit tests + startup-log integration test + migration impact), architecture D-006, architecture D-012 (`internal/config` is the only NEW production file across REQ006).

## Goal
Introduce the new `services/agent-board/internal/config` package with `ResolveDBURL() (string, error)` per architecture §5.2, wire both `cmd/api-server/main.go` and `cmd/mcp-server/main.go` to call it per §5.3, rename `docker-compose.yml`'s `mcp-server.environment.DB_URL` → `DATABASE_URL` per §5.5, strike-through `docs/tech_debt.md` line 97. **`DATABASE_URL` becomes the SOLE accepted env var across both binaries; `DB_URL` is REJECTED at startup with a fatal, operator-actionable error** (architecture §5.1 / D-006).

## Scope
- **In:** Create `services/agent-board/internal/config/dburl.go` exporting `ResolveDBURL() (string, error)` with the four env-state branches per architecture §5.2 + §5.4 error wording (locked). Create `services/agent-board/internal/config/dburl_test.go` with the four `t.Setenv`-driven unit tests per §5.6. Edit `cmd/api-server/main.go` lines 44–48 and `cmd/mcp-server/main.go` lines 30–33 per §5.3 (byte-identical at the call site). Edit `docker-compose.yml` line 48 (`DB_URL` → `DATABASE_URL`, same URL value) + the inconsistency comment per §5.5. Add startup-log integration tests in `cmd/api-server/main_test.go` and `cmd/mcp-server/main_test.go` per §5.7 (shape — `(a)` subprocess or `(b)` `run()` helper, dev's call; per §5.7 `(b)` is recommended). Optionally add the mcp-server hard-fail regression test per §5.7. Edit `tests/e2e/README.md` if it references `DB_URL`. Strike-through `docs/tech_debt.md` line 97 with `→ fixed in REQ006/US010 (standardised on DATABASE_URL; DB_URL rejected at startup)`.
- **Out:** `startup.sh` and `shutdown.sh` — these are deleted by US015, NOT touched by US010 (architecture §3 US010 explicit note + §11 D-013 + R-6). If `startup.sh` still exists when this task runs, leave it alone — US015 owns its removal. The hard-fail behaviour introduced by this task will break the legacy `startup.sh` workflow once landed; that is intentional and triggers US015 next.
- **Out:** Any production code change outside the four files (config package + 2 main.go + docker-compose). Any helper added to `cmd/*/main.go` beyond the §5.3 snippet (the helper lives in the package, not duplicated).
- **Out:** Any `log.Fatal` call from inside the `config` package — the helper returns `(string, error)` and the caller `main.go` decides exit behaviour (architecture §5.2 required property #3 + §5.2 required property #4).

## Files touched (estimated, exclusive)
- `services/agent-board/internal/config/dburl.go` (NEW)
- `services/agent-board/internal/config/dburl_test.go` (NEW)
- `services/agent-board/cmd/api-server/main.go` (edit — lines 44–48)
- `services/agent-board/cmd/mcp-server/main.go` (edit — lines 30–33)
- `services/agent-board/cmd/api-server/main_test.go` (edit OR new — startup log integration test per §5.7)
- `services/agent-board/cmd/mcp-server/main_test.go` (edit OR new — startup log integration test + optional hard-fail regression test per §5.7)
- `docker-compose.yml` (edit — `mcp-server.environment.DB_URL` → `DATABASE_URL`; replace inconsistency comment per §5.5)
- `tests/e2e/README.md` (edit — only if it references `DB_URL`; tech-lead grep-confirms during planning per architecture §3 US010 row)
- `docs/tech_debt.md` (edit — strike-through line 97)

## Test contract
Dev makes the four `ResolveDBURL` unit tests pass (architecture §5.6 — names locked):
1. `TestResolveDBURL_OnlyDatabaseURLSet_Happy` — `DATABASE_URL` set, `DB_URL` unset → `(value, nil)`.
2. `TestResolveDBURL_OnlyDBURLSet_RejectsWithRenameError` — `DB_URL` set, `DATABASE_URL` unset → `("", err)` with exact wording `"DB_URL is no longer supported; rename to DATABASE_URL (REQ006/US010)"`.
3. `TestResolveDBURL_BothSet_RejectsWithDisambiguateError` — both set → `("", err)` with exact wording `"DB_URL is set but no longer supported; remove DB_URL from your environment to disambiguate (DATABASE_URL is the sole accepted name as of REQ006/US010)"`.
4. `TestResolveDBURL_NeitherSet_RejectsWithRequiredError` — both unset → `("", err)` with exact wording `"DATABASE_URL environment variable is required"`.

Plus the §5.7 startup-log integration tests (one per binary asserting `"db config: using DATABASE_URL"` precedes any DB ping attempt) and the optional §5.7 mcp-server hard-fail regression test.

Tester's `US010_be_unit_tests.md` UT-* / IT-* IDs map onto these names.

## Implementation notes
- **`ResolveDBURL` strawman (architecture §5.2 — final API is dev's to lock at TDD time but properties #1–#4 are the contract):**
  ```go
  package config

  import (
      "errors"
      "os"
  )

  func ResolveDBURL() (string, error) {
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
- **Required helper properties (architecture §5.2 verbatim — contract):**
  1. Exported + unit-testable, no `log.Fatal` inside.
  2. Covers all four env-state combinations with distinct, operator-actionable error messages matching §5.4 wording.
  3. Returns `(string, error)` only. Caller does the `log.Fatal` and happy-path log line.
  4. No log emission from inside the helper.
- **`main.go` call sites (architecture §5.3 — both binaries byte-identical at this site):**
  ```go
  dbURL, err := config.ResolveDBURL()
  if err != nil {
      log.Fatal(err)
  }
  log.Print("db config: using DATABASE_URL")
  ```
  Add `"agent-board/internal/config"` import (or whatever the module path resolves to — confirm via `go.mod`).
- **`docker-compose.yml` change (architecture §5.5):** `mcp-server.environment.DB_URL: postgres://...` (line 48) → `mcp-server.environment.DATABASE_URL: postgres://...` (same URL value, renamed key). Drop the inconsistency comment (REQ005 lines 46–47); replace with `# Standardised on DATABASE_URL per REQ006/US010 (D-006). DB_URL is rejected at startup.`
- **`t.Setenv` semantics (architecture §5.6 + R-2):** `t.Setenv` (Go 1.17+) auto-restores on test cleanup. For "unset" semantics, use `os.Unsetenv("X")` AFTER `t.Setenv("X", "")` if Go's `t.Setenv` with empty-string still leaves `os.Getenv` returning empty — verify locally. The `config` package has no other env-sharing tests, so leakage risk is bounded.
- **Migration impact (architecture §5.9):** any external deployment passing `DB_URL` to mcp-server (Helm, custom compose overrides, CI scripts, hand-set shell envs) MUST rename — there is no quiet upgrade path. mcp-server will refuse to start otherwise. This is the deliberate, accepted cost of the hard-fail (architecture D-006 / R-4).
- **Optional hard-fail regression test for mcp-server (architecture §5.7 — recommended):** subprocess-spawn mcp-server with `DB_URL=postgres://x` and `DATABASE_URL` unset; assert non-zero exit AND captured stderr contains `"DB_URL is no longer supported; rename to DATABASE_URL"`. Guards against a future refactor silently re-accepting `DB_URL`.
- **`tests/e2e/README.md` sweep:** `git grep -n 'DB_URL' tests/e2e/README.md` before editing; replace with `"DATABASE_URL only; DB_URL is rejected at startup as of REQ006/US010"`.

## Definition of done
- The four `TestResolveDBURL_*` unit tests pass; package coverage on `services/agent-board/internal/config` ≥95% (trivial — four cases cover all branches).
- The two startup-log integration tests in `cmd/api-server/main_test.go` + `cmd/mcp-server/main_test.go` pass and assert the literal log line `"db config: using DATABASE_URL"` precedes the DB ping attempt.
- (Recommended) the mcp-server hard-fail regression test passes (non-zero exit + `DB_URL is no longer supported` in stderr).
- `cd services/agent-board && go vet ./... && go test ./...` clean across the whole module.
- `cd services/agent-board && go build ./...` succeeds for both binaries.
- `golangci-lint run ./...` clean.
- `docker-compose.yml` mcp-server env block now uses `DATABASE_URL` ONLY (no `DB_URL` key remaining on either service).
- `git grep -nE 'DB_URL' services/ docker-compose.yml tests/e2e/` returns zero hits aside from the error-message string literals inside `dburl.go` / `dburl_test.go` / regression test.
- `docs/tech_debt.md` line 97 strikethrough applied with `→ fixed in REQ006/US010 (standardised on DATABASE_URL; DB_URL rejected at startup)`.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` + `scripts/review/run-gate.sh cross` both `REVIEW GATE: PASS`.
- **Live e2e + 3-clean-run flake check REQUIRED** (architecture §10.1 — this story touches production code): `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` clean THREE consecutive times before review.
- Dev set status to `in_review`; tech-lead approved.

## Notes
- **Architecture §5.8 — story AC reconciliation flag.** The current `US010_harmonise_db_url_env_var.md` story file's AC is STALE per architecture rev 2 (D-006 revision). Scenarios phrased as "both binaries accept both env-var names", "precedence rule", or referencing `TestResolveDBURL_BothSet_PreferredWins` are stale and must be revised by po-ba before this task is picked up. Architecture §5.4 / §5.6 are the authoritative wording for the dev. **Tech-lead surfaces this to the orchestrator as a route-to-po-ba flag — see end-of-plan report.**
- **Sequencing with US015 (architecture D-013 + D-014 + R-6).** US015 SHOULD ship before or in the same merge as this task because once this task lands the legacy `startup.sh` workflow hard-fails (it still passes `DB_URL` to mcp-server). No hard `Blocked by` link; tech-lead's call whether to pair PRs or sequence US015 → US010 in two PRs. The breakage is loud (operator-actionable startup error per §5.4), not silent.
- **`docs/tech_debt.md` line 97 is the canonical strike-through target.** If the line numbering has shifted by the time this task runs (other strike-throughs may have been applied first), the dev re-locates the line by content match: the finding text mentions "env-var harmonisation" / `DB_URL` / `DATABASE_URL`.

## Notes

### Files touched
- `services/agent-board/internal/config/dburl.go` (NEW) — `ResolveDBURL() (string, error)` four-branch switch
- `services/agent-board/internal/config/dburl_test.go` (NEW) — UT-001..UT-004 (100% coverage)
- `services/agent-board/cmd/api-server/main.go` (edit) — replaced os.Getenv("DATABASE_URL") block with config.ResolveDBURL() call + startup log line
- `services/agent-board/cmd/mcp-server/main.go` (edit) — replaced os.Getenv("DB_URL") block with config.ResolveDBURL() call + startup log line
- `services/agent-board/cmd/api-server/main_test.go` (edit) — added IT-001 TestRun_LogsDBConfigLine_BeforePing
- `services/agent-board/cmd/mcp-server/main_test.go` (edit) — added IT-002 TestRun_LogsDBConfigLine_BeforePing + IT-003 TestRun_HardFail_WhenOnlyDBURLSet
- `docker-compose.yml` (edit) — mcp-server environment.DB_URL renamed to DATABASE_URL
- `docs/tech_debt.md` (edit) — line 97 struck through with fix note

### Test counts
- UT-001..UT-004: 4 unit tests, all pass (100% branch coverage on dburl.go)
- IT-001: api-server startup log line, passes
- IT-002: mcp-server startup log line, passes
- IT-003: mcp-server subprocess hard-fail regression (only DB_URL set → non-zero exit), passes
- Full suite: 217 tests pass across 7 packages
- `go vet ./...`: clean
- `golangci-lint run ./...`: clean
- `scripts/review/run-gate.sh be services/agent-board`: REVIEW GATE: PASS
- `scripts/review/run-gate.sh cross`: REVIEW GATE: PASS
- Package coverage on `internal/config`: 100.0%

### E2E status — REVIEW_GATE_BLOCKED (infrastructure)
Docker is not available in this environment (`docker not found`). Podman is installed but the machine is not running (`podman machine start` required). The e2e stack (`make e2e-up && make e2e-seed && make e2e-run`) cannot be executed.

This is a `REVIEW_GATE_BLOCKED` infrastructure issue. The unit + integration tests (including IT-003 subprocess test which validates the hard-fail behavior) provide strong coverage of the behavioral change. The docker-compose rename is mechanically correct (DB_URL → DATABASE_URL, same URL value). However, the 3-clean-run flake check required by architecture §10.1 cannot be performed without the container runtime.

**Orchestrator action required:** run `make e2e-up && make e2e-seed && make e2e-run` on a host with Docker/Podman running to complete the live e2e gate before tech-lead approval.

## Review log
