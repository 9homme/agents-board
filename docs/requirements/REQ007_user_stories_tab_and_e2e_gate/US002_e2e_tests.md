# US002 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ007_user_stories_tab_and_e2e_gate/US002_makefile_healthcheck_fixes.robot`.

## Why e2e
US002 alters developer tooling (`Makefile`). While not a traditional web flow, proving the tooling updates prevents the e2e stack itself from hanging. Robot Framework checks the Makefile content and dry-run output to verify the bounded probe and data-only seed constraints are met.

## Scenarios
### E2E-US002-001 — mcp-server health-check is bounded and e2e-seed is data-only
- **Tag:** US002
- **Preconditions:** The `Makefile` exists in the repository root.
- **Steps:** 
  1. Read the `Makefile` contents.
  2. Verify the `e2e-up` target includes a `curl` to mcp-server with `--max-time 5`.
  3. Verify the `e2e-seed` target does NOT contain a migration step.
- **Expected:** Makefile changes conform to the ACs.
- **Cleanup:** None.
