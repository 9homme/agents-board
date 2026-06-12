# US009 — Story state machine

**Requirement:** REQ003 — status_state_machine
**Status:** done

## Story
As an API client, I want the backend to enforce the User Story status state machine, so that stories follow the designated workflow.

## Acceptance criteria
- **Scenario: Valid forward transitions**
  - Given a story in a specific state
  - When I update its status to the next valid state (`draft` -> `in_development` -> `in_signoff` -> `done`)
  - Then the transition is successful
- **Scenario: Sign-off cycle transitions**
  - Given a story in `in_signoff` or `changes_requested`
  - When I update its status between these states or back to `in_development` (e.g. `in_signoff` -> `changes_requested`, `changes_requested` -> `in_development`, `changes_requested` -> `in_signoff`)
  - Then the transition is successful
- **Scenario: Circuit breaker transition**
  - Given a story in `changes_requested`
  - When I update its status to `blocked_circuit_breaker`
  - Then the transition is successful
- **Scenario: Invalid transitions are rejected**
  - Given a story in a specific state
  - When I attempt an invalid status update (e.g., `draft` -> `done`, `done` -> `in_development`)
  - Then the backend rejects the update with a clear error message indicating an invalid transition
- **Scenario: Enforce initial state on creation**
  - Given I am creating a new story
  - When I provide any status other than `draft` (or omit it)
  - Then the backend either defaults it to `draft` or rejects the request with a validation error

## UI / UX flow expectations
No UI: Operations are exposed via MCP/Backend APIs. Any UI calling these APIs will simply receive validation errors for invalid transitions and must handle them accordingly.

## Out of scope
- Automatic cascading state changes (e.g. automatically setting story to `in_development` when first task becomes `in_progress`).

## Dependencies
- REQ001 US011 (User Story CRUD)

## Notes for the team
- The exact state machine is: `draft` -> `in_development` -> `in_signoff` <-> `changes_requested` -> `done` (and `blocked_circuit_breaker`).
- Ensure API error messages are descriptive so clients understand which transitions are allowed.

## Sign-off log

### Sign-off pass 1 — 2026-05-19 — verdict: approved
- **Spec review:** All five acceptance criteria are mapped to concrete tests in the specs.
  - AC "Valid forward transitions" → UT-001 (covers `draft` -> `in_development` -> `in_signoff` -> `done` and `changes_requested` -> `done`) + E2E-001 steps 1-3, 6.
  - AC "Sign-off cycle transitions" → UT-002 (covers `in_signoff` <-> `changes_requested` and `changes_requested` -> `in_development`) + E2E-001 steps 4-5.
  - AC "Circuit breaker transition" → UT-003 (`changes_requested` -> `blocked_circuit_breaker`).
  - AC "Invalid transitions are rejected" → UT-004 (domain layer) + IT-001 (MCP layer returns `isError: true`) + E2E-002 over live HTTP/SSE.
  - AC "Enforce initial state on creation" → UT-005 (`NewUserStory` / `create_user_story` handler) + E2E-003.
  - E2E justification is honest: only the user-observable lifecycle and error-marshalling cases are promoted to e2e; transition rule permutations stay at unit layer.
- **Result review:** Test report (commit `f86fa43`) shows UT-001..005 + IT-001 all PASS, supporting repo-layer tests `TestUserStoryRepo_UpdateUserStoryStatus` and `TestUserStoryRepo_UpdateUserStoryStatus_RollbackOnAuditFailure` PASS, and E2E-001..003 all PASS (3/3 after the tester fixed the JSON-path regression). No skipped tests. Counts match specs.
- **Routed to:** none — story approved.
