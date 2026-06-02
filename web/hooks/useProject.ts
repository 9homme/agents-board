import { useState, useEffect, useRef } from 'react';
import { fetchProject } from '../lib/api/projects';
import { ApiError } from '../lib/api/client';
import { Project } from '../lib/api/types';

/**
 * Return type for the useProject hook.
 */
export interface UseProjectResult {
  data: Project | null;
  isLoading: boolean;
  error: ApiError | Error | null;
  isNotFound: boolean;
}

/**
 * Race-safe hook that fetches a single project by id.
 *
 * Uses AbortController + a stale-id ref to ensure that rapid id changes always
 * end with the most-recently-requested project in state — never a stale response
 * from an earlier-started but later-resolved request (parity with useDocument,
 * D-005 in architecture.md / US006 AbortController harmonisation).
 *
 * Pattern (mirrors useDocument exactly):
 *  1. On each new id, abort the prior controller (cancels in-flight network request).
 *  2. Create a new AbortController and store the current id in latestIdRef.
 *  3. Issue fetchProject(id, controller.signal).
 *  4. On resolve: only commit state if id === latestIdRef.current (belt-and-braces guard).
 *  5. On error: ignore if signal is already aborted (stale request); otherwise set error.
 *
 * Skips the fetch when `id` is undefined (Next.js Pages Router first render
 * before `router.isReady` — id comes from useRouter().query).
 *
 * Discriminates ApiError.code === 'NOT_FOUND' to set `isNotFound: true`
 * so the page can render a friendly 404 message rather than the generic
 * error state.
 *
 * @param id - The project id, or undefined when router is not yet ready.
 */
export const useProject = (id: string | undefined): UseProjectResult => {
  const [data, setData] = useState<Project | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<ApiError | Error | null>(null);
  const [isNotFound, setIsNotFound] = useState<boolean>(false);

  const controllerRef = useRef<AbortController | null>(null);
  const latestIdRef = useRef<string | undefined>(undefined);

  useEffect(() => {
    if (id === undefined) {
      // Router not ready yet — skip fetch
      return;
    }

    // Abort any in-flight request from the previous id
    controllerRef.current?.abort();

    const controller = new AbortController();
    controllerRef.current = controller;
    latestIdRef.current = id;

    setIsLoading(true);
    setError(null);
    setIsNotFound(false);
    setData(null);

    fetchProject(id, controller.signal)
      .then((project) => {
        // Belt-and-braces: only commit if this id is still the latest
        if (latestIdRef.current === id) {
          setData(project);
          setIsLoading(false);
        }
      })
      .catch((err: unknown) => {
        // Ignore errors from aborted (superseded) requests
        if (controller.signal.aborted) return;

        // Only commit error if this id is still the latest
        if (latestIdRef.current === id) {
          if (err instanceof ApiError) {
            setError(err);
            if (err.code === 'NOT_FOUND') {
              setIsNotFound(true);
            }
          } else if (err instanceof Error) {
            setError(err);
          } else {
            setError(new Error('Failed to load project'));
          }
          setIsLoading(false);
        }
      });

    return () => {
      controller.abort();
    };
  }, [id]);

  return { data, isLoading, error, isNotFound };
};
