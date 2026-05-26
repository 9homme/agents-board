import React from 'react';
import { render, screen } from '@testing-library/react';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../../test/msw/server';
import ProjectDetailPage from './[id]';

// Mock next/router
const mockReplace = jest.fn();
let mockQuery: Record<string, string> = {};

jest.mock('next/router', () => ({
  useRouter: () => ({
    query: mockQuery,
    pathname: '/projects/[id]',
    replace: mockReplace,
    isReady: true,
  }),
}));

beforeEach(() => {
  mockQuery = {};
  mockReplace.mockClear();
});

describe('ProjectDetailPage', () => {
  describe('FCT-US001-005 — header renders project name and description', () => {
    it('shows project name in h1 and description text after loading', async () => {
      mockQuery = { id: 'proj-001' };
      render(<ProjectDetailPage />);

      const heading = await screen.findByRole('heading', {
        level: 1,
        name: /E-commerce Website/i,
      });
      expect(heading).toBeInTheDocument();
      expect(
        await screen.findByText(/A new online store for electronics/i)
      ).toBeInTheDocument();
    });
  });

  describe('FCT-US001-007 — loading skeleton visible during fetch', () => {
    it('shows a loading skeleton before the fetch resolves', () => {
      mockQuery = { id: 'proj-001' };

      server.use(
        http.get('*/api/v1/projects/proj-001', async () => {
          await delay('infinite');
          return HttpResponse.json({});
        })
      );

      render(<ProjectDetailPage />);

      // Before fetch resolves, skeleton must be present
      expect(screen.getByTestId('project-header-skeleton')).toBeInTheDocument();
      // Project name must not be present yet
      expect(screen.queryByText('E-commerce Website')).not.toBeInTheDocument();
    });
  });

  describe('FCT-US001-008 — tab switcher: Documents active by default (no ?tab= param)', () => {
    it('renders two tabs with Documents aria-selected=true when no tab param', async () => {
      mockQuery = { id: 'proj-001' };
      render(<ProjectDetailPage />);

      // Wait for page load
      await screen.findByRole('heading', { level: 1 });

      const tablist = screen.getByRole('tablist');
      expect(tablist).toBeInTheDocument();

      const tabs = screen.getAllByRole('tab');
      expect(tabs).toHaveLength(2);

      expect(screen.getByRole('tab', { name: /Documents/i })).toHaveAttribute(
        'aria-selected',
        'true'
      );
      expect(screen.getByRole('tab', { name: /User Stories/i })).toHaveAttribute(
        'aria-selected',
        'false'
      );
    });
  });

  describe('FCT-US001-009 — clicking "User Stories" updates URL', () => {
    it('calls router.replace with tab=user-stories and shallow:true on click', async () => {
      mockQuery = { id: 'proj-001' };
      render(<ProjectDetailPage />);

      await screen.findByRole('heading', { level: 1 });

      const userStoriesTab = screen.getByRole('tab', { name: /User Stories/i });
      userStoriesTab.click();

      expect(mockReplace).toHaveBeenCalledWith(
        expect.objectContaining({
          query: expect.objectContaining({ tab: 'user-stories' }),
        }),
        undefined,
        { shallow: true }
      );
    });
  });

  describe('FCT-US001-011 — ?tab=user-stories activates User Stories tab on mount', () => {
    it('shows User Stories tab as active when query.tab = "user-stories"', async () => {
      mockQuery = { id: 'proj-001', tab: 'user-stories' };
      render(<ProjectDetailPage />);

      await screen.findByRole('heading', { level: 1 });

      expect(screen.getByRole('tab', { name: /User Stories/i })).toHaveAttribute(
        'aria-selected',
        'true'
      );

      expect(
        screen.getByText(
          'Coming soon — user stories will appear here in a future release.'
        )
      ).toBeInTheDocument();
    });
  });

  describe('FCT-US001-014 — 404 renders "Project not found" + "Back to dashboard" + hides tabs', () => {
    it('shows "Project not found" and hides tab switcher on 404', async () => {
      mockQuery = { id: 'no-such-project' };
      render(<ProjectDetailPage />);

      expect(await screen.findByText(/Project not found/i)).toBeInTheDocument();

      // Tab switcher must be hidden
      expect(screen.queryByRole('tablist')).not.toBeInTheDocument();

      // "Back to dashboard" link must be present
      const backLink = screen.getByRole('link', { name: /Back to dashboard/i });
      expect(backLink).toHaveAttribute('href', '/');
    });
  });

  describe('FCT-US001-015 — 500 renders "Failed to load project" + "Back to dashboard"', () => {
    it('shows "Failed to load project" and back link on 500', async () => {
      mockQuery = { id: 'broken-project' };
      render(<ProjectDetailPage />);

      expect(await screen.findByText(/Failed to load project/i)).toBeInTheDocument();

      const backLink = screen.getByRole('link', { name: /Back to dashboard/i });
      expect(backLink).toBeInTheDocument();
    });
  });
});
