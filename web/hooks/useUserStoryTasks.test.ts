/**
 * Tests for web/hooks/useUserStoryTasks.ts
 * Mirrors the pattern established in useDocument.test.ts
 */
import { renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../test/msw/server';
import { useUserStoryTasks } from './useUserStoryTasks';
import { ApiError } from '../lib/api/client';

const STORY_ID = 'us-001';
const TASKS_FIXTURE = [
  {
    id: 'task-001',
    userStoryId: 'us-001',
    title: 'be_basket_repo',
    description: 'Implement the basket repository layer.',
    status: 'completed',
    createdAt: '2024-01-01T10:00:00Z',
    updatedAt: '2024-01-02T11:00:00Z',
  },
  {
    id: 'task-002',
    userStoryId: 'us-001',
    title: 'fe_basket_button',
    description: 'Add the add-to-basket button.',
    status: 'in_review',
    createdAt: '2024-01-02T10:00:00Z',
    updatedAt: '2024-01-03T11:00:00Z',
  },
];

describe('useUserStoryTasks', () => {
  beforeEach(() => {
    server.use(
      http.get('*/api/v1/user-stories/us-001/tasks', () => {
        return HttpResponse.json({ tasks: TASKS_FIXTURE });
      })
    );
  });

  it('returns isLoading:true initially then tasks when resolved', async () => {
    const { result } = renderHook(() => useUserStoryTasks(STORY_ID));

    expect(result.current.isLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).not.toBeNull();
    expect(result.current.data?.tasks).toHaveLength(2);
    expect(result.current.error).toBeNull();
  });

  it('returns empty tasks array when story has no tasks', async () => {
    server.use(
      http.get('*/api/v1/user-stories/us-empty/tasks', () => {
        return HttpResponse.json({ tasks: [] });
      })
    );

    const { result } = renderHook(() => useUserStoryTasks('us-empty'));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data?.tasks).toHaveLength(0);
    expect(result.current.error).toBeNull();
  });

  it('skips fetch when storyId is undefined', () => {
    const { result } = renderHook(() => useUserStoryTasks(undefined));

    expect(result.current.isLoading).toBe(false);
    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
  });

  it('returns error when 500 response', async () => {
    server.use(
      http.get('*/api/v1/user-stories/broken-story/tasks', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed to fetch tasks' },
          { status: 500 }
        );
      })
    );

    const { result } = renderHook(() => useUserStoryTasks('broken-story'));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeInstanceOf(ApiError);
  });

  it('returns error with NOT_FOUND code when 404', async () => {
    server.use(
      http.get('*/api/v1/user-stories/not-found-story/tasks', () => {
        return HttpResponse.json(
          { code: 'NOT_FOUND', message: 'User story not found' },
          { status: 404 }
        );
      })
    );

    const { result } = renderHook(() => useUserStoryTasks('not-found-story'));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.error).toBeInstanceOf(ApiError);
    expect((result.current.error as ApiError).code).toBe('NOT_FOUND');
  });

  it('aborts prior request when storyId changes', async () => {
    server.use(
      http.get('*/api/v1/user-stories/story-slow/tasks', async () => {
        await delay('infinite');
        return HttpResponse.json({ tasks: [] });
      }),
      http.get('*/api/v1/user-stories/story-fast/tasks', () => {
        return HttpResponse.json({
          tasks: [
            {
              id: 'task-fast',
              userStoryId: 'story-fast',
              title: 'fast task',
              description: '',
              status: 'pending',
              createdAt: '2024-01-01T00:00:00Z',
              updatedAt: '2024-01-01T00:00:00Z',
            },
          ],
        });
      })
    );

    const abortSpy = jest.spyOn(AbortController.prototype, 'abort');

    let storyId = 'story-slow';
    const { result, rerender } = renderHook(() => useUserStoryTasks(storyId));

    expect(result.current.isLoading).toBe(true);
    abortSpy.mockClear();

    storyId = 'story-fast';
    rerender();

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.data?.tasks[0].id).toBe('task-fast');
    expect(abortSpy).toHaveBeenCalled();

    abortSpy.mockRestore();
  });
});
