import { useState, useEffect, useRef, useCallback } from 'react';
import { fetchDocument } from '../lib/api/documents';
import { ApiError } from '../lib/api/client';
import { Document } from '../lib/api/types';

/**
 * Return type for the useDocument hook.
 */
export interface UseDocumentResult {
  data: Document | null;
  isLoading: boolean;
  error: ApiError | Error | null;
  /** Re-issues the most recent fetch. Wire to the Retry button in the previewer error state. */
  refetch: () => void;
}

/**
 * Race-safe hook that fetches a single document by id.
 *
 * Uses AbortController + a stale-id ref to ensure that rapid consecutive calls
 * (rapid sidebar clicks) always end with the most-recently-requested document
 * in state — never a stale response from an earlier-started but later-resolved
 * request (D-005 in architecture.md).
 *
 * Pattern:
 *  1. On each new documentId, abort the prior controller (cancels in-flight network request).
 *  2. Create a new AbortController and store the current id in latestIdRef.
 *  3. Issue fetchDocument(id, controller.signal).
 *  4. On resolve: only commit state if documentId === latestIdRef.current (belt-and-braces
 *     guard — abort handles the network layer; stale-id handles any late resolution).
 *  5. On error: ignore if signal is already aborted (stale request); otherwise set error.
 *
 * Skips the fetch when `documentId` is undefined.
 *
 * @param documentId - The document id to fetch, or undefined to skip.
 */
export const useDocument = (documentId: string | undefined): UseDocumentResult => {
  const [data, setData] = useState<Document | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<ApiError | Error | null>(null);

  const controllerRef = useRef<AbortController | null>(null);
  const latestIdRef = useRef<string | undefined>(undefined);
  const [fetchCount, setFetchCount] = useState(0);

  const refetch = useCallback(() => {
    setFetchCount((c) => c + 1);
  }, []);

  useEffect(() => {
    if (documentId === undefined) {
      return;
    }

    // Abort any in-flight request from the previous documentId
    controllerRef.current?.abort();

    const controller = new AbortController();
    controllerRef.current = controller;
    latestIdRef.current = documentId;

    setIsLoading(true);
    setError(null);
    setData(null);

    fetchDocument(documentId, controller.signal)
      .then((doc) => {
        // Belt-and-braces: only commit if this id is still the latest
        if (latestIdRef.current === documentId) {
          setData(doc);
          setIsLoading(false);
        }
      })
      .catch((err: unknown) => {
        // Ignore errors from aborted (superseded) requests
        if (controller.signal.aborted) return;

        // Only commit error if this id is still the latest
        if (latestIdRef.current === documentId) {
          if (err instanceof ApiError) {
            setError(err);
          } else if (err instanceof Error) {
            setError(err);
          } else {
            setError(new Error('Failed to load document'));
          }
          setIsLoading(false);
        }
      });

    return () => {
      controller.abort();
    };
  }, [documentId, fetchCount]);

  return { data, isLoading, error, refetch };
};
