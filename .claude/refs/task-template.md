# Shared ref — engineering task file template

Authoritative template for a `Track: BE` / `Track: FE` task file written by **tech-lead-planner** in Phase 2 and consumed by **be-dev** / **fe-dev** in Phase 3 and **tech-lead-reviewer** at review time.

Single source of truth. Do not re-inline this template into agent definitions — link here.

Path convention: `docs/requirements/REQ[ID]_*/US[ID]_[task_name].md` (`snake_case` task name; BE tasks conventionally `be_<thing>`, FE `fe_<thing>` — the `Track:` field is authoritative).

```markdown
# US[ID]/[task_name]

**Requirement:** REQ[ID]
**Story:** US[ID]
**Track:** BE | FE
**Service:** services/<name>   (only for Track: BE; omit for FE)
**Status:** pending | in_progress | in_review | changes_requested | completed | blocked_circuit_breaker | blocked_review_gate
**Blocked by:** [list of other task filenames, or none]
**Worked-by:** [filled in by the dev when they claim the task — leave blank]
**Implements:** [architecture decision IDs / API contract endpoints / FE surface rows / data-model items this task realises]

## Goal
One sentence: what changes when this task is merged.

## Scope
- **In:** [files/packages to touch, interfaces to add, migrations to write]
- **Out:** [explicit non-goals — keep PRs small]

## Files touched (estimated, exclusive)
List the concrete file paths this task is expected to create or modify. The orchestrator's 3a tick uses this list to ensure no two parallel-spawned devs collide on the same file (worktree isolation prevents torn writes, but a merge conflict at integration still wastes a spawn). Be conservative — overestimating costs nothing; underestimating causes a re-queue.
- e.g. `services/basket/internal/repo/basket_repo.go`
- e.g. `services/basket/internal/repo/basket_repo_test.go`
- e.g. `services/basket/migrations/0003_basket_items.up.sql`

If this task touches a shared scaffold file that other tasks in the same story would otherwise need (`go.mod`, `web/package.json`, `web/lib/api/types.ts`, `web/lib/api/client.ts`, `tests/e2e/resources/common.resource`, or migration-number space), say so explicitly and mark the task as a **scaffold task** in this section. Other tasks in the same story should `Blocked by:` this scaffold task so the orchestrator runs it solo before parallelising the rest.

## Architecture extract
**Self-contained.** tech-lead-planner copies the exact architecture content this task implements directly here, so the dev never opens `architecture.md`. Copy verbatim (do not paraphrase or re-design):
- the **exact JSON request/response contract** for every endpoint the task touches, per status code, field-for-field, with the architect's example values;
- the **error envelope shape** and every status code's body;
- the **data-model rows** (tables/columns/indexes) the task creates or reads;
- the full text of each cited **decision (D-NNN)** entry;
- any relevant **FE surface row** / data-flow note (FE tasks).

If this section is empty or vague, the dev reports it back rather than opening `architecture.md` — a populated extract is the planner's deliverable, not optional.

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US[ID]_be_unit_tests.md`: UT-00X, IT-00Y
- (Track: FE) from `US[ID]_fe_unit_tests.md`: FCT-00X

A task lists test IDs from only its track's spec file. If new cases are needed beyond the spec, the dev writes them but flags the addition back to tester for review.

## Implementation notes
- [package layout, suggested function signatures, error types, logging keys]
- [migration SQL outline if applicable]
- [config/env vars to add]

## Definition of done
- All listed tests green.
- (Track: BE) `go vet ./...` and `go test ./...` clean inside the task's service module.
- (Track: BE) `go test -coverprofile=/tmp/cov.out ./... && go tool cover -func=/tmp/cov.out` — every production `.go` file in this task's `## Files touched` clears ≥ 80% line coverage, OR the task has a written `## Coverage exemption` section justifying each below-threshold file.
- (Track: FE) `npm run typecheck` and `npm test` clean in `web/`. No `any` types added without justification.
- (Track: FE) `cd web && npm test -- --coverage --watchAll=false --forceExit` — every non-test `.ts` / `.tsx` file in this task's `## Files touched` clears ≥ 80% line coverage, OR the task has a written `## Coverage exemption` section.
- No new public exports / public components without a doc comment.
- Code matches the `## Architecture extract` above (no silent deviation).
- (Track: FE) react-doctor evidence in `## Notes` (verbatim `--diff` score line, no regression).
- **Review gate green (dev runs it once before hand-off):** the dev runs the per-track gate + cross gate per `.claude/refs/review-gate.md`, both emit `REVIEW GATE: PASS`, and the dev pastes those lines into `## Notes`. A gate that cannot reach PASS → `blocked_review_gate`, never `in_review`.
- Dev set status to `in_review` and reported back.

## Notes
(dev appends here at hand-off: files touched, tests added, gate PASS lines, FE react-doctor score line, follow-ups, and on rework a per-item response to the previous review pass)

## Review log
(tech-lead-reviewer appends here on each review pass)

### Review pass N — YYYY-MM-DD — verdict: approved | changes_requested | blocked_review_gate
- [observation / required change / file:line]
- ...
```
