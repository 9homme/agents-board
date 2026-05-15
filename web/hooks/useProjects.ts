import { useState, useEffect } from 'react';
import { fetchProjects } from '../lib/api/projects';
import { Project } from '../lib/api/types';

/**
 * Hook to manage loading, data, and error state for fetching projects.
 * @returns An object containing the projects data, isLoading, isError, and error state.
 */
export const useProjects = () => {
  const [data, setData] = useState<Project[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isError, setIsError] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);

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
      } catch (err: any) {
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
  }, []);

  return { data, isLoading, isError, error };
};
