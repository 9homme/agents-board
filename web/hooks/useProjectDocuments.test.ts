/**
 * Tests for web/hooks/useProjectDocuments.ts
 */
import { renderHook, waitFor, act } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
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
});
