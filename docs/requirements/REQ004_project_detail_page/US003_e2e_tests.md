# US003 — E2E test specification (Robot Framework)

**Owner:** tester. Implemented in `tests/e2e/REQ004_project_detail_page/US003_markdown_mermaid.robot`.

## Why e2e

The following scenarios cannot be fully verified at the component layer:

1. **Syntax-highlighted code fence rendered in a real browser:** `rehype-highlight` emits CSS classes, but whether the `highlight.js` CSS theme actually loads and whether the resulting DOM looks correct can only be confirmed in a real browser (Playwright). JSDOM-based tests verify class presence, not visual rendering.
2. **Mermaid SVG visible in a real browser:** `mermaid.render()` is called in a real browser DOM (not JSDOM), uses native SVG APIs, and depends on the `next/dynamic` lazy-load executing in a real runtime. The SVG cannot be verified in JSDOM.

XSS sanitization, error fallbacks, and individual rendering correctness assertions live at the unit layer (FCT-US003-012 through FCT-US003-015). They are NOT promoted to e2e.

## Scenarios

### E2E-US003-001 — Previewer renders syntax-highlighted code fence and mermaid SVG for a real document
- **Tag:** US003, smoke, regression
- **Preconditions:**
  - Next.js frontend at `${WEB_BASE_URL}`, `api-server` at `${API_BASE_URL}`, MCP server running.
- **Setup (data):**
  1. Via MCP `create_project`: create project `"REQ004 US003 E2E <random>"`. Record `projectId`.
  2. Via MCP `create_document`: create one document:
     - `title = "Rendering test document"`
     - `content`:
       ```
       # Rendering test

       ## Code block

       ```go
       func main() { fmt.Println("hello") }
       ```

       ## Diagram

       ```mermaid
       graph TD; Start-->End;
       ```
       ```
  - Record `documentId`.
- **Steps:**
  1. Navigate to `${WEB_BASE_URL}/projects/{projectId}?tab=documents&doc={documentId}`.
  2. Wait for the previewer to show the document title "Rendering test document".
  3. Assert an `<h1>` or `<h2>` heading with text "Rendering test" is visible in the previewer.
  4. Assert a `<pre>` element containing a `<code>` element with a CSS class matching `language-go` or `hljs` is present in the previewer DOM.
  5. Wait for the mermaid SVG to appear (mermaid is async — use a generous timeout, e.g. 15 seconds).
  6. Assert an `<svg>` element is present inside the previewer.
  7. Assert the raw mermaid source text `"graph TD; Start-->End;"` is NOT visible as readable text on the page (it has been replaced by the SVG).
- **Expected:**
  - Syntax highlighting classes applied to the Go code fence in a real browser.
  - Mermaid diagram rendered as SVG (not as raw source text).
- **Cleanup:** none.
