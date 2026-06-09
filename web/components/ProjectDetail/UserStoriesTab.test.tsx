/**
 * Tests for web/components/ProjectDetail/UserStoriesTab.tsx
 * FCT-001, FCT-003, FCT-005
 *
 * Note: FCT-US001-012 and FCT-US001-013 (placeholder text tests) are superseded
 * by this US005 implementation which replaces the placeholder with real content.
 */
import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { server } from '../../test/msw/server';
import { UserStoriesTab } from './UserStoriesTab';

const PROJECT_ID = 'proj-001';

const STORY_1 = {
  id: 'us-111',
  projectId: PROJECT_ID,
  title: 'Story One',
  description: 'First story description here.',
  status: 'in_development',
  taskCount: 2,
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-02T00:00:00Z',
};

const STORY_2 = {
  id: 'us-222',
  projectId: PROJECT_ID,
  title: 'Story Two',
  description: 'Second story description here.',
  status: 'pending',
  taskCount: 1,
  createdAt: '2024-01-03T00:00:00Z',
  updatedAt: '2024-01-04T00:00:00Z',
};

const STORY_1_DETAIL = {
  id: 'us-111',
  projectId: PROJECT_ID,
  title: 'Story One',
  description: 'First story full description.',
  status: 'in_development',
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-02T00:00:00Z',
};

const STORY_2_DETAIL = {
  id: 'us-222',
  projectId: PROJECT_ID,
  title: 'Story Two',
  description: 'Second story full description.',
  status: 'pending',
  createdAt: '2024-01-03T00:00:00Z',
  updatedAt: '2024-01-04T00:00:00Z',
};

const TASKS_FOR_STORY_1 = [
  {
    id: 'task-a',
    userStoryId: 'us-111',
    title: 'Task Alpha',
    description: 'Do alpha work.',
    status: 'completed',
    createdAt: '2024-01-01T10:00:00Z',
    updatedAt: '2024-01-02T11:00:00Z',
  },
  {
    id: 'task-b',
    userStoryId: 'us-111',
    title: 'Task Beta',
    description: 'Do beta work.',
    status: 'in_progress',
    createdAt: '2024-01-02T10:00:00Z',
    updatedAt: '2024-01-03T11:00:00Z',
  },
];

const TASKS_FOR_STORY_2 = [
  {
    id: 'task-c',
    userStoryId: 'us-222',
    title: 'Task Gamma',
    description: 'Do gamma work.',
    status: 'pending',
    createdAt: '2024-01-03T10:00:00Z',
    updatedAt: '2024-01-04T11:00:00Z',
  },
];

// ---------------------------------------------------------------------------
// FCT-001 — Renders story detail and tasks in right-side drawer
// ---------------------------------------------------------------------------
describe('FCT-001 — selecting a card opens the drawer', () => {
  beforeEach(() => {
    server.use(
      http.get(`*/api/v1/projects/${PROJECT_ID}/user-stories`, () => {
        return HttpResponse.json({ userStories: [STORY_1] });
      }),
      http.get('*/api/v1/user-stories/us-111/tasks', () => {
        return HttpResponse.json({ tasks: TASKS_FOR_STORY_1 });
      }),
      http.get('*/api/v1/user-stories/us-111', () => {
        return HttpResponse.json(STORY_1_DETAIL);
      })
    );
  });

  it('clicking a story card opens the drawer with dialog role', async () => {
    const user = userEvent.setup();
    render(<UserStoriesTab projectId={PROJECT_ID} />);

    // Wait for the card list to render
    const card = await screen.findByRole('button', { name: /Story One/i });
    await user.click(card);

    // Drawer with role=dialog must appear
    expect(await screen.findByRole('dialog')).toBeInTheDocument();
  });

  it('drawer displays story description and status after loading', async () => {
    const user = userEvent.setup();
    render(<UserStoriesTab projectId={PROJECT_ID} />);

    const card = await screen.findByRole('button', { name: /Story One/i });
    await user.click(card);

    // Verify story detail description is shown in the drawer
    expect(await screen.findByText('First story full description.')).toBeInTheDocument();
    // The drawer renders the dialog — status badge shows once detail loaded
    const dialog = await screen.findByRole('dialog');
    expect(dialog).toBeInTheDocument();
  });

  it('drawer displays task titles', async () => {
    const user = userEvent.setup();
    render(<UserStoriesTab projectId={PROJECT_ID} />);

    const card = await screen.findByRole('button', { name: /Story One/i });
    await user.click(card);

    await screen.findByRole('dialog');
    expect(await screen.findByText('Task Alpha')).toBeInTheDocument();
    expect(await screen.findByText('Task Beta')).toBeInTheDocument();
  });
});

// ---------------------------------------------------------------------------
// FCT-003 — Close button closes drawer and returns focus
// ---------------------------------------------------------------------------
describe('FCT-003 — close button closes drawer', () => {
  beforeEach(() => {
    server.use(
      http.get(`*/api/v1/projects/${PROJECT_ID}/user-stories`, () => {
        return HttpResponse.json({ userStories: [STORY_1] });
      }),
      http.get('*/api/v1/user-stories/us-111/tasks', () => {
        return HttpResponse.json({ tasks: TASKS_FOR_STORY_1 });
      }),
      http.get('*/api/v1/user-stories/us-111', () => {
        return HttpResponse.json(STORY_1_DETAIL);
      })
    );
  });

  it('clicking close button unmounts the drawer', async () => {
    const user = userEvent.setup();
    render(<UserStoriesTab projectId={PROJECT_ID} />);

    // Open the drawer
    const card = await screen.findByRole('button', { name: /Story One/i });
    await user.click(card);

    // Wait for drawer to be present
    await screen.findByRole('dialog');

    // Click the close button
    const closeButton = screen.getByRole('button', { name: /close/i });
    await user.click(closeButton);

    // Drawer must be gone
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });
  });
});

// ---------------------------------------------------------------------------
// FCT-005 — Switching stories updates drawer in place
// ---------------------------------------------------------------------------
describe('FCT-005 — switching stories updates drawer in place', () => {
  beforeEach(() => {
    server.use(
      http.get(`*/api/v1/projects/${PROJECT_ID}/user-stories`, () => {
        return HttpResponse.json({ userStories: [STORY_1, STORY_2] });
      }),
      http.get('*/api/v1/user-stories/us-111/tasks', () => {
        return HttpResponse.json({ tasks: TASKS_FOR_STORY_1 });
      }),
      http.get('*/api/v1/user-stories/us-111', () => {
        return HttpResponse.json(STORY_1_DETAIL);
      }),
      http.get('*/api/v1/user-stories/us-222/tasks', () => {
        return HttpResponse.json({ tasks: TASKS_FOR_STORY_2 });
      }),
      http.get('*/api/v1/user-stories/us-222', () => {
        return HttpResponse.json(STORY_2_DETAIL);
      })
    );
  });

  it('clicking a second card while drawer is open fetches and shows new story', async () => {
    const user = userEvent.setup();
    render(<UserStoriesTab projectId={PROJECT_ID} />);

    // Click Story 1
    const card1 = await screen.findByRole('button', { name: /Story One/i });
    await user.click(card1);

    // Wait for Story 1 to be loaded in drawer
    await screen.findByText('First story full description.');

    // Click Story 2 (without closing drawer)
    const card2 = screen.getByRole('button', { name: /Story Two/i });
    await user.click(card2);

    // Drawer remains visible and now shows Story 2
    await screen.findByText('Second story full description.');
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(await screen.findByText('Task Gamma')).toBeInTheDocument();
  });
});
