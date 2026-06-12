# US044 — Introduce Requirement entity + re-parent model

**Requirement:** REQ008 — Requirement entity + project local-path linking
**Status:** done

## Story
As a system maintainer, I want a first-class Requirement (REQ) entity between Project and User Story, with existing User Stories and Documents re-parented under a Requirement and a `path` column added to Project, so that the data model matches how work is organised on disk (`docs/requirements/REQ[ID]_*/`) and nothing in the current dataset is lost.

## Acceptance criteria
- **Scenario: Requirement table exists**
  - Given a fresh database
  - When migrations run
  - Then a `requirements` table exists with columns: `id`, `project_id` (FK → projects, ON DELETE CASCADE), `name`, `description`, `status`, `created_at`, `updated_at`
- **Scenario: Requirement status is a stored enum defaulting to draft**
  - Given the `requirements` table
  - When a requirement is inserted without specifying status
  - Then `status` defaults to `draft`, and the column accepts the values `draft`, `in_progress`, `done` (stored enum, no state-machine enforcement this REQ)
- **Scenario: User stories belong to a Requirement**
  - Given the migration has run
  - When inspecting `user_stories`
  - Then it has a `requirement_id` FK referencing `requirements(id)` ON DELETE CASCADE, declared NOT NULL, and every existing row has a non-null `requirement_id`
- **Scenario: Documents belong to a Requirement**
  - Given the migration has run
  - When inspecting `documents`
  - Then it has a `requirement_id` FK referencing `requirements(id)` ON DELETE CASCADE, declared NOT NULL, and every existing row has a non-null `requirement_id`
- **Scenario: Existing projects auto-migrated into a Default Requirement**
  - Given existing projects each with user stories and/or documents
  - When the migration runs
  - Then exactly one Requirement named "Default" (status `draft`) is created per existing project, and that project's existing User Stories and Documents are re-parented under it, with zero data loss and no orphaned rows
- **Scenario: Project gains a unique NOT NULL path column**
  - Given the migration has run
  - When inspecting `projects`
  - Then it has a `path` `TEXT NOT NULL` column with a uniqueness constraint (two projects cannot share the same path)
- **Scenario: Domain models updated**
  - Given the Go domain layer
  - When code is compiled
  - Then a `Requirement` domain type exists (with `Id`, `ProjectId`, `Name`, `Description`, `Status`, `CreatedAt`, `UpdatedAt`), `UserStory` and `Document` carry a `RequirementId`, and `Project` carries a required `Path` (non-pointer `string`)
- **Scenario: Migration is reversible**
  - Given the up migration has run
  - When the down migration runs
  - Then the schema returns to its prior shape without error (data-loss expectations for the down path documented in the migration)

## UI / UX flow expectations
No UI: this story is the data-model + migration + domain-type foundation only. User-facing surfaces are delivered in US045 (API), US046 and US047 (web).

## Out of scope
- HTTP endpoints / API changes (US045).
- Any web UI (US046, US047).
- Reading or importing the contents of the linked local path.

## Dependencies
- None (foundation story). US045, US046, US047 depend on this.

## Notes for the team
- Mirror the on-disk contract: a Requirement maps to a `REQ[ID]_*/` folder; its User Stories map to `US*.md`; its Documents map to `README.md`/other docs.
- The migration backfill creates exactly one Requirement named "Default" per project that has children; re-parent all of that project's User Stories and Documents under it. Zero data loss is mandatory — verify counts before/after.
- `requirement_id` is NOT NULL on both `user_stories` and `documents` (confirmed: REQ is mandatory). Backfill must complete before the NOT NULL constraint is enforced (add column nullable → backfill → set NOT NULL, or equivalent).
- `Requirement.status` is a plain stored enum (`draft` | `in_progress` | `done`, default `draft`). No transition rules this REQ.
- `Project.path` is `TEXT NOT NULL` + unique — every project links to a distinct local directory; no path-less projects. Add the column with a transient `DEFAULT ''` only to satisfy NOT NULL on the ALTER over existing rows (the application layer always requires a non-blank path, so `''` is never reachable via the API); the uniqueness constraint rejects duplicate paths.

## Sign-off log
(po-ba appends here on each sign-off pass)

### Sign-off pass 1 — 2026-06-12 — verdict: approved
- **Spec review:** All 8 AC scenarios map to tests. Requirement table + columns + FK/CASCADE/index → IT-044-001. Status enum default → UT-044-001/002 + IT-044-001 CHECK constraint. user_stories.requirement_id NOT NULL + FK → IT-044-002 (incl. NULL-insert violation edge). documents.requirement_id NOT NULL + FK → IT-044-003. Default-requirement backfill (one per project) → IT-044-004; zero-data-loss re-parenting → IT-044-005 (stories) + IT-044-006 (documents); no-orphans invariant → IT-044-010. projects.path TEXT NOT NULL + unique → IT-044-007, duplicate-rejection edge → IT-044-008. Domain models (Requirement type, RequirementId on UserStory/Document, Path on Project) → UT-044-003/004/005/006. Reversible migration → IT-044-009 (down path correctly noted as documentation-only). Pyramid is honest: schema/backfill at integration, struct shape at unit, only a single read-path smoke at e2e.
- **Result review:** Report shows UT-044-001..006 PASS and IT-044-001–010 (migration suite) PASS; E2E-044-001 (path field on read path) and E2E-044-002 (Default requirement post-migration) PASS. No skips. Counts consistent with the spec.
- **Routed to:** none
