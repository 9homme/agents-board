# US003 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live next to the code they exercise inside the relevant `services/<service-name>/` module.

## Coverage matrix
| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| Audit record created on task status change | integration | IT-001 | services/agent-board / internal/repo | `UpdateTaskStatus(...)` |
| Audit record created on story status change | integration | IT-002 | services/agent-board / internal/repo | `UpdateUserStoryStatus(...)` |
| Audit record not created on invalid transition | integration | IT-003 | services/agent-board / internal/handler | `update_task` / `update_user_story` MCP tools |
| Retrieve task audit trail | integration | IT-004 | services/agent-board / internal/handler | `get_task_audit_trail` MCP tool |
| Retrieve story audit trail | integration | IT-005 | services/agent-board / internal/handler | `get_user_story_audit_trail` MCP tool |

## Integration tests
### IT-001 — Audit record created on task status change
- **Service:** `services/agent-board`
- **Boundary:** repo ↔ DB
- **Setup:** A test DB with an existing task in `pending` state.
- **Endpoint exercised:** Repo method responsible for updating task status (e.g., `UpdateTaskStatus`).
- **Given:** Update task status to `in_progress`.
- **Expect:** Task status is updated to `in_progress`. A new record is inserted into `status_audit_trail` table within the same transaction, containing `entity_id` (task ID), `entity_type` ('task'), `from_status` ('pending'), `to_status` ('in_progress').
- **Architecture cite:** D-002, D-003.

### IT-002 — Audit record created on story status change
- **Service:** `services/agent-board`
- **Boundary:** repo ↔ DB
- **Setup:** A test DB with an existing story in `draft` state.
- **Endpoint exercised:** Repo method responsible for updating story status (e.g., `UpdateUserStoryStatus`).
- **Given:** Update story status to `in_development`.
- **Expect:** Story status is updated to `in_development`. A new record is inserted into `status_audit_trail` table within the same transaction, containing `entity_id` (story ID), `entity_type` ('user_story'), `from_status` ('draft'), `to_status` ('in_development').
- **Architecture cite:** D-002, D-003.

### IT-003 — Audit record not created on invalid transition
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ DB
- **Setup:** A test DB with an existing task in `pending` state.
- **Endpoint exercised:** MCP tool `update_task`.
- **Given:** Request to update task status to `completed` (invalid).
- **Expect:** MCP tool returns error. No record is inserted into `status_audit_trail` for this task.
- **Architecture cite:** US003 AC "Audit record not created on invalid transition".

### IT-004 — Retrieve task audit trail
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ DB
- **Setup:** A test DB with a task that has undergone multiple valid state transitions (e.g., `pending` -> `in_progress` -> `in_review`).
- **Endpoint exercised:** MCP tool `get_task_audit_trail`.
- **Request body:** `{"taskId": "..."}`
- **Expect:** Response contains `auditTrail` array with entries in chronological order, containing fields: `id`, `entityId`, `entityType` ('task'), `fromStatus`, `toStatus`, `changedAt`.
- **Architecture cite:** API contracts (exact) - `get_task_audit_trail`.

### IT-005 — Retrieve story audit trail
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ DB
- **Setup:** A test DB with a story that has undergone multiple valid state transitions.
- **Endpoint exercised:** MCP tool `get_user_story_audit_trail`.
- **Request body:** `{"userStoryId": "..."}`
- **Expect:** Response contains `auditTrail` array with entries in chronological order.
- **Architecture cite:** API contracts (exact) - `get_user_story_audit_trail`.
