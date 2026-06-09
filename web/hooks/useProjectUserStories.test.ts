/**
 * Tests for web/hooks/useProjectUserStories.ts
 */
import { renderHook, waitFor, act } from '@testing-library/react';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../test/msw/server';
import { useProjectUserStories } from './useProjectUserStories';
import { ApiError } from '../lib/api/client';

describe('useProjectUserStories', () => {
  it('returns loading:true initially then stories when resolved', async () => {
    const { result } = renderHook(() => useProjectUserStories('p1'));

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.stories).toHaveLength(1);
    expect(result.current.error).toBeNull();
  });

  it('returns error when 500 response', async () => {
    const { result } = renderHook(() => useProjectUserStories('broken-project'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.stories).toHaveLength(0);
    expect(result.current.error).not.toBeNull();
    expect(result.current.error).toBeInstanceOf(ApiError);
  });

  it('refresh re-issues the list fetch', async () => {
    let callCount = 0;
    server.use(
      http.get('*/api/v1/projects/p1/user-stories', () => {
        callCount++;
        return HttpResponse.json({
          userStories: [
            {
              id: 'us1',
              projectId: 'p1',
              title: 'Story 1',
              description: 'A story.',
              status: 'pending',
              taskCount: 0,
              createdAt: '2024-01-01T00:00:00Z',
              updatedAt: '2024-01-01T00:00:00Z',
            },
          ],
        });
      })
    );

    const { result } = renderHook(() => useProjectUserStories('p1'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(callCount).toBe(1);

    act(() => {
      result.current.refresh();
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(callCount).toBe(2);
  });

  // FCT-006 — Hook aborts pending fetch on remount/id change
  describe('FCT-006 — hook aborts pending fetch on unmount or projectId change', () => {
    it('calls controller.abort() when the hook unmounts mid-flight', async () => {
      const abortSpy = jest.spyOn(AbortController.prototype, 'abort');

      server.use(
        http.get('*/api/v1/projects/proj-stories-abort/user-stories', async () => {
          await delay('infinite');
          return HttpResponse.json({ userStories: [] });
        })
      );

      const { result, unmount } = renderHook(() =>
        useProjectUserStories('proj-stories-abort')
      );

      expect(result.current.loading).toBe(true);

      abortSpy.mockClear();

      // Unmount — cleanup should call controller.abort()
      unmount();

      expect(abortSpy).toHaveBeenCalled();
      abortSpy.mockRestore();
    });

    it('aborts prior fetch when projectId changes and only latest projectId ends in state', async () => {
      const abortSpy = jest.spyOn(AbortController.prototype, 'abort');

      server.use(
        http.get('*/api/v1/projects/proj-slow-stories/user-stories', async () => {
          await delay('infinite');
          return HttpResponse.json({ userStories: [] });
        })
      );

      let projectId = 'proj-slow-stories';
      const { result, rerender } = renderHook(() => useProjectUserStories(projectId));

      expect(result.current.loading).toBe(true);
      abortSpy.mockClear();

      // Switch to p1 before proj-slow-stories resolves
      act(() => {
        projectId = 'p1';
        rerender();
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Only p1 data should be in state
      expect(result.current.stories).toHaveLength(1);

      // abort() must have been called for the prior in-flight request
      expect(abortSpy).toHaveBeenCalled();

      abortSpy.mockRestore();
    });

    it('does not set error state when the in-flight request is aborted by projectId change', async () => {
      server.use(
        http.get('*/api/v1/projects/proj-stories-will-abort/user-stories', async () => {
          await delay('infinite');
          return HttpResponse.json({ userStories: [] });
        })
      );

      let projectId = 'proj-stories-will-abort';
      const { result, rerender } = renderHook(() => useProjectUserStories(projectId));

      expect(result.current.loading).toBe(true);

      // Switch projectId — aborts the in-flight request
      act(() => {
        projectId = 'p1';
        rerender();
      });

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Error must NOT be set (AbortError is control flow, not user-visible)
      expect(result.current.error).toBeNull();
      // Data is from the new p1 request
      expect(result.current.stories).toHaveLength(1);
    });
  });
});
