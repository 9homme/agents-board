# US009 — Test Report — Story State Machine

**Requirement:** REQ003 — status_state_machine
**Story:** US009_story_state_machine
**Captured:** 2026-05-19
**Commit:** `f86fa43`
**Captured by:** orchestrator (Phase 3c)

## Toolchain

- Go `go1.26.3` — `go test ./...` inside `services/agent-board`
- Robot Framework `7.4.2` (Python 3.10.11) — `tests/e2e/REQ003_status_state_machine/US009_story_state_machine.robot`, run against a live `mcp-server` on `:8080` + Postgres `agent_board`
- Jest — N/A (REQ003 has no FE tasks)

## BE summary (UT / IT — from `US009_be_unit_tests.md`)

| Test ID | Scenario | Go test | Result |
|---|---|---|---|
| UT-001 | Valid forward transitions | `TestUserStory_IsValidTransition` | PASS |
| UT-002 | Review cycle transitions | `TestUserStory_IsValidTransition` | PASS |
| UT-003 | Circuit breaker transition | `TestUserStory_IsValidTransition` | PASS |
| UT-004 | Invalid transitions are rejected | `TestUserStory_IsValidTransition` | PASS |
| UT-005 | Enforce initial state on creation | `TestNewUserStory_EnforceInitialState`, `TestUserStoryTools_CreateUserStory_InvalidInitialStatus` | PASS |
| IT-001 | Reject invalid transitions at MCP layer | `TestUserStoryTools_UpdateUserStory_InvalidTransition` | PASS |

Supporting repo-layer tests for the transactional status + audit write: `TestUserStoryRepo_UpdateUserStoryStatus`, `TestUserStoryRepo_UpdateUserStoryStatus_RollbackOnAuditFailure` — PASS.

## FE summary (FCT)

N/A — REQ003 has no frontend tasks.

## E2E summary (E2E — from `US009_e2e_tests.md`)

| Test ID | Scenario | Result |
|---|---|---|
| E2E-001 | Valid story state machine transitions | PASS |
| E2E-002 | Invalid story state machine transition rejected | PASS |
| E2E-003 | Enforce initial state on story creation | PASS |

## Skipped tests

None.

## Notes

- All BE unit/integration tests green; `go vet` and the BE + cross review gates passed at review time.
- E2E suite previously failed to parse responses due to a wrong JSON path (`['params']['result']`) in the REQ003 `.robot` files; corrected by the tester (revision) to `['result']`. Re-run is 3/3 PASS.
