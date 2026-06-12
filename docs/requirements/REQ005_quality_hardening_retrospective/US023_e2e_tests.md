# US023 — E2E test specification

**Owner:** tester.

## Why e2e does not apply

US023 ships path (b) — appending the canonical-path edit policy block to six agent definition files (`.claude/agents/*.md`) and optionally to `CLAUDE.md`. This is a documentation-only change:

- There is no web service surface, no HTTP endpoint, no browser interaction.
- The correctness guarantee is purely structural: the six files contain the policy block with identical wording (UT-US023-001 through UT-US023-007 in `US023_be_unit_tests.md`).
- The docker-compose stack (US022) is not involved.
- The Robot Framework's Browser and RequestsLibrary have no applicable surface.
- The worktree behaviour US023 fixes (add/add merge conflicts) is not observable via e2e test — it is an orchestrator-level operational concern.

**Verdict: No e2e scenarios. Documentation-only change; covered by UT-US023-001 through UT-US023-007.**
