/**
 * Tests for web/hooks/useRequirementDocuments.ts
 * FCT-047-018 — fetches by /projects/:pid/requirements/:rid/documents
 */
import { renderHook, waitFor, act } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { server } from '../test/msw/server';

const PROJECT_ID = '11111111-1111-1111-1111-111111111111';
const REQ_ID = 'b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f';

describe('useRequirementDocuments', () => {
  // FCT-047-018 — fetches from canonical /requirements/:rid/documents URL
  it('FCT-047-018: fetches from /projects/:pid/requirements/:rid/documents (not flat route)', async () => {
    const { useRequirementDocuments } = await import('./useRequirementDocuments');

    let canonicalHit = false;
    let flatRouteHit = false;

    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/documents', ({ params }) => {
        canonicalHit = true;
        const pid = typeof params.pid === 'string' ? params.pid : String(params.pid);
        const rid = typeof params.rid === 'string' ? params.rid : String(params.rid);
        return HttpResponse.json({
          documents: [
            {
              id: 'cccccccc-cccc-cccc-cccc-cccccccccccc',
              projectId: pid,
              requirementId: rid,
              title: 'README',
              createdAt: '2026-06-02T09:00:00Z',
              updatedAt: '2026-06-02T09:00:00Z',
            },
          ],
        });
      }),
      http.get('*/api/v1/projects/:pid/documents', () => {
        flatRouteHit = true;
        return HttpResponse.json({ documents: [] });
      })
    );

    const { result } = renderHook(() =>
      useRequirementDocuments(PROJECT_ID, REQ_ID)
    );

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(canonicalHit).toBe(true);
    expect(flatRouteHit).toBe(false);
    expect(result.current.documents).toHaveLength(1);
    expect(result.current.documents[0].requirementId).toBe(REQ_ID);
  });

  it('returns loading=true initially', async () => {
    const { useRequirementDocuments } = await import('./useRequirementDocuments');

    const { result } = renderHook(() =>
      useRequirementDocuments(PROJECT_ID, REQ_ID)
    );

    expect(result.current.loading).toBe(true);

    await waitFor(() => expect(result.current.loading).toBe(false));
  });

  it('returns empty documents array when response is empty', async () => {
    const { useRequirementDocuments } = await import('./useRequirementDocuments');

    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/documents', () => {
        return HttpResponse.json({ documents: [] });
      })
    );

    const { result } = renderHook(() =>
      useRequirementDocuments(PROJECT_ID, REQ_ID)
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.documents).toHaveLength(0);
    expect(result.current.error).toBeNull();
  });

  it('sets error state on 404 NOT_FOUND', async () => {
    const { useRequirementDocuments } = await import('./useRequirementDocuments');
    const { ApiError } = await import('../lib/api/client');

    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/documents', () => {
        return HttpResponse.json(
          { code: 'NOT_FOUND', message: 'Requirement not found' },
          { status: 404 }
        );
      })
    );

    const { result } = renderHook(() =>
      useRequirementDocuments(PROJECT_ID, REQ_ID)
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBeInstanceOf(ApiError);
    expect(result.current.documents).toHaveLength(0);
  });

  it('sets error state on 500', async () => {
    const { useRequirementDocuments } = await import('./useRequirementDocuments');
    const { ApiError } = await import('../lib/api/client');

    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/documents', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed to fetch documents' },
          { status: 500 }
        );
      })
    );

    const { result } = renderHook(() =>
      useRequirementDocuments(PROJECT_ID, REQ_ID)
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBeInstanceOf(ApiError);
  });

  it('skips fetch when projectId is undefined', async () => {
    const { useRequirementDocuments } = await import('./useRequirementDocuments');

    const { result } = renderHook(() =>
      useRequirementDocuments(undefined, REQ_ID)
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.documents).toHaveLength(0);
    expect(result.current.error).toBeNull();
  });

  it('refresh re-triggers the fetch', async () => {
    const { useRequirementDocuments } = await import('./useRequirementDocuments');

    let callCount = 0;
    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/documents', ({ params }) => {
        callCount++;
        const pid = typeof params.pid === 'string' ? params.pid : String(params.pid);
        const rid = typeof params.rid === 'string' ? params.rid : String(params.rid);
        return HttpResponse.json({
          documents: [
            {
              id: 'cccccccc-cccc-cccc-cccc-cccccccccccc',
              projectId: pid,
              requirementId: rid,
              title: 'README',
              createdAt: '2026-06-02T09:00:00Z',
              updatedAt: '2026-06-02T09:00:00Z',
            },
          ],
        });
      })
    );

    const { result } = renderHook(() =>
      useRequirementDocuments(PROJECT_ID, REQ_ID)
    );

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(callCount).toBe(1);

    act(() => {
      result.current.refresh();
    });

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(callCount).toBe(2);
  });
});
