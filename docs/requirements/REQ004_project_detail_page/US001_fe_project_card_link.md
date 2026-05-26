---
US: US001
Title: Make ProjectCard clickable as a link to /projects/{id}
Status: in_review
Track: FE
Implements: US001 AC "Dashboard project cards are clickable (in-scope for this story)", "Click a project card to open the detail page"
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
The dev must make the matching cases in `US001_fe_unit_tests.md` for "ProjectCard becomes a link" / clickable card AC pass (FCT-* IDs assigned by tester). If the tester has not yet authored the relevant IDs at the time the dev picks this up, the dev flags it back to tester rather than skipping coverage.

## Implementation notes
- Import `Link` from `next/link`. The Pages Router pattern is `<Link href="/projects/abc"><article …>…</article></Link>` — no `<a>` wrapping needed in modern Next.js (it renders the anchor itself).
- `aria-label={project.name}` belongs on the `<Link>` so screen readers announce the project name.
- Preserve **all** existing `className` strings on the `<article>` and inner elements — the REQ002 sign-off cited "minimal beautiful" and this change is additive.
- Add a `focus-visible:` ring class (e.g. `focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2`) on the `<Link>` itself for keyboard focus visibility.
- Tests should assert: (a) the rendered element has `role="link"` (implicit on the anchor) with `aria-label` matching `project.name`; (b) `href` is `/projects/${project.id}`; (c) the existing visual children (name `<h3>`, description `<p>`, dates) are still present.

## Definition of Done
- All matching unit tests in `US001_fe_unit_tests.md` pass.
- `cd web && npm run typecheck && npm test -- --watchAll=false` clean.
- No `any` introduced.
- ProjectCard's existing visual design is preserved (Tailwind class strings unchanged on `<article>` and its children).
- **Review gate green:** `scripts/review/run-gate.sh fe` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved.

## Notes

### Files touched
- `web/components/Dashboard/ProjectCard.tsx` — wrapped `<article>` in `<Link href={/projects/${project.id}} aria-label={project.name}>` with `focus-visible:ring` classes; all existing `<article>` class names preserved verbatim.
- `web/components/Dashboard/ProjectCard.test.tsx` — new file (not extending `ProjectList.test.tsx` — card-specific concerns are cleaner in their own file). Covers FCT-US001-001 through FCT-US001-004.

### Tests added
- FCT-US001-001: 2 tests — link with role="link" and name matching project.name; href = `/projects/proj-001`.
- FCT-US001-002: 2 tests — `<a>` tag confirms keyboard-reachability (native Tab focus); `focus-visible` CSS class present on Link.
- FCT-US001-003: 2 tests — element is `<a>` (not `<div>`); href confirms native Link behavior.
- FCT-US001-004: 2 tests — all REQ002 Tailwind classes on `<article>` retained; `<h3>`, `<p>`, date `<span>`s still render.

Total: 8 new tests; 16 total tests green (typecheck clean).

### Scope note
FCT-US001-005 through FCT-US001-015 belong to the sibling task `us001_fe_detail_page_with_tabs` (those cover `web/pages/projects/[id].tsx` and `web/components/ProjectDetail/*` components). This task is limited to FCT-US001-001–004 per the `## Files touched` constraint.

### MSW
No MSW handlers needed for this task (no API calls in the ProjectCard link change).

## Review log
(left for tech-lead review pass entries)
