import { renderHook, waitFor } from '@testing-library/react';
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
});
