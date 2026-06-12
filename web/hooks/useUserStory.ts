import { useReducer, useEffect, useRef, useCallback, useState } from 'react';
import { fetchUserStory } from '../lib/api/userStories';
import { ApiError } from '../lib/api/client';
import { UserStory } from '../lib/api/types';

/**
 * Return type for the useUserStory hook.
 */
export interface UseUserStoryResult {
  data: UserStory | null;
  isLoading: boolean;
  error: ApiError | Error | null;
  /** Re-issues the most recent fetch. Wire to the Retry button if needed. */
  refetch: () => void;
}

// ---------------------------------------------------------------------------
// Reducer state and actions
// ---------------------------------------------------------------------------

type State = {
  data: UserStory | null;
  isLoading: boolean;
  error: ApiError | Error | null;
};

type Action =
  | { type: 'FETCH_STARTED' }
  | { type: 'FETCH_SUCCEEDED'; story: UserStory }
  | { type: 'FETCH_FAILED'; error: ApiError | Error }
  | { type: 'ABORTED' };

const initialState: State = { data: null, isLoading: false, error: null };

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'FETCH_STARTED':
      return { data: null, isLoading: true, error: null };
    case 'FETCH_SUCCEEDED':
      return { data: action.story, isLoading: false, error: null };
    case 'FETCH_FAILED':
      return { data: null, isLoading: false, error: action.error };
    case 'ABORTED':
      return state; // no-op; aborts are control flow, not failures
    default:
      return state;
  }
}

/**
 * Race-safe hook that fetches a single user story by id.
 *
 * Skips the fetch when any of projectId, requirementId, or storyId is undefined.
 */
export const useUserStory = (
  projectId: string | undefined,
  requirementId: string | undefined,
  storyId: string | undefined
): UseUserStoryResult => {
  const [state, dispatch] = useReducer(reducer, initialState);

  const controllerRef = useRef<AbortController | null>(null);
  const latestKeyRef = useRef<string | undefined>(undefined);
  const [fetchCount, setFetchCount] = useState(0);

  const refetch = useCallback(() => {
    setFetchCount((c) => c + 1);
  }, []);

  useEffect(() => {
    if (!projectId || !requirementId || !storyId) {
      return;
    }

    const compositeKey = `${projectId}|${requirementId}|${storyId}`;

    controllerRef.current?.abort();

    const controller = new AbortController();
    controllerRef.current = controller;
    latestKeyRef.current = compositeKey;

    dispatch({ type: 'FETCH_STARTED' });

    fetchUserStory(projectId, requirementId, storyId, controller.signal)
      .then((story) => {
        if (latestKeyRef.current === compositeKey) {
          dispatch({ type: 'FETCH_SUCCEEDED', story });
        }
      })
      .catch((err: unknown) => {
        if (controller.signal.aborted) {
          dispatch({ type: 'ABORTED' });
          return;
        }

        if (latestKeyRef.current !== compositeKey) return;

        const error: ApiError | Error =
          err instanceof ApiError
            ? err
            : err instanceof Error
              ? err
              : new Error('Failed to load user story');
        dispatch({ type: 'FETCH_FAILED', error });
      });

    return () => {
      controller.abort();
    };
  }, [projectId, requirementId, storyId, fetchCount]);

  return { data: state.data, isLoading: state.isLoading, error: state.error, refetch };
};
