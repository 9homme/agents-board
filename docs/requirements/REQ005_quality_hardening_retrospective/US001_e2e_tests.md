# US001 — E2E test specification

**Owner:** tester.

## Why e2e does not apply

US001 fixes a `printf` bug in `scripts/review/run-gate.sh`. The behaviour is entirely script-level:

- The gate script is not a web service and has no HTTP surface to exercise.
- It is not a React component, so the Browser library is irrelevant.
- The RequestsLibrary has no applicable use against a local shell script.
- The full correctness contract is verifiable through shell-harness integration tests (IT-US001-001 through IT-US001-004 in `US001_be_unit_tests.md`) that invoke the script in a non-TTY context and grep the output.

Promoting this to an e2e scenario would require the live docker-compose stack (from US008) to be running, and Robot Framework would still need to shell-exec the script — which is exactly what IT-US001-001/002 already do at the script-harness layer. There is no additional end-user-observable behaviour to verify at the e2e layer.

**Verdict: No e2e scenarios. Script-level concern; covered by IT-US001-001 through IT-US001-004.**
