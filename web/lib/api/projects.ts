import { fetchClient } from './client';
import { ProjectsResponse } from './types';

export const fetchProjects = async (): Promise<ProjectsResponse> => {
  return fetchClient<ProjectsResponse>('/api/v1/projects');
};
