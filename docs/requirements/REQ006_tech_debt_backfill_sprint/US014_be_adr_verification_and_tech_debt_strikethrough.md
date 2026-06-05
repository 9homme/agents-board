# US014/be_adr_verification_and_tech_debt_strikethrough

**Requirement:** REQ006
**Story:** US014
**Track:** BE
**Service:** services/agent-board   (nominal — this task touches NO Go code; meta/docs-only — see Notes)
**Status:** pending
**Blocked by:** none
**Worked-by:**
**Implements:** REQ006/US014 AC (all scenarios — architecture.md contains an explicit ADR section, decision stated, four read-only endpoints enumerated, five MCP tool families enumerated, rationale documented, alternatives considered, conditions-for-revisiting documented, tech-debt strike-through applied, no executable tests required), architecture §3 US014 touch row, architecture §9 (ADR-001 text — INLINE per D-008), architecture §10.3 (no TDG, no live e2e, no react-doctor; grep + manual read verification).

## Goal
**This is a documentation-verification + tech-debt-strike-through task.** The ADR text was authored by system-architect in `docs/requirements/REQ006_tech_debt_backfill_sprint/architecture.md` §9 (REQ006 rev 5, approved). The "dev" work for US014 reduces to:

1. **Verify** `architecture.md` §9 satisfies every AC scenario in `US014_adr_mcp_only_writes.md` — if any AC is missing, raise `ARCHITECTURE_GAP_FOUND` to route back to system-architect (do NOT patch the ADR yourself).
2. **Strike-through** `docs/tech_debt.md` line 98 with `→ won't-fix per REQ006/US014 ADR-001 (MCP-only-writes is permanent)`, and any other line that exists today and reflects "we should add REST writes" (US014 AC explicit).

No code, no tests, no react-doctor sweep, no live e2e (architecture §10.3).

## Scope
- **In:** Read `architecture.md` §9 (lines that hold ADR-001) and check it against each AC scenario in `US014_adr_mcp_only_writes.md`:
  - §9.1 Status (Accepted, effective date, author, approver).
  - §9.3 Decision (verbatim or close-paraphrase of "api-server is intentionally read-only…REST POST/PUT/DELETE will NOT be added unless a future REQ explicitly overrides").
  - §9.4 The four read-only `GET` endpoints enumerated.
  - §9.5 The MCP tool families enumerated.
  - §9.6 Rationale (≥ the 3 bullets from AC).
  - §9.7 Alternatives considered (≥ the 3 bullets from AC).
  - §9.8 Conditions for revisiting (≥ the 3 bullets from AC + the "not because it'd be easy" disclaimer).
- **In:** Edit `docs/tech_debt.md` to strike-through line 98 with `→ won't-fix per REQ006/US014 ADR-001 (MCP-only-writes is permanent)`. Sweep for any other tech-debt line that proposes "add REST writes" and strike-through with a pointer to the ADR.
- **Out:** Any edit to `architecture.md` itself — that's system-architect's domain. If a gap is found, raise `ARCHITECTURE_GAP_FOUND` and STOP. Any code, test, MSW handler, Robot file. Any change to `web/`, `services/`, `cmd/`. Any new ADR convention under `docs/adr/` — D-008 explicitly chose inline.

## Files touched (estimated, exclusive)
- `docs/tech_debt.md` (edit — strike-through line 98 + any other "add REST writes" line)

(Zero Go source files. Zero TS/TSX source files. Zero test files. Zero `services/`, zero `web/`, zero `tests/`. The architecture file is READ, not edited.)

## Test contract
**No executable tests** (architecture §10.3 + US014 AC scenario "tester confirms no tests are required"). Tester's `US014_be_unit_tests.md`, `US014_fe_unit_tests.md`, `US014_e2e_tests.md` will each contain exactly one disclaimer line. No `.robot` file is created for US014.

Verification reduces to architecture §10.3's grep assertions:
```
grep -q "^## 9\. ADR-001 — MCP-only-writes is the permanent write API" docs/requirements/REQ006_tech_debt_backfill_sprint/architecture.md
grep -q "^### 9.3 Decision" docs/requirements/REQ006_tech_debt_backfill_sprint/architecture.md
grep -q "Conditions for revisiting" docs/requirements/REQ006_tech_debt_backfill_sprint/architecture.md
```
Plus a manual scenario-by-scenario walkthrough of the US014 AC against §9.

## Implementation notes
- **The ADR is already written.** `architecture.md` §9 (lines ~745–831) is the deliverable. The dev's job is to confirm it's complete vs. the US014 AC checklist — NOT to author or edit any ADR content.
- **If a gap is found:** raise `ARCHITECTURE_GAP_FOUND` with the specific AC scenario number + the §9 subsection it should map to but doesn't. Do NOT patch §9 yourself — the orchestrator routes back to system-architect for an architecture rev.
- **Strike-through format** (matches convention in `docs/tech_debt.md`):
  ```
  - <existing line content> → won't-fix per REQ006/US014 ADR-001 (MCP-only-writes is permanent)
  ```
  Verify with `git diff docs/tech_debt.md` that line 98 has only the trailing strike annotation appended; the original finding text stays for audit history.
- **Sweep for "add REST writes" lines** (US014 AC explicit): `git grep -nE 'REST.{0,10}(POST|PUT|DELETE|writes?|write API)' docs/tech_debt.md` to find any sibling. Each gets the same strike-through.
- **No `tests/e2e/` ADR file.** US014 AC scenario "tester confirms no tests are required" explicitly precludes a Robot Framework file.

## Definition of done
- `architecture.md` §9 verified against every AC scenario in `US014_adr_mcp_only_writes.md`. No gaps surfaced (or if gaps exist, `ARCHITECTURE_GAP_FOUND` raised and this task pauses).
- `docs/tech_debt.md` line 98 strike-through applied with `→ won't-fix per REQ006/US014 ADR-001 (MCP-only-writes is permanent)`.
- Any sibling "add REST writes" lines in `docs/tech_debt.md` similarly struck through with the same pointer.
- **The three architecture §10.3 grep assertions all pass.**
- **Review gate green:** `scripts/review/run-gate.sh cross` exits 0 with `REVIEW GATE: PASS`. (No BE-track or FE-track gate applies — no code touched.)
- **No `scripts/review/run-gate.sh be` / `scripts/review/run-gate.sh fe` requirement** — this task touches no Go and no TS/TSX. Tech-lead-side review consists of a re-read of architecture §9 + verification of the strike-throughs.
- **Live e2e NOT required** (architecture §10.3 explicit).
- **react-doctor NOT required** (architecture §10.3 explicit).
- Dev set status to `in_review`; tech-lead approved.

## Notes
- **Track field clarification.** `Track: BE` is set above because the project's task contract only allows `BE | FE` (per `.claude/agents/tech-lead.md` template). The REQ006 README's per-story track column refers to this as `BE-meta (docs)` — that maps to BE-track here, but the dev should treat this as a pure docs/verification task. No Go code is involved.
- **Dev type for orchestrator routing.** Orchestrator should spawn a `be-dev` (lower setup cost than `fe-dev` for a docs touch). The `be-dev` skill is irrelevant to the actual work — the dev's task is `Read` + `Edit` on `docs/tech_debt.md` plus a structured walkthrough of architecture §9 vs. US014 AC. Any be-dev who would otherwise want to write Go code SHOULD instead stop and re-read this `Notes` block.
- **Architecture is the deliverable.** If `architecture.md` §9 is complete and the strike-through is applied, the story is effectively done. Most of the work happened during Phase 1 architecture authoring; this Phase-3 task is the closing housekeeping.

## Review log
