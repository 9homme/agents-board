/**
 * Tests for web/hooks/useProjectRequirements.ts
 * FCT-047-011 — loading=true initially
 * FCT-047-012 — success: requirements array populated
 * FCT-047-013 — empty: empty array, no error
 * FCT-047-014 — 404 project not found → error state
 * FCT-047-015 — 500 error → error state
 * FCT-047-016 — AbortController cleanup on unmount
 */
import { renderHook, waitFor, act } from '@testing-library/react';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../test/msw/server';
import { ApiError } from '../lib/api/client';

const PROJECT_ID = '11111111-1111-1111-1111-111111111111';
const REQ_ID = 'b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f';

describe('useProjectRequirements', () => {
  // FCT-047-011 — loading=true initially, no error, no requirements
  it('FCT-047-011: loading=true initially before response resolves', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements', async () => {
        await delay('infinite');
        return HttpResponse.json({ requirements: [] });
      })
    );

    const { useProjectRequirements } = await import('./useProjectRequirements');
    const { result } = renderHook(() => useProjectRequirements(PROJECT_ID));

    expect(result.current.loading).toBe(true);
    expect(result.current.error).toBeNull();
    expect(result.current.requirements).toHaveLength(0);
  });

  // FCT-047-012 — success: requirements populated with all §4 fields
  it('FCT-047-012: requirements array populated on success with correct §4 fields', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements', ({ params }) => {
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

    const { useProjectRequirements } = await import('./useProjectRequirements');
    const { result } = renderHook(() => useProjectRequirements(PROJECT_ID));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBeNull();
    expect(result.current.requirements).toHaveLength(1);

    const req = result.current.requirements[0];
    expect(req.id).toBe(REQ_ID);
    expect(req.projectId).toBe(PROJECT_ID);
    expect(req.name).toBe('Default');
    expect(req.description).toBe('');
    expect(req.status).toBe('draft');
    expect(req.createdAt).toBe('2026-06-09T10:00:00Z');
    expect(req.updatedAt).toBe('2026-06-09T10:00:00Z');
  });

  // FCT-047-013 — empty array, no error
  it('FCT-047-013: empty requirements array returned, no error', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements', () => {
        return HttpResponse.json({ requirements: [] });
      })
    );

    const { useProjectRequirements } = await import('./useProjectRequirements');
    const { result } = renderHook(() => useProjectRequirements(PROJECT_ID));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.requirements).toEqual([]);
    expect(result.current.error).toBeNull();
  });

  // FCT-047-014 — 404 project not found → error state
  it('FCT-047-014: error state from 404 project not found', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements', () => {
        return HttpResponse.json(
          { code: 'NOT_FOUND', message: 'Project not found' },
          { status: 404 }
        );
      })
    );

    const { useProjectRequirements } = await import('./useProjectRequirements');
    const { result } = renderHook(() => useProjectRequirements(PROJECT_ID));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).not.toBeNull();
    expect(result.current.error).toBeInstanceOf(ApiError);
    expect((result.current.error as ApiError).code).toBe('NOT_FOUND');
    expect(result.current.requirements).toHaveLength(0);
  });

  // FCT-047-015 — 500 error → error state
  it('FCT-047-015: error state from 500 server error', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed to fetch requirements' },
          { status: 500 }
        );
      })
    );

    const { useProjectRequirements } = await import('./useProjectRequirements');
    const { result } = renderHook(() => useProjectRequirements(PROJECT_ID));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).not.toBeNull();
    expect(result.current.error).toBeInstanceOf(ApiError);
    expect(result.current.requirements).toHaveLength(0);
  });

  // skip fetch when projectId is undefined
  it('skips fetch and sets loading=false when projectId is undefined', async () => {
    const { useProjectRequirements } = await import('./useProjectRequirements');

    const { result } = renderHook(() => useProjectRequirements(undefined));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.requirements).toHaveLength(0);
    expect(result.current.error).toBeNull();
  });

  // refresh re-triggers fetch
  it('refresh re-triggers the fetch', async () => {
    const { useProjectRequirements } = await import('./useProjectRequirements');

    let callCount = 0;
    server.use(
      http.get('*/api/v1/projects/:pid/requirements', ({ params }) => {
        callCount++;
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

    const { result } = renderHook(() => useProjectRequirements(PROJECT_ID));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(callCount).toBe(1);

    act(() => {
      result.current.refresh();
    });

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(callCount).toBe(2);
  });

  // FCT-047-016 — AbortController cleanup on unmount
  it('FCT-047-016: abort is called on unmount, no state-update-after-unmount warning', async () => {
    const abortSpy = jest.spyOn(AbortController.prototype, 'abort');

    server.use(
      http.get('*/api/v1/projects/abort-test-pid/requirements', async () => {
        await delay('infinite');
        return HttpResponse.json({ requirements: [] });
      })
    );

    const { useProjectRequirements } = await import('./useProjectRequirements');
    const { result, unmount } = renderHook(() =>
      useProjectRequirements('abort-test-pid')
    );

    expect(result.current.loading).toBe(true);

    abortSpy.mockClear();
    unmount();

    expect(abortSpy).toHaveBeenCalled();
    abortSpy.mockRestore();
  });
});
