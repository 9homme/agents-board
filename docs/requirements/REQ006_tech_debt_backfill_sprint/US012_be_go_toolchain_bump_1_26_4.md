# US012/be_go_toolchain_bump_1_26_4

**Requirement:** REQ006
**Story:** US012
**Track:** BE
**Service:** services/agent-board
**Status:** in_progress
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

## Review log
