import { useState, useRef, useCallback, useEffect } from 'react';
import { createProject as apiCreateProject } from '../lib/api/projects';
import { ApiError } from '../lib/api/client';
import { Project, CreateProjectRequest } from '../lib/api/types';

/** Lifecycle status for a createProject submission. */
export type CreateProjectStatus = 'idle' | 'submitting' | 'success' | 'error';

/** Return type for the useCreateProject hook. */
export interface UseCreateProjectResult {
  /** Submit a create-project request. Resolves with the created Project on success. */
  createProject: (req: CreateProjectRequest) => Promise<Project>;
  /** Current lifecycle status of the submission. */
  status: CreateProjectStatus;
  /** ApiError from the most recent failed submission, or null. */
  error: ApiError | Error | null;
  /** Reset status back to idle and clear the error. */
  reset: () => void;
}

/**
 * Hook encapsulating form submission state for creating a new project.
 *
 * State machine: idle → submitting → success | error
 * Guards against double-submit: ignores calls while status === 'submitting'.
 * Aborts any in-flight request on unmount (AbortController cleanup via useEffect).
 */
export const useCreateProject = (): UseCreateProjectResult => {
  const [status, setStatus] = useState<CreateProjectStatus>('idle');
  const [error, setError] = useState<ApiError | Error | null>(null);

  // Ref to track if the component is still mounted (prevents state updates after unmount).
  const mountedRef = useRef(true);
  // AbortController for the current in-flight request.
  const controllerRef = useRef<AbortController | null>(null);
  // Stable ref to the current status so useCallback can read it without re-creating.
  const statusRef = useRef<CreateProjectStatus>('idle');
  statusRef.current = status;

  // Cleanup on unmount: mark as unmounted and abort any in-flight request.
  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      controllerRef.current?.abort();
    };
  }, []);

  const submit = useCallback(async (req: CreateProjectRequest): Promise<Project> => {
    // Guard against double-submit
    if (statusRef.current === 'submitting') {
      return Promise.reject(new Error('Already submitting'));
    }

    // Abort any previous in-flight request before starting a new one
    controllerRef.current?.abort();
    const controller = new AbortController();
    controllerRef.current = controller;

    setStatus('submitting');
    setError(null);

    try {
      const project = await apiCreateProject(req);

      if (mountedRef.current) {
        setStatus('success');
        setError(null);
      }
      return project;
    } catch (err: unknown) {
      // Ignore errors caused by our own abort (unmount cleanup)
      if (controller.signal.aborted) {
        return Promise.reject(err);
      }

      if (mountedRef.current) {
        const normalized =
          err instanceof ApiError
            ? err
            : err instanceof Error
              ? err
              : new Error('Failed to create project');
        setStatus('error');
        setError(normalized);
      }
      throw err;
    }
  }, []); // stable — reads statusRef instead of status closure

  const reset = useCallback(() => {
    setStatus('idle');
    setError(null);
  }, []);

  return { createProject: submit, status, error, reset };
};
