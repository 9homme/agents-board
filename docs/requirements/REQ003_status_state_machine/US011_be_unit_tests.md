# US011 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). This story is a quality refinement — there is no new production behaviour to add, so no new behavioural unit tests are required. Instead, each UT-* below corresponds to a static-analysis or grep-style verification that a specific lint finding category has been resolved correctly. Implement all verification commands inside `services/agent-board/`.

> **Note — baseline drift acknowledgement:** Before starting any fixes, re-run `golangci-lint run ./...` from inside `services/agent-board/` and record the per-linter counts. If they differ from the 2026-05-19 baseline (noctx=11, unused=9, gocritic=5, errcheck=4, errorlint=3, gosec=1, revive=1), document the drift in a comment at the top of your PR. The tester does not gate on the exact per-linter counts; the gate is the **final state** (zero findings). Confirming or documenting drift is the dev's responsibility, not a blocker.

## Coverage matrix

| AC scenario | Layer | Test ID | Service / package | Verification method |
|---|---|---|---|---|
| Lint exits clean | static analysis | UT-001 | services/agent-board | `golangci-lint run ./...` exits 0, zero findings |
| Race tests pass | runtime | UT-002 | services/agent-board | `go test ./... -race` exits 0, no race reports |
| noctx findings resolved | static analysis | UT-003 | services/agent-board / internal/handler, internal/repo | zero `noctx` findings; `httptest.NewRequest(` absent from non-fixture test files; `db.Ping(` absent from production code |
| unused findings resolved | static analysis | UT-004 | services/agent-board / internal/handler | zero `unused` findings; dead helpers in `handler_test.go` either removed or provably referenced |
| gocritic findings resolved | static analysis | UT-005 | services/agent-board / internal/repo | zero `gocritic` findings; sloppyReassign patterns rewritten |
| errcheck findings resolved | static analysis | UT-006 | services/agent-board / internal/repo | zero `errcheck` findings; `tx.Rollback()` returns explicitly discarded with justification |
| errorlint findings resolved | static analysis | UT-007 | services/agent-board / internal/repo | zero `errorlint` findings; `errors.Is(err, sql.ErrNoRows)` used throughout |
| gosec finding resolved | static analysis | UT-008 | services/agent-board / cmd | zero `gosec` findings; log-injection addressed by sanitisation or justified suppression |
| revive finding resolved | static analysis | UT-009 | services/agent-board | zero `revive` findings |
| Suppression hygiene | static analysis | UT-010 | services/agent-board | all remaining `// nolint` directives name linter(s) and carry a justification; no blanket `// nolint`; no file-level disables |

## Unit tests

### UT-001 — Lint exits clean

- **Service:** `services/agent-board`
- **Layer:** static analysis
- **Command:** `golangci-lint run ./...` executed from inside `services/agent-board/`
- **Given:** The v2 `.golangci.yml` at the repo root is present and unchanged.
- **When:** The command completes.
- **Then:**
  - Exit status is 0.
  - Standard output contains zero finding lines (no lines matching the pattern `<filename>:<line>:<col>:`).
- **How to capture for test report:** Redirect stdout + stderr into `golangci-lint-output.txt`; assert the file contains no lines matching `^\S.*:\d+:\d+:`.
- **Architecture cite:** US011 AC "Lint exits clean".

---

### UT-002 — Race tests pass

- **Service:** `services/agent-board`
- **Layer:** runtime
- **Command:** `go test ./... -race` executed from inside `services/agent-board/`
- **Given:** All existing tests for US008/US009/US010 are present and unmodified in behaviour.
- **When:** The command completes.
- **Then:**
  - Exit status is 0.
  - Output contains `ok` for every package line; no `FAIL` lines.
  - Output contains no `DATA RACE` blocks.
- **How to capture for test report:** Redirect full output into `go-test-race-output.txt`; assert no line starts with `--- FAIL` and no line contains `DATA RACE`.
- **Watch-out note:** After applying `noctx` fixes that introduce a context into `httptest.NewRequestWithContext`, re-run this command specifically before committing the noctx batch — adding a real context with a deadline can occasionally surface timeout-related races under `-race`.
- **Architecture cite:** US011 AC "No behaviour change in production code under -race".

---

### UT-003 — noctx findings resolved

- **Service:** `services/agent-board`
- **Layer:** static analysis (grep-level contract)
- **Baseline count:** 11 findings.
- **What a passing fix looks like:**
  - Every `httptest.NewRequest(` call site in test files under `internal/handler/` is replaced with `httptest.NewRequestWithContext(t.Context(), ...)` or the request is followed by `.WithContext(ctx)` before it is passed to the handler.
  - The `db.Ping()` call site(s) in production code (likely `cmd/agent-board/main.go` or a DB initialisation helper) are replaced with `db.PingContext(ctx)` where `ctx` is a `context.Background()` or a startup context.
- **Verification:**
  - `golangci-lint run --enable=noctx ./...` reports zero findings.
  - `grep -rn "httptest\.NewRequest(" services/agent-board/` should show no hits outside of test fixture helpers that already receive a context parameter (i.e., no bare two-argument form).
  - `grep -rn "\.Ping()" services/agent-board/internal/ services/agent-board/cmd/` returns 0 lines.
- **Edge cases:** Anywhere a helper function constructs a request internally, the context must be threaded in — not just the call site the linter flagged directly.
- **Architecture cite:** US011 AC "Specific finding categories are resolved correctly — noctx (11)".

---

### UT-004 — unused findings resolved

- **Service:** `services/agent-board`
- **Layer:** static analysis
- **Baseline count:** 9 findings (most concentrated in `internal/handler/handler_test.go`).
- **What a passing fix looks like:**
  - Each unused symbol is either: (a) deleted because it is genuinely dead code left over from the SSE-race fix, or (b) referenced from at least one test case so the linter no longer considers it unused.
  - Suppression via `// nolint:unused` is a last resort and requires a justification comment — prefer deletion.
- **Verification:**
  - `golangci-lint run --enable=unused ./...` reports zero findings.
  - If any `// nolint:unused` suppressions were added: `grep -n "nolint:unused" services/agent-board/` shows each line is immediately preceded by or shares a line with a justification comment.
- **Suggested triage order (per story AC):** Review the `unused` cluster before any other category; prune dead helpers first; only suppress what has a genuine reason to exist but cannot be referenced in a test build.
- **Architecture cite:** US011 AC "unused cluster in handler_test.go is triaged first".

---

### UT-005 — gocritic findings resolved

- **Service:** `services/agent-board`
- **Layer:** static analysis
- **Baseline count:** 5 findings (`sloppyReassign` and related checks in the repo layer).
- **What a passing fix looks like:**
  - `sloppyReassign`: a pattern such as `x, err = f()` followed immediately by `x, err = g()` where the first result `x` is overwritten before use is rewritten so each assignment is on a separate statement with its intermediate result used or explicitly discarded before the next assignment proceeds.
  - Other `gocritic` subchecks: apply the idiomatic rewrite suggested in the finding message. No suppression unless the linter is producing a demonstrable false positive for that specific subcheck.
- **Verification:**
  - `golangci-lint run --enable=gocritic ./...` reports zero findings.
- **Architecture cite:** US011 AC "Specific finding categories are resolved correctly — gocritic (5)".

---

### UT-006 — errcheck findings resolved

- **Service:** `services/agent-board`
- **Layer:** static analysis
- **Baseline count:** 4 findings (ignored `tx.Rollback()` returns in `internal/repo/user_story_repo.go` and `internal/repo/task_repo.go`).
- **What a passing fix looks like:** One of the two forms below, applied consistently across all four sites:
  - **Form A — explicit discard with justification (acceptable, conventional Go pattern):**
    ```go
    // Rollback is best-effort cleanup after an earlier error; the original
    // error is already being returned to the caller.
    _ = tx.Rollback()
    ```
  - **Form B — handle the rollback error (preferred if rollback failures should be visible in logs):**
    ```go
    if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
        // log or wrap, but still return the original error
    }
    ```
- **Note on `.golangci.yml`:** `errcheck.check-blank: true` is enabled in the project config. This means `_ = tx.Rollback()` is also checked. Verify against the installed linter version: if `check-blank` rejects Form A, Form B is required. The dev must confirm this before submitting.
- **Verification:**
  - `golangci-lint run --enable=errcheck ./...` reports zero findings.
  - `grep -n "tx\.Rollback()" services/agent-board/internal/repo/` — every match is either `_ = tx.Rollback()` or inside an `if rbErr :=` block; no bare `tx.Rollback()` statement on its own line.
- **Architecture cite:** US011 AC "Specific finding categories are resolved correctly — errcheck (4)".

---

### UT-007 — errorlint findings resolved

- **Service:** `services/agent-board`
- **Layer:** static analysis
- **Baseline count:** 3 findings (`err == sql.ErrNoRows` comparisons in the repo layer).
- **What a passing fix looks like:**
  ```go
  // Before (flagged by errorlint):
  if err == sql.ErrNoRows {

  // After (correct):
  if errors.Is(err, sql.ErrNoRows) {
  ```
  The `errors` standard library package must be imported in each file where this pattern appears (it may already be imported for other uses).
- **Verification:**
  - `golangci-lint run --enable=errorlint ./...` reports zero findings.
  - `grep -rn "== sql\.ErrNoRows" services/agent-board/` returns 0 lines.
  - `grep -rn "!= sql\.ErrNoRows" services/agent-board/` returns 0 lines (the inequality form carries the same wrapping risk and must be converted to `!errors.Is(...)`).
- **Architecture cite:** US011 AC "Specific finding categories are resolved correctly — errorlint (3)".

---

### UT-008 — gosec finding resolved

- **Service:** `services/agent-board`
- **Layer:** static analysis
- **Baseline count:** 1 finding (log-injection warning in `cmd/agent-board/main.go` or equivalent `cmd/*/main.go`).
- **What a passing fix looks like:** One of two forms:
  - **Form A — sanitise the logged value** by stripping or replacing newline characters and other control characters before passing the value to the logger, eliminating the injection vector.
  - **Form B — justified suppression** where the input is provably safe (e.g. comes from a trusted env var whose valid values cannot contain newlines):
    ```go
    // nolint:gosec // G104: addr derives from APP_PORT which is validated to be a port number; newline injection is not possible.
    log.Printf("listening on %s", addr)
    ```
- **Verification:**
  - `golangci-lint run --enable=gosec ./...` reports zero findings.
  - If Form B was used: `grep -n "nolint:gosec" services/agent-board/cmd/` shows the suppression line; manual inspection confirms a justification comment is present on the same or immediately preceding line.
- **Architecture cite:** US011 AC "Specific finding categories are resolved correctly — gosec (1)"; US011 AC "Any remaining suppressions are explicit and justified".

---

### UT-009 — revive finding resolved

- **Service:** `services/agent-board`
- **Layer:** static analysis
- **Baseline count:** 1 finding.
- **What a passing fix looks like:** The specific `revive` rule triggered (one of the rules enabled in `.golangci.yml`: `exported`, `var-naming`, `package-comments`, `context-as-argument`, `error-return`, `error-naming`) is addressed idiomatically:
  - `exported` / `package-comments`: add or correct the missing doc comment on the exported symbol or package declaration.
  - `var-naming`: rename to the conventional Go form (e.g., `id` not `ID` for unexported, `ID` not `Id` for exported acronyms).
  - `context-as-argument`: move the `context.Context` parameter to be the first argument of the function signature.
  - `error-return`: reorder return values so `error` is last.
  - `error-naming`: rename the error variable or type to the idiomatic form (e.g., `ErrFoo` for sentinel errors).
  - No suppression unless the finding is a demonstrable false positive for this specific rule.
- **Verification:**
  - `golangci-lint run --enable=revive ./...` reports zero findings.
- **Architecture cite:** US011 AC "Specific finding categories are resolved correctly — revive (1)".

---

### UT-010 — Suppression hygiene

- **Service:** `services/agent-board`
- **Layer:** static analysis (grep contract enforced at tech-lead review gate)
- **Contract:** After the story is complete, every `// nolint` directive remaining anywhere under `services/agent-board/` must satisfy all three rules:
  1. Names at least one specific linter: `// nolint:gosec` is valid; `// nolint` alone is not.
  2. Has a justification comment on the same line (after the directive) or on the immediately preceding line.
  3. Is not a file-level pragma silencing an entire file (no `// nolint` or `//nolint` appearing as the first line of a `.go` file with the intent to suppress all findings in that file).
- **Verification (tech-lead review gate):**
  - `grep -rn "// nolint$" services/agent-board/` returns 0 lines.
  - `grep -rn "//nolint$" services/agent-board/` returns 0 lines.
  - Manual scan: each line produced by `grep -rn "nolint:" services/agent-board/` has a justification either on the same line following the directive or on the line immediately above it.
- **Architecture cite:** US011 AC "Any remaining suppressions are explicit and justified".
