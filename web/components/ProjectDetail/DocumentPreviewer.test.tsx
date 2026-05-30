/**
 * Tests for DocumentPreviewer component
 * FCT-US002-008: Loading indicator during content fetch
 * FCT-US002-009: Renders title, updatedAt, and content when loaded
 * FCT-US002-010: In-pane error + Retry when content fetch fails; sidebar unaffected
 * FCT-US002-011: Retry button calls onRetry
 * FCT-US003-011: Switching documents resets mermaid state (no stale SVG)
 */

// Mock mermaid so MermaidDiagram works in tests without the real mermaid library.
jest.mock('mermaid', () => ({
  __esModule: true,
  default: {
    initialize: jest.fn(),
    render: jest.fn().mockResolvedValue({
      svg: '<svg id="svg-placeholder"></svg>',
    }),
  },
}));

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import mermaid from 'mermaid';
import { DocumentPreviewer } from './DocumentPreviewer';
import { Document } from '../../lib/api/types';

const mockDocument: Document = {
  id: 'd111',
  projectId: 'p1',
  title: 'Architecture overview',
  content: '# Architecture\n\nThis project uses…',
  createdAt: '2026-05-18T08:30:00Z',
  updatedAt: '2026-05-20T09:45:00Z',
};

// FCT-US002-008 — Loading indicator during content fetch
describe('FCT-US002-008 — DocumentPreviewer: loading indicator', () => {
  it('shows loading indicator when isLoading is true', () => {
    render(
      <DocumentPreviewer
        document={null}
        isLoading={true}
        error={null}
        isNotFound={false}
        onRetry={jest.fn()}
      />
    );

    // A loading indicator should be visible
    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('does not render sidebar DOM (previewer is isolated)', () => {
    render(
      <DocumentPreviewer
        document={null}
        isLoading={true}
        error={null}
        isNotFound={false}
        onRetry={jest.fn()}
      />
    );

    // Sidebar-specific elements must not appear inside the previewer
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
  });
});

// FCT-US002-009 — Renders title, updatedAt, and content when loaded
describe('FCT-US002-009 — DocumentPreviewer: renders document content', () => {
  it('shows heading, updatedAt date, and content text', () => {
    render(
      <DocumentPreviewer
        document={mockDocument}
        isLoading={false}
        error={null}
        isNotFound={false}
        onRetry={jest.fn()}
      />
    );

    // Title as h2
    expect(screen.getByRole('heading', { name: /Architecture overview/i })).toBeInTheDocument();

    // updatedAt display must include the date
    expect(screen.getByText(/2026-05-20/i)).toBeInTheDocument();

    // Content text visible
    expect(screen.getByText(/This project uses/i)).toBeInTheDocument();
  });
});

// FCT-US002-010 — In-pane error + Retry when content fetch fails; sidebar unaffected
describe('FCT-US002-010 — DocumentPreviewer: in-pane error state', () => {
  it('shows "Failed to load document" and Retry button on error', () => {
    const onRetry = jest.fn();
    render(
      <DocumentPreviewer
        document={null}
        isLoading={false}
        error={new Error('Failed to fetch document')}
        isNotFound={false}
        onRetry={onRetry}
      />
    );

    expect(screen.getByText(/Failed to load document/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Retry/i })).toBeInTheDocument();
  });

  it('does not render sidebar DOM (previewer is isolated)', () => {
    render(
      <DocumentPreviewer
        document={null}
        isLoading={false}
        error={new Error('Failed')}
        isNotFound={false}
        onRetry={jest.fn()}
      />
    );

    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
  });
});

// FCT-US002-011 — Retry button calls onRetry
describe('FCT-US002-011 — DocumentPreviewer: Retry button calls onRetry', () => {
  it('clicking Retry calls onRetry exactly once', async () => {
    const onRetry = jest.fn();
    render(
      <DocumentPreviewer
        document={null}
        isLoading={false}
        error={new Error('Failed')}
        isNotFound={false}
        onRetry={onRetry}
      />
    );

    await userEvent.click(screen.getByRole('button', { name: /Retry/i }));

    expect(onRetry).toHaveBeenCalledTimes(1);
  });
});

// Additional: isNotFound state
describe('DocumentPreviewer: isNotFound state', () => {
  it('shows "Document not found" message when isNotFound is true', () => {
    render(
      <DocumentPreviewer
        document={null}
        isLoading={false}
        error={null}
        isNotFound={true}
        onRetry={jest.fn()}
      />
    );

    expect(screen.getByText(/Document not found/i)).toBeInTheDocument();
    // No Retry button for not-found state
    expect(screen.queryByRole('button', { name: /Retry/i })).not.toBeInTheDocument();
  });
});

// ============================================================
// FCT-US003-011 — DocumentPreviewer: switching documents resets mermaid state
// ============================================================
describe('FCT-US003-011 — DocumentPreviewer: switching documents resets mermaid state', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('after key change old SVG is gone and new SVG is present', async () => {
    // Configure mermaid mock to return svg-A for the first call
    (mermaid.render as jest.Mock).mockResolvedValueOnce({
      svg: '<svg id="svg-A"></svg>',
    });

    const docA: Document = {
      id: 'doc-A',
      projectId: 'p1',
      title: 'Doc A',
      content: '```mermaid\ngraph TD; A-->B;\n```',
      createdAt: '2026-05-18T08:30:00Z',
      updatedAt: '2026-05-20T09:45:00Z',
    };

    const { container, rerender } = render(
      <DocumentPreviewer
        key="doc-A"
        document={docA}
        isLoading={false}
        error={null}
        isNotFound={false}
        onRetry={jest.fn()}
      />
    );

    // Wait for doc-A's SVG to render
    await waitFor(() => {
      expect(container.querySelector('#svg-A')).not.toBeNull();
    });

    // Configure mermaid mock to return svg-B for the second document
    (mermaid.render as jest.Mock).mockResolvedValueOnce({
      svg: '<svg id="svg-B"></svg>',
    });

    const docB: Document = {
      id: 'doc-B',
      projectId: 'p1',
      title: 'Doc B',
      content: '```mermaid\ngraph TD; X-->Y;\n```',
      createdAt: '2026-05-18T08:30:00Z',
      updatedAt: '2026-05-20T09:45:00Z',
    };

    // React's key prop forces unmount + remount, simulating what DocumentsTab does.
    // With RTL we do this by rerendering with a different key via a wrapper.
    rerender(
      <DocumentPreviewer
        key="doc-B"
        document={docB}
        isLoading={false}
        error={null}
        isNotFound={false}
        onRetry={jest.fn()}
      />
    );

    await waitFor(() => {
      expect(container.querySelector('#svg-B')).not.toBeNull();
    });

    // Old SVG must be gone
    expect(container.querySelector('#svg-A')).toBeNull();
  });
});
