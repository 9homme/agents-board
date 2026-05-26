import { useState, useEffect } from 'react';
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
 * Hook that fetches a single project by id.
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

  useEffect(() => {
    if (id === undefined) {
      // Router not ready yet — skip fetch
      return;
    }

    let mounted = true;

    const loadProject = async () => {
      setIsLoading(true);
      setError(null);
      setIsNotFound(false);
      setData(null);

      try {
        const project = await fetchProject(id);
        if (mounted) {
          setData(project);
        }
      } catch (err: unknown) {
        if (mounted) {
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
        }
      } finally {
        if (mounted) {
          setIsLoading(false);
        }
      }
    };

    loadProject();

    return () => {
      mounted = false;
    };
  }, [id]);

  return { data, isLoading, error, isNotFound };
};
