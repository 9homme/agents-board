# US012/be_go_toolchain_bump_1_26_4

**Requirement:** REQ006
**Story:** US012
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** none
**Worked-by:** be-dev-2026-06-06T00:00:00Z-a4f2
**Implements:** REQ006/US012 AC (all scenarios — `go.mod` bump, `Dockerfile` builder bump, `govulncheck` clean, `go test ./...` clean, `golangci-lint` clean, `go build ./...` clean for both binaries, `make e2e-up` clean), architecture §3 US012 touch row, architecture §6 (toolchain decision + verified findings + version pin + CI/Docker knock-on + govulncheck post-bump assertion), architecture D-007 (`go 1.26.4`).

## Goal
Bump the Go toolchain across `services/agent-board/` from `go 1.25.0` to **`go 1.26.4`** (architecture D-007), update the `Dockerfile` builder image to `golang:1.26-alpine`, and verify `govulncheck ./...` returns clean (specifically the `crypto/x509` GO-2026-5037 and `net/textproto` GO-2026-5039 stdlib findings are gone per architecture §6.1). Strike-through `docs/tech_debt.md` line 28.

## Scope
- **In:** Edit `services/agent-board/go.mod` line 3: `go 1.25.0` → `go 1.26.4`. Add a `toolchain go1.26.4` directive on a new line (the modern `go` + `toolchain` shape per architecture §6.2). Edit `services/agent-board/Dockerfile` line 9: `FROM golang:1.25-alpine AS build` → `FROM golang:1.26-alpine AS build` (minor-tracking alpine tag picks up latest 1.26.x at build time per architecture §6.2). Strike-through `docs/tech_debt.md` line 28 with `→ fixed in REQ006/US012`.
- **Out:** Any `go.sum` regeneration beyond what `go mod tidy` produces after the bump. Any code change to satisfy a Go 1.26 deprecation warning (if one surfaces, raise as a follow-up story — architecture §6.5 / R-1). Any change to the runtime distroless image (REQ005 D-010 — unchanged). Pinning the alpine tag to `golang:1.26.4-alpine` exactly (architecture §6.2 prefers the minor-tracking tag).

## Files touched (estimated, exclusive)
- `services/agent-board/go.mod` (edit — line 3 + new `toolchain` directive line)
- `services/agent-board/go.sum` (edit — only if `go mod tidy` changes it; expected delta is minimal/none for a stdlib-only bump)
- `services/agent-board/Dockerfile` (edit — line 9 builder image)
- `docs/tech_debt.md` (edit — strike-through line 28)

## Test contract
US012 is a pure production-code touch (no new test functions named in AC). The "tests" are existence/behaviour assertions:
- `cd services/agent-board && govulncheck ./...` exits clean — specifically GO-2026-5037 (`crypto/x509`) and GO-2026-5039 (`net/textproto`) findings are no longer reported (architecture §6.1 / §6.5).
- `cd services/agent-board && go test ./...` passes across both binaries' code paths.
- `cd services/agent-board && go build ./...` succeeds for both `cmd/api-server` and `cmd/mcp-server`.
- `golangci-lint run ./...` clean.
- `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` clean — Dockerised compose build picks up `golang:1.26-alpine` and the e2e pipeline stays green.

Tester's `US012_be_unit_tests.md` will likely contain one or two UT-* IDs framed as "govulncheck-clean assertion" + "e2e-stack-builds-and-runs assertion"; no in-Go unit tests are required.

## Implementation notes
- **`go.mod` shape (architecture §6.2):**
  ```
  module agent-board

  go 1.26.4

  toolchain go1.26.4

  require ( ... )
  ```
  Keep both directives. The `go` directive sets the language version; the `toolchain` directive pins the build.
- **Local toolchain install (architecture §6.4):** `go install golang.org/dl/go1.26.4@latest && go1.26.4 download` — gives the dev a `go1.26.4` binary alongside the system toolchain. Alternatively upgrade the system `go` if the dev prefers. The CI image / Docker builder pulls `golang:1.26-alpine` independently.
- **govulncheck assertion (architecture §6.5):** after the bump, `cd services/agent-board && govulncheck ./...` MUST exit clean (zero `Your code is affected by N vulnerabilities` lines). The two specific findings the bump targets:
  - GO-2026-5039 (`net/textproto`) — fixed in `1.26.4`.
  - GO-2026-5037 (`crypto/x509`) — fixed in `1.26.4`.
  Other findings flagged "vulnerabilities in modules you require but your code doesn't appear to call these" are noise per architecture §6.1 — acceptable.
- **If a new finding surfaces post-bump:** raise as a follow-up story per architecture §6.5 + R-1 — do NOT silently widen US012's scope.
- **`Dockerfile` line 9 (architecture §6.2):** `FROM golang:1.25-alpine AS build` → `FROM golang:1.26-alpine AS build`. The runtime stage (REQ005 D-010 distroless) stays unchanged.
- **No other `golang:<ver>` references** under `services/agent-board/` (verified during architecture authoring per §6.4 — only line 9 of `Dockerfile`).
- **`scripts/review/run-gate.sh` does NOT pin a Go version** (architecture §6.4) — no change needed in the gate script.
- **`docs/tech_debt.md` line 28** strike-through: re-locate by content match if numbering has drifted (the line mentions the Go toolchain / `crypto/x509` govulncheck finding).

## Definition of done
- `services/agent-board/go.mod` declares `go 1.26.4` + `toolchain go1.26.4`.
- `services/agent-board/Dockerfile` line 9 uses `golang:1.26-alpine`.
- `cd services/agent-board && govulncheck ./...` exits 0 (zero findings reachable from project code).
- `cd services/agent-board && go test ./...` passes.
- `cd services/agent-board && go build ./...` passes for both binaries.
- `golangci-lint run ./...` clean.
- `docs/tech_debt.md` line 28 strike-through applied.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` + `scripts/review/run-gate.sh cross` both `REVIEW GATE: PASS`.
- **Live e2e + 3-clean-run flake check REQUIRED** (architecture §10.1 — toolchain touches everything that ships): `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` clean THREE consecutive times.
- Dev set status to `in_review`; tech-lead approved.

## Notes
- **govulncheck output capture is the load-bearing evidence.** Tester / dev capture the pre-bump and post-bump `govulncheck ./...` output in the test report so the diff is auditable (architecture §6.5 + §10.1).
- **No story dependency.** US012 is independent of every other REQ006 story. It can land in any tick.

### Implementation notes (2026-06-07)

**Files touched:**
- `services/agent-board/go.mod` — `go 1.25.0` → `go 1.26.4`
- `services/agent-board/Dockerfile` line 9 — `golang:1.25-alpine` → `golang:1.26-alpine`
- `docs/tech_debt.md` line 28 — struck through with `→ fixed in REQ006/US012`

**Tests added:** None (per spec — existing suite is the regression vehicle).

**Test results (IT-001 through IT-004):**
- IT-001 `govulncheck ./...`: EXIT 0 — "Your code is affected by 0 vulnerabilities." GO-2026-5037 and GO-2026-5039 no longer reachable. 1 vulnerability in imported packages (noise per architecture §6.1). Clean.
- IT-002 `go test ./...`: 6 packages — `ok agent-board/cmd/api-server`, `ok agent-board/cmd/mcp-server`, `ok agent-board/internal/domain`, `ok agent-board/internal/handler`, `ok agent-board/internal/mcp`, `ok agent-board/internal/repo`. All pass.
- IT-003 `go build ./...`: EXIT 0 — both api-server and mcp-server compile clean.
- IT-004 `golangci-lint run ./...`: "No issues found." EXIT 0.
- `scripts/review/run-gate.sh be services/agent-board`: REVIEW GATE: PASS
- `scripts/review/run-gate.sh cross`: REVIEW GATE: PASS

**IT-005 live e2e (3 runs) — 2026-06-07 re-attempt with Podman running:**

Stack: Podman machine running (`podman machine list` shows `Currently running`); `podman-compose` v1.5.0 available. The compose images (api-server, mcp-server) were built using `golang:1.26-alpine` (Dockerfile line 9 as per US012 change). All 4 containers started successfully.

**Pre-existing infrastructure issue (not caused by US012):** `make e2e-up` has two bugs that prevent it from completing on a fresh stack:
1. The api-server health check polls `GET /api/v1/projects` which returns HTTP 500 until migrations run — but migrations run in `e2e-seed` which comes AFTER `e2e-up` completes. Circular dependency.
2. The mcp-server health check uses `curl http://localhost:8081/sse` without `--max-time`. The SSE endpoint returns HTTP 200 but holds the connection open indefinitely. `curl` hangs, the `grep` never receives input, and the `until` loop always times out after 120s.

Both bugs were present in the Makefile before US012 (since commit `1ba4793` / REQ005/US008). US012's go toolchain bump does not cause or affect these bugs. The go.mod and Dockerfile changes have been verified correct by tech-lead (Review pass 1).

**Workaround applied to get e2e evidence:** started the compose stack (`podman-compose up -d`), waited for postgres to become healthy, applied migrations and seed via `make e2e-seed`, verified api-server returns HTTP 200, then ran `make e2e-run` three consecutive times. `make e2e-down` was called after all three runs.

**IT-005 live e2e (3 runs):**

Run 1 — `make e2e-run` → Robot Framework output.xml `<stat pass="23" fail="0" skip="0">All Tests</stat>`:
**23 tests, 23 passed, 0 failed**

Run 2 — `make e2e-run` → Robot Framework output.xml `<stat pass="23" fail="0" skip="0">All Tests</stat>`:
**23 tests, 23 passed, 0 failed**

Run 3 — `make e2e-run` → Robot Framework output.xml `<stat pass="23" fail="0" skip="0">All Tests</stat>`:
**23 tests, 23 passed, 0 failed**

All three runs clean. No flakes. The `golang:1.26-alpine` builder image works correctly for both api-server and mcp-server binaries.

**`make e2e-down` ran successfully** after all three runs — volumes removed, containers stopped.

**Note for tech-lead (re: `make e2e-up` bug):** The `make e2e-up` mcp-server health check is a pre-existing Makefile infrastructure bug (curl hangs on SSE endpoint without `--max-time`). This affects every task requiring the full `make e2e-up && ... && make e2e-down` DoD sequence, not just US012. Recommending a follow-up Makefile fix (add `--max-time 5` to the mcp-server curl health check and fix the api-server circular dependency). Filed to tech_debt.md if tech-lead agrees.

**`toolchain go1.26.4` directive — spec discrepancy note:**
UT-001 asserts "a `toolchain go1.26.4` directive exists on the immediately following line." In practice, `go mod tidy` with go1.26.4 (the toolchain triggered by `go 1.26.4` in go.mod) removes the `toolchain go1.26.4` directive as redundant when `go == toolchain`. This is per Go 1.21+ module toolchain semantics documented at https://go.dev/doc/toolchain. The canonical and fully-equivalent form is `go 1.26.4` alone. Adding `toolchain go1.26.4` manually after tidy causes `go build ./...` to fail with "go: updates to go.mod needed; to update it: go mod tidy". The committed state (`go 1.26.4` only) is the correct Go-idiomatic form and satisfies the intent of architecture §6.2 (toolchain pinned to 1.26.4). Tech-lead: please confirm whether the UT-001 assertion about the `toolchain` directive should be waived given Go toolchain semantics, or whether a different `go`/`toolchain` version split (e.g. `go 1.26.0` + `toolchain go1.26.4`) was intended.

## Review log

### Review pass 1 — 2026-06-07 — verdict: changes_requested

**What passed (verified on review host):**
- `go.mod:3` — `go 1.26.4` ✓ (architecture §6.2 / D-007). See `toolchain` note below.
- `Dockerfile:9` — `FROM golang:1.26-alpine AS build` ✓ (UT-002); no other `golang:1.25-alpine` references remain ✓.
- `docs/tech_debt.md:28` — struck through with `→ fixed in REQ006/US012` ✓.
- IT-002 `go test ./...` — `301 passed in 7 packages`, exit 0 ✓.
- IT-003 `go build ./...` — `Go build: Success`, exit 0; both `cmd/api-server` and `cmd/mcp-server` compile ✓.
- IT-004 `golangci-lint run ./...` — PASS (via `run-gate.sh be`) ✓.
- `scripts/review/run-gate.sh be services/agent-board` — `REVIEW GATE: PASS` ✓
  (gosec + govulncheck WARN-skipped on review host — not installed; gate treats as WARN not FAIL, so gate is not blocked).
- `scripts/review/run-gate.sh cross` — `REVIEW GATE: PASS` ✓ (semgrep + gitleaks clean).
- IT-001 `govulncheck ./...` — NOT re-runnable on review host (`govulncheck not found`). Dev captured EXIT 0 / "0 vulnerabilities" / GO-2026-5037 + GO-2026-5039 no longer reachable in `## Notes`. Accepted as dev evidence; the orchestrator's Phase 3c test report should re-capture pre/post govulncheck diff on a govulncheck-enabled host per architecture §6.5.
- TDG conformance — commit subjects `green: ... (US012)` + `refactor: ... (US012)` follow tdg prefix + traceability convention ✓.

**BLOCKER — IT-005 (live e2e) evidence absent.**
The task DoD (line 67) and `US012_be_unit_tests.md` IT-005 require:
`make e2e-up && make e2e-seed && make e2e-run && make e2e-down` clean THREE consecutive times,
with the verbatim Robot Framework summary (`N tests, N passed, 0 failed`) for each run captured in `## Notes`
(architecture §10.1 — toolchain touches everything that ships, so the live-e2e + 3-clean-run flake bar applies).
The dev's `## Notes` (IT-005) states Docker was not found and NO e2e run was performed — zero of the three required runs,
and no Robot summary lines are present. This is a mandatory DoD gate and CANNOT be waived even though every other
check passed. Per the project anti-pattern rule, a missing mandatory DoD step is `changes_requested`, not a pass-with-note.
**Podman is now running on the review host**, so the e2e stack is available. Re-attempt: run the four-target e2e flow
THREE consecutive times and paste all three verbatim `N tests, N passed, 0 failed` summary lines into `## Notes`.

**`toolchain go1.26.4` directive question — RESOLVED, UT-001 assertion partially waived.**
The dev correctly identified that `go mod tidy` under go1.26.4 strips a redundant `toolchain go1.26.4` directive when
`go == toolchain`, and that manually re-adding it breaks `go build` with "updates to go.mod needed". Confirmed: **`go 1.26.4`
alone is the correct, idiomatic Go 1.21+ module form when the language version equals the desired toolchain** — it pins the
build to 1.26.4 exactly as architecture §6.2 intends. The committed `go.mod` (`go 1.26.4`, no `toolchain` line) is accepted.
The UT-001 sub-assertion "a `toolchain go1.26.4` directive exists on the immediately following line" is **waived** as
incompatible with Go toolchain semantics. This is a spec-text nuance, not a code defect, and does NOT contribute to the
changes_requested verdict — it is recorded here so the next pass is not re-flagged on it. (Filed to tech_debt for the
tester to reconcile the spec wording on next touch.)

**Tech-debt filed this pass:** 1 line appended to `docs/tech_debt.md` (UT-001 spec-wording reconciliation).

### Review pass 2 — 2026-06-07 — verdict: approved

**Scope of this pass:** the sole pass-1 blocker was missing IT-005 live-e2e evidence. Re-verified the implementation is still intact and the IT-005 evidence is now present.

**Implementation re-verified (unchanged from pass 1):**
- `go.mod:3` — `go 1.26.4` ✓ (architecture §6.2 / D-007). The `go 1.26.4`-only form (no separate `toolchain` line) was accepted/waived at pass 1 as the idiomatic Go 1.21+ shape — re-confirmed.
- `Dockerfile:9` — `FROM golang:1.26-alpine AS build` ✓; runtime distroless stage unchanged ✓.
- `docs/tech_debt.md:28` — struck through with `→ fixed in REQ006/US012` ✓.

**Test + gate results (re-run on review host):**
- `cd services/agent-board && go test ./...` — `Go test: 301 passed in 7 packages`, exit 0 ✓ (IT-002).
- `scripts/review/run-gate.sh be services/agent-board` — `REVIEW GATE: PASS` ✓
  (gofmt -s / go vet / golangci-lint / go test all PASS; gosec + govulncheck WARN-skipped on review host — not installed; gate treats as WARN not FAIL, so gate is not blocked).
- `scripts/review/run-gate.sh cross` — `REVIEW GATE: PASS` ✓ (semgrep owasp/golang/typescript + gitleaks clean).

**IT-005 live e2e evidence — NOW PRESENT and ACCEPTED.**
The dev's `## Notes` (IT-005, 2026-06-07 re-attempt) contains the three required verbatim Robot Framework summary lines, all clean:
- Run 1 — **23 tests, 23 passed, 0 failed**
- Run 2 — **23 tests, 23 passed, 0 failed**
- Run 3 — **23 tests, 23 passed, 0 failed**
No flakes across the three consecutive runs. The compose images were built from `golang:1.26-alpine` (the US012 Dockerfile change), confirming the new builder image works for both api-server and mcp-server binaries.

**Workaround accepted (per human direction + documented pre-existing bugs).** The exact DoD sequence `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` could not run because `make e2e-up` has two pre-existing infrastructure bugs (api-server health-check circular dependency on migrations; mcp-server SSE curl hang with no `--max-time`). Both bugs predate US012 (since REQ005/US008 commit `1ba4793`) and are NOT caused by the go toolchain bump. They are filed in `docs/tech_debt.md:113` (REQ006/US012). The equivalent workaround (`podman-compose up -d` + `make e2e-seed` + `make e2e-run` ×3 + `make e2e-down`) was explicitly accepted by the human as equivalent IT-005 evidence. This is an accepted deviation, not a waived mandatory gate — the three-clean-run flake bar (architecture §10.1) was met; only the orchestration entrypoint differed.

**TDG conformance:** pass-2 rework commits `39310e6 refactor: chore: hand off go toolchain bump 1.26.4 for review (US012)` carries the `(US012)` traceability tag; `9001ddd chore: document make e2e-up health-check bugs in tech_debt.md` is a tech_debt.md documentation commit (not test/production code). The `refactor: chore:` prefix drift on hand-off commits is the same recurring housekeeping-prefix pattern already filed for REQ006/US001, US004, US007 (tech_debt.md:105,107,110) — red→green→refactor ordering is honored; tolerated as the established precedent on this REQ, not a new `changes_requested`.

**Tech-debt filed this pass:** none filed this pass — the two relevant items (UT-001 spec wording; `make e2e-up` health-check bugs) were already filed at pass 1 / by the dev (tech_debt.md:112-113). No new non-blocking findings.
