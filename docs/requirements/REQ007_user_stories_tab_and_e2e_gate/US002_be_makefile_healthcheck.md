# US002/be_makefile_healthcheck

**Requirement:** REQ007
**Story:** US002
**Track:** BE
**Service:** services/agent-board
**Status:** pending
**Blocked by:** US001_be_migrations_at_startup.md
**Worked-by:** 
**Implements:** US002, US002 Makefile changes, D-002

## Goal
Fix `e2e-up` mcp-server health check to not hang by adding a timeout, and make `e2e-seed` data-only by removing its migration step. Also resolve tech debt line 113.

## Scope
- **In:** `Makefile` updates for `e2e-up` and `e2e-seed`. Resolving line 113 in `docs/tech_debt.md`.
- **Out:** Other Makefile targets. Changing the api-server health check target.

## Files touched (estimated, exclusive)
- `Makefile`
- `docs/tech_debt.md`

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US002_be_unit_tests.md`: All applicable UT/IT tests.

## Implementation notes
- Add `--max-time 5` to the `curl` commands polling the mcp-server `/sse` endpoint in `make e2e-up`.
- Remove the migration step (`psql "$(PG_CONN)" -f migrations/...`) from `make e2e-seed`. Keep the seed loop. Update target help text.
- Remove line 113 from `docs/tech_debt.md` and append a resolved note or delete it per repo convention.

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- (Track: BE) `go test -coverprofile=/tmp/cov.out ./... && go tool cover -func=/tmp/cov.out` — every production `.go` file in this task's `## Files touched` clears ≥ 80% line coverage, OR the task has a written `## Coverage exemption` section justifying each below-threshold file.
- No new public exports / public components without a doc comment.
- Code matches the cited architecture entries (no silent deviation).
- **Review gate green:** `scripts/review/run-gate.sh be services/agent-board` exits 0 AND emits `REVIEW GATE: PASS` on stdout. Also `scripts/review/run-gate.sh cross` exits 0 AND emits `REVIEW GATE: PASS`. If the REQ has Robot e2e suites, `robot --dryrun tests/e2e/REQ007_*/` also passes.
- Dev set status to `in_review` and reported back.

## Review log
