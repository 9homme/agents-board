# US009 — E2E test specification

**Owner:** tester.

## Why e2e does not apply

US009 ships path (b) — appending the canonical-path edit policy block to six agent definition files (`.claude/agents/*.md`) and optionally to `CLAUDE.md`. This is a documentation-only change:

- There is no web service surface, no HTTP endpoint, no browser interaction.
- The correctness guarantee is purely structural: the six files contain the policy block with identical wording (UT-US009-001 through UT-US009-007 in `US009_be_unit_tests.md`).
- The docker-compose stack (US008) is not involved.
- The Robot Framework's Browser and RequestsLibrary have no applicable surface.
- The worktree behaviour US009 fixes (add/add merge conflicts) is not observable via e2e test — it is an orchestrator-level operational concern.

**Verdict: No e2e scenarios. Documentation-only change; covered by UT-US009-001 through UT-US009-007.**
