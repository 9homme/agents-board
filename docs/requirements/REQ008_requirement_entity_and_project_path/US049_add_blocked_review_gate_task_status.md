# US049 — Add `blocked_review_gate` task status to the state machine

**Requirement:** REQ008 — Requirement entity + project local-path linking
**Status:** draft

## Story
As an agent tracking task state via MCP, I want a `blocked_review_gate` task status with valid transitions into it from `in_review` and `changes_requested`, so that I can record when a task is blocked by a failing review-gate tool (a tooling problem, not a code problem) instead of being rejected by `IsValidTransition`.

## Background
CLAUDE.md's task state machine documents `blocked_review_gate` as the status a reviewer sets when the review gate did NOT emit `REVIEW GATE: PASS` because the gate/tooling is at fault (not the code). The Go domain layer in `services/agent-board/internal/domain/status_machine.go` currently defines only `blocked_circuit_breaker` among the blocked states. Because `blocked_review_gate` is not a known transition target, `Task.IsValidTransition` returns `false` for it, so MCP `update_task` cannot move a task into that status.

## Acceptance criteria
- **Scenario: Constant exists**
  - Given the Go domain layer
  - When the code is compiled
  - Then a `TaskStatusBlockedReviewGate` constant exists with the string value `"blocked_review_gate"`
- **Scenario: Transition from in_review is valid**
  - Given a `Task` with `Status` = `in_review`
  - When `IsValidTransition("blocked_review_gate")` is called
  - Then it returns `true`
- **Scenario: Transition from changes_requested is valid**
  - Given a `Task` with `Status` = `changes_requested`
  - When `IsValidTransition("blocked_review_gate")` is called
  - Then it returns `true`
- **Scenario: blocked_review_gate is terminal (no transitions out)**
  - Given a `Task` with `Status` = `blocked_review_gate`
  - When `IsValidTransition` is called with any status value (e.g. `in_progress`, `in_review`, `completed`, `changes_requested`, `pending`, `done`, `blocked_circuit_breaker`)
  - Then it returns `false` for every value
- **Scenario: Cannot reach blocked_review_gate from non-allowed states**
  - Given a `Task` with `Status` in {`pending`, `in_progress`, `completed`, `blocked_circuit_breaker`}
  - When `IsValidTransition("blocked_review_gate")` is called
  - Then it returns `false` (only `in_review` and `changes_requested` may transition into `blocked_review_gate`)
- **Scenario: Existing transitions are unaffected**
  - Given the existing task transitions (`pending → in_progress`, `in_progress → in_review`, `in_review → {completed, changes_requested}`, `changes_requested → {in_progress, in_review, completed, blocked_circuit_breaker}`, and the terminal `completed` / `blocked_circuit_breaker` states)
  - When `IsValidTransition` is evaluated for each of those existing edges
  - Then every previously-valid transition still returns `true` and every previously-invalid transition still returns `false` (adding `blocked_review_gate` introduces no regressions)
- **Scenario: MCP update_task accepts the new status**
  - Given a task in `in_review` (or `changes_requested`)
  - When the MCP `update_task` tool is invoked with `status` = `blocked_review_gate`
  - Then the update succeeds (no `invalid status transition` error) and the task's persisted status becomes `blocked_review_gate`
- **Scenario: MCP update_task rejects an invalid transition into the new status**
  - Given a task in `pending`
  - When the MCP `update_task` tool is invoked with `status` = `blocked_review_gate`
  - Then the update is rejected with the invalid-status-transition error and the task's status is unchanged

## UI / UX flow expectations
No UI: this is a backend-only change to the domain state machine and the MCP `update_task` tool. There is no web surface for setting task status to `blocked_review_gate` in this story.

## Out of scope
- Adding `blocked_review_gate` (or any equivalent) to the **UserStory** state machine — this story covers Task only.
- Any DB schema change or migration (see Dependencies / Notes — the status column is free-text at the DB level).
- Web UI to display or set the `blocked_review_gate` status.
- Circuit-breaker logic changes — `blocked_circuit_breaker` is untouched.

## Dependencies
- None. Independent of US044–US048. This story only touches `services/agent-board/internal/domain/status_machine.go` (and the MCP `update_task` path that already calls `IsValidTransition`); it does not depend on the Requirement entity, re-parenting, or the `Project.path` work.
- **No DB migration needed:** the task `status` column is `TEXT` and already accepts any string at the database level; the only enforcement is in the Go domain layer (`IsValidTransition`). This story adds enforcement, not a schema change.

## Notes for the team
- Track: BE. Service: `services/agent-board`.
- Mirror the existing `blocked_circuit_breaker` treatment exactly:
  - In `IsValidTransition`, add `TaskStatusBlockedReviewGate` as an allowed target from the `in_review` case and from the `changes_requested` case.
  - Add `TaskStatusBlockedReviewGate` to the terminal-states `case` alongside `TaskStatusCompleted, TaskStatusBlockedCircuitBreaker` so it returns `false` (no transitions out).
- Tester: assert the full transition matrix to lock the "existing transitions unaffected" AC — a table-driven test enumerating every (from, to) pair is the cleanest way to prove no regression.
- No new MCP tool is needed — `update_task` already delegates to `IsValidTransition`; the new status flows through once the domain layer allows it. Add an MCP-level test (UT/IT) covering the accept and reject scenarios above.

## Sign-off log
(po-ba appends here on each sign-off pass)
