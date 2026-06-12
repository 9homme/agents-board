/**
 * Tests for web/hooks/useUserStory.ts
 * Mirrors the pattern established in useDocument.test.ts
 */
import { renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../test/msw/server';
import { useUserStory } from './useUserStory';
import { ApiError } from '../lib/api/client';

const PROJECT_ID = 'proj-001';
const REQUIREMENT_ID = 'req-001';
const STORY_ID = 'us-001';
const STORY_FIXTURE = {
  id: 'us-001',
  projectId: 'proj-001',
  requirementId: 'req-001',
  title: 'Add item to basket',
  description: 'As a shopper I want to add an item to my basket.',
  status: 'in_development',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-02T09:30:00Z',
};

describe('useUserStory', () => {
  beforeEach(() => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories/us-001', () => {
        return HttpResponse.json(STORY_FIXTURE);
      })
    );
  });

  it('returns isLoading:true initially then data when resolved', async () => {
    const { result } = renderHook(() => useUserStory(PROJECT_ID, REQUIREMENT_ID, STORY_ID));

    expect(result.current.isLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).not.toBeNull();
    expect(result.current.data?.id).toBe(STORY_ID);
    expect(result.current.error).toBeNull();
  });

  it('skips fetch when params are undefined', () => {
    const { result } = renderHook(() => useUserStory(undefined, undefined, undefined));

    expect(result.current.isLoading).toBe(false);
    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
  });

  it('returns error when 500 response', async () => {
    const { result } = renderHook(() => useUserStory(PROJECT_ID, REQUIREMENT_ID, 'broken-story'));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeInstanceOf(ApiError);
  });

  it('returns error with NOT_FOUND code when 404', async () => {
    const { result } = renderHook(() => useUserStory(PROJECT_ID, REQUIREMENT_ID, 'not-found-story'));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.error).toBeInstanceOf(ApiError);
    expect((result.current.error as ApiError).code).toBe('NOT_FOUND');
  });

  it('aborts prior request when storyId changes', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories/story-slow', async () => {
        await delay('infinite');
        return HttpResponse.json({ id: 'story-slow' });
      }),
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories/story-fast', () => {
        return HttpResponse.json({
          id: 'story-fast',
          projectId: PROJECT_ID,
          requirementId: REQUIREMENT_ID,
          title: 'Fast Story',
          description: '',
          status: 'pending',
          createdAt: '2024-01-01T00:00:00Z',
          updatedAt: '2024-01-01T00:00:00Z',
        });
      })
    );

    const abortSpy = jest.spyOn(AbortController.prototype, 'abort');

    let storyId = 'story-slow';
    const { result, rerender } = renderHook(() => useUserStory(PROJECT_ID, REQUIREMENT_ID, storyId));

    expect(result.current.isLoading).toBe(true);
    abortSpy.mockClear();

    storyId = 'story-fast';
    rerender();

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data?.id).toBe('story-fast');
    expect(abortSpy).toHaveBeenCalled();

    abortSpy.mockRestore();
  });
});
