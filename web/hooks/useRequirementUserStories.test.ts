/**
 * Tests for web/hooks/useRequirementUserStories.ts
 * FCT-047-017 — fetches by /projects/:pid/requirements/:rid/user-stories
 */
import { renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { server } from '../test/msw/server';

const PROJECT_ID = '11111111-1111-1111-1111-111111111111';
const REQ_ID = 'b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f';

describe('useRequirementUserStories', () => {
  // FCT-047-017 — fetches from canonical /requirements/:rid/user-stories URL
  it('FCT-047-017: fetches from /projects/:pid/requirements/:rid/user-stories (not flat route)', async () => {
    const { useRequirementUserStories } = await import('./useRequirementUserStories');

    let canonicalHit = false;
    let flatRouteHit = false;

    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories', ({ params }) => {
        canonicalHit = true;
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
      }),
      http.get('*/api/v1/projects/:pid/user-stories', () => {
        flatRouteHit = true;
        return HttpResponse.json({ userStories: [] });
      })
    );

    const { result } = renderHook(() =>
      useRequirementUserStories(PROJECT_ID, REQ_ID)
    );

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(canonicalHit).toBe(true);
    expect(flatRouteHit).toBe(false);
    expect(result.current.stories).toHaveLength(1);
    expect(result.current.stories[0].requirementId).toBe(REQ_ID);
  });

  it('returns loading=true initially', async () => {
    const { useRequirementUserStories } = await import('./useRequirementUserStories');

    const { result } = renderHook(() =>
      useRequirementUserStories(PROJECT_ID, REQ_ID)
    );

    expect(result.current.loading).toBe(true);

    await waitFor(() => expect(result.current.loading).toBe(false));
  });

  it('returns empty stories array when response is empty', async () => {
    const { useRequirementUserStories } = await import('./useRequirementUserStories');

    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories', () => {
        return HttpResponse.json({ userStories: [] });
      })
    );

    const { result } = renderHook(() =>
      useRequirementUserStories(PROJECT_ID, REQ_ID)
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.stories).toHaveLength(0);
    expect(result.current.error).toBeNull();
  });
});
