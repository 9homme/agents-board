import { useState, useEffect, useRef, useCallback } from 'react';
import { UserStoryListItem } from '../lib/api/types';
import { ApiError } from '../lib/api/client';
import { fetchProjectUserStories } from '../lib/api/userStories';

/**
 * Return type for the useProjectUserStories hook.
 */
export interface UseProjectUserStoriesResult {
  stories: UserStoryListItem[];
  loading: boolean;
  error: ApiError | Error | null;
  /** Re-issues the list fetch. Wire to a Retry button if needed. */
  refresh: () => void;
}

/**
 * Race-safe hook that fetches the user story list for a project.
 *
 * Uses AbortController + a stale-id ref to ensure rapid projectId changes
 * always end with the most-recently-requested project's stories in state —
 * mirrors the useProjectDocuments pattern (D-005 / US004 AbortController harmonisation).
 *
 * Pattern:
 *  1. On each new projectId (or fetchCount increment), abort the prior controller.
 *  2. Create a new AbortController and store the current projectId in latestIdRef.
 *  3. Issue fetchProjectUserStories(projectId, controller.signal).
 *  4. On resolve: only commit state if projectId === latestIdRef.current.
 *  5. On error: ignore if signal is already aborted (stale request); otherwise set error.
 *
 * `refresh()` increments `fetchCount`, triggering a new effect run.
 *
 * @param projectId - The project id.
 */
export function useProjectUserStories(projectId: string): UseProjectUserStoriesResult {
  const [stories, setStories] = useState<UserStoryListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<ApiError | Error | null>(null);
  const [fetchCount, setFetchCount] = useState(0);

  const controllerRef = useRef<AbortController | null>(null);
  const latestIdRef = useRef<string | undefined>(undefined);

  const refresh = useCallback(() => {
    setFetchCount((c) => c + 1);
  }, []);

  useEffect(() => {
    // Abort any in-flight request from the previous projectId or previous refresh
    controllerRef.current?.abort();

    const controller = new AbortController();
    controllerRef.current = controller;
    latestIdRef.current = projectId;

    setLoading(true);
    setError(null);

    fetchProjectUserStories(projectId, controller.signal)
      .then((result) => {
        // Only commit if this projectId is still the latest
        if (latestIdRef.current === projectId) {
          setStories(result.userStories);
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
            setError(new Error('Failed to load user stories'));
          }
          setLoading(false);
        }
      });

    return () => {
      controller.abort();
    };
  }, [projectId, fetchCount]);

  return { stories, loading, error, refresh };
}
