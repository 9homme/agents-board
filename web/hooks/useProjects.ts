import { useState, useEffect, useCallback } from 'react';
import { fetchProjects } from '../lib/api/projects';
import { Project } from '../lib/api/types';

/**
 * Hook to manage loading, data, and error state for fetching projects.
 * Exposes a `refetch` function to re-trigger the data load (e.g. after project creation).
 * @returns An object containing the projects data, isLoading, isError, error, and refetch.
 */
export const useProjects = () => {
  const [data, setData] = useState<Project[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isError, setIsError] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);
  const [fetchCount, setFetchCount] = useState(0);

  useEffect(() => {
    let mounted = true;

    const loadProjects = async () => {
      setIsLoading(true);
      setIsError(false);
      setError(null);

      try {
        const response = await fetchProjects();
        if (mounted) {
          setData(response.projects);
        }
      } catch (err: unknown) {
        if (mounted) {
          setIsError(true);
          setError(err instanceof Error ? err : new Error('Failed to load projects'));
        }
      } finally {
        if (mounted) {
          setIsLoading(false);
        }
      }
    };

    loadProjects();

    return () => {
      mounted = false;
    };
  }, [fetchCount]);

  const refetch = useCallback(() => {
    setFetchCount((c) => c + 1);
  }, []);

  return { data, isLoading, isError, error, refetch };
};
