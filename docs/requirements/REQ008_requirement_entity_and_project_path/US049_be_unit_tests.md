# US049 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live next to the code they exercise inside `services/agent-board/internal/domain/` and `internal/handler/`.

## Coverage matrix

| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| Constant exists with correct value | unit | UT-049-001 | services/agent-board / internal/domain | `TaskStatusBlockedReviewGate` constant |
| Transition from in_review is valid | unit | UT-049-002 | services/agent-board / internal/domain | `Task.IsValidTransition` |
| Transition from changes_requested is valid | unit | UT-049-003 | services/agent-board / internal/domain | `Task.IsValidTransition` |
| blocked_review_gate is terminal (no transitions out) | unit | UT-049-004 | services/agent-board / internal/domain | `Task.IsValidTransition` |
| Cannot reach blocked_review_gate from pending | unit | UT-049-005 | services/agent-board / internal/domain | `Task.IsValidTransition` |
| Cannot reach blocked_review_gate from in_progress | unit | UT-049-006 | services/agent-board / internal/domain | `Task.IsValidTransition` |
| Cannot reach blocked_review_gate from completed | unit | UT-049-007 | services/agent-board / internal/domain | `Task.IsValidTransition` |
| Cannot reach blocked_review_gate from blocked_circuit_breaker | unit | UT-049-008 | services/agent-board / internal/domain | `Task.IsValidTransition` |
| Full transition matrix — no regression | unit | UT-049-009 | services/agent-board / internal/domain | `Task.IsValidTransition` (table-driven) |
| MCP update_task accepts blocked_review_gate from in_review | unit | UT-049-010 | services/agent-board / internal/handler | `update_task` MCP tool |
| MCP update_task accepts blocked_review_gate from changes_requested | unit | UT-049-011 | services/agent-board / internal/handler | `update_task` MCP tool |
| MCP update_task rejects blocked_review_gate from pending | unit | UT-049-012 | services/agent-board / internal/handler | `update_task` MCP tool |
| MCP update_task — task status persisted as blocked_review_gate | integration | IT-049-001 | services/agent-board / internal/handler | `update_task` → repo → DB |
| MCP update_task — blocked_review_gate task cannot transition further | integration | IT-049-002 | services/agent-board / internal/handler | `update_task` terminal state |

---

## Unit tests

### UT-049-001 — TaskStatusBlockedReviewGate constant has value "blocked_review_gate"
- **Service:** `services/agent-board`
- **Package under test:** `internal/domain`
- **Given:** The domain package is compiled.
- **When:** `domain.TaskStatusBlockedReviewGate` is evaluated.
- **Then:** Its value equals the string `"blocked_review_gate"`. (Compile-time or runtime assertion using `assert.Equal(t, "blocked_review_gate", domain.TaskStatusBlockedReviewGate)`.)
- **Edge cases:** Verify it is a named constant (not a variable), so it can be used in `switch` cases without shadowing.
- **Architecture cite:** US049 AC "Constant exists"; architecture scope `TaskStatusBlockedReviewGate = "blocked_review_gate"`

### UT-049-002 — IsValidTransition("blocked_review_gate") from in_review returns true
- **Service:** `services/agent-board`
- **Package under test:** `internal/domain`
- **Given:** `task := &domain.Task{Status: domain.TaskStatusInReview}`
- **When:** `task.IsValidTransition(domain.TaskStatusBlockedReviewGate)` is called.
- **Then:** Returns `true`.
- **Architecture cite:** US049 AC "Transition from in_review is valid"

### UT-049-003 — IsValidTransition("blocked_review_gate") from changes_requested returns true
- **Service:** `services/agent-board`
- **Package under test:** `internal/domain`
- **Given:** `task := &domain.Task{Status: domain.TaskStatusChangesRequested}`
- **When:** `task.IsValidTransition(domain.TaskStatusBlockedReviewGate)`.
- **Then:** Returns `true`.
- **Architecture cite:** US049 AC "Transition from changes_requested is valid"

### UT-049-004 — IsValidTransition from blocked_review_gate to any status returns false (terminal)
- **Service:** `services/agent-board`
- **Package under test:** `internal/domain`
- **Given:** `task := &domain.Task{Status: domain.TaskStatusBlockedReviewGate}`
- **When:** `task.IsValidTransition(newStatus)` is called for each of: `"pending"`, `"in_progress"`, `"in_review"`, `"completed"`, `"changes_requested"`, `"blocked_review_gate"` (self), `"blocked_circuit_breaker"`, `""` (empty), `"unknown_status"`.
- **Then:** Returns `false` for every value.
- **Architecture cite:** US049 AC "blocked_review_gate is terminal — no transitions out"

### UT-049-005 — Cannot reach blocked_review_gate from pending
- **Service:** `services/agent-board`
- **Package under test:** `internal/domain`
- **Given:** `task := &domain.Task{Status: domain.TaskStatusPending}`
- **When:** `task.IsValidTransition(domain.TaskStatusBlockedReviewGate)`.
- **Then:** Returns `false`.
- **Architecture cite:** US049 AC "Cannot reach blocked_review_gate from non-allowed states"

### UT-049-006 — Cannot reach blocked_review_gate from in_progress
- **Given:** `task.Status = "in_progress"`.
- **When:** `task.IsValidTransition("blocked_review_gate")`.
- **Then:** Returns `false`.

### UT-049-007 — Cannot reach blocked_review_gate from completed
- **Given:** `task.Status = "completed"`.
- **When:** `task.IsValidTransition("blocked_review_gate")`.
- **Then:** Returns `false`.

### UT-049-008 — Cannot reach blocked_review_gate from blocked_circuit_breaker
- **Given:** `task.Status = "blocked_circuit_breaker"`.
- **When:** `task.IsValidTransition("blocked_review_gate")`.
- **Then:** Returns `false`.

### UT-049-009 — Full transition matrix table-driven test (no regression)
- **Service:** `services/agent-board`
- **Package under test:** `internal/domain`
- **Given:** A table of `(fromStatus, toStatus, expectedResult)` triples covering the COMPLETE existing state machine PLUS the new `blocked_review_gate` transitions. The table must enumerate every pair of (existing from, existing to) that was previously tested in `status_machine_test.go` PLUS the new entries for `blocked_review_gate`.
- **Minimum required table entries:**

| from | to | expected |
|---|---|---|
| pending | in_progress | true |
| pending | in_review | false |
| pending | completed | false |
| pending | changes_requested | false |
| pending | blocked_circuit_breaker | false |
| pending | blocked_review_gate | false |
| in_progress | in_review | true |
| in_progress | pending | false |
| in_progress | completed | false |
| in_progress | blocked_review_gate | false |
| in_review | completed | true |
| in_review | changes_requested | true |
| in_review | blocked_review_gate | true |
| in_review | in_progress | false |
| in_review | pending | false |
| in_review | blocked_circuit_breaker | false |
| changes_requested | in_progress | true |
| changes_requested | in_review | true |
| changes_requested | completed | true |
| changes_requested | blocked_circuit_breaker | true |
| changes_requested | blocked_review_gate | true |
| changes_requested | pending | false |
| completed | in_progress | false |
| completed | in_review | false |
| completed | blocked_review_gate | false |
| completed | (any) | false |
| blocked_circuit_breaker | (any) | false |
| blocked_review_gate | in_progress | false |
| blocked_review_gate | in_review | false |
| blocked_review_gate | completed | false |
| blocked_review_gate | changes_requested | false |
| blocked_review_gate | pending | false |
| blocked_review_gate | blocked_circuit_breaker | false |
| blocked_review_gate | blocked_review_gate | false |

- **When:** `task.IsValidTransition(to)` is called for each row.
- **Then:** Returns `expected` for each row. Any deviation from the table is a regression.
- **Architecture cite:** US049 AC "Existing transitions are unaffected"; US049 notes — "table-driven test … is the cleanest way to prove no regression"

### UT-049-010 — MCP update_task accepts blocked_review_gate from in_review
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler` (MCP tools)
- **Given:**
  - Mock task repo `GetTask` returns a task with `Status = "in_review"`.
  - Mock task repo `UpdateTask` expects to be called with `Status = "blocked_review_gate"` and returns the updated task.
- **When:** MCP `update_task` tool called with `id = <task-id>`, `status = "blocked_review_gate"`.
- **Then:**
  - Tool returns successfully (no tool error).
  - Returned object has `"status": "blocked_review_gate"`.
  - `UpdateTask` was called once with the new status.
- **Architecture cite:** US049 AC "MCP update_task accepts the new status"

### UT-049-011 — MCP update_task accepts blocked_review_gate from changes_requested
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock `GetTask` returns task with `Status = "changes_requested"`. Mock `UpdateTask` returns updated task.
- **When:** `update_task` called with `status = "blocked_review_gate"`.
- **Then:** Tool returns success with `status = "blocked_review_gate"`. No invalid-transition error.
- **Architecture cite:** US049 AC "MCP update_task accepts the new status"

### UT-049-012 — MCP update_task rejects blocked_review_gate from pending
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock `GetTask` returns task with `Status = "pending"`.
- **When:** `update_task` called with `status = "blocked_review_gate"`.
- **Then:**
  - Tool returns a tool error containing "invalid status transition" (or equivalent).
  - `UpdateTask` is NOT called (no write happens).
  - Task status remains `"pending"` (mock verifies no update call).
- **Architecture cite:** US049 AC "MCP update_task rejects an invalid transition"; domain `IsValidTransition` returns false for pending → blocked_review_gate

---

## Integration tests

### IT-049-001 — MCP update_task persists blocked_review_gate status to DB
- **Service:** `services/agent-board`
- **Boundary:** MCP tool handler → `TaskRepository.UpdateTask` → Postgres (testcontainer)
- **Setup:**
  1. Insert a project, user story (with a requirement after US044 migration), and task with `status = "in_review"` into the test DB.
  2. Capture the task's `id`.
- **When:** Invoke the `update_task` MCP tool handler directly (or via httptest MCP endpoint) with `id = <task-id>`, `status = "blocked_review_gate"`.
- **Then:**
  - Tool call succeeds (no error returned).
  - `SELECT status FROM tasks WHERE id = <task-id>` returns `"blocked_review_gate"`.
  - `updated_at` in the DB is later than the initial `created_at`.
- **Teardown:** Drop test container.
- **Architecture cite:** US049 AC "MCP update_task accepts the new status" + "task's persisted status becomes blocked_review_gate"

### IT-049-002 — MCP update_task rejects further transitions out of blocked_review_gate
- **Service:** `services/agent-board`
- **Boundary:** MCP tool → domain → repo → DB
- **Setup:** Same as IT-049-001; task starts at `"in_review"`, transition to `"blocked_review_gate"` (step 1).
- **When:** Call `update_task` again with `status = "in_progress"` (attempt to escape the terminal state).
- **Then:**
  - Tool returns a tool error containing "invalid status transition".
  - `SELECT status FROM tasks WHERE id = <task-id>` still returns `"blocked_review_gate"` (unchanged).
- **Architecture cite:** US049 AC "blocked_review_gate is terminal — no transitions out"
