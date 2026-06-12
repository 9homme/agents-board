import React from 'react';
import { useRouter } from 'next/router';
import { DocumentSidebar } from './DocumentSidebar';
import { DocumentPreviewer } from './DocumentPreviewer';
import { useProjectDocuments } from '../../hooks/useProjectDocuments';
import { useRequirementDocuments } from '../../hooks/useRequirementDocuments';
import { useDocument } from '../../hooks/useDocument';
import { ApiError } from '../../lib/api/client';
import { DocumentListItem } from '../../lib/api/types';

interface DocumentsTabProps {
  /** The project id whose documents are being browsed. */
  projectId: string;
  /**
   * The requirement id that scopes documents via the §10 canonical path.
   * When provided, fetches from /requirements/:rid/documents.
   * When absent, falls back to the project-scoped endpoint (legacy).
   */
  requirementId?: string;
}

/**
 * DocumentsTab component.
 *
 * Orchestrates DocumentSidebar + DocumentPreviewer for the Documents tab.
 * When `requirementId` is provided, fetches from the canonical §10 path:
 *   GET /api/v1/projects/{projectId}/requirements/{requirementId}/documents
 *
 * State strategy (architecture §11.3):
 * - URL `?doc=` is the source of truth for the selected document.
 * - **Render-time selection:** when `doc` is absent and the list is non-empty,
 *   `selectedDocId` is computed as `documents[0].id` at render time. No URL write
 *   occurs on initial load (architecture §11.3.3 / OQ-6: bare URL is acceptable).
 * - **Bogus deep-link:** when `doc` is present but not in the list, render
 *   `isNotFound=true` in the previewer; do NOT auto-select.
 * - **Content 404:** if `useDocument` returns ApiError with code NOT_FOUND,
 *   treat the same as `isNotFound` (no Retry).
 * - **User-driven URL writes:** `handleSelectDoc` continues to call `router.replace`
 *   on every sidebar click — this is the only path that writes `?doc=` to the URL.
 *
 * The component passes `key={selectedDocId}` to DocumentPreviewer so US003's
 * mermaid mount/unmount cleanup works without further changes here.
 */
export const DocumentsTab: React.FC<DocumentsTabProps> = ({ projectId, requirementId }) => {
  const router = useRouter();
  const docParam = typeof router.query.doc === 'string' ? router.query.doc : undefined;

  // Requirement-scoped hook — uses §10 canonical path when requirementId is set
  const {
    documents: reqDocuments,
    loading: reqLoading,
    error: reqError,
    refresh: reqRefresh,
  } = useRequirementDocuments(
    requirementId !== undefined ? projectId : undefined,
    requirementId
  );

  // Legacy project-scoped fetch when no requirementId
  const {
    data: legacyListData,
    isLoading: legacyListLoading,
    error: legacyListError,
    refetch: legacyRefetch,
  } = useProjectDocuments(requirementId === undefined ? projectId : undefined);

  // Unify the two data sources
  const listLoading = requirementId !== undefined ? reqLoading : legacyListLoading;
  const listError = requirementId !== undefined ? reqError : legacyListError;
  const documents: DocumentListItem[] | null = requirementId !== undefined
    ? reqDocuments
    : (legacyListData?.documents ?? null);
  const refetchList = requirementId !== undefined ? reqRefresh : legacyRefetch;

  // Determine the selected document id:
  // - If docParam is set and is in the list → use it.
  // - If docParam is set but NOT in the list → bogus deep-link (isNotFound).
  // - If docParam is absent → fall back to documents[0].id at render time (architecture §11.3.2).
  const docInList =
    docParam !== undefined && documents !== null
      ? documents.find((d) => d.id === docParam)
      : undefined;

  const isBogusDeepLink =
    docParam !== undefined &&
    documents !== null &&
    documents.length > 0 &&
    docInList === undefined;

  // Render-time selection: when docParam is absent, fall back to the first document.
  const selectedDocId = isBogusDeepLink ? undefined : (docParam ?? documents?.[0]?.id);

  // Fetch the selected document (only when we have a valid, non-bogus id)
  const {
    data: documentData,
    isLoading: docLoading,
    error: docError,
    refetch: refetchDoc,
  } = useDocument(projectId, requirementId, isBogusDeepLink ? undefined : selectedDocId);

  const handleSelectDoc = (id: string) => {
    void router.replace(
      {
        pathname: router.pathname,
        query: { ...router.query, doc: id },
      },
      undefined,
      { shallow: true }
    );
  };

  // Determine isNotFound for the previewer
  const isContentNotFound =
    isBogusDeepLink ||
    (docError instanceof ApiError && docError.code === 'NOT_FOUND');

  // Determine the error to pass to the previewer (exclude 404 — those go to isNotFound)
  const previewerError =
    docError !== null && !isContentNotFound ? docError : null;

  // --- Render ---

  if (listLoading) {
    return (
      <div
        role="tabpanel"
        id="tabpanel-documents"
        aria-labelledby="tab-documents"
        className="flex h-full"
      >
        {/* Sidebar loading state */}
        <div className="w-64 flex-shrink-0 border-r border-gray-200 p-3">
          <div data-testid="documents-list-loading" aria-label="Loading documents">
            <div className="h-4 bg-gray-200 rounded w-32 mb-3 animate-pulse" />
            <div className="space-y-2">
              {[1, 2, 3].map((i) => (
                <div key={i} className="h-8 bg-gray-100 rounded animate-pulse" />
              ))}
            </div>
          </div>
        </div>
        {/* Previewer neutral state while list loads */}
        <div className="flex-1 flex items-center justify-center p-8">
          <p className="text-sm text-gray-400">Loading documents…</p>
        </div>
      </div>
    );
  }

  if (listError) {
    return (
      <div
        role="tabpanel"
        id="tabpanel-documents"
        aria-labelledby="tab-documents"
        className="flex h-full"
      >
        {/* Sidebar error state */}
        <div className="w-64 flex-shrink-0 border-r border-gray-200 p-4">
          <p className="text-sm text-red-600 mb-3">Couldn&apos;t load documents</p>
          <button
            type="button"
            onClick={() => refetchList()}
            className="px-3 py-1.5 text-sm font-medium text-white bg-blue-600 rounded hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            Retry
          </button>
        </div>
        {/* Previewer neutral message when list is unavailable */}
        <div className="flex-1 flex items-center justify-center p-8">
          <p className="text-sm text-gray-400">Document list unavailable</p>
        </div>
      </div>
    );
  }

  if (documents !== null && documents.length === 0) {
    return (
      <div
        role="tabpanel"
        id="tabpanel-documents"
        aria-labelledby="tab-documents"
        className="flex h-full"
      >
        {/* Sidebar empty state */}
        <div
          data-testid="documents-sidebar-area"
          className="w-64 flex-shrink-0 border-r border-gray-200 p-4"
        >
          <p className="text-sm text-gray-500">No documents yet</p>
        </div>
        {/* Previewer empty state */}
        <div className="flex-1 flex items-center justify-center p-8">
          <p className="text-sm text-gray-400">This project has no documents yet</p>
        </div>
      </div>
    );
  }

  // Add requirementId data attribute to sidebar items for FCT-047-026 assertability
  const docsWithRequirementId = documents ?? [];

  return (
    <div
      role="tabpanel"
      id="tabpanel-documents"
      aria-labelledby="tab-documents"
      className="flex h-full"
      data-requirement-id={requirementId}
    >
      <DocumentSidebar
        documents={docsWithRequirementId}
        selectedId={selectedDocId}
        onSelect={handleSelectDoc}
      />
      <DocumentPreviewer
        key={selectedDocId}
        document={documentData}
        isLoading={docLoading}
        error={previewerError}
        isNotFound={isContentNotFound}
        onRetry={refetchDoc}
      />
    </div>
  );
};
