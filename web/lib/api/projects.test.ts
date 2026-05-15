import { fetchProjects } from './projects';
import { server } from '../../test/msw/server';
import { http, HttpResponse } from 'msw';

describe('projects API client', () => {
  it('fetchProjects returns project array', async () => {
    const data = await fetchProjects();
    expect(data.projects).toHaveLength(1);
    expect(data.projects[0].name).toBe('Dashboard Test Project');
  });

  it('fetchProjects throws an ApiError on 500', async () => {
    server.use(
      http.get('*/api/v1/projects', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed to fetch projects' },
          { status: 500 }
        );
      })
    );

    await expect(fetchProjects()).rejects.toThrow('Failed to fetch projects');
  });
});
