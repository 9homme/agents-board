# US045 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live next to the code they exercise inside `services/agent-board/`.

## Coverage matrix

| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| List requirements 200 | integration | IT-045-001 | services/agent-board / internal/handler | `GET /api/v1/projects/:pid/requirements` |
| List requirements empty | integration | IT-045-002 | services/agent-board / internal/handler | `GET /api/v1/projects/:pid/requirements` |
| List requirements 404 unknown project | integration | IT-045-003 | services/agent-board / internal/handler | `GET /api/v1/projects/:pid/requirements` |
| List requirements 500 DB error | unit | UT-045-001 | services/agent-board / internal/handler | `RequirementHandler.ListRequirements` |
| No HTTP POST create requirement | integration | IT-045-004 | services/agent-board / internal/handler | `POST /api/v1/projects/:pid/requirements` |
| RequirementRepo.ListByProject happy path | unit | UT-045-002 | services/agent-board / internal/repo | `RequirementRepository.ListByProject` |
| RequirementRepo.ListByProject — DB Query error | unit | UT-045-003 | services/agent-board / internal/repo | `RequirementRepository.ListByProject` |
| RequirementRepo.ListByProject — rows.Scan error | unit | UT-045-004 | services/agent-board / internal/repo | `RequirementRepository.ListByProject` |
| RequirementRepo.ListByProject — rows.Err error | unit | UT-045-005 | services/agent-board / internal/repo | `RequirementRepository.ListByProject` |
| RequirementRepo.ListByProject — project not found | unit | UT-045-006 | services/agent-board / internal/repo | `RequirementRepository.ListByProject` |
| RequirementRepo.Create happy path | unit | UT-045-007 | services/agent-board / internal/repo | `RequirementRepository.Create` |
| RequirementRepo.Create — QueryRow Scan error | unit | UT-045-008 | services/agent-board / internal/repo | `RequirementRepository.Create` |
| RequirementRepo.Create — project FK violation | unit | UT-045-009 | services/agent-board / internal/repo | `RequirementRepository.Create` |
| RequirementRepo.Update happy path | unit | UT-045-010 | services/agent-board / internal/repo | `RequirementRepository.Update` |
| RequirementRepo.Update — not found | unit | UT-045-011 | services/agent-board / internal/repo | `RequirementRepository.Update` |
| RequirementRepo.Update — QueryRow Scan error | unit | UT-045-012 | services/agent-board / internal/repo | `RequirementRepository.Update` |
| POST /api/v1/projects — 201 valid path | integration | IT-045-005 | services/agent-board / internal/handler | `ProjectHandler.CreateProject` |
| POST /api/v1/projects — 400 missing path | integration | IT-045-006 | services/agent-board / internal/handler | `ProjectHandler.CreateProject` |
| POST /api/v1/projects — 400 blank path | integration | IT-045-007 | services/agent-board / internal/handler | `ProjectHandler.CreateProject` |
| POST /api/v1/projects — 400 missing name | integration | IT-045-008 | services/agent-board / internal/handler | `ProjectHandler.CreateProject` |
| POST /api/v1/projects — 400 blank name | integration | IT-045-009 | services/agent-board / internal/handler | `ProjectHandler.CreateProject` |
| POST /api/v1/projects — 400 path not a directory (file) | integration | IT-045-010 | services/agent-board / internal/handler | `ProjectHandler.CreateProject` |
| POST /api/v1/projects — 400 path not on disk | integration | IT-045-011 | services/agent-board / internal/handler | `ProjectHandler.CreateProject` |
| POST /api/v1/projects — 409 duplicate path | integration | IT-045-012 | services/agent-board / internal/handler | `ProjectHandler.CreateProject` |
| POST /api/v1/projects — 500 DB error | unit | UT-045-013 | services/agent-board / internal/handler | `ProjectHandler.CreateProject` |
| ProjectRepo.CreateProject — ErrInvalidPath from fsutil | unit | UT-045-014 | services/agent-board / internal/repo | `ProjectRepository.CreateProject` |
| ProjectRepo.CreateProject — ErrDuplicatePath (23505) | unit | UT-045-015 | services/agent-board / internal/repo | `ProjectRepository.CreateProject` |
| ProjectRepo.CreateProject — QueryRow Scan error | unit | UT-045-016 | services/agent-board / internal/repo | `ProjectRepository.CreateProject` |
| fsutil.ValidatePath — exists + is dir | unit | UT-045-017 | services/agent-board / internal/fsutil | `ValidatePath` |
| fsutil.ValidatePath — does not exist | unit | UT-045-018 | services/agent-board / internal/fsutil | `ValidatePath` |
| fsutil.ValidatePath — exists but is a file | unit | UT-045-019 | services/agent-board / internal/fsutil | `ValidatePath` |
| fsutil.ValidatePath — empty path | unit | UT-045-020 | services/agent-board / internal/fsutil | `ValidatePath` |
| MCP create_requirement happy path | unit | UT-045-021 | services/agent-board / internal/handler | `RegisterRequirementTools` → `create_requirement` |
| MCP create_requirement — blank name | unit | UT-045-022 | services/agent-board / internal/handler | `create_requirement` |
| MCP create_requirement — project not found | unit | UT-045-023 | services/agent-board / internal/handler | `create_requirement` |
| MCP create_requirement — invalid status | unit | UT-045-024 | services/agent-board / internal/handler | `create_requirement` |
| MCP create_requirement — explicit status in_progress | unit | UT-045-025 | services/agent-board / internal/handler | `create_requirement` |
| MCP create_requirement — explicit status done | unit | UT-045-026 | services/agent-board / internal/handler | `create_requirement` |
| MCP create_requirement — repo error | unit | UT-045-027 | services/agent-board / internal/handler | `create_requirement` |
| MCP list_requirements happy path | unit | UT-045-028 | services/agent-board / internal/handler | `list_requirements` |
| MCP list_requirements — unknown project_id | unit | UT-045-029 | services/agent-board / internal/handler | `list_requirements` |
| MCP list_requirements — repo error | unit | UT-045-030 | services/agent-board / internal/handler | `list_requirements` |
| MCP update_requirement happy path (status change) | unit | UT-045-031 | services/agent-board / internal/handler | `update_requirement` |
| MCP update_requirement — name update | unit | UT-045-032 | services/agent-board / internal/handler | `update_requirement` |
| MCP update_requirement — description update | unit | UT-045-033 | services/agent-board / internal/handler | `update_requirement` |
| MCP update_requirement — invalid status value | unit | UT-045-034 | services/agent-board / internal/handler | `update_requirement` |
| MCP update_requirement — blank name provided | unit | UT-045-035 | services/agent-board / internal/handler | `update_requirement` |
| MCP update_requirement — requirement not found | unit | UT-045-036 | services/agent-board / internal/handler | `update_requirement` |
| MCP update_requirement — all-empty (no-op) | unit | UT-045-037 | services/agent-board / internal/handler | `update_requirement` |
| MCP update_requirement — repo error | unit | UT-045-038 | services/agent-board / internal/handler | `update_requirement` |
| MCP create_user_story with requirement_id | unit | UT-045-039 | services/agent-board / internal/handler | `create_user_story` (BREAKING change) |
| MCP create_user_story — missing requirement_id | unit | UT-045-040 | services/agent-board / internal/handler | `create_user_story` |
| MCP create_user_story — requirement not in project | unit | UT-045-041 | services/agent-board / internal/handler | `create_user_story` |
| MCP create_document with requirement_id | unit | UT-045-042 | services/agent-board / internal/handler | `create_document` (BREAKING change) |
| MCP create_document — missing requirement_id | unit | UT-045-043 | services/agent-board / internal/handler | `create_document` |
| MCP create_document — requirement not in project | unit | UT-045-044 | services/agent-board / internal/handler | `create_document` |
| MCP create_project now requires path | unit | UT-045-045 | services/agent-board / internal/handler | `create_project` MCP tool (D-008) |
| MCP create_project — missing path | unit | UT-045-046 | services/agent-board / internal/handler | `create_project` MCP tool |
| MCP create_project — invalid path (not a dir) | unit | UT-045-047 | services/agent-board / internal/handler | `create_project` MCP tool |
| MCP create_project — duplicate path | unit | UT-045-048 | services/agent-board / internal/handler | `create_project` MCP tool |
| GET /api/v1/projects — items now include path | integration | IT-045-013 | services/agent-board / internal/handler | `GET /api/v1/projects` |
| GET /api/v1/projects/:pid — item includes path | integration | IT-045-014 | services/agent-board / internal/handler | `GET /api/v1/projects/:pid` |
| context.Done propagation in ListByProject | unit | UT-045-049 | services/agent-board / internal/repo | `RequirementRepository.ListByProject` |

---

## Unit tests

### UT-045-001 — RequirementHandler returns 500 on repo error for list
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** A mock `RequirementRepository` whose `ListByProject` returns an `errors.New("db error")`; handler wired with it.
- **When:** `GET /api/v1/projects/:pid/requirements` is served via `httptest`.
- **Then:** HTTP 500; body `{"code":"INTERNAL_ERROR","message":"Failed to fetch requirements"}`.
- **Architecture cite:** §4 500 response

### UT-045-002 — RequirementRepository.ListByProject returns ordered list
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Given:** A `sqlmock` DB that returns two requirement rows ordered by `created_at ASC`.
- **When:** `ListByProject(ctx, projectID)` is called.
- **Then:** Returns a `[]domain.Requirement` with two items; `items[0].CreatedAt` ≤ `items[1].CreatedAt`.
- **Architecture cite:** §4 "ordered by createdAt ASC"

### UT-045-003 — RequirementRepository.ListByProject — DB Query error
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Given:** `sqlmock` configured so the SELECT query returns `driver.ErrBadConn`.
- **When:** `ListByProject(ctx, projectID)` is called.
- **Then:** Returns a non-nil error. Returns nil slice.

### UT-045-004 — RequirementRepository.ListByProject — rows.Scan error
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Given:** `sqlmock` returns a row but the column type is deliberately wrong (causes Scan to fail).
- **When:** `ListByProject(ctx, projectID)` is called.
- **Then:** Returns a non-nil error wrapping the scan failure.

### UT-045-005 — RequirementRepository.ListByProject — rows.Err() error
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Given:** `sqlmock` returns rows successfully but injects a `rows.Err()` after iteration.
- **When:** `ListByProject(ctx, projectID)` is called.
- **Then:** Returns a non-nil error from the rows.Err() path.

### UT-045-006 — RequirementRepository.ListByProject — project not found returns empty list (not error)
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Given:** `sqlmock` returns zero rows for the project.
- **When:** `ListByProject(ctx, unknownProjectID)` is called.
- **Then:** Returns `([]domain.Requirement{}, nil)` — an empty non-nil slice and no error. (The handler separately validates project existence before calling the repo; the repo returns empty for missing projects to keep it simple.)
- **Note:** The handler's 404 on unknown project is tested in IT-045-003.

### UT-045-007 — RequirementRepository.Create happy path
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Given:** `sqlmock` expects an INSERT with `project_id`, `name`, `description`, `status` and returns a full row including a generated `id`, `created_at`, `updated_at`.
- **When:** `Create(ctx, domain.Requirement{ProjectID: "...", Name: "Default", Status: "draft"})` is called.
- **Then:** Returns a `domain.Requirement` with all fields populated (non-empty `ID`, correct `ProjectID`, `Name`, `Status = "draft"`, non-zero timestamps). No error.
- **Architecture cite:** MCP `create_requirement` output shape §5

### UT-045-008 — RequirementRepository.Create — QueryRow.Scan error
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Given:** `sqlmock` returns a scan-incompatible row.
- **When:** `Create(ctx, ...)` is called.
- **Then:** Returns a non-nil error.

### UT-045-009 — RequirementRepository.Create — project FK violation
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Given:** `sqlmock` returns a Postgres FK violation error (error code `23503`) for the `project_id` FK.
- **When:** `Create(ctx, domain.Requirement{ProjectID: "nonexistent"})` is called.
- **Then:** Returns a sentinel error `repo.ErrProjectNotFound` (or equivalent) — distinct from generic DB errors.

### UT-045-010 — RequirementRepository.Update happy path
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Given:** `sqlmock` expects an UPDATE and returns the updated row with `status = "in_progress"`, bumped `updated_at`.
- **When:** `Update(ctx, requirementID, patch{Status: ptr("in_progress")})` is called.
- **Then:** Returns the updated `domain.Requirement` with `Status = "in_progress"` and non-zero `UpdatedAt`. No error.
- **Architecture cite:** MCP `update_requirement` output shape §5

### UT-045-011 — RequirementRepository.Update — not found (sql.ErrNoRows)
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Given:** `sqlmock` returns `sql.ErrNoRows` from the UPDATE's Scan.
- **When:** `Update(ctx, "nonexistent-id", patch{})` is called.
- **Then:** Returns `repo.ErrNotFound` (or equivalent sentinel).
- **Architecture cite:** MCP `update_requirement` — "requirement not found → tool error"

### UT-045-012 — RequirementRepository.Update — QueryRow Scan error (non-ErrNoRows)
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Given:** `sqlmock` returns a generic scan error.
- **When:** `Update(ctx, ...)` is called.
- **Then:** Returns a non-nil error that is NOT `repo.ErrNotFound`.

### UT-045-013 — ProjectHandler.CreateProject — 500 on repo DB error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock `ProjectRepository` whose `CreateProject` returns a generic `errors.New("db error")` (not `ErrInvalidPath`, not `ErrDuplicatePath`). `fsutil.ValidatePath` stubbed to return nil.
- **When:** Valid `POST /api/v1/projects` body sent via httptest.
- **Then:** HTTP 500; body `{"code":"INTERNAL_ERROR","message":"Failed to create project"}`.
- **Architecture cite:** §3 500 response

### UT-045-014 — ProjectRepository.CreateProject — ErrInvalidPath from fsutil validation
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo` (or handler, where fsutil is called)
- **Given:** `fsutil.ValidatePath` returns `ErrInvalidPath` (path does not exist / not a dir).
- **When:** `CreateProject(ctx, ...)` (or the handler) is called.
- **Then:** Returns (or propagates) `ErrInvalidPath` — no INSERT is attempted.
- **Architecture cite:** §3 400 "path does not exist or is not a directory"

### UT-045-015 — ProjectRepository.CreateProject — ErrDuplicatePath (Postgres 23505)
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Given:** `sqlmock` returns a unique-violation error (`pq.Error{Code: "23505"}`) on the INSERT.
- **When:** `CreateProject(ctx, domain.Project{Path: "/already/taken"})` is called.
- **Then:** Returns `repo.ErrDuplicatePath` sentinel (not a generic error).
- **Architecture cite:** §3 409 response — `"DUPLICATE_PATH"`

### UT-045-016 — ProjectRepository.CreateProject — QueryRow Scan error
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Given:** `sqlmock` returns a scan error (not a constraint violation).
- **When:** `CreateProject(ctx, ...)` is called.
- **Then:** Returns a non-nil generic error (not `ErrDuplicatePath`, not `ErrInvalidPath`).

### UT-045-017 — fsutil.ValidatePath — exists and is a directory (happy path)
- **Service:** `services/agent-board`
- **Package under test:** `internal/fsutil`
- **Given:** A real temporary directory created by `os.MkdirTemp`.
- **When:** `ValidatePath(ctx, tempDir)` is called.
- **Then:** Returns `nil` (no error).
- **Architecture cite:** §3 "must exist on disk (os.Stat) and be a directory (IsDir())"

### UT-045-018 — fsutil.ValidatePath — path does not exist
- **Service:** `services/agent-board`
- **Package under test:** `internal/fsutil`
- **Given:** A path string that does not exist on disk (e.g. `/tmp/does-not-exist-abc123`).
- **When:** `ValidatePath(ctx, path)` is called.
- **Then:** Returns `ErrInvalidPath` (or equivalent sentinel). No panic.

### UT-045-019 — fsutil.ValidatePath — path exists but is a regular file
- **Service:** `services/agent-board`
- **Package under test:** `internal/fsutil`
- **Given:** A real temporary file created by `os.CreateTemp`.
- **When:** `ValidatePath(ctx, filePath)` is called.
- **Then:** Returns `ErrInvalidPath` — `IsDir()` returns false for a file.

### UT-045-020 — fsutil.ValidatePath — empty path
- **Service:** `services/agent-board`
- **Package under test:** `internal/fsutil`
- **Given:** `path = ""`.
- **When:** `ValidatePath(ctx, "")` is called.
- **Then:** Returns `ErrInvalidPath` immediately (no syscall needed).

### UT-045-021 — MCP create_requirement happy path (status defaults to draft)
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler` (MCP tools)
- **Given:** Mock repo; a valid `project_id` UUID, `name = "REQ008 Requirement entity"`, no status supplied.
- **When:** `create_requirement` MCP tool handler is called with those args.
- **Then:** Repo `Create` is called with `Status = "draft"`; tool returns the created requirement object with all fields: `id`, `projectId`, `name`, `description = ""`, `status = "draft"`, `createdAt`, `updatedAt`.
- **Architecture cite:** §5 `create_requirement` output shape

### UT-045-022 — MCP create_requirement — blank name returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** `name = ""`  (or whitespace-only).
- **When:** `create_requirement` is called.
- **Then:** Returns a tool error; repo `Create` is NOT called.
- **Architecture cite:** §5 "blank name → tool error"

### UT-045-023 — MCP create_requirement — project not found returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repo `Create` returns `repo.ErrProjectNotFound` (FK violation mapped to sentinel).
- **When:** `create_requirement` called with a non-existent `project_id`.
- **Then:** Returns a tool error containing "project not found" (or equivalent). Persists nothing.
- **Architecture cite:** §5 "project not found → tool error"

### UT-045-024 — MCP create_requirement — invalid status returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** `status = "invalid_status"`.
- **When:** `create_requirement` is called.
- **Then:** Returns a tool error. Repo `Create` NOT called.
- **Architecture cite:** §5 "invalid status → tool error"

### UT-045-025 — MCP create_requirement — explicit status in_progress
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** `status = "in_progress"`.
- **When:** `create_requirement` is called.
- **Then:** Repo `Create` called with `Status = "in_progress"`. Tool returns requirement with `status = "in_progress"`.

### UT-045-026 — MCP create_requirement — explicit status done
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** `status = "done"`.
- **When:** `create_requirement` is called.
- **Then:** Repo `Create` called with `Status = "done"`. Tool returns requirement with `status = "done"`.

### UT-045-027 — MCP create_requirement — generic repo error returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repo `Create` returns a generic `errors.New("db error")`.
- **When:** `create_requirement` is called with valid inputs.
- **Then:** Returns a tool error. Not a panic.

### UT-045-028 — MCP list_requirements happy path
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repo `ListByProject` returns two requirements.
- **When:** `list_requirements` MCP tool called with valid `project_id`.
- **Then:** Returns `{"requirements": [...]}` with two items, each matching the requirement object shape (`id`, `projectId`, `name`, `description`, `status`, `createdAt`, `updatedAt`).
- **Architecture cite:** §5 list_requirements output — "same shape as GET /api/v1/projects/:pid/requirements"

### UT-045-029 — MCP list_requirements — unknown project_id returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repo returns `repo.ErrProjectNotFound` (or the handler validates project existence first).
- **When:** `list_requirements` called with a non-existent `project_id`.
- **Then:** Returns a tool error indicating project not found.
- **Architecture cite:** §5 "unknown project_id → tool error"

### UT-045-030 — MCP list_requirements — generic repo error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repo `ListByProject` returns `errors.New("db error")`.
- **When:** `list_requirements` called.
- **Then:** Returns a tool error.

### UT-045-031 — MCP update_requirement happy path (status change)
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repo `Update` returns the updated requirement with `status = "in_progress"` and bumped `updated_at`.
- **When:** `update_requirement` called with `requirement_id = "..."`, `status = "in_progress"`.
- **Then:** Returns the updated requirement object with all seven fields. `updatedAt` is later than `createdAt`.
- **Architecture cite:** §5 `update_requirement` output shape

### UT-045-032 — MCP update_requirement — name update only
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repo `Update` returns requirement with `name = "New Name"`.
- **When:** `update_requirement` called with only `name = "New Name"` (no status or description).
- **Then:** Returns requirement with `name = "New Name"`. Status unchanged.

### UT-045-033 — MCP update_requirement — description update only
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repo `Update` returns requirement with updated `description`.
- **When:** `update_requirement` called with only `description = "New desc"`.
- **Then:** Returns requirement with updated description. Name and status unchanged.

### UT-045-034 — MCP update_requirement — invalid status value returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** `status = "not_a_valid_status"`.
- **When:** `update_requirement` called.
- **Then:** Returns a tool error. Repo `Update` NOT called.
- **Architecture cite:** §5 "invalid status value → tool error"

### UT-045-035 — MCP update_requirement — blank name when provided returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** `name = ""` or `name = "   "` (trimmed = blank).
- **When:** `update_requirement` called.
- **Then:** Returns a tool error. Repo NOT called.
- **Architecture cite:** §5 "`name`, if provided, is trimmed and must be non-blank"

### UT-045-036 — MCP update_requirement — requirement not found returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repo `Update` returns `repo.ErrNotFound`.
- **When:** `update_requirement` called with a non-existent `requirement_id`.
- **Then:** Returns a tool error indicating not found.
- **Architecture cite:** §5 "requirement not found → tool error"

### UT-045-037 — MCP update_requirement — all-empty patch is a no-op returning current object
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** `update_requirement` called with only `requirement_id` (no other fields).
- **When:** Called.
- **Then:** Repo `Update` is called (or no call if the handler short-circuits) and returns the current requirement object unchanged. No error.
- **Architecture cite:** §5 "all-empty update is a no-op returning the current object"

### UT-045-038 — MCP update_requirement — generic repo error returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repo `Update` returns `errors.New("db error")`.
- **When:** `update_requirement` called with valid inputs.
- **Then:** Returns a tool error. No panic.

### UT-045-039 — MCP create_user_story now includes requirement_id in INSERT
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock `UserStoryRepository.CreateUserStory` expects a `UserStory` with non-empty `RequirementID`. A valid `requirement_id` belonging to `project_id` is supplied in the tool call.
- **When:** `create_user_story` called with `project_id`, `requirement_id`, `title`.
- **Then:** Repo is called with `RequirementID` set; tool returns the created story object including `requirementId`.
- **Architecture cite:** §12 BREAKING CHANGE — "Repo INSERT becomes INSERT INTO user_stories (project_id, requirement_id, title, description, status)"

### UT-045-040 — MCP create_user_story — missing requirement_id returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** `requirement_id` omitted from the tool call arguments.
- **When:** `create_user_story` called.
- **Then:** Returns a tool error indicating `requirement_id` is required. Repo NOT called.
- **Architecture cite:** §12 "missing requirement_id → tool error"

### UT-045-041 — MCP create_user_story — requirement does not belong to project returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** A `requirement_id` that exists in the DB but whose `project_id` ≠ the supplied `project_id`.
- **When:** `create_user_story` called.
- **Then:** Returns tool error "requirement does not belong to project". Repo NOT called.
- **Architecture cite:** §12 "requirement's project_id must equal the supplied project_id — mismatch → tool error"

### UT-045-042 — MCP create_document now includes requirement_id in INSERT
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock `DocumentRepository.CreateDocument` expects a `Document` with non-empty `RequirementID`. Valid `requirement_id` belonging to `project_id` supplied.
- **When:** `create_document` called with `project_id`, `requirement_id`, `title`.
- **Then:** Repo called with `RequirementID` set; tool returns document object including `requirementId`.
- **Architecture cite:** §13 BREAKING CHANGE

### UT-045-043 — MCP create_document — missing requirement_id returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** `requirement_id` omitted.
- **When:** `create_document` called.
- **Then:** Returns tool error indicating `requirement_id` required. Repo NOT called.
- **Architecture cite:** §13 "missing requirement_id → tool error"

### UT-045-044 — MCP create_document — requirement not in project returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** `requirement_id` belongs to a different project.
- **When:** `create_document` called.
- **Then:** Returns tool error. Repo NOT called.
- **Architecture cite:** §13 "mismatch → tool error"

### UT-045-045 — MCP create_project now requires path (D-008)
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock fsutil `ValidatePath` returns nil. Mock repo `CreateProject` returns a project with `Path` set. Valid `path` supplied in MCP args.
- **When:** `create_project` MCP tool called with `name`, `path`.
- **Then:** `ValidatePath` is called with the supplied path; repo `CreateProject` is called with non-empty `Path`; tool returns project object including `path`.
- **Architecture cite:** D-008 — "MCP create_project gains required, non-blank path"

### UT-045-046 — MCP create_project — missing path returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** `path` argument omitted from the tool call.
- **When:** `create_project` called.
- **Then:** Returns tool error. `ValidatePath` NOT called. Repo NOT called.
- **Architecture cite:** D-008

### UT-045-047 — MCP create_project — invalid path (not a directory) returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock `ValidatePath` returns `ErrInvalidPath`.
- **When:** `create_project` called with an invalid path.
- **Then:** Returns tool error. Repo NOT called.
- **Architecture cite:** D-008

### UT-045-048 — MCP create_project — duplicate path returns tool error
- **Service:** `services/agent-board`
- **Package under test:** `internal/handler`
- **Given:** Mock repo `CreateProject` returns `repo.ErrDuplicatePath`.
- **When:** `create_project` called with an already-used path.
- **Then:** Returns tool error indicating duplicate path.
- **Architecture cite:** D-008

### UT-045-049 — RequirementRepository.ListByProject — context cancellation
- **Service:** `services/agent-board`
- **Package under test:** `internal/repo`
- **Given:** A cancelled context (`ctx, cancel := context.WithCancel(context.Background()); cancel()`).
- **When:** `ListByProject(ctx, projectID)` is called.
- **Then:** Returns a non-nil error (context cancelled). Does not hang.

---

## Integration tests

### IT-045-001 — GET /api/v1/projects/:pid/requirements — 200 with requirements list
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ in-memory/testcontainer DB
- **Setup:** Insert project `P1`, insert two requirements under `P1` with different `created_at` values.
- **Endpoint:** `GET /api/v1/projects/{P1}/requirements`
- **Expect:**
  - HTTP 200
  - Body: `{"requirements": [<req1>, <req2>]}` ordered by `created_at` ASC
  - Each item: `id` (uuid string), `projectId` (== P1 id), `name` (string), `description` (string), `status` ("draft"|"in_progress"|"done"), `createdAt` (ISO-8601), `updatedAt` (ISO-8601)
  - No extra fields; no null fields.
- **Architecture cite:** §4 200 response shape

### IT-045-002 — GET /api/v1/projects/:pid/requirements — 200 empty list for new project
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ DB
- **Setup:** Insert project `P2` with no requirements.
- **Endpoint:** `GET /api/v1/projects/{P2}/requirements`
- **Expect:**
  - HTTP 200
  - Body: `{"requirements": []}` — key present, value empty array (NOT null).
- **Architecture cite:** §4 "Empty project → { requirements: [] }"

### IT-045-003 — GET /api/v1/projects/:pid/requirements — 404 unknown project
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ DB
- **Setup:** No project with the given UUID.
- **Endpoint:** `GET /api/v1/projects/00000000-0000-0000-0000-000000000000/requirements`
- **Expect:**
  - HTTP 404
  - Body: `{"code":"NOT_FOUND","message":"Project not found"}`
- **Architecture cite:** §4 404 response

### IT-045-004 — POST /api/v1/projects/:pid/requirements — 404 (no HTTP create endpoint)
- **Service:** `services/agent-board`
- **Boundary:** router
- **Setup:** Any running server.
- **Endpoint:** `POST /api/v1/projects/{anyUUID}/requirements` (with any body)
- **Expect:** HTTP 404 or 405 — the route is not registered. The router returns its default unmatched-route response. No requirement is created.
- **Architecture cite:** D-004 "No POST /api/v1/projects/:id/requirements HTTP endpoint"

### IT-045-005 — POST /api/v1/projects — 201 with valid name and valid path
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ fsutil ↔ DB
- **Setup:** A real temporary directory created by the test (`os.MkdirTemp`). DB has no project with that path.
- **Endpoint:** `POST /api/v1/projects`
- **Request body:** `{"name": "Test Project", "description": "", "path": "/tmp/testdir-xxx"}`
- **Expect:**
  - HTTP 201
  - Body matches: `{"id": "<uuid>", "name": "Test Project", "description": "", "path": "/tmp/testdir-xxx", "createdAt": "<ISO-8601>", "updatedAt": "<ISO-8601>"}`
  - `id` is a valid UUID string; `path` equals the submitted value; both timestamps are non-zero.
- **Teardown:** Remove temp dir; delete the inserted project row.
- **Architecture cite:** §3 201 response

### IT-045-006 — POST /api/v1/projects — 400 when path field is missing
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ fsutil
- **Request body:** `{"name": "No Path Project"}`
- **Expect:**
  - HTTP 400
  - Body: `{"code":"VALIDATION_ERROR","message":"path is required"}`
  - No row inserted in DB.
- **Architecture cite:** §3 400 — "path is required"

### IT-045-007 — POST /api/v1/projects — 400 when path is blank string
- **Service:** `services/agent-board`
- **Request body:** `{"name": "Blank Path", "path": ""}`
- **Expect:**
  - HTTP 400
  - Body: `{"code":"VALIDATION_ERROR","message":"path is required"}`
- **Architecture cite:** §3 400 — blank path treated same as absent

### IT-045-008 — POST /api/v1/projects — 400 when name is missing
- **Service:** `services/agent-board`
- **Request body:** `{"path": "/tmp/some-dir"}`
- **Expect:**
  - HTTP 400
  - Body: `{"code":"VALIDATION_ERROR","message":"name is required"}`
- **Architecture cite:** §3 400 — "name is required"

### IT-045-009 — POST /api/v1/projects — 400 when name is blank
- **Service:** `services/agent-board`
- **Request body:** `{"name": "   ", "path": "/tmp/some-dir"}`
- **Expect:**
  - HTTP 400
  - Body: `{"code":"VALIDATION_ERROR","message":"name is required"}`
- **Note:** Trimmed name must be non-blank.
- **Architecture cite:** §3 "name string required, non-blank (trimmed)"

### IT-045-010 — POST /api/v1/projects — 400 when path exists but is a regular file
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ fsutil (real os.Stat)
- **Setup:** Create a real temporary file (`os.CreateTemp`).
- **Request body:** `{"name": "File Path", "path": "<path-to-temp-file>"}`
- **Expect:**
  - HTTP 400
  - Body: `{"code":"VALIDATION_ERROR","message":"path does not exist or is not a directory"}`
  - No row inserted.
- **Teardown:** Remove temp file.
- **Architecture cite:** §3 400 — "path does not exist / is not a directory"

### IT-045-011 — POST /api/v1/projects — 400 when path does not exist on disk
- **Service:** `services/agent-board`
- **Request body:** `{"name": "Ghost Path", "path": "/tmp/this-does-not-exist-xxxxxxxxxxx"}`
- **Expect:**
  - HTTP 400
  - Body: `{"code":"VALIDATION_ERROR","message":"path does not exist or is not a directory"}`
- **Architecture cite:** §3 400

### IT-045-012 — POST /api/v1/projects — 409 when path already linked to another project
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ fsutil ↔ DB
- **Setup:**
  1. Create a real temp directory.
  2. Insert a project row with `path = <tempdir>` directly into the DB.
- **Request body:** `{"name": "Duplicate", "path": "<tempdir>"}`
- **Expect:**
  - HTTP 409
  - Body: `{"code":"DUPLICATE_PATH","message":"path already linked to another project"}`
  - No second row inserted.
- **Teardown:** Remove temp dir; delete the inserted project.
- **Architecture cite:** §3 409 — `"DUPLICATE_PATH"`

### IT-045-013 — GET /api/v1/projects — response items include `path` field
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ DB
- **Setup:** At least one project in DB with a non-empty `path`.
- **Endpoint:** `GET /api/v1/projects`
- **Expect:**
  - HTTP 200
  - Each item in `projects[]` has a `"path"` key with a non-null string value.
- **Architecture cite:** §1 modified — adds `path` field to each item

### IT-045-014 — GET /api/v1/projects/:pid — response includes `path` field
- **Service:** `services/agent-board`
- **Boundary:** handler ↔ DB
- **Setup:** Project with known `path` inserted.
- **Endpoint:** `GET /api/v1/projects/{pid}`
- **Expect:**
  - HTTP 200
  - Body has `"path": "<expected-path>"` (exact string match).
- **Architecture cite:** §2 modified — adds `path` field
