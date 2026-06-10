# US035 — Backend unit & integration test specification
# Go toolchain bump to `go 1.26.4`

**For BE Dev:** this story is a toolchain and Dockerfile edit, not a test-file story. The "test spec" is the govulncheck-clean smoke + regression bar defined below. Production code changes: `go.mod` (version directive + toolchain directive) and `services/agent-board/Dockerfile` (builder image). No new `*_test.go` files are created by this story; the existing test suite is the regression vehicle.

## Coverage matrix

| AC scenario | Layer | Test ID | File / command | What it asserts |
|---|---|---|---|---|
| `go.mod` declares `go 1.26.4` + `toolchain go1.26.4` | unit | UT-001 | `go.mod` inspection | version directives correct |
| `Dockerfile` builder image updated to `golang:1.26-alpine` | unit | UT-002 | `Dockerfile:9` inspection | builder image tag updated |
| `govulncheck ./...` exits clean (zero reachable findings) | integration | IT-001 | `govulncheck` | no CVEs reachable from project code |
| `go test ./...` passes under new toolchain | integration | IT-002 | `go test` | regression suite green |
| `go build ./...` succeeds for both binaries | integration | IT-003 | `go build` | binaries compile cleanly |
| `golangci-lint run ./...` clean | integration | IT-004 | `golangci-lint` | no new lint issues from toolchain bump |
| `make e2e-up && make e2e-run && make e2e-down` passes | integration | IT-005 | `make` e2e flow | compose stack + e2e suite unaffected |

## Unit tests (structural assertions — BE Dev verifies during implementation)

### UT-001 — `go.mod` version directives
- **File:** `services/agent-board/go.mod`
- **Assert:** line 3 reads `go 1.26.4` (exact string)
- **Assert:** a `toolchain go1.26.4` directive exists on the immediately following line
- **Architecture cite:** architecture.md §6.2 (D-007); target version locked to `go 1.26.4`

---

### UT-002 — Dockerfile builder image
- **File:** `services/agent-board/Dockerfile`
- **Assert:** line 9 reads `FROM golang:1.26-alpine AS build` (exact string — minor-tracking tag, not patch-pinned)
- **Assert:** no other `FROM golang:1.25-alpine` references remain in the file
- **Architecture cite:** architecture.md §6.2; §6.4

## Integration tests

### IT-001 — `govulncheck` clean post-bump
- **Precondition:** dev has installed `go1.26.4` locally (`go install golang.org/dl/go1.26.4@latest && go1.26.4 download`)
- **Command:** `cd services/agent-board && govulncheck ./...`
- **Expect:**
  - Exit code 0
  - Zero lines matching `Your code is affected by N vulnerabilities` in the output
  - Specifically: `GO-2026-5039` (`net/textproto`) is no longer reported as reachable
  - Specifically: `GO-2026-5037` (`crypto/x509`) is no longer reported as reachable
- **If a new finding surfaces:** do NOT fold the fix into US035. Raise a follow-up story per architecture.md §6.5.
- **Architecture cite:** architecture.md §6.1; §6.5

---

### IT-002 — `go test ./...` regression under new toolchain
- **Command:** `cd services/agent-board && go test ./...`
- **Expect:** all tests pass (same outcome as pre-bump)
- **Architecture cite:** architecture.md §10.1 (3-clean-run bar); US035 AC "all existing tests pass"

---

### IT-003 — `go build ./...` succeeds
- **Command:** `cd services/agent-board && go build ./...`
- **Expect:** zero build errors for both `api-server` and `mcp-server` binaries
- **Architecture cite:** US035 AC "go build ./... succeeds"

---

### IT-004 — `golangci-lint` clean
- **Command:** `cd services/agent-board && golangci-lint run ./...`
- **Expect:** no new lint findings introduced by the toolchain bump
- **Architecture cite:** architecture.md §10.1

---

### IT-005 — e2e stack regression
- **Command:** `make e2e-up && make e2e-seed && make e2e-run && make e2e-down` (3 consecutive clean runs per architecture.md §10.1)
- **Expect:** all Robot Framework tests pass; compose stack builds cleanly with the new `golang:1.26-alpine` builder image
- **Architecture cite:** architecture.md §6.4; §10.1 live-e2e mandate for production-code touches

## Coverage exemptions

N/A — this story does not add test files. The regression bar is the full existing test suite passing under the new toolchain plus `govulncheck` clean.
