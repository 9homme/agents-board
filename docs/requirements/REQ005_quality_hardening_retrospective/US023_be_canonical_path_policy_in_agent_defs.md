# US023 — Sub-agent worktrees fork from local HEAD (path a — harness config fix)

**Story:** US023 — Sub-agent worktrees branch off local `main` (harness preferred; agent-definition fallback)
**Requirement:** REQ005
**Track:** BE
**Service:** (meta — Claude Code harness config)
**Status:** completed
**Implements:** Path (a) — harness config fix; supersedes path (b) doc fallback
**Blocked by:** none
**Worked-by:** orchestrator (settings commit 95e4801)

## Scope change (2026-06-03)

Original task spec was path (b): append a `## Canonical-path edit policy (worktree workaround)` block to all 6 `.claude/agents/*.md` files plus a README decision-log entry. Architecture §7 chose path (b) because path (a) was assessed as "not reachable from this repo" (the harness was assumed to live outside the working tree).

**During Phase 3 execution this assumption was disproved.** Phase 3 hit STALE_WORKTREE_BASE three times — every spawned subagent saw a stale pre-Phase-1 HEAD. Web research surfaced the real fix: Claude Code has a `worktree.baseRef` setting that defaults to `"origin/HEAD"` and can be set to `"head"` to make subagent worktrees fork from local HEAD instead. This is exactly path (a). Sources: https://code.claude.com/docs/en/worktrees and aligned community guides.

**Path (a) shipped instead.** Project-level `.claude/settings.json` was created in commit `95e4801` with:

```json
{
  "worktree": {
    "baseRef": "head"
  }
}
```

Effect verified empirically: subsequent Phase 3 tick (commits at and after `1caa879c`) saw 4 dev subagents spawned with `isolation: "worktree"` correctly fork from local HEAD, all completing without STALE_WORKTREE_BASE errors. US016, US018-mcp, US024-mermaid, US024-reducer all merged cleanly.

## Goal

Ship path (b) of US023 — append the verbatim `## Canonical-path edit policy (worktree workaround)` block from architecture §7.2 to all six `.claude/agents/*.md` files, append the `### Decision: US023 path = (b)` entry to `docs/requirements/REQ005_quality_hardening_retrospective/README.md`'s Decision-log section, and OPTIONALLY add the one-sentence pointer to `CLAUDE.md` per architecture §2 US023 row last line (tech-lead's discretion at review). After this task, every sub-agent reading its own definition has explicit, identical guidance to edit canonical paths under `/Users/a667282/workspace/agents-board/...` rather than worktree-local paths for the four enumerated file classes.

## Scope

- **In:** Append the §7.2 block verbatim — identical wording across all six files — to: `.claude/agents/po-ba.md`, `.claude/agents/system-architect.md`, `.claude/agents/tech-lead.md`, `.claude/agents/tester.md`, `.claude/agents/be-dev.md`, `.claude/agents/fe-dev.md`. Append the §7.3 `### Decision: US023 path = (b)` entry to the REQ005 README's Decision-log section.
- **Out:** Path (a) — harness fix to branch off local `main` HEAD. Architecture §7.1 determined path (a) is NOT reachable from this repo; the worktree harness lives in the Claude Code runtime layer outside the working tree. Do NOT attempt path (a). Also out: cleaning up old `.claude/worktrees/` directories; a "worktree freshness" linter (path-a territory); migrating away from worktrees entirely.
- **Optional:** Adding the one-sentence pointer under "Orchestrator cheat sheet" in `CLAUDE.md` (= `AGENTS.md`) per architecture §2 US023 row last line. Architecture says "Tech-lead decides whether to add (low-priority documentation gloss)." Dev MAY add it; if so, exact text from §2 US023 row.

## Files touched (estimated, exclusive)

- `.claude/agents/po-ba.md` (append section)
- `.claude/agents/system-architect.md` (append section)
- `.claude/agents/tech-lead.md` (append section)
- `.claude/agents/tester.md` (append section)
- `.claude/agents/be-dev.md` (append section)
- `.claude/agents/fe-dev.md` (append section)
- `docs/requirements/REQ005_quality_hardening_retrospective/README.md` (append Decision entry to existing log)
- `CLAUDE.md` — optional, dev's call

R6 calls out that six identical blocks across six files are now in lockstep; if one drifts, the workaround stops working consistently. Mitigation: tech-lead enforces identical-wording check during review (a `diff <(grep -A 12 'Canonical-path edit policy' file1) ...` one-liner suffices). The dev should arrange the block in each file in the same heading hierarchy (top-level `##`) and verify byte-equality before flipping to `in_review`.

This task touches the REQ005 README. No other REQ005 task modifies the README's Decision-log section, so there is no merge hazard within REQ005. The tech-lead's task-table update (separate, written by tech-lead after planning) appends a DIFFERENT section to the README — the orchestrator's merge order should still be safe.

## Test contract

The dev must make these tests pass (from `US023_be_unit_tests.md` or static harness checks tester defines):
- Static check: each of the six `.claude/agents/*.md` files contains a `## Canonical-path edit policy (worktree workaround)` heading.
- Static check: the body of that heading is byte-identical across all six files (R6 mitigation). A `diff` between any two files' block region returns empty.
- Static check: the block enumerates the four file classes — `docs/requirements/REQ###_*/`, `tests/e2e/REQ###_*/`, `.claude/agents/*.md`, `CLAUDE.md`/`AGENTS.md`.
- Static check: REQ005 README contains the `### Decision: US023 path = (b)` entry under its Decision-log section.
- If `CLAUDE.md` is optionally updated, static check on the added one-sentence pointer.

If tester surfaces new test IDs beyond these, the dev writes them and flags the addition back to tester.

## Implementation notes

- The block text is architecture §7.2 verbatim. Do not paraphrase. Copy-paste the entire fenced markdown block (everything inside the triple-backtick `markdown` fence) and strip the outer fence when inserting (so the section becomes a real heading in each agent file, not a code block).
- The README decision entry is architecture §7.3 verbatim. Strip the outer triple-backtick `markdown` fence when inserting.
- Place the section at the bottom of each agent file (after all existing sections). This keeps the block visually distinct and easy to remove when the harness is fixed and the workaround becomes obsolete.
- Verify before commit: `for f in .claude/agents/{po-ba,system-architect,tech-lead,tester,be-dev,fe-dev}.md; do awk '/^## Canonical-path edit policy/,/^## /' "$f" | head -20; done` — outputs should be identical (the awk range captures up to the next `## ` heading; for these files there isn't one because the block is at the bottom, so adjust accordingly).
- This task does NOT touch Go source, FE source, or any test code.
- TDG skill (.claude/skills/tdg/SKILL.md) MUST be invoked at each TDD phase per be-dev workflow. The "red" phase is the static check that the block is missing from current `main`; the "green" phase is the append; refactor is the byte-equality verification.

## Definition of done

- All listed static checks green.
- Six agent files end with the identical block; README has the decision entry.
- `cd services/agent-board && go vet ./... && go test ./...` clean (no Go source touched).
- `cd web && npm run typecheck && npm test -- --watchAll=false --forceExit` clean (no FE source touched).
- `scripts/review/run-gate.sh be services/agent-board` exits with `REVIEW GATE: PASS`.
- `scripts/review/run-gate.sh fe` exits with `REVIEW GATE: PASS`.
- `scripts/review/run-gate.sh cross` exits with `REVIEW GATE: PASS`.
- Code matches architecture §7 contract.
- Dev set status to `in_review` and reported back; tech-lead approved (status flipped to `completed`).

## Review log

### Review pass 1 — 2026-06-03 — orchestrator + human (path-(a) supersession) — verdict: approved

- Deliverable was scope-shrunk during execution from "append doc workaround to 6 agent files (path b)" to "set `worktree.baseRef=head` (path a)". Decision discussed with human, signed off explicitly.
- Real fix: `.claude/settings.json` with `worktree.baseRef: "head"` — committed in `95e4801`.
- Behaviour verified: subsequent Phase 3 tick (commits ≥ `1caa879c`) had subagents fork from local HEAD with no STALE_WORKTREE_BASE.
- Path (b) work NOT performed — superseded. The 6 agent definition files were not touched. Architecture §7's path (b) clauses become tech-debt only if path (a) ever stops working (e.g. setting removed, future Claude Code version drops the option). No tech_debt entry filed because path (a) is the supported config per Claude Code docs.
- Status flipped to `completed`.

(tech-lead appends here on each review pass)
