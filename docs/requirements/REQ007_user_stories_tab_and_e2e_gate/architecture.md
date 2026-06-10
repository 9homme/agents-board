# Architecture — REQ007 User Stories Tab + E2E Quality Gate + Health-Check Fixes

**Approval:** approved
**Approved-by:** human
**Approved-at:** 2026-06-08T00:00:00Z

## Scope
- **In:**
  - **US039** — `services/agent-board` api-server auto-applies `.up.sql` migrations at boot, before listening, idempotently, aborting startup on failure.
  - **US040** — `Makefile` `e2e-up` mcp-server probe bounded with `--max-time 5`; `e2e-seed` becomes data-only (migration step removed). `docs/tech_debt.md` line 113 marked resolved.
  - **US041** — `.github/workflows/e2e.yml` GitHub Actions workflow: `pull_request → main` runs the full Robot e2e suite, blocks merge on failure, always uploads artifacts, always tears down.
  - **US042 / US043** — three new read-only REST endpoints on api-server exposing user stories and tasks, plus the FE User Stories tab (card list + right-side detail drawer with tasks).
- **Out:**
  - No `/healthz` / readiness endpoint (D-001 REVISED supersedes the original healthz story).
  - No down-migrations / rollback at startup (`.up.sql` only).
  - No standalone migrate CLI (migrations run inside the api-server boot; the migration logic lives in a reusable package but is not wired to a separate binary in this requirement).
  - No create/edit/delete of stories or tasks (read-only views).
  - No filtering / sorting / search / pagination (D-008).
  - No deep-link/URL for a selected story (in-tab selection state only — D-006).
  - No branch-protection "required check" toggle (repo-admin action in GitHub settings, external dependency).
  - No compose-level `healthcheck:` rewrite (US040 targets the Makefile probe loops only).
  - No mcp-server changes beyond the Makefile probe (no new mcp endpoint).

## Service topology
| Service | New / Modified | Responsibility | Inter-service calls |
|---|---|---|---|
| `services/agent-board` (api-server binary) | modified | Boot-time DB migrations (US039); new read-only REST endpoints for user stories + tasks (US042/US043) | none (talks only to Postgres) |
| `services/agent-board` (mcp-server binary) | unchanged | MCP SSE/tools — untouched by this requirement | none |
| Postgres | existing | Single shared DB for the agent-board module | — |
| GitHub Actions (CI) | new | PR-to-main e2e quality gate orchestrating the compose stack + Robot suite via Makefile targets | drives `docker compose` + `make` |

Both binaries share the same Go module (`services/agent-board`), the same DB, and the same `internal/...` packages. Only the api-server boot path and the api-server HTTP routes change.

## Frontend surface
| Route (`web/pages/...`) | New / Modified | Owns these user actions | Backend endpoints used |
|---|---|---|---|
| `/projects/[id]?tab=user-stories` | modified (wiring only) | switches to the User Stories tab | — (delegates to `UserStoriesTab`) |

The route file `web/pages/projects/[id].tsx` changes in exactly one place: it passes `projectId` into `UserStoriesTab` (currently rendered with no props). No new page/route is added — the drawer is in-tab state, not a route (D-006).

| Component group | Path | New / Modified | Responsibility |
|---|---|---|---|
| Tab body | `web/components/ProjectDetail/UserStoriesTab.tsx` | modified (replaces placeholder) | Orchestrates list + drawer; owns selection state |
| Card list | `web/components/ProjectDetail/UserStoryCardList.tsx` | new | Renders the grid of cards; loading/empty/error states |
| Card | `web/components/ProjectDetail/UserStoryCard.tsx` | new | One story card (title, status badge, "N tasks", ~80-char description preview); focusable button |
| Drawer | `web/components/ProjectDetail/UserStoryDrawer.tsx` | new | Right-side drawer: story detail + tasks list; close button + Escape; focus management |

- **API client layer:** `web/lib/api/userStories.ts` (new) — every backend call lives here; components never call `fetch` directly. Types in `web/lib/api/types.ts` extended to match the contract field-for-field. MSW handlers in `web/test/msw/` mirror the exact JSON. (Note: existing tests use per-file MSW server setup as in `documents.test.ts`; follow that precedent.)
- **Hooks (new):** `web/hooks/useProjectUserStories.ts`, `web/hooks/useUserStory.ts`, `web/hooks/useUserStoryTasks.ts` — each mirrors the AbortController + stale-id race-safe pattern already established in `useProjectDocuments.ts` / `useDocument.ts`.

## Data flow

Two representative flows: (A) boot-time migrations (US039), (B) the card-list-then-drawer FE flow (US042/US043).

### A. Migration at startup (US039)
```mermaid
sequenceDiagram
    participant Boot as api-server run()
    participant DB as Postgres
    Boot->>DB: sql.Open + PingContext (existing)
    Boot->>DB: SELECT to_regclass('schema_migrations') / CREATE TABLE IF NOT EXISTS schema_migrations
    Boot->>DB: SELECT version FROM schema_migrations (applied set)
    loop each embedded .up.sql in lexical filename order
        alt version already applied
            Boot->>Boot: skip (idempotent)
        else not applied
            Boot->>DB: BEGIN; exec file SQL; INSERT version; COMMIT
            alt SQL error
                DB-->>Boot: error
                Boot->>Boot: log.Fatal / return err -> os.Exit(1) (NO listen)
            end
        end
    end
    Boot->>Boot: e.Start(":8080")  (only reached if all migrations succeeded)
```

### B. Card list → drawer (US042 + US043)
```mermaid
sequenceDiagram
    participant U as User (web)
    participant Tab as UserStoriesTab
    participant API as web/lib/api/userStories
    participant SVC as api-server handler
    participant DB as Postgres
    U->>Tab: open User Stories tab
    Tab->>API: fetchProjectUserStories(projectId)
    API->>SVC: GET /api/v1/projects/{id}/user-stories
    SVC->>DB: verify project exists; SELECT stories + LEFT JOIN task counts
    DB-->>SVC: rows
    SVC-->>API: 200 {"userStories":[{...,taskCount}]}
    API-->>Tab: render cards
    U->>Tab: click a card (storyId)
    Tab->>API: fetchUserStory(storyId)
    Tab->>API: fetchUserStoryTasks(storyId)
    API->>SVC: GET /api/v1/user-stories/{id}
    API->>SVC: GET /api/v1/user-stories/{id}/tasks
    SVC->>DB: SELECT story; SELECT tasks WHERE user_story_id=$1
    DB-->>SVC: rows
    SVC-->>API: 200 bare UserStory ; 200 {"tasks":[...]}
    API-->>Tab: drawer renders detail + tasks
    U->>Tab: click X / press Escape
    Tab->>Tab: clear selection -> drawer unmounts; list stays mounted
```

**Decision (D-005 below): the drawer makes two parallel requests** — `GET /api/v1/user-stories/{id}` (story detail) and `GET /api/v1/user-stories/{id}/tasks` (tasks). Tasks are NOT embedded in the story-detail response, keeping each endpoint single-responsibility and matching the existing bare-object / wrapped-list conventions.

## Components

### Backend (`services/agent-board`)
| Package | New / Modified | Responsibility |
|---|---|---|
| `internal/migrate` (new pkg, e.g. `internal/migrate/migrate.go`) | new | `//go:embed` the `migrations/*.up.sql` files; `Run(ctx, *sql.DB) error` applies pending migrations idempotently in lexical order, tracked in a `schema_migrations` table; returns error on any failure |
| `cmd/api-server/main.go` | modified | After the existing DB ping and before `e.Start(...)`, call `migrate.Run(ctx, db)`; on error, return it so `run()`'s caller exits non-zero (no listen) |
| `internal/handler/user_story_handler.go` | new | `GetProjectUserStories`, `GetUserStory`, `GetUserStoryTasks` HTTP handlers + their response DTOs; reuses existing repos |
| `internal/repo/user_story_repo.go` | modified | Add `ListUserStoriesWithTaskCount(ctx, projectID) ([]*UserStoryWithCount, error)` (single query with `LEFT JOIN tasks` GROUP BY — avoids N+1). `GetUserStory` and `ListUserStories` already exist and are reused as-is |
| `internal/repo/task_repo.go` | unchanged | `ListTasks(ctx, userStoryID)` already exists and is reused |
| `internal/domain/*` | unchanged | `UserStory` and `Task` structs already exist |

The `migrations/` directory's location matters: `//go:embed` requires the embedded files to live **inside or below the embedding package's directory tree**. The migration files currently live at `services/agent-board/migrations/`. The embedding package must reference them via a path it can see — see **D-002** for the chosen layout.

### Frontend (`web/`)
| Group | Path | New/Mod | Responsibility |
|---|---|---|---|
| Tab body | `web/components/ProjectDetail/UserStoriesTab.tsx` | mod | Owns `selectedStoryId` state; renders `UserStoryCardList` always; renders `UserStoryDrawer` when a story is selected |
| Card list | `web/components/ProjectDetail/UserStoryCardList.tsx` | new | Maps stories to cards; loading spinner / empty message / error message (mirrors DocumentsTab states) |
| Card | `web/components/ProjectDetail/UserStoryCard.tsx` | new | `<button>` (role=button, accessible name = title); shows title, status badge, "N tasks", truncated description; calls `onSelect(storyId)` |
| Drawer | `web/components/ProjectDetail/UserStoryDrawer.tsx` | new | role=dialog, aria-modal; story title/status/full description + tasks list; loading/empty/error; X button + Escape; focus moves in on open, returns to triggering card on close |
| API client | `web/lib/api/userStories.ts` | new | `fetchProjectUserStories`, `fetchUserStory`, `fetchUserStoryTasks` (each accepts optional `AbortSignal`) |
| Types | `web/lib/api/types.ts` | mod | Add `UserStoryListItem`, `UserStoriesListResponse`, `UserStory`, `Task`, `TasksListResponse` |
| Hooks | `web/hooks/useProjectUserStories.ts`, `useUserStory.ts`, `useUserStoryTasks.ts` | new | Race-safe fetch hooks mirroring `useProjectDocuments` / `useDocument` |
| MSW | `web/test/msw/` handlers (per-test setup as in existing tests) | new | Mirror the exact JSON contracts below |

## Infrastructure
- **Databases:** single existing Postgres (`agent_board`). No schema additions beyond a new **bookkeeping table** `schema_migrations` (created by the migration runner itself — see Data model). The application tables (`projects`, `documents`, `user_stories`, `tasks`, `status_audit_trail`) are unchanged and continue to be defined by the existing `.up.sql` files.
- **Caches / queues:** none.
- **External services:** GitHub Actions (CI), GitHub-hosted Ubuntu runner (Docker preinstalled).
- **Container runtime constraint (load-bearing):** api-server runs on `gcr.io/distroless/static-debian12` — **no shell, no psql, no curl** in the runtime image. Therefore migrations CANNOT be run by shelling out to psql at container start; they MUST run in-process over the existing `database/sql` connection using embedded SQL. This is the central driver for D-001/D-002.
- **Env vars added:** none. Migrations use the existing `DATABASE_URL`. FE uses the existing `NEXT_PUBLIC_API_BASE_URL`.
- **CORS:** unchanged — the new endpoints are served by the same api-server Echo instance with the existing `FRONTEND_URL` CORS config; they inherit it automatically.
- **Deployment surface change:** the api-server image now needs the migration `.up.sql` files compiled into the binary via `//go:embed` (no extra COPY in the Dockerfile if the files are embedded from within the module — see D-002). The new `.github/workflows/e2e.yml` is the only new deployment-adjacent artifact.

## API contracts (exact)

All three endpoints are **GET, read-only, no auth** (consistent with the existing `/api/v1/projects`, `/api/v1/documents` endpoints, which carry no auth). Timestamps are ISO-8601 UTC strings formatted `2006-01-02T15:04:05Z` (matching existing handlers). List responses are **wrapped**; single-resource responses are **bare objects** (existing convention, README note). The shared error envelope is `{ "code": string, "message": string }`.

### 1. GET /api/v1/projects/{id}/user-stories
- **Service:** `services/agent-board` (api-server)
- **Auth:** none
- **Path params:** `id` — project id (string, uuid)
- **Request body:** none
- **Behaviour:** verifies the project exists first (404 if not — mirrors `ListProjectDocuments`), then returns all user stories for the project, **each including `taskCount`** (count of tasks whose `userStoryId` = the story id) and the **full `description`** (truncation to ~80 chars is an FE concern). Order: `created_at DESC` (matches existing `ListUserStories`).
- **Responses:**
  - **200 OK** — array always present, never null; empty project returns `{"userStories":[]}`:
    ```json
    {
      "userStories": [
        {
          "id": "11111111-1111-1111-1111-111111111111",
          "projectId": "00000000-0000-0000-0000-000000000001",
          "title": "Add item to basket",
          "description": "As a shopper I want to add an item to my basket so that I can purchase it later.",
          "status": "in_development",
          "taskCount": 3,
          "createdAt": "2024-01-01T00:00:00Z",
          "updatedAt": "2024-01-02T09:30:00Z"
        }
      ]
    }
    ```
    Field types: `id` string(uuid), `projectId` string(uuid), `title` string, `description` string (MAY be `""`), `status` string, `taskCount` integer ≥ 0, `createdAt` string(ISO-8601 UTC), `updatedAt` string(ISO-8601 UTC).
  - **404 Not Found** — project does not exist:
    ```json
    { "code": "NOT_FOUND", "message": "Project not found" }
    ```
  - **500 Internal Server Error** — unexpected failure:
    ```json
    { "code": "INTERNAL_ERROR", "message": "Failed to fetch user stories" }
    ```
- **Idempotency:** safe GET; naturally idempotent.

### 2. GET /api/v1/user-stories/{id}
- **Service:** `services/agent-board` (api-server)
- **Auth:** none
- **Path params:** `id` — user story id (string, uuid)
- **Request body:** none
- **Behaviour:** returns the bare `UserStory` object. Tasks are NOT embedded (fetched separately — D-005).
- **Responses:**
  - **200 OK** — bare object:
    ```json
    {
      "id": "11111111-1111-1111-1111-111111111111",
      "projectId": "00000000-0000-0000-0000-000000000001",
      "title": "Add item to basket",
      "description": "As a shopper I want to add an item to my basket so that I can purchase it later.",
      "status": "in_development",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-02T09:30:00Z"
    }
    ```
    (Note: the single-story response does NOT include `taskCount` — that is a list-only aggregate field for the card. The drawer derives the task count from the tasks endpoint response length.)
  - **404 Not Found** — story does not exist:
    ```json
    { "code": "NOT_FOUND", "message": "User story not found" }
    ```
  - **500 Internal Server Error**:
    ```json
    { "code": "INTERNAL_ERROR", "message": "Failed to fetch user story" }
    ```
- **Idempotency:** safe GET.

### 3. GET /api/v1/user-stories/{id}/tasks
- **Service:** `services/agent-board` (api-server)
- **Auth:** none
- **Path params:** `id` — user story id (string, uuid)
- **Request body:** none
- **Behaviour:** verifies the user story exists first (404 if not — mirrors the documents-list-checks-project pattern), then returns all tasks for that story. Order: `created_at DESC` (a tester may pin exact order; the contract guarantees the FE renders in returned order — D-008 no client sorting).
- **Responses:**
  - **200 OK** — array always present, never null; story with no tasks returns `{"tasks":[]}`:
    ```json
    {
      "tasks": [
        {
          "id": "22222222-2222-2222-2222-222222222222",
          "userStoryId": "11111111-1111-1111-1111-111111111111",
          "title": "be_basket_repo",
          "description": "Implement the basket repository layer.",
          "status": "completed",
          "createdAt": "2024-01-01T10:00:00Z",
          "updatedAt": "2024-01-02T11:00:00Z"
        }
      ]
    }
    ```
    Field types: `id` string(uuid), `userStoryId` string(uuid), `title` string, `description` string (MAY be `""`), `status` string, `createdAt`/`updatedAt` string(ISO-8601 UTC).
  - **404 Not Found** — story does not exist:
    ```json
    { "code": "NOT_FOUND", "message": "User story not found" }
    ```
  - **500 Internal Server Error**:
    ```json
    { "code": "INTERNAL_ERROR", "message": "Failed to fetch tasks" }
    ```
- **Idempotency:** safe GET.

> **Contract freeze note for FE/BE parity:** the FE `web/lib/api/types.ts` `UserStoryListItem` includes `taskCount`; the `UserStory` (detail) type does NOT. MSW handlers and the BE DTOs must reflect that asymmetry exactly. The drawer's task-count display, if any, comes from `tasks.length`, not from the detail object.

## Data model

No application table changes. One new bookkeeping table, created by the migration runner before it applies any user migration:

```sql
-- created by internal/migrate at startup (idempotent), NOT a numbered .up.sql file
CREATE TABLE IF NOT EXISTS schema_migrations (
    version     TEXT PRIMARY KEY,   -- the .up.sql filename, e.g. "000001_init_schema.up.sql"
    applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Migration runner algorithm (per D-001):
1. `CREATE TABLE IF NOT EXISTS schema_migrations (...)`.
2. `SELECT version FROM schema_migrations` → in-memory applied set.
3. Read embedded `*.up.sql` entries, **sort by filename lexically** (`000001_…`, `000002_…`).
4. For each not in the applied set: run in a single transaction — `BEGIN; <file SQL>; INSERT INTO schema_migrations(version) VALUES($1); COMMIT;`.
5. Any error → return it (caller exits non-zero, server never listens).

**Idempotency strategy decision (D-003):** use the `schema_migrations` bookkeeping table as the primary idempotency mechanism (a fresh start applies all; a restart skips all). The existing `.up.sql` files use bare `CREATE TABLE` (no `IF NOT EXISTS`), so the bookkeeping table is required — re-running `000001_init_schema.up.sql` against a populated DB would otherwise error on "table already exists". The `.up.sql` files are NOT modified.

**Index for `taskCount` aggregate:** the existing `idx_tasks_user_story_id` index already supports the `LEFT JOIN tasks ... GROUP BY user_stories.id` count query efficiently. No new index needed.

`ListUserStoriesWithTaskCount` query sketch:
```sql
SELECT us.id, us.project_id, us.title, us.description, us.status,
       us.created_at, us.updated_at, COUNT(t.id) AS task_count
FROM user_stories us
LEFT JOIN tasks t ON t.user_story_id = us.id
WHERE us.project_id = $1
GROUP BY us.id
ORDER BY us.created_at DESC;
```

## US040 — Makefile changes (exact)

Two edits to `Makefile`, no behavioural change to the web/api probes:

1. **mcp-server probe bounded.** The current `e2e-up` mcp loop (lines ~49-53) already has a fallback that bounds the curl, but the **first** curl in the `until` (`curl -sf http://localhost:8081/sse`) is **unbounded** and can hang on the SSE stream. Add `--max-time 5` to that curl (and keep treating an accepted HTTP status as healthy, per US040 AC — health is "connection established / accepted status", not "clean exit"). The bounded form:
   ```make
   @i=0; until curl -sf --max-time 5 http://localhost:8081/sse >/dev/null 2>&1 || \
       curl -s --max-time 5 -o /dev/null -w "%{http_code}" http://localhost:8081/sse 2>&1 | grep -qE "^(200|405|404)$$"; do \
   ```
2. **`e2e-seed` becomes data-only.** Remove the migration loop (current lines ~60-63) from `e2e-seed`; keep only the seed loop over `$(SEEDS_DIR)/*.sql`. Update the target's `##` help text from "Apply migrations then seed fixtures" to "Seed data fixtures (migrations now run at api-server startup)". Migrations are applied by the api-server at boot (US039), so by the time `e2e-seed` runs (after `e2e-up` gates the api-server healthy) the schema already exists.

The api-server probe target (`GET /api/v1/projects`) and the web probe (`GET http://localhost:3000/`) are **unchanged** (US040 AC: no regression to the web probe; api poll stays as-is and now works because of US039).

> **Dependency note for `e2e-seed` after the split:** `e2e-seed` writes to the DB via `psql "$(PG_CONN)"` against the host-published Postgres port (`localhost:15432`). It depends on the api-server having already run migrations. Since `e2e-up` blocks until `GET /api/v1/projects` returns 200 (which is only true after migrations completed), the natural `make e2e-up && make e2e-seed` ordering is safe. The `dev-migrate` / `dev-seed` native targets are **out of scope** for this requirement and remain unchanged (they target a different DB/port and a non-container flow); a follow-up may retire `dev-migrate` once migrations-at-startup proves out, but that is not part of REQ007.

> **tech_debt.md:** line 113 (the REQ006/US012 deferred fix) is marked resolved by US040 per the story note; the dev/tech-lead handles that one-line edit.

## US041 — GitHub Actions workflow (shape)

- **File:** `.github/workflows/e2e.yml` (the `.github/` directory does not yet exist — it is created by this story).
- **Trigger:** `on: pull_request:` with `branches: [main]` only. No `push`. (D-003)
- **Runner:** `ubuntu-latest` (Docker + `docker compose` preinstalled).
- **Single job, steps in order:**
  1. `actions/checkout@v4`.
  2. Bring the stack up via the Makefile family (`make e2e-up`, which runs `docker compose up -d` then the health-gated probe loops fixed in US040). The runner must have `psql` available for `make e2e-seed` and Python/Robot for `make e2e-run` — install via `actions/setup-python` + `pip install` of the Robot toolchain, plus the postgres client (apt) — OR drive everything through containers; tech-lead/tester pick the simplest path that keeps CI and local in lockstep (README note). The architecture mandates: **reuse the Makefile targets, do not duplicate their commands inline.**
  3. `make e2e-seed` (data-only after US040).
  4. `make e2e-run` (runs all `tests/e2e/REQ*/` Robot suites; no `REQ`/`US` narrowing → full suite).
  5. **Upload artifacts** — `actions/upload-artifact@v4` with `if: always()`, uploading `tests/e2e/results/output.xml`, `log.html`, `report.html`.
  6. **Teardown** — `make e2e-down` with `if: always()`.
- **Gate strictness (D-004):** no retries anywhere. Any Robot failure (or stack-bring-up failure) makes a step exit non-zero → job fails → check fails → (once branch protection is enabled by a repo admin) merge is blocked.
- **External dependency:** enabling the branch-protection "required check" is a repo-admin action in GitHub settings, not a repo file — noted, not implemented here.

## Key decisions (ADR-lite)

### D-001 — Migrations run in-process at api-server boot via an embedded SQL runner + a `schema_migrations` table
- **Context:** The api-server must have application tables present the moment it accepts traffic, to break the `make e2e-up` circular dependency. The runtime container is distroless (no shell/psql), so an external migrate step at container start is impossible. Migrations must be deterministic, idempotent across restarts, and abort startup on failure.
- **Decision:** Add an `internal/migrate` package that `//go:embed`s the `.up.sql` files, creates a `schema_migrations` bookkeeping table, applies pending files in lexical filename order each inside its own transaction, and records applied versions. `cmd/api-server/main.go` calls `migrate.Run(ctx, db)` after the DB ping and before `e.Start(...)`; any error returns up through `run()` so the process exits non-zero and never listens.
- **Alternatives rejected:**
  - *Shell out to psql / a sidecar init container:* impossible/awkward — distroless has no psql; an init container adds compose complexity and a second source of truth for migration ordering.
  - *A third-party migration library (golang-migrate, goose):* heavier dependency for two tiny migrations; the embed + bookkeeping approach is ~60 lines, dependency-free, and fully testable against a throwaway Postgres. (Re-evaluate if migration count grows large or down-migrations become needed.)
  - *Rely on `CREATE TABLE IF NOT EXISTS` in the SQL instead of a tracking table:* the existing `.up.sql` files use bare `CREATE TABLE`; rewriting them is riskier and still doesn't give a clean "already applied" signal for future data-mutating migrations. The tracking table generalises better.
- **Consequences:** A new `schema_migrations` table appears in every DB. The migration files become compile-time embedded — adding a migration requires a rebuild (acceptable; the Dockerfile already rebuilds from full source). The `e2e-seed`/`dev-migrate` psql-based migration steps become redundant for the api-server path (US040 removes the `e2e-seed` one). Integration tests need a real/throwaway Postgres (tester already flagged this).

### D-002 — Embed migrations from a package whose directory can see `migrations/`
- **Context:** `//go:embed` can only embed files at or below the embedding `.go` file's directory. The migration files live at `services/agent-board/migrations/`.
- **Decision:** Place the embedding source file at the **module root level relative to migrations** — i.e. the `internal/migrate` package embeds via a small indirection: put a `migrations.go` containing `//go:embed migrations/*.up.sql` **at `services/agent-board/migrations.go`** (package `migrate` re-exported, or an `embed.FS` var in the root package consumed by `internal/migrate`). Concretely, the recommended layout: a file `services/agent-board/internal/migrate/embed.go` with the migrations copied under `internal/migrate/files/` is rejected (duplicates the dir); instead embed from a top-level file. **tech-lead/be-dev choose between:** (a) an `embed.go` at `services/agent-board/` (package `main`-adjacent isn't allowed across packages, so use a dedicated package at root, e.g. `package agentboard`), or (b) move the embed declaration into a file colocated such that `//go:embed migrations/*.up.sql` resolves. The architectural constraint is fixed: **the embedded FS must contain exactly the `*.up.sql` files, sorted lexically, with no `.down.sql` files included.** The exact file placement is an implementation detail for be-dev as long as it satisfies the embed-path rule and excludes down-migrations.
- **Alternatives rejected:** Copying the SQL into the package dir (two copies drift); reading from disk via `filepath.Glob` at runtime (fails in distroless — the files aren't in the image unless explicitly COPYed; embedding is cleaner and guarantees presence).
- **Consequences:** The migrations directory location is now load-bearing for the embed path; moving it requires updating the `//go:embed` directive. Down-migration files must be explicitly excluded from the glob (`*.up.sql`, not `*.sql`).

### D-003 — Idempotency via `schema_migrations` table, `.up.sql` files left untouched
- **Context:** Restart-safety requires detecting already-applied migrations; existing SQL uses bare `CREATE TABLE`.
- **Decision:** Track applied filenames in `schema_migrations`; skip any already recorded. Do not modify existing `.up.sql` files.
- **Alternatives rejected:** Rewriting all DDL to `IF NOT EXISTS` (touches working migrations, weaker generality for future data migrations).
- **Consequences:** First-ever boot against a DB that already has tables (e.g. one previously migrated by the old `e2e-seed` psql step) would attempt to re-run `000001` because `schema_migrations` is empty → "table already exists" error. **Mitigation / open question OQ-1:** for clean e2e/CI runs this never happens (fresh volume each time via `e2e-down -v`). For pre-existing dev DBs, the operator starts from a fresh DB or the runner backfills `schema_migrations` — flagged as an open question for the human.

### D-004 — Three separate read-only endpoints; tasks fetched separately from story detail
- **Context:** US042 needs a list with `taskCount` + full description; US043 needs single-story detail and the story's tasks.
- **Decision:** `GET /projects/{id}/user-stories` (wrapped list with `taskCount`), `GET /user-stories/{id}` (bare object, no tasks embedded), `GET /user-stories/{id}/tasks` (wrapped list). The drawer issues the detail + tasks calls in parallel.
- **Alternatives rejected:**
  - *Embed tasks in the detail response:* couples two concerns, breaks the bare-object convention for single resources, and forces a single larger payload even when only detail is needed. Two small parallel requests keep each endpoint single-purpose and let the FE show partial loading.
  - *Derive `taskCount` on the FE via N+1 per-story task calls:* N+1, slow, rejected — `taskCount` is a server aggregate on the list (single GROUP BY query).
- **Consequences:** The drawer makes two requests; FE must handle their loading/error independently (the spec already anticipates a single "Couldn't load this user story." drawer error — the FE may surface either failure under that message). The `taskCount`/no-`taskCount` asymmetry between list-item and detail must be honored exactly by types + MSW.

### D-005 — FE selection state lives in `UserStoriesTab`; drawer is conditional render, not a route
- **Context:** D-006 mandates a right-side drawer with the list staying mounted behind it; switching cards updates the drawer in place; Escape/X closes it.
- **Decision:** `UserStoriesTab` owns `selectedStoryId: string | null`. `UserStoryCardList` is always rendered; `UserStoryDrawer` renders only when `selectedStoryId !== null`. Selecting a different card sets a new id (drawer stays mounted, re-fetches via hook keyed on id). Close sets id to `null` (drawer unmounts; list never unmounted). No URL/router involvement (D-006 — in-tab state only, no deep-link).
- **Alternatives rejected:** URL `?story=` like DocumentsTab's `?doc=` (D-006 explicitly excludes deep-linking for v1); a dedicated route (excluded by D-006).
- **Consequences:** Selection is lost on full page reload (acceptable per D-006). Focus management: on open, focus moves into the drawer (heading or close button); on close, focus returns to the triggering card (accessibility AC). Drawer uses role=dialog + Escape handler + focus management; a full focus-trap is "sensible focus management" per the AC, not strictly required.

## Cross-cutting
- **Config / env vars:** none added. `DATABASE_URL` (existing) for migrations; `NEXT_PUBLIC_API_BASE_URL` (existing) for FE; `FRONTEND_URL` (existing) for CORS.
- **Logging keys:** migration runner logs one line per applied migration (`-> applying migration <filename>`) and a clear fatal on failure (mirrors the existing `e2e-seed` echo style and the existing `log.Print`/`log.Fatal` patterns in `main.go`). New handlers use the existing `log.Printf("Failed to ...: %v", err)` pattern on 500s.
- **Metrics:** none (project has none today).
- **Error model:** shared envelope `{ "code": string, "message": string }`; codes used here: `NOT_FOUND`, `INTERNAL_ERROR` (matching existing handlers). FE `ApiError` maps `code`/`message` as today; FE shows generic friendly copy ("Couldn't load user stories." / "Couldn't load this user story.") regardless of code, per the stories.
- **Observability:** Robot artifacts (`output.xml`, `log.html`, `report.html`) uploaded on every CI run (D-004) are the e2e observability surface.
- **CORS:** inherited — new routes share the api-server Echo instance and its existing CORS middleware; no change.

## Risks & open questions
- **OQ-1 (human) — pre-existing dev DB + bookkeeping backfill (D-003 consequence):** If an operator points the new migrations-at-startup api-server at a DB that already has the application tables but no `schema_migrations` row, boot will fail re-running `000001` ("table already exists"). For e2e/CI this never occurs (fresh volume). For local/dev DBs migrated by the old psql path, do we (a) require a fresh DB, (b) have the runner detect existing tables and backfill `schema_migrations`, or (c) accept the failure and document "drop and recreate"? Recommendation: (a) for this requirement (simplest, matches CI), document the reset. Please confirm.
- **OQ-2 (human) — CI toolchain path for Robot (US041):** install Python+Robot+psql on the runner and reuse `make e2e-seed`/`e2e-run` (keeps Makefile as single source of truth, my recommendation), OR run Robot from a container. Confirm the install-on-runner approach is acceptable, or state a preference.
- **Risk — `//go:embed` path placement (D-002):** if be-dev places the embed declaration wrong, the build fails fast (compile error) — low risk, caught immediately. Mitigation: the constraint (embed `*.up.sql` only, lexical order, exclude `.down.sql`) is fixed here; placement is the dev's choice.
- **Risk — e2e fixtures need user stories + tasks (US042/US043 e2e):** the current baseline seed (`REQ000_baseline.sql`) seeds a project + documents but **no user stories or tasks**. The new User Stories e2e suite will need seed rows. This is a tester concern (a new seed file under `tests/e2e/data/seeds/`), flagged here so the tester designs fixtures matching the contract. Not an architecture blocker.
- **Risk — branch protection not auto-enabled (US041):** the workflow alone does not block merges until a repo admin marks the check required. Documented as an external dependency; surface to the human at rollout.

## Approval log
### Revision 1 — 2026-06-07 — author: system-architect
- Initial draft covering US039–US043: migrations-at-startup (embed + `schema_migrations` table), Makefile mcp `--max-time` + data-only `e2e-seed`, GitHub Actions PR-to-main e2e gate, and three new read-only REST endpoints (exact JSON contracts) plus the FE card-list + side-drawer design. Decisions D-001–D-005 recorded; open questions OQ-1 (bookkeeping backfill) and OQ-2 (CI Robot toolchain) raised for the human.

### Revision 2 — 2026-06-08 — driver: human approval
- Approved by human at 2026-06-08T00:00:00Z.
