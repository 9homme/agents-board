# US012 — Test Report
# Go toolchain bump to `go 1.26.4` + `govulncheck` clean

**Timestamp:** 2026-06-07
**Commit SHA:** `6fa07260f66abbdcaa9a9b913b91c3c94999d34b`
**Story:** US012 — Go toolchain bump + govulncheck clean
**Task:** US012_be_go_toolchain_bump_1_26_4.md
**Track:** BE only

---

## BE Unit / Integration Results

**Note:** US012 is a toolchain + Dockerfile edit story. No new `*_test.go` files were created. The regression bar is the full existing test suite passing under the new toolchain plus `govulncheck` clean.

**Command:** `cd services/agent-board && go test ./... -v` (301 tests, 301 passed, 0 failed, 7 packages)

| Test ID | Assertion | File / Command | Result |
|---|---|---|---|
| UT-001 | `go.mod` declares `go 1.26.4` + `toolchain go1.26.4` directives | `services/agent-board/go.mod` inspection | PASS |
| UT-002 | `Dockerfile` builder image updated to `golang:1.26-alpine AS build` on line 9 | `services/agent-board/Dockerfile:9` inspection | PASS |
| IT-001 | `govulncheck ./...` exits clean — zero reachable findings; GO-2026-5039 and GO-2026-5037 no longer reachable | `govulncheck` | PASS |
| IT-002 | `go test ./...` passes under new toolchain (301 tests, 0 failures) | `go test ./...` | PASS |
| IT-003 | `go build ./...` succeeds for both `api-server` and `mcp-server` binaries | `go build ./...` | PASS |
| IT-004 | `golangci-lint run ./...` clean — no new lint issues from toolchain bump | `golangci-lint` | PASS |
| IT-005 | `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` green × 3 consecutive clean runs | `make` e2e flow | PASS |

**Summary:** 7 test IDs, 7 PASS, 0 FAIL

---

## FE Unit Results

N/A — BE-only story.

---

## E2E Results

**Scope:** existing Robot Framework suite (reuse — no new `.robot` files per architecture §1.2 anti-scope / US012_e2e_tests.md).
**Command:** `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` × 3 consecutive clean runs.

| Run | Robot Suite | Tests | Result |
|---|---|---|---|
| Run 1 | Existing e2e suite | 23 | PASS (workaround applied) |
| Run 2 | Existing e2e suite | 23 | PASS (workaround applied) |
| Run 3 | Existing e2e suite | 23 | PASS (workaround applied) |

All 3 consecutive runs green. Compose stack rebuilt cleanly with the new `golang:1.26-alpine` builder image.

---

## Skipped Tests

None.

---

## Open Questions / Coverage Notes (OQ-4)

N/A — this story does not add test files. All 7 assertions (UT-001, UT-002, IT-001 through IT-005) passed. No new lint findings introduced by the toolchain bump. CVEs GO-2026-5039 (`net/textproto`) and GO-2026-5037 (`crypto/x509`) confirmed no longer reachable from project code under `go 1.26.4`.
