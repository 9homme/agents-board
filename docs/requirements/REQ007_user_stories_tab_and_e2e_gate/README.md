# REQ007 — User Stories Tab + E2E Quality Gate + Health-Check Fixes

**Status:** draft (intake complete + human revisions applied 2026-06-07; awaiting `system-architect` → `architecture.md`)
**Source:** human requirement (2026-06-07) + `docs/tech_debt.md` line 113 (REQ006/US012 deferred fix)
**Audience:** mixed — end users (User Stories tab) + internal (CI gate, Makefile health-checks).
**Track mix:** BE-prod + FE + CI/infra.

## Business goal

Three related threads bundled into one requirement:

1. **Unblock the e2e stack.** Close the two pre-existing `make e2e-up` health-check bugs recorded in `docs/tech_debt.md` (api-server polls a DB-dependent endpoint; mcp-server curl hangs on the SSE stream). Once green, the stack can come up unattended — which is the prerequisite for thread 2.
2. **Make e2e a PR quality gate.** Add a GitHub Actions workflow so every pull request runs the Robot Framework e2e suite and blocks merge on failure. This gives the same "equivalent evidence" the workaround produces, but automatically and on every PR.
3. **Build out the User Stories tab.** The tab is currently a static "Coming soon" placeholder. Turn it into a real feature: a list of user-story cards that, on click, opens a right-side drawer showing the full story plus all of its tasks with detail (the card list stays visible behind the drawer).

Threads 1 and 2 are tightly coupled (CI cannot rely on `make e2e-up` until the health-checks are fixed). Thread 3 is independent and can run in parallel.

## Decisions confirmed by the human (2026-06-07 revision)

The decisions below were confirmed (and several revised) by the human after the initial intake. Stories have been revised to match.

- **D-001 REVISED — migrations at startup, no `/healthz` (US001).** The original `/healthz` endpoint is **dropped**. Instead, the api-server binary auto-runs all `.up.sql` files from `services/agent-board/migrations/` on boot, before accepting traffic. Application tables therefore exist as soon as the server is up, so the existing `make e2e-up` poll on `GET /api/v1/projects` works without a prior seed. (US001 was completely rewritten from the healthz story to the migration-at-startup story.)
- **D-002 — mcp-server health-check (US002).** Add `--max-time 5` to the existing SSE health-check curl so it cannot hang. No new mcp endpoint. (Matches the tech-debt note.)
- **D-003 CONFIRMED — CI trigger (US003).** Workflow runs on `pull_request` targeting `main` **only** — not on push-to-main and not on other branches. Uses `docker compose` (GitHub-hosted Ubuntu runners ship Docker). The branch-protection "required check" toggle is a human/repo-admin action, noted as a dependency.
- **D-004 — Gate strictness (US003).** All e2e tests must pass; **no retries**. Robot artifacts (`output.xml`, `log.html`, `report.html`) are **always uploaded** (even on failure) for debugging.
- **D-005 CONFIRMED — Card fields (US004).** Each card shows: story **title**, **status** badge, a **task count** ("N tasks"), and a **description preview** (first ~80 chars). Clicking a card opens the detail drawer (US005).
- **D-006 REVISED — Detail layout (US005).** Detail is shown in a **right-side drawer/panel** (NOT an in-tab view that replaces the list, NOT a dedicated route). The card list stays visible behind the drawer. The drawer shows the story's title, status, full description, and a list of all tasks (each with title, status, description), and has a **close button** (Escape also closes it).
- **D-007 — New REST endpoints required.** No REST endpoints currently expose user stories or tasks (they exist only as MCP tools + repos). The FE work depends on the System Architect specifying new `GET` endpoints on api-server. Proposed (architect to finalise exact contract): `GET /api/v1/projects/{id}/user-stories` (must include `taskCount` and full `description`), `GET /api/v1/user-stories/{id}`, `GET /api/v1/user-stories/{id}/tasks`. Domain types already exist (`domain.UserStory`, `domain.Task`).
- **D-008 — No filtering/sorting in v1.** Cards render in the API's returned order. Filtering, sorting, and search are out of scope for this requirement.

## Open questions for the human

None outstanding — all intake decisions (D-001 through D-008) are confirmed or revised as above.

## Stories

| US ID | Title | Track | Service / Area |
|---|---|---|---|
| US001 | api-server runs DB migrations at startup | BE | services/agent-board |
| US002 | Makefile `e2e-up` mcp `--max-time` fix + data-only `e2e-seed` | CI/infra | Makefile |
| US003 | GitHub Actions e2e PR quality gate (PR-to-main only) | CI/infra | .github/workflows |
| US004 | User Stories tab — story cards list (title, status, task count, description preview) | FE + BE | web + services/agent-board |
| US005 | User story detail + tasks in a side drawer on card click | FE + BE | web + services/agent-board |

## Tasks

| Task | Title | Blocked by | Status |
|---|---|---|---|
| `US001_be_migrations_at_startup.md` | Auto-run migrations | | pending |
| `US002_be_makefile_healthcheck.md` | Makefile health-check fix | `US001_be_migrations_at_startup.md` | pending |
| `US003_be_github_actions_gate.md` | GitHub Actions gate | `US002_be_makefile_healthcheck.md` | pending |
| `US004_be_user_stories_list.md` | BE: User stories list | | completed |
| `US004_fe_user_stories_list.md` | FE: User stories card list | | pending |
| `US005_be_user_story_detail.md` | BE: User story detail | `US004_be_user_stories_list.md` | pending |
| `US005_fe_user_story_detail.md` | FE: User story drawer | `US004_fe_user_stories_list.md` | pending |

## Dependencies & sequencing

- **US002 depends on US001** (migrations-at-startup makes the existing `GET /api/v1/projects` poll succeed on a fresh stack and lets `e2e-seed` drop its migration step).
- **US003 depends on US002** (CI relies on a working `make e2e-up`/`e2e-seed`/`e2e-run` chain).
- **US004 and US005 depend on the architecture** specifying the new REST endpoints (D-007). US005 depends on US004 (the drawer is reached by clicking a card).
- US001/US002/US003 (the infra thread) and US004/US005 (the FE thread) are independent and can be developed in parallel.

## Notes for the architect

- api-server is Echo; boot/wiring is in `services/agent-board/cmd/api-server/main.go`. US001 runs `.up.sql` migrations from `services/agent-board/migrations/` in the boot sequence (after the DB connection, before `e.Start(...)` begins listening). No `/healthz` endpoint (D-001 revised).
- mcp-server is a **separate binary** (`cmd/mcp-server`) on `:8081` exposing SSE; D-002 keeps the SSE poll, just bounded with `--max-time 5`.
- New user-story/task REST endpoints must define **exact JSON contracts** so BE and FE can develop in parallel. Domain shapes: `UserStory{id,projectId,title,description,status,createdAt,updatedAt}`, `Task{id,userStoryId,title,description,status,createdAt,updatedAt}`. Existing list responses are wrapped (e.g. `{"projects":[...]}`, `{"documents":[...]}`) and single-resource responses are bare objects — keep that convention.
- FE talks to the backend only through `web/lib/api/`; types in `web/lib/api/types.ts` must match the contract field-for-field; MSW handlers mirror the exact JSON.
