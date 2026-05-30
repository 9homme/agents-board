import { fetchClient } from './client';
import { Project, ProjectsResponse } from './types';

/**
 * Fetch the list of all projects.
 * Corresponds to GET /api/v1/projects.
 */
export const fetchProjects = async (): Promise<ProjectsResponse> => {
  return fetchClient<ProjectsResponse>('/api/v1/projects');
};

/**
 * Fetch a single project by id.
 * Corresponds to GET /api/v1/projects/{id}.
 * Returns a bare Project object (not wrapped in { project: ... }).
 * Throws ApiError with code 'NOT_FOUND' when the project does not exist.
 */
export const fetchProject = async (id: string): Promise<Project> => {
  return fetchClient<Project>(`/api/v1/projects/${encodeURIComponent(id)}`);
};
