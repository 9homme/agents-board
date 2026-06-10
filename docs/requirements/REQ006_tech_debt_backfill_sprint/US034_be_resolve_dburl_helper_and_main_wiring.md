# US034/be_resolve_dburl_helper_and_main_wiring

**Requirement:** REQ006
**Story:** US034
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** none (soft sequencing — see Notes below; US038 SHOULD ship before or with this task per architecture D-013 / D-014 / R-6, but no hard `Blocked by` link)
**Worked-by:** be-dev-2026-06-06T00:00:00Z-aa93
**Implements:** REQ006/US034 AC (all scenarios per architecture §5.8 reconciliation — see Notes), architecture §3 US034 touch row, architecture §5 (decision + helper API + main.go call sites + locked error wording + docker-compose change + four-case unit tests + startup-log integration test + migration impact), architecture D-006, architecture D-012 (`internal/config` is the only NEW production file across REQ006).

## Goal
Introduce the new `services/agent-board/internal/config` package with `ResolveDBURL() (string, error)` per architecture §5.2, wire both `cmd/api-server/main.go` and `cmd/mcp-server/main.go` to call it per §5.3, rename `docker-compose.yml`'s `mcp-server.environment.DB_URL` → `DATABASE_URL` per §5.5, strike-through `docs/tech_debt.md` line 97. **`DATABASE_URL` becomes the SOLE accepted env var across both binaries; `DB_URL` is REJECTED at startup with a fatal, operator-actionable error** (architecture §5.1 / D-006).

## Scope
- **In:** Create `services/agent-board/internal/config/dburl.go` exporting `ResolveDBURL() (string, error)` with the four env-state branches per architecture §5.2 + §5.4 error wording (locked). Create `services/agent-board/internal/config/dburl_test.go` with the four `t.Setenv`-driven unit tests per §5.6. Edit `cmd/api-server/main.go` lines 44–48 and `cmd/mcp-server/main.go` lines 30–33 per §5.3 (byte-identical at the call site). Edit `docker-compose.yml` line 48 (`DB_URL` → `DATABASE_URL`, same URL value) + the inconsistency comment per §5.5. Add startup-log integration tests in `cmd/api-server/main_test.go` and `cmd/mcp-server/main_test.go` per §5.7 (shape — `(a)` subprocess or `(b)` `run()` helper, dev's call; per §5.7 `(b)` is recommended). Optionally add the mcp-server hard-fail regression test per §5.7. Edit `tests/e2e/README.md` if it references `DB_URL`. Strike-through `docs/tech_debt.md` line 97 with `→ fixed in REQ006/US034 (standardised on DATABASE_URL; DB_URL rejected at startup)`.
- **Out:** `startup.sh` and `shutdown.sh` — these are deleted by US038, NOT touched by US034 (architecture §3 US034 explicit note + §11 D-013 + R-6). If `startup.sh` still exists when this task runs, leave it alone — US038 owns its removal. The hard-fail behaviour introduced by this task will break the legacy `startup.sh` workflow once landed; that is intentional and triggers US038 next.
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
- `tests/e2e/README.md` (edit — only if it references `DB_URL`; tech-lead grep-confirms during planning per architecture §3 US034 row)
- `docs/tech_debt.md` (edit — strike-through line 97)

## Test contract
Dev makes the four `ResolveDBURL` unit tests pass (architecture §5.6 — names locked):
1. `TestResolveDBURL_OnlyDatabaseURLSet_Happy` — `DATABASE_URL` set, `DB_URL` unset → `(value, nil)`.
2. `TestResolveDBURL_OnlyDBURLSet_RejectsWithRenameError` — `DB_URL` set, `DATABASE_URL` unset → `("", err)` with exact wording `"DB_URL is no longer supported; rename to DATABASE_URL (REQ006/US034)"`.
3. `TestResolveDBURL_BothSet_RejectsWithDisambiguateError` — both set → `("", err)` with exact wording `"DB_URL is set but no longer supported; remove DB_URL from your environment to disambiguate (DATABASE_URL is the sole accepted name as of REQ006/US034)"`.
4. `TestResolveDBURL_NeitherSet_RejectsWithRequiredError` — both unset → `("", err)` with exact wording `"DATABASE_URL environment variable is required"`.

Plus the §5.7 startup-log integration tests (one per binary asserting `"db config: using DATABASE_URL"` precedes any DB ping attempt) and the optional §5.7 mcp-server hard-fail regression test.

Tester's `US034_be_unit_tests.md` UT-* / IT-* IDs map onto these names.

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
          return "", errors.New("DB_URL is set but no longer supported; remove DB_URL from your environment to disambiguate (DATABASE_URL is the sole accepted name as of REQ006/US034)")
      case dbURL != "":
          return dbURL, nil
      case legacyURL != "":
          return "", errors.New("DB_URL is no longer supported; rename to DATABASE_URL (REQ006/US034)")
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
- **`docker-compose.yml` change (architecture §5.5):** `mcp-server.environment.DB_URL: postgres://...` (line 48) → `mcp-server.environment.DATABASE_URL: postgres://...` (same URL value, renamed key). Drop the inconsistency comment (REQ005 lines 46–47); replace with `# Standardised on DATABASE_URL per REQ006/US034 (D-006). DB_URL is rejected at startup.`
- **`t.Setenv` semantics (architecture §5.6 + R-2):** `t.Setenv` (Go 1.17+) auto-restores on test cleanup. For "unset" semantics, use `os.Unsetenv("X")` AFTER `t.Setenv("X", "")` if Go's `t.Setenv` with empty-string still leaves `os.Getenv` returning empty — verify locally. The `config` package has no other env-sharing tests, so leakage risk is bounded.
- **Migration impact (architecture §5.9):** any external deployment passing `DB_URL` to mcp-server (Helm, custom compose overrides, CI scripts, hand-set shell envs) MUST rename — there is no quiet upgrade path. mcp-server will refuse to start otherwise. This is the deliberate, accepted cost of the hard-fail (architecture D-006 / R-4).
- **Optional hard-fail regression test for mcp-server (architecture §5.7 — recommended):** subprocess-spawn mcp-server with `DB_URL=postgres://x` and `DATABASE_URL` unset; assert non-zero exit AND captured stderr contains `"DB_URL is no longer supported; rename to DATABASE_URL"`. Guards against a future refactor silently re-accepting `DB_URL`.
- **`tests/e2e/README.md` sweep:** `git grep -n 'DB_URL' tests/e2e/README.md` before editing; replace with `"DATABASE_URL only; DB_URL is rejected at startup as of REQ006/US034"`.

## Definition of done
- The four `TestResolveDBURL_*` unit tests pass; package coverage on `services/agent-board/internal/config` ≥95% (trivial — four cases cover all branches).
- The two startup-log integration tests in `cmd/api-server/main_test.go` + `cmd/mcp-server/main_test.go` pass and assert the literal log line `"db config: using DATABASE_URL"` precedes the DB ping attempt.
- (Recommended) the mcp-server hard-fail regression test passes (non-zero exit + `DB_URL is no longer supported` in stderr).
- `cd services/agent-board && go vet ./... && go test ./...` clean across the whole module.
- `cd services/agent-board && go build ./...` succeeds for both binaries.
- `golangci-lint run ./...` clean.
- `docker-compose.yml` mcp-server env block now uses `DATABASE_URL` ONLY (no `DB_URL` key remaining on either service).
- `git grep -nE 'DB_URL' services/ docker-compose.yml tests/e2e/` returns zero hits aside from the error-message string literals inside `dburl.go` / `dburl_test.go` / regression test.
- `docs/tech_debt.md` line 97 strikethrough applied with `→ fixed in REQ006/US034 (standardised on DATABASE_URL; DB_URL rejected at startup)`.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` + `scripts/review/run-gate.sh cross` both `REVIEW GATE: PASS`.
- **Live e2e + 3-clean-run flake check REQUIRED** (architecture §10.1 — this story touches production code): `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` clean THREE consecutive times before review.
- Dev set status to `in_review`; tech-lead approved.

## Notes
- **Architecture §5.8 — story AC reconciliation flag.** The current `US034_harmonise_db_url_env_var.md` story file's AC is STALE per architecture rev 2 (D-006 revision). Scenarios phrased as "both binaries accept both env-var names", "precedence rule", or referencing `TestResolveDBURL_BothSet_PreferredWins` are stale and must be revised by po-ba before this task is picked up. Architecture §5.4 / §5.6 are the authoritative wording for the dev. **Tech-lead surfaces this to the orchestrator as a route-to-po-ba flag — see end-of-plan report.**
- **Sequencing with US038 (architecture D-013 + D-014 + R-6).** US038 SHOULD ship before or in the same merge as this task because once this task lands the legacy `startup.sh` workflow hard-fails (it still passes `DB_URL` to mcp-server). No hard `Blocked by` link; tech-lead's call whether to pair PRs or sequence US038 → US034 in two PRs. The breakage is loud (operator-actionable startup error per §5.4), not silent.
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

### Review pass 1 — 2026-06-07 — verdict: blocked_review_gate

**Why blocked (NOT changes_requested, NOT approved):** the code is correct on every check I was able to run, but the mandatory **live-e2e + 3-clean-run flake verification** required by architecture §10.1 (which names US034 explicitly: "Any story that touches `services/<>/` production code (US034, US035) requires a `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` clean run three times in a row before code review can pass") and by this task's own DoD line 99 **cannot be performed on the review host** — the container runtime is unavailable. Per tech-lead.md review-mode rule (REQ005/US032 follow-up): "If the e2e stack itself is unavailable on the review host, that's `blocked_review_gate` — NOT `approved`." The three live-e2e summary lines are required evidence for an `approved` verdict and cannot be produced; I may not approve around a missing gate, and I may not fabricate the evidence. This routes to the gate-fix / infra track, not back to the dev — the dev's code is not at fault.

**Container-runtime probe on review host (verbatim):**
- `which docker` → `docker not found` (exit 1)
- `docker-compose` → `not found` (exit 1)
- `podman-compose` → not installed
- `podman machine list` → `podman-machine-default ... LAST UP: Never`; `podman info` → exit 125 (machine not started)
- Makefile `_check-compose` guard would fail at recipe time; `make e2e-up` cannot run.

**Everything that COULD be verified — all green (recorded so the next reviewer does not re-run the code-level checks, only the live e2e):**

Code-level conformance:
- `dburl.go` matches architecture §5.2 contract exactly: package `config`, exported `ResolveDBURL() (string, error)`, four-branch switch, no `log.Fatal`/no log emission inside the helper (properties #1–#4 hold). All three error strings are byte-identical to the locked §5.4 wording.
- `cmd/api-server/main.go` (lines 45–50) and `cmd/mcp-server/main.go` (lines 31–36) call sites are byte-identical to §5.3, with the happy-path log line `"db config: using DATABASE_URL"` emitted from `main.go` (not the helper).
- `docker-compose.yml`: both `api-server` and `mcp-server` set `DATABASE_URL` only; `DB_URL` key removed; §5.5 comment present verbatim (`# Standardised on DATABASE_URL per REQ006/US034 (D-006). DB_URL is rejected at startup.`).
- `git grep DB_URL services/ docker-compose.yml tests/e2e/`: every remaining hit is an error-message string literal, test code (UT-002/UT-003 + IT-003 subprocess setup), or an explanatory comment — zero live env-var usages. As DoD line 96 requires.
- `tests/e2e/README.md`: no `DB_URL` reference (grep exit 1) — correctly left untouched.
- `startup.sh`/`shutdown.sh`: not touched (already removed by US038, which landed first in this merge sequence — §3 US034 note honoured).
- `docs/tech_debt.md` line 97: struck through with the exact required suffix `→ fixed in REQ006/US034 (standardised on DATABASE_URL; DB_URL rejected at startup)`.

Test contract (all named tests present + passing):
- UT-001..UT-004 (`internal/config/dburl_test.go`): 4/4 PASS, `coverage: 100.0% of statements`.
- IT-001 `TestRun_LogsDBConfigLine_BeforePing` (api-server): PASS.
- IT-002 `TestRun_LogsDBConfigLine_BeforePing` (mcp-server): PASS.
- IT-003 `TestRun_HardFail_WhenOnlyDBURLSet` (mcp-server subprocess, non-zero exit + rename message): PASS.
- Full module suite `go test ./...`: all 7 packages `ok` (api-server, mcp-server, config, domain, handler, mcp, repo).
- `go vet ./...`: clean. `go build ./...`: both binaries build.

Gates:
- `scripts/review/run-gate.sh be services/agent-board` → `REVIEW GATE: PASS` (gofmt -s, go vet, golangci-lint, go test all PASS; gosec/govulncheck WARN-skipped by the gate itself — not installed — with gosec coverage via the golangci-lint gosec linter, the gate's designed fallback).
- `scripts/review/run-gate.sh cross` → `REVIEW GATE: PASS` (semgrep owasp/golang/typescript PASS, gitleaks no secrets).
- No REQ006 Robot e2e suite exists (`tests/e2e/REQ006_*` absent), so robot `--dryrun` is N/A for this task.

Per-file coverage (recorded for the next reviewer — NOT the blocker, see note below):
- `internal/config/dburl.go`: 100.0% — clears ≥80%.
- `cmd/api-server/main.go`: `main` 0.0%, `run` 52.6%, `pingDB` 100.0% (file 50.0%).
- `cmd/mcp-server/main.go`: `main` 0.0%, `run` 34.2%, `pingDB` 100.0% (file 33.3%).
  - The two `main.go` files are below the 80% per-file threshold, but the uncovered portion is the pre-existing server-wiring/`e.Start` path that is inherently e2e-only (not unit-testable) and predates this task. The lines US034 actually added (the `ResolveDBURL()` call + happy-path log line) ARE exercised by IT-001/IT-002. This is the standard `main`-package exemption pattern. I am NOT treating this as the blocking issue — it is the live-e2e UNAVAILABILITY that blocks. When the e2e gate is satisfied on a runtime-capable host, the next reviewer should note the main.go sub-threshold coverage as an accepted exemption (main wiring, validated by e2e) rather than `changes_requested`.

TDG discipline (verified on branch commits): clean red → green → refactor cycles, all tagged `(US034)`:
- `red: test spec for ResolveDBURL four env-state branches (US034)`
- `green: implement ResolveDBURL four-branch config helper (US034)`
- `red: test spec for startup log line integration tests IT-001 IT-002 IT-003 (US034)`
- `green: wire both main.go files to use config.ResolveDBURL (US034)`
- `refactor: rename DB_URL to DATABASE_URL in docker-compose; strike tech_debt.md line 97 (US034)`
- `refactor: use exec.CommandContext in IT-003 to satisfy noctx linter (US034)`

**Action required (orchestrator → gate-fix / infra track, NOT the dev):** start a container runtime on the review host (`podman machine start` or install Docker + a compose provider), then run `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` THREE consecutive times and confirm all three are 100% green (`N tests, N passed, 0 failed`). Paste the three summary lines, then re-review for `approved`. No code change is needed for the verdict to flip — the implementation is already correct.

Tech-debt: none filed this pass. The implementation is clean — no style nits, no scope creep, no sibling-pattern divergence, no dead code, no unowned TODOs in the US034 diff.

### Review pass 2 — 2026-06-07 — verdict: approved

**Context:** pass 1 blocked solely on the unavailable container runtime (no Docker/Podman on the review host) preventing the mandatory live-e2e + 3-clean-run flake check (architecture §10.1, DoD line 99). Podman machine is now running (`DOCKER_HOST=unix:///.../podman-machine-default-api.sock`, `podman info` → UP). The code was already verified correct in pass 1; this pass re-confirms the code-level gates AND completes the previously-impossible live-e2e gate.

**Implementation re-verification (D-003 / D-005 / D-006):**
- `internal/config/dburl.go`: `ResolveDBURL() (string, error)` present, four-branch switch, no `log.Fatal`/no log emission inside helper, all three error strings byte-identical to locked §5.4 wording (D-003).
- `cmd/api-server/main.go` lines 45–50: calls `config.ResolveDBURL()`, `log.Fatal(err)` on error (fails loud), happy-path log line `"db config: using DATABASE_URL"` from main, not the helper (D-005).
- `cmd/mcp-server/main.go` lines 31–36: identical call site, fails loud on missing env (D-006).
- Neither binary reads `DB_URL` directly — `git grep -nE 'DB_URL' services/ docker-compose.yml tests/e2e/` returns only error-string literals, test code, and the helper's own legacy-detection `os.Getenv("DB_URL")`. Zero live env-var consumption outside the helper. `docker-compose.yml` mcp-server uses `DATABASE_URL` only.

**Tests + gates (re-run on review host):**
- `cd services/agent-board && go test ./...` → `Go test: 301 passed in 7 packages` (UT-001..UT-004, IT-001/IT-002/IT-003 all green; full module clean). `go vet` clean.
- `scripts/review/run-gate.sh be services/agent-board` → final line `REVIEW GATE: PASS` (gofmt -s, go vet, golangci-lint, go test all PASS; gosec/govulncheck WARN-skipped by the gate itself — coverage via golangci-lint gosec linter, the gate's designed fallback).
- `scripts/review/run-gate.sh cross` → final line `REVIEW GATE: PASS` (semgrep owasp/golang/typescript PASS, gitleaks no secrets).

**Per-file coverage (`go tool cover -func`):**
- `internal/config/dburl.go`: `ResolveDBURL` 100.0% — clears ≥80%.
- `cmd/api-server/main.go`: `main` 0.0%, `run` 52.6%, `pingDB` 100.0% (file ~50%). Accepted main-wiring exemption — the uncovered lines are the pre-existing `e.Start`/server-bootstrap path that is inherently e2e-only (not unit-testable) and predates this task; the lines US034 added (the `ResolveDBURL()` call + happy-path log line) ARE exercised by IT-001 and validated live by the e2e runs below.
- `cmd/mcp-server/main.go`: `main` 0.0%, `run` 34.2%, `pingDB` 100.0% (file ~33%). Same accepted main-wiring exemption; US034's added lines exercised by IT-002 + IT-003 + the live e2e runs.

**Robot e2e dryrun:** N/A — no `tests/e2e/REQ006_*` suite exists for this REQ. The existing REQ001–005 suites exercise the live api-server + mcp-server stack this task wired.

**Live e2e + 3-clean-run flake verification (architecture §10.1 — THE pass-1 blocker, now satisfied):**
Container runtime: Podman 5.8.2 via `podman-compose` (`make COMPOSE="podman-compose" ...`, `DOCKER_HOST` exported to the podman machine API socket). Images rebuilt from current source with `podman build` (fresh `go build` of both binaries — confirmed compiling the new `config` package + wired main.go files; the initially-running stale containers from a prior session were crash-looping on the OLD `"DB_URL environment variable is required"` code and were fully torn down with `down -v` before the rebuild). Stack brought up fresh from a clean postgres volume, then migrated + seeded (`make e2e-seed`: migrations 000001/000002 applied + REQ000_baseline seed inserted). In the live stack the **mcp-server (the binary this task wired) starts healthy and serves `/sse` 200** — direct proof `DATABASE_URL` resolution works end-to-end, where the pre-fix container had been crash-looping.

Three consecutive `make e2e-run` invocations against the running stack — Robot Framework summary lines verbatim:
- Run 1: `23 tests, 23 passed, 0 failed`
- Run 2: `23 tests, 23 passed, 0 failed`
- Run 3: `23 tests, 23 passed, 0 failed`

All three runs 100% green — no failures, no flakes. Stack torn down with `make e2e-down` (volume removed) after run 3.

**Verdict:** approved. Status → completed. Every DoD line is now satisfied including the previously-impossible live-e2e gate. No code change was required between pass 1 and pass 2 — the implementation was correct; only the infra was missing.

Tech-debt: none filed this pass. (Note for the next REQ retrospective, NOT a code finding: the migrations `000001`/`000002` `.up.sql` are not idempotent under `e2e-seed`'s `ON_ERROR_STOP=1` when run against a non-empty volume — `relation already exists`. This is pre-existing infra ergonomics outside US034's scope and is already implicitly tracked by the `make e2e-down -v` clean-volume workflow; not filing a new tech_debt line for it as it is not introduced by this task.)
