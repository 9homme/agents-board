import { ErrorResponse } from './types';

export class ApiError extends Error {
  public code: string;

  constructor(message: string, code: string) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
  }
}

/**
 * Base HTTP client for all backend API calls.
 *
 * Accepts an optional `signal: AbortSignal` (via `options`) and forwards it
 * to `fetch` so callers can cancel in-flight requests (e.g. useDocument's
 * AbortController pattern — see D-005 in architecture.md).
 *
 * Existing callers that omit `signal` continue to work unchanged.
 */
export const fetchClient = async <T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> => {
  const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || '';
  const url = `${baseUrl}${endpoint}`;

  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    // signal is already included via ...options spread above; explicitly
    // listing it here makes the pass-through intent clear for readers.
    signal: options.signal,
  });

  if (!response.ok) {
    let errorRes: ErrorResponse;
    try {
      errorRes = await response.json();
    } catch {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    throw new ApiError(errorRes.message || 'An error occurred', errorRes.code || 'UNKNOWN_ERROR');
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return {} as T;
  }

  return response.json();
};
