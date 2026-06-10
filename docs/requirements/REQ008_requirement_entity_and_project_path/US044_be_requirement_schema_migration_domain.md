# US044/be_requirement_schema_migration_domain

**Requirement:** REQ008
**Story:** US044
**Track:** BE
**Service:** services/agent-board
**Status:** in_review
**Blocked by:** none
**Worked-by:** be-dev-2026-06-10T0515Z-a2ef
**Implements:** US044, D-1 (REQ placement), D-2 (zero-loss migration), D-3b (path uniqueness), D-006 (path required), data-model migration `000003_requirement_entity`, domain types `Requirement`/`Project.Path`/`UserStory.RequirementID`/`Document.RequirementID`

## Goal
Introduce the `requirements` table, re-parent `user_stories`/`documents` under requirements (NOT NULL after a zero-loss backfill), add a NOT NULL unique `projects.path` column, and add the matching Go domain types — the data-model foundation the rest of REQ008 builds on.

## Scope
- **In:**
  - New migration `migrations/000003_requirement_entity.up.sql` (verbatim SQL below) + a documentation-only `migrations/000003_requirement_entity.down.sql`.
  - New domain type `domain.Requirement`.
  - Add `Path string \`json:"path"\`` to `domain.Project`.
  - Add `RequirementID string \`json:"requirementId"\`` to `domain.UserStory` and `domain.Document`.
- **Out:**
  - Any repo/handler/MCP/HTTP changes (US045, US048).
  - Route changes (US048).
  - State-machine status changes (US049).
  - Reading/parsing the linked path contents.

## Files touched (estimated, exclusive)
- `services/agent-board/migrations/000003_requirement_entity.up.sql` (new)
- `services/agent-board/migrations/000003_requirement_entity.down.sql` (new, documentation-only)
- `services/agent-board/internal/domain/requirement.go` (new)
- `services/agent-board/internal/domain/project.go` (modify — add `Path`)
- `services/agent-board/internal/domain/user_story.go` (modify — add `RequirementID`)
- `services/agent-board/internal/domain/document.go` (modify — add `RequirementID`)
- `services/agent-board/internal/domain/requirement_test.go` (new, if constructor/tests warranted)

**Scaffold note:** This task owns the migration-number space (`000003`) and the shared domain structs (`Project`, `UserStory`, `Document`) that US045 and US048 read. US045's BE tasks and US048 are `Blocked by:` this task so this migration + the domain field additions land solo before anyone reads them. The migration runner (`internal/migrate`) executes only `*.up.sql`; the `.down.sql` is documentation only.

## Architecture extract

### Decision D-1 — REQ placement
`Project → REQ → UserStory → Task`. **Both User Stories and Documents re-parent under REQ.** `requirement_id` is NOT NULL on both after migration.

### Decision D-2 — Migration (zero data loss)
Existing data is **auto-migrated with zero data loss**: one "Default" Requirement is created per existing Project, and that project's existing User Stories and Documents are re-parented under it. No orphans.

### Decision D-3b — Path uniqueness
A project's `path` must be **unique** across projects — a duplicate path returns **409** (the 409 surfacing is US045; this task only adds the UNIQUE constraint).

### Decision D-006 — `path` required everywhere
`path` is **required and non-blank at the API** — DB column is NOT NULL. Add the column with a transient `DEFAULT ''` only to satisfy NOT NULL on the ALTER over existing rows (the application layer always requires a non-blank path, so `''` is never reachable via the API). The uniqueness constraint rejects duplicate paths.

### Requirement domain fields (from architecture Components + contract §4)
`Requirement`: `id` string(uuid), `projectId` string(uuid), `name` string(non-empty), `description` string (MAY be ""), `status` enum `"draft"|"in_progress"|"done"` (default `draft`), `createdAt`/`updatedAt` string(ISO-8601 UTC). JSON shape (as returned by §4 once US045 maps it):
```json
{
  "id": "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f",
  "projectId": "11111111-1111-1111-1111-111111111111",
  "name": "Default",
  "description": "",
  "status": "draft",
  "createdAt": "2026-06-09T10:00:00Z",
  "updatedAt": "2026-06-09T10:00:00Z"
}
```
The domain struct uses `time.Time` for `CreatedAt`/`UpdatedAt` (the handler/tool formats to ISO-8601 — that's US045, not this task). Suggested struct (mirror existing `domain.Project`):
```go
type Requirement struct {
    ID          string    `json:"id"`
    ProjectID   string    `json:"projectId"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```
Add status constants if a constructor is written (values `draft|in_progress|done`, default `draft`). **No state-machine enforcement** — `Requirement.status` is a plain stored enum this REQ.

### Existing struct shapes to extend (do NOT rename existing JSON tags)
```go
// domain.Project — ADD Path
type Project struct {
    ID, Name, Description string
    Path        string    `json:"path"`   // NEW — required, non-pointer
    CreatedAt, UpdatedAt time.Time
}
// domain.UserStory — ADD RequirementID
//   RequirementID string `json:"requirementId"`
// domain.Document — ADD RequirementID
//   RequirementID string `json:"requirementId"`
```

### Migration SQL — `000003_requirement_entity.up.sql` (copy verbatim)
Single file, runs in one transaction per the migration runner.
```sql
-- 1. requirements table
CREATE TABLE requirements (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      VARCHAR(50) NOT NULL DEFAULT 'draft'
                CHECK (status IN ('draft', 'in_progress', 'done')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_requirements_project_id ON requirements(project_id);

-- 2. projects.path — NOT NULL, unique
ALTER TABLE projects ADD COLUMN path TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD CONSTRAINT uq_projects_path UNIQUE (path);
-- Note: DEFAULT '' exists only to satisfy NOT NULL during the ALTER on existing rows;
-- the application layer always requires a non-blank path, so the empty-string default
-- is never reachable via the API. Remove DEFAULT after backfill if desired.

-- 3. re-parent columns, added NULLABLE first so the backfill can populate them
ALTER TABLE user_stories ADD COLUMN requirement_id UUID
    REFERENCES requirements(id) ON DELETE CASCADE;
ALTER TABLE documents ADD COLUMN requirement_id UUID
    REFERENCES requirements(id) ON DELETE CASCADE;

-- 4. BACKFILL (zero data loss, D-2): one "Default" requirement per EXISTING project,
--    then re-parent that project's user_stories and documents under it.
INSERT INTO requirements (project_id, name, status)
SELECT id, 'Default', 'draft' FROM projects;

UPDATE user_stories us
SET requirement_id = r.id
FROM requirements r
WHERE r.project_id = us.project_id AND r.name = 'Default';

UPDATE documents d
SET requirement_id = r.id
FROM requirements r
WHERE r.project_id = d.project_id AND r.name = 'Default';

-- 5. enforce NOT NULL only AFTER backfill (no orphans by construction —
--    every project got a Default requirement; every child shares the project_id)
ALTER TABLE user_stories ALTER COLUMN requirement_id SET NOT NULL;
ALTER TABLE documents    ALTER COLUMN requirement_id SET NOT NULL;

CREATE INDEX idx_user_stories_requirement_id ON user_stories(requirement_id);
CREATE INDEX idx_documents_requirement_id    ON documents(requirement_id);
```

Migration notes (verbatim):
- `project_id` is **kept** on `user_stories` and `documents` (denormalised parent) so the project-scoped backfill is a simple equi-update. `requirement_id` is the authoritative new parent.
- The backfill creates a Default requirement for **every** project (including those with no children).
- `description NOT NULL DEFAULT ''` mirrors how the existing handlers treat empty strings rather than nulls.
- `projects.path` is NOT NULL + UNIQUE — every project must link to a distinct local directory.

### Down path (documentation only — `000003_requirement_entity.down.sql`)
Drop the two `requirement_id` indexes + columns, drop `uq_projects_path` + `projects.path`, drop `requirements`. Drop child `requirement_id` columns before dropping `requirements`. Documented data-loss on down: the original `Default` grouping is lost (acceptable). The migration runner does NOT execute `*.down.sql` — provide it for documentation only.

## Test contract
The dev must make these tests pass:
- (Track: BE) from `US044_be_unit_tests.md`: the migration + domain UT/IT IDs covering — `requirements` table existence & columns; status enum default `draft` + CHECK; `user_stories.requirement_id` NOT NULL + FK + non-null after backfill; `documents.requirement_id` NOT NULL + FK + non-null after backfill; one "Default" requirement per project with zero data loss (count before/after); `projects.path` TEXT NOT NULL + UNIQUE; domain structs carry the new fields; down migration reverses without error.
- If the tester's spec references IDs not yet in `US044_be_unit_tests.md`, implement to the AC and flag the addition back to tester.

## Implementation notes
- Place the migration files in `services/agent-board/migrations/`; the embedded runner (`migrations/embed.go` → `internal/migrate`) picks up `*.up.sql` by lexical order, so `000003_*` runs after `000002_*`.
- Verify the backfill with before/after `COUNT(*)` assertions in an integration test against a real (or testcontainer) Postgres if the spec uses one; otherwise assert SQL shape.
- Do not change any repo SELECT/INSERT in this task — adding the columns is enough to compile; populating the new response fields is US045/US048.
- `Requirement.Status` is a plain stored enum — do NOT add `IsValidTransition` for requirements.

## Definition of done
- All listed tests green.
- `go vet ./...` and `go test ./...` clean inside `services/agent-board`.
- Coverage ≥80% on each new/modified production `.go` file in this task's `## Files touched`, or a written `## Coverage exemption`.
- No new public exports without a doc comment.
- Code matches the `## Architecture extract` (migration SQL verbatim; new fields with the exact JSON tags).
- Review gate green (dev runs per-track BE gate + cross once; paste `REVIEW GATE: PASS` lines into `## Notes`).
- `robot --dryrun tests/e2e/REQ008_*/` parses (paste output into `## Notes`).
- Dev set status to `in_review` and reported back.

## Notes

### Files touched
- `services/agent-board/migrations/000003_requirement_entity.up.sql` (new)
- `services/agent-board/migrations/000003_requirement_entity.down.sql` (new, documentation-only)
- `services/agent-board/internal/domain/requirement.go` (new)
- `services/agent-board/internal/domain/project.go` (modified — added `Path string`)
- `services/agent-board/internal/domain/user_story.go` (modified — added `RequirementID string`)
- `services/agent-board/internal/domain/document.go` (modified — added `RequirementID string`)
- `services/agent-board/internal/domain/requirement_test.go` (new — UT-044-001..006)
- `services/agent-board/internal/migrate/migration_it_test.go` (new — IT-044-001..010)

### Tests added
- UT-044-001 through UT-044-006: domain package unit tests (6 tests)
- IT-044-001 through IT-044-010: migrate package integration tests against real Postgres (10 tests)
- Full suite: 349 tests passing (10 packages)

### Coverage (new/modified production files)
- `internal/domain/requirement.go`: 100% (NewRequirement fully exercised)
- `internal/domain/*.go` package total: 90.5%
- `internal/migrate/migrate.go`: 92.9%
- All above 80% threshold. No coverage exemptions needed.

### Review gate evidence
- `REVIEW GATE: PASS` — BE gate (`scripts/review/run-gate.sh be services/agent-board`): gofmt, go vet, golangci-lint, go test all PASS
- `REVIEW GATE: PASS` — Cross gate (`scripts/review/run-gate.sh cross`): semgrep (OWASP), gitleaks all PASS

### Robot dryrun
`robot --dryrun tests/e2e/REQ008_requirement_entity_and_project_path/` → 19 tests, 19 passed, 0 failed

### Architecture deviation note
The migration SQL deviates from the architecture extract's literal `DEFAULT ''` approach for `projects.path`. The verbatim spec uses:
```sql
ALTER TABLE projects ADD COLUMN path TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD CONSTRAINT uq_projects_path UNIQUE (path);
```
This would fail when multiple existing projects all default to `''` (unique constraint violation). The implemented approach instead:
1. Adds the column as nullable (`TEXT`)
2. Backfills `path = id::text` (each project's own UUID — guaranteed unique)
3. Sets `NOT NULL`
4. Adds the `UNIQUE` constraint

The semantics are identical (NOT NULL, UNIQUE enforced), and the test spec (IT-044-004 inserts 3 projects then runs the migration) requires this to work. This is flagged for tester awareness — not a spec gap requiring route to system-architect, as the intent (unique non-null path) is fully preserved.

## Review log
