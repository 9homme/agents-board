# US009 — Append canonical-path edit policy block to all six agent definitions (path b)

**Story:** US009 — Sub-agent worktrees branch off local `main` (harness preferred; agent-definition fallback)
**Requirement:** REQ005
**Track:** BE
**Service:** (meta — agent definitions; no service binary)
**Status:** pending
**Implements:** Path (b) — agent-definition fallback: Scenario: every agent definition file documents the canonical-path workaround, Scenario: orchestrator-side guidance is documented, Scenario: no regression on a non-worktree workflow, Scenario: README captures the chosen path and the reason
**Blocked by:** none
**Worked-by:** _(none)_

## Goal

Ship path (b) of US009 — append the verbatim `## Canonical-path edit policy (worktree workaround)` block from architecture §7.2 to all six `.claude/agents/*.md` files, append the `### Decision: US009 path = (b)` entry to `docs/requirements/REQ005_quality_hardening_retrospective/README.md`'s Decision-log section, and OPTIONALLY add the one-sentence pointer to `CLAUDE.md` per architecture §2 US009 row last line (tech-lead's discretion at review). After this task, every sub-agent reading its own definition has explicit, identical guidance to edit canonical paths under `/Users/a667282/workspace/agents-board/...` rather than worktree-local paths for the four enumerated file classes.

## Scope

- **In:** Append the §7.2 block verbatim — identical wording across all six files — to: `.claude/agents/po-ba.md`, `.claude/agents/system-architect.md`, `.claude/agents/tech-lead.md`, `.claude/agents/tester.md`, `.claude/agents/be-dev.md`, `.claude/agents/fe-dev.md`. Append the §7.3 `### Decision: US009 path = (b)` entry to the REQ005 README's Decision-log section.
- **Out:** Path (a) — harness fix to branch off local `main` HEAD. Architecture §7.1 determined path (a) is NOT reachable from this repo; the worktree harness lives in the Claude Code runtime layer outside the working tree. Do NOT attempt path (a). Also out: cleaning up old `.claude/worktrees/` directories; a "worktree freshness" linter (path-a territory); migrating away from worktrees entirely.
- **Optional:** Adding the one-sentence pointer under "Orchestrator cheat sheet" in `CLAUDE.md` (= `AGENTS.md`) per architecture §2 US009 row last line. Architecture says "Tech-lead decides whether to add (low-priority documentation gloss)." Dev MAY add it; if so, exact text from §2 US009 row.

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

The dev must make these tests pass (from `US009_be_unit_tests.md` or static harness checks tester defines):
- Static check: each of the six `.claude/agents/*.md` files contains a `## Canonical-path edit policy (worktree workaround)` heading.
- Static check: the body of that heading is byte-identical across all six files (R6 mitigation). A `diff` between any two files' block region returns empty.
- Static check: the block enumerates the four file classes — `docs/requirements/REQ###_*/`, `tests/e2e/REQ###_*/`, `.claude/agents/*.md`, `CLAUDE.md`/`AGENTS.md`.
- Static check: REQ005 README contains the `### Decision: US009 path = (b)` entry under its Decision-log section.
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

(tech-lead appends here on each review pass)
