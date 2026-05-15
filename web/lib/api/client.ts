import { ErrorResponse } from './types';

export class ApiError extends Error {
  public code: string;

  constructor(message: string, code: string) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
  }
}

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
