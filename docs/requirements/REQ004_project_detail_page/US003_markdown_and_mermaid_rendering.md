# US003 — Markdown and mermaid rendering in the previewer

**Requirement:** REQ004 — project_detail_page
**Status:** draft

## Story
As a user viewing a document on the project detail page, I want the previewer to render the document's content as beautifully-formatted markdown — including GitHub-flavored extensions, syntax-highlighted code fences, and mermaid diagrams — so that documents are pleasant and useful to read instead of looking like raw source.

> Scope note: this story upgrades the previewer's internals only. It does NOT touch the sidebar, navigation, fetching, error/loading states, or routing — those are owned by US002. From the user's point of view, US003 changes "what the content area looks like" once a document has loaded.

## Acceptance criteria

- **Scenario: Headings render as headings**
  - Given a document whose `content` includes `# Heading 1`, `## Heading 2`, and `### Heading 3` lines
  - When the previewer renders that document
  - Then the rendered DOM contains an `<h1>`, an `<h2>`, and an `<h3>` with matching text
  - And visual hierarchy is preserved (h1 is largest, h3 is smallest)

- **Scenario: Paragraphs, bold, italic, and inline code render**
  - Given a document whose content includes a paragraph with `**bold**`, `*italic*`, and `` `inline code` ``
  - When the previewer renders that document
  - Then the rendered DOM contains a `<strong>`, an `<em>`, and an inline `<code>` with the corresponding text
  - And the inline `<code>` is visually distinguishable (monospace + subtle background)

- **Scenario: Lists render (ordered + unordered + task lists)**
  - Given a document with a bullet list, a numbered list, and a GFM task list (`- [ ]` / `- [x]`)
  - When the previewer renders
  - Then unordered lists render as `<ul><li>`, ordered as `<ol><li>`, and task list items render with a (disabled) checkbox reflecting checked/unchecked state

- **Scenario: Tables render**
  - Given a document with a GFM pipe-table
  - When the previewer renders
  - Then the rendered DOM contains a `<table>` with `<thead>`, `<tbody>`, and cells matching the source

- **Scenario: Links render and are safe**
  - Given a document with `[text](https://example.com)`
  - When the previewer renders
  - Then the rendered DOM contains an `<a>` with `href="https://example.com"` and text "text"
  - And external links open in a new tab (`target="_blank"` with `rel="noopener noreferrer"`)
  - And any `javascript:` URIs in the source are sanitized away (do not appear as `href` attributes in the rendered DOM)

- **Scenario: Code fences render syntax-highlighted**
  - Given a document with a fenced code block tagged with a language, e.g.:
    ````
    ```go
    func main() { fmt.Println("hi") }
    ```
    ````
  - When the previewer renders
  - Then the code block renders inside a `<pre><code>` structure
  - And the `<code>` element carries a class indicating the language (e.g. `language-go` or `hljs language-go`)
  - And tokens inside are wrapped in `<span>` elements with highlight classes (i.e. syntax highlighting is actually applied, not just CSS-class noise)
  - And a code block with no language tag renders as a plain `<pre><code>` without highlighting (and without throwing)

- **Scenario: Mermaid diagrams render as SVG**
  - Given a document with a fenced code block tagged `mermaid`:
    ````
    ```mermaid
    graph TD; A-->B; B-->C;
    ```
    ````
  - When the previewer renders that document
  - Then a `<svg>` element corresponding to the diagram appears in the DOM in place of the code block
  - And the raw mermaid source is NOT shown to the user (it has been replaced by the rendered diagram)
  - And the diagram is visually centered / sized to fit the previewer width without horizontal overflow on a typical desktop viewport (≥ 1024px)

- **Scenario: Invalid mermaid source does not crash the page**
  - Given a document with a `mermaid` code block whose source is malformed (e.g. random non-mermaid text)
  - When the previewer renders that document
  - Then the rest of the document still renders normally
  - And in place of the diagram, the user sees a small inline error indicator (e.g. "Could not render diagram") and/or the raw mermaid source as a fallback
  - And the application does not throw an unhandled error

- **Scenario: Switching documents re-renders mermaid for the new content**
  - Given the previewer is showing document A which contains a mermaid diagram
  - When I click document B in the sidebar (which contains a different mermaid diagram)
  - Then the previewer shows document B's diagram, not document A's
  - And there are no stale SVGs left over from document A

- **Scenario: XSS — script tags in content are not executed**
  - Given a document whose `content` includes `<script>alert(1)</script>` as raw text
  - When the previewer renders the document
  - Then no `<script>` element is added to the live DOM (the markdown renderer must sanitize HTML)
  - And no `alert` fires (i.e. the test should not observe the side effect)
  - And the raw `<script>` source either appears as escaped text or is stripped entirely — but never executed

## UI / UX flow expectations

- **Entry points:** Same as US002 — user is on the Documents tab and a document is selected. This story changes only what the user sees inside the previewer pane.

- **Happy-path flow (visual upgrade only):**
  1. User selects a document (US002 behavior unchanged).
  2. Previewer fetches content (US002 behavior unchanged).
  3. **NEW in US003:** once content arrives, the previewer renders it as styled markdown with mermaid diagrams and syntax-highlighted code, rather than as plain text / `<pre>`.

- **Rendering coverage required:**
  - Headings `# … ######`
  - Paragraphs, bold, italic, strikethrough, inline code
  - Bulleted and numbered lists, including nested
  - GFM task lists (`- [ ]` / `- [x]`)
  - GFM tables
  - Blockquotes
  - Horizontal rules
  - Links (rendered with safe `target`/`rel` and javascript-URI sanitization)
  - Images (rendered as `<img>` with `alt`)
  - Fenced code blocks: syntax-highlighted when a language is provided
  - Inline code
  - **Mermaid diagrams** for fenced blocks tagged `mermaid`

- **Out of rendering scope (this story):**
  - LaTeX / KaTeX math.
  - Custom GFM alerts (`> [!NOTE]`) — nice-to-have, not required.
  - PlantUML or other diagramming languages.
  - Footnotes — nice-to-have, not required.
  - Full-bleed images, image carousels, lightboxes.

- **Loading / error states:**
  - Loading state for the content fetch itself is unchanged from US002.
  - **Mermaid render in progress:** acceptable to show the code fence briefly before the SVG replaces it, or to show a lightweight placeholder. Must not show a permanent flash of unstyled markdown.
  - **Mermaid render error:** see XSS / invalid-mermaid scenarios above.

- **Validation rules visible to the user:** none (no input).

- **Accessibility expectations:**
  - Code blocks: maintain semantic `<pre><code>` so screen readers handle them correctly.
  - Mermaid SVGs: include an accessible name where possible (e.g. via `<title>` element in the SVG or `aria-label` on a wrapper); at minimum, do not hide the surrounding content.
  - Tables: real `<table>` semantics (not divs styled as tables).

- **Out of UI scope:**
  - Anything outside the previewer pane.
  - Theming (light/dark mode).
  - Print stylesheet.

## Out of scope
- Math rendering (KaTeX).
- Editing documents.
- Custom directives, embeds, OEmbeds.
- Mobile-specific markdown rendering polish.

## Dependencies
- **US002** — provides the previewer container, the document fetch, the selection/loading/error states. US003 replaces the previewer's internal rendering only.
- **No new backend endpoints.** US003 is FE-only (mermaid + markdown rendering is a client-side concern; mermaid renders in the browser and must run via dynamic import to keep the bundle small).

## Notes for the team
- **Library selection is the System Architect's call** in Phase 1. Reasonable choices for the team to evaluate:
  - Markdown: `react-markdown` + `remark-gfm` + `rehype-raw` (with sanitizer) + `rehype-highlight` for code; alternative: `marked` + `DOMPurify`. The choice should keep bundle size reasonable and play nicely with Next.js CSR.
  - Mermaid: the official `mermaid` package, loaded via dynamic import on first use to avoid bloating initial bundle. Either render via a custom code-block renderer plugin or by post-processing rendered markdown HTML.
  - Whatever the architect picks must be SSR-safe in the sense that it doesn't break Next.js build — even though we are CSR-only, builds should not fail on `window`/`document` references at import time. Dynamic import (`next/dynamic` with `ssr: false`) is the standard escape hatch.
- **Sanitization is a hard requirement** because the document content originates from MCP-authored input and could contain anything. The XSS AC above is testable and must pass — do not use `dangerouslySetInnerHTML` without a sanitizer in the chain.
- **Mermaid is the visually heaviest dependency.** Lazy-load it (e.g. `next/dynamic` with `ssr: false`) so projects with no mermaid documents don't pay the cost. The test contract may need to mock or stub mermaid; the tester will design that.
- **Re-rendering across document switches:** mermaid mutates the DOM (it injects SVGs into the markup). When the previewer's content changes, the component must guarantee that prior SVGs are cleaned up — the simplest pattern is to give the renderer a `key={document.id}` so React unmounts and remounts.
- **"Render beautifully" sanity check:** the tester's spec for this story should include not just "elements appear" assertions but also at least one visual / golden-style check that headings have hierarchy spacing and code blocks have a visible background — otherwise a totally unstyled renderer could technically pass.

## Sign-off log
(po-ba appends here on each sign-off pass)
