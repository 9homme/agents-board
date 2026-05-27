---
US: US003
Title: MermaidDiagram lazy-loaded component + wire into MarkdownRenderer code override
Status: pending
Track: FE
Implements: US003 AC "Mermaid diagrams render as SVG", "Invalid mermaid source does not crash the page", "Switching documents re-renders mermaid for the new content"
Blocked by: US003_fe_markdown_renderer.md
Worked-by:
---

## Goal
Add a `MermaidDiagram` component that lazy-loads the `mermaid` library via `next/dynamic({ ssr: false })`, renders the diagram as inline SVG (sanitized), falls back to a friendly inline error + raw source when rendering fails, and wire it into the `code` component override of `MarkdownRenderer` so fenced blocks tagged `mermaid` render as diagrams instead of code.

## Architecture references
- `architecture.md` §"Components → Frontend" → row `MermaidDiagram.tsx` (new — US003).
- `architecture.md` §"Markdown rendering plan → Mermaid mechanics" — lazy-load, re-render on document switch via parent `key`, error handling (catches both import failure and render exception → small `<div role="alert">Could not render diagram</div>` + `<pre>` fallback), accessibility (wrapper carries `role="img"` + `aria-label` from first line of source if SVG has no `<title>`).
- `architecture.md` §"Markdown rendering plan → Bundle-size posture" — mermaid is the heaviest dep; lazy chunk only loaded when a `mermaid` fence is encountered.
- `architecture.md` §"Key decisions" → D-004 (mermaid via `next/dynamic({ ssr: false })`, custom `code` override routes `language-mermaid` to `MermaidDiagram`).

## Scope
- **In:**
  - Add `mermaid ^11` to `web/package.json`. Run `npm install` and commit `package-lock.json`.
  - Create `web/components/ProjectDetail/MermaidDiagram.tsx`:
    - Default export `MermaidDiagram({ source }: { source: string })`.
    - Lazy-loads the `mermaid` library. Two acceptable patterns; either is fine — the dev picks the cleaner fit:
      1. `const MermaidImpl = dynamic(() => import('./MermaidImpl'), { ssr: false })` where `MermaidImpl` is the actual component that does the `import('mermaid')` and `mermaid.render(...)` work; OR
      2. Inline `useEffect(() => { import('mermaid').then(...) }, [source])` inside `MermaidDiagram` itself, with the import promise cached at module scope on first success.
    - On mount / `source` change: call `mermaid.render(uniqueId, source)` (mermaid v11 returns `{ svg, bindFunctions }`); set the resulting SVG string into React state. Render it via `dangerouslySetInnerHTML` (the only sanctioned use in REQ004 — the SVG is produced by mermaid from the source string AND the sanitizer's allow-list already covers the SVG element set; for belt-and-braces optionally pipe the SVG through `rehype-sanitize`'s `sanitize` function before injecting).
    - Wrap the SVG output in a `<div role="img" aria-label={ariaLabel}>` where `ariaLabel` is derived from the first non-empty line of the source if mermaid's SVG does not already carry a `<title>` element. Width fits the previewer container (CSS `max-width: 100%`).
    - Error handling: wrap the `import('mermaid')` AND the `mermaid.render(...)` calls in try/catch. On either failure, render `<div role="alert">Could not render diagram</div>` followed by the raw `source` in a `<pre><code>` fallback. The component must NOT throw — failures stay contained so the surrounding markdown still renders.
    - Use a per-instance unique `id` for `mermaid.render` (e.g. `useId()` from React, prefixed `mermaid-`) so multiple diagrams in the same document don't collide on mermaid's internal SVG id namespace.
    - Initialise mermaid once per module load (`mermaid.initialize({ startOnLoad: false, securityLevel: 'strict' })`) — `securityLevel: 'strict'` prevents mermaid from emitting click handlers / arbitrary HTML.
  - Modify `web/components/ProjectDetail/MarkdownRenderer.tsx` `code` component override:
    - When `inline === false` AND `className?.includes('language-mermaid')`, render `<MermaidDiagram source={String(children).replace(/\n$/, '')} />` instead of the default `<pre><code>`.
    - All other (`language-*` and no-language) code blocks fall through to the existing default rendering (highlighted by `rehype-highlight`).
    - The order matters: detect mermaid FIRST, fall through to the default branch SECOND. This is the architecture's pipeline note: "the custom `code` component override intercepts mermaid first" so the highlighter never sees it.
    - Optionally tune `rehype-highlight` config to `ignoreMissing: true` (already done in the prior task) so a missing `mermaid` language registration doesn't throw — the override should fire before the highlighter on a mermaid block anyway, but `ignoreMissing` makes the pipeline robust.
  - Jest tests for `MermaidDiagram`:
    - **Mock `mermaid`** in tests (`jest.mock('mermaid', ...)` or via the dynamic-import boundary). Mocking is essential — real mermaid renders to DOM with timers and is impractical in jsdom. The mock's `render(id, src)` returns `{ svg: '<svg data-testid="fake-mermaid"><title>fake</title></svg>' }` for the happy path; throws for the failure-path test.
    - Happy path: passing a valid `source` results in an `<svg>` appearing in the DOM (selected by `data-testid` from the mock or by `role="img"` on the wrapper).
    - Failure path: when the mocked `mermaid.render` throws, the component renders `role="alert"` "Could not render diagram" AND a `<pre>` containing the raw source. No exception escapes the component.
    - Multiple diagrams: rendering two `MermaidDiagram` instances in the same test does NOT throw on duplicate ids (proves the per-instance `useId` worked).
  - Jest tests for the `MarkdownRenderer` `code` override:
    - A ```mermaid``` fence in `source` results in a `MermaidDiagram` being rendered (assert by `role="img"` from the mock) AND the raw mermaid source is NOT present as a `<code>` block (per the AC "the raw mermaid source is NOT shown to the user").
    - A ```go``` fence still renders as a highlighted `<pre><code class="hljs language-go">` (the override fell through correctly).
  - Document-switch test (component-level, in `DocumentPreviewer.test.tsx` or `DocumentsTab.test.tsx` — dev's call):
    - Render the previewer with document A containing a mermaid block, switch to document B containing a different mermaid block, assert that B's mock-svg is shown and A's is gone. This validates the `key={document.id}` cleanup (already wired by US002 FE; this test confirms it still holds after the markdown swap).
- **Out:**
  - Any change to `DocumentsTab.tsx` / `DocumentSidebar.tsx` / hooks / API client / BE.
  - Configuring mermaid themes, fonts, or icons beyond the defaults.
  - Pre-rendering / SSR of mermaid — explicitly forbidden by D-004.
  - PlantUML / other diagram languages — out of REQ004 scope.

## Files touched (estimated, exclusive)
- `web/package.json` (modified — add `mermaid ^11`)
- `web/package-lock.json` (modified — regenerated)
- `web/components/ProjectDetail/MermaidDiagram.tsx` (new)
- `web/components/ProjectDetail/MermaidDiagram.test.tsx` (new)
- `web/components/ProjectDetail/MarkdownRenderer.tsx` (modified — `code` override branches on `language-mermaid`)
- `web/components/ProjectDetail/MarkdownRenderer.test.tsx` (modified — add the override-routing tests)
- `web/components/ProjectDetail/DocumentPreviewer.test.tsx` OR `web/components/ProjectDetail/DocumentsTab.test.tsx` (modified — document-switch mermaid cleanup test; dev's call which file is the natural home)
- `web/jest.config.js` and/or `web/jest.setup.ts` (possibly modified — to install the `mermaid` mock and handle the `next/dynamic` boundary if needed; dev's call)

## Test contract
The dev must make the matching cases in `US003_fe_unit_tests.md` pass — covering: "Mermaid diagrams render as SVG", "Invalid mermaid source does not crash the page", "Switching documents re-renders mermaid for the new content", and the `code`-override routing tests. (FCT-* IDs assigned by tester.) If the tester has not yet authored the relevant IDs at the time the dev picks this up, the dev flags it back to tester rather than skipping coverage.

## Implementation notes
- **CSR-only.** `mermaid` references `window` / `document` at module-load time on older versions; even on v11 the safe posture is `next/dynamic({ ssr: false })` OR `await import('mermaid')` inside a `useEffect`. Either way, the page must build cleanly (`npm run build`).
- **Don't use raw `dangerouslySetInnerHTML` on user input.** What goes through `dangerouslySetInnerHTML` here is the SVG **emitted by mermaid from the source**, not the source itself. Mermaid's `securityLevel: 'strict'` plus the sanitizer's SVG allow-list (already extended in the prior task) is the trust chain. Architecture is explicit that the optional belt-and-braces re-sanitization of mermaid's SVG is allowed but not required if `securityLevel: 'strict'` is set.
- The unique-id requirement: mermaid v11's `render(id, source)` writes a hidden iframe / dom element keyed on `id`; duplicate ids cause one diagram to overwrite the other's container. Use `useId()` (React 18+) prefixed with `mermaid-` to be safe.
- The "switching documents re-renders mermaid" AC is satisfied by US002 FE's existing `key={document.id}` on `DocumentPreviewer` — this task adds a test that proves it still works after the markdown swap. Do NOT add a separate cleanup mechanism inside `MermaidDiagram` — the parent `key` is the architecture's chosen pattern.
- Mocking strategy: in Jest, the simplest approach is `jest.mock('mermaid', () => ({ default: { initialize: jest.fn(), render: jest.fn().mockResolvedValue({ svg: '<svg role="img" aria-label="diagram"></svg>' }) } }))` (mermaid v11's API is async `render`). For failure-path tests, override per-test with `(mermaid.render as jest.Mock).mockRejectedValueOnce(new Error('bad'))`. If the dev uses `next/dynamic`, also mock the dynamic boundary or set `jest.mock('next/dynamic', () => (loader) => loader())` to make the lazy load synchronous in tests.
- The "raw source not shown" AC matters: the override must REPLACE the code block, not append the diagram next to it. Returning `<MermaidDiagram source={...} />` from the `code` override component does this naturally because react-markdown uses the override's return value in place of the default `<pre><code>` tree.

## Definition of Done
- All matching unit tests in `US003_fe_unit_tests.md` pass.
- `cd web && npm run typecheck && npm test -- --watchAll=false` clean.
- `cd web && npm run build` succeeds (no SSR-time `window`/`document` reference broke the build).
- No `any` introduced (where unavoidable for the mermaid mock, scope it tightly).
- No `getServerSideProps` / `getStaticProps` / `getInitialProps` / `web/pages/api/*` introduced (CSR-only).
- `MermaidDiagram` has a doc comment summarising the lazy-load + error-handling contract.
- **Review gate green:** `scripts/review/run-gate.sh fe` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved.

## Review log
(left for tech-lead review pass entries)
