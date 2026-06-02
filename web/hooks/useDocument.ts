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
 * Race-safe hook that fetches a single document by id.
 *
 * Uses AbortController + a stale-id ref to ensure that rapid consecutive calls
 * (rapid sidebar clicks) always end with the most-recently-requested document
 * in state — never a stale response from an earlier-started but later-resolved
 * request (D-005 in architecture.md).
 *
 * Internal state is managed via useReducer (architecture §11.2) — collapses the
 * prior useState×3 cascade into single-dispatch state transitions, clearing the
 * no-cascading-set-state, no-adjust-state-on-prop-change, and
 * rendering-usetransition-loading react-doctor rules.
 *
 * Pattern:
 *  1. On each new documentId, abort the prior controller (cancels in-flight network request).
 *  2. Create a new AbortController and store the current id in latestIdRef.
 *  3. Dispatch FETCH_STARTED (single state transition replaces three setState calls).
 *  4. Issue fetchDocument(id, controller.signal).
 *  5. On resolve: only commit state if documentId === latestIdRef.current; dispatch FETCH_SUCCEEDED.
 *  6. On error: dispatch ABORTED (no-op) if signal aborted; otherwise dispatch FETCH_FAILED.
 *
 * Skips the fetch when `documentId` is undefined.
 *
 * @param documentId - The document id to fetch, or undefined to skip.
 */
export const useDocument = (documentId: string | undefined): UseDocumentResult => {
  const [state, dispatch] = useReducer(reducer, initialState);

  const controllerRef = useRef<AbortController | null>(null);
  const latestIdRef = useRef<string | undefined>(undefined);
  // fetchCount drives manual refetch; stays as separate useState per §11.2.1
  const [fetchCount, setFetchCount] = useState(0);

  const refetch = useCallback(() => {
    setFetchCount((c) => c + 1);
  }, []);

  useEffect(() => {
    if (documentId === undefined) {
      return;
    }

    // Abort any in-flight request from the previous documentId
    controllerRef.current?.abort();

    const controller = new AbortController();
    controllerRef.current = controller;
    latestIdRef.current = documentId;

    // Single dispatch replaces the prior setIsLoading(true); setError(null); setData(null) cascade
    dispatch({ type: 'FETCH_STARTED' });

    fetchDocument(documentId, controller.signal)
      .then((doc) => {
        // Belt-and-braces: only commit if this id is still the latest
        if (latestIdRef.current === documentId) {
          dispatch({ type: 'FETCH_SUCCEEDED', document: doc });
        }
      })
      .catch((err: unknown) => {
        // Ignore errors from aborted (superseded) requests
        if (controller.signal.aborted) {
          dispatch({ type: 'ABORTED' });
          return;
        }

        // Only commit error if this id is still the latest (stale-id guard)
        if (latestIdRef.current !== documentId) return;

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
  }, [documentId, fetchCount]);

  return { data: state.data, isLoading: state.isLoading, error: state.error, refetch };
};
