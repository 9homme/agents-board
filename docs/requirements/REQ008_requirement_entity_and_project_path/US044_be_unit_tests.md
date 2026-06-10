# US044 — Backend unit & integration test specification

**For BE Dev:** these are the tests you write FIRST (TDD red). Implement in Go using `testing` + `github.com/stretchr/testify`. Tests live next to the code they exercise inside `services/agent-board/`.

## Coverage matrix

| AC scenario | Layer | Test ID | Service / package | Function or endpoint under test |
|---|---|---|---|---|
| requirements table exists post-migration | integration | IT-044-001 | services/agent-board / migrate | `000003_requirement_entity.up.sql` |
| status defaults to draft | unit | UT-044-001 | services/agent-board / internal/domain | `Requirement` struct zero-value |
| status enum validated | unit | UT-044-002 | services/agent-board / internal/domain | `Requirement` status constant values |
| user_stories.requirement_id NOT NULL after migration | integration | IT-044-002 | services/agent-board / migrate | `000003_requirement_entity.up.sql` |
| documents.requirement_id NOT NULL after migration | integration | IT-044-003 | services/agent-board / migrate | `000003_requirement_entity.up.sql` |
| existing projects get Default requirement | integration | IT-044-004 | services/agent-board / migrate | backfill in `000003_requirement_entity.up.sql` |
| zero data loss — user stories re-parented | integration | IT-044-005 | services/agent-board / migrate | backfill UPDATE in `000003_requirement_entity.up.sql` |
| zero data loss — documents re-parented | integration | IT-044-006 | services/agent-board / migrate | backfill UPDATE in `000003_requirement_entity.up.sql` |
| projects.path NOT NULL + unique | integration | IT-044-007 | services/agent-board / migrate | `000003_requirement_entity.up.sql` |
| path uniqueness constraint prevents duplicate | integration | IT-044-008 | services/agent-board / migrate | `uq_projects_path` constraint |
| Requirement domain type compiles with all fields | unit | UT-044-003 | services/agent-board / internal/domain | `domain.Requirement` type |
| UserStory carries RequirementId | unit | UT-044-004 | services/agent-board / internal/domain | `domain.UserStory.RequirementID` |
| Document carries RequirementId | unit | UT-044-005 | services/agent-board / internal/domain | `domain.Document.RequirementID` |
| Project carries Path non-pointer string | unit | UT-044-006 | services/agent-board / internal/domain | `domain.Project.Path` |
| down migration reverses schema cleanly | integration | IT-044-009 | services/agent-board / migrate | `000003_requirement_entity.down.sql` |
| no orphaned rows after backfill | integration | IT-044-010 | services/agent-board / migrate | row-count invariant |

---

## Unit tests

### UT-044-001 — Requirement status constant: draft is the zero-value default
- **Service:** `services/agent-board`
- **Package under test:** `internal/domain`
- **Given:** A `Requirement` struct initialized with only required fields (`ID`, `ProjectID`, `Name`), no status set
- **When:** `Status` field is accessed
- **Then:** The `status` field (when set to the default in `NewRequirement` or equivalent) equals `"draft"`; the constant `RequirementStatusDraft` has value `"draft"`
- **Edge cases:** Verify constants: `RequirementStatusDraft = "draft"`, `RequirementStatusInProgress = "in_progress"`, `RequirementStatusDone = "done"`. No other values are defined.
- **Architecture cite:** §4 requirements item shape — `"status": "draft"|"in_progress"|"done"`

### UT-044-002 — Requirement status enum completeness
- **Service:** `services/agent-board`
- **Package under test:** `internal/domain`
- **Given:** The domain constant block for `Requirement` statuses
- **When:** Each constant is evaluated
- **Then:** Exactly three constants exist: `RequirementStatusDraft = "draft"`, `RequirementStatusInProgress = "in_progress"`, `RequirementStatusDone = "done"`. No `blocked_*` status on Requirement (out of scope this REQ — architecture Scope section).
- **Architecture cite:** Architecture scope: "no state-machine enforcement on Requirement.status"

### UT-044-003 — Requirement domain type has all required fields
- **Service:** `services/agent-board`
- **Package under test:** `internal/domain`
- **Given:** The `Requirement` struct definition
- **When:** The struct is compiled and its fields are inspected
- **Then:** The struct has exactly: `ID string`, `ProjectID string`, `Name string`, `Description string`, `Status string`, `CreatedAt time.Time`, `UpdatedAt time.Time`. JSON tags must match: `id`, `projectId`, `name`, `description`, `status`, `createdAt`, `updatedAt`.
- **Architecture cite:** §4 response item shape

### UT-044-004 — UserStory gains RequirementID field
- **Service:** `services/agent-board`
- **Package under test:** `internal/domain`
- **Given:** The `UserStory` struct definition after US044 changes
- **When:** The struct is compiled
- **Then:** `UserStory.RequirementID string` field exists with JSON tag `requirementId`. It is a plain non-pointer `string` (NOT NULL semantics at the Go level). The previously existing fields (`ID`, `ProjectID`, `Title`, `Description`, `Status`, `TaskCount`, `CreatedAt`, `UpdatedAt`) remain present and unchanged.
- **Architecture cite:** §6 user-story list item shape — `requirementId` always present

### UT-044-005 — Document gains RequirementID field
- **Service:** `services/agent-board`
- **Package under test:** `internal/domain`
- **Given:** The `Document` struct definition after US044 changes
- **When:** The struct is compiled
- **Then:** `Document.RequirementID string` field exists with JSON tag `requirementId`. Non-pointer `string`. Existing fields unchanged.
- **Architecture cite:** §10 document list item shape — `requirementId` always present

### UT-044-006 — Project gains Path field (non-pointer, required)
- **Service:** `services/agent-board`
- **Package under test:** `internal/domain`
- **Given:** The `Project` struct definition after US044 changes
- **When:** The struct is compiled
- **Then:** `Project.Path string` field exists with JSON tag `path`. It is a non-pointer `string` (never `null` in JSON responses). Existing fields (`ID`, `Name`, `Description`, `CreatedAt`, `UpdatedAt`) remain present and unchanged.
- **Architecture cite:** §1 project list item — `"path"` always present, §2 project get — `"path"` always present

---

## Integration tests

> **Setup note:** All IT-044-* tests require a real PostgreSQL instance (testcontainers-go or a Docker-in-Docker pattern matching the existing `*_test.go` files in `internal/repo/`). Each test runs the migration runner against a clean schema and inspects the result via direct SQL queries.

### IT-044-001 — requirements table exists with correct columns post-migration
- **Service:** `services/agent-board`
- **Boundary:** migration runner ↔ real Postgres (testcontainers)
- **Setup:** Start a fresh Postgres container; run all migrations up to and including `000003_requirement_entity.up.sql`.
- **When:** Query `information_schema.columns` for the `requirements` table.
- **Then:**
  - Table `requirements` exists.
  - Columns present: `id` (uuid / pg type), `project_id` (uuid, NOT NULL), `name` (varchar(255), NOT NULL), `description` (text, NOT NULL), `status` (varchar(50), NOT NULL), `created_at` (timestamptz, NOT NULL), `updated_at` (timestamptz, NOT NULL).
  - `id` has `DEFAULT gen_random_uuid()`.
  - `project_id` has a foreign key referencing `projects(id)` with `ON DELETE CASCADE`.
  - `status` has a CHECK constraint accepting `'draft'`, `'in_progress'`, `'done'` only.
  - Index `idx_requirements_project_id` exists on `requirements(project_id)`.
- **Teardown:** Drop the container.
- **Architecture cite:** Data model — `CREATE TABLE requirements ...`

### IT-044-002 — user_stories.requirement_id NOT NULL after migration
- **Service:** `services/agent-board`
- **Boundary:** migration runner ↔ real Postgres
- **Setup:** Same fresh migration as IT-044-001.
- **When:** Query `information_schema.columns` for `user_stories.requirement_id`.
- **Then:**
  - Column `requirement_id` exists with `is_nullable = 'NO'` (NOT NULL).
  - Column has FK referencing `requirements(id)` with `ON DELETE CASCADE`.
  - Index `idx_user_stories_requirement_id` exists.
- **Edge case:** Attempt to INSERT a user_story with `requirement_id = NULL`; expect a NOT NULL violation error from Postgres.
- **Architecture cite:** Data model step 3+5

### IT-044-003 — documents.requirement_id NOT NULL after migration
- **Service:** `services/agent-board`
- **Boundary:** migration runner ↔ real Postgres
- **Setup:** Same fresh migration.
- **When:** Query `information_schema.columns` for `documents.requirement_id`.
- **Then:**
  - Column `requirement_id` exists with `is_nullable = 'NO'`.
  - FK referencing `requirements(id)` ON DELETE CASCADE.
  - Index `idx_documents_requirement_id` exists.
- **Edge case:** Attempt to INSERT a document with `requirement_id = NULL`; expect NOT NULL violation.
- **Architecture cite:** Data model step 3+5

### IT-044-004 — existing projects get exactly one "Default" requirement per backfill
- **Service:** `services/agent-board`
- **Boundary:** migration runner ↔ real Postgres
- **Setup:**
  1. Run migrations 000001 + 000002 against a fresh Postgres.
  2. Insert 3 projects (P1, P2, P3) with no requirements yet.
  3. Insert 2 user stories under P1, 1 document under P2, nothing under P3.
  4. Run migration 000003.
- **When:** Query `SELECT COUNT(*) FROM requirements WHERE name = 'Default'`.
- **Then:**
  - Count = 3 (one per project, including P3 which has no children — see architecture note: "creates a Default requirement for every project including those with no children").
  - Each `Default` requirement has `status = 'draft'`.
  - Each `Default` requirement has `project_id` matching its project.
- **Architecture cite:** Data model backfill step 4

### IT-044-005 — zero data loss: user stories re-parented to Default requirement
- **Service:** `services/agent-board`
- **Boundary:** migration runner ↔ real Postgres
- **Setup:** Same as IT-044-004.
- **When:** After migration 000003, query `user_stories`.
- **Then:**
  - Count before migration = 2 (P1's stories); count after = 2 (no loss).
  - Both rows have `requirement_id = <P1's Default requirement id>` (non-null).
  - Both rows still have their original `project_id = P1`.
  - No orphaned user_stories (zero rows where `requirement_id IS NULL` — impossible after SET NOT NULL but verify).
- **Architecture cite:** Data model backfill step 4; "zero data loss is mandatory"

### IT-044-006 — zero data loss: documents re-parented to Default requirement
- **Service:** `services/agent-board`
- **Boundary:** migration runner ↔ real Postgres
- **Setup:** Same as IT-044-004.
- **When:** After migration 000003, query `documents`.
- **Then:**
  - Count before = 1 (P2's document); count after = 1 (no loss).
  - Row has `requirement_id = <P2's Default requirement id>` (non-null).
  - Row retains original `project_id = P2`.
- **Architecture cite:** Data model backfill step 4

### IT-044-007 — projects.path column is TEXT NOT NULL with unique constraint
- **Service:** `services/agent-board`
- **Boundary:** migration runner ↔ real Postgres
- **Setup:** Fresh migration including 000003.
- **When:** Query `information_schema.columns` for `projects.path` and `information_schema.table_constraints` for `uq_projects_path`.
- **Then:**
  - Column `path` exists on `projects` with data type `text`, `is_nullable = 'NO'`.
  - Unique constraint `uq_projects_path` exists on `projects(path)`.
- **Architecture cite:** Data model step 2

### IT-044-008 — path uniqueness constraint rejects duplicate paths
- **Service:** `services/agent-board`
- **Boundary:** migration runner ↔ real Postgres
- **Setup:** Fresh migration; insert one project with `path = '/tmp/test-project'`.
- **When:** Attempt to INSERT a second project with `path = '/tmp/test-project'`.
- **Then:** Postgres returns a unique-violation error (error code `23505`).
- **Edge case:** Same path with trailing slash vs without is treated as different by the constraint (string equality, not path normalization — the application layer normalizes if needed; the DB sees only the stored string).
- **Architecture cite:** `uq_projects_path`; D-006

### IT-044-009 — down migration reverses schema cleanly
- **Service:** `services/agent-board`
- **Boundary:** migration runner ↔ real Postgres
- **Setup:**
  1. Fresh Postgres; run all migrations through 000003 up.
  2. Insert a project with a `path`, a requirement, a user story, a document (all linked).
- **When:** Run `000003_requirement_entity.down.sql`.
- **Then:**
  - `requirements` table is gone (or the equivalent reversal as documented in the down file).
  - `user_stories.requirement_id` column is gone.
  - `documents.requirement_id` column is gone.
  - `projects.path` column is gone (or null-able, depending on down script) and `uq_projects_path` constraint is removed.
  - `user_stories` and `documents` rows (minus re-parenting columns) are still present — no data loss of core content.
  - No Postgres errors during the down execution.
- **Architecture cite:** "Migration is reversible" AC; Data model down-path note
- **Coverage exemption:** The down migration is documentation-only per architecture ("runner only executes `*.up.sql`"). The down migration is tested here for correctness guarantees only; it is not run in production automation. Test still runs manually to confirm the SQL is syntactically and semantically valid.

### IT-044-010 — no orphaned rows after backfill (row-count invariant)
- **Service:** `services/agent-board`
- **Boundary:** migration runner ↔ real Postgres
- **Setup:** Same as IT-044-004 (3 projects, 2 stories, 1 document pre-migration).
- **When:** After migration 000003 completes.
- **Then:**
  - `SELECT COUNT(*) FROM user_stories WHERE requirement_id IS NULL` = 0 (would be rejected by NOT NULL anyway, but verify via EXPLAIN-friendly direct query before SET NOT NULL to confirm backfill was complete).
  - `SELECT COUNT(*) FROM documents WHERE requirement_id IS NULL` = 0.
  - `SELECT COUNT(*) FROM requirements` = 3 (one per project).
  - `SELECT COUNT(*) FROM user_stories` = 2 (unchanged).
  - `SELECT COUNT(*) FROM documents` = 1 (unchanged).
- **Architecture cite:** US044 AC — "zero data loss, no orphaned rows"
