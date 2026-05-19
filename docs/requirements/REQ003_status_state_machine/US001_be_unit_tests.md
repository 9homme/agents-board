# US001 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live next to the code they exercise inside the relevant `services/<service-name>/` module.

## Coverage matrix
| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| Valid forward transitions | unit | UT-001 | services/agent-board / internal/domain | `Task.IsValidTransition(...)` |
| Review cycle transitions | unit | UT-002 | services/agent-board / internal/domain | `Task.IsValidTransition(...)` |
| Circuit breaker transition | unit | UT-003 | services/agent-board / internal/domain | `Task.IsValidTransition(...)` |
| Invalid transitions are rejected | unit | UT-004 | services/agent-board / internal/domain | `Task.IsValidTransition(...)` |
| Enforce initial state on creation | unit | UT-005 | services/agent-board / internal/domain | `NewTask(...)` or `CreateTask` handler |
| Reject invalid transitions at MCP layer | integration | IT-001 | services/agent-board / internal/handler | `update_task` MCP tool |

## Unit tests
### UT-001 — Valid forward transitions
- **Service:** `services/agent-board`
- **Function under test:** `internal/domain.IsValidTransition` (or equivalent method on Task)
- **Given:** A Task entity in a specific status.
- **When:** Checking transitions: `pending` -> `in_progress`, `in_progress` -> `in_review`, `in_review` -> `completed`, `changes_requested` -> `completed`. (Note: `changes_requested` -> `completed` might not be valid, check AC. Actually AC says `pending` -> `in_progress` -> `in_review` <-> `changes_requested` -> `completed`). Let's test `in_review` -> `completed` and `changes_requested` -> `completed`.
- **Then:** Returns `true`.
- **Architecture cite:** D-001.

### UT-002 — Review cycle transitions
- **Service:** `services/agent-board`
- **Function under test:** `internal/domain.IsValidTransition`
- **Given:** A Task entity in `in_review` or `changes_requested` status.
- **When:** Checking transitions: `in_review` -> `changes_requested`, `changes_requested` -> `in_progress`, `changes_requested` -> `in_review`.
- **Then:** Returns `true`.
- **Architecture cite:** US001 AC "Review cycle transitions".

### UT-003 — Circuit breaker transition
- **Service:** `services/agent-board`
- **Function under test:** `internal/domain.IsValidTransition`
- **Given:** A Task entity in `changes_requested` status.
- **When:** Checking transition to `blocked_circuit_breaker`.
- **Then:** Returns `true`.
- **Architecture cite:** US001 AC "Circuit breaker transition".

### UT-004 — Invalid transitions are rejected
- **Service:** `services/agent-board`
- **Function under test:** `internal/domain.IsValidTransition`
- **Given:** A Task entity in a specific status.
- **When:** Checking invalid transitions (e.g., `pending` -> `completed`, `completed` -> `in_progress`, `completed` -> `pending`).
- **Then:** Returns `false`.
- **Architecture cite:** D-001.

### UT-005 — Enforce initial state on creation
- **Service:** `services/agent-board`
- **Function under test:** `internal/domain.NewTask` or handler layer `create_task`.
- **Given:** Request to create a new task.
- **When:** Status provided is not `pending` (e.g., `completed`, `in_progress`).
- **Then:** Returns validation error or defaults to `pending`.
- **Architecture cite:** US001 AC "Enforce initial state on creation".

## Integration tests
### IT-001 — Reject invalid transitions at MCP layer
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ domain
- **Setup:** A mock or test DB with an existing task in `pending` state.
- **Endpoint exercised:** MCP tool `update_task`
- **Request body:** `{"id": "...", "status": "completed"}`
- **Expect:** Response contains `isError: true` and descriptive error message (e.g., "Invalid transition").
- **Architecture cite:** API contracts (exact) - MCP Tools.
