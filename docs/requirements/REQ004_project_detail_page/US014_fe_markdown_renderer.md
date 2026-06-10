---
US: US014
Title: MarkdownRenderer + sanitization + syntax highlighting + DocumentPreviewer body swap
Status: completed
Track: FE
Implements: US014 AC "Headings render as headings", "Paragraphs, bold, italic, and inline code render", "Lists render (ordered + unordered + task lists)", "Tables render", "Links render and are safe", "Code fences render syntax-highlighted", "XSS — script tags in content are not executed"
Blocked by: US013_fe_documents_tab.md
Worked-by: fe-dev-2026-05-30T000000Z-a1b2
---

## Goal
Replace the plain `<pre>` / `<div>` content rendering inside `DocumentPreviewer` with a real GFM markdown renderer (`react-markdown` + `remark-gfm` + `rehype-sanitize` + `rehype-highlight`), wrapped in an error boundary, with a custom `code` component override that will (in the mermaid task) route `language-mermaid` fences to `MermaidDiagram` and otherwise fall through to the highlighter. This task wires everything except mermaid (the next task wires mermaid into the override that this task builds).

## Architecture references
- `architecture.md` §"Components → Frontend" → row `MarkdownRenderer.tsx` (new — US014) and `DocumentPreviewer.tsx` (upgraded — US014 swaps the body).
- `architecture.md` §"Markdown rendering plan" — full pipeline diagram, library choices, sanitization schema patches (allow `language-*` / `hljs` and the SVG element set mermaid emits), highlight.js language subset (Go, TypeScript, JSON, Bash, SQL, YAML, Markdown).
- `architecture.md` §"Key decisions" → D-004 (markdown stack rationale, MDX/Shiki/marked+DOMPurify rejection).
- `architecture.md` §"Risks & open questions" → "untrusted markdown source — FE-only sanitization (Option D)" (the XSS AC is testable and load-bearing; sanitizer is the authoritative trust boundary for REQ004).

## Scope
- **In:**
  - Add new dependencies to `web/package.json`: `react-markdown ^9`, `remark-gfm ^4`, `rehype-sanitize ^6`, `rehype-highlight ^7`, `highlight.js` (for the CSS theme + language registration). Run `npm install` and commit `package-lock.json`.
  - Import a `highlight.js` CSS theme in `web/pages/_app.tsx` (or `web/styles/globals.css`) so highlighted tokens are visible — choose a single theme (e.g. `highlight.js/styles/github.css`). Configure `rehype-highlight` to register only the curated language subset (Go, TypeScript, JSON, Bash, SQL, YAML, Markdown) per architecture's bundle-size posture.
  - Create `web/components/ProjectDetail/MarkdownRenderer.tsx`:
    - Exports a default React component `MarkdownRenderer({ source }: { source: string })`.
    - Builds the `react-markdown` pipeline: `remarkPlugins=[remarkGfm]`, `rehypePlugins=[[rehypeSanitize, extendedSchema], [rehypeHighlight, { ignoreMissing: true }]]`. **Order matters** — sanitize must run before highlight so highlight's class names are preserved by the sanitizer's allow-list (architecture's pipeline diagram is explicit).
    - Extends the `defaultSchema` from `hast-util-sanitize`: allow `className` containing `language-*` / `hljs` on `<code>` and `<pre>` elements; allow the SVG element set mermaid will emit (`svg`, `g`, `path`, `rect`, `circle`, `line`, `polyline`, `polygon`, `text`, `tspan`, `defs`, `marker`, `title`, `desc`, plus the attributes mermaid uses); MUST NOT allow `<script>`, `on*=` event handlers, or `javascript:` / `vbscript:` / `data:text/html` URIs. Anchor tags must support `target` and `rel` so links can be safely opened in a new tab.
    - `components` override for `a`: external links (`href` starts with `http`) get `target="_blank"` and `rel="noopener noreferrer"`; internal / relative / `mailto:` left as-is. (Sanitizer already strips `javascript:` URIs — this override is for opening behavior only.)
    - `components` override for `code`: stub for now — accept `inline`, `className`, `children`; for the non-inline case, render the default `<pre><code className={className}>{children}</code></pre>` (the highlighter still applies via its rehype pass). The mermaid task wires the `language-mermaid` branch into this override.
  - Create `web/components/ProjectDetail/MarkdownErrorBoundary.tsx` — a small React class error boundary (or use a `useErrorBoundary`-style functional approach, dev's call) that catches throws from inside `MarkdownRenderer` and renders `<div role="alert">Failed to render document</div>`. Wrap `<MarkdownRenderer source={...} />` inside `<MarkdownErrorBoundary>...</MarkdownErrorBoundary>` at the DocumentPreviewer call site.
  - Update `web/components/ProjectDetail/DocumentPreviewer.tsx`: replace the plain `<pre>` / `<div>` body with `<MarkdownErrorBoundary><MarkdownRenderer source={document.content} /></MarkdownErrorBoundary>`. The component's prop surface stays exactly the same — DocumentsTab needs no change. The `key={document.id}` from DocumentsTab already guarantees clean remount on document switch (architecture: simplest pattern for mermaid SVG cleanup).
  - Jest tests for `MarkdownRenderer` covering: headings (h1/h2/h3 with text), bold/em/inline-code, ordered + unordered + task list (disabled checkbox reflects `[x]`), GFM table → `<table><thead><tbody>`, external link gets `target="_blank" rel="noopener noreferrer"`, `javascript:` URI is stripped (`href` attribute does not contain `javascript:`), fenced ```go``` block carries `language-go` (and `hljs` class via the highlighter) with tokenised `<span>` children proving highlighting actually applied, fenced no-language block renders plain `<pre><code>` without highlighting and without throwing, **XSS AC: `<script>alert(1)</script>` in source results in no `<script>` element in the rendered DOM and no executed alert**.
  - Jest tests for `MarkdownErrorBoundary` (renders fallback when child throws).
  - Update / extend `DocumentPreviewer.test.tsx` if its existing assertions assumed plain rendering — switch them to assert against the rendered markdown DOM (e.g. find `<h1>` instead of looking for raw `#`). The test contract from US014 is the authority for new assertions.
- **Out:**
  - Mermaid diagram rendering — `us003_fe_mermaid_diagram` wires `MermaidDiagram` into the `code` override of this MarkdownRenderer.
  - Any change to `DocumentsTab.tsx` / `DocumentSidebar.tsx` / hooks / API client (US013 owns those; this story does not touch them).
  - LaTeX / KaTeX, footnotes, GFM alerts, MDX, custom directives — explicitly out of REQ004 scope.

## Files touched (estimated, exclusive)
- `web/package.json` (modified — add markdown stack deps)
- `web/package-lock.json` (modified — regenerated)
- `web/pages/_app.tsx` OR `web/styles/globals.css` (modified — import highlight.js CSS theme; dev picks the cleaner location)
- `web/components/ProjectDetail/MarkdownRenderer.tsx` (new)
- `web/components/ProjectDetail/MarkdownRenderer.test.tsx` (new — covers GFM + sanitization + highlight + XSS)
- `web/components/ProjectDetail/MarkdownErrorBoundary.tsx` (new)
- `web/components/ProjectDetail/MarkdownErrorBoundary.test.tsx` (new)
- `web/components/ProjectDetail/DocumentPreviewer.tsx` (modified — swap body for MarkdownRenderer wrapped in error boundary)
- `web/components/ProjectDetail/DocumentPreviewer.test.tsx` (modified — assertions now target the rendered markdown DOM, not plain text)

> **Scaffold posture for `web/package.json` and `web/package-lock.json`:** this task is the only REQ004 FE task that adds new dependencies. It is sequenced **after** `US013_fe_documents_tab.md` because it modifies `DocumentPreviewer.tsx` (created by US013 FE). The next task (`us003_fe_mermaid_diagram`) ALSO adds a dep (`mermaid`) AND modifies `MarkdownRenderer.tsx` — to avoid parallel writes to `package.json` / `package-lock.json` / `MarkdownRenderer.tsx`, the mermaid task `Blocked by:` is set to this task. This task is the **scaffold task for the markdown-stack additions to `web/package.json`** for REQ004.

## Test contract
The dev must make the matching cases in `US014_fe_unit_tests.md` pass — covering: all "headings / paragraphs / lists / tables / links / code fences / XSS" ACs from US014. Specifically, the XSS AC is load-bearing: a `<script>alert(1)</script>` in `content` MUST result in zero `<script>` elements in the rendered tree and zero `alert` side effects. (FCT-* IDs assigned by tester.) If the tester has not yet authored the relevant IDs at the time the dev picks this up, the dev flags it back to tester rather than skipping coverage.

## Implementation notes
- **CSR-only.** The markdown stack is fine to import at module top — `react-markdown` and its plugins do not touch `window`/`document` at import time, so Next.js build will not break. (Mermaid is the one that does — handled in the next task with `next/dynamic({ ssr: false })`.)
- Sanitization schema patch — start from `defaultSchema` (`hast-util-sanitize` re-exports it via `rehype-sanitize`):
  ```ts
  const schema = {
    ...defaultSchema,
    attributes: {
      ...defaultSchema.attributes,
      code: [...(defaultSchema.attributes?.code || []), ['className', /^language-/, 'hljs', /^hljs/]],
      pre:  [...(defaultSchema.attributes?.pre  || []), ['className', /^language-/, 'hljs', /^hljs/]],
      span: [...(defaultSchema.attributes?.span || []), ['className', /^hljs/]],
      a:    [...(defaultSchema.attributes?.a    || []), 'target', 'rel'],
      // SVG allow-list extension for mermaid (used by the next task too):
      svg:  ['xmlns', 'viewBox', 'width', 'height', 'role', 'aria-label', 'className'],
      g:    ['transform', 'className', 'style'],
      path: ['d', 'fill', 'stroke', 'strokeWidth', 'className', 'style'],
      // ...add other SVG elements per architecture's allow-list as needed
    },
    tagNames: [...(defaultSchema.tagNames || []), 'svg', 'g', 'path', 'rect', 'circle', 'line', 'polyline', 'polygon', 'text', 'tspan', 'defs', 'marker', 'title', 'desc'],
  };
  ```
  Validate by running a test with malicious source (`<script>`, `onerror=`, `javascript:` URI) and asserting they are stripped.
- Pipeline order in `rehypePlugins`: `rehypeSanitize` FIRST, then `rehypeHighlight`. If you swap the order, sanitize will strip the `hljs` / `language-*` classes that highlight just added — architecture's pipeline diagram is the contract.
- `rehype-highlight` language registration — to keep the bundle smaller, configure with the curated subset:
  ```ts
  import langGo from 'highlight.js/lib/languages/go';
  // ... etc
  const rehypeHighlightOpts = {
    ignoreMissing: true,
    languages: { go: langGo, typescript: langTs, json: langJson, bash: langBash, sql: langSql, yaml: langYaml, markdown: langMd },
  };
  ```
  Architecture says the dev may extend the subset with justification — if you need more, document it in the task notes for the reviewer.
- The `code` override stub: this task leaves the default behavior intact for ALL languages (including a no-op for `language-mermaid` — render it as a normal code block for now; the mermaid task will replace this branch). This keeps the markdown renderer fully working and shippable on its own, with the mermaid task layering on top.
- `<a>` external-link rule: detect `href.startsWith('http')` or `URL` parsing as absolute; if external, add `target="_blank" rel="noopener noreferrer"`. Architecture allow-list already permits `target` + `rel` attributes on `<a>` post-sanitization, so the override sets them on the React element and they survive.
- The MarkdownErrorBoundary is a defence-in-depth measure — `react-markdown` is well-tested but the highlighter and sanitizer chain can throw on pathological input; the boundary keeps the rest of the page alive (architecture explicitly calls this out).
- Do NOT use `dangerouslySetInnerHTML` anywhere in this task. `react-markdown` returns a React tree — that's the entire point of choosing it over `marked`.

## Definition of Done
- All matching unit tests in `US014_fe_unit_tests.md` pass — including the XSS test.
- `cd web && npm run typecheck && npm test -- --watchAll=false` clean.
- No `any` introduced.
- No `dangerouslySetInnerHTML` in `MarkdownRenderer` (architecture D-004).
- No `getServerSideProps` / `getStaticProps` / `getInitialProps` / `web/pages/api/*` introduced (CSR-only).
- `MarkdownRenderer` has a doc comment summarising the pipeline. `MarkdownErrorBoundary` has a doc comment.
- Bundle-size check: `npm run build` succeeds (no broken imports). The dev should glance at the build's per-page JS size to confirm the curated language registration kept things sane.
- **Review gate green:** `scripts/review/run-gate.sh fe` exits 0, and `scripts/review/run-gate.sh cross` exits 0.
- Dev set status to `in_review` and reported back; tech-lead approved.

## Notes
### Worked-by: fe-dev-2026-05-30T000000Z-a1b2

**Files touched:**
- `/Users/a667282/workspace/agents-board/web/package.json` — added react-markdown ^9, remark-gfm ^4, rehype-sanitize ^6, rehype-highlight ^7, highlight.js
- `/Users/a667282/workspace/agents-board/web/package-lock.json` — regenerated (104 packages added)
- `/Users/a667282/workspace/agents-board/web/pages/_app.tsx` — added `import "highlight.js/styles/github.css"` for syntax highlight theme
- `/Users/a667282/workspace/agents-board/web/components/ProjectDetail/MarkdownRenderer.tsx` — new; implements full GFM pipeline with sanitization and highlight
- `/Users/a667282/workspace/agents-board/web/components/ProjectDetail/MarkdownRenderer.test.tsx` — new; covers FCT-US014-001 through FCT-US014-007, FCT-US014-012 through FCT-US014-015 (12 tests)
- `/Users/a667282/workspace/agents-board/web/components/ProjectDetail/MarkdownErrorBoundary.tsx` — new; class error boundary renders "Failed to render document" fallback
- `/Users/a667282/workspace/agents-board/web/components/ProjectDetail/DocumentPreviewer.tsx` — modified; swapped `<pre>` body for `<MarkdownErrorBoundary><MarkdownRenderer source={document.content} /></MarkdownErrorBoundary>`
- `/Users/a667282/workspace/agents-board/web/components/ProjectDetail/DocumentPreviewer.test.tsx` — existing tests all pass unchanged (markdown renders transparently)
- `/Users/a667282/workspace/agents-board/web/jest.config.js` — extended `transformIgnorePatterns` to include all ESM-only remark/rehype/micromark/hast packages; added CSS mock via `moduleNameMapper`
- `/Users/a667282/workspace/agents-board/web/__mocks__/styleMock.js` — new; maps CSS imports to empty object in Jest

**Tests added:** 12 new tests (FCT-US014-001 through FCT-US014-007, FCT-US014-012 through FCT-US014-015)
**Tests total:** 98 tests across 16 suites — all green

**FCT-US014-008 through FCT-US014-011** (mermaid) are out of scope for this task — they belong to the sibling mermaid task (`us003_fe_mermaid_diagram`) which is `Blocked by:` this task.

**Pipeline order note:** rehype-sanitize runs BEFORE rehype-highlight (as architecture §"Markdown rendering plan" specifies), so the sanitizer's allow-list preserves the `language-*` and `hljs` class names that the highlighter adds. The extended schema allows those class names on `<code>`, `<pre>`, `<span>`.

**jest.config.js change:** The remark/rehype ecosystem is ESM-only (all packages declare `"type": "module"`). The `transformIgnorePatterns` was updated to include all transitive ESM dependencies. This is a one-time cost; no future markdown-ecosystem packages should need re-addition unless a new dep is brought in.

**Lint gate:** `eslint` no issues. `run-gate.sh fe` and `run-gate.sh cross` both exit 0.

## Review log

### Review pass 1 — 2026-05-30 — verdict: approved

**Gate summary:**
- `cd web && npm run typecheck` → clean (no output, exit 0).
- `cd web && npm test -- --watchAll=false --forceExit` → `Test Suites: 16 passed, 16 total` / `Tests: 98 passed, 98 total` / `Time: 4.767 s`. (The Jest "worker process has failed to exit gracefully" line is the MSW open-handle workaround for which `--forceExit` is mandated by the review brief — not a real failure.)
- `cd web && npm run lint -- --max-warnings=0` → `ESLint: No issues found`.
- `scripts/review/run-gate.sh cross` → `REVIEW GATE: PASS` (semgrep + gitleaks both PASS).

**FCT coverage verified (all in `web/components/ProjectDetail/MarkdownRenderer.test.tsx`):**
- FCT-US014-001 headings h1/h2/h3 — green.
- FCT-US014-002 strong/em/del/inline code — green.
- FCT-US014-003 GFM task list disabled checkboxes — green.
- FCT-US014-004 GFM table thead/tbody/columnheader/cell — green.
- FCT-US014-005 safe link `target=_blank rel=noopener noreferrer` + `javascript:` stripped — green.
- FCT-US014-006 `language-go` + `hljs` classes + `<span>` tokens — green (proves sanitize-then-highlight pipeline order preserved the highlighter's classes).
- FCT-US014-007 unlabeled fence renders `<pre><code>` without `hljs` and without throw — green.
- FCT-US014-012 `<script>alert(1)</script>` → 0 script elements AND `alert` spy not called AND safe paragraph survives — green; this is the load-bearing XSS test and it exercises an actual attacker payload, not just an allow-list check.
- FCT-US014-013 `[link](javascript:alert('xss'))` → no anchor has `href` starting with `javascript:` — green.
- FCT-US014-014 `<img onerror="alert(...)">` → 0 `[onerror]` elements AND `alert` not called — green.
- FCT-US014-015 error boundary catches throw and renders `role="alert"` "Failed to render document" fallback (plus passthrough when no error) — green.

Mermaid FCTs (008-011) are explicitly out of scope for this task and correctly deferred to the sibling `us003_fe_mermaid_diagram` task per `Blocked by:` chain.

**Architecture conformance (D-004 + §"Markdown rendering plan"):**
- `MarkdownRenderer.tsx:204-207` — `rehypePlugins=[[rehypeSanitize, extendedSchema], [rehypeHighlight, rehypeHighlightOpts]]`. Sanitize FIRST, highlight SECOND — matches the architecture pipeline diagram exactly. If swapped, the sanitizer would strip the `hljs`/`language-*` classes the highlighter adds.
- `MarkdownRenderer.tsx:38-94` — `extendedSchema` patches `defaultSchema` to allow `language-*`/`hljs(-*)` on `<code>`/`<pre>`/`<span>`, `target`/`rel` on `<a>`, and the SVG element set mermaid will emit. `<script>` is NOT added to `tagNames`; no `on*` attribute is added to any element's allow-list. javascript:/vbscript:/data:text/html URI rejection comes free from `defaultSchema`'s built-in URL protocol allow-list (verified by FCT-US014-013).
- `MarkdownRenderer.tsx:96-107` — `rehypeHighlight` configured with the curated language subset (Go/TS/JSON/Bash/SQL/YAML/Markdown) per the architecture's bundle-size posture.
- `MarkdownRenderer.tsx:153-174` — anchor override sets `target=_blank rel=noopener noreferrer` only for `http(s)` prefixes; relative/`mailto:` left untouched.
- `MarkdownRenderer.tsx:129-141` — `code` override is the stub the task specified; it renders default `<code>` and leaves the `language-mermaid` branch open for the sibling mermaid task to fill in.
- No `dangerouslySetInnerHTML` anywhere — confirmed.
- `MarkdownErrorBoundary.tsx` — proper React class boundary (`getDerivedStateFromError` + `componentDidCatch`); renders `<div role="alert">Failed to render document</div>` fallback.
- `DocumentPreviewer.tsx:107-109` — body correctly swapped to `<MarkdownErrorBoundary><MarkdownRenderer source={document.content} /></MarkdownErrorBoundary>`. Prop surface unchanged. The `key={document.id}` invariant lives at the DocumentsTab call site (US013) and is preserved.
- `pages/_app.tsx:3` — `highlight.js/styles/github.css` imported once globally per the task's CSS-theme choice. Mocked in Jest via `__mocks__/styleMock.js` (sensible).

**Hard invariants:**
- CSR-only: no `getServerSideProps`/`getStaticProps`/`getInitialProps` introduced; no `web/pages/api/*` added. `MarkdownRenderer` is a pure function component with no SSR coupling.
- No `dangerouslySetInnerHTML` in the markdown renderer.
- No `any` introduced (typecheck clean; explicit interfaces for `CodeProps`, `MarkdownRendererProps`, `MarkdownErrorBoundaryProps`/`State`). The two `as Components['code']` / `as Components['a']` casts at lines 180-181 are narrow and well-justified by the comment immediately above them (react-markdown's typed `ComponentProps` differ slightly from raw HTML element props); these are not `any` leaks.
- Scope discipline: no edits under `services/`; DocumentsTab / DocumentSidebar / hooks / API client untouched. Mermaid is correctly deferred.

**Cosmetic deviation (not a finding):**
- The scope list mentioned a separate `MarkdownErrorBoundary.test.tsx` file; the dev folded both boundary cases (FCT-US014-015 + a passthrough case) into `MarkdownRenderer.test.tsx`. Both required cases are present and green — the contract is the FCT IDs, not the file split. Approved as-is.

**Jest config / ESM transform broadening (`jest.config.js`):** the dev extended `transformIgnorePatterns` to whitelist the entire remark/rehype/micromark/hast/unist ecosystem so SWC transpiles their ESM. This is a one-time scaffold cost that future markdown-stack additions will inherit. Acceptable and well-documented in the task notes.

Verdict: **approved**. Status flipped to `completed`. No follow-up required for this task; mermaid wiring is the next sibling task's responsibility per the existing `Blocked by:` chain.

