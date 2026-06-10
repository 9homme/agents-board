# REQ008 — Requirement entity + project local-path linking

## Summary
The data model is missing a **Requirement (REQ)** level between Project and User Story. Today the hierarchy is `Project → Documents / User Stories → Tasks`. This requirement introduces a first-class **Requirement** entity so the model matches the on-disk `docs/requirements/REQ[ID]_*/` contract, and adds the ability to **create a new Project from the web by linking it to a required local path** on disk (so a project points at a real multi-agents project directory). The user types the full path into a plain text field; the backend validates it via `os.Stat`. Requirement records are created by the po-ba agent via **MCP tools**, not the web.

## Business goal
- Correct the data model so a Project groups Requirements, and each Requirement owns its User Stories and Documents — mirroring how work is actually organised on disk (`REQ[ID]_*/README.md`, `US*.md`, ...).
- Let a human create/register a project from the web by pointing it at a local directory (a required, validated path), laying the foundation for the (out-of-scope) live multi-agent control feature.

## Confirmed data model (this REQ)
```
Project ──< Requirement ──< UserStory ──< Task
                      └────< Document
Project: + path (NOT NULL, required local path to the multi-agents project on disk)
```
- `Requirement`: `id`, `project_id` (FK → projects), `name`, `description`, `status`, `created_at`, `updated_at`. Mirrors a `REQ[ID]_*/` folder.
- `UserStory.requirement_id`: parent FK → requirements (NOT NULL after migration).
- `Document.requirement_id`: parent FK → requirements (NOT NULL after migration). Both User Stories and Documents live under a Requirement, mirroring `US*.md` + `README.md` inside `REQ[ID]_*/`.
- `Project.path`: new `TEXT NOT NULL` column — required local path string. Unique across projects. Every project links to a distinct directory; no path-less projects.
- `Requirement.status`: a stored enum column (workflow state), same style as task/story status. Values: `draft`, `in_progress`, `done`. Default `draft`. No state-machine enforcement this REQ — it is a plain stored enum.

## Confirmed decisions
- **D-1 — REQ placement:** `Project → REQ → UserStory → Task`. **Both User Stories and Documents re-parent under REQ.** `requirement_id` is NOT NULL on both after migration.
- **D-2 — Migration:** Existing data is **auto-migrated with zero data loss**: one "Default" Requirement is created per existing Project, and that project's existing User Stories and Documents are re-parented under it. No orphans.
- **D-3 — Path validation:** The backend **stores the path string and verifies via `os.Stat` that it exists on disk and is a directory**; otherwise the create request is rejected (400). Path is **required** (NOT NULL) — absent/blank `path` → 400.
- **D-3b — Path uniqueness:** A project's `path` must be **unique** across projects — a duplicate path returns **409**.
- **D-4 — Requirement create is MCP-only (revised):** There is **no** HTTP `POST` to create requirements. Requirements are created via the MCP tools `create_requirement` / `list_requirements`, called by po-ba (agent) at the end of Phase 1. The web reads requirements via `GET /api/v1/projects/:id/requirements` only — the web is read-only for requirements.
- **D-5 — No filesystem autocomplete (revised):** No `fs/suggestions` endpoint, no autocomplete dropdown, no `PathAutocomplete` component, no `usePathSuggestions` hook, no `web/lib/api/fs.ts`. The "Add Project" form uses a **plain `<input type="text">`**; the user types the full path manually; project **name auto-fills from the path basename** and stays editable; **both name and path are required**.
- **D-6 — Navigation:** Project detail page shows its Requirements; selecting a Requirement shows that Requirement's User Stories + Documents. The linked path (always present) is displayed read-only.

## NOT in scope (explicitly excluded)
- **Live Activity** (subagent live feed / response stream).
- **Token tracking** (token-consumed meters).
- **Permission UI** (pending-permission allow/deny).
- **Agent chat** (chat-to-agents textbox).
- Any execution/control of the multi-agents virtual team.
- Reading/parsing/importing the contents of the linked local path (file sync). REQ008 only stores and validates the path.
- Editing or deleting a project's path after creation.

## User stories
| ID | Title | Track(s) | Status |
|---|---|---|---|
| US044 | Introduce Requirement entity, re-parent US + Documents, add Project.path NOT NULL (schema + migration + domain) | BE | draft |
| US045 | Requirement read API (HTTP list + MCP create/list) + Project-create-with-required-path API | BE | draft |
| US046 | Add Project from web by linking a local path (plain text input) | FE | draft |
| US047 | Requirement-level navigation on the project detail page | FE | draft |
| US048 | Fix inconsistent REST design: add nested detail endpoints for documents and user stories (keep top-level routes for backward compat) | BE | draft |
| US049 | Add `blocked_review_gate` task status to the domain state machine (constant + `in_review`/`changes_requested` → `blocked_review_gate`, terminal; MCP `update_task` accepts it; no migration) | BE | draft |

## Engineering tasks (tech-lead-planner, Phase 2)

| Task file | Title | Track | Blocked by | Status |
|---|---|---|---|---|
| US044_be_requirement_schema_migration_domain.md | Migration 000003 + Requirement domain + Project.path + re-parent US/Documents | BE | none | pending |
| US045_be_requirement_repo_and_list_api.md | RequirementRepository + HTTP `GET /projects/:pid/requirements` (§4) | BE | US044_be_requirement_schema_migration_domain | pending |
| US045_be_requirement_mcp_tools.md | MCP create/list/update_requirement (§5) + breaking `create_user_story`/`create_document` requirement_id (§12/§13) | BE | US045_be_requirement_repo_and_list_api | pending |
| US045_be_project_create_with_path.md | fsutil + ProjectRepo path + `POST /projects` (§3) + MCP create_project path (D-008) | BE | US045_be_requirement_repo_and_list_api | pending |
| US046_fe_add_project_dialog.md | Add Project dialog: plain-text path, basename auto-fill, createProject (§3), inline 400/409 | FE | none | pending |
| US047_fe_requirement_navigation.md | Requirement selector + read-only path + tabs re-scoped to canonical hierarchy (§2/§4/§6/§10) | FE | none | pending |
| US048_be_nested_hierarchy_routes.md | Full canonical hierarchy routes (§6–§11) + chain guards; remove 8 flat routes (D-009) | BE | US044_be_requirement_schema_migration_domain | pending |
| US049_be_blocked_review_gate_status.md | `blocked_review_gate` task status in state machine (terminal; from in_review/changes_requested) | BE | none | pending |

**Dependency graph:**
- `US044_be_requirement_schema_migration_domain` → blocks `US045_be_requirement_repo_and_list_api` and `US048_be_nested_hierarchy_routes`.
- `US045_be_requirement_repo_and_list_api` → blocks `US045_be_requirement_mcp_tools` and `US045_be_project_create_with_path` (both consume the RequirementRepository/repo wiring it creates).
- `US046_fe_add_project_dialog`, `US047_fe_requirement_navigation`, `US049_be_blocked_review_gate_status` have no blockers (parallel-ready immediately, alongside US044).
- BE↔FE for the same story have no cross-track `Blocked by` — they meet only at the locked API contract.

**Shared-file collision notes (for the orchestrator's 3a co-pick avoidance):**
- `cmd/api-server/main.go` — edited by US045_be_requirement_repo_and_list_api (add §4 GET), US045_be_project_create_with_path (add §3 POST), and US048 (remove 8 / add 6). Sequence or accept small merges; US048 must NOT delete the §4 route.
- `internal/repo/requirement_repo.go` & `internal/handler/requirement_handler.go` — created by US045_be_requirement_repo_and_list_api; consumed by US045_be_requirement_mcp_tools and US048 (which may add `GetRequirement`).
- `internal/repo/user_story_repo.go` & `document_repo.go` — US045_be_requirement_mcp_tools changes INSERTs; US048 changes SELECTs/adds ListByRequirement (same files, different methods).
- `web/lib/api/types.ts` & `web/test/msw/handlers.ts` — both US046 and US047 add disjoint declarations; `Project.path` added by whichever lands first.
