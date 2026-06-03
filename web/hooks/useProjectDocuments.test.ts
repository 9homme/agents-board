/**
 * Tests for web/hooks/useProjectDocuments.ts
 */
import { renderHook, waitFor, act } from '@testing-library/react';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../test/msw/server';
import { useProjectDocuments } from './useProjectDocuments';
import { ApiError } from '../lib/api/client';

describe('useProjectDocuments', () => {
  it('returns isLoading:true initially then data when resolved', async () => {
    const { result } = renderHook(() => useProjectDocuments('p1'));

    expect(result.current.isLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).not.toBeNull();
    expect(result.current.data?.documents).toHaveLength(2);
    expect(result.current.error).toBeNull();
  });

  it('skips fetch when projectId is undefined', () => {
    const { result } = renderHook(() => useProjectDocuments(undefined));

    expect(result.current.isLoading).toBe(false);
    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
  });

  it('returns error when 500 response', async () => {
    const { result } = renderHook(() => useProjectDocuments('broken-project'));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toBeNull();
    expect(result.current.error).not.toBeNull();
    expect(result.current.error).toBeInstanceOf(ApiError);
  });

  it('refetch re-issues the list fetch', async () => {
    let callCount = 0;
    server.use(
      http.get('*/api/v1/projects/p1/documents', () => {
        callCount++;
        return HttpResponse.json({
          documents: [
            {
              id: 'd111aaaa-1111-1111-1111-111111111111',
              projectId: 'p1',
              title: 'Architecture overview',
              createdAt: '2026-05-18T08:30:00Z',
              updatedAt: '2026-05-20T09:45:00Z',
            },
          ],
        });
      })
    );

    const { result } = renderHook(() => useProjectDocuments('p1'));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(callCount).toBe(1);

    act(() => {
      result.current.refetch();
    });

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(callCount).toBe(2);
  });

  it('returns error when 404 (NOT_FOUND)', async () => {
    const { result } = renderHook(() => useProjectDocuments('ghost-project'));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.error).toBeInstanceOf(ApiError);
    expect((result.current.error as ApiError).code).toBe('NOT_FOUND');
  });

  // FCT-US006-006 — useProjectDocuments uses AbortController; cleanup calls controller.abort()
  describe('FCT-US006-006 — useProjectDocuments AbortController cleanup on unmount', () => {
    it('calls controller.abort() when the hook unmounts mid-flight', async () => {
      const abortSpy = jest.spyOn(AbortController.prototype, 'abort');

      server.use(
        http.get('*/api/v1/projects/proj-docs-abort/documents', async () => {
          await delay('infinite');
          return HttpResponse.json({ documents: [] });
        })
      );

      const { result, unmount } = renderHook(() => useProjectDocuments('proj-docs-abort'));

      expect(result.current.isLoading).toBe(true);

      abortSpy.mockClear();

      // Unmount — cleanup should call controller.abort()
      unmount();

      expect(abortSpy).toHaveBeenCalled();
      abortSpy.mockRestore();
    });
  });

  // FCT-US006-007 — useProjectDocuments aborts prior request on rapid projectId change
  describe('FCT-US006-007 — useProjectDocuments aborts prior fetch on rapid projectId change', () => {
    it('switching projectId aborts prior fetch and only latest projectId ends in state', async () => {
      const abortSpy = jest.spyOn(AbortController.prototype, 'abort');

      server.use(
        http.get('*/api/v1/projects/proj-slow-docs/documents', async () => {
          await delay('infinite');
          return HttpResponse.json({ documents: [] });
        })
      );

      let projectId = 'proj-slow-docs';
      const { result, rerender } = renderHook(() => useProjectDocuments(projectId));

      expect(result.current.isLoading).toBe(true);
      abortSpy.mockClear();

      // Switch to p1 before proj-slow-docs resolves
      act(() => {
        projectId = 'p1';
        rerender();
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      // Only p1 data should be in state
      expect(result.current.data?.documents).toHaveLength(2);

      // abort() must have been called for the prior in-flight request
      expect(abortSpy).toHaveBeenCalled();

      abortSpy.mockRestore();
    });
  });

  // FCT-US006-010 — useProjectDocuments does NOT surface AbortError as error state
  describe('FCT-US006-010 — useProjectDocuments does not surface AbortError as error state', () => {
    it('does not set error state when the in-flight request is aborted by projectId change', async () => {
      server.use(
        http.get('*/api/v1/projects/proj-docs-will-abort/documents', async () => {
          await delay('infinite');
          return HttpResponse.json({ documents: [] });
        })
      );

      let projectId: string | undefined = 'proj-docs-will-abort';
      const { result, rerender } = renderHook(() => useProjectDocuments(projectId));

      expect(result.current.isLoading).toBe(true);

      // Switch projectId — this aborts the in-flight request
      act(() => {
        projectId = 'p1';
        rerender();
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      // Error must NOT be set (AbortError is control flow, not a user-visible error)
      expect(result.current.error).toBeNull();
      // Data is from the new p1 request
      expect(result.current.data?.documents).toHaveLength(2);
    });
  });
});
