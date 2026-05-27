# US003 — Frontend component test specification

**For FE Dev:** these are the tests you write FIRST (TDD red). Implement in TypeScript using **Jest + React Testing Library**. Mock the backend at the API client layer (`web/lib/api/`) using **MSW** with handlers that match the architecture's exact JSON request/response shapes.

## Coverage matrix

| AC / UI flow | Test ID | Component / hook under test | What it asserts |
|---|---|---|---|
| GFM: headings render as `<h1>` / `<h2>` / `<h3>` | FCT-US003-001 | `web/components/ProjectDetail/MarkdownRenderer.tsx` | correct heading elements with matching text |
| GFM: bold, italic, strikethrough, inline code | FCT-US003-002 | `web/components/ProjectDetail/MarkdownRenderer.tsx` | `<strong>`, `<em>`, `<del>`, inline `<code>` present |
| GFM: task list renders disabled checkboxes | FCT-US003-003 | `web/components/ProjectDetail/MarkdownRenderer.tsx` | checked and unchecked `<input type="checkbox" disabled>` present |
| GFM: table renders `<table>` with `<thead>` and `<tbody>` | FCT-US003-004 | `web/components/ProjectDetail/MarkdownRenderer.tsx` | semantic table elements present with correct cell text |
| GFM: links — safe `<a>` with `target="_blank" rel="noopener noreferrer"` | FCT-US003-005 | `web/components/ProjectDetail/MarkdownRenderer.tsx` | link href correct; `target` and `rel` set; `javascript:` href sanitized away |
| Code fence: syntax-highlighted block has `language-X` + `hljs` classes and `<span>` tokens | FCT-US003-006 | `web/components/ProjectDetail/MarkdownRenderer.tsx` | `<pre><code class="hljs language-go">` + child `<span>` elements |
| Code fence: no language tag renders plain `<pre><code>` without error | FCT-US003-007 | `web/components/ProjectDetail/MarkdownRenderer.tsx` | `<pre><code>` present; no crash; no `hljs` class |
| Mermaid fence: renders `<svg>` in place of raw source | FCT-US003-008 | `web/components/ProjectDetail/MermaidDiagram.tsx` | `<svg>` present; raw mermaid source text NOT visible to user |
| Invalid mermaid: renders error fallback without crashing | FCT-US003-009 | `web/components/ProjectDetail/MermaidDiagram.tsx` | error fallback text visible; page does not throw; rest of doc renders |
| Mermaid: lazy-loaded — dynamic import not executed until a mermaid fence present | FCT-US003-010 | `web/components/ProjectDetail/MermaidDiagram.tsx` | mermaid dynamic import mock is NOT called for non-mermaid docs |
| Document switch resets mermaid: no stale SVG from prior document | FCT-US003-011 | `web/components/ProjectDetail/DocumentPreviewer.tsx` | after key change, old SVG gone; new SVG present |
| XSS: `<script>alert(1)</script>` in content — no `<script>` in DOM | FCT-US003-012 | `web/components/ProjectDetail/MarkdownRenderer.tsx` | no `<script>` element in container; no alert fired |
| XSS: `javascript:` href is sanitized — not rendered as `href` | FCT-US003-013 | `web/components/ProjectDetail/MarkdownRenderer.tsx` | link with `javascript:` URI does not appear as an `<a href="javascript:...">` |
| XSS: `onerror=` attribute stripped from rendered elements | FCT-US003-014 | `web/components/ProjectDetail/MarkdownRenderer.tsx` | no `onerror` attribute on any DOM element in the rendered output |
| Error boundary: thrown error in MarkdownRenderer shows fallback | FCT-US003-015 | `web/components/ProjectDetail/MarkdownRenderer.tsx` (MarkdownErrorBoundary) | "Failed to render document" fallback shown instead of crashing page |

## Component tests

### FCT-US003-001 — MarkdownRenderer: headings render as `<h1>`, `<h2>`, `<h3>`
- **Component / hook under test:** `web/components/ProjectDetail/MarkdownRenderer.tsx`
- **Render with:**
  ```
  source = "# Heading One\n\n## Heading Two\n\n### Heading Three\n"
  ```
- **MSW handlers:** none (pure rendering component).
- **User interactions (RTL):** none.
- **Expect:**
  - `screen.getByRole('heading', { level: 1, name: /Heading One/i })` present.
  - `screen.getByRole('heading', { level: 2, name: /Heading Two/i })` present.
  - `screen.getByRole('heading', { level: 3, name: /Heading Three/i })` present.
- **Architecture cite:** US003 AC "Headings render as headings"; D-004 — `react-markdown` + `remark-gfm`.

### FCT-US003-002 — MarkdownRenderer: bold, italic, strikethrough, inline code
- **Component / hook under test:** `web/components/ProjectDetail/MarkdownRenderer.tsx`
- **Render with:**
  ```
  source = "**bold** *italic* ~~strikethrough~~ `inline code`"
  ```
- **MSW handlers:** none.
- **User interactions (RTL):** none.
- **Expect:**
  - `container.querySelector('strong')` text equals `"bold"`.
  - `container.querySelector('em')` text equals `"italic"`.
  - `container.querySelector('del')` text equals `"strikethrough"` (GFM strikethrough via `remark-gfm`).
  - `container.querySelector('code:not(pre code)')` text equals `"inline code"` (inline code, not inside a `<pre>`).
- **Architecture cite:** US003 AC "Paragraphs, bold, italic, and inline code render"; §"Rendering coverage required".

### FCT-US003-003 — MarkdownRenderer: GFM task list renders disabled checkboxes
- **Component / hook under test:** `web/components/ProjectDetail/MarkdownRenderer.tsx`
- **Render with:**
  ```
  source = "- [ ] Unchecked task\n- [x] Checked task\n"
  ```
- **MSW handlers:** none.
- **User interactions (RTL):** none.
- **Expect:**
  - Two `<input type="checkbox">` elements are present.
  - The checkbox for "Unchecked task" has `checked === false`.
  - The checkbox for "Checked task" has `checked === true`.
  - Both checkboxes are `disabled` (the user should not be able to interact with them — read-only previewer).
- **Architecture cite:** US003 AC "Lists render (ordered + unordered + task lists)"; `remark-gfm` task list extension.

### FCT-US003-004 — MarkdownRenderer: GFM table renders semantic `<table>`
- **Component / hook under test:** `web/components/ProjectDetail/MarkdownRenderer.tsx`
- **Render with:**
  ```
  source = "| Name | Age |\n|------|-----|\n| Alice | 30 |\n| Bob | 25 |\n"
  ```
- **MSW handlers:** none.
- **User interactions (RTL):** none.
- **Expect:**
  - `container.querySelector('table')` is not null.
  - `container.querySelector('thead')` is not null.
  - `container.querySelector('tbody')` is not null.
  - `screen.getByRole('columnheader', { name: /Name/i })` present.
  - `screen.getByRole('columnheader', { name: /Age/i })` present.
  - `screen.getAllByRole('cell')` has 4 elements (2 data rows × 2 columns).
- **Architecture cite:** US003 AC "Tables render"; §"Rendering coverage required".

### FCT-US003-005 — MarkdownRenderer: safe links + `javascript:` href sanitization
- **Component / hook under test:** `web/components/ProjectDetail/MarkdownRenderer.tsx`
- **Render with:**
  ```
  source = "[safe link](https://example.com)\n\n[unsafe link](javascript:alert(1))\n"
  ```
- **MSW handlers:** none.
- **User interactions (RTL):** none.
- **Expect:**
  - Safe link: `screen.getByRole('link', { name: /safe link/i })` has `href = "https://example.com"` and `target = "_blank"` and `rel` containing `"noopener"` and `"noreferrer"`.
  - Unsafe link: no `<a>` element in the DOM has `href` starting with `javascript:` (use `container.querySelectorAll('a')` and assert none have `href` attribute starting with `javascript:`).
- **Architecture cite:** US003 AC "Links render and are safe"; §"Markdown rendering plan" — `rehype-sanitize` "does NOT allow ... `javascript:` URIs"; D-004.

### FCT-US003-006 — MarkdownRenderer: fenced code block with language gets `language-X` + `hljs` classes and `<span>` tokens
- **Component / hook under test:** `web/components/ProjectDetail/MarkdownRenderer.tsx`
- **Render with:**
  ````
  source = "```go\nfunc main() { fmt.Println(\"hi\") }\n```\n"
  ````
- **MSW handlers:** none.
- **User interactions (RTL):** none.
- **Expect:**
  - `container.querySelector('pre > code')` exists.
  - That `<code>` element has class `language-go` (or `hljs language-go` — both acceptable).
  - That `<code>` element has class `hljs`.
  - `container.querySelectorAll('pre > code span').length > 0` (syntax highlighting produced at least one `<span>` token inside the code block).
- **Architecture cite:** US003 AC "Code fences render syntax-highlighted"; D-004 — `rehype-highlight` "emits the `language-xxx` + `hljs` classes that the US003 AC literally tests for".

### FCT-US003-007 — MarkdownRenderer: code fence with no language tag renders plain `<pre><code>` without error
- **Component / hook under test:** `web/components/ProjectDetail/MarkdownRenderer.tsx`
- **Render with:**
  ````
  source = "```\nplain text code\n```\n"
  ````
- **MSW handlers:** none.
- **User interactions (RTL):** none.
- **Expect:**
  - `container.querySelector('pre > code')` exists.
  - `container.querySelector('pre > code')` does NOT have `hljs` class (no highlighting applied).
  - No error thrown during render.
- **Architecture cite:** US003 AC "a code block with no language tag renders as a plain `<pre><code>` without highlighting (and without throwing)".

### FCT-US003-008 — MermaidDiagram: mermaid fence renders `<svg>`; raw source not visible
- **Component / hook under test:** `web/components/ProjectDetail/MermaidDiagram.tsx`
- **Render with:**
  - `source = "graph TD; A-->B; B-->C;"`.
  - Mock `mermaid.render` (via `jest.mock('mermaid')`) to return a valid SVG string: `'<svg><g id="test"></g></svg>'`.
- **MSW handlers:** none.
- **User interactions (RTL):** await render (mermaid render is async).
- **Expect:**
  - `container.querySelector('svg')` is present.
  - The text `"graph TD; A-->B;"` does NOT appear anywhere in the rendered DOM (raw mermaid source replaced by SVG).
- **Architecture cite:** US003 AC "Mermaid diagrams render as SVG"; §"Markdown rendering plan" — mermaid renders to SVG strings, set via React state.

### FCT-US003-009 — MermaidDiagram: invalid mermaid source renders error fallback without crashing
- **Component / hook under test:** `web/components/ProjectDetail/MermaidDiagram.tsx`
- **Render with:**
  - `source = "this is not valid mermaid syntax at all !@#$"`.
  - Mock `mermaid.render` to throw `new Error("Parse error")`.
- **MSW handlers:** none.
- **User interactions (RTL):** await render.
- **Expect:**
  - No unhandled error propagates (component catches the throw internally).
  - `screen.getByRole('alert')` (or `screen.getByText(/Could not render diagram/i)`) visible.
  - The raw source appears in a `<pre>` fallback (so the user can still read the original text).
  - The rest of the document (if `MermaidDiagram` is rendered inside `MarkdownRenderer`) continues to render normally.
- **Architecture cite:** US003 AC "Invalid mermaid source does not crash the page"; §"Mermaid mechanics" — "Catches render errors and renders 'Could not render diagram' + raw source as a `<pre>` fallback".

### FCT-US003-010 — MermaidDiagram: mermaid dynamic import is NOT executed for non-mermaid documents
- **Component / hook under test:** `web/components/ProjectDetail/MermaidDiagram.tsx` (by asserting it is never mounted when there is no mermaid fence)
- **Render with:**
  - `web/components/ProjectDetail/MarkdownRenderer.tsx` with:
    ```
    source = "# No mermaid here\n\n```go\nfmt.Println(\"hi\")\n```\n"
    ```
  - Mock the dynamic import via `jest.mock` — spy on the factory function for `mermaid`.
- **MSW handlers:** none.
- **User interactions (RTL):** render and await.
- **Expect:**
  - The mermaid dynamic import mock factory is NOT called.
  - `container.querySelector('svg')` is null.
- **Notes:** This verifies the lazy-load contract — mermaid is only imported when a `language-mermaid` code block is encountered. The custom `code` component override routes `language-mermaid` to `<MermaidDiagram>`; for all other languages it falls through to the highlighter. If `MermaidDiagram` is never mounted, mermaid's `import()` is never called.
- **Architecture cite:** §"Markdown rendering plan" — "Lazy-loaded so users on `/` (dashboard) and on projects with no mermaid code blocks do not pay the bundle cost"; D-004 — "mermaid loaded via `next/dynamic({ ssr: false })` inside `MermaidDiagram`".

### FCT-US003-011 — DocumentPreviewer: switching documents resets mermaid state (no stale SVG)
- **Component / hook under test:** `web/components/ProjectDetail/DocumentPreviewer.tsx`
- **Render with:**
  - Initially `key = "doc-A"`, `document.content = "```mermaid\ngraph TD; A-->B;\n```"`.
  - Mock `mermaid.render` for doc-A to return `<svg id="svg-A"></svg>`.
- **MSW handlers:** none (document fed via props).
- **User interactions (RTL):**
  1. Assert `container.querySelector('#svg-A')` is present.
  2. `rerender` with `key = "doc-B"`, `document.content = "```mermaid\ngraph TD; X-->Y;\n```"`.
  - Mock `mermaid.render` for doc-B to return `<svg id="svg-B"></svg>`.
  3. Assert again after rerender.
- **Expect:**
  - After rerender with `key="doc-B"`:
    - `container.querySelector('#svg-A')` is null (old SVG is gone — React unmounted the old subtree due to `key` change).
    - `container.querySelector('#svg-B')` is present.
- **Architecture cite:** US003 AC "Switching documents re-renders mermaid for the new content"; §"Mermaid mechanics" — "`<DocumentPreviewer>` a `key={document.id}` from its parent — React unmounts the old subtree, mermaid's old SVGs go with it".

### FCT-US003-012 — MarkdownRenderer XSS: `<script>alert(1)</script>` not added to DOM
- **Component / hook under test:** `web/components/ProjectDetail/MarkdownRenderer.tsx`
- **Render with:**
  ```
  source = "<script>alert(1)</script>\n\nSafe paragraph after the script tag."
  ```
- **MSW handlers:** none.
- **User interactions (RTL):** render.
- **Expect:**
  - `container.querySelectorAll('script').length === 0` — no `<script>` element in the rendered container.
  - `window.alert` is NOT called (install a `jest.spyOn(window, 'alert')` before rendering and assert it was not called).
  - The text "Safe paragraph after the script tag." is still rendered (the sanitizer strips the script, not the whole document).
- **Architecture cite:** US003 AC "XSS — script tags in content are not executed"; §"Markdown rendering plan" — `rehype-sanitize` "explicitly does NOT allow `<script>`"; Risks — "XSS through markdown".

### FCT-US003-013 — MarkdownRenderer XSS: `javascript:` href is sanitized
- **Component / hook under test:** `web/components/ProjectDetail/MarkdownRenderer.tsx`
- **Render with:**
  ```
  source = "[click me](javascript:alert('xss'))"
  ```
- **MSW handlers:** none.
- **User interactions (RTL):** render.
- **Expect:**
  - No `<a>` element has `href` starting with `javascript:` (use `container.querySelectorAll('a')` and assert each `href` attribute does not start with `"javascript:"`).
  - Either the link is rendered without `href` or it is stripped entirely — but NOT rendered with a `javascript:` URI.
- **Architecture cite:** US003 AC "any `javascript:` URIs in the source are sanitized away"; §"Markdown rendering plan" — `rehype-sanitize` "does NOT allow ... `javascript:` / `vbscript:` / `data:text/html` URI prefixes".

### FCT-US003-014 — MarkdownRenderer XSS: `on*=` event handler attribute stripped
- **Component / hook under test:** `web/components/ProjectDetail/MarkdownRenderer.tsx`
- **Render with:**
  ```
  source = '<img src="https://example.com/img.png" onerror="alert(\'xss\')" alt="test">'
  ```
- **MSW handlers:** none.
- **User interactions (RTL):** render.
- **Expect:**
  - `container.querySelectorAll('[onerror]').length === 0` — no element in the rendered output carries an `onerror` attribute.
  - If the `<img>` renders at all (architecture allows `<img>` for the `alt` text rendering), it must not carry `onerror`.
  - `window.alert` not called.
- **Architecture cite:** §"Markdown rendering plan" — `rehype-sanitize` "explicitly does NOT allow ... `on*` event handlers"; Risks — "XSS through markdown".

### FCT-US003-015 — MarkdownRenderer error boundary: unhandled throw shows "Failed to render document" fallback
- **Component / hook under test:** `web/components/ProjectDetail/MarkdownRenderer.tsx` (wrapping `MarkdownErrorBoundary`)
- **Render with:**
  - Mock one of the internal rehype/remark plugins (or the `react-markdown` render) to throw synchronously during render: `jest.mock('react-markdown', () => { throw new Error('Render failure'); })`.
  - Or: trigger via a prop that causes the internal pipeline to throw (implementation-specific).
- **MSW handlers:** none.
- **User interactions (RTL):** render.
- **Expect:**
  - `screen.getByText(/Failed to render document/i)` visible.
  - No unhandled error propagates to the test runner (the error boundary catches it).
  - The rest of the page (sidebar, header) is not affected (error boundary is scoped to the previewer body).
- **Architecture cite:** §"Mermaid mechanics" — "React error boundary (`MarkdownErrorBoundary`) wrapping `<MarkdownRenderer>` so any unexpected throw inside the rendering pipeline shows 'Failed to render document'".
