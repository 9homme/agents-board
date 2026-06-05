# US002/be_user_story_repo_error_tests

**Requirement:** REQ006
**Story:** US002
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** none
**Worked-by:**
**Implements:** REQ006/US002 AC (all four scenarios — 12 verbatim test function names, ≥95% per-file coverage modulo §4.5 exemptions, no production-code change, existing suite still green). Architecture §3 US002 touch row + §4.2 cluster-1 sqlmock pattern + §4.5 exemption mechanism + §4.6 local verification command (US002 row).

## Goal
Backfill `user_story_repo.go` error-branch tests so per-file statement coverage clears ≥95% (modulo enumerated unreachable lines), with the 12 verbatim test functions named in the US002 AC, following the architecture §4.2 sqlmock pattern. Tests-only — `user_story_repo.go` is byte-for-byte unchanged.

## Scope
- **In:** Edit `services/agent-board/internal/repo/user_story_repo_test.go` to add the 12 functions enumerated in US002 AC. Use the sqlmock branch→shape mapping from architecture §4.2. For `_AuditInsertError` / `_CommitError` declare `ExpectRollback()` per the §4.2 note on `user_story_repo.go:65-71`.
- **Out:** Any change to `user_story_repo.go` itself (production code untouched). Any change to `task_repo*`, `audit_repo*`, `document_repo*`, `project_repo*`. Doc-comment-vs-code mismatches — raise `ARCHITECTURE_GAP_FOUND` if surfaced.

## Files touched (estimated, exclusive)
- `services/agent-board/internal/repo/user_story_repo_test.go` (edit — add 12 test functions)

## Test contract
Dev makes these spec IDs pass (from `US002_be_unit_tests.md` once tester authors it). The 12 verbatim function names from US002 AC are the authoritative list. Same shape as US001 (Create/Get/Update split, plus `UpdateUserStoryStatus` 5-branch transactional set, plus `ListUserStories` Query/Scan/RowsErr triplet). Tester's UT-* IDs map 1:1 onto these names.

## Implementation notes
- **Reference pattern:** architecture §4.1 — `document_repo_test.go` and `project_repo_test.go` from REQ005/US005; this story is its lineal descendant.
- **Branch → sqlmock idiom:** identical to US001 — see architecture §4.2 mapping table verbatim.
- **`UpdateUserStoryStatus` runs inside a transaction** (`BeginTx` → `QueryRowContext` → `ExecContext` → `Commit`) with five distinct error exits. Tester may consolidate via `t.Run` per US002 AC; either shape acceptable.
- **The rollback `log.Printf` path (`user_story_repo.go:68`) is unreachable via sqlmock** — leave uncovered, name in test report under OQ-4 / architecture §4.5.
- **Read the source first** for the exact `fmt.Errorf("failed to ...: %w", err)` wrap prefixes — assertion strings must match the production source literally.
- **Coverage check command** (architecture §4.6, US002 row):
  ```
  cd services/agent-board && go test ./internal/repo -coverprofile=/tmp/repo.out -run TestUserStoryRepo
  go tool cover -func=/tmp/repo.out | grep user_story_repo.go
  ```
  Must show ≥95% statement coverage on `user_story_repo.go`.

## Definition of done
- All 12 new test functions present with the exact verbatim names from US002 AC; all green via `cd services/agent-board && go test ./internal/repo -cover -v -run TestUserStoryRepo`.
- `cd services/agent-board && go vet ./... && go test ./...` clean across the whole module.
- `go tool cover -func=/tmp/repo.out | grep user_story_repo.go` shows **≥95%** statement coverage on `user_story_repo.go` (modulo line :68 per §4.5).
- `user_story_repo.go` byte-for-byte unchanged.
- `golangci-lint run ./...` clean.
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` + `scripts/review/run-gate.sh cross` — both emit `REVIEW GATE: PASS`.
- **Live e2e NOT required** (tests-only, prod unchanged — architecture §10.4); instead 3 clean runs of `cd services/agent-board && go test -count=3 ./internal/repo -race`.
- Dev set status to `in_review`; tech-lead approved.

## Review log
