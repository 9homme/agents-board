import { fetchClient } from './client';
import { Document, DocumentsListResponse } from './types';

/**
 * Fetch the list of document metadata for a project.
 * Corresponds to GET /api/v1/projects/{projectId}/documents.
 *
 * The response contains only metadata (no `content` field per D-002).
 * Order is updatedAt DESC, id DESC (enforced by the backend).
 *
 * @param projectId - The project id (will be URL-encoded).
 * @param signal    - Optional AbortSignal for request cancellation (D-005).
 */
export const fetchProjectDocuments = async (
  projectId: string,
  signal?: AbortSignal
): Promise<DocumentsListResponse> => {
  return fetchClient<DocumentsListResponse>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/documents`,
    { signal }
  );
};

/**
 * Fetch a single document including its raw markdown content.
 * Corresponds to GET /api/v1/documents/{documentId}.
 *
 * @param documentId - The document id (will be URL-encoded).
 * @param signal     - Optional AbortSignal for request cancellation (D-005).
 */
export const fetchDocument = async (
  documentId: string,
  signal?: AbortSignal
): Promise<Document> => {
  return fetchClient<Document>(
    `/api/v1/documents/${encodeURIComponent(documentId)}`,
    { signal }
  );
};
