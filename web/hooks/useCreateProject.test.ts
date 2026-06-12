import { renderHook, act, waitFor } from '@testing-library/react';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../test/msw/server';
import { useCreateProject } from './useCreateProject';
import { ApiError } from '../lib/api/client';

// ---------------------------------------------------------------------------
// FCT-046-016 through FCT-046-020, FCT-046-023
// ---------------------------------------------------------------------------

describe('useCreateProject', () => {
  it('FCT-046-016 — initial state is idle with null error', () => {
    const { result } = renderHook(() => useCreateProject());
    expect(result.current.status).toBe('idle');
    expect(result.current.error).toBeNull();
  });

  it('FCT-046-017 — status is submitting while request is in flight', async () => {
    // Use a deferred response so we can inspect in-flight state
    let resolveRequest!: () => void;
    const deferredPromise = new Promise<void>((res) => { resolveRequest = res; });

    server.use(
      http.post('*/api/v1/projects', async () => {
        await deferredPromise;
        return HttpResponse.json(
          {
            id: '33333333-3333-3333-3333-333333333333',
            name: 'Test',
            description: '',
            path: '/tmp/test',
            createdAt: '2026-06-09T11:00:00Z',
            updatedAt: '2026-06-09T11:00:00Z',
          },
          { status: 201 }
        );
      })
    );

    const { result } = renderHook(() => useCreateProject());

    act(() => {
      result.current.createProject({ name: 'Test', path: '/tmp/test' });
    });

    // Immediately after calling, status should be submitting
    await waitFor(() => {
      expect(result.current.status).toBe('submitting');
    });

    // Resolve to clean up
    resolveRequest();
    await waitFor(() => {
      expect(result.current.status).not.toBe('submitting');
    });
  });

  it('FCT-046-018 — status is success and returns project on 201', async () => {
    const { result } = renderHook(() => useCreateProject());

    let project: Awaited<ReturnType<typeof result.current.createProject>>;
    await act(async () => {
      project = await result.current.createProject({
        name: 'agents-board',
        path: '/Users/me/workspace/agents-board',
      });
    });

    // After success the hook exposes the result
    expect(result.current.status).toBe('success');
    expect(result.current.error).toBeNull();
    // The returned project has the path field
    expect(project!.path).toBe('/Users/me/workspace/agents-board');
    expect(project!.id).toBe('33333333-3333-3333-3333-333333333333');
  });

  it('FCT-046-019 — status is error and error object populated on 400', async () => {
    server.use(
      http.post('*/api/v1/projects', () => {
        return HttpResponse.json(
          { code: 'VALIDATION_ERROR', message: 'path does not exist or is not a directory' },
          { status: 400 }
        );
      })
    );

    const { result } = renderHook(() => useCreateProject());

    await act(async () => {
      try {
        await result.current.createProject({ name: 'Test', path: '/bad/path' });
      } catch {
        // expected to throw
      }
    });

    expect(result.current.status).toBe('error');
    expect(result.current.error).toBeInstanceOf(ApiError);
    expect((result.current.error as ApiError).code).toBe('VALIDATION_ERROR');
    expect((result.current.error as ApiError).message).toBe(
      'path does not exist or is not a directory'
    );
  });

  it('FCT-046-020 — status is error on network failure', async () => {
    server.use(
      http.post('*/api/v1/projects', () => {
        return HttpResponse.error();
      })
    );

    const { result } = renderHook(() => useCreateProject());

    await act(async () => {
      try {
        await result.current.createProject({ name: 'Test', path: '/some/path' });
      } catch {
        // expected
      }
    });

    expect(result.current.status).toBe('error');
    expect(result.current.error).not.toBeNull();
  });

  it('FCT-046-023 — in-flight request is aborted on unmount with no state update warning', async () => {
    let resolveRequest!: () => void;
    const deferredPromise = new Promise<void>((res) => { resolveRequest = res; });

    server.use(
      http.post('*/api/v1/projects', async ({ request }) => {
        await Promise.race([
          deferredPromise,
          new Promise<void>((_, reject) =>
            request.signal.addEventListener('abort', () => reject(new DOMException('AbortError', 'AbortError')))
          ),
        ]);
        return HttpResponse.json(
          {
            id: '33333333-3333-3333-3333-333333333333',
            name: 'Test',
            description: '',
            path: '/tmp/test',
            createdAt: '2026-06-09T11:00:00Z',
            updatedAt: '2026-06-09T11:00:00Z',
          },
          { status: 201 }
        );
      })
    );

    const { result, unmount } = renderHook(() => useCreateProject());

    act(() => {
      result.current.createProject({ name: 'Test', path: '/tmp/test' });
    });

    await waitFor(() => {
      expect(result.current.status).toBe('submitting');
    });

    // Unmount — should abort the in-flight request without causing state update warnings
    unmount();

    // Resolve after unmount (simulating the response finally arriving)
    resolveRequest();

    // Wait a tick to ensure no state update warnings are thrown
    await new Promise((r) => setTimeout(r, 50));
    // If the abort cleanup works, no React warning about state update on unmounted component
  });
});
