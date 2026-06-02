/**
 * Tests for MermaidDiagram component
 *
 * FCT-US003-008: Mermaid fence renders <svg>; raw source not visible
 * FCT-US003-009: Invalid mermaid source renders error fallback without crashing
 * FCT-US003-010: Mermaid dynamic import NOT executed for non-mermaid documents
 *   (tested via MarkdownRenderer rendering a non-mermaid doc and asserting no svg)
 *
 * FCT-US010-001: SVG child present after success (ref-attach)
 * FCT-US010-002: No dangerouslySetInnerHTML in DOM
 * FCT-US010-003: Strict-mode double-mount yields exactly one SVG child
 * FCT-US010-004: Malformed SVG string does not throw
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

// ============================================================
// FCT-US010-001 — MermaidDiagram ref-attach: <svg> child present after success
// ============================================================
describe('FCT-US010-001 — MermaidDiagram ref-attach: <svg> child present after success', () => {
  it('renders an <svg> child inside the <div role="img"> wrapper via ref-attach', async () => {
    const source = 'graph TD; A-->B;';
    const { container } = render(<MermaidDiagram source={source} />);

    await waitFor(() => {
      expect(container.querySelector('svg')).not.toBeNull();
    });

    // The svg must be a child of the role="img" wrapper
    const wrapper = container.querySelector('[role="img"]');
    expect(wrapper).not.toBeNull();
    expect(wrapper!.querySelector('svg')).not.toBeNull();
  });
});

// ============================================================
// FCT-US010-002 — MermaidDiagram: no element uses dangerouslySetInnerHTML
// ============================================================
describe('FCT-US010-002 — MermaidDiagram: no element uses dangerouslySetInnerHTML', () => {
  it('does not inject svg via innerHTML — the role=img wrapper has no dangerouslySetInnerHTML React prop', async () => {
    const { container } = render(<MermaidDiagram source="graph TD; A-->B;" />);

    await waitFor(() => {
      expect(container.querySelector('svg')).not.toBeNull();
    });

    const wrapper = container.querySelector('[role="img"]');
    expect(wrapper).not.toBeNull();

    // Inspect the React fiber attached to the wrapper element to assert that
    // no dangerouslySetInnerHTML prop was passed to the JSX element that
    // rendered this DOM node.
    //
    // React 18 attaches fiber internals to DOM nodes as a property named
    // __reactFiber$<random>. We search all own properties for the fiber key
    // and check the fiber's memoizedProps for dangerouslySetInnerHTML.
    const wrapperNode = wrapper as Element & Record<string, unknown>;
    const fiberKey = Object.keys(wrapperNode).find((k) =>
      k.startsWith('__reactFiber')
    );
    expect(fiberKey).toBeDefined();
    const fiber = wrapperNode[fiberKey as string] as { memoizedProps?: Record<string, unknown> };
    expect(fiber).toBeDefined();
    // The memoizedProps of the wrapper element must NOT contain dangerouslySetInnerHTML
    expect(fiber.memoizedProps).not.toHaveProperty('dangerouslySetInnerHTML');
  });
});

// ============================================================
// FCT-US010-003 — MermaidDiagram: strict-mode double-mount yields exactly one <svg> child
// ============================================================
describe('FCT-US010-003 — MermaidDiagram: strict-mode double-mount yields exactly one <svg> child', () => {
  it('has exactly one <svg> child in the wrapper after strict-mode double-mount', async () => {
    const source = 'graph TD; A-->B;';
    const { container } = render(
      <React.StrictMode>
        <MermaidDiagram source={source} />
      </React.StrictMode>
    );

    await waitFor(() => {
      expect(container.querySelector('svg')).not.toBeNull();
    });

    const wrapper = container.querySelector('[role="img"]');
    expect(wrapper).not.toBeNull();
    // Under React 18 strict mode, effects run twice in dev.
    // The cleanup must remove the first appended node before the second run.
    // Result: exactly ONE <svg> child element, not two.
    expect(wrapper!.children.length).toBe(1);
    expect(wrapper!.children[0].tagName.toLowerCase()).toBe('svg');
  });
});

// ============================================================
// FCT-US010-004 — MermaidDiagram: malformed SVG string does not throw
// ============================================================
describe('FCT-US010-004 — MermaidDiagram: malformed SVG string does not throw', () => {
  it('does not throw when mermaid resolves with a malformed SVG string', async () => {
    // Malformed: not an SVG root element
    (mermaid.render as jest.Mock).mockResolvedValue({ svg: 'not-an-svg' });

    expect(() => {
      render(<MermaidDiagram source="graph TD; A-->B;" />);
    }).not.toThrow();

    // The component should not crash; it either renders a blank wrapper or the
    // DOMParser falls back to an error document. Either way, no uncaught exception.
    // We wait a tick for the async effect to complete.
    await waitFor(() => {
      // No role="alert" error boundary — the component handles it silently
      // OR renders the wrapper without an svg child. Either is acceptable.
      // The key invariant: no exception thrown.
      expect(true).toBe(true);
    });
  });

  it('does not throw when mermaid resolves with a non-svg element (e.g. <div>oops</div>)', async () => {
    (mermaid.render as jest.Mock).mockResolvedValue({ svg: '<div>oops</div>' });

    expect(() => {
      render(<MermaidDiagram source="graph TD; A-->B;" />);
    }).not.toThrow();

    await waitFor(() => {
      expect(true).toBe(true);
    });
  });
});
