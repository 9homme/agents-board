import React from 'react';
import { Document } from '../../lib/api/types';
import { MarkdownRenderer } from './MarkdownRenderer';
import { MarkdownErrorBoundary } from './MarkdownErrorBoundary';

interface DocumentPreviewerProps {
  /** The loaded document, or null when loading / error / not-found. */
  document: Document | null;
  /** True while the content fetch is in flight. */
  isLoading: boolean;
  /** Non-null when the content fetch failed (non-404). */
  error: Error | null;
  /**
   * True when the document id is not in the project's document list
   * (deep-link-to-bogus-doc case), or when the detail endpoint returned 404.
   * No Retry is offered in this state.
   */
  isNotFound: boolean;
  /** Callback invoked when the user clicks the Retry button (error state only). */
  onRetry: () => void;
}

/**
 * DocumentPreviewer component — US002 plain rendering variant.
 *
 * Renders one of four mutually-exclusive states:
 * 1. **Loading** — spinner/skeleton.
 * 2. **Error (non-404)** — "Failed to load document" + Retry button.
 * 3. **Not found** — "Document not found" friendly message; no Retry.
 * 4. **Loaded** — `<h2>` title, muted updatedAt, markdown content rendered
 *    via `<MarkdownErrorBoundary><MarkdownRenderer source={content} /></MarkdownErrorBoundary>`.
 *
 * The parent (`DocumentsTab`) passes `key={document.id}` so mermaid SVG
 * state (US003) is cleaned up correctly on document switch.
 */
export const DocumentPreviewer: React.FC<DocumentPreviewerProps> = ({
  document,
  isLoading,
  error,
  isNotFound,
  onRetry,
}) => {
  if (isLoading) {
    return (
      <div
        className="flex-1 flex items-center justify-center p-8"
        role="status"
        aria-label="Loading document"
      >
        <div className="flex flex-col items-center gap-3">
          <div
            data-testid="document-loading-spinner"
            className="h-8 w-8 rounded-full border-4 border-gray-200 border-t-blue-500 animate-spin"
          />
          <p className="text-sm text-gray-500">Loading document…</p>
        </div>
      </div>
    );
  }

  if (isNotFound) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="text-center">
          <h2 className="text-lg font-semibold text-gray-700 mb-2">Document not found</h2>
          <p className="text-sm text-gray-500">
            The document you requested does not exist in this project.
          </p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="text-center">
          <p className="text-sm text-red-600 mb-4">Failed to load document</p>
          <button
            onClick={onRetry}
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  if (!document) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <p className="text-sm text-gray-400">Select a document to view it.</p>
      </div>
    );
  }

  return (
    <article
      className="flex-1 overflow-auto p-8"
      aria-label={document.title}
    >
      <header className="mb-6">
        <h2 className="text-2xl font-bold text-gray-900">{document.title}</h2>
        <p className="text-sm text-gray-500 mt-1">Updated {document.updatedAt}</p>
      </header>
      <MarkdownErrorBoundary>
        <MarkdownRenderer source={document.content} />
      </MarkdownErrorBoundary>
    </article>
  );
};
