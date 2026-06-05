# US014 — ADR — formalise MCP-only-writes as the permanent write API

**Requirement:** REQ006 — tech debt backfill sprint
**Status:** draft

## Story
As a **future contributor (human or sub-agent) considering whether to add `POST` / `PUT` / `DELETE` endpoints to `api-server`**, I want **a written Architecture Decision Record (ADR) that explicitly states "MCP-only-writes" is the permanent architectural intent — not an oversight or a TODO**, so that I do not waste cycles proposing REST-writes again without first reading the rationale, the alternatives that were considered, and the conditions under which the decision could be revisited.

This story is a **documentation deliverable**. po-ba authors the AC; the system-architect writes the ADR text in `architecture.md` (REQ006) during Phase 1.

## Acceptance criteria

- **Scenario: `architecture.md` (REQ006) contains an explicit ADR section**
  - Given the system-architect is authoring `docs/requirements/REQ006_tech_debt_backfill_sprint/architecture.md` during Phase 1
  - When the architecture is approved (`Approval: approved`)
  - Then the document contains a clearly-marked ADR section (heading shape: `## ADR-001 — MCP-only-writes is the permanent write API` OR equivalent architect-chosen convention — see OQ-3 in README)
  - And the section appears in the document's table of contents (if a TOC is present)

- **Scenario: ADR states the decision explicitly**
  - Given the ADR section
  - When read
  - Then it contains a "**Decision**" subsection that states verbatim or in close paraphrase: "The `api-server` is intentionally read-only. All create / update / delete operations on projects, documents, user stories, tasks (and any future write surface) are exposed exclusively through MCP tools (registered in `internal/handler/*_tools.go` and served by `cmd/mcp-server`). REST `POST`/`PUT`/`DELETE` endpoints will NOT be added to `api-server` unless a future requirement explicitly overrides this ADR."

- **Scenario: ADR enumerates the four current `api-server` endpoints (the read-only surface)**
  - Given the ADR section
  - When read
  - Then it explicitly names the four `GET` endpoints currently exposed:
    1. `GET /api/v1/projects`
    2. `GET /api/v1/projects/:id`
    3. `GET /api/v1/projects/:id/documents`
    4. `GET /api/v1/documents/:id`
  - And confirms these (and any new read-only `GET` endpoints added in future REQs) are the entirety of `api-server`'s public surface

- **Scenario: ADR enumerates the MCP tool families (the write surface)**
  - Given the ADR section
  - When read
  - Then it names the MCP tool families that own all writes:
    1. `RegisterProjectTools` (5 tools: create / get / update / delete / list)
    2. `RegisterDocumentTools` (5 tools)
    3. `RegisterTaskTools` (5 tools)
    4. `RegisterUserStoryTools` (5 tools)
    5. `RegisterAuditTools` (read-side audit trail — included for completeness even though it's a read tool)
  - And confirms that any future entity requiring CUD operations adds an `*_tools.go` family, NOT REST endpoints

- **Scenario: ADR documents the rationale ("why this decision was taken")**
  - Given the ADR section
  - When read
  - Then it contains a "**Rationale**" subsection covering at least:
    1. **Single source of truth for writes.** All write paths go through the MCP tool registry, which means audit-trail invariants (`status_audit_trail` rows), domain-level transition guards (`IsValidTransition`), and transactional consistency (`BeginTx`/`Commit` in `*_repo.UpdateXxxStatus`) are uniformly enforced. A second write surface (REST) would require re-implementing all of this or risk drift.
    2. **Sub-agent-first design.** The project's primary client is the team of sub-agents (po-ba, system-architect, tech-lead, tester, be-dev, fe-dev, orchestrator). Sub-agents speak MCP natively. Adding REST writes for a hypothetical browser client is premature when no such client exists yet.
    3. **Operational surface area.** Fewer endpoints = fewer surfaces to monitor, fewer auth concerns, smaller CORS attack surface.

- **Scenario: ADR documents alternatives considered**
  - Given the ADR section
  - When read
  - Then it contains an "**Alternatives considered**" subsection covering at least:
    1. **REST writes added to `api-server`.** Rejected per the rationale above. Cost: doubles the write surface, requires audit/transition logic duplication or extraction into a shared service layer.
    2. **MCP-as-write + REST-as-read with shared service layer.** Considered but deferred indefinitely. Cost: introduces a `services/agent-board/internal/service/` layer that does not exist today; adds an abstraction step with no concrete consumer.
    3. **Bidirectional REST + MCP.** Rejected; same as #1 plus the operational surface concern.

- **Scenario: ADR documents the conditions under which the decision could be revisited**
  - Given the ADR section
  - When read
  - Then it contains a "**Conditions for revisiting**" subsection naming at least:
    1. A concrete non-sub-agent browser-direct client requirement (i.e. a user-facing UI that needs to create/update/delete without an MCP proxy).
    2. An external integration partner who cannot integrate MCP.
    3. A measured performance or operational benefit to a REST-write path that outweighs the duplication cost.
  - And explicitly states that "this decision is NOT revisited just because adding a REST endpoint would be technically easy in an isolated PR"

- **Scenario: tech-debt items are closed with `won't-fix` strike-throughs pointing to this ADR**
  - Given `docs/tech_debt.md` line 98 contains the finding `api-server only exposes 4 GET endpoints; ALL data writes go through MCP. Architecturally, this is by design (MCP-as-write-API). BUT it means: ... If the project ever wants browser-direct CRUD, add REST POST/PUT/DELETE endpoints`
  - And `docs/tech_debt.md` line 89 was already struck through (the REQ005/US006 `/api/v1/projects` POST → 405 finding) — verify with the related "no REST writes" finding
  - When this story is `done`
  - Then `docs/tech_debt.md` line 98 is struck through with `→ won't-fix per REQ006/US014 ADR (MCP-only-writes is permanent)`
  - And any other `tech_debt.md` line that exists today and reflects "we should add REST writes" is similarly struck with a pointer to the ADR

- **Scenario: tester confirms no tests are required for this story**
  - Given this is a pure-documentation story
  - When the tester reviews the story for spec authoring
  - Then `US014_be_unit_tests.md` is written but contains exactly ONE row asserting that "this story is documentation-only; no executable tests apply"
  - And `US014_fe_unit_tests.md` contains the same disclaimer
  - And `US014_e2e_tests.md` contains the same disclaimer
  - And **no Robot Framework `.robot` file is created** for US014
  - And the test report (Phase 3c) for US014 confirms zero tests, zero failures, zero skipped — by design

- **Scenario: tech-lead's plan for this story is "no implementation tasks"**
  - Given this story's deliverable is entirely in `architecture.md` (which the architect writes during Phase 1, BEFORE Phase 2)
  - When tech-lead reviews the story for task decomposition
  - Then tech-lead writes a single task file (`US014_be_adr_verification.md` or similar) whose `Track: BE-meta` task description is: "Verify `architecture.md` (REQ006) ADR section meets all AC scenarios above. If any AC is missing, raise `ARCHITECTURE_GAP_FOUND` to route back to system-architect."
  - And no `services/` or `web/` code changes are involved
  - And the task is a documentation review, not implementation

- **Scenario: sign-off verifies the ADR exists and is complete**
  - Given the architecture is approved AND the verification task is `completed`
  - When po-ba reviews for sign-off
  - Then po-ba reads the `architecture.md` ADR section and verifies each of the AC scenarios above
  - And po-ba verifies the strike-throughs in `docs/tech_debt.md` are in place
  - And if all AC pass, the story flips to `Status: done`

## UI / UX flow expectations

**No UI: documentation-only story.** Operational expectations:

- **Reader flow:** a future contributor (sub-agent or human) reads `architecture.md` (REQ006) before considering REST-write additions and is directed to the ADR section.
- **Discoverability:** the ADR is in `architecture.md` (REQ006) per OQ-3. If the architect prefers a separate `docs/adr/0001-mcp-only-writes.md`, that is acceptable as long as `architecture.md` (REQ006) links to it prominently.
- **No log lines, no startup messages, no API behaviour change.**

## Out of scope
- **Removing any existing REST endpoint.** The four `GET` endpoints stay.
- **Adding REST writes.** That is the opposite of this story.
- **Migrating MCP tools to a different protocol.** MCP stays.
- **Extracting a shared service layer.** Per "Alternatives considered" #2, deferred.
- **Writing the actual ADR text in the story file.** The story file (this one) authors the AC; the architect authors the ADR text in `architecture.md`.

## Dependencies
- None directly. Architect resolves OQ-3 (ADR location) in `architecture.md` during Phase 1.

## Notes for the team

- **This is a "meta" story.** Most of the deliverable is in `architecture.md`, which the architect produces during Phase 1 (BEFORE the human approves the architecture in the HARD STOP). That means when the architecture is approved, this story is effectively 90% done — the remaining 10% is the tech-debt strike-throughs and the tech-lead's verification task.
- **Story IS a documentation deliverable.** po-ba intentionally wrote the AC scenarios to be verifiable against the document text. The architect's job is to write the document; the tech-lead's job is to verify; po-ba's job is to sign off.
- **`won't-fix` strike-through phrasing.** Use the existing convention from `docs/tech_debt.md` line 14 / line 79: prefix with `~~`, suffix with `~~ → won't-fix per REQ006/US014 ADR`. Match the prevailing style.
- **Why this story is "BE-meta" not "FE" or "BE-prod".** The artefact is `architecture.md`. The verification is reading a markdown file. There is no service or frontend involved. The track label is a hint to the orchestrator; the architect-then-tech-lead-then-po-ba flow is the same as any other story.
- **No `*_test.go` / no `*.test.tsx` / no `*.robot`.** Confirmed in the AC. tester writes the disclaimer-only spec files; orchestrator's Phase 3c test report records zero tests.

## Sign-off log
(po-ba appends here on each sign-off pass)
