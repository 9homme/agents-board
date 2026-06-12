/**
 * Tests for web/components/ProjectDetail/UserStoryDrawer.tsx
 * FCT-002, FCT-004, FCT-006, FCT-007
 */
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../../test/msw/server';
import { UserStoryDrawer } from './UserStoryDrawer';

const PROJECT_ID = 'proj-001';
const REQUIREMENT_ID = 'req-001';
const STORY_ID = 'us-001';
const STORY_FIXTURE = {
  id: 'us-001',
  projectId: 'proj-001',
  requirementId: 'req-001',
  title: 'Add item to basket',
  description: 'As a shopper I want to add an item to my basket so that I can purchase it later.',
  status: 'in_development',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-02T09:30:00Z',
};
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
];

const mockOnClose = jest.fn();

// ---------------------------------------------------------------------------
// FCT-002 — Drawer shows empty state when no tasks
// ---------------------------------------------------------------------------
describe('FCT-002 — empty state when no tasks', () => {
  it('shows "No tasks for this story." when tasks array is empty', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories/us-001', () => {
        return HttpResponse.json(STORY_FIXTURE);
      }),
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories/us-001/tasks', () => {
        return HttpResponse.json({ tasks: [] });
      })
    );

    render(<UserStoryDrawer projectId={PROJECT_ID} requirementId={REQUIREMENT_ID} storyId={STORY_ID} onClose={mockOnClose} />);

    expect(await screen.findByText(/No tasks for this story/i)).toBeInTheDocument();
  });
});

// ---------------------------------------------------------------------------
// FCT-004 — Escape key closes drawer
// ---------------------------------------------------------------------------
describe('FCT-004 — Escape key closes drawer', () => {
  it('calls onClose when Escape is pressed', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories/us-001', () => {
        return HttpResponse.json(STORY_FIXTURE);
      }),
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories/us-001/tasks', () => {
        return HttpResponse.json({ tasks: TASKS_FIXTURE });
      })
    );

    const onClose = jest.fn();
    render(<UserStoryDrawer projectId={PROJECT_ID} requirementId={REQUIREMENT_ID} storyId={STORY_ID} onClose={onClose} />);

    fireEvent.keyDown(document, { key: 'Escape', code: 'Escape' });

    expect(onClose).toHaveBeenCalledTimes(1);
  });
});

// ---------------------------------------------------------------------------
// FCT-006 — Loading state in drawer
// ---------------------------------------------------------------------------
describe('FCT-006 — loading state in drawer', () => {
  it('shows spinner while requests are pending', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories/us-001', async () => {
        await delay(500);
        return HttpResponse.json(STORY_FIXTURE);
      }),
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories/us-001/tasks', async () => {
        await delay(500);
        return HttpResponse.json({ tasks: TASKS_FIXTURE });
      })
    );

    render(<UserStoryDrawer projectId={PROJECT_ID} requirementId={REQUIREMENT_ID} storyId={STORY_ID} onClose={mockOnClose} />);

    // Spinner should be visible immediately (loading state)
    expect(screen.getByRole('status')).toBeInTheDocument();
  });
});

// ---------------------------------------------------------------------------
// FCT-007 — Error state in drawer
// ---------------------------------------------------------------------------
describe('FCT-007 — error state in drawer', () => {
  it('shows error message and close button when story fetch fails', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories/us-001', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed to fetch user story' },
          { status: 500 }
        );
      }),
      http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories/us-001/tasks', () => {
        return HttpResponse.json({ tasks: [] });
      })
    );

    render(<UserStoryDrawer projectId={PROJECT_ID} requirementId={REQUIREMENT_ID} storyId={STORY_ID} onClose={mockOnClose} />);

    expect(
      await screen.findByText(/Couldn't load this user story/i)
    ).toBeInTheDocument();

    // Close button must still be visible in error state
    expect(screen.getByRole('button', { name: /close/i })).toBeInTheDocument();
  });
});
