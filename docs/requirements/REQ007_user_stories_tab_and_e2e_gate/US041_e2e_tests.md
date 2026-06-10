# US041 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ007_user_stories_tab_and_e2e_gate/US041_e2e_pr_quality_gate.robot`.

## Why e2e
US041 introduces a GitHub Actions workflow. E2E verifies that the `.github/workflows/e2e.yml` file is created with the correct triggers, setup steps, and teardown logic per the ACs.

## Scenarios
### E2E-US041-001 — GitHub Actions workflow is correctly configured
- **Tag:** US041
- **Preconditions:** None.
- **Steps:** 
  1. Read `.github/workflows/e2e.yml`.
  2. Verify trigger is `pull_request` on `branches: [main]`.
  3. Verify it runs `make e2e-up`, `make e2e-seed`, and `make e2e-run`.
  4. Verify artifact upload uses `if: always()`.
  5. Verify teardown `make e2e-down` uses `if: always()`.
- **Expected:** Workflow file conforms exactly to the architectural and story requirements.
- **Cleanup:** None.
