# US003 — E2E test specification

**Owner:** tester.

## Why e2e does not apply

US003 replaces two `require_tool` calls in `scripts/review/run-gate.sh` with inline `command -v` checks. The entire behaviour is at the shell-script level:

- The gate script is not a web service; the docker-compose stack (US008) is irrelevant.
- The Browser library has no surface to exercise.
- The RequestsLibrary has no applicable HTTP endpoints.
- The correctness guarantees (soft-warn emitted; exit code is 0 or 1 not 2; hard tools remain hard; README content) are all static-file or shell-execution assertions, fully covered by IT-US003-001 through IT-US003-006 in `US003_be_unit_tests.md`.

**Verdict: No e2e scenarios. Script-level concern; covered by IT-US003-001 through IT-US003-006.**
