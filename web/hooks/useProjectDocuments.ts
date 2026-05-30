import { useState, useEffect, useCallback } from 'react';
import { fetchProjectDocuments } from '../lib/api/documents';
import { ApiError } from '../lib/api/client';
import { DocumentsListResponse } from '../lib/api/types';

/**
 * Return type for the useProjectDocuments hook.
 */
export interface UseProjectDocumentsResult {
  data: DocumentsListResponse | null;
  isLoading: boolean;
  error: ApiError | Error | null;
  /** Re-issues the list fetch. Wire to the Retry button in the sidebar error state. */
  refetch: () => void;
}

/**
 * Hook that fetches the document list for a project.
 *
 * Skips the fetch when `projectId` is undefined (router not ready yet).
 * Exposes `refetch()` so callers can wire a Retry button.
 *
 * @param projectId - The project id, or undefined to skip.
 */
export const useProjectDocuments = (
  projectId: string | undefined
): UseProjectDocumentsResult => {
  const [data, setData] = useState<DocumentsListResponse | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<ApiError | Error | null>(null);
  const [fetchCount, setFetchCount] = useState(0);

  const refetch = useCallback(() => {
    setFetchCount((c) => c + 1);
  }, []);

  useEffect(() => {
    if (projectId === undefined) {
      return;
    }

    let mounted = true;

    const loadDocuments = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const result = await fetchProjectDocuments(projectId);
        if (mounted) {
          setData(result);
        }
      } catch (err: unknown) {
        if (mounted) {
          if (err instanceof ApiError) {
            setError(err);
          } else if (err instanceof Error) {
            setError(err);
          } else {
            setError(new Error('Failed to load documents'));
          }
        }
      } finally {
        if (mounted) {
          setIsLoading(false);
        }
      }
    };

    loadDocuments();

    return () => {
      mounted = false;
    };
  }, [projectId, fetchCount]);

  return { data, isLoading, error, refetch };
};
