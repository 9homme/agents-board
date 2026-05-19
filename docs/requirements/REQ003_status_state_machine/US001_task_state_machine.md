# US001 — Task state machine

**Requirement:** REQ003 — status_state_machine
**Status:** done

## Story
As an API client, I want the backend to enforce the Task status state machine, so that tasks cannot bypass required states or enter invalid states.

## Acceptance criteria
- **Scenario: Valid forward transitions**
  - Given a task in a specific state
  - When I update its status to the next valid forward state (e.g., `pending` -> `in_progress` -> `in_review` -> `completed`)
  - Then the transition is successful
- **Scenario: Review cycle transitions**
  - Given a task in `in_review` or `changes_requested`
  - When I update its status between these states or back to `in_progress` (e.g. `in_review` -> `changes_requested`, `changes_requested` -> `in_progress`, `changes_requested` -> `in_review`)
  - Then the transition is successful
- **Scenario: Circuit breaker transition**
  - Given a task in `changes_requested`
  - When I update its status to `blocked_circuit_breaker`
  - Then the transition is successful
- **Scenario: Invalid transitions are rejected**
  - Given a task in a specific state
  - When I attempt an invalid status update (e.g., `pending` -> `completed`, `completed` -> `in_progress`)
  - Then the backend rejects the update with a clear error message indicating an invalid transition
- **Scenario: Enforce initial state on creation**
  - Given I am creating a new task
  - When I provide any status other than `pending` (or omit it)
  - Then the backend either defaults it to `pending` or rejects the request with a validation error

## UI / UX flow expectations
No UI: Operations are exposed via MCP/Backend APIs. Any UI calling these APIs will simply receive validation errors for invalid transitions and must handle them accordingly.

## Out of scope
- Cascading state changes.

## Dependencies
- REQ001 US005 (Task CRUD)

## Notes for the team
- The exact state machine is: `pending` -> `in_progress` -> `in_review` <-> `changes_requested` -> `completed` (and `blocked_circuit_breaker`).
- Ensure API error messages are descriptive so clients understand which transitions are allowed.

## Sign-off log

### Sign-off pass 1 — 2026-05-19 — verdict: approved
- **Spec review:** All five AC scenarios map to spec cases:
  - "Valid forward transitions" -> UT-001 + E2E-001 (steps 1-3, 5-6).
  - "Review cycle transitions" -> UT-002 + E2E-001 (steps 4-5 cover `in_review` -> `changes_requested` -> `in_progress`).
  - "Circuit breaker transition" -> UT-003 (`changes_requested` -> `blocked_circuit_breaker`).
  - "Invalid transitions are rejected" -> UT-004 (domain) + IT-001 (handler layer) + E2E-002 (over-the-wire MCP error envelope).
  - "Enforce initial state on creation" -> UT-005 + E2E-003.
  - No FE spec is required (story declares "No UI"); the e2e justification is honest — IT-001 covers the handler-domain seam, E2E-002 only covers the over-the-wire error marshaling that unit tests can't reach.
- **Result review:** Per `US001_test_report.md` (commit `f86fa43`):
  - BE: UT-001..005 + IT-001 all PASS, plus supporting repo-layer tests `TestTaskRepo_UpdateTaskStatus` and `TestTaskRepo_UpdateTaskStatus_RollbackOnAuditFail` PASS.
  - E2E: E2E-001, E2E-002, E2E-003 all PASS (re-run after tester's JSON-path fix).
  - No skipped tests; FE table is N/A as expected.
- **Routed to:** none.
