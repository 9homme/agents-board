# US016 — E2E test specification

**Owner:** tester.

## Why e2e does not apply

US016 adds `--forceExit` to the Jest invocation inside `scripts/review/run-gate.sh`. This is a single-line script edit with no web-application surface. The correctness guarantee is:

1. The flag is present in the script (IT-US016-001 — static grep).
2. The gate terminates in finite time when Jest finishes (IT-US016-002 — shell-harness timeout check).

Both are fully verifiable at the script-harness integration layer. The docker-compose stack (US022) is not needed; no browser interaction is involved; the RequestsLibrary has no applicable surface. Promoting to e2e would add zero coverage beyond what IT-US016-001/002 already provide.

**Verdict: No e2e scenarios. Script-level concern; covered by IT-US016-001 through IT-US016-003.**
