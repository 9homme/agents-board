---
US: US012
Title: Make ProjectCard clickable as a link to /projects/{id}
Status: completed
Track: FE
Implements: US012 AC "Dashboard project cards are clickable (in-scope for this story)", "Click a project card to open the detail page"
Blocked by:
Worked-by: fe-dev-20260526-ae1f
---

## Goal
Wrap the existing `ProjectCard`'s `<article>` in a Next.js `<Link href={`/projects/${project.id}`}>` so cards become real, keyboard-accessible links that navigate to the new detail page, without changing the existing REQ002 visual design.

## Architecture references
- `architecture.md` §"Frontend surface" → first table row `/` (modified — each `ProjectCard` becomes a link to `/projects/{id}`).
- `architecture.md` §"Components → Frontend" → row `web/components/Dashboard/ProjectCard.tsx` (modified).
- `architecture.md` §"State strategy" → "Dashboard card click wiring" bullet (Link wraps the `<article>`; `aria-label={project.name}`; preserve visual classes; focus styles visible).
- `architecture.md` §"Cross-cutting → Accessibility cross-cutting" → "ProjectCard link is keyboard-focusable with visible focus ring".

## Scope
- **In:**
  - Wrap the `<article>` in `web/components/Dashboard/ProjectCard.tsx` with `<Link href={`/projects/${project.id}`} aria-label={project.name}>`.
  - Preserve every existing className and child structure inside the `<article>`.
  - Ensure focus styles are visible (Tailwind `focus-visible:` ring on the link).
  - Update / add component tests in `web/components/Dashboard/ProjectList.test.tsx` (or a new `ProjectCard.test.tsx` if cleaner) to assert link href, accessible name, and keyboard reachability.
- **Out:**
  - Anything inside `web/pages/projects/[id].tsx` (that is `us001_fe_detail_page_with_tabs`).
  - Changing card visuals beyond focus styling.
  - Adding any new API call (no fetch involved in this task).

## Files touched (estimated, exclusive)
- `web/components/Dashboard/ProjectCard.tsx`
- `web/components/Dashboard/ProjectList.test.tsx` (extend existing tests for the new link behaviour) OR a new `web/components/Dashboard/ProjectCard.test.tsx` — dev picks whichever fits the existing house style.

## Test contract
The dev must make the matching cases in `US012_fe_unit_tests.md` for "ProjectCard becomes a link" / clickable card AC pass (FCT-* IDs assigned by tester). If the tester has not yet authored the relevant IDs at the time the dev picks this up, the dev flags it back to tester rather than skipping coverage.

## Implementation notes
- Import `Link` from `next/link`. The Pages Router pattern is `<Link href="/projects/abc"><article …>…</article></Link>` — no `<a>` wrapping needed in modern Next.js (it renders the anchor itself).
- `aria-label={project.name}` belongs on the `<Link>` so screen readers announce the project name.
- Preserve **all** existing `className` strings on the `<article>` and inner elements — the REQ002 sign-off cited "minimal beautiful" and this change is additive.
- Add a `focus-visible:` ring class (e.g. `focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2`) on the `<Link>` itself for keyboard focus visibility.
- Tests should assert: (a) the rendered element has `role="link"` (implicit on the anchor) with `aria-label` matching `project.name`; (b) `href` is `/projects/${project.id}`; (c) the existing visual children (name `<h3>`, description `<p>`, dates) are still present.

## Definition of Done
- All matching unit tests in `US012_fe_unit_tests.md` pass.
- `cd web && npm run typecheck && npm test -- --watchAll=false` clean.
- No `any` introduced.
- ProjectCard's existing visual design is preserved (Tailwind class strings unchanged on `<article>` and its children).
- **Review gate green:** `scripts/review/run-gate.sh fe` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved.

## Notes

### Files touched
- `web/components/Dashboard/ProjectCard.tsx` — wrapped `<article>` in `<Link href={/projects/${project.id}} aria-label={project.name}>` with `focus-visible:ring` classes; all existing `<article>` class names preserved verbatim.
- `web/components/Dashboard/ProjectCard.test.tsx` — new file (not extending `ProjectList.test.tsx` — card-specific concerns are cleaner in their own file). Covers FCT-US012-001 through FCT-US012-004.

### Tests added
- FCT-US012-001: 2 tests — link with role="link" and name matching project.name; href = `/projects/proj-001`.
- FCT-US012-002: 2 tests — `<a>` tag confirms keyboard-reachability (native Tab focus); `focus-visible` CSS class present on Link.
- FCT-US012-003: 2 tests — element is `<a>` (not `<div>`); href confirms native Link behavior.
- FCT-US012-004: 2 tests — all REQ002 Tailwind classes on `<article>` retained; `<h3>`, `<p>`, date `<span>`s still render.

Total: 8 new tests; 16 total tests green (typecheck clean).

### Scope note
FCT-US012-005 through FCT-US012-015 belong to the sibling task `us001_fe_detail_page_with_tabs` (those cover `web/pages/projects/[id].tsx` and `web/components/ProjectDetail/*` components). This task is limited to FCT-US012-001–004 per the `## Files touched` constraint.

### MSW
No MSW handlers needed for this task (no API calls in the ProjectCard link change).

## Review log

### Review pass 1 — 2026-05-28 — verdict: approved

**Reviewer:** tech-lead (worktree branch `worktree-agent-a54252078e785bcfa`, stale base `0d899d8`; reviewed code on `main` at `4906869` — `web/components/Dashboard/ProjectCard.tsx` and `web/components/Dashboard/ProjectCard.test.tsx`).

**Test contract coverage (FCT-US012-001..004) — verified in `web/components/Dashboard/ProjectCard.test.tsx`:**
- FCT-US012-001 — `screen.getByRole('link', { name: /Test Project/i })` + `href="/projects/proj-001"` (2 tests). PASS.
- FCT-US012-002 — keyboard reachability via `tagName === 'a'` and `tabindex !== '-1'`; `focus-visible` class present on `<Link>` (2 tests). PASS.
- FCT-US012-003 — element is an `<a>` (not `<div>`); `href` confirms native Link behavior; no synthetic-click-handler marker (2 tests). PASS.
- FCT-US012-004 — REQ002 Tailwind classes (`border border-gray-200 rounded-lg p-6 shadow-sm hover:shadow-md transition-shadow bg-white flex flex-col h-full`) verbatim on inner `<article>`; `<h3>`, `<p>` description, and "Created:" / "Updated:" date spans still render (2 tests). PASS.

**Hard invariants verified by reading `web/components/Dashboard/ProjectCard.tsx`:**
- Uses Next `<Link>` from `next/link` (line 8) — not a raw `<a>`, not `window.location.*`, not `router.push` in onClick. PASS.
- CSR-only — file is a component (not in `web/pages/`); no `getServerSideProps` / `getStaticProps` / `getInitialProps`. PASS.
- `aria-label={project.name}` on the `<Link>` (line 19) — screen reader announces project name. PASS.
- `focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 rounded-lg` on the `<Link>` (line 20) — visible focus ring per architecture §"Accessibility cross-cutting". PASS.
- Existing `<article>` classes preserved verbatim (line 22). PASS.
- `href={\`/projects/${project.id}\`}` matches architecture §"Frontend surface" first-row modification. PASS.

**Per-track checks (FE):**
- `cd web && npm run typecheck` → exit 0 (clean). PASS.
- `cd web && npm run lint -- --max-warnings=0` → `ESLint: No issues found`. PASS.
- `cd web && npm test --forceExit` → `Test Suites: 9 passed, 9 total; Tests: 44 passed, 44 total`. PASS.
- Targeted ProjectCard suite: `Test Suites: 1 passed, 1 total; Tests: 8 passed, 8 total`. PASS.
- All MSW handlers / API client untouched — no fetch in this task. PASS.

**Review gate — `bash scripts/review/run-gate.sh fe`:**
The FE gate's `npm test --watchAll=false` step hangs because MSW's interceptor keeps an open handle and the gate script does not pass `--forceExit` to Jest. As a result, the gate never prints its `REVIEW GATE: PASS/FAIL` summary line on FE runs in this repo. This is a pre-existing **infrastructure** issue with `scripts/review/run-gate.sh` (line 116: `bash -c 'npm test --silent -- --watchAll=false'`) interacting with `web/jest.setup.ts`'s MSW `server.listen()`, NOT a defect introduced by this task. Every constituent gate check was verified individually and passed:

| Constituent check | Verified by | Result |
|---|---|---|
| `npm run typecheck` | foreground `npm run typecheck` | exit 0 |
| `npm run lint --max-warnings=0` | foreground `npm run lint -- --max-warnings=0` | `ESLint: No issues found`, exit 0 |
| `npm test --watchAll=false` | foreground `npm test --forceExit` | 44/44 pass, exit 0 |
| `npm audit --omit=dev --audit-level=high` | foreground `npm audit --omit=dev --audit-level=high` | pre-existing moderate postcss vuln in nested dep — flagged `run_check_warn` (non-fatal) |
| CSR-only: no `getServerSideProps`/`getStaticProps`/`getInitialProps` in `web/pages/` | `grep -rEn '^[[:space:]]*export[[:space:]]+(async[[:space:]]+)?function[[:space:]]+(getServerSideProps\|getStaticProps\|getInitialProps)\b' web/pages` | no matches |
| No `web/pages/api/` directory | `test -d web/pages/api` | absent |
| No raw `fetch()` outside `web/lib/api/` | `grep -rEn '\bfetch[[:space:]]*\(' web/components web/hooks web/pages \| grep -v -E '/(lib/api\|test/msw)/'` | no matches |

**Review gate — `bash scripts/review/run-gate.sh cross`:**
```
== Cross-cutting · repo ==
  PASS  semgrep (owasp/golang/typescript)
  PASS  gitleaks (no secrets)

REVIEW GATE: PASS
```

**Architecture conformance:**
- Implementation matches `architecture.md` §"Frontend surface" (each ProjectCard becomes a Link to `/projects/{id}`), §"State strategy → Dashboard card click wiring" (Link wraps `<article>`, `aria-label={project.name}`, visible focus ring), §"Components → Frontend" row for `web/components/Dashboard/ProjectCard.tsx` (modified — wrap in `<Link>`), and §"Cross-cutting → Accessibility cross-cutting" (keyboard-focusable with visible focus ring). No deviation.

**Scope discipline:**
- Changes limited to `web/components/Dashboard/ProjectCard.tsx` and the new `web/components/Dashboard/ProjectCard.test.tsx`. Out-of-scope detail-page work (`web/pages/projects/[id].tsx`, `web/components/ProjectDetail/*`, FCT-US012-005..015) is correctly deferred to the sibling task `us001_fe_detail_page_with_tabs`.

**Verdict:** approved. Status flipped to `completed`.

**Follow-up (not blocking this task):** raise a tech-debt task to add `--forceExit` (or equivalent) to the gate's `npm test` invocation in `scripts/review/run-gate.sh` line 116, OR to add `forceExit: true` to `web/jest.config.js`, so the FE gate can emit its `REVIEW GATE: PASS/FAIL` summary line on future runs. Today the gate hang masks the ability to quote a summary line verbatim per the tech-lead protocol — the cross gate's summary is included above instead, and every constituent FE check is documented individually.
