/**
 * MarkdownRenderer — renders raw GFM markdown as a sanitized React tree.
 *
 * Pipeline (order is load-bearing per architecture §"Markdown rendering plan"):
 *   remark-parse → remark-gfm → remark-rehype → rehype-sanitize → rehype-highlight → React
 *
 * Key points:
 * - rehype-sanitize runs BEFORE rehype-highlight so that hljs/language-* class names
 *   added by the highlighter are preserved in the allow-list.
 * - The sanitization schema extends `defaultSchema` to allow:
 *     • className with `language-*` / `hljs` / `hljs-*` prefixes on `<code>`, `<pre>`, `<span>`
 *     • `target` and `rel` on `<a>` for safe external link opening
 *     • SVG element set emitted by mermaid (wired in the next task)
 *   and explicitly does NOT allow <script>, on*= event handlers, or javascript:/vbscript: URIs.
 * - The `code` component override is a stub for now; the mermaid task wires the
 *   `language-mermaid` branch into this override.
 * - Does NOT use `dangerouslySetInnerHTML` (architecture D-004).
 *
 * Architecture ref: §"Components → Frontend" row MarkdownRenderer.tsx; D-004; §"Markdown rendering plan".
 */
import React from 'react';
import ReactMarkdown, { Components } from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeSanitize, { defaultSchema } from 'rehype-sanitize';
import rehypeHighlight from 'rehype-highlight';

// Curated language subset per architecture's bundle-size posture.
import langGo from 'highlight.js/lib/languages/go';
import langTs from 'highlight.js/lib/languages/typescript';
import langJson from 'highlight.js/lib/languages/json';
import langBash from 'highlight.js/lib/languages/bash';
import langSql from 'highlight.js/lib/languages/sql';
import langYaml from 'highlight.js/lib/languages/yaml';
import langMd from 'highlight.js/lib/languages/markdown';

// Extended sanitization schema — starts from the safe default and adds the allow-list
// entries required for syntax-highlighted code blocks and mermaid-generated SVGs.
const extendedSchema = {
  ...defaultSchema,
  attributes: {
    ...defaultSchema.attributes,
    // Allow highlight.js classes on code blocks
    code: [
      ...(defaultSchema.attributes?.code ?? []),
      ['className', /^language-/, 'hljs', /^hljs/],
    ],
    pre: [
      ...(defaultSchema.attributes?.pre ?? []),
      ['className', /^language-/, 'hljs', /^hljs/],
    ],
    span: [
      ...(defaultSchema.attributes?.span ?? []),
      ['className', /^hljs/],
    ],
    // Allow target + rel on anchors so external links can open safely in a new tab
    a: [
      ...(defaultSchema.attributes?.a ?? []),
      'target',
      'rel',
    ],
    // SVG element allow-list for mermaid (next task wires MermaidDiagram)
    svg: ['xmlns', 'viewBox', 'width', 'height', 'role', 'aria-label', 'className'],
    g: ['transform', 'className', 'style'],
    path: ['d', 'fill', 'stroke', 'strokeWidth', 'className', 'style'],
    rect: ['x', 'y', 'width', 'height', 'rx', 'ry', 'fill', 'stroke', 'className', 'style'],
    circle: ['cx', 'cy', 'r', 'fill', 'stroke', 'className', 'style'],
    line: ['x1', 'y1', 'x2', 'y2', 'stroke', 'className', 'style'],
    polyline: ['points', 'fill', 'stroke', 'className', 'style'],
    polygon: ['points', 'fill', 'stroke', 'className', 'style'],
    text: ['x', 'y', 'dx', 'dy', 'textAnchor', 'className', 'style'],
    tspan: ['x', 'y', 'dx', 'dy', 'className', 'style'],
    defs: [],
    marker: ['id', 'viewBox', 'refX', 'refY', 'markerWidth', 'markerHeight', 'orient', 'className'],
    title: [],
    desc: [],
  },
  tagNames: [
    ...(defaultSchema.tagNames ?? []),
    'svg',
    'g',
    'path',
    'rect',
    'circle',
    'line',
    'polyline',
    'polygon',
    'text',
    'tspan',
    'defs',
    'marker',
    'title',
    'desc',
  ],
};

const rehypeHighlightOpts = {
  ignoreMissing: true,
  languages: {
    go: langGo,
    typescript: langTs,
    json: langJson,
    bash: langBash,
    sql: langSql,
    yaml: langYaml,
    markdown: langMd,
  },
};

interface CodeProps {
  className?: string;
  children?: React.ReactNode;
}

/**
 * Custom `code` component override.
 *
 * In react-markdown v9, the `code` component is called for BOTH inline code
 * and block code (fenced). For block code, react-markdown's `pre` component
 * wraps the output — so this component should only return `<code>`, never
 * wrapping in `<pre>` itself.
 *
 * Inline code has no className; block code from `rehype-highlight` carries
 * `className="hljs language-*"`.
 *
 * STUB: the `language-mermaid` branch is intentionally a no-op for now —
 * it renders as a normal code block. The mermaid task
 * (`us003_fe_mermaid_diagram`) will replace this branch with <MermaidDiagram>.
 */
const CodeBlock = ({
  className,
  children,
  ...rest
}: CodeProps & React.HTMLAttributes<HTMLElement>): React.ReactElement => {
  // language-mermaid stub: render as plain code for now.
  // The next task will intercept this and route to <MermaidDiagram>.
  return (
    <code className={className} {...rest}>
      {children}
    </code>
  );
};

/**
 * Custom `a` (anchor) component override.
 *
 * External links (href starts with "http") get `target="_blank"` and
 * `rel="noopener noreferrer"` for safe new-tab opening.
 * Internal / relative / `mailto:` links are left as-is.
 *
 * Note: the sanitizer already strips `javascript:` URIs upstream, so this
 * override only handles the opening behaviour for legitimate external links.
 */
const Anchor = ({
  href,
  children,
  ...rest
}: React.AnchorHTMLAttributes<HTMLAnchorElement>): React.ReactElement => {
  const isExternal =
    typeof href === 'string' && (href.startsWith('http://') || href.startsWith('https://'));

  if (isExternal) {
    return (
      <a href={href} target="_blank" rel="noopener noreferrer" {...rest}>
        {children}
      </a>
    );
  }

  return (
    <a href={href} {...rest}>
      {children}
    </a>
  );
};

const markdownComponents: Components = {
  // Cast needed because react-markdown's typed ComponentProps differ slightly
  // from the raw HTML element props; the cast is safe here since we are merely
  // passing through to standard HTML elements with controlled extra props.
  code: CodeBlock as Components['code'],
  a: Anchor as Components['a'],
};

interface MarkdownRendererProps {
  /** Raw GFM markdown string to render. */
  source: string;
}

/**
 * MarkdownRenderer component.
 *
 * Renders a GFM markdown string into a sanitized React tree using:
 * - `react-markdown` (no dangerouslySetInnerHTML)
 * - `remark-gfm` for tables, task lists, strikethrough, autolinks
 * - `rehype-sanitize` for XSS protection (allow-listed schema)
 * - `rehype-highlight` for syntax highlighting (curated language subset)
 *
 * Should be wrapped by `<MarkdownErrorBoundary>` at the call site.
 */
export const MarkdownRenderer = ({ source }: MarkdownRendererProps): React.ReactElement => {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      rehypePlugins={[
        [rehypeSanitize, extendedSchema],
        [rehypeHighlight, rehypeHighlightOpts],
      ]}
      components={markdownComponents}
    >
      {source}
    </ReactMarkdown>
  );
};
