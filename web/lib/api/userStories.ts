import { fetchClient } from './client';
import { UserStoriesListResponse, UserStory, TasksListResponse } from './types';

/**
 * Fetches user stories for a given project.
 */
export async function fetchProjectUserStories(
  projectId: string,
  signal?: AbortSignal
): Promise<UserStoriesListResponse> {
  return fetchClient<UserStoriesListResponse>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/user-stories`,
    { signal }
  );
}

/**
 * Fetches a single user story by ID.
 */
export async function fetchUserStory(
  id: string,
  signal?: AbortSignal
): Promise<UserStory> {
  return fetchClient<UserStory>(`/api/v1/user-stories/${encodeURIComponent(id)}`, {
    signal,
  });
}

/**
 * Fetches tasks for a given user story.
 */
export async function fetchUserStoryTasks(
  userStoryId: string,
  signal?: AbortSignal
): Promise<TasksListResponse> {
  return fetchClient<TasksListResponse>(
    `/api/v1/user-stories/${encodeURIComponent(userStoryId)}/tasks`,
    { signal }
  );
}
