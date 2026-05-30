/**
 * Tests for web/hooks/useDocument.ts
 * Includes FCT-US002-007: race-cancellation via AbortController + stale-id ref
 */
import { renderHook, waitFor, act } from '@testing-library/react';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../test/msw/server';
import { useDocument } from './useDocument';
import { ApiError } from '../lib/api/client';

describe('useDocument', () => {
  it('returns isLoading:true initially then data when resolved', async () => {
    const { result } = renderHook(() => useDocument('d111aaaa-1111-1111-1111-111111111111'));

    expect(result.current.isLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).not.toBeNull();
    expect(result.current.data?.id).toBe('d111aaaa-1111-1111-1111-111111111111');
    expect(result.current.data?.content).toBeDefined();
    expect(result.current.error).toBeNull();
  });

  it('skips fetch when documentId is undefined', () => {
    const { result } = renderHook(() => useDocument(undefined));

    expect(result.current.isLoading).toBe(false);
    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
  });

  it('returns error when 500 response', async () => {
    const { result } = renderHook(() => useDocument('broken-document'));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toBeNull();
    expect(result.current.error).not.toBeNull();
    expect(result.current.error).toBeInstanceOf(ApiError);
  });

  it('returns error with NOT_FOUND code when 404', async () => {
    const { result } = renderHook(() => useDocument('not-found-document'));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.error).toBeInstanceOf(ApiError);
    expect((result.current.error as ApiError).code).toBe('NOT_FOUND');
  });

  it('refetch re-issues the document fetch', async () => {
    let callCount = 0;
    server.use(
      http.get('*/api/v1/documents/d111aaaa-1111-1111-1111-111111111111', () => {
        callCount++;
        return HttpResponse.json({
          id: 'd111aaaa-1111-1111-1111-111111111111',
          projectId: 'p1',
          title: 'Architecture overview',
          content: '# Architecture',
          createdAt: '2026-05-18T08:30:00Z',
          updatedAt: '2026-05-20T09:45:00Z',
        });
      })
    );

    const { result } = renderHook(() => useDocument('d111aaaa-1111-1111-1111-111111111111'));

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

  // FCT-US002-007 — Rapid click: abort prior controller before new request; only latest doc shown
  describe('FCT-US002-007 — rapid click cancels in-flight request via AbortController + stale-id', () => {
    it('switching documentId aborts prior fetch and only latest doc ends in state', async () => {
      // doc-A: delay indefinitely so the initial request stays in-flight.
      // Spy on AbortController.abort to verify the hook calls it when switching docs.
      const abortSpy = jest.spyOn(AbortController.prototype, 'abort');

      server.use(
        http.get('*/api/v1/documents/doc-A', async () => {
          // Simulate an indefinitely-pending request
          await delay('infinite');
          return HttpResponse.json({
            id: 'doc-A',
            projectId: 'p1',
            title: 'Doc A',
            content: 'Doc A content',
            createdAt: '2026-05-20T10:00:00Z',
            updatedAt: '2026-05-20T10:00:00Z',
          });
        })
      );

      let documentId = 'doc-A';
      const { result, rerender } = renderHook(() => useDocument(documentId));

      // doc-A fetch is in-flight
      expect(result.current.isLoading).toBe(true);

      // Clear calls from initial render (effect cleanup from unmounts etc.)
      abortSpy.mockClear();

      // Switch to doc-B before doc-A resolves — this triggers controllerRef.current.abort()
      act(() => {
        documentId = 'doc-B';
        rerender();
      });

      // doc-B resolves immediately via the default handler
      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      // Final state must be doc-B (the most recently requested doc)
      expect(result.current.data?.id).toBe('doc-B');

      // State must NOT contain doc-A (stale-id guard proves race safety)
      expect(result.current.data?.id).not.toBe('doc-A');

      // The hook must have called .abort() on the prior AbortController
      // (could be called once on the doc-A controller when switching to doc-B)
      expect(abortSpy).toHaveBeenCalled();

      abortSpy.mockRestore();
    });
  });
});
