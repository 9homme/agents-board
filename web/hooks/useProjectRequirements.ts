import { useState, useEffect, useRef, useCallback } from 'react';
import { Requirement } from '../lib/api/types';
import { ApiError } from '../lib/api/client';
import { fetchProjectRequirements } from '../lib/api/requirements';

/**
 * Return type for the useProjectRequirements hook.
 */
export interface UseProjectRequirementsResult {
  requirements: Requirement[];
  loading: boolean;
  error: ApiError | Error | null;
  /** Re-issues the list fetch. Wire to a Retry button if needed. */
  refresh: () => void;
}

/**
 * Race-safe hook that fetches the requirements list for a project.
 *
 * Uses AbortController + a stale-id ref to ensure rapid projectId changes
 * always end with the most-recently-requested project's requirements in state —
 * mirrors the useProjectUserStories / useProjectDocuments pattern (D-005).
 *
 * Pattern:
 *  1. On each new projectId (or fetchCount increment), abort the prior controller.
 *  2. Create a new AbortController and store the current projectId in latestIdRef.
 *  3. Issue fetchProjectRequirements(projectId, controller.signal).
 *  4. On resolve: only commit state if projectId === latestIdRef.current.
 *  5. On error: ignore if signal is already aborted (stale request); otherwise set error.
 *
 * `refresh()` increments `fetchCount`, triggering a new effect run.
 *
 * Skips the fetch when `projectId` is undefined (router not ready yet).
 *
 * @param projectId - The project id, or undefined to skip.
 */
export function useProjectRequirements(
  projectId: string | undefined
): UseProjectRequirementsResult {
  const [requirements, setRequirements] = useState<Requirement[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<ApiError | Error | null>(null);
  const [fetchCount, setFetchCount] = useState(0);

  const controllerRef = useRef<AbortController | null>(null);
  const latestIdRef = useRef<string | undefined>(undefined);

  const refresh = useCallback(() => {
    setFetchCount((c) => c + 1);
  }, []);

  useEffect(() => {
    if (projectId === undefined) {
      setLoading(false);
      return;
    }

    // Abort any in-flight request from the previous projectId or previous refresh
    controllerRef.current?.abort();

    const controller = new AbortController();
    controllerRef.current = controller;
    latestIdRef.current = projectId;

    setLoading(true);
    setError(null);

    fetchProjectRequirements(projectId, controller.signal)
      .then((result) => {
        // Only commit if this projectId is still the latest
        if (latestIdRef.current === projectId) {
          setRequirements(result.requirements);
          setLoading(false);
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
            setError(new Error('Failed to load requirements'));
          }
          setLoading(false);
        }
      });

    return () => {
      controller.abort();
    };
  }, [projectId, fetchCount]);

  return { requirements, loading, error, refresh };
}
