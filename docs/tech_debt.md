# Tech Debt Backlog

Durable backlog of non-blocking findings raised at code review or sign-off. Tech-lead appends one line per finding BEFORE issuing `approved` on a task (see `.claude/agents/tech-lead.md` §Verdict). Burying findings inside per-task review logs is explicitly disallowed — review logs scatter, this file does not.

**Format (one line per finding):**

```
- YYYY-MM-DD — <file:line> — <what's wrong> — <suggested fix> — REQ[ID]/US[ID]/<task-name>
```

Re-file fixed items by striking through, not by deleting:

```
- ~~2026-05-30 — web/package.json:14 — @testing-library/dom in dependencies instead of devDependencies — move it — REQ004/US003/fe_mermaid_diagram~~ → fixed in REQ005/US007
```

A retrospective at the end of each REQ scans this file and decides which items become a new tech-debt REQ.

---

## Findings

(none yet — first finding lands here)
