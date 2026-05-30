/**
 * Tests for MarkdownRenderer component and MarkdownErrorBoundary
 *
 * FCT-US003-001: Headings render as h1/h2/h3
 * FCT-US003-002: Bold, italic, strikethrough, inline code
 * FCT-US003-003: GFM task list renders disabled checkboxes
 * FCT-US003-004: GFM table renders semantic <table>
 * FCT-US003-005: Safe links + javascript: href sanitization
 * FCT-US003-006: Fenced code block with language gets language-X + hljs classes and <span> tokens
 * FCT-US003-007: Code fence with no language tag renders plain <pre><code> without error
 * FCT-US003-012: XSS — <script>alert(1)</script> not added to DOM
 * FCT-US003-013: XSS — javascript: href is sanitized
 * FCT-US003-014: XSS — onerror= attribute stripped
 * FCT-US003-015: Error boundary: unhandled throw shows "Failed to render document" fallback
 */
import React from 'react';
import { render, screen } from '@testing-library/react';
import { MarkdownRenderer } from './MarkdownRenderer';
import { MarkdownErrorBoundary } from './MarkdownErrorBoundary';

// Suppress error boundary console errors in test output
beforeEach(() => {
  jest.spyOn(console, 'error').mockImplementation(() => {});
});
afterEach(() => {
  (console.error as jest.Mock).mockRestore();
});

// ============================================================
// FCT-US003-001 — MarkdownRenderer: headings render as h1, h2, h3
// ============================================================
describe('FCT-US003-001 — MarkdownRenderer: headings render as h1/h2/h3', () => {
  it('renders h1, h2, h3 heading elements with matching text', () => {
    const source = '# Heading One\n\n## Heading Two\n\n### Heading Three\n';
    const { container } = render(<MarkdownRenderer source={source} />);

    expect(screen.getByRole('heading', { level: 1, name: /Heading One/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 2, name: /Heading Two/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 3, name: /Heading Three/i })).toBeInTheDocument();

    // Container should exist and have rendered content
    expect(container.querySelector('h1')).not.toBeNull();
    expect(container.querySelector('h2')).not.toBeNull();
    expect(container.querySelector('h3')).not.toBeNull();
  });
});

// ============================================================
// FCT-US003-002 — MarkdownRenderer: bold, italic, strikethrough, inline code
// ============================================================
describe('FCT-US003-002 — MarkdownRenderer: bold, italic, strikethrough, inline code', () => {
  it('renders <strong>, <em>, <del>, and inline <code>', () => {
    const source = '**bold** *italic* ~~strikethrough~~ `inline code`';
    const { container } = render(<MarkdownRenderer source={source} />);

    const strong = container.querySelector('strong');
    expect(strong).not.toBeNull();
    expect(strong!.textContent).toBe('bold');

    const em = container.querySelector('em');
    expect(em).not.toBeNull();
    expect(em!.textContent).toBe('italic');

    const del = container.querySelector('del');
    expect(del).not.toBeNull();
    expect(del!.textContent).toBe('strikethrough');

    // Inline code: <code> not inside <pre>
    const inlineCode = container.querySelector('code:not(pre code)');
    expect(inlineCode).not.toBeNull();
    expect(inlineCode!.textContent).toBe('inline code');
  });
});

// ============================================================
// FCT-US003-003 — MarkdownRenderer: GFM task list renders disabled checkboxes
// ============================================================
describe('FCT-US003-003 — MarkdownRenderer: GFM task list renders disabled checkboxes', () => {
  it('renders checked and unchecked disabled checkboxes for task list items', () => {
    const source = '- [ ] Unchecked task\n- [x] Checked task\n';
    const { container } = render(<MarkdownRenderer source={source} />);

    const checkboxes = container.querySelectorAll('input[type="checkbox"]');
    expect(checkboxes).toHaveLength(2);

    // Both should be disabled (read-only previewer)
    checkboxes.forEach((cb) => {
      expect(cb).toBeDisabled();
    });

    // Check the unchecked state
    const unchecked = container.querySelector('input[type="checkbox"]:not(:checked)');
    expect(unchecked).not.toBeNull();

    // Check the checked state
    const checked = container.querySelector('input[type="checkbox"]:checked');
    expect(checked).not.toBeNull();
  });
});

// ============================================================
// FCT-US003-004 — MarkdownRenderer: GFM table renders semantic <table>
// ============================================================
describe('FCT-US003-004 — MarkdownRenderer: GFM table renders semantic <table>', () => {
  it('renders a table with thead, tbody, column headers and data cells', () => {
    const source = '| Name | Age |\n|------|-----|\n| Alice | 30 |\n| Bob | 25 |\n';
    const { container } = render(<MarkdownRenderer source={source} />);

    expect(container.querySelector('table')).not.toBeNull();
    expect(container.querySelector('thead')).not.toBeNull();
    expect(container.querySelector('tbody')).not.toBeNull();

    expect(screen.getByRole('columnheader', { name: /Name/i })).toBeInTheDocument();
    expect(screen.getByRole('columnheader', { name: /Age/i })).toBeInTheDocument();

    // 2 data rows × 2 columns = 4 cells
    const cells = screen.getAllByRole('cell');
    expect(cells).toHaveLength(4);
  });
});

// ============================================================
// FCT-US003-005 — MarkdownRenderer: safe links + javascript: href sanitization
// ============================================================
describe('FCT-US003-005 — MarkdownRenderer: safe links + javascript: href sanitization', () => {
  it('external link has target=_blank rel=noopener noreferrer, javascript: href is stripped', () => {
    const source = '[safe link](https://example.com)\n\n[unsafe link](javascript:alert(1))\n';
    const { container } = render(<MarkdownRenderer source={source} />);

    // Safe link assertions
    const safeLink = screen.getByRole('link', { name: /safe link/i });
    expect(safeLink).toBeInTheDocument();
    expect(safeLink).toHaveAttribute('href', 'https://example.com');
    expect(safeLink).toHaveAttribute('target', '_blank');
    const rel = safeLink.getAttribute('rel') ?? '';
    expect(rel).toContain('noopener');
    expect(rel).toContain('noreferrer');

    // No <a> element should have a javascript: href
    const allLinks = container.querySelectorAll('a');
    allLinks.forEach((a) => {
      const href = a.getAttribute('href') ?? '';
      expect(href.toLowerCase().startsWith('javascript:')).toBe(false);
    });
  });
});

// ============================================================
// FCT-US003-006 — MarkdownRenderer: fenced code block with language gets language-X + hljs classes
// ============================================================
describe('FCT-US003-006 — MarkdownRenderer: fenced code block with language tag', () => {
  it('renders <pre><code> with language-go and hljs class, plus <span> tokens', () => {
    const source = '```go\nfunc main() { fmt.Println("hi") }\n```\n';
    const { container } = render(<MarkdownRenderer source={source} />);

    const codeBlock = container.querySelector('pre > code');
    expect(codeBlock).not.toBeNull();

    // Should have language-go class
    expect(codeBlock!.className).toMatch(/language-go/);

    // Should have hljs class (applied by rehype-highlight)
    expect(codeBlock!.className).toMatch(/hljs/);

    // Should have at least one <span> token from highlighting
    const spans = container.querySelectorAll('pre > code span');
    expect(spans.length).toBeGreaterThan(0);
  });
});

// ============================================================
// FCT-US003-007 — MarkdownRenderer: code fence with no language tag
// ============================================================
describe('FCT-US003-007 — MarkdownRenderer: code fence with no language tag', () => {
  it('renders plain <pre><code> without hljs class and without throwing', () => {
    const source = '```\nplain text code\n```\n';
    const { container } = render(<MarkdownRenderer source={source} />);

    const codeBlock = container.querySelector('pre > code');
    expect(codeBlock).not.toBeNull();

    // Should NOT have hljs class (no highlighting applied for unlabeled fences)
    expect(codeBlock!.className).not.toMatch(/hljs/);
  });
});

// ============================================================
// FCT-US003-012 — MarkdownRenderer XSS: <script> not added to DOM
// ============================================================
describe('FCT-US003-012 — MarkdownRenderer XSS: <script>alert(1)</script> not added to DOM', () => {
  it('strips script tags and does not call alert', () => {
    const alertSpy = jest.spyOn(window, 'alert').mockImplementation(() => {});

    const source = '<script>alert(1)</script>\n\nSafe paragraph after the script tag.';
    const { container } = render(<MarkdownRenderer source={source} />);

    // No <script> element in the rendered container
    expect(container.querySelectorAll('script')).toHaveLength(0);

    // alert must not have been called
    expect(alertSpy).not.toHaveBeenCalled();

    // The safe paragraph text should still appear
    expect(screen.getByText(/Safe paragraph after the script tag\./i)).toBeInTheDocument();

    alertSpy.mockRestore();
  });
});

// ============================================================
// FCT-US003-013 — MarkdownRenderer XSS: javascript: href is sanitized
// ============================================================
describe("FCT-US003-013 — MarkdownRenderer XSS: javascript: href is sanitized", () => {
  it('does not render any <a> with href starting with javascript:', () => {
    const source = "[click me](javascript:alert('xss'))";
    const { container } = render(<MarkdownRenderer source={source} />);

    const allLinks = container.querySelectorAll('a');
    allLinks.forEach((a) => {
      const href = a.getAttribute('href') ?? '';
      expect(href.toLowerCase().startsWith('javascript:')).toBe(false);
    });
  });
});

// ============================================================
// FCT-US003-014 — MarkdownRenderer XSS: onerror= attribute stripped
// ============================================================
describe('FCT-US003-014 — MarkdownRenderer XSS: onerror= attribute stripped', () => {
  it('strips onerror attribute and does not call alert', () => {
    const alertSpy = jest.spyOn(window, 'alert').mockImplementation(() => {});

    const source = '<img src="https://example.com/img.png" onerror="alert(\'xss\')" alt="test">';
    const { container } = render(<MarkdownRenderer source={source} />);

    // No element should have an onerror attribute
    const withOnerror = container.querySelectorAll('[onerror]');
    expect(withOnerror).toHaveLength(0);

    // alert must not have been called
    expect(alertSpy).not.toHaveBeenCalled();

    alertSpy.mockRestore();
  });
});

// ============================================================
// FCT-US003-015 — MarkdownErrorBoundary: unhandled throw shows fallback
// ============================================================
describe('FCT-US003-015 — MarkdownErrorBoundary: shows fallback on thrown error', () => {
  it('renders "Failed to render document" when child throws', () => {
    // Component that always throws during render
    const ThrowingComponent = (): React.ReactElement => {
      throw new Error('Render failure');
    };

    render(
      <MarkdownErrorBoundary>
        <ThrowingComponent />
      </MarkdownErrorBoundary>
    );

    expect(screen.getByText(/Failed to render document/i)).toBeInTheDocument();
  });

  it('renders children normally when no error is thrown', () => {
    render(
      <MarkdownErrorBoundary>
        <div>Normal content</div>
      </MarkdownErrorBoundary>
    );

    expect(screen.getByText('Normal content')).toBeInTheDocument();
  });
});
