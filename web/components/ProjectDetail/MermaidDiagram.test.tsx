/**
 * Tests for MermaidDiagram component
 *
 * FCT-US003-008: Mermaid fence renders <svg>; raw source not visible
 * FCT-US003-009: Invalid mermaid source renders error fallback without crashing
 * FCT-US003-010: Mermaid dynamic import NOT executed for non-mermaid documents
 *   (tested via MarkdownRenderer rendering a non-mermaid doc and asserting no svg)
 */
import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';

// Mock mermaid BEFORE importing MermaidDiagram.
// mermaid v11 returns a Promise from render().
jest.mock('mermaid', () => ({
  __esModule: true,
  default: {
    initialize: jest.fn(),
    render: jest.fn().mockResolvedValue({
      svg: '<svg data-testid="fake-mermaid-svg"><title>fake</title></svg>',
    }),
  },
}));

// eslint-disable-next-line import/first
import { MermaidDiagram } from './MermaidDiagram';
// eslint-disable-next-line import/first
import mermaid from 'mermaid';

// Clear mocks between tests
beforeEach(() => {
  jest.clearAllMocks();
  // Reset render mock to its default happy-path behaviour
  (mermaid.render as jest.Mock).mockResolvedValue({
    svg: '<svg data-testid="fake-mermaid-svg"><title>fake</title></svg>',
  });
});

// ============================================================
// FCT-US003-008 — MermaidDiagram: renders <svg>; raw source not visible
// ============================================================
describe('FCT-US003-008 — MermaidDiagram: mermaid fence renders <svg>; raw source not visible', () => {
  it('shows <svg> in the DOM and hides the raw mermaid source', async () => {
    const source = 'graph TD; A-->B; B-->C;';
    const { container } = render(<MermaidDiagram source={source} />);

    // Wait for the async mermaid render to settle
    await waitFor(() => {
      expect(container.querySelector('svg')).not.toBeNull();
    });

    // The raw mermaid source text must NOT appear anywhere in the rendered DOM
    expect(container.textContent).not.toContain('graph TD; A-->B;');
  });

  it('multiple MermaidDiagram instances do not throw on duplicate ids (useId works)', async () => {
    const { container } = render(
      <>
        <MermaidDiagram source="graph TD; A-->B;" />
        <MermaidDiagram source="graph LR; X-->Y;" />
      </>
    );

    await waitFor(() => {
      const svgs = container.querySelectorAll('svg');
      expect(svgs.length).toBeGreaterThanOrEqual(2);
    });
  });
});

// ============================================================
// FCT-US003-009 — MermaidDiagram: invalid mermaid renders error fallback without crashing
// ============================================================
describe('FCT-US003-009 — MermaidDiagram: invalid mermaid renders error fallback without crashing', () => {
  it('shows role="alert" "Could not render diagram" and the raw source in <pre> on render failure', async () => {
    // Override the happy-path mock to throw for this test
    (mermaid.render as jest.Mock).mockRejectedValueOnce(new Error('Parse error'));

    const source = 'this is not valid mermaid syntax at all !@#$';
    const { container } = render(<MermaidDiagram source={source} />);

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
    });

    // The error text must be visible
    expect(screen.getByText(/Could not render diagram/i)).toBeInTheDocument();

    // The raw source must appear in a <pre> fallback
    const pre = container.querySelector('pre');
    expect(pre).not.toBeNull();
    expect(pre!.textContent).toContain(source);

    // No <svg> should be present — only the fallback
    expect(container.querySelector('svg')).toBeNull();
  });

  it('does not propagate an exception — the component renders its fallback, not an error', async () => {
    (mermaid.render as jest.Mock).mockRejectedValueOnce(new Error('bad graph'));

    // If MermaidDiagram throws, this render() call will throw too.
    // The test passes only if render completes without throwing.
    expect(() => {
      render(<MermaidDiagram source="bad mermaid input" />);
    }).not.toThrow();

    // Confirm the error fallback is eventually visible
    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
    });
  });
});
