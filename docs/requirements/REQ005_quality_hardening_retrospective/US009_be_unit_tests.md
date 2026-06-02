# US009 — Backend unit & integration test specification

**For BE Dev:** US009 ships path (b) — appending the canonical-path edit policy block to six agent definition files. Tests are grep-style structural assertions verifying the block is present and has consistent wording across all six files.

No Go packages are touched. No Dockerfiles, no compose files. These are documentation-only file edits plus optional CLAUDE.md pointer.

## Coverage matrix

| AC scenario | Layer | Test ID | File | Behaviour under test |
|---|---|---|---|---|
| `po-ba.md` contains canonical-path policy block | unit | UT-US009-001 | `.claude/agents/po-ba.md` | required heading and body text present |
| `system-architect.md` contains canonical-path policy block | unit | UT-US009-002 | `.claude/agents/system-architect.md` | same |
| `tech-lead.md` contains canonical-path policy block | unit | UT-US009-003 | `.claude/agents/tech-lead.md` | same |
| `tester.md` contains canonical-path policy block | unit | UT-US009-004 | `.claude/agents/tester.md` | same |
| `be-dev.md` contains canonical-path policy block | unit | UT-US009-005 | `.claude/agents/be-dev.md` | same |
| `fe-dev.md` contains canonical-path policy block | unit | UT-US009-006 | `.claude/agents/fe-dev.md` | same |
| All six blocks have identical wording | unit | UT-US009-007 | all six agent files | diff of the extracted blocks is empty |

## Unit tests

These are shell-harness or script assertions (bash scripts in `scripts/review/test/` or as standalone verifiers). They may also be implemented as Jest structural tests reading the filesystem — the implementation choice is the dev's.

### UT-US009-001 through UT-US009-006 — Each agent file contains the policy block

For each of the six files (`po-ba.md`, `system-architect.md`, `tech-lead.md`, `tester.md`, `be-dev.md`, `fe-dev.md`):

- **File under test:** `.claude/agents/<agent-name>.md`
- **When:** the file is read
- **Then:**
  - The file contains the exact heading `## Canonical-path edit policy (worktree workaround)` (verbatim, per architecture §7.2).
  - The file contains the sentence `ALWAYS edit the following file classes at their canonical path under `/Users/a667282/workspace/agents-board/`` (verbatim substring match).
  - The file contains the four bullet-point paths: `docs/requirements/REQ###_*/`, `tests/e2e/REQ###_*/`, `.claude/agents/*.md`, and `CLAUDE.md` / `AGENTS.md`.
  - The file contains the instruction `Do NOT \`git add\` the worktree-local copies of these files`.
- **Architecture cite:** architecture §7.2 — canonical-path edit policy block (exact wording given there).

### UT-US009-007 — All six blocks are word-for-word identical

- **Files under test:** all six agent definition files
- **When:** the `## Canonical-path edit policy (worktree workaround)` section is extracted from each file (from the heading to the next `##` heading or end-of-file)
- **Then:**
  - `diff` of any two extracted blocks is empty (byte-for-byte identical content after the heading line).
- **Notes:** Identical wording is intentional per architecture §7.2 — "uniformity ... is intentional and is enforceable by a future lint check." This test IS that lint check for the initial landing.
- **Architecture cite:** architecture §7.2 — "Identical wording across files (a copy-paste-able block)".
