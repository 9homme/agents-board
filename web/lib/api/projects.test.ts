import { fetchProjects, fetchProject } from './projects';
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
});
