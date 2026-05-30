import React from 'react';

interface MarkdownErrorBoundaryProps {
  children: React.ReactNode;
}

interface MarkdownErrorBoundaryState {
  hasError: boolean;
}

/**
 * MarkdownErrorBoundary — React class error boundary for the markdown rendering pipeline.
 *
 * Catches any unhandled throws from inside MarkdownRenderer (or its child plugins) and
 * renders a safe fallback UI: `<div role="alert">Failed to render document</div>`.
 * This keeps the rest of the page (sidebar, header, tab switcher) alive even if the
 * rehype/remark pipeline throws on pathological input.
 *
 * Architecture ref: §"Mermaid mechanics" — "React error boundary (MarkdownErrorBoundary)
 * wrapping <MarkdownRenderer> so any unexpected throw inside the rendering pipeline shows
 * 'Failed to render document' rather than blanking the whole page."
 */
export class MarkdownErrorBoundary extends React.Component<
  MarkdownErrorBoundaryProps,
  MarkdownErrorBoundaryState
> {
  constructor(props: MarkdownErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(): MarkdownErrorBoundaryState {
    return { hasError: true };
  }

  componentDidCatch(error: Error, info: React.ErrorInfo): void {
    // Log in development; in production a real logger would go here.
    console.error('[MarkdownErrorBoundary] caught render error:', error, info);
  }

  render(): React.ReactNode {
    if (this.state.hasError) {
      return (
        <div role="alert" className="p-4 text-sm text-red-600">
          Failed to render document
        </div>
      );
    }
    return this.props.children;
  }
}
