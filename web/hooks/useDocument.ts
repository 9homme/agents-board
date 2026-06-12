import { useReducer, useEffect, useRef, useCallback, useState } from 'react';
import { fetchDocument } from '../lib/api/documents';
import { ApiError } from '../lib/api/client';
import { Document } from '../lib/api/types';

/**
 * Return type for the useDocument hook.
 * Shape is frozen per architecture §11.2.4 — DO NOT change.
 */
export interface UseDocumentResult {
  data: Document | null;
  isLoading: boolean;
  error: ApiError | Error | null;
  /** Re-issues the most recent fetch. Wire to the Retry button in the previewer error state. */
  refetch: () => void;
}

// ---------------------------------------------------------------------------
// Reducer state and actions (architecture §11.2.1 – §11.2.3)
// ---------------------------------------------------------------------------

type State = {
  data: Document | null;
  isLoading: boolean;
  error: ApiError | Error | null;
};

type Action =
  | { type: 'FETCH_STARTED' }
  | { type: 'FETCH_SUCCEEDED'; document: Document }
  | { type: 'FETCH_FAILED'; error: ApiError | Error }
  | { type: 'ABORTED' };

const initialState: State = { data: null, isLoading: false, error: null };

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'FETCH_STARTED':
      return { data: null, isLoading: true, error: null };
    case 'FETCH_SUCCEEDED':
      return { data: action.document, isLoading: false, error: null };
    case 'FETCH_FAILED':
      return { data: null, isLoading: false, error: action.error };
    case 'ABORTED':
      return state; // no-op; aborts are control flow, not failures
    default:
      return state;
  }
}

/**
 * Race-safe hook that fetches a single document by id via the hierarchical endpoint.
 *
 * Skips the fetch when any of projectId, requirementId, or documentId is undefined.
 */
export const useDocument = (
  projectId: string | undefined,
  requirementId: string | undefined,
  documentId: string | undefined
): UseDocumentResult => {
  const [state, dispatch] = useReducer(reducer, initialState);

  const controllerRef = useRef<AbortController | null>(null);
  const latestKeyRef = useRef<string | undefined>(undefined);
  const [fetchCount, setFetchCount] = useState(0);

  const refetch = useCallback(() => {
    setFetchCount((c) => c + 1);
  }, []);

  useEffect(() => {
    if (!projectId || !requirementId || !documentId) {
      return;
    }

    const compositeKey = `${projectId}|${requirementId}|${documentId}`;

    controllerRef.current?.abort();

    const controller = new AbortController();
    controllerRef.current = controller;
    latestKeyRef.current = compositeKey;

    dispatch({ type: 'FETCH_STARTED' });

    fetchDocument(projectId, requirementId, documentId, controller.signal)
      .then((doc) => {
        if (latestKeyRef.current === compositeKey) {
          dispatch({ type: 'FETCH_SUCCEEDED', document: doc });
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
              : new Error('Failed to load document');
        dispatch({ type: 'FETCH_FAILED', error });
      });

    return () => {
      controller.abort();
    };
  }, [projectId, requirementId, documentId, fetchCount]);

  return { data: state.data, isLoading: state.isLoading, error: state.error, refetch };
};
