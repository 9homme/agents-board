import { useState } from 'react';
import { UserStoryListItem } from '../lib/api/types';
import { ApiError } from '../lib/api/client';

/**
 * Return type for the useProjectUserStories hook.
 */
export interface UseProjectUserStoriesResult {
  stories: UserStoryListItem[];
  loading: boolean;
  error: ApiError | Error | null;
  refresh: () => void;
}

/**
 * Stub — will be replaced with full AbortController implementation.
 */
export function useProjectUserStories(_projectId: string): UseProjectUserStoriesResult {
  const [stories] = useState<UserStoryListItem[]>([]);
  const [loading] = useState(false);
  const [error] = useState<ApiError | Error | null>(null);

  return { stories, loading, error, refresh: () => {} };
}
