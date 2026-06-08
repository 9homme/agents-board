import React from 'react';
import { render, screen } from '@testing-library/react';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../../test/msw/server';
import { UserStoryCardList } from './UserStoryCardList';

const projectId = 'p1';

const longDesc =
  'As a shopper I want to add an item to my basket so that I can purchase it later.'.repeat(2);

const shortDesc = 'Short description.';

describe('UserStoryCardList', () => {
  // FCT-001 — Renders user story cards with correct details
  describe('FCT-001 — renders user story cards with correct details', () => {
    it('renders 2 cards with title, status, taskCount, and full description text', async () => {
      server.use(
        http.get(`*/api/v1/projects/${projectId}/user-stories`, () => {
          return HttpResponse.json({
            userStories: [
              {
                id: 'us1',
                projectId,
                title: 'Add item to basket',
                description: longDesc,
                status: 'in_development',
                taskCount: 3,
                createdAt: '2024-01-01T00:00:00Z',
                updatedAt: '2024-01-02T09:30:00Z',
              },
              {
                id: 'us2',
                projectId,
                title: 'Remove item from basket',
                description: shortDesc,
                status: 'pending',
                taskCount: 1,
                createdAt: '2024-01-03T00:00:00Z',
                updatedAt: '2024-01-04T09:30:00Z',
              },
            ],
          });
        })
      );

      render(<UserStoryCardList projectId={projectId} onSelect={() => {}} />);

      // Wait for cards to load — expect 2 cards
      const cards = await screen.findAllByRole('button');
      expect(cards).toHaveLength(2);

      // First card details
      expect(screen.getByText('Add item to basket')).toBeInTheDocument();
      expect(screen.getByText('3 tasks')).toBeInTheDocument();
      expect(screen.getByText(/in_development/i)).toBeInTheDocument();
      // Full text in DOM (CSS truncation not testable in jsdom)
      expect(screen.getByText(longDesc)).toBeInTheDocument();

      // Second card details
      expect(screen.getByText('Remove item from basket')).toBeInTheDocument();
      expect(screen.getByText(shortDesc)).toBeInTheDocument();
    });
  });

  // FCT-002 — Empty state when no stories
  describe('FCT-002 — empty state when no stories', () => {
    it('shows "No user stories yet for this project." and renders no cards', async () => {
      server.use(
        http.get(`*/api/v1/projects/${projectId}/user-stories`, () => {
          return HttpResponse.json({ userStories: [] });
        })
      );

      render(<UserStoryCardList projectId={projectId} onSelect={() => {}} />);

      expect(
        await screen.findByText(/No user stories yet for this project/i)
      ).toBeInTheDocument();

      expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });
  });

  // FCT-003 — Loading state
  describe('FCT-003 — loading state', () => {
    it('shows a loading indicator while fetching and no stale cards or error text', async () => {
      server.use(
        http.get(`*/api/v1/projects/${projectId}/user-stories`, async () => {
          await delay('infinite');
          return HttpResponse.json({ userStories: [] });
        })
      );

      render(<UserStoryCardList projectId={projectId} onSelect={() => {}} />);

      // Loading indicator must be visible immediately
      expect(screen.getByText(/Loading/i)).toBeInTheDocument();

      // No stale cards or error text during loading
      expect(screen.queryByRole('button')).not.toBeInTheDocument();
      expect(screen.queryByText(/Couldn't load user stories/i)).not.toBeInTheDocument();
    });
  });

  // FCT-004 — Error state
  describe('FCT-004 — error state', () => {
    it('shows "Couldn\'t load user stories." on API 500 and renders no cards', async () => {
      server.use(
        http.get(`*/api/v1/projects/${projectId}/user-stories`, () => {
          return HttpResponse.json(
            { code: 'INTERNAL_ERROR', message: 'Failed to fetch user stories' },
            { status: 500 }
          );
        })
      );

      render(<UserStoryCardList projectId={projectId} onSelect={() => {}} />);

      expect(
        await screen.findByText(/Couldn't load user stories/i)
      ).toBeInTheDocument();

      expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });
  });
});
