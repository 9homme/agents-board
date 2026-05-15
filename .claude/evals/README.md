# evals — quality bar for vibe-commerce's skills and agents

Mirror of [`claude-skills/eval-workspace/`](../../../claude-skills/eval-workspace/) with an added agent track.
Symlinked into `.gemini/evals` so one set of files covers both platforms.

## Two layers

1. **Structural** (`tests/`) — pytest. Fast, deterministic, CI-friendly.
   - `test_skill_integrity.py` — every `SKILL.md` under `.claude/skills/` has frontmatter, an H1, references/scripts exist if cited, no empty files, names are unique.
   - `test_agent_integrity.py` — every `*.md` under `.claude/agents/` AND `.gemini/agents/` has frontmatter with `name` (= filename), `description`, `model`, `tools`; tools come from the per-platform allowlist; the two platforms stay in parity (same 6 agents, same descriptions, equivalent tool capability); models match the role matrix in `CLAUDE.md` (opus for po-ba / system-architect / tech-lead).

2. **Behavioral** (`iteration-N/`) — JSON fixtures hand-graded into `grading-results.md`. Same shape as `claude-skills`.
   - `skill_evals.json` — one or more prompts per skill, with `expected_output` written as graded assertions, not as a model answer.
   - `agent_evals.json` — scenarios per agent, each with `setup` (fixture state on disk), `prompt` (what the orchestrator would send), `expected_artifacts` (files the agent must create/modify), `expected_behavior` (rules from `CLAUDE.md`), and `anti_patterns` (must NOT happen — these come straight out of the CLAUDE.md anti-patterns list).
   - `grading-results.md` — manual scorecard per iteration. Verdict: PASS / PARTIAL / FAIL with notes.

## Run

```sh
pip install pytest pyyaml
pytest .claude/evals/tests/ -v
```

Behavioral evals are graded by hand. From a fresh session:
1. Pick a fixture in `iteration-N/agent_evals.json`. Materialize its `setup` block on disk under a throwaway worktree.
2. Spawn the named agent with the `prompt` field (mirroring how the orchestrator would).
3. Compare what was produced against `expected_artifacts` + `expected_behavior` + `anti_patterns`.
4. Record verdict + notes in `iteration-N/grading-results.md`.

## Iterating

Copy `iteration-N/` → `iteration-(N+1)/`, edit prompts (e.g. after an agent rewrite, or to harden a regression), rerun the manual grade. Keep prior iterations as history.

## What is intentionally NOT here

- No automated runner that spawns Claude Code via `-p`. claude-skills documented `-p` hangs on long system prompts, and our agents have long prompts. Add later only if reliability improves.
- No e2e of the full phase 1 → 3 pipeline — that belongs in `tests/e2e/` (Robot), not here. Agent evals are scoped to **one spawn at a time**.
