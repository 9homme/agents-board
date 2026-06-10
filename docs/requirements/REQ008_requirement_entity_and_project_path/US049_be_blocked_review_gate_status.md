# US049/be_blocked_review_gate_status

**Requirement:** REQ008
**Story:** US049
**Track:** BE
**Service:** services/agent-board
**Status:** in_review
**Blocked by:** none
**Worked-by:** be-dev-20260610T000000Z-af21
**Implements:** US049, architecture scope item "`blocked_review_gate` task status (US049)" (Revision 8)

## Goal
Add the `blocked_review_gate` task status to the Go domain state machine — a terminal state reachable only from `in_review` and `changes_requested` — so MCP `update_task` can move a task there when the review-gate tool fails. No DB migration.

## Scope
- **In:**
  - Add `TaskStatusBlockedReviewGate = "blocked_review_gate"` constant to `internal/domain/status_machine.go`.
  - Allow `in_review → blocked_review_gate` and `changes_requested → blocked_review_gate` in `Task.IsValidTransition`.
  - Make `blocked_review_gate` terminal (no transitions out) by adding it to the terminal-states case.
- **Out:**
  - Any DB migration (status column is TEXT; enforcement is Go-domain-only).
  - UserStory state machine (Task only).
  - New MCP tool — `update_task` already delegates to `IsValidTransition`; the new status flows through once the domain allows it. Add MCP-level accept/reject tests only.
  - Web UI for the status.
  - Circuit-breaker logic (`blocked_circuit_breaker` untouched).

## Files touched (estimated, exclusive)
- `services/agent-board/internal/domain/status_machine.go` (modify)
- `services/agent-board/internal/domain/status_machine_test.go` (modify — full transition matrix)
- `services/agent-board/internal/handler/task_tools_test.go` (modify — MCP `update_task` accept/reject for the new status)

**No shared-scaffold collision.** This task is fully independent of US044–US048 (it touches only `status_machine.go` and `task_tools_test.go`). It can be picked first, in parallel with US044.

## Architecture extract

### Scope item — `blocked_review_gate` task status (US049)
Add `TaskStatusBlockedReviewGate = "blocked_review_gate"` constant to `internal/domain/status_machine.go`. Valid transitions: `in_review → blocked_review_gate` and `changes_requested → blocked_review_gate`. Terminal — no transitions out. No DB migration (status column is TEXT; enforcement is Go-domain-only). Enables agents to set this status via MCP `update_task` when the review gate tool fails (distinct from `blocked_circuit_breaker` which is a code-review failure).

### Current state machine (verbatim — `internal/domain/status_machine.go`)
```go
// Task Statuses
const (
	TaskStatusPending               = "pending"
	TaskStatusInProgress            = "in_progress"
	TaskStatusInReview              = "in_review"
	TaskStatusChangesRequested      = "changes_requested"
	TaskStatusCompleted             = "completed"
	TaskStatusBlockedCircuitBreaker = "blocked_circuit_breaker"
)

func (t *Task) IsValidTransition(newStatus string) bool {
	switch t.Status {
	case TaskStatusPending:
		return newStatus == TaskStatusInProgress
	case TaskStatusInProgress:
		return newStatus == TaskStatusInReview
	case TaskStatusInReview:
		return newStatus == TaskStatusCompleted || newStatus == TaskStatusChangesRequested
	case TaskStatusChangesRequested:
		return newStatus == TaskStatusInProgress || newStatus == TaskStatusInReview || newStatus == TaskStatusCompleted || newStatus == TaskStatusBlockedCircuitBreaker
	case TaskStatusCompleted, TaskStatusBlockedCircuitBreaker:
		return false // Terminal states
	default:
		return false
	}
}
```

### Required change (mirror the `blocked_circuit_breaker` treatment exactly)
- Add `TaskStatusBlockedReviewGate = "blocked_review_gate"` to the Task status const block.
- In the `TaskStatusInReview` case: add `|| newStatus == TaskStatusBlockedReviewGate`.
- In the `TaskStatusChangesRequested` case: add `|| newStatus == TaskStatusBlockedReviewGate`.
- In the terminal case: `case TaskStatusCompleted, TaskStatusBlockedCircuitBreaker, TaskStatusBlockedReviewGate:` → `return false`.
- Resulting valid transitions into `blocked_review_gate`: only from `in_review` and `changes_requested`. From `pending`/`in_progress`/`completed`/`blocked_circuit_breaker`: `false`. From `blocked_review_gate`: `false` to every target.

### MCP `update_task` (no code change needed — already delegates)
`task_tools.go` `update_task` already calls `existing.IsValidTransition(*req.Status)` and returns `invalid transition from X to Y` on `false`. Once the domain allows the new edges, `update_task` accepts `blocked_review_gate` from `in_review`/`changes_requested` and rejects it from others. Add tests only.

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US049_be_unit_tests.md`: UT/IT IDs covering — `TaskStatusBlockedReviewGate == "blocked_review_gate"`; `in_review → blocked_review_gate` true; `changes_requested → blocked_review_gate` true; `blocked_review_gate → *` false for every status; `{pending,in_progress,completed,blocked_circuit_breaker} → blocked_review_gate` false; full existing-transition matrix unchanged (no regressions); MCP `update_task` accepts `blocked_review_gate` from `in_review`/`changes_requested` and rejects it from `pending` (status unchanged on reject).
- Flag any spec gaps back to tester.

## Implementation notes
- A table-driven test enumerating every (from, to) pair is the cleanest way to prove "existing transitions unaffected" (tester guidance) — the dev writes production code; the matrix lives in the test spec.
- Do NOT add `blocked_review_gate` to the UserStory state machine.

## Definition of done
- All listed tests green.
- `go vet ./...` and `go test ./...` clean inside `services/agent-board`.
- Coverage ≥80% on `status_machine.go` (or written `## Coverage exemption`).
- No new public exports without a doc comment.
- Code matches the `## Architecture extract`.
- Review gate green (BE + cross; paste `REVIEW GATE: PASS` into `## Notes`).
- `robot --dryrun tests/e2e/REQ008_*/` parses (paste output into `## Notes`).
- Dev set status to `in_review` and reported back.

## Notes

### Files touched
- `services/agent-board/internal/domain/status_machine.go` — added `TaskStatusBlockedReviewGate = "blocked_review_gate"` constant (with doc comment); added `blocked_review_gate` as valid target in `in_review` and `changes_requested` switch cases; added `TaskStatusBlockedReviewGate` to the terminal-states case.
- `services/agent-board/internal/domain/status_machine_test.go` — added UT-049-001 through UT-049-009 (constant value, individual valid/invalid transitions, full table-driven regression matrix with 40 entries).
- `services/agent-board/internal/handler/task_tools_test.go` — added UT-049-010, UT-049-011, UT-049-012 (MCP accept from in_review/changes_requested, reject from pending) and IT-049-001, IT-049-002 (persistence and terminal-state enforcement via mock repo).

### Tests added
- Domain: 50 new test cases (UT-049-001 through UT-049-009, including 40-entry table-driven matrix). Total domain: 64 passing.
- Handler: 5 new test functions (UT-049-010/011/012 + IT-049-001/002). Total suite: 388 passing.

### Coverage
- `status_machine.go` `IsValidTransition` (Task): **85.7%** — above 80% threshold.
- `status_machine.go` `NewTask`: 100%.

### Review gate evidence
```
REVIEW GATE: PASS  (be services/agent-board)
REVIEW GATE: PASS  (cross)
```

### Robot dryrun
```
19 tests, 19 passed, 0 failed
```
(`robot --dryrun tests/e2e/REQ008_requirement_entity_and_project_path/`)

### Implementation notes
No production changes to `task_tools.go` were required — the handler already delegates status validation to `IsValidTransition`. Once the domain constant and transitions were in place, all MCP handler tests passed without further handler changes, exactly as the architecture extract specified.

## Review log
