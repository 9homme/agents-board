import { useReducer, useEffect, useRef, useCallback, useState } from 'react';
import { fetchUserStoryTasks } from '../lib/api/userStories';
import { ApiError } from '../lib/api/client';
import { TasksListResponse } from '../lib/api/types';

/**
 * Return type for the useUserStoryTasks hook.
 */
export interface UseUserStoryTasksResult {
  data: TasksListResponse | null;
  isLoading: boolean;
  error: ApiError | Error | null;
  /** Re-issues the most recent fetch. Wire to the Retry button if needed. */
  refetch: () => void;
}

// ---------------------------------------------------------------------------
// Reducer state and actions
// ---------------------------------------------------------------------------

type State = {
  data: TasksListResponse | null;
  isLoading: boolean;
  error: ApiError | Error | null;
};

type Action =
  | { type: 'FETCH_STARTED' }
  | { type: 'FETCH_SUCCEEDED'; tasks: TasksListResponse }
  | { type: 'FETCH_FAILED'; error: ApiError | Error }
  | { type: 'ABORTED' };

const initialState: State = { data: null, isLoading: false, error: null };

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'FETCH_STARTED':
      return { data: null, isLoading: true, error: null };
    case 'FETCH_SUCCEEDED':
      return { data: action.tasks, isLoading: false, error: null };
    case 'FETCH_FAILED':
      return { data: null, isLoading: false, error: action.error };
    case 'ABORTED':
      return state; // no-op; aborts are control flow, not failures
    default:
      return state;
  }
}

/**
 * Race-safe hook that fetches all tasks for a user story by id.
 *
 * Uses AbortController + a stale-id ref to ensure that rapid id changes
 * always end with the most-recently-requested story's tasks in state.
 * Mirrors the pattern established in useDocument.ts (D-005).
 *
 * Skips the fetch when `storyId` is undefined.
 *
 * @param storyId - The user story id whose tasks to fetch, or undefined to skip.
 */
export const useUserStoryTasks = (
  storyId: string | undefined
): UseUserStoryTasksResult => {
  const [state, dispatch] = useReducer(reducer, initialState);

  const controllerRef = useRef<AbortController | null>(null);
  const latestIdRef = useRef<string | undefined>(undefined);
  const [fetchCount, setFetchCount] = useState(0);

  const refetch = useCallback(() => {
    setFetchCount((c) => c + 1);
  }, []);

  useEffect(() => {
    if (storyId === undefined) {
      return;
    }

    // Abort any in-flight request from the previous storyId
    controllerRef.current?.abort();

    const controller = new AbortController();
    controllerRef.current = controller;
    latestIdRef.current = storyId;

    dispatch({ type: 'FETCH_STARTED' });

    fetchUserStoryTasks(storyId, controller.signal)
      .then((tasks) => {
        if (latestIdRef.current === storyId) {
          dispatch({ type: 'FETCH_SUCCEEDED', tasks });
        }
      })
      .catch((err: unknown) => {
        if (controller.signal.aborted) {
          dispatch({ type: 'ABORTED' });
          return;
        }

        if (latestIdRef.current !== storyId) return;

        const error: ApiError | Error =
          err instanceof ApiError
            ? err
            : err instanceof Error
              ? err
              : new Error('Failed to load tasks');
        dispatch({ type: 'FETCH_FAILED', error });
      });

    return () => {
      controller.abort();
    };
  }, [storyId, fetchCount]);

  return { data: state.data, isLoading: state.isLoading, error: state.error, refetch };
};
