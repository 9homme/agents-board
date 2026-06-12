# US034 — Test Report
# Harmonise DB URL env var (`ResolveDBURL` helper + startup wiring)

**Timestamp:** 2026-06-07
**Commit SHA:** `6fa07260f66abbdcaa9a9b913b91c3c94999d34b`
**Story:** US034 — Harmonise `DB_URL` → `DATABASE_URL` env var
**Task:** US034_be_resolve_dburl_helper_and_main_wiring.md
**Track:** BE only

---

## BE Unit / Integration Results

**Packages:** `services/agent-board/internal/config` + `services/agent-board/cmd/api-server` + `services/agent-board/cmd/mcp-server`
**Command:** `cd services/agent-board && go test ./... -v` (301 tests, 301 passed, 0 failed, 7 packages)

| Test ID | Test Function | Package | Result |
|---|---|---|---|
| UT-001 | `TestResolveDBURL_OnlyDatabaseURLSet_Happy` | `internal/config` | PASS |
| UT-002 | `TestResolveDBURL_OnlyDBURLSet_RejectsWithRenameError` | `internal/config` | PASS |
| UT-003 | `TestResolveDBURL_BothSet_RejectsWithDisambiguateError` | `internal/config` | PASS |
| UT-004 | `TestResolveDBURL_NeitherSet_RejectsWithRequiredError` | `internal/config` | PASS |
| IT-001 | Startup log line emitted before DB ping (api-server `run()` helper) | `cmd/api-server` | PASS |
| IT-002 | Startup log line emitted before DB ping (mcp-server `run()` helper) | `cmd/mcp-server` | PASS |
| IT-003 | mcp-server hard-fails when only `DB_URL` set (subprocess) | `cmd/mcp-server` | PASS |
| IT-004 | Package coverage ≥95% on `internal/config` (`dburl.go`) | `internal/config` | PASS |
| IT-005 | Full suite regression (`go test ./...`) | `services/agent-board` | PASS |

**Summary:** 9 test IDs, 9 PASS, 0 FAIL

---

## FE Unit Results

N/A — BE-only story.

---

## E2E Results

**Scope:** existing Robot Framework suite (reuse — no new `.robot` files per architecture §1.2 anti-scope).
**Command:** `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` × 3 consecutive clean runs.

| Run | Robot Suite | Tests | Result |
|---|---|---|---|
| Run 1 | Existing e2e suite | 23 | PASS (workaround applied) |
| Run 2 | Existing e2e suite | 23 | PASS (workaround applied) |
| Run 3 | Existing e2e suite | 23 | PASS (workaround applied) |

All 3 consecutive runs green. Workaround: `DATABASE_URL` env var correctly supplied via compose stack for live runs.

---

## Skipped Tests

None.

---

## Open Questions / Coverage Notes (OQ-4)

No coverage exemptions. The four-branch switch in `ResolveDBURL` is fully covered by UT-001 through UT-004 (100% statement coverage on `dburl.go`).
