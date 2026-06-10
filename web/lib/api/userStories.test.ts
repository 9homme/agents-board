/**
 * Tests for web/lib/api/userStories.ts — requirement-scoped fetcher
 * FCT-047-023 — fetchRequirementUserStories sends correct URL
 */
import { server } from '../../test/msw/server';
import { http, HttpResponse, delay } from 'msw';
import { ApiError } from './client';

const PROJECT_ID = '11111111-1111-1111-1111-111111111111';
const REQ_ID = 'b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f';

describe('fetchRequirementUserStories', () => {
  // FCT-047-023 — sends GET /api/v1/projects/:pid/requirements/:rid/user-stories
  it('FCT-047-023: sends correct canonical URL and parses §6 shape', async () => {
    const { fetchRequirementUserStories } = await import('./userStories');

    let capturedUrl: string | undefined;
    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories', ({ request, params }) => {
        capturedUrl = request.url;
        const pid = typeof params.pid === 'string' ? params.pid : String(params.pid);
        const rid = typeof params.rid === 'string' ? params.rid : String(params.rid);
        return HttpResponse.json({
          userStories: [
            {
              id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
              projectId: pid,
              requirementId: rid,
              title: 'Add item to basket',
              description: '',
              status: 'in_progress',
              taskCount: 3,
              createdAt: '2026-06-02T09:00:00Z',
              updatedAt: '2026-06-02T09:00:00Z',
            },
          ],
        });
      })
    );

    const result = await fetchRequirementUserStories(PROJECT_ID, REQ_ID);

    expect(capturedUrl).toContain(
      `/api/v1/projects/${PROJECT_ID}/requirements/${REQ_ID}/user-stories`
    );
    expect(result.userStories).toHaveLength(1);
    const story = result.userStories[0];
    expect(story.id).toBe('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa');
    expect(story.projectId).toBe(PROJECT_ID);
    expect(story.requirementId).toBe(REQ_ID);
    expect(story.title).toBe('Add item to basket');
    expect(story.status).toBe('in_progress');
    expect(story.taskCount).toBe(3);
  });

  it('does NOT call the old flat route /api/v1/projects/:pid/user-stories', async () => {
    const { fetchRequirementUserStories } = await import('./userStories');

    let flatRouteHit = false;
    server.use(
      http.get('*/api/v1/projects/:pid/user-stories', () => {
        flatRouteHit = true;
        return HttpResponse.json({ userStories: [] });
      })
    );

    await fetchRequirementUserStories(PROJECT_ID, REQ_ID);
    expect(flatRouteHit).toBe(false);
  });

  it('returns empty array on 200 with no stories', async () => {
    const { fetchRequirementUserStories } = await import('./userStories');

    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories', () => {
        return HttpResponse.json({ userStories: [] });
      })
    );

    const result = await fetchRequirementUserStories(PROJECT_ID, REQ_ID);
    expect(result.userStories).toHaveLength(0);
  });

  it('throws ApiError with NOT_FOUND on 404', async () => {
    const { fetchRequirementUserStories } = await import('./userStories');

    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories', () => {
        return HttpResponse.json(
          { code: 'NOT_FOUND', message: 'Requirement not found' },
          { status: 404 }
        );
      })
    );

    await expect(fetchRequirementUserStories(PROJECT_ID, REQ_ID)).rejects.toBeInstanceOf(ApiError);
    try {
      await fetchRequirementUserStories(PROJECT_ID, REQ_ID);
    } catch (err) {
      expect((err as ApiError).code).toBe('NOT_FOUND');
    }
  });

  it('throws ApiError on 500', async () => {
    const { fetchRequirementUserStories } = await import('./userStories');

    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed to fetch user stories' },
          { status: 500 }
        );
      })
    );

    await expect(fetchRequirementUserStories(PROJECT_ID, REQ_ID)).rejects.toBeInstanceOf(ApiError);
  });

  it('passes AbortSignal to fetch', async () => {
    const { fetchRequirementUserStories } = await import('./userStories');

    const controller = new AbortController();
    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories', async () => {
        await delay('infinite');
        return HttpResponse.json({ userStories: [] });
      })
    );

    const promise = fetchRequirementUserStories(PROJECT_ID, REQ_ID, controller.signal);
    controller.abort();
    await expect(promise).rejects.toThrow();
  });
});
