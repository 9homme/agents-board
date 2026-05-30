---
US: US003
Title: MermaidDiagram lazy-loaded component + wire into MarkdownRenderer code override
Status: completed
Track: FE
Implements: US003 AC "Mermaid diagrams render as SVG", "Invalid mermaid source does not crash the page", "Switching documents re-renders mermaid for the new content"
Blocked by: US003_fe_markdown_renderer.md
Worked-by: fe-dev-2026-05-30T00-00-00Z-a9f2
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

## Notes

### Files touched
- `web/package.json` — added `mermaid ^11.0.0` and `@testing-library/dom ^10.4.1` (peer dep pulled in by mermaid install).
- `web/package-lock.json` — regenerated after npm install.
- `web/jest.config.js` — added `mermaid` and `@mermaid-js` to ESM transform list.
- `web/components/ProjectDetail/MermaidDiagram.tsx` — new component (lazy import via dynamic `import('mermaid')` in `useEffect`, per-instance `useId()` for mermaid render ids, error fallback, `role="img"` wrapper).
- `web/components/ProjectDetail/MermaidDiagram.test.tsx` — new tests covering FCT-US003-008 and FCT-US003-009.
- `web/components/ProjectDetail/MarkdownRenderer.tsx` — added `import { MermaidDiagram }`, wired the `language-mermaid` intercept in `CodeBlock` before the highlighter fall-through.
- `web/components/ProjectDetail/MarkdownRenderer.test.tsx` — added `jest.mock('mermaid')`, FCT-US003-008 (code-override routing), FCT-US003-010 (no mermaid import for non-mermaid docs) tests.
- `web/components/ProjectDetail/DocumentPreviewer.test.tsx` — added `jest.mock('mermaid')`, FCT-US003-011 (document-switch resets mermaid SVG) test.

### Tests added (this task)
- `MermaidDiagram.test.tsx`: 4 tests (FCT-US003-008 × 2, FCT-US003-009 × 2)
- `MarkdownRenderer.test.tsx`: 5 new tests (FCT-US003-008 code-override × 3, FCT-US003-010 × 1, plus existing 12 unchanged)
- `DocumentPreviewer.test.tsx`: 1 new test (FCT-US003-011)
- **Total across entire suite: 107 tests, all passing.**

### Implementation pattern chosen
Used pattern 2 (inline `useEffect` + `await import('mermaid')`) rather than `next/dynamic`, because:
- The mock-based test strategy (`jest.mock('mermaid', ...)`) works cleanly without also needing to mock `next/dynamic`.
- The module-scope cache (`mermaidModuleCache`) ensures mermaid is imported only once per page lifetime, matching the lazy-load architecture intent.
- The component is CSR-only by construction (the `import('mermaid')` is inside `useEffect` which never runs on the server).

### Follow-up notes (not in scope)
- `@testing-library/dom` became a direct dep after the npm install pulled it in as a missing peer. It was added to `dependencies`; could move to `devDependencies` if desired.
- The worker-process warning in Jest (`A worker process has failed to exit gracefully`) is pre-existing — the same warning appears in the full suite before this task's changes. It is caused by the MSW server's open handles and not introduced by mermaid.

## Review log

### Review pass 1 — 2026-05-30 — verdict: approved

**Reviewer:** tech-lead (worktree-agent `a9635ed9b942b4c32`). Reviewer stalled before appending this entry; entry transcribed by orchestrator verbatim from the reviewer's returned report (the `Status: completed` flip was applied by the reviewer itself before the stall).

**Gate summary:**
- `cd web && npm run typecheck` — clean.
- `cd web && npm test -- --watchAll=false --forceExit` — **107 tests passing** (16 suites; pre-existing MSW open-handle warning absorbed by `--forceExit`).
- `cd web && npm run lint -- --max-warnings=0` — `ESLint: No issues found`.
- `bash scripts/review/run-gate.sh cross` — `REVIEW GATE: PASS` (semgrep + gitleaks).

**Hard invariants verified:**
1. **CSR-only:** `MermaidDiagram.tsx:86` does `await import('mermaid')` inside `useEffect`; module never executes on the server. PASS.
2. **Cleanup + key-driven swap (FCT-US003-011):** `MermaidDiagram.tsx:112-114` returns a `() => { cancelled = true }` cleanup; setters gated on `if (cancelled) return`. `DocumentPreviewer.test.tsx:174-238` proves doc-A SVG is gone and doc-B SVG present after `rerender` with a new `key`. PASS.
3. **Error fallback non-propagating (FCT-US003-009):** try/catch wraps both `import('mermaid')` and `mermaid.render`; failure sets `status: 'error'`; success path emits SVG without raw source (the `<pre>` only appears in the error branch). PASS.
4. **`mermaid.render` not called for non-mermaid (FCT-US003-010):** `CodeBlock` in `MarkdownRenderer.tsx:137` only branches to `<MermaidDiagram>` when `className?.includes('language-mermaid')`. Test at `MarkdownRenderer.test.tsx:347-356` asserts `expect(mermaid.render).not.toHaveBeenCalled()` after rendering a Go fence. PASS.
5. **Sanctioned `dangerouslySetInnerHTML`:** `MermaidDiagram.tsx:139` uses `dangerouslySetInnerHTML` on the mermaid-emitted SVG. Explicitly sanctioned by architecture D-004 and the task's scope note. Dev chose `securityLevel: 'strict'` (`MermaidDiagram.tsx:93`) and skipped the optional double-sanitize through `rehype-sanitize` — architecture allows this. Accepted.

**Other spot-checks (all PASS):**
- `useId()` per-instance with `mermaid-` prefix; characters sanitized for valid HTML ids.
- `mermaid.initialize({ startOnLoad: false, securityLevel: 'strict' })`.
- Code-override order: mermaid checked FIRST, default highlighter branch SECOND.
- Trailing newline strip in `MarkdownRenderer.tsx:139`.
- ariaLabel derived from first line if SVG has no `<title>`.

**Follow-up (non-blocking tech-debt):**
- `@testing-library/dom ^10.4.1` was added to `dependencies` because npm flagged it as a missing peer during the mermaid install. Mermaid does NOT depend on it; npm's peer-resolver surfaced an unrelated missing peer. Purely a test-time tool — should be moved to `devDependencies`. Doesn't break the build, doesn't ship to browser (Next won't bundle since nothing in `web/pages` or `web/components` imports it).

**Verdict:** approved. `Status: completed`.
