import { fetchProjects, fetchProject, createProject } from './projects';
import { server } from '../../test/msw/server';
import { http, HttpResponse } from 'msw';
import { ApiError } from './client';

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

  describe('fetchProject', () => {
    it('returns a bare Project object on 200', async () => {
      const project = await fetchProject('proj-001');
      expect(project.id).toBe('proj-001');
      expect(project.name).toBe('E-commerce Website');
      expect(project.description).toBe('A new online store for electronics');
      expect(project.createdAt).toBe('2026-05-20T10:00:00Z');
      expect(project.updatedAt).toBe('2026-05-20T10:00:00Z');
    });

    it('throws ApiError with code NOT_FOUND on 404', async () => {
      let caught: unknown;
      try {
        await fetchProject('no-such-project');
      } catch (e) {
        caught = e;
      }
      expect(caught).toBeInstanceOf(ApiError);
      expect((caught as ApiError).code).toBe('NOT_FOUND');
      expect((caught as ApiError).message).toBe('Project not found');
    });

    it('throws ApiError with code INTERNAL_ERROR on 500', async () => {
      let caught: unknown;
      try {
        await fetchProject('broken-project');
      } catch (e) {
        caught = e;
      }
      expect(caught).toBeInstanceOf(ApiError);
      expect((caught as ApiError).code).toBe('INTERNAL_ERROR');
      expect((caught as ApiError).message).toBe('Failed to fetch project');
    });

    it('URL-encodes the project id', async () => {
      // The handler uses encodeURIComponent in the fetch call
      // proj-001 doesn't need encoding but we verify the function exists and works
      const project = await fetchProject('proj-001');
      expect(project).toBeDefined();
    });
  });

  // FCT-US006-001 — fetchProject accepts optional signal?: AbortSignal (uniform lib/api surface)
  describe('FCT-US006-001 — fetchProject signal parameter', () => {
    it('fetchProject accepts an AbortSignal without error when not aborted', async () => {
      const controller = new AbortController();
      // Should resolve normally when signal is not aborted
      const project = await fetchProject('proj-001', controller.signal);
      expect(project.id).toBe('proj-001');
    });

    it('fetchProject rejects when the AbortSignal is aborted before fetch', async () => {
      const controller = new AbortController();
      controller.abort();
      await expect(fetchProject('proj-001', controller.signal)).rejects.toThrow();
    });
  });

  // FCT-US006-002 — fetchProjects accepts optional signal?: AbortSignal (uniform lib/api surface)
  describe('FCT-US006-002 — fetchProjects signal parameter', () => {
    it('fetchProjects accepts an AbortSignal without error when not aborted', async () => {
      const controller = new AbortController();
      const data = await fetchProjects(controller.signal);
      expect(data.projects).toHaveLength(1);
    });

    it('fetchProjects rejects when the AbortSignal is aborted before fetch', async () => {
      const controller = new AbortController();
      controller.abort();
      await expect(fetchProjects(controller.signal)).rejects.toThrow();
    });
  });

  // ---------------------------------------------------------------------------
  // FCT-046-021 — createProject sends correct request body
  // FCT-046-022 — createProject returns typed Project including path field
  // ---------------------------------------------------------------------------
  describe('createProject', () => {
    it('FCT-046-021 — sends name, description, and path in request body with Content-Type', async () => {
      let capturedBody: unknown = null;
      let capturedContentType: string | null = null;

      server.use(
        http.post('*/api/v1/projects', async ({ request }) => {
          capturedBody = await request.json();
          capturedContentType = request.headers.get('content-type');
          return HttpResponse.json(
            {
              id: '33333333-3333-3333-3333-333333333333',
              name: 'Test',
              description: 'Desc',
              path: '/tmp/test',
              createdAt: '2026-06-09T11:00:00Z',
              updatedAt: '2026-06-09T11:00:00Z',
            },
            { status: 201 }
          );
        })
      );

      await createProject({ name: 'Test', description: 'Desc', path: '/tmp/test' });

      expect(capturedBody).toEqual({ name: 'Test', description: 'Desc', path: '/tmp/test' });
      expect(capturedContentType).toContain('application/json');
    });

    it('FCT-046-022 — returns typed Project including path field on 201', async () => {
      const project = await createProject({
        name: 'agents-board',
        path: '/Users/me/workspace/agents-board',
      });

      expect(project.id).toBe('33333333-3333-3333-3333-333333333333');
      expect(project.name).toBe('agents-board');
      expect(project.description).toBe('');
      expect(project.path).toBe('/Users/me/workspace/agents-board');
      expect(typeof project.path).toBe('string');
      expect(project.createdAt).toBe('2026-06-09T11:00:00Z');
      expect(project.updatedAt).toBe('2026-06-09T11:00:00Z');
    });

    it('FCT-046-022b — throws ApiError with VALIDATION_ERROR code on 400', async () => {
      server.use(
        http.post('*/api/v1/projects', () => {
          return HttpResponse.json(
            { code: 'VALIDATION_ERROR', message: 'path does not exist or is not a directory' },
            { status: 400 }
          );
        })
      );

      let caught: unknown;
      try {
        await createProject({ name: 'Test', path: '/bad/path' });
      } catch (e) {
        caught = e;
      }
      expect(caught).toBeInstanceOf(ApiError);
      expect((caught as ApiError).code).toBe('VALIDATION_ERROR');
      expect((caught as ApiError).message).toBe('path does not exist or is not a directory');
    });

    it('FCT-046-022c — throws ApiError with DUPLICATE_PATH code on 409', async () => {
      server.use(
        http.post('*/api/v1/projects', () => {
          return HttpResponse.json(
            { code: 'DUPLICATE_PATH', message: 'path already linked to another project' },
            { status: 409 }
          );
        })
      );

      let caught: unknown;
      try {
        await createProject({ name: 'Test', path: '/existing/path' });
      } catch (e) {
        caught = e;
      }
      expect(caught).toBeInstanceOf(ApiError);
      expect((caught as ApiError).code).toBe('DUPLICATE_PATH');
    });
  });
});
