# US002/be_story_state_machine

**Requirement:** REQ003
**Story:** US002
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** US001_be_scaffold_domain_and_migration.md
**Worked-by:** 
**Implements:** US002, D-003, part of US003

## Goal
Enforce the User Story status state machine on creation and updates, and transactionally write the state transitions to the audit trail.

## Scope
- **In:** Update `internal/repo` and `internal/handler` for User Stories to validate state transitions and initial state, and write to the audit trail on change.
- **Out:** Reading from the audit trail (handled in US003). Task status state machine.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/repo/user_story_repo.go`
- `services/agent-board/internal/repo/user_story_repo_test.go`
- `services/agent-board/internal/handler/user_story_tools.go`
- `services/agent-board/internal/handler/user_story_tools_test.go`

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US002_be_unit_tests.md`: Check specific UT and IT IDs for the story state machine and audit writing.

## Implementation notes
- Handlers should use the domain logic from `status_machine.go` to validate transitions/initial state before applying them.
- Updates must be transactional (update story status and insert into `status_audit_trail` in one DB transaction).
- `create_user_story` should ensure the initial status defaults to or is explicitly set to `draft`.

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Review log
