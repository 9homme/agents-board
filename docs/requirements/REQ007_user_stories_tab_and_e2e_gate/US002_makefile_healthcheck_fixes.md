# US002 — Makefile `e2e-up` health-check fix + data-only `e2e-seed`

**Requirement:** REQ007 — User Stories Tab + E2E Quality Gate + Health-Check Fixes
**Status:** draft

## Story
As a developer, I want `make e2e-up` to wait on a bounded mcp-server probe (so it cannot hang on the SSE stream) and `make e2e-seed` to be data-only (migrations now run at api-server startup), so that `make e2e-up && make e2e-seed` completes on a fresh stack instead of hanging.

## Acceptance criteria
- **Scenario: mcp-server health-check cannot hang on the SSE stream**
  - Given a fresh stack started by `make e2e-up`
  - When the mcp-server health-check curl runs against `http://localhost:8081/sse`
  - Then the curl invocation includes `--max-time 5` (or an equivalent time bound) so it returns within the bound even though SSE holds the connection open
  - And the health-check still treats the mcp-server as healthy on the expected response (connection established / accepted status), not solely on a clean exit
- **Scenario: api-server health-check poll succeeds on a fresh stack**
  - Given a fresh stack started by `make e2e-up` (DB not yet seeded)
  - When the api-server health-check loop polls `GET http://localhost:8080/api/v1/projects`
  - Then the poll succeeds (`200`) as soon as the api-server is up
  - And it does NOT time out waiting for tables, because the api-server applied migrations at startup (US001)
  - (no change to the poll target is required by this story — US001 makes the existing poll work)
- **Scenario: `e2e-seed` is data-only, no migration step**
  - Given the `make e2e-seed` target
  - When it runs against a stack already brought up by `make e2e-up`
  - Then it seeds application/test data only
  - And it does NOT run a migration step (migrations are applied by the api-server at startup — US001)
- **Scenario: full bring-up + seed completes end-to-end on a fresh stack**
  - Given a clean environment (no prior containers/volumes)
  - When a developer runs `make e2e-up` followed by `make e2e-seed`
  - Then `make e2e-up` reaches its "stack is healthy" success line without timing out or hanging
  - And `make e2e-seed` then seeds data successfully against the already-migrated schema
- **Scenario: no regression to the web health-check**
  - Given the existing `make e2e-up` web probe (`GET http://localhost:3000/`)
  - When US002 changes are applied
  - Then the web probe behavior is unchanged

## UI / UX flow expectations
No UI: this is a Makefile / developer-tooling change with no frontend surface.

## Out of scope
- Changing the api-server health-check poll target — it stays `GET /api/v1/projects` and now works because of US001 (migrations at startup). This story only fixes the mcp probe and makes `e2e-seed` data-only.
- Rewriting the compose file's container-level `healthcheck:` directives (this story targets the Makefile probe loops; if the architect prefers compose-level health-gating, that is a separate decision to surface via `ARCHITECTURE_GAP_FOUND`).
- Changing timeouts/retry counts beyond what is needed to fix the mcp hang.
- The mcp-server gaining a `/healthz` endpoint (D-002 keeps the bounded SSE poll).

## Dependencies
- **US001** — migrations-at-startup makes the existing `GET /api/v1/projects` poll succeed on a fresh stack and lets `e2e-seed` drop its migration step.

## Notes for the team
- **D-002 — CONFIRMED:** add `--max-time 5` to the mcp-server SSE health-check curl. This is the only health-check curl fix in this story.
- **D-001 REVISED — CONFIRMED:** the original api-server circular-dependency fix (re-pointing the poll at a DB-free endpoint) is **removed** from this story — it is now resolved by US001 (migrations at startup). The api-server poll stays on `GET /api/v1/projects`.
- Current offending line is in `Makefile` `e2e-up`: the mcp-server loop curling `http://localhost:8081/sse` without `--max-time`. Also update `e2e-seed` to remove its migration invocation (migrations now run at api-server startup).
- **Testability note for the tester:** these are shell/Make changes. Prefer asserting on the Makefile content / `make -n` dry-run output (mcp curl carries `--max-time`; `e2e-seed` has no migration step) plus an e2e bring-up smoke check where the environment allows it. Keep heavy container bring-up in e2e, not unit.
- After this lands, **update `docs/tech_debt.md` line 113** to mark the item resolved (the dev/tech-lead handles the debt-log update per project convention; po-ba notes it here for traceability).

## Sign-off log
(po-ba appends here on each sign-off pass)
