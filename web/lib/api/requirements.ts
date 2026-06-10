import { fetchClient } from './client';
import { RequirementsResponse } from './types';

/**
 * Fetch the list of requirements for a project.
 * Corresponds to GET /api/v1/projects/{projectId}/requirements (§4).
 *
 * Order is createdAt ASC (enforced by the backend).
 * Empty project → `{ "requirements": [] }`.
 *
 * @param projectId - The project id (will be URL-encoded).
 * @param signal    - Optional AbortSignal for request cancellation.
 */
export const fetchProjectRequirements = async (
  projectId: string,
  signal?: AbortSignal
): Promise<RequirementsResponse> => {
  return fetchClient<RequirementsResponse>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/requirements`,
    { signal }
  );
};
