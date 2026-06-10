# US023 — Sub-agent worktrees branch off local `main` (harness preferred; agent-definition fallback)

**Requirement:** REQ005 — quality hardening retrospective
**Status:** draft

## Story
As an **orchestrator (and as a reviewer in a worktree)**, I want **sub-agent worktrees to branch from the local `main` HEAD** (not from `origin/main` which is stale by N commits during an active REQ), so that worktree edits don't generate add/add merge conflicts on docs/spec files that already exist on local `main`, and I no longer have to resolve them manually with `git checkout --theirs` or instruct agents to edit canonical paths instead of worktree paths.

## Acceptance criteria

This story has **two acceptance paths**. Either one satisfies the AC. **Path (a) is preferred** if the harness layer is reachable from this repo.

### Path (a) — harness fix (preferred)

- **Scenario: new worktree branches off local `main`**
  - Given an active REQ where local `main` is N commits ahead of `origin/main` (N > 0)
  - When the orchestrator (or harness) creates a new sub-agent worktree (e.g. `git worktree add .claude/worktrees/agent-<id> -b agent-<id>-branch`)
  - Then the new branch is created off `refs/heads/main` (local HEAD)
  - And NOT off `refs/remotes/origin/main`
  - And the worktree's working tree contains all files that local `main` has, including any docs/specs/tasks created by previous sub-agent runs in this REQ that have been merged to local main

- **Scenario: previously-stuck add/add conflict no longer reproduces**
  - Given the scenario that prompted this story: REQ004 sign-off, worktree `worktree-agent-ab9da34ff39564aa2` branched off stale `origin/main`, agent tried to create `docs/requirements/REQ004_*/REQ004_quality_audit.md` that already existed on local main, merge into main produced add/add conflict
  - When the new harness behaviour is in place
  - Then a worktree created in the same scenario branches off the up-to-date local main
  - And the file already exists in the worktree
  - And the agent edits it in place (no add/add at merge time)

- **Scenario: worktree creation is deterministic and visible**
  - Given a worktree is created via the harness
  - When the orchestrator runs `git -C <worktree-path> log -1 --oneline`
  - Then the HEAD commit matches local `main`'s HEAD at the time of worktree creation
  - And `git -C <worktree-path> rev-parse --abbrev-ref HEAD` shows the agent-specific branch name

### Path (b) — agent-definition fallback (if harness is not reachable)

- **Scenario: every agent definition file documents the canonical-path workaround**
  - Given the harness cannot be modified from this repo (architect / tech-lead determines this is the case)
  - When the story is complete via path (b)
  - Then each of the following files contains an explicit "canonical-path edit policy" section explaining: (i) why the workaround exists (worktree branches off stale `origin/main`), (ii) when to use it (edits to `docs/requirements/REQ###_*/`, `tests/e2e/REQ###_*/`, `.claude/agents/*.md`, and any other shared spec-style file), (iii) the exact path to edit (always under `/Users/a667282/workspace/agents-board/...`, NEVER the worktree-local path), (iv) what NOT to do (do not `git add` the worktree-local copy of those files):
    - `.claude/agents/po-ba.md`
    - `.claude/agents/system-architect.md`
    - `.claude/agents/tech-lead.md`
    - `.claude/agents/tester.md`
    - `.claude/agents/be-dev.md`
    - `.claude/agents/fe-dev.md`
  - And the wording is consistent across all six files (a copy-paste-able block).

- **Scenario: orchestrator-side guidance is documented**
  - Given path (b) is the chosen route
  - When `CLAUDE.md` is checked
  - Then it either (i) already contains the orchestrator-side guidance (the "edit canonical paths" expectation is implicit in current routing) — in which case nothing more is needed, OR (ii) gains a short subsection under the "Orchestrator cheat sheet" naming the canonical-path workaround as the expected pattern until path (a) is feasible.

### Common to both paths

- **Scenario: no regression on a non-worktree workflow**
  - Given a developer working directly in `/Users/a667282/workspace/agents-board/` (no worktree)
  - When they run normal `git` commands
  - Then nothing about their workflow changes

- **Scenario: README captures the chosen path and the reason**
  - Given REQ005 README's decision log
  - When the story is complete
  - Then a `### Decision: US023 path` entry is appended to the README (or to a `docs/operations/harness.md` file) naming which path (a or b) shipped, and why the other was infeasible if applicable

## UI / UX flow expectations

**No UI:** orchestrator / agent harness change. For completeness:

- **Entry points:** orchestrator spawns a sub-agent (`Agent` tool); the harness creates a worktree behind the scenes.
- **Happy-path flow (path a):** orchestrator → harness → `git worktree add` off local main → agent works in worktree → diffs merge cleanly into main.
- **Happy-path flow (path b):** orchestrator → harness → worktree off stale origin/main → agent reads its `.md` definition → notices the canonical-path policy → edits `/Users/.../docs/requirements/REQ###_*/...` directly → merge-conflict-free because the file is touched only once.
- **Out of UI scope:** any visual indicator of worktree freshness. CLI + git status is enough.

## Out of scope
- **Migrating away from worktrees entirely.** Worktrees are the chosen isolation mechanism; this story makes them work, not replace them.
- **A separate "worktree linter" CI check.** Nice-to-have, but new scope.
- **Touching every existing worktree under `.claude/worktrees/`.** Only new worktrees benefit from path (a); old worktrees can be left alone or cleaned up separately.
- **Documenting the workaround in `CLAUDE.md`'s top-level anti-patterns list** unless path (b) is the only feasible route AND the team wants to make it more discoverable.
- **Solving every conceivable merge-conflict scenario.** This story fixes the specific add/add-on-spec-files class. Other conflict classes (legitimate concurrent edits to the same line of source code) are handled by tech-lead review and are not in scope.

## Dependencies
- None within REQ005. May coordinate with anyone who's currently maintaining the worktree harness (architect can identify in their phase).

## Notes for the team

- **Architect / tech-lead must determine path-feasibility first.** Path (a) requires reaching whatever script or hook creates the worktree (likely a shell wrapper around `git worktree add` somewhere outside this repo, possibly in `~/.claude/` or a Claude Code internal). If reachable, path (a). If not, path (b).
- **Path (a) one-liner** (illustrative): replace `git worktree add <path> -b <branch>` with `git worktree add <path> -b <branch> main` (the third positional arg is the start point — defaults to HEAD, which is what we want, but the orchestrator's harness may currently be passing `origin/main` explicitly).
- **Path (b) wording block** (so all six agent files say the same thing — tech-lead can refine):
  ```
  ## Canonical-path edit policy (worktree workaround)

  Sub-agent worktrees currently branch off `origin/main`, which may be stale by N commits
  during an active REQ. To avoid add/add merge conflicts on spec / docs / agent-definition
  files that already exist on local `main`, ALWAYS edit the following file classes at their
  canonical path under `/Users/a667282/workspace/agents-board/`, never at the worktree-local
  path:
    - docs/requirements/REQ###_*/
    - tests/e2e/REQ###_*/
    - .claude/agents/*.md
    - CLAUDE.md
  Do NOT `git add` the worktree-local copies of these files. The orchestrator runs `git status`
  and rejects worktrees that touch the local copy of a docs/spec/agent file.
  ```
- **Evidence the bug is real:** orchestrator log entries during REQ004 sign-off documented multiple `git checkout --theirs` resolutions on `docs/requirements/REQ004_*/REQ004_quality_audit.md` and adjacent spec files — pure friction, no design value.
- **Smallest viable path (b) shipping:** add the wording block to the six agent files, append the decision entry to REQ005 README. ≤2 points of work.
- **Smallest viable path (a) shipping:** find the worktree-creation site (likely outside this repo), change the start point, verify with one fresh agent spawn during a multi-commit REQ. ≤2 points if the harness is reachable; ∞ if it's not.

## Sign-off log
(po-ba appends here on each sign-off pass)
