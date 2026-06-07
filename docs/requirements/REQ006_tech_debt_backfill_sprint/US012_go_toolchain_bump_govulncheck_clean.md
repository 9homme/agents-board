# US012 — Go toolchain bump to clear transitive `crypto/x509` govulncheck finding

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** done

## Story
As a **security-conscious operator running `govulncheck` on every push**, I want **the Go toolchain in `services/agent-board/go.mod` bumped to the next minor version that contains the fix for the transitive `crypto/x509` finding**, so that `cd services/agent-board && govulncheck ./...` returns clean and the BE quality gate does not silently soft-warn on a known stdlib vulnerability.

## Acceptance criteria

- **Scenario: Go toolchain version is bumped in `go.mod`**
  - Given `services/agent-board/go.mod` currently declares `go 1.25.0` (verified at intake)
  - When this story is complete
  - Then `services/agent-board/go.mod` declares a Go version that is the **lowest released minor version** that fixes the stdlib `crypto/x509` finding flagged by `govulncheck` against the current `go 1.25.0`
  - And the exact target version is determined by the system-architect (OQ-2 in README) and documented in `architecture.md`
  - And the `toolchain` directive (if present) is set consistently with the new `go` directive

- **Scenario: `govulncheck` is clean after the bump**
  - Given the new Go toolchain is installed locally (the tech-lead / dev runs `go install golang.org/dl/go<NEW>@latest` and `go<NEW> download` as needed)
  - When `cd services/agent-board && govulncheck ./...` runs
  - Then the output is clean — specifically, the previously-flagged `crypto/x509` finding (per `docs/tech_debt.md` line 28) is no longer reported
  - And no new findings have been introduced (if the bump surfaces a new finding, raise as a follow-up story or fold into this one with a `Notes for the team` entry)

- **Scenario: all existing tests pass under the new toolchain**
  - Given the new Go toolchain
  - When `cd services/agent-board && go test ./...` runs
  - Then all tests pass (BE unit + BE integration)
  - And `golangci-lint run ./...` is clean
  - And `cd services/agent-board && go build ./...` succeeds for both binaries (`api-server` and `mcp-server`)

- **Scenario: Dockerfile builder image is updated to match**
  - Given the existing Dockerfile(s) under `services/agent-board/` (or repo root, per current layout) reference a `golang:<version>` builder image
  - When this story is complete
  - Then the Dockerfile builder image tag is updated to a version compatible with the new `go.mod` directive (e.g. `golang:1.<NEW>-alpine` or `golang:1.<NEW>-bookworm` per current convention)
  - And the runtime image (distroless per REQ005 D-010) is unchanged
  - And `make e2e-up` continues to succeed (compose builds + starts cleanly)

- **Scenario: CI / quality-gate continues to pass**
  - Given `scripts/review/run-gate.sh` (or whatever the current BE gate is)
  - When the gate runs end-to-end
  - Then it passes green
  - And the `govulncheck` step (per REQ005/US003 soft-warn behaviour) either passes silently OR is no longer needed to soft-warn the `crypto/x509` finding
  - And the gate's `make e2e-run` smoke succeeds (no regression)

- **Scenario: closes tech-debt finding**
  - Given `docs/tech_debt.md` line 28 contains the finding `pre-existing govulncheck finding on stdlib crypto/x509 ... bump Go toolchain or pin a fix when a runtime upgrade is available`
  - When this story is `done`
  - Then `docs/tech_debt.md` line 28 is struck through with `→ fixed in REQ006/US012`

- **Scenario: documented in architecture**
  - Given `architecture.md` (REQ006) is being authored as part of Phase 1
  - When this story's AC is being planned
  - Then `architecture.md` documents the chosen target Go version, the rationale (which CVE / advisory the bump addresses), and any operator-visible consequence

## UI / UX flow expectations
**No UI: BE-prod (toolchain) only.** Operational expectations:

- **No runtime behaviour change** beyond the underlying Go stdlib's bug fixes.
- **Deployment compatibility:** the runtime image (distroless) is unchanged; only the builder image changes. Existing `make e2e-up` flow is unaffected.

## Out of scope
- **Bumping `go.mod` dependencies** (e.g. `github.com/labstack/echo/v4`, `github.com/jackc/pgx/v5`, `github.com/stretchr/testify`). Dependency bumps are a separate concern; raise as new debt if needed.
- **Bumping Go major version** (1.x → 2.x). When Go 2 ships, that's a new REQ.
- **Adding a `gosec` or `staticcheck` policy change.** Quality-gate behaviour stays the same.
- **Multi-version compatibility matrix** (testing against 1.24 and 1.<NEW>). Pick one target; tests run against that.

## Dependencies
- None directly. The architect resolves OQ-2 (target version) in `architecture.md` before tech-lead plans this story.

## Notes for the team

- **Architect picks the exact version.** po-ba does not pick the version because the choice depends on (a) which Go release fixes the `crypto/x509` finding, (b) which version is `govulncheck`-clean against the rest of the dep tree, and (c) which version is supported by `golangci-lint` and the Docker image we use. The architect is in the best position to resolve.
- **Audit reference.** `docs/tech_debt.md` line 28: `services/agent-board/cmd/mcp-server/sse.go,message.go — pre-existing govulncheck finding on stdlib crypto/x509`.
- **Local verification commands:**
  - `cd services/agent-board && go install golang.org/x/vuln/cmd/govulncheck@latest`
  - `cd services/agent-board && govulncheck ./...`
  - `cd services/agent-board && go test ./...`
  - `cd services/agent-board && go build ./...`
  - `make e2e-up && make e2e-run && make e2e-down`
- **Closes tech-debt.** Strike `docs/tech_debt.md` line 28 in the same commit (or sign-off commit) as the bump.
- **Rollback plan.** If the bump surfaces an unexpected regression, the rollback is `git revert` on the `go.mod` / `Dockerfile` commit — the change is self-contained.

## Sign-off log
(po-ba appends here on each sign-off pass)

### Sign-off pass 1 — 2026-06-07 — verdict: approved

**Spec review (US012_be_unit_tests.md — no e2e spec file; existing suite reused per architecture §1.2):**
- Every AC scenario maps to at least one UT-*/IT-* case:
  - "go.mod bumped" → UT-001 (with `toolchain go1.26.4` sub-assertion **waived** — `go 1.26.4` alone is the correct, idiomatic Go 1.21+ module form when language version == desired toolchain; manually re-adding the directive breaks `go build`. Confirmed by tech-lead Review pass 1, documented in tech_debt.md. Spec-text nuance, not a code defect — does not affect verdict).
  - "Dockerfile builder image" → UT-002.
  - "govulncheck clean, GO-2026-5037 (`crypto/x509`) + GO-2026-5039 (`net/textproto`) no longer reachable" → IT-001.
  - "all tests pass under new toolchain" → IT-002.
  - "go build both binaries" → IT-003.
  - "golangci-lint clean" → IT-004.
  - "Dockerfile/e2e regression, 3-clean-run flake bar" → IT-005.
  - "closes tech-debt line 28" → struck through with `→ fixed in REQ006/US012`.
  - "documented in architecture" → architecture §6 cited throughout the task.
- No edge case or error path implied by the AC is skipped. Pyramid is honest — a toolchain touch is genuinely an integration/e2e-level regression concern, not a unit one.

**Result review (US012_test_report.md, commit `6fa0726`):**
- 7 test IDs (UT-001, UT-002, IT-001–IT-005) — all PASS. Counts match the spec; no silent dropping.
- `go test ./...` — 301 tests, 0 failures, 7 packages.
- govulncheck — EXIT 0; both target stdlib CVEs confirmed no longer reachable from project code.
- E2E — 3 consecutive runs, 23/23/23 passed, 0 failed each; no flakes. Compose images rebuilt from `golang:1.26-alpine`.
- **Skipped tests: none.** No `t.Skip`, no `[Tags] skip`.
- **IT-005 workaround accepted.** The literal `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` sequence could not run because `make e2e-up` has two pre-existing infrastructure bugs (api-server health-check circular dependency on migrations; mcp-server SSE curl hang with no `--max-time`), both predating US012 (REQ005/US008 commit `1ba4793`) and documented in tech_debt.md. The equivalent flow (`podman-compose up -d` + `make e2e-seed` + `make e2e-run` ×3 + `make e2e-down`) met the architecture §10.1 three-clean-run flake bar; only the orchestration entrypoint differed. Per human direction, accepted as equivalent IT-005 evidence — an accepted deviation, not a waived mandatory gate.

**Routed to:** none — story approved, `Status: done`.
