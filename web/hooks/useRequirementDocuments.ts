import { useState, useEffect, useRef, useCallback } from 'react';
import { DocumentListItem } from '../lib/api/types';
import { ApiError } from '../lib/api/client';
import { fetchRequirementDocuments } from '../lib/api/documents';

/**
 * Return type for the useRequirementDocuments hook.
 */
export interface UseRequirementDocumentsResult {
  documents: DocumentListItem[];
  loading: boolean;
  error: ApiError | Error | null;
  /** Re-issues the list fetch. Wire to a Retry button if needed. */
  refresh: () => void;
}

/**
 * Race-safe hook that fetches documents scoped to a requirement.
 *
 * Uses AbortController + a stale-key ref to ensure rapid requirement changes
 * always end with the most-recently-requested requirement's documents in state —
 * mirrors the useProjectDocuments pattern (D-005).
 *
 * Calls GET /api/v1/projects/{projectId}/requirements/{requirementId}/documents (§10).
 *
 * Skips the fetch when either `projectId` or `requirementId` is undefined.
 *
 * @param projectId     - The project id.
 * @param requirementId - The requirement id.
 */
export function useRequirementDocuments(
  projectId: string | undefined,
  requirementId: string | undefined
): UseRequirementDocumentsResult {
  const [documents, setDocuments] = useState<DocumentListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<ApiError | Error | null>(null);
  const [fetchCount, setFetchCount] = useState(0);

  const controllerRef = useRef<AbortController | null>(null);
  const latestKeyRef = useRef<string | undefined>(undefined);

  const refresh = useCallback(() => {
    setFetchCount((c) => c + 1);
  }, []);

  useEffect(() => {
    if (projectId === undefined || requirementId === undefined) {
      setLoading(false);
      return;
    }

    const key = `${projectId}:${requirementId}`;

    // Abort any in-flight request from the previous key
    controllerRef.current?.abort();

    const controller = new AbortController();
    controllerRef.current = controller;
    latestKeyRef.current = key;

    setLoading(true);
    setError(null);

    fetchRequirementDocuments(projectId, requirementId, controller.signal)
      .then((result) => {
        if (latestKeyRef.current === key) {
          setDocuments(result.documents);
          setLoading(false);
        }
      })
      .catch((err: unknown) => {
        if (controller.signal.aborted) return;

        if (latestKeyRef.current === key) {
          if (err instanceof ApiError) {
            setError(err);
          } else if (err instanceof Error) {
            setError(err);
          } else {
            setError(new Error('Failed to load documents'));
          }
          setLoading(false);
        }
      });

    return () => {
      controller.abort();
    };
  }, [projectId, requirementId, fetchCount]);

  return { documents, loading, error, refresh };
}
