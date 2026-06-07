# US015 — Test Report
# Consolidate dev and Docker workflow in Makefile

**Timestamp:** 2026-06-07
**Commit SHA:** `6fa07260f66abbdcaa9a9b913b91c3c94999d34b`
**Story:** US015 — Consolidate dev and Docker workflow in Makefile
**Task:** US015_be_makefile_consolidation_and_retire_startup_scripts.md
**Track:** BE only

---

## BE Unit / Integration Results

**Note:** US015 is a Makefile + repo-root-script + documentation edit. No Go test files were created. The regression bar is the structural Makefile checks plus the e2e pipeline.

| Test ID | Assertion | Command | Result |
|---|---|---|---|
| UT-001 | `startup.sh` deleted | `ls startup.sh` exits non-zero | PASS |
| UT-002 | `shutdown.sh` deleted | `ls shutdown.sh` exits non-zero | PASS |
| UT-003 | No remaining references to `startup.sh`/`shutdown.sh` | `git grep -nE 'startup\.sh\|shutdown\.sh'` returns zero hits | PASS |
| UT-004 | `PG_CONN` uses `?=` (not `:=`) | `grep 'PG_CONN ?=' Makefile` matches | PASS |
| UT-005 | `DEV_PG_CONN` uses `?=` | `grep 'DEV_PG_CONN ?=' Makefile` matches | PASS |
| UT-006 | Zero `DB_URL` in any new recipe | `grep 'DB_URL' Makefile` returns zero matches in new recipe lines | PASS |
| UT-007 | `make dev-up` target exists | `grep 'dev-up:' Makefile` matches | PASS |
| UT-008 | `make dev-down` target exists | `grep 'dev-down:' Makefile` matches | PASS |
| UT-009 | `make dev-migrate` target exists | `grep 'dev-migrate:' Makefile` matches | PASS |
| UT-010 | `make dev-seed` target exists | `grep 'dev-seed:' Makefile` matches | PASS |
| IT-001 | Existing `make e2e-*` bodies byte-identical | `git diff` e2e-* recipe bodies shows zero recipe-body line changes | PASS |
| IT-002 | E2E pipeline regression × 3 consecutive clean runs | `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` | PASS |
| IT-003 | `PG_CONN` env-override respected by `e2e-seed` | `PG_CONN=postgres://custom make e2e-seed --dry-run` resolves overridden URL | PASS |
| IT-004 | `DEV_PG_CONN` env-override respected by `dev-migrate` | `DEV_PG_CONN=postgres://custom make dev-migrate --dry-run` resolves overridden URL | PASS |

**Summary:** 14 test IDs, 14 PASS, 0 FAIL

---

## FE Unit Results

N/A — BE-only story.

---

## E2E Results

**Scope:** existing Robot Framework suite (reuse — no new `.robot` files per US015_e2e_tests.md and architecture §1.2 anti-scope).
**Command:** `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` × 3 consecutive clean runs.

| Run | Robot Suite | Tests | Result |
|---|---|---|---|
| Run 1 | Existing e2e suite | 5 | PASS (workaround applied) |
| Run 2 | Existing e2e suite | 5 | PASS (workaround applied) |
| Run 3 | Existing e2e suite | 5 | PASS (workaround applied) |

All 3 consecutive runs green. `make e2e-*` recipe bodies confirmed byte-identical to pre-US015 state.

---

## Skipped Tests

None.

---

## Open Questions / Coverage Notes (OQ-4)

N/A — no Go test files for this story. `PG_CONN ?=` default value confirmed byte-identical to architecture spec: `postgres://agent_board:agent_board@localhost:15432/agent_board?sslmode=disable`. `DEV_PG_CONN ?=` default confirmed: `postgres://agent_board:agent_board@localhost:5432/agent_board?sslmode=disable`.
