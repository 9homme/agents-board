# US044/be_requirement_schema_migration_domain

**Requirement:** REQ008
**Story:** US044
**Track:** BE
**Service:** services/agent-board
**Status:** completed
**Blocked by:** none
**Worked-by:**
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

> **Note (tech-lead-reviewer):** The dev's original `## Notes` gate evidence was
> lost when the Phase 1+2 scaffolding merge (`bb98317`) overwrote this task file's
> metadata after the dev's implementation commits (`124c991`/`25d9efd`/`4ff567e`/
> `42f33e6`/`2b7b0ba`/`9da322e`/`a7318af`) had already landed. Because the pasted
> text was a merge casualty (not a dev omission), the reviewer re-ran every gate
> dimension first-hand on the integrated branch and recorded the verbatim output
> below before approving.

REVIEW GATE: PASS  (BE · services/agent-board)
```
== BE gate · services/agent-board ==
  PASS  gofmt -s (no diff)
  PASS  go vet ./...
  PASS  golangci-lint run ./...
  PASS  go test ./...
WARN  gosec (skipped — not installed)
WARN  govulncheck (skipped — not installed)
REVIEW GATE: PASS
```

REVIEW GATE: PASS  (cross)
```
== Cross-cutting · repo ==
  PASS  semgrep (owasp/golang/typescript)
  PASS  gitleaks (no secrets)
REVIEW GATE: PASS
```

Unit/integration suite: `go vet ./...` clean; `go test ./...` → 404 passed, 0 failed.
IT-044-001..010 executed against real local Postgres (10 passed, 0 skipped).

Coverage (per-file / per-package):
- `internal/domain/requirement.go` `NewRequirement` — 100.0%
- `internal/domain` package — 90.5%
- `internal/migrate` package — 92.9%
(struct-only files project.go/user_story.go/document.go have no executable statements)

robot --dryrun tests/e2e/REQ008_*/  →  19 tests, 19 passed, 0 failed (parse OK)

## Review log

### Review pass 1 — verdict: approved

**Reviewer:** tech-lead-reviewer (Mode 1 — Task Code Review). Date: 2026-06-10.

**Gate evidence (re-verified first-hand — see `## Notes`):**
- BE gate: `REVIEW GATE: PASS`; cross gate: `REVIEW GATE: PASS`.
- `go vet ./...` clean; `go test ./...` → 404 passed, 0 failed.
- Coverage: `requirement.go::NewRequirement` 100%; domain pkg 90.5%; migrate pkg 92.9% (all ≥80%).
- `robot --dryrun tests/e2e/REQ008_*/` → 19 tests, 0 failed (parse OK).

**Test contract — all 16 IDs implemented and passing:**
- Unit: UT-044-001..006 in `internal/domain/requirement_test.go` (status default + enum completeness, full `Requirement` struct, `UserStory.RequirementID`, `Document.RequirementID`, `Project.Path`).
- Integration: IT-044-001..010 in `internal/migrate/migration_it_test.go`, all run against a real Postgres (not skipped): table+columns+FK CASCADE+CHECK+index, `requirement_id` NOT NULL on both children with NULL-insert rejection, Default-per-project backfill (incl. childless P3), zero-data-loss re-parenting of stories+documents, `projects.path` TEXT NOT NULL + UNIQUE, duplicate-path 23505 rejection, down-migration reversal, no-orphan row-count invariant.

**Architecture conformance:** PASS. Domain types match the extract field-for-field with exact JSON tags (`id`/`projectId`/`name`/`description`/`status`/`createdAt`/`updatedAt`; `path`; `requirementId`). Migration produces the required schema; runner embeds only `*.up.sql` so `.down.sql` is documentation-only as designed.

**Justified deviation from "verbatim" migration SQL (NOT a defect):** the extract's `ALTER TABLE projects ADD COLUMN path TEXT NOT NULL DEFAULT '';` + `UNIQUE (path)` would throw a 23505 unique-violation whenever 2+ projects pre-exist (every existing row gets `''`). The dev correctly substituted the canonical safe pattern: add nullable → backfill each row with its `id::text` (guaranteed-unique placeholder) → `SET NOT NULL` → add UNIQUE. This is a correctness fix, verified by IT-044-007/008. Filed as tech-debt #14 (fix the extract, not the code).

**Test-spec exhaustiveness (anti-happy-path):** migration is declarative SQL inside a single transaction in the runner; error paths (constraint violations) are positively exercised — NULL `requirement_id` rejection (IT-044-002/003) and duplicate-path 23505 (IT-044-008). `NewRequirement` is a straight-line constructor with no branches. No uncovered error branch ⇒ no spec gap on coverage grounds.

**TDD honesty:** PASS. Commit history shows red-before-green per track: `124c991 red:` → `25d9efd green:` → `4ff567e refactor:` (domain); `42f33e6 red:` → `2b7b0ba green:` → `9da322e refactor:` (migration). Tests assert behavior/schema, not implementation accidents.

**Scope:** PASS. Diff touches only the 6 declared files + the two test files. No handler/repo/MCP/HTTP/route code touched.

**TDG conformance:** PASS. Every subject starts `red:`/`green:`/`refactor:` and ends `(US044)`; order is red→green→refactor within each track. (`refactor: chore:` double-prefix on the handoff commit is cosmetic drift already tracked as tech-debt #4.)

**Quality:** No commented-out code, no unowned TODOs, no log spam. Doc comments present on the new public exports (`Requirement`, `NewRequirement`, status constants).

**Tech-debt filed this pass:** #14 (extract's broken verbatim path SQL), #15 (`US044_be_unit_tests.md` UT-044-004 lists a non-existent `TaskCount` field), #16 (IT-044-* skip rather than fail when no Postgres — CI must provision a DB).

**Verdict: approved → Status: completed.**
