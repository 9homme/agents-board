# US005 — E2E test specification

**Owner:** tester.

## Why e2e does not apply

US005 adds 16 `sqlmock`-driven unit tests to `internal/repo/document_repo_test.go` and `internal/repo/project_repo_test.go`. This is a test-only addition — no production code changes. The correctness guarantees are:

- Each error branch in the repo layer returns a non-nil error with the correct wrapping (UT-US005-001 through UT-US005-016 in `US005_be_unit_tests.md`).
- Coverage thresholds are verified via `go test -coverprofile`.

These are pure Go unit tests verifiable with `go test ./internal/repo`. No web service is needed; no browser interaction; no HTTP surface. The docker-compose stack (US008) is irrelevant for test-only changes.

**Verdict: No e2e scenarios. Repo-layer test backfill is unit-level; covered by UT-US005-001 through UT-US005-016.**
