# US010 — Fix React-Doctor baseline regressions (top-3 state/effect + 1 security)

**Requirement:** REQ005 — quality hardening retrospective
**Status:** draft
**Track hint:** FE

## Story

As a fe-dev finishing REQ005, I want the React-Doctor baseline regressions introduced by REQ004 cleared, so that (a) the security `no-danger` issue in `MermaidDiagram` is closed, (b) `useDocument` stops cascading state and adjusting state on prop change, (c) `DocumentsTab` stops performing client-side redirects inside `useEffect`, and (d) future REQs are gated against a clean, smaller baseline rather than today's 92/100 with four errors.

## Acceptance criteria

- **Scenario: MermaidDiagram no longer uses unsanitised `dangerouslySetInnerHTML`**
  - Given the current implementation at `web/components/ProjectDetail/MermaidDiagram.tsx:139` injects mermaid-rendered SVG via `dangerouslySetInnerHTML`
  - When the file is refactored to either (a) sanitise the mermaid SVG output with DOMPurify before injection, or (b) parse the rendered SVG into a DOM node and append it via a React `ref`
  - Then `npx react-doctor scan web/` reports zero `react-doctor/no-danger` findings on `web/components/ProjectDetail/MermaidDiagram.tsx`
  - And the diagram still renders visually identically to today for every existing mermaid input used in the app's tests and fixtures.

- **Scenario: `useDocument` stops cascading state and adjusting state on prop change**
  - Given `web/hooks/useDocument.ts` lines 39–64 currently trigger `react-doctor/no-cascading-set-state` (7 setStates in one effect at line 50), `react-doctor/no-adjust-state-on-prop-change` × 3 at lines 62/63/64, and `react-doctor/rendering-usetransition-loading` at line 39
  - When the hook is refactored to use `useReducer` with a single dispatched action per logical state transition, and any state previously derived from props is computed inline at render-time rather than stored
  - Then `npx react-doctor scan web/` reports zero findings on `web/hooks/useDocument.ts` for all four rule IDs: `no-cascading-set-state`, `no-adjust-state-on-prop-change`, `rendering-usetransition-loading`, and the related `exhaustive-deps` knock-ons.

- **Scenario: `DocumentsTab` no longer redirects from inside `useEffect`**
  - Given `web/components/ProjectDetail/DocumentsTab.tsx` lines 40/69/78 currently trigger `react-doctor/no-event-handler` (line 40), `react-doctor/nextjs-no-client-side-redirect` (line 69 — `router.replace` inside `useEffect`), and `react-doctor/exhaustive-deps` (line 78 — missing `router.*` deps)
  - When the `router.replace` call is moved out of `useEffect` and into the click handler that owns the redirect intent, and any remaining `exhaustive-deps` violations are corrected
  - Then `npx react-doctor scan web/` reports zero findings on `web/components/ProjectDetail/DocumentsTab.tsx` for `no-event-handler`, `nextjs-no-client-side-redirect`, and `exhaustive-deps`.

- **Scenario: react-doctor score recovers**
  - Given all three fixes above land
  - When `npx react-doctor scan web/` runs
  - Then the reported score is **at least 96/100**
  - And none of the following seven rule IDs appear anywhere in the output: `react-doctor/no-danger`, `react-doctor/no-cascading-set-state`, `react-doctor/no-adjust-state-on-prop-change`, `react-doctor/rendering-usetransition-loading`, `react-doctor/no-event-handler`, `react-doctor/nextjs-no-client-side-redirect`, `react-doctor/exhaustive-deps`.

- **Scenario: no regressions in baseline noise**
  - Given the 15 lower-severity findings from today's scan are recorded as the baseline (button-type, prefer-tag-over-role, design-no-three-period-ellipsis, design-no-em-dash-in-jsx-text, design-no-redundant-size-axes, no-noninteractive-element-to-interactive-role, prefer-module-scope-static-value, deslop/unused-dev-dependency `whatwg-fetch`, and other recorded items)
  - When tech-lead reviews the change using `react-doctor --diff` against the recorded baseline
  - Then no NEW rule fires that wasn't present in the original baseline (the diff shows only removals — the four errors targeted above — not additions).

- **Scenario: tests still pass and hook contract is unchanged**
  - Given the `useDocument` refactor to `useReducer` may require updates to existing Jest tests that assert on internal state-update sequencing
  - When `npm test` and `npm run typecheck` are run inside `web/`
  - Then both exit 0
  - And every existing test that consumes `useDocument`'s returned object (the public surface — fields, types, and observable transitions like loading → loaded / error) passes without changing those assertions; tests may only be updated to drive the new `useReducer` shape internally, never to weaken external behaviour.

## UI / UX flow expectations

**No UI change.** This is internal refactor plus a security tightening. User-visible behaviour of `MermaidDiagram`, document loading via `useDocument`, and tab switching / redirect flow in `DocumentsTab` MUST remain identical to today. No new screens, no new copy, no new states, no new keyboard behaviour, no new loading indicators.

## Out of scope

- The 15 lower-severity React-Doctor baseline findings: `button-type`, `prefer-tag-over-role`, `design-no-three-period-ellipsis`, `design-no-em-dash-in-jsx-text`, `design-no-redundant-size-axes`, `no-noninteractive-element-to-interactive-role`, `prefer-module-scope-static-value`, `deslop/unused-dev-dependency` (`whatwg-fetch`). These remain recorded baseline noise and are tracked via `react-doctor --diff` going forward, not in this story.
- Adopting DOMPurify as a project-wide policy. If the dev chooses the sanitise path for `MermaidDiagram`, the dependency lands in `web/package.json` for this use only; broader policy decisions belong elsewhere.
- Introducing `useReducer` everywhere in the codebase. Only `useDocument` is refactored in this story.
- Any change to the mermaid library version, configuration, or load mode (eager vs lazy). The lazy-load investigation belongs to the deferred MSW leak hunt, not US010.
- Any change to the public TypeScript shape of `useDocument`'s returned object (fields, types, ordering of observable transitions). Behaviour-preserving refactor only.

## Dependencies

- **Soft dependency on US006** (FE hooks AbortController harmonisation). If US006 lands first, the `useDocument` refactor in US010 MUST NOT regress its AbortController contract — signal threading through `lib/api/`, cleanup on unmount, and abort-on-prop-change semantics must be preserved across the `useReducer` rewrite. If US010 lands first, US006 picks up the already-reduced `useDocument` and just confirms the signal threading is still in place. The system architect locks the ordering.

## Notes for the team

- Today's recorded baseline (tech-lead full scan, 2026-06-01): **92/100, 4 errors, 19 warnings.** ~70% of the state/effect findings live in REQ004-shipped files. US010 deliberately targets only the 4 errors (3 state/effect clusters + 1 security) plus their immediate `exhaustive-deps` knock-ons.
- The fix direction for `MermaidDiagram` is architect/dev choice (sanitise vs ref-append) — the AC binds the outcome (`no-danger` no longer fires + visual parity), not the mechanism.
- The fix direction for `useDocument` is locked to `useReducer` because the cascading-setState + adjust-on-prop-change combination is exactly the pattern `useReducer` resolves; derived state from props should be computed inline at render-time, not stored in reducer state.
- The fix for `DocumentsTab` is locked: `router.replace` moves from `useEffect` into the click handler that owns the redirect intent. The `exhaustive-deps` fix should fall out naturally once the effect is simplified or removed; if a residual effect remains, deps must be exhaustive.
- Tech-lead enforces "no NEW rule fires" via `react-doctor --diff` at code review. Devs should run `npx react-doctor scan web/` locally before pushing.

## Sign-off log

(po-ba appends here on each sign-off pass)
