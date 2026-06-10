# US037/be_adr_verification_and_tech_debt_strikethrough

**Requirement:** REQ006
**Story:** US037
**Track:** BE
**Service:** services/agent-board   (nominal — this task touches NO Go code; meta/docs-only — see Notes)
**Status:** completed
**Blocked by:** none
**Worked-by:** be-dev-2026-06-06T00:00:00Z-a7f3
**Implements:** REQ006/US037 AC (all scenarios — architecture.md contains an explicit ADR section, decision stated, four read-only endpoints enumerated, five MCP tool families enumerated, rationale documented, alternatives considered, conditions-for-revisiting documented, tech-debt strike-through applied, no executable tests required), architecture §3 US037 touch row, architecture §9 (ADR-001 text — INLINE per D-008), architecture §10.3 (no TDG, no live e2e, no react-doctor; grep + manual read verification).

## Goal
**This is a documentation-verification + tech-debt-strike-through task.** The ADR text was authored by system-architect in `docs/requirements/REQ006_tech_debt_backfill_sprint/architecture.md` §9 (REQ006 rev 5, approved). The "dev" work for US037 reduces to:

1. **Verify** `architecture.md` §9 satisfies every AC scenario in `US037_adr_mcp_only_writes.md` — if any AC is missing, raise `ARCHITECTURE_GAP_FOUND` to route back to system-architect (do NOT patch the ADR yourself).
2. **Strike-through** `docs/tech_debt.md` line 98 with `→ won't-fix per REQ006/US037 ADR-001 (MCP-only-writes is permanent)`, and any other line that exists today and reflects "we should add REST writes" (US037 AC explicit).

No code, no tests, no react-doctor sweep, no live e2e (architecture §10.3).

## Scope
- **In:** Read `architecture.md` §9 (lines that hold ADR-001) and check it against each AC scenario in `US037_adr_mcp_only_writes.md`:
  - §9.1 Status (Accepted, effective date, author, approver).
  - §9.3 Decision (verbatim or close-paraphrase of "api-server is intentionally read-only…REST POST/PUT/DELETE will NOT be added unless a future REQ explicitly overrides").
  - §9.4 The four read-only `GET` endpoints enumerated.
  - §9.5 The MCP tool families enumerated.
  - §9.6 Rationale (≥ the 3 bullets from AC).
  - §9.7 Alternatives considered (≥ the 3 bullets from AC).
  - §9.8 Conditions for revisiting (≥ the 3 bullets from AC + the "not because it'd be easy" disclaimer).
- **In:** Edit `docs/tech_debt.md` to strike-through line 98 with `→ won't-fix per REQ006/US037 ADR-001 (MCP-only-writes is permanent)`. Sweep for any other tech-debt line that proposes "add REST writes" and strike-through with a pointer to the ADR.
- **Out:** Any edit to `architecture.md` itself — that's system-architect's domain. If a gap is found, raise `ARCHITECTURE_GAP_FOUND` and STOP. Any code, test, MSW handler, Robot file. Any change to `web/`, `services/`, `cmd/`. Any new ADR convention under `docs/adr/` — D-008 explicitly chose inline.

## Files touched (estimated, exclusive)
- `docs/tech_debt.md` (edit — strike-through line 98 + any other "add REST writes" line)

(Zero Go source files. Zero TS/TSX source files. Zero test files. Zero `services/`, zero `web/`, zero `tests/`. The architecture file is READ, not edited.)

## Test contract
**No executable tests** (architecture §10.3 + US037 AC scenario "tester confirms no tests are required"). Tester's `US037_be_unit_tests.md`, `US037_fe_unit_tests.md`, `US037_e2e_tests.md` will each contain exactly one disclaimer line. No `.robot` file is created for US037.

Verification reduces to architecture §10.3's grep assertions:
```
grep -q "^## 9\. ADR-001 — MCP-only-writes is the permanent write API" docs/requirements/REQ006_tech_debt_backfill_sprint/architecture.md
grep -q "^### 9.3 Decision" docs/requirements/REQ006_tech_debt_backfill_sprint/architecture.md
grep -q "Conditions for revisiting" docs/requirements/REQ006_tech_debt_backfill_sprint/architecture.md
```
Plus a manual scenario-by-scenario walkthrough of the US037 AC against §9.

## Implementation notes
- **The ADR is already written.** `architecture.md` §9 (lines ~745–831) is the deliverable. The dev's job is to confirm it's complete vs. the US037 AC checklist — NOT to author or edit any ADR content.
- **If a gap is found:** raise `ARCHITECTURE_GAP_FOUND` with the specific AC scenario number + the §9 subsection it should map to but doesn't. Do NOT patch §9 yourself — the orchestrator routes back to system-architect for an architecture rev.
- **Strike-through format** (matches convention in `docs/tech_debt.md`):
  ```
  - <existing line content> → won't-fix per REQ006/US037 ADR-001 (MCP-only-writes is permanent)
  ```
  Verify with `git diff docs/tech_debt.md` that line 98 has only the trailing strike annotation appended; the original finding text stays for audit history.
- **Sweep for "add REST writes" lines** (US037 AC explicit): `git grep -nE 'REST.{0,10}(POST|PUT|DELETE|writes?|write API)' docs/tech_debt.md` to find any sibling. Each gets the same strike-through.
- **No `tests/e2e/` ADR file.** US037 AC scenario "tester confirms no tests are required" explicitly precludes a Robot Framework file.

## Definition of done
- `architecture.md` §9 verified against every AC scenario in `US037_adr_mcp_only_writes.md`. No gaps surfaced (or if gaps exist, `ARCHITECTURE_GAP_FOUND` raised and this task pauses).
- `docs/tech_debt.md` line 98 strike-through applied with `→ won't-fix per REQ006/US037 ADR-001 (MCP-only-writes is permanent)`.
- Any sibling "add REST writes" lines in `docs/tech_debt.md` similarly struck through with the same pointer.
- **The three architecture §10.3 grep assertions all pass.**
- **Review gate green:** `scripts/review/run-gate.sh cross` exits 0 with `REVIEW GATE: PASS`. (No BE-track or FE-track gate applies — no code touched.)
- **No `scripts/review/run-gate.sh be` / `scripts/review/run-gate.sh fe` requirement** — this task touches no Go and no TS/TSX. Tech-lead-side review consists of a re-read of architecture §9 + verification of the strike-throughs.
- **Live e2e NOT required** (architecture §10.3 explicit).
- **react-doctor NOT required** (architecture §10.3 explicit).
- Dev set status to `in_review`; tech-lead approved.

## Notes
- **Track field clarification.** `Track: BE` is set above because the project's task contract only allows `BE | FE` (per `.claude/agents/tech-lead.md` template). The REQ006 README's per-story track column refers to this as `BE-meta (docs)` — that maps to BE-track here, but the dev should treat this as a pure docs/verification task. No Go code is involved.
- **Dev type for orchestrator routing.** Orchestrator should spawn a `be-dev` (lower setup cost than `fe-dev` for a docs touch). The `be-dev` skill is irrelevant to the actual work — the dev's task is `Read` + `Edit` on `docs/tech_debt.md` plus a structured walkthrough of architecture §9 vs. US037 AC. Any be-dev who would otherwise want to write Go code SHOULD instead stop and re-read this `Notes` block.
- **Architecture is the deliverable.** If `architecture.md` §9 is complete and the strike-through is applied, the story is effectively done. Most of the work happened during Phase 1 architecture authoring; this Phase-3 task is the closing housekeeping.

## Notes

**Files touched:**
- `docs/tech_debt.md` — line 98 struck through with `→ won't-fix per REQ006/US037 ADR-001 (MCP-only-writes is permanent)`. Line 91 ("Suggested fold-into-REQ006: REST surface") also struck through as a sibling "add REST writes" proposal.

**AC verification results (architecture §9 vs US037_adr_mcp_only_writes.md scenarios):**
- Scenario 1 (explicit ADR section): PASS — `## 9. ADR-001 — MCP-only-writes is the permanent write API` at line 746; referenced in executive summary (line 20).
- Scenario 2 (Decision subsection — verbatim): PASS — §9.3 contains the exact required text.
- Scenario 3 (four GET endpoints enumerated): PASS — §9.4 lists all four `/api/v1/` GET endpoints.
- Scenario 4 (MCP tool families enumerated): PASS — §9.5 lists all five families including RegisterAuditTools.
- Scenario 5 (Rationale ≥3 bullets): PASS — §9.6 has 4 bullets covering all 3 required plus composability.
- Scenario 6 (Alternatives ≥3 bullets): PASS — §9.7 has 4 entries covering all 3 required plus GraphQL.
- Scenario 7 (Conditions for revisiting ≥3 + disclaimer): PASS — §9.8 has all three conditions plus "NOT revisited just because adding a REST endpoint would be technically easy" verbatim.
- Scenario 8 (tech-debt strike-throughs): PASS — applied in this task.
- Scenario 9 (no tests required): N/A (tester's domain, confirmed in US037_be_unit_tests.md).

**Three §10.3 grep assertions:** all PASS.

**Review gate:** `scripts/review/run-gate.sh cross` → `REVIEW GATE: PASS` (semgrep OWASP/Go/TS + gitleaks both clean).

**No executable tests.** Documentation-verification + strike-through only, as per architecture §10.3.

**Live e2e:** NOT required (architecture §10.3 explicit).

## Review log

### Review pass 1 — 2026-06-07 — verdict: approved

**Track:** BE (docs-only / meta-verification — NO Go code, NO web/, NO tests, NO e2e per architecture §10.3). Standard BE/FE/coverage/live-e2e/react-doctor DoD gates do not apply; only the cross review gate + manual ADR-vs-AC walkthrough apply.

**Architecture §9 vs US037 AC walkthrough (all 7 verifiable scenarios + strike-through):**
- Scenario 1 (explicit ADR section): PASS — `## 9. ADR-001 — MCP-only-writes is the permanent write API` at architecture.md:746.
- Scenario 2 (Decision verbatim): PASS — §9.3 (architecture.md:768) contains the required "api-server is intentionally read-only … REST POST/PUT/DELETE will NOT be added unless a future requirement explicitly overrides this ADR" verbatim.
- Scenario 3 (four read-only GET endpoints): PASS — §9.4 (architecture.md:774–777) enumerates all four `/api/v1/` GET endpoints.
- Scenario 4 (MCP tool families): PASS — §9.5 (architecture.md:785–789) enumerates all five families (Project/Document/UserStory/Task + RegisterAuditTools).
- Scenario 5 (Rationale ≥3 bullets): PASS — §9.6 (architecture.md:793–798) has 4 bullets covering all 3 required.
- Scenario 6 (Alternatives ≥3 bullets): PASS — §9.7 (architecture.md:800–805) has 4 entries covering all 3 required.
- Scenario 7 (Conditions ≥3 + disclaimer): PASS — §9.8 (architecture.md:807–815) has 3 conditions plus the verbatim "This decision is NOT revisited just because adding a REST endpoint would be technically easy" disclaimer at architecture.md:815.
- Scenario 8 (tech-debt strike-throughs): PASS — `docs/tech_debt.md:98` (primary "add REST POST/PUT/DELETE endpoints" finding) struck through with `→ won't-fix per REQ006/US037 ADR-001 (MCP-only-writes is permanent)`; sibling `docs/tech_debt.md:91` ("...adds the missing REST endpoints...") also struck through with the same pointer. No other "add REST writes" line remains unstruck (verified by grep sweep).

**Three §10.3 grep assertions:** all PASS (grep1 §9 heading, grep2 §9.3 Decision, grep3 "Conditions for revisiting").

**Scope / boundary checks:**
- architecture.md NOT modified by the dev (system-architect owns it) — confirmed: US037 dev commit `9e43840` touched ONLY the task file + `docs/tech_debt.md` (2 lines). VERIFIED.
- No Go code, no `*_test.go`, no `web/`, no `*.tsx`, no `tests/`/`.robot` changes. VERIFIED.
- TDG: not applicable — pure docs-verification task, no red/green/refactor cycles (architecture §10.3 waives tests). Dev commits carry `(US037)` traceability tags.

**Review gate (mandatory):**
```
== Cross-cutting · repo ==
  PASS  semgrep (owasp/golang/typescript)
  PASS  gitleaks (no secrets)

REVIEW GATE: PASS
```
(`scripts/review/run-gate.sh cross` exit 0. No BE/FE-track gate applies — no code touched. No coverage gate — no production source files. No robot --dryrun — no e2e suite for US037 by design. No live-e2e — architecture §10.3 explicit.)

**Tech-debt:** none filed this pass — the only findings (lines 91, 98) were the won't-fix strike-throughs that ARE the deliverable of this task.

**Verdict:** approved. Status → completed.
