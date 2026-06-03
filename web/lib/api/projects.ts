import { fetchClient } from './client';
import { Project, ProjectsResponse } from './types';

/**
 * Fetch the list of all projects.
 * Corresponds to GET /api/v1/projects.
 *
 * @param signal - Optional AbortSignal for request cancellation (uniform lib/api surface per D-005).
 */
export const fetchProjects = async (signal?: AbortSignal): Promise<ProjectsResponse> => {
  return fetchClient<ProjectsResponse>('/api/v1/projects', { signal });
};

/**
 * Fetch a single project by id.
 * Corresponds to GET /api/v1/projects/{id}.
 * Returns a bare Project object (not wrapped in { project: ... }).
 * Throws ApiError with code 'NOT_FOUND' when the project does not exist.
 *
 * @param id     - The project id (will be URL-encoded).
 * @param signal - Optional AbortSignal for request cancellation (D-005).
 */
export const fetchProject = async (id: string, signal?: AbortSignal): Promise<Project> => {
  return fetchClient<Project>(`/api/v1/projects/${encodeURIComponent(id)}`, { signal });
};
