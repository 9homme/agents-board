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

- 2026-06-02 — scripts/review/test/test_run_gate.sh:UT-001 — UT-001 only asserts banner presence, missing body-text + rc value visibility — tighten assertion to also check captured body string and `rc=1` per IT-US001-001 — REQ005/US001/be_fix_printf_double_dash
- 2026-06-02 — scripts/review/test/test_run_gate.sh:UT-003 — UT-003 symmetric tightening required for IT-US001-002 (same as above for `run_check_warn`) — REQ005/US001/be_fix_printf_double_dash
- 2026-06-02 — scripts/review/test/test_run_gate.sh:* — Test-ID naming uses UT-NNN instead of canonical IT-US001-NNN — re-map naming on next touch — REQ005/US001/be_fix_printf_double_dash
- 2026-06-02 — docs/requirements/REQ005_*/US001_be_unit_tests.md:IT-US001-004 — TTY-mode test remains manual (spec allows it; promote to automated when a portable TTY harness is added) — REQ005/US001/be_fix_printf_double_dash
- 2026-06-02 — docs/requirements/REQ005_*/US001_be_fix_printf_double_dash.md:DoD — BE-gate precondition implicitly couples to US003 (gosec/govulncheck soft-warn); add explicit `Depends-on: US003` or rephrase DoD — REQ005/US001/be_fix_printf_double_dash
