import { useState, useEffect, useRef, useCallback } from 'react';
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
 * Race-safe hook that fetches the document list for a project.
 *
 * Uses AbortController + a stale-id ref to ensure that rapid projectId changes
 * always end with the most-recently-requested project's documents in state —
 * parity with useDocument (D-005 / US006 AbortController harmonisation).
 *
 * Pattern (mirrors useDocument exactly):
 *  1. On each new projectId (or fetchCount increment), abort the prior controller.
 *  2. Create a new AbortController and store the current projectId in latestIdRef.
 *  3. Issue fetchProjectDocuments(projectId, controller.signal).
 *  4. On resolve: only commit state if projectId === latestIdRef.current.
 *  5. On error: ignore if signal is already aborted (stale request); otherwise set error.
 *
 * `refetch()` increments `fetchCount`, triggering a new effect run (new controller,
 * aborting the previous one if still in-flight) — so retry semantics are preserved.
 *
 * Skips the fetch when `projectId` is undefined (router not ready yet).
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

  const controllerRef = useRef<AbortController | null>(null);
  const latestIdRef = useRef<string | undefined>(undefined);

  const refetch = useCallback(() => {
    setFetchCount((c) => c + 1);
  }, []);

  useEffect(() => {
    if (projectId === undefined) {
      return;
    }

    // Abort any in-flight request from the previous projectId or previous refetch
    controllerRef.current?.abort();

    const controller = new AbortController();
    controllerRef.current = controller;
    latestIdRef.current = projectId;

    setIsLoading(true);
    setError(null);

    fetchProjectDocuments(projectId, controller.signal)
      .then((result) => {
        // Belt-and-braces: only commit if this projectId is still the latest
        if (latestIdRef.current === projectId) {
          setData(result);
          setIsLoading(false);
        }
      })
      .catch((err: unknown) => {
        // Ignore errors from aborted (superseded) requests
        if (controller.signal.aborted) return;

        // Only commit error if this projectId is still the latest
        if (latestIdRef.current === projectId) {
          if (err instanceof ApiError) {
            setError(err);
          } else if (err instanceof Error) {
            setError(err);
          } else {
            setError(new Error('Failed to load documents'));
          }
          setIsLoading(false);
        }
      });

    return () => {
      controller.abort();
    };
  }, [projectId, fetchCount]);

  return { data, isLoading, error, refetch };
};
