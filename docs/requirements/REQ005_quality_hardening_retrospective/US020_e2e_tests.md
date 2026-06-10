# US020 — E2E test specification

**Owner:** tester. Implemented in `tests/e2e/REQ005_quality_hardening_retrospective/US020_abort_controller_hooks.robot`.

## Why e2e

The AbortController + signal-thread refactor is an internal correctness improvement — the happy path is functionally identical to today. Most of the correctness guarantees (abort on unmount, no stale state, AbortError swallowed) are at the hook/unit layer (FCT-US020-003 through FCT-US020-011).

However, one scenario belongs at e2e: **rapid navigation on the live stack must not produce stale data or visible errors**. This cannot be fully proven at the unit layer alone because:

- MSW mocks simulate delay; real network latency patterns differ.
- The React lifecycle (strict-mode double-invocation in dev, router transitions in Next.js Pages Router CSR) interacts with AbortController in ways that are easier to observe on the full stack.
- The only end-user-observable regression would be: navigate to Project A, quickly navigate to Project B, see Project A's stale data rendered under Project B's URL — a visual regression that only the live stack can surface.

One e2e scenario is justified. The remaining abort semantics are fully covered by FCT-US020-003 through FCT-US020-010.

## Scenarios

### E2E-US020-001 — Rapid project navigation does not show stale data or error state

- **Tag:** US020, regression
- **Preconditions:**
  - `make e2e-up && make e2e-seed` has run successfully.
  - `WEB_BASE_URL` = `http://localhost:3000` (default).
  - `API_BASE_URL` = `http://localhost:8080` (default).
  - Seed data: at least two distinct projects exist in the DB. The baseline seed (`tests/e2e/data/seeds/REQ000_baseline.sql`) provides project-1 and project-2 (or create equivalent in a REQ005-specific seed if baseline is not sufficient — see `tests/e2e/data/seeds/` contract in architecture §6.5).
  - The seed provides: project-1 `name = "Sample Project"`, project-2 `name = "Second Project"` (or whatever the baseline seed uses — tester adjusts variable references in the Robot file to match the actual seed data).
- **Steps (Browser library — UI flow):**
  1. `New Browser` / `New Page` (or use existing session).
  2. Navigate to `${WEB_BASE_URL}/` (dashboard).
  3. Wait for projects list to render: `Wait For Elements State    css=[data-testid="project-card"]    visible`.
  4. Click on project-1's card to navigate to the project detail page.
  5. Immediately (within ~100 ms) click the browser's back button OR directly navigate to project-2's URL: `Go To    ${WEB_BASE_URL}/projects/${PROJECT_2_ID}`.
  6. `Wait For Elements State    css=h1    visible    timeout=5s`.
- **Expected:**
  - The `<h1>` text matches project-2's name (not project-1's — no stale data).
  - No element matching `text="Failed to load project"` is visible.
  - No element matching `text="Project not found"` is visible.
  - `npm test`-style console errors are NOT surfaced in the browser console (no uncaught AbortError bubbling).
- **Cleanup:** `Close Browser` (or leave open for next suite if session is shared).
- **Notes:** If the Navigation timing is not fast enough to exercise the in-flight-abort scenario, this test still has value as a "rapid navigation does not break the page" smoke test. The unit tests (FCT-US020-004/006) are the authoritative abort-correctness proofs.
