# US048 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live next to the code they exercise inside `services/agent-board/`.

## Coverage matrix

| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| List user-stories under requirement 200 | integration | IT-048-001 | services/agent-board / internal/handler | `GET /api/v1/projects/:pid/requirements/:rid/user-stories` |
| List user-stories — chain mismatch (req not in project) | integration | IT-048-002 | services/agent-board / internal/handler | same |
| List user-stories — project not found | integration | IT-048-003 | services/agent-board / internal/handler | same |
| List user-stories — empty list | integration | IT-048-004 | services/agent-board / internal/handler | same |
| List user-stories — 500 DB error | unit | UT-048-001 | services/agent-board / internal/handler | `UserStoryHandler.ListByRequirement` |
| Get user story 200 | integration | IT-048-005 | services/agent-board / internal/handler | `GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid` |
| Get user story — story not in requirement | integration | IT-048-006 | services/agent-board / internal/handler | same |
| Get user story — requirement not in project | integration | IT-048-007 | services/agent-board / internal/handler | same |
| Get user story — story not found | integration | IT-048-008 | services/agent-board / internal/handler | same |
| Get user story — 500 DB error | unit | UT-048-002 | services/agent-board / internal/handler | `UserStoryHandler.GetUserStory` (hierarchy) |
| List tasks for story 200 | integration | IT-048-009 | services/agent-board / internal/handler | `GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks` |
| List tasks — story not in requirement | integration | IT-048-010 | services/agent-board / internal/handler | same |
| List tasks — requirement not in project | integration | IT-048-011 | services/agent-board / internal/handler | same |
| List tasks — empty task list | integration | IT-048-012 | services/agent-board / internal/handler | same |
| List tasks — 500 DB error | unit | UT-048-003 | services/agent-board / internal/handler | task list (hierarchy) |
| Get task 200 | integration | IT-048-013 | services/agent-board / internal/handler | `GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid` |
| Get task — task not in story | integration | IT-048-014 | services/agent-board / internal/handler | same |
| Get task — story not in requirement | integration | IT-048-015 | services/agent-board / internal/handler | same |
| Get task — not found | integration | IT-048-016 | services/agent-board / internal/handler | same |
| Get task — 500 DB error | unit | UT-048-004 | services/agent-board / internal/handler | get task (hierarchy) |
| List documents under requirement 200 | integration | IT-048-017 | services/agent-board / internal/handler | `GET /api/v1/projects/:pid/requirements/:rid/documents` |
| List documents — chain mismatch | integration | IT-048-018 | services/agent-board / internal/handler | same |
| List documents — empty list | integration | IT-048-019 | services/agent-board / internal/handler | same |
| List documents — 500 DB error | unit | UT-048-005 | services/agent-board / internal/handler | document list (hierarchy) |
| Get document 200 (incl. content) | integration | IT-048-020 | services/agent-board / internal/handler | `GET /api/v1/projects/:pid/requirements/:rid/documents/:docid` |
| Get document — doc not in requirement | integration | IT-048-021 | services/agent-board / internal/handler | same |
| Get document — requirement not in project | integration | IT-048-022 | services/agent-board / internal/handler | same |
| Get document — not found | integration | IT-048-023 | services/agent-board / internal/handler | same |
| Get document — 500 DB error | unit | UT-048-006 | services/agent-board / internal/handler | get document (hierarchy) |
| Old flat route: /projects/:id/user-stories → 404 | integration | IT-048-024 | services/agent-board / internal/handler | removed route |
| Old flat route: /projects/:id/documents → 404 | integration | IT-048-025 | services/agent-board / internal/handler | removed route |
| Old flat route: /user-stories/:id → 404 | integration | IT-048-026 | services/agent-board / internal/handler | removed route |
| Old flat route: /user-stories/:id/tasks → 404 | integration | IT-048-027 | services/agent-board / internal/handler | removed route |
| Old flat route: /tasks/:id → 404 | integration | IT-048-028 | services/agent-board / internal/handler | removed route |
| Old flat route: /documents/:id → 404 | integration | IT-048-029 | services/agent-board / internal/handler | removed route |
| Old intermediate route: /requirements/:rid/user-stories → 404 | integration | IT-048-030 | services/agent-board / internal/handler | removed route |
| Old intermediate route: /requirements/:rid/documents → 404 | integration | IT-048-031 | services/agent-board / internal/handler | removed route |
| Ownership chain guard: wrong project for requirement | unit | UT-048-007 | services/agent-board / internal/handler | chain guard helper |
| Ownership chain guard: wrong requirement for story | unit | UT-048-008 | services/agent-board / internal/handler | chain guard helper |
| Ownership chain guard: wrong story for task | unit | UT-048-009 | services/agent-board / internal/handler | chain guard helper |
| Ownership chain guard: wrong requirement for document | unit | UT-048-010 | services/agent-board / internal/handler | chain guard helper |
| User-story list item includes requirementId | unit | UT-048-011 | services/agent-board / internal/handler | response mapping |
| User-story detail includes requirementId (no taskCount) | unit | UT-048-012 | services/agent-board / internal/handler | response mapping |
| Document list item includes requirementId | unit | UT-048-013 | services/agent-board / internal/handler | response mapping |
| Document detail includes requirementId + content | unit | UT-048-014 | services/agent-board / internal/handler | response mapping |
| Task item shape unchanged (no requirementId on task) | unit | UT-048-015 | services/agent-board / internal/handler | response mapping |
| Context cancelled propagated to repo query | unit | UT-048-016 | services/agent-board / internal/handler | context propagation |

---

## Unit tests

### UT-048-001 — UserStoryHandler.ListByRequirement: 500 on repo error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock `UserStoryRepository` whose `ListByRequirement` returns `errors.New("db error")`. Mock `RequirementRepository` whose `GetByID` returns a valid requirement with `project_id == :pid`.
- **When:** `GET /api/v1/projects/:pid/requirements/:rid/user-stories` served via `httptest`.
- **Then:** HTTP 500; body `{"code":"INTERNAL_ERROR","message":"Failed to fetch user stories"}`.
- **Architecture cite:** §6 500 response

### UT-048-002 — UserStoryHandler.GetUserStory (hierarchy): 500 on repo error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock `UserStoryRepository` whose `GetUserStory` returns `errors.New("db error")`. Chain guard repos return valid objects.
- **When:** `GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid` via httptest.
- **Then:** HTTP 500; body `{"code":"INTERNAL_ERROR","message":"Internal server error"}`.
- **Architecture cite:** §7 500 response

### UT-048-003 — Task list (hierarchy): 500 on repo error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock task repo `ListByUserStory` returns error. Chain guard repos return valid.
- **When:** `GET .../user-stories/:usid/tasks` via httptest.
- **Then:** HTTP 500; body `{"code":"INTERNAL_ERROR","message":"Failed to fetch tasks"}`.
- **Architecture cite:** §8 500 response

### UT-048-004 — Get task (hierarchy): 500 on repo error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock task repo `GetTask` returns error. Chain guard repos return valid.
- **When:** `GET .../tasks/:tid` via httptest.
- **Then:** HTTP 500; body `{"code":"INTERNAL_ERROR","message":"Internal server error"}`.
- **Architecture cite:** §9 500 response

### UT-048-005 — Document list (hierarchy): 500 on repo error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock document repo `ListByRequirement` returns error. Chain guard repos return valid.
- **When:** `GET .../requirements/:rid/documents` via httptest.
- **Then:** HTTP 500; body `{"code":"INTERNAL_ERROR","message":"Failed to fetch documents"}`.
- **Architecture cite:** §10 500 response

### UT-048-006 — Get document (hierarchy): 500 on repo error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock document repo `GetDocument` returns error. Chain guard repos return valid.
- **When:** `GET .../requirements/:rid/documents/:docid` via httptest.
- **Then:** HTTP 500; body `{"code":"INTERNAL_ERROR","message":"Failed to fetch document"}`.
- **Architecture cite:** §11 500 response

### UT-048-007 — Chain guard: requirement.ProjectID != :pid → 404
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock `RequirementRepository.GetByID` returns a requirement whose `ProjectID = "other-project-uuid"` (not the `:pid` path param).
- **When:** Any hierarchy endpoint with `:pid = "target-project"` and `:rid = "the-requirement"` is called.
- **Then:** HTTP 404; body `{"code":"NOT_FOUND","message":"Requirement not found"}`. The child resource is never fetched (mock asserts `GetUserStory` / `ListByRequirement` / etc. is not called).
- **Architecture cite:** §6 "verify requirement.project_id == :pid; any mismatch → 404"; D-009 "no cross-resource leakage"

### UT-048-008 — Chain guard: userStory.RequirementID != :rid → 404
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Chain guard for requirement passes (requirement belongs to project). Mock `UserStoryRepository.GetUserStory` returns a story whose `RequirementID = "other-requirement-uuid"`.
- **When:** `GET .../requirements/:rid/user-stories/:usid` called.
- **Then:** HTTP 404; body `{"code":"NOT_FOUND","message":"User story not found"}`.
- **Architecture cite:** §7 "verify userStory.requirement_id == :rid → 404"

### UT-048-009 — Chain guard: task.UserStoryID != :usid → 404
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** All parent chain guards pass. Mock `TaskRepository.GetTask` returns a task whose `UserStoryID = "other-story-uuid"`.
- **When:** `GET .../user-stories/:usid/tasks/:tid` called.
- **Then:** HTTP 404; body `{"code":"NOT_FOUND","message":"Task not found"}`.
- **Architecture cite:** §9 "verify task.user_story_id == :usid → 404"

### UT-048-010 — Chain guard: document.RequirementID != :rid → 404
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Chain guard for requirement passes. Mock `DocumentRepository.GetDocument` returns a document whose `RequirementID = "other-requirement-uuid"`.
- **When:** `GET .../requirements/:rid/documents/:docid` called.
- **Then:** HTTP 404; body `{"code":"NOT_FOUND","message":"Document not found"}`.
- **Architecture cite:** §11 "verify document.requirement_id == :rid → 404"

### UT-048-011 — User story list item response includes requirementId
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repos return a user story with `RequirementID = "b2e9d0c1-..."`. Chain guard passes.
- **When:** `GET .../requirements/:rid/user-stories` via httptest.
- **Then:** HTTP 200; body `{"userStories":[{"id":"...","requirementId":"b2e9d0c1-...","projectId":"...","title":"...","status":"...","taskCount":N,"createdAt":"...","updatedAt":"..."}]}`. The `requirementId` field is present and correct.
- **Architecture cite:** §6 user-story list item shape

### UT-048-012 — User story detail response includes requirementId but NOT taskCount
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repos return user story with `RequirementID`.
- **When:** `GET .../requirements/:rid/user-stories/:usid` via httptest.
- **Then:** HTTP 200; body includes `"requirementId"` field. Body does NOT include `"taskCount"` (the detail endpoint omits it per architecture §7).
- **Architecture cite:** §7 "bare user story detail object (no taskCount)"

### UT-048-013 — Document list item response includes requirementId
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repos return a document with `RequirementID`.
- **When:** `GET .../requirements/:rid/documents` via httptest.
- **Then:** HTTP 200; body `{"documents":[{"id":"...","requirementId":"...","projectId":"...","title":"...","createdAt":"...","updatedAt":"..."}]}`. No `content` field in list item.
- **Architecture cite:** §10 "metadata-only item shape (no content)"

### UT-048-014 — Document detail response includes requirementId AND content
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock document repo returns document with `RequirementID` and `Content = "# README\n..."`.
- **When:** `GET .../requirements/:rid/documents/:docid` via httptest.
- **Then:** HTTP 200; body includes `"requirementId"`, `"content": "# README\n..."`, and all other fields. No field is null.
- **Architecture cite:** §11 "full document object including content and requirementId"

### UT-048-015 — Task item shape is unchanged (no requirementId on task)
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repos return a task.
- **When:** `GET .../user-stories/:usid/tasks` via httptest.
- **Then:** HTTP 200; body `{"tasks":[{"id":"...","userStoryId":"...","title":"...","description":"...","status":"...","track":"...","createdAt":"...","updatedAt":"..."}]}`. No `requirementId` field on task items — the task shape is unchanged per architecture §8.
- **Architecture cite:** §8 "task item shape mirrors existing task list response exactly"

### UT-048-016 — Context cancellation propagated to all repo calls
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** A cancelled context injected via httptest request. Mock chain guard repo calls check `ctx.Err()`.
- **When:** Any hierarchy handler is called with a cancelled context.
- **Then:** Handler returns before calling deep repos; mock asserts that the first context-aware repo call propagates the cancelled context (does not call subsequent repos with a live context).

---

## Integration tests

> **Setup note:** IT-048-* use httptest with a real or in-memory Postgres (testcontainers-go) to test handler ↔ repo ↔ DB chains. Fixture setup inserts projects, requirements, user stories, tasks, and documents as needed.

### IT-048-001 — GET /api/v1/projects/:pid/requirements/:rid/user-stories — 200 with user stories
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ repo ↔ DB
- **Setup:** Project P, Requirement R (project_id=P), UserStory S1 (requirement_id=R), Task T1 (user_story_id=S1).
- **Endpoint:** `GET /api/v1/projects/{P}/requirements/{R}/user-stories`
- **Expect:**
  - HTTP 200
  - Body: `{"userStories":[{"id":"S1-uuid","projectId":"P-uuid","requirementId":"R-uuid","title":"...","description":"...","status":"...","taskCount":1,"createdAt":"...","updatedAt":"..."}]}`
  - `requirementId = R`; `taskCount = 1` (one task).
  - Ordered by `createdAt DESC` (existing convention).
- **Architecture cite:** §6 200 response

### IT-048-002 — GET .../requirements/:rid/user-stories — 404 when requirement does not belong to project
- **Service:** `services/agent-board`
- **Setup:** Project P1, Project P2, Requirement R (project_id=P2). UserStory under R.
- **Endpoint:** `GET /api/v1/projects/{P1}/requirements/{R}/user-stories`
- **Expect:**
  - HTTP 404
  - Body: `{"code":"NOT_FOUND","message":"Requirement not found"}`
- **Architecture cite:** §6 "requirement not found or belongs to different project → 404, indistinguishable"

### IT-048-003 — GET .../requirements/:rid/user-stories — 404 when project does not exist
- **Setup:** No project with the given UUID.
- **Endpoint:** `GET /api/v1/projects/00000000-0000-0000-0000-000000000000/requirements/{R}/user-stories`
- **Expect:** HTTP 404; not-found envelope.
- **Architecture cite:** §6 chain validation

### IT-048-004 — GET .../requirements/:rid/user-stories — 200 empty list
- **Setup:** Project P, Requirement R (project_id=P), no user stories.
- **Expect:**
  - HTTP 200
  - Body: `{"userStories": []}`
- **Architecture cite:** §6 "Empty → { userStories: [] }"

### IT-048-005 — GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid — 200 user story detail
- **Service:** `services/agent-board`
- **Setup:** Project P, Requirement R (project_id=P), UserStory S (requirement_id=R).
- **Endpoint:** `GET /api/v1/projects/{P}/requirements/{R}/user-stories/{S}`
- **Expect:**
  - HTTP 200
  - Body: bare object `{"id":"S-uuid","projectId":"P-uuid","requirementId":"R-uuid","title":"...","description":"...","status":"...","createdAt":"...","updatedAt":"..."}` — no `taskCount`.
- **Architecture cite:** §7 200 response

### IT-048-006 — GET .../user-stories/:usid — 404 when story belongs to different requirement
- **Setup:** Project P, Requirement R1 (project_id=P), Requirement R2 (project_id=P), UserStory S (requirement_id=R2).
- **Endpoint:** `GET /api/v1/projects/{P}/requirements/{R1}/user-stories/{S}`
- **Expect:** HTTP 404; `{"code":"NOT_FOUND","message":"User story not found"}`.
- **Architecture cite:** §7 "userStory.requirement_id == :rid; any mismatch → 404"

### IT-048-007 — GET .../user-stories/:usid — 404 when requirement not in project (story mismatch also hidden)
- **Setup:** Project P1, Project P2, Requirement R (project_id=P2), UserStory S (requirement_id=R).
- **Endpoint:** `GET /api/v1/projects/{P1}/requirements/{R}/user-stories/{S}`
- **Expect:** HTTP 404; not-found envelope — cross-project leakage is hidden.
- **Architecture cite:** §7 D-009

### IT-048-008 — GET .../user-stories/:usid — 404 for non-existent user story
- **Setup:** Project P, Requirement R (project_id=P). No user story with the given UUID.
- **Endpoint:** `GET /api/v1/projects/{P}/requirements/{R}/user-stories/00000000-0000-0000-0000-000000000000`
- **Expect:** HTTP 404; not-found envelope.
- **Architecture cite:** §7 "not found → 404"

### IT-048-009 — GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks — 200 task list
- **Setup:** Full chain: P → R → S → T1, T2.
- **Endpoint:** `GET /api/v1/projects/{P}/requirements/{R}/user-stories/{S}/tasks`
- **Expect:**
  - HTTP 200
  - Body: `{"tasks":[{"id":"...","userStoryId":"S-uuid","title":"...","description":"...","status":"...","track":"...","createdAt":"...","updatedAt":"..."},...]}` — two tasks.
  - No `requirementId` on task items.
- **Architecture cite:** §8 200 response

### IT-048-010 — GET .../user-stories/:usid/tasks — 404 when story not in requirement
- **Setup:** P → R1, R2 → S (requirement_id=R2).
- **Endpoint:** `GET /api/v1/projects/{P}/requirements/{R1}/user-stories/{S}/tasks`
- **Expect:** HTTP 404; not-found envelope.
- **Architecture cite:** §8 chain validation

### IT-048-011 — GET .../user-stories/:usid/tasks — 404 when requirement not in project
- **Setup:** P1 → (none); P2 → R → S.
- **Endpoint:** `GET /api/v1/projects/{P1}/requirements/{R}/user-stories/{S}/tasks`
- **Expect:** HTTP 404.

### IT-048-012 — GET .../user-stories/:usid/tasks — 200 empty list
- **Setup:** P → R → S (no tasks).
- **Expect:**
  - HTTP 200; body `{"tasks": []}`.
- **Architecture cite:** §8 "Array never null"

### IT-048-013 — GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid — 200 task detail
- **Setup:** Full chain P → R → S → T.
- **Endpoint:** `GET /api/v1/projects/{P}/requirements/{R}/user-stories/{S}/tasks/{T}`
- **Expect:**
  - HTTP 200
  - Body: bare task object `{"id":"T-uuid","userStoryId":"S-uuid","title":"...","description":"...","status":"...","track":"...","createdAt":"...","updatedAt":"..."}`.
- **Architecture cite:** §9 200 response

### IT-048-014 — GET .../tasks/:tid — 404 when task not in story
- **Setup:** P → R → S1, S2; T (user_story_id=S2).
- **Endpoint:** `GET /api/v1/projects/{P}/requirements/{R}/user-stories/{S1}/tasks/{T}`
- **Expect:** HTTP 404; `{"code":"NOT_FOUND","message":"Task not found"}`.
- **Architecture cite:** §9 "task.user_story_id == :usid → 404"

### IT-048-015 — GET .../tasks/:tid — 404 when story not in requirement
- **Setup:** P → R1, R2; R2 → S → T.
- **Endpoint:** `GET /api/v1/projects/{P}/requirements/{R1}/user-stories/{S}/tasks/{T}`
- **Expect:** HTTP 404.

### IT-048-016 — GET .../tasks/:tid — 404 for non-existent task
- **Setup:** Valid P → R → S chain. No task with given UUID.
- **Expect:** HTTP 404; not-found envelope.

### IT-048-017 — GET /api/v1/projects/:pid/requirements/:rid/documents — 200 document list
- **Setup:** P → R → D1, D2.
- **Endpoint:** `GET /api/v1/projects/{P}/requirements/{R}/documents`
- **Expect:**
  - HTTP 200
  - Body: `{"documents":[{"id":"D1-uuid","projectId":"P-uuid","requirementId":"R-uuid","title":"...","createdAt":"...","updatedAt":"..."},...]}` — no `content` field on list items.
- **Architecture cite:** §10 200 response; "metadata-only (no content)"

### IT-048-018 — GET .../requirements/:rid/documents — 404 when requirement not in project
- **Setup:** P1 → (none); P2 → R → D.
- **Endpoint:** `GET /api/v1/projects/{P1}/requirements/{R}/documents`
- **Expect:** HTTP 404; `{"code":"NOT_FOUND","message":"Requirement not found"}`.
- **Architecture cite:** §10 "verify requirement.project_id == :pid; mismatch → 404"

### IT-048-019 — GET .../requirements/:rid/documents — 200 empty list
- **Setup:** P → R (no documents).
- **Expect:** HTTP 200; body `{"documents": []}`.
- **Architecture cite:** §10 "Empty → { documents: [] }"

### IT-048-020 — GET /api/v1/projects/:pid/requirements/:rid/documents/:docid — 200 with content
- **Setup:** P → R → D (content = "# README\n...").
- **Endpoint:** `GET /api/v1/projects/{P}/requirements/{R}/documents/{D}`
- **Expect:**
  - HTTP 200
  - Body: `{"id":"D-uuid","projectId":"P-uuid","requirementId":"R-uuid","title":"...","content":"# README\n...","createdAt":"...","updatedAt":"..."}`.
  - `content` field present and non-empty.
  - `requirementId` present and equals R.
- **Architecture cite:** §11 200 response "incl. content and requirementId"

### IT-048-021 — GET .../documents/:docid — 404 when document not in requirement
- **Setup:** P → R1, R2 → D (requirement_id=R2).
- **Endpoint:** `GET /api/v1/projects/{P}/requirements/{R1}/documents/{D}`
- **Expect:** HTTP 404; `{"code":"NOT_FOUND","message":"Document not found"}`.
- **Architecture cite:** §11 "document.requirement_id == :rid; mismatch → 404"

### IT-048-022 — GET .../documents/:docid — 404 when requirement not in project
- **Setup:** P1 → (none); P2 → R → D.
- **Endpoint:** `GET /api/v1/projects/{P1}/requirements/{R}/documents/{D}`
- **Expect:** HTTP 404.

### IT-048-023 — GET .../documents/:docid — 404 for non-existent document
- **Setup:** Valid P → R chain. No document with given UUID.
- **Expect:** HTTP 404; not-found envelope.

### IT-048-024 — Removed route: GET /api/v1/projects/:id/user-stories → 404/405
- **Service:** `services/agent-board`
- **Boundary:** router (no handler registered)
- **Endpoint:** `GET /api/v1/projects/{anyUUID}/user-stories`
- **Expect:** HTTP 404 or 405 (router default unmatched-route response — NOT 200; the old handler does NOT execute).
- **Note:** If the router returns 405 when a different method matches the same path pattern, that is also acceptable — the important assertion is that no 200 response with user story data is returned.
- **Architecture cite:** Breaking changes — "GET /api/v1/projects/:id/user-stories removed"

### IT-048-025 — Removed route: GET /api/v1/projects/:id/documents → 404/405
- **Endpoint:** `GET /api/v1/projects/{anyUUID}/documents`
- **Expect:** HTTP 404 or 405.
- **Architecture cite:** Breaking changes

### IT-048-026 — Removed route: GET /api/v1/user-stories/:id → 404/405
- **Endpoint:** `GET /api/v1/user-stories/{anyUUID}`
- **Expect:** HTTP 404 or 405.
- **Architecture cite:** Breaking changes

### IT-048-027 — Removed route: GET /api/v1/user-stories/:id/tasks → 404/405
- **Endpoint:** `GET /api/v1/user-stories/{anyUUID}/tasks`
- **Expect:** HTTP 404 or 405.
- **Architecture cite:** Breaking changes

### IT-048-028 — Removed route: GET /api/v1/tasks/:id → 404/405
- **Endpoint:** `GET /api/v1/tasks/{anyUUID}`
- **Expect:** HTTP 404 or 405.
- **Architecture cite:** Breaking changes

### IT-048-029 — Removed route: GET /api/v1/documents/:id → 404/405
- **Endpoint:** `GET /api/v1/documents/{anyUUID}`
- **Expect:** HTTP 404 or 405.
- **Architecture cite:** Breaking changes

### IT-048-030 — Removed intermediate route: GET /api/v1/requirements/:rid/user-stories → 404/405
- **Endpoint:** `GET /api/v1/requirements/{anyUUID}/user-stories`
- **Expect:** HTTP 404 or 405.
- **Architecture cite:** Breaking changes — "intermediate draft routes removed"

### IT-048-031 — Removed intermediate route: GET /api/v1/requirements/:rid/documents → 404/405
- **Endpoint:** `GET /api/v1/requirements/{anyUUID}/documents`
- **Expect:** HTTP 404 or 405.
- **Architecture cite:** Breaking changes
