# US048 — Migrate to fully-nested REST hierarchy (breaking change): remove flat routes, replace with Project → Requirement → UserStory → Task and Project → Requirement → Document

**Requirement:** REQ008 — Requirement entity + project local-path linking
**Status:** draft

**Track:** BE

## Story
As an API consumer (MCP tool, FE client, or e2e test), I want every nested resource to be reachable only through its full parent chain (`Project → Requirement → UserStory → Task` and `Project → Requirement → Document`), so that the API exposes one consistent, unambiguous REST hierarchy instead of a mix of flat, project-scoped, and shorthand routes.

## Context / why
The current API is internally inconsistent: list endpoints are scoped under `projects` or `requirements`, while detail endpoints are top-level/shorthand (`/user-stories/:id`, `/tasks/:id`, `/documents/:id`). The REQ008 architecture review confirmed a full migration to a single fully-nested hierarchy.

This is a **breaking change**. All old flat and shorthand routes are **removed** — there is no backward compatibility, no deprecation window, and the new nested paths do **not** run "alongside" the old ones. Callers (FE, MCP tools, e2e) must move to the new paths; migrating those callers is tracked separately and is out of scope here.

## Confirmed decisions (final — not negotiable in this story)
- **Full REST hierarchy migration. Breaking change. No backward compat.**
- The two canonical hierarchies are:
  - `Project → Requirement → UserStory → Task`
  - `Project → Requirement → Document`
- **Remove** all 8 old routes from the router (see "Routes removed").
- **Add** the 6 new fully-nested routes (see "Routes added").
- Response bodies of the new endpoints are **identical** to what the corresponding old endpoint returned for the same resource — same bare-object shape, same fields, same ISO-8601 timestamp format — **plus** a `requirementId` field on user-story, task, and document objects where that field exists on the entity post-US044.
- **Full-chain ownership validation** on every new endpoint: project exists → requirement belongs to that project → resource belongs to that requirement (and, for tasks, the task belongs to that user story). Any break in the chain → **404** with the existing not-found envelope. "Wrong owner anywhere in the chain" and "resource missing" are indistinguishable to the caller — no cross-resource information leakage.

## Routes removed (in scope for US048 to delete)
```
GET /api/v1/projects/:id/user-stories
GET /api/v1/projects/:id/documents
GET /api/v1/requirements/:rid/user-stories
GET /api/v1/requirements/:rid/documents
GET /api/v1/user-stories/:id
GET /api/v1/user-stories/:id/tasks
GET /api/v1/tasks/:id
GET /api/v1/documents/:id
```

## Routes added (canonical — in scope for US048 to implement)
```
GET /api/v1/projects/:pid/requirements/:rid/user-stories
GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid
GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks
GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid
GET /api/v1/projects/:pid/requirements/:rid/documents
GET /api/v1/projects/:pid/requirements/:rid/documents/:docid
```

## Acceptance criteria

### Happy paths (new canonical routes)

- **Scenario: List user stories under a requirement**
  - Given project `P` exists, requirement `R` belongs to `P`, and user stories `S1, S2` belong to `R`
  - When a client calls `GET /api/v1/projects/{P}/requirements/{R}/user-stories`
  - Then the response is `200` with the same list shape the old `GET /api/v1/requirements/{R}/user-stories` returned, each item carrying its existing fields plus `requirementId == R`, scoped to user stories of `R` only.

- **Scenario: Get a single user story**
  - Given project `P`, requirement `R` (in `P`), and user story `S` (in `R`)
  - When a client calls `GET /api/v1/projects/{P}/requirements/{R}/user-stories/{S}`
  - Then the response is `200` with the same bare-object body the old `GET /api/v1/user-stories/{S}` returned (all existing fields, same ISO-8601 timestamps) plus `requirementId == R`.

- **Scenario: List tasks under a user story**
  - Given project `P`, requirement `R` (in `P`), user story `S` (in `R`), and tasks `T1, T2` (in `S`)
  - When a client calls `GET /api/v1/projects/{P}/requirements/{R}/user-stories/{S}/tasks`
  - Then the response is `200` with the same list shape the old `GET /api/v1/user-stories/{S}/tasks` returned, scoped to tasks of `S` only.

- **Scenario: Get a single task**
  - Given project `P`, requirement `R` (in `P`), user story `S` (in `R`), and task `T` (in `S`)
  - When a client calls `GET /api/v1/projects/{P}/requirements/{R}/user-stories/{S}/tasks/{T}`
  - Then the response is `200` with the same bare-object body the old `GET /api/v1/tasks/{T}` returned.

- **Scenario: List documents under a requirement**
  - Given project `P`, requirement `R` (in `P`), and documents `D1, D2` (in `R`)
  - When a client calls `GET /api/v1/projects/{P}/requirements/{R}/documents`
  - Then the response is `200` with the same list shape the old `GET /api/v1/requirements/{R}/documents` returned, each item carrying its existing fields plus `requirementId == R`, scoped to documents of `R` only.

- **Scenario: Get a single document**
  - Given project `P`, requirement `R` (in `P`), and document `D` (in `R`)
  - When a client calls `GET /api/v1/projects/{P}/requirements/{R}/documents/{D}`
  - Then the response is `200` with the same bare-object body the old `GET /api/v1/documents/{D}` returned (all existing fields, same ISO-8601 timestamps) plus `requirementId == R`.

### Ownership-mismatch paths (404, no info leakage)

- **Scenario: Project does not exist → 404**
  - Given no project with id `P` exists
  - When a client calls any new endpoint with `:pid = P`
  - Then the response is `404` with the existing not-found envelope; no resource is returned.

- **Scenario: Requirement does not belong to the project → 404**
  - Given project `P` exists and requirement `R` exists but belongs to a different project `P2 != P`
  - When a client calls `GET /api/v1/projects/{P}/requirements/{R}/...` (any sub-resource list or detail)
  - Then the response is `404` with the existing not-found envelope; the requirement and its children are not leaked.

- **Scenario: User story does not belong to the requirement → 404**
  - Given project `P`, requirement `R` (in `P`), and user story `S` that belongs to a different requirement `R2 != R`
  - When a client calls `GET /api/v1/projects/{P}/requirements/{R}/user-stories/{S}` (or its `/tasks` sub-paths)
  - Then the response is `404`; `S` is not returned.

- **Scenario: Task does not belong to the user story → 404**
  - Given project `P`, requirement `R` (in `P`), user story `S` (in `R`), and task `T` that belongs to a different user story `S2 != S`
  - When a client calls `GET /api/v1/projects/{P}/requirements/{R}/user-stories/{S}/tasks/{T}`
  - Then the response is `404`; `T` is not returned.

- **Scenario: Document does not belong to the requirement → 404**
  - Given project `P`, requirement `R` (in `P`), and document `D` that belongs to a different requirement `R2 != R`
  - When a client calls `GET /api/v1/projects/{P}/requirements/{R}/documents/{D}`
  - Then the response is `404`; `D` is not returned.

- **Scenario: Resource id does not exist → 404**
  - Given a valid project `P` and requirement `R` (in `P`) but no user story / task / document with the requested id
  - When a client calls the corresponding new detail endpoint
  - Then the response is `404` with the existing not-found envelope.

### Old routes removed

- **Scenario: All 8 old routes are gone from the router**
  - Given the API server is running with US048 applied
  - When a client calls any of the 8 removed routes (`GET /api/v1/projects/:id/user-stories`, `/projects/:id/documents`, `/requirements/:rid/user-stories`, `/requirements/:rid/documents`, `/user-stories/:id`, `/user-stories/:id/tasks`, `/tasks/:id`, `/documents/:id`)
  - Then the route is not registered (router returns its standard unmatched-route response — `404`/`405` per the router's default) and no old handler executes. None of these paths resolve to the migrated handlers.

## UI / UX flow expectations
No UI: backend-only router migration. No FE component, page, or hook changes are in scope. (FE/MCP callers must migrate to the new nested paths — tracked separately, out of scope here.)

## Out of scope
- Migrating FE or MCP callers to the new nested routes (separate story).
- Nested write verbs (POST/PUT/DELETE) — this story migrates the `GET` read surface only.
- Any deprecation shim, alias, or redirect from old paths to new ones — old paths are removed outright.
- Changes to the response envelope/error formats beyond adding `requirementId` to the affected entities.

## Dependencies
- **Blocked by US044** — the `requirement_id` migration must be complete first. The new hierarchy requires every user story, task, and document to resolve through a requirement, and the new response bodies include `requirementId`; both depend on US044's schema/data migration being in place.

## Notes for the team
- **Implementation shape (BE):** in the router (`cmd/api-server/main.go` or equivalent), delete the 8 old route registrations and register the 6 new fully-nested routes. Handlers reuse the existing fetch + response-mapping logic; the new behavior is the full-chain ownership guard.
- **Ownership guard:** validate top-down — load project by `:pid` (404 if missing); load requirement by `:rid` and assert `requirement.projectId == :pid` (404 on mismatch); load the target resource and assert it belongs to `:rid` (and for tasks, to `:usid`) (404 on mismatch). A single not-found envelope covers every failure point so "wrong owner" and "missing" are indistinguishable — no cross-resource leakage.
- **Error envelopes / 500 path:** match the existing handlers exactly. Do not introduce new error envelopes.
- **For the tester:** keep ownership-mismatch permutations at the unit/integration (handler-via-`httptest`) level — there are many and they are not user-observable at the routing layer. A small set of e2e cases is sufficient: one happy-path 200 per hierarchy leaf (user-story detail, task detail, document detail) and one representative 404 ownership-mismatch e2e to prove the chain is wired. Assert that the 8 removed routes no longer resolve (e.g. status is not 200 / route unregistered). Assert the new response bodies match the pre-migration shapes plus `requirementId`.

## Sign-off log
(po-ba appends here on each sign-off pass)
