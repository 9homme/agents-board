# US009 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live next to the code they exercise inside the relevant `services/<service-name>/` module.

## Coverage matrix
| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| Valid forward transitions | unit | UT-001 | services/agent-board / internal/domain | `UserStory.IsValidTransition(...)` |
| Sign-off cycle transitions | unit | UT-002 | services/agent-board / internal/domain | `UserStory.IsValidTransition(...)` |
| Circuit breaker transition | unit | UT-003 | services/agent-board / internal/domain | `UserStory.IsValidTransition(...)` |
| Invalid transitions are rejected | unit | UT-004 | services/agent-board / internal/domain | `UserStory.IsValidTransition(...)` |
| Enforce initial state on creation | unit | UT-005 | services/agent-board / internal/domain | `NewUserStory(...)` or `CreateUserStory` handler |
| Reject invalid transitions at MCP layer | integration | IT-001 | services/agent-board / internal/handler | `update_user_story` MCP tool |

## Unit tests
### UT-001 — Valid forward transitions
- **Service:** `services/agent-board`
- **Function under test:** `internal/domain.IsValidTransition` (or equivalent method on UserStory)
- **Given:** A UserStory entity in a specific status.
- **When:** Checking transitions: `draft` -> `in_development`, `in_development` -> `in_signoff`, `in_signoff` -> `done`, `changes_requested` -> `done`.
- **Then:** Returns `true`.
- **Architecture cite:** D-001.

### UT-002 — Sign-off cycle transitions
- **Service:** `services/agent-board`
- **Function under test:** `internal/domain.IsValidTransition`
- **Given:** A UserStory entity in `in_signoff` or `changes_requested` status.
- **When:** Checking transitions: `in_signoff` -> `changes_requested`, `changes_requested` -> `in_development`, `changes_requested` -> `in_signoff`.
- **Then:** Returns `true`.
- **Architecture cite:** US009 AC "Sign-off cycle transitions".

### UT-003 — Circuit breaker transition
- **Service:** `services/agent-board`
- **Function under test:** `internal/domain.IsValidTransition`
- **Given:** A UserStory entity in `changes_requested` status.
- **When:** Checking transition to `blocked_circuit_breaker`.
- **Then:** Returns `true`.
- **Architecture cite:** US009 AC "Circuit breaker transition".

### UT-004 — Invalid transitions are rejected
- **Service:** `services/agent-board`
- **Function under test:** `internal/domain.IsValidTransition`
- **Given:** A UserStory entity in a specific status.
- **When:** Checking invalid transitions (e.g., `draft` -> `done`, `done` -> `in_development`, `done` -> `draft`).
- **Then:** Returns `false`.
- **Architecture cite:** D-001.

### UT-005 — Enforce initial state on creation
- **Service:** `services/agent-board`
- **Function under test:** `internal/domain.NewUserStory` or handler layer `create_user_story`.
- **Given:** Request to create a new user story.
- **When:** Status provided is not `draft` (e.g., `done`, `in_development`).
- **Then:** Returns validation error or defaults to `draft`.
- **Architecture cite:** US009 AC "Enforce initial state on creation".

## Integration tests
### IT-001 — Reject invalid transitions at MCP layer
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ domain
- **Setup:** A mock or test DB with an existing user story in `draft` state.
- **Endpoint exercised:** MCP tool `update_user_story`
- **Request body:** `{"id": "...", "status": "done"}`
- **Expect:** Response contains `isError: true` and descriptive error message (e.g., "Invalid transition").
- **Architecture cite:** API contracts (exact) - MCP Tools.
