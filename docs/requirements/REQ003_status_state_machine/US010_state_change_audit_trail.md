# US010 — State change audit trail

**Requirement:** REQ003 — status_state_machine
**Status:** done

## Story
As an API client, I want the backend to track the history of state changes for tasks and user stories, so that there is an audit trail of when a status was changed and what the transitions were.

## Acceptance criteria
- **Scenario: Audit record created on status change**
  - Given an existing task or user story
  - When its status is successfully updated to a new valid state
  - Then an audit trail record is created, containing the timestamp, the previous state, and the new state
- **Scenario: Audit record not created on invalid transition**
  - Given an existing task or user story
  - When a status update fails due to an invalid transition
  - Then no audit trail record is created
- **Scenario: Retrieve audit trail**
  - Given a task or user story with a history of state changes
  - When I query for its audit trail
  - Then the backend returns a chronological list of state changes for that item

## UI / UX flow expectations
No UI: Operations are exposed via MCP/Backend APIs. Any UI calling these APIs will be able to retrieve and display the historical audit trail if needed.

## Out of scope
- Tracking non-status changes (e.g., changes to description or title).

## Dependencies
- US008_task_state_machine
- US009_story_state_machine

## Notes for the team
- Ensure the audit trail includes timestamps for accurate chronological tracking.

## Sign-off log

### Sign-off pass 1 — 2026-05-19 — verdict: approved
- **Spec review:** All three acceptance criteria are covered by the test specs:
  - AC "Audit record created on status change" → IT-001 (task repo) + IT-002 (story repo), both asserting the audit row is written in the same transaction with the correct `entity_id`, `entity_type`, `from_status`, `to_status`.
  - AC "Audit record not created on invalid transition" → IT-003 (handler-level, both `update_task` and `update_user_story`) + E2E-003 (live invalid transition + empty audit-trail check).
  - AC "Retrieve audit trail" → IT-004 (`get_task_audit_trail`) + IT-005 (`get_user_story_audit_trail`) covering chronological order and the exact response shape from the architecture, reinforced by E2E-001 / E2E-002 over the live MCP server.
  - No UI surface, so no FE/FCT coverage required (story correctly declares "No UI").
  - Pyramid is honest: chronological-order and JSON-shape assertions are validated at the integration layer; e2e only confirms the end-to-end MCP round-trip (justified, since previous review pass uncovered that the tools had not been registered with the live server).
- **Result review:** `US010_test_report.md` shows IT-001..IT-005 all PASS plus supporting empty/missing-ID/repo-error cases, and E2E-001..E2E-003 all PASS. No skipped tests, no `t.Skip`, no `[Tags] skip`. Test counts match the specs.
- **Routed to:** none (approved).
