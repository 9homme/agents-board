/**
 * Tests for web/lib/api/requirements.ts
 * FCT-047-022 — fetchProjectRequirements sends correct URL and parses §4 response
 */
import { server } from '../../test/msw/server';
import { http, HttpResponse, delay } from 'msw';
import { ApiError } from './client';

const PROJECT_ID = '11111111-1111-1111-1111-111111111111';
const REQ_ID = 'b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f';

describe('fetchProjectRequirements', () => {
  // FCT-047-022 — sends correct URL
  it('FCT-047-022: sends GET /api/v1/projects/:pid/requirements and parses §4 shape', async () => {
    // dynamic import so the module doesn't exist yet (red phase)
    const { fetchProjectRequirements } = await import('./requirements');

    let capturedUrl: string | undefined;
    server.use(
      http.get('*/api/v1/projects/:pid/requirements', ({ request, params }) => {
        capturedUrl = request.url;
        const pid = typeof params.pid === 'string' ? params.pid : String(params.pid);
        return HttpResponse.json({
          requirements: [
            {
              id: REQ_ID,
              projectId: pid,
              name: 'Default',
              description: '',
              status: 'draft',
              createdAt: '2026-06-09T10:00:00Z',
              updatedAt: '2026-06-09T10:00:00Z',
            },
          ],
        });
      })
    );

    const result = await fetchProjectRequirements(PROJECT_ID);

    expect(capturedUrl).toContain(`/api/v1/projects/${PROJECT_ID}/requirements`);
    expect(result.requirements).toHaveLength(1);
    const req = result.requirements[0];
    expect(req.id).toBe(REQ_ID);
    expect(req.projectId).toBe(PROJECT_ID);
    expect(req.name).toBe('Default');
    expect(req.description).toBe('');
    expect(req.status).toBe('draft');
    expect(req.createdAt).toBe('2026-06-09T10:00:00Z');
    expect(req.updatedAt).toBe('2026-06-09T10:00:00Z');
  });

  it('returns empty requirements array on 200 with empty list', async () => {
    const { fetchProjectRequirements } = await import('./requirements');

    server.use(
      http.get('*/api/v1/projects/:pid/requirements', () => {
        return HttpResponse.json({ requirements: [] });
      })
    );

    const result = await fetchProjectRequirements(PROJECT_ID);
    expect(result.requirements).toHaveLength(0);
  });

  it('throws ApiError with NOT_FOUND on 404', async () => {
    const { fetchProjectRequirements } = await import('./requirements');

    server.use(
      http.get('*/api/v1/projects/:pid/requirements', () => {
        return HttpResponse.json(
          { code: 'NOT_FOUND', message: 'Project not found' },
          { status: 404 }
        );
      })
    );

    await expect(fetchProjectRequirements(PROJECT_ID)).rejects.toBeInstanceOf(ApiError);
    try {
      await fetchProjectRequirements(PROJECT_ID);
    } catch (err) {
      expect((err as ApiError).code).toBe('NOT_FOUND');
    }
  });

  it('throws ApiError on 500', async () => {
    const { fetchProjectRequirements } = await import('./requirements');

    server.use(
      http.get('*/api/v1/projects/:pid/requirements', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed to fetch requirements' },
          { status: 500 }
        );
      })
    );

    await expect(fetchProjectRequirements(PROJECT_ID)).rejects.toBeInstanceOf(ApiError);
  });

  it('passes AbortSignal to fetch', async () => {
    const { fetchProjectRequirements } = await import('./requirements');

    const controller = new AbortController();
    server.use(
      http.get('*/api/v1/projects/:pid/requirements', async () => {
        await delay('infinite');
        return HttpResponse.json({ requirements: [] });
      })
    );

    const promise = fetchProjectRequirements(PROJECT_ID, controller.signal);
    controller.abort();
    await expect(promise).rejects.toThrow();
  });
});
