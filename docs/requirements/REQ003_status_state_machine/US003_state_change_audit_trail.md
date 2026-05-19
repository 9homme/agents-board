# US003 — State change audit trail

**Requirement:** REQ003 — status_state_machine
**Status:** draft

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
- US001_task_state_machine
- US002_story_state_machine

## Notes for the team
- Ensure the audit trail includes timestamps for accurate chronological tracking.

## Sign-off log
