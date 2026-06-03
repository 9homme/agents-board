import { renderHook, waitFor, act } from '@testing-library/react';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../test/msw/server';
import { useProject } from './useProject';
import { ApiError } from '../lib/api/client';

describe('useProject hook', () => {
  it('returns isLoading:true initially and then data when resolved', async () => {
    const { result } = renderHook(() => useProject('proj-001'));

    expect(result.current.isLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).not.toBeNull();
    expect(result.current.data?.id).toBe('proj-001');
    expect(result.current.error).toBeNull();
    expect(result.current.isNotFound).toBe(false);
  });

  it('returns isNotFound:true when API returns 404 NOT_FOUND', async () => {
    const { result } = renderHook(() => useProject('no-such-project'));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toBeNull();
    expect(result.current.isNotFound).toBe(true);
    expect(result.current.error).not.toBeNull();
    expect(result.current.error).toBeInstanceOf(ApiError);
  });

  it('returns error (not isNotFound) when API returns 500', async () => {
    const { result } = renderHook(() => useProject('broken-project'));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toBeNull();
    expect(result.current.isNotFound).toBe(false);
    expect(result.current.error).not.toBeNull();
    expect(result.current.error).toBeInstanceOf(ApiError);
  });

  it('skips fetch when id is undefined (router not ready)', () => {
    const { result } = renderHook(() => useProject(undefined));

    // isLoading is false because no fetch was started
    expect(result.current.isLoading).toBe(false);
    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
    expect(result.current.isNotFound).toBe(false);
  });

  // FCT-US006-003 — useProject uses AbortController; cleanup calls controller.abort()
  describe('FCT-US006-003 — useProject AbortController cleanup on unmount', () => {
    it('calls controller.abort() when the hook unmounts mid-flight', async () => {
      const abortSpy = jest.spyOn(AbortController.prototype, 'abort');

      server.use(
        http.get('*/api/v1/projects/proj-abort-test', async () => {
          await delay('infinite');
          return HttpResponse.json({ id: 'proj-abort-test', name: 'X', description: '', createdAt: '', updatedAt: '' });
        })
      );

      const { result, unmount } = renderHook(() => useProject('proj-abort-test'));

      // Fetch is in-flight
      expect(result.current.isLoading).toBe(true);

      abortSpy.mockClear();

      // Unmount — cleanup should call controller.abort()
      unmount();

      expect(abortSpy).toHaveBeenCalled();
      abortSpy.mockRestore();
    });
  });

  // FCT-US006-004 — useProject aborts prior request on id change; only latest id ends in state
  describe('FCT-US006-004 — useProject aborts prior fetch on rapid id change', () => {
    it('switching id aborts prior fetch and only latest id ends in state', async () => {
      const abortSpy = jest.spyOn(AbortController.prototype, 'abort');

      server.use(
        http.get('*/api/v1/projects/proj-slow', async () => {
          await delay('infinite');
          return HttpResponse.json({ id: 'proj-slow', name: 'Slow', description: '', createdAt: '', updatedAt: '' });
        })
      );

      let projectId = 'proj-slow';
      const { result, rerender } = renderHook(() => useProject(projectId));

      expect(result.current.isLoading).toBe(true);
      abortSpy.mockClear();

      // Switch to proj-001 before proj-slow resolves
      act(() => {
        projectId = 'proj-001';
        rerender();
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      // Only proj-001 data should be in state
      expect(result.current.data?.id).toBe('proj-001');
      expect(result.current.data?.id).not.toBe('proj-slow');

      // abort() must have been called for the prior in-flight request
      expect(abortSpy).toHaveBeenCalled();

      abortSpy.mockRestore();
    });
  });

  // FCT-US006-009 — AbortError from useProject is NOT surfaced as error state
  describe('FCT-US006-009 — useProject does not surface AbortError as error state', () => {
    it('does not set error state when the in-flight request is aborted by id change', async () => {
      server.use(
        http.get('*/api/v1/projects/proj-will-be-aborted', async () => {
          await delay('infinite');
          return HttpResponse.json({ id: 'proj-will-be-aborted', name: 'X', description: '', createdAt: '', updatedAt: '' });
        })
      );

      let projectId: string | undefined = 'proj-will-be-aborted';
      const { result, rerender } = renderHook(() => useProject(projectId));

      expect(result.current.isLoading).toBe(true);

      // Switch id — this aborts the in-flight request
      act(() => {
        projectId = 'proj-001';
        rerender();
      });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      // Error must NOT be set (AbortError is control flow, not an error)
      expect(result.current.error).toBeNull();
      // Data is the new proj-001 result
      expect(result.current.data?.id).toBe('proj-001');
    });
  });
});
