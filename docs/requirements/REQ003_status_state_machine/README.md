# REQ003 — status_state_machine

## Summary
The backend must enforce a strict state machine for Tasks and User Stories, ensuring that their statuses follow the exact workflow used by our agent engineering team. This prevents invalid state jumps (e.g., moving directly from 'pending' to 'completed') and enforces correct initial states upon creation.

## Business Goal
To ensure data integrity and process compliance by forcing all work items to go through the correct lifecycle stages.

## Decisions
- State transitions will be enforced within the backend domain logic/API validation.
- Attempting an invalid transition will result in a clear error (rejecting invalid updates to existing CRUD endpoints).
- Creation of new items will enforce their respective initial states ('pending' for tasks, 'draft' for stories).
- State change history (audit trail) will be tracked for all transitions.
- Cascading state changes are out of scope (e.g., tasks do not automatically update story state).

## User Stories
- **US001_task_state_machine**: Enforce Task status state machine.
- **US002_story_state_machine**: Enforce User Story status state machine.
- **US003_state_change_audit_trail**: Track historical state change audit trail for tasks and stories.

## Tasks
| Task File | Title | Blocked By | Status |
|---|---|---|---|
| `US001_be_scaffold_domain_and_migration.md` | Scaffold Domain and Migration | None | pending |
| `US001_be_task_state_machine.md` | Task State Machine | `US001_be_scaffold_domain_and_migration.md` | pending |
| `US002_be_story_state_machine.md` | Story State Machine | `US001_be_scaffold_domain_and_migration.md` | pending |
| `US003_be_audit_trail_mcp.md` | Audit Trail MCP | `US001_be_scaffold_domain_and_migration.md` | pending |
