import { fetchClient } from './client';
import { UserStoriesListResponse, UserStory, TasksListResponse } from './types';

/**
 * Fetch the list of user stories scoped to a requirement.
 * Corresponds to GET /api/v1/projects/{projectId}/requirements/{requirementId}/user-stories (§6).
 *
 * Each item includes `requirementId`. Order is createdAt DESC.
 *
 * @param projectId     - The project id (will be URL-encoded).
 * @param requirementId - The requirement id (will be URL-encoded).
 * @param signal        - Optional AbortSignal for request cancellation.
 */
export const fetchRequirementUserStories = async (
  projectId: string,
  requirementId: string,
  signal?: AbortSignal
): Promise<UserStoriesListResponse> => {
  return fetchClient<UserStoriesListResponse>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/requirements/${encodeURIComponent(requirementId)}/user-stories`,
    { signal }
  );
};

/**
 * Fetch the list of user stories for a project, including taskCount.
 * Corresponds to GET /api/v1/projects/{projectId}/user-stories.
 *
 * Order is createdAt DESC (enforced by the backend).
 *
 * @param projectId - The project id (will be URL-encoded).
 * @param signal    - Optional AbortSignal for request cancellation.
 */
export const fetchProjectUserStories = async (
  projectId: string,
  signal?: AbortSignal
): Promise<UserStoriesListResponse> => {
  return fetchClient<UserStoriesListResponse>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/user-stories`,
    { signal }
  );
};

/**
 * Fetch a single user story's detail (no tasks embedded, no taskCount).
 * Corresponds to GET /api/v1/projects/{projectId}/requirements/{requirementId}/user-stories/{storyId} (§6).
 */
export const fetchUserStory = async (
  projectId: string,
  requirementId: string,
  storyId: string,
  signal?: AbortSignal
): Promise<UserStory> => {
  return fetchClient<UserStory>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/requirements/${encodeURIComponent(requirementId)}/user-stories/${encodeURIComponent(storyId)}`,
    { signal }
  );
};

/**
 * Fetch all tasks for a user story.
 * Corresponds to GET /api/v1/projects/{projectId}/requirements/{requirementId}/user-stories/{storyId}/tasks (§6).
 */
export const fetchUserStoryTasks = async (
  projectId: string,
  requirementId: string,
  storyId: string,
  signal?: AbortSignal
): Promise<TasksListResponse> => {
  return fetchClient<TasksListResponse>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/requirements/${encodeURIComponent(requirementId)}/user-stories/${encodeURIComponent(storyId)}/tasks`,
    { signal }
  );
};
