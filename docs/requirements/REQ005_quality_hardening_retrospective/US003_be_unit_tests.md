# US003 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). This story edits `scripts/review/run-gate.sh` and `scripts/review/README.md`. No Go packages are touched. Tests are shell-harness tests in `scripts/review/test/`, asserting soft-warn behaviour under tool-absent conditions.

## Coverage matrix

| AC scenario | Layer | Test ID | Script | Behaviour under test |
|---|---|---|---|---|
| BE gate completes when `gosec` is absent | integration | IT-US003-001 | `scripts/review/run-gate.sh` | soft-warn WARN line emitted; gate does not exit 2; continues |
| BE gate completes when `govulncheck` is absent | integration | IT-US003-002 | `scripts/review/run-gate.sh` | soft-warn WARN line emitted; gate does not exit 2; continues |
| When both ARE installed, behaviour unchanged (no spurious WARN) | integration | IT-US003-003 | `scripts/review/run-gate.sh` | `gosec` and `govulncheck` run_check steps execute; no WARN line |
| `go` and `golangci-lint` remain hard-required (exit 2) | integration | IT-US003-004 | `scripts/review/run-gate.sh` | missing `go` causes exit 2; missing `golangci-lint` causes exit 2 |
| Cross and FE gates unaffected | integration | IT-US003-005 | `scripts/review/run-gate.sh` | soft-warn only in `gate_be()`; cross/fe tracks unchanged |
| README documents hard vs soft-warn tools | integration | IT-US003-006 | `scripts/review/README.md` | file contains required subsection content |

## Integration tests

### IT-US003-001 — BE gate soft-warns when `gosec` is absent

- **Script under test:** `scripts/review/run-gate.sh`
- **Boundary:** full `be services/agent-board` gate invocation with `gosec` absent from PATH
- **Setup:**
  - Override PATH to exclude any `gosec` binary: `PATH=$(echo $PATH | tr ':' '\n' | grep -v gosec | tr '\n' ':')`.
  - Ensure `govulncheck` IS on PATH (use the real binary or a stub that exits 0).
  - Invoke `bash scripts/review/run-gate.sh be services/agent-board 2>&1 | cat`.
- **When:** gate runs `gate_be()` with `gosec` absent
- **Then:**
  - Exit code is 0 or 1 (gate pass or fail from other checks) — NOT 2 (MISSING TOOL).
  - Output contains a line matching `WARN.*gosec.*skipped` (case-insensitive substring match acceptable).
  - The WARN line does NOT indicate a gate failure — it is advisory only.
  - `golangci-lint`, `go vet`, `go test`, and `govulncheck` steps still execute (their output lines appear in captured output).
- **Architecture cite:** architecture §2 US003 row — inline `command -v` check with WARN print + skip for `gosec`.

### IT-US003-002 — BE gate soft-warns when `govulncheck` is absent

- **Script under test:** `scripts/review/run-gate.sh`
- **Boundary:** full `be services/agent-board` gate invocation with `govulncheck` absent from PATH
- **Setup:**
  - Override PATH to exclude `govulncheck`.
  - Ensure `gosec` IS on PATH (use the real binary or a stub that exits 0).
  - Invoke `bash scripts/review/run-gate.sh be services/agent-board 2>&1 | cat`.
- **When:** gate runs `gate_be()` with `govulncheck` absent
- **Then:**
  - Exit code is 0 or 1 — NOT 2.
  - Output contains a line matching `WARN.*govulncheck.*skipped` (or `not installed`).
  - The line also mentions the install one-liner (`go install golang.org/x/vuln`) per architecture.
  - All other BE gate steps still execute.
- **Architecture cite:** architecture §2 US003 row; US003 AC "Scenario: BE gate completes when `govulncheck` is absent".

### IT-US003-003 — When both tools ARE installed, no WARN line appears

- **Script under test:** `scripts/review/run-gate.sh`
- **Boundary:** BE gate invocation with both `gosec` and `govulncheck` on PATH (stub binaries that exit 0 with empty output are acceptable)
- **Setup:**
  - Create temporary stub binaries `gosec` and `govulncheck` that `exit 0` in a temp directory.
  - Prepend that directory to PATH.
  - Invoke `bash scripts/review/run-gate.sh be services/agent-board 2>&1 | cat`.
- **When:** gate runs with both tools present
- **Then:**
  - No line matching `WARN.*gosec` or `WARN.*govulncheck` appears in the output.
  - Both `gosec` and `govulncheck` `run_check` execution lines appear (i.e. the tools were invoked, not skipped).
- **Architecture cite:** US003 AC "Scenario: when both binaries ARE installed, behaviour unchanged".

### IT-US003-004 — `go` and `golangci-lint` remain hard-required (exit 2)

- **Script under test:** `scripts/review/run-gate.sh`
- **Boundary:** BE gate invocation with `go` absent from PATH
- **Setup:**
  - Override PATH to exclude `go`.
- **When:** `bash scripts/review/run-gate.sh be services/agent-board`
- **Then:** exit code is 2 with a MISSING TOOL message.
- **Repeat:** same test with `golangci-lint` absent.
- **Architecture cite:** US003 AC "Scenario: `go` and `golangci-lint` remain hard-required".

### IT-US003-005 — Cross and FE gate tracks unaffected

- **Script under test:** `scripts/review/run-gate.sh`
- **Boundary:** static content — grep for `require_tool` usage in `gate_cross()` and `gate_fe()` sections
- **Setup:** Read the script; no invocation needed.
- **When:** grep for `require_tool` outside the `gate_be()` function body
- **Then:**
  - `semgrep`, `gitleaks`, and `npm` `require_tool` calls are present and unmodified in the `gate_cross()` and `gate_fe()` sections.
  - No `gosec` or `govulncheck` soft-warn logic appears outside `gate_be()`.
- **Architecture cite:** US003 AC "Scenario: cross-cutting and FE gates unaffected".

### IT-US003-006 — README documents hard vs soft-warn distinction

- **Script under test:** `scripts/review/README.md`
- **Boundary:** static file content assertion
- **Setup:** Read the file.
- **When:** inspected after the story
- **Then:**
  - The README contains a subsection titled "Soft-warn vs. hard-required" (or equivalent heading) under "What it runs".
  - The section lists hard-required tools: `go`, `golangci-lint`, `npm`, `semgrep`, `gitleaks`.
  - The section lists soft-warn tools: `gosec`, `govulncheck`.
  - The section includes install one-liners for `gosec` and `govulncheck`.
  - The section explains that `gosec` rules are covered by `golangci-lint`'s built-in gosec linter.
- **Architecture cite:** architecture §2 US003 row — README update; US003 AC "Scenario: README documents the substitution".
