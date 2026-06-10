import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../test/msw/server';
import Home from './index';

describe('Dashboard Page', () => {
  it('FCT-001 - Successfully load project list', async () => {
    render(<Home />);
    
    // Wait for the mock data to load
    await waitFor(() => {
      expect(screen.queryByText(/loading/i)).not.toBeInTheDocument();
    });

    // Check if the mock project is rendered
    expect(screen.getByText('Dashboard Test Project')).toBeVisible();
    expect(screen.getByText('A minimal beautiful dashboard')).toBeVisible();
  });

  it('FCT-002 - Empty state', async () => {
    server.use(
      http.get('*/api/v1/projects', () => {
        return HttpResponse.json({ projects: [] });
      })
    );

    render(<Home />);
    
    await waitFor(() => {
      expect(screen.queryByText(/loading/i)).not.toBeInTheDocument();
    });

    expect(screen.getByText(/no projects/i)).toBeVisible();
  });

  it('FCT-003 - Loading state', async () => {
    server.use(
      http.get('*/api/v1/projects', async () => {
        await delay(100);
        return HttpResponse.json({ projects: [] });
      })
    );

    render(<Home />);
    
    expect(screen.getByText(/loading/i)).toBeVisible();

    await waitFor(() => {
      expect(screen.queryByText(/loading/i)).not.toBeInTheDocument();
    });
  });

  it('FCT-004 - Error state', async () => {
    server.use(
      http.get('*/api/v1/projects', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed to fetch projects' },
          { status: 500 }
        );
      })
    );

    render(<Home />);

    await waitFor(() => {
      expect(screen.queryByText(/loading/i)).not.toBeInTheDocument();
    });

    expect(screen.getByText(/failed to load projects/i)).toBeVisible();
  });

  // ---------------------------------------------------------------------------
  // FCT-046-001 — "Add Project" button is visible on the dashboard
  // ---------------------------------------------------------------------------
  it('FCT-046-001 — Add Project button is visible on the dashboard', async () => {
    server.use(
      http.get('*/api/v1/projects', () => {
        return HttpResponse.json({ projects: [] });
      })
    );

    render(<Home />);

    await waitFor(() => {
      expect(screen.queryByText(/loading/i)).not.toBeInTheDocument();
    });

    expect(screen.getByRole('button', { name: /add project/i })).toBeInTheDocument();
  });

  // ---------------------------------------------------------------------------
  // FCT-046-002 — Clicking "Add Project" opens the dialog
  // ---------------------------------------------------------------------------
  it('FCT-046-002 — clicking Add Project opens the dialog with path and name inputs', async () => {
    const user = userEvent.setup();
    server.use(
      http.get('*/api/v1/projects', () => {
        return HttpResponse.json({ projects: [] });
      })
    );

    render(<Home />);

    await waitFor(() => {
      expect(screen.queryByText(/loading/i)).not.toBeInTheDocument();
    });

    await user.click(screen.getByRole('button', { name: /add project/i }));

    expect(screen.getByRole('dialog')).toBeVisible();
    expect(screen.getByLabelText(/path/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/name/i)).toBeInTheDocument();
  });

  // ---------------------------------------------------------------------------
  // FCT-046-012 — Happy path: 201 → dialog closes + projects list refreshes
  // ---------------------------------------------------------------------------
  it('FCT-046-012 — 201 response closes dialog and refreshes project list', async () => {
    const user = userEvent.setup();

    // Initial list: empty
    server.use(
      http.get('*/api/v1/projects', () => {
        return HttpResponse.json({ projects: [] });
      })
    );

    render(<Home />);

    await waitFor(() => {
      expect(screen.queryByText(/loading/i)).not.toBeInTheDocument();
    });

    // Open dialog
    await user.click(screen.getByRole('button', { name: /add project/i }));
    expect(screen.getByRole('dialog')).toBeVisible();

    // Fill form
    await user.type(screen.getByLabelText(/path/i), '/Users/me/workspace/agents-board');

    // After POST success, GET returns the new project
    server.use(
      http.get('*/api/v1/projects', () => {
        return HttpResponse.json({
          projects: [
            {
              id: '33333333-3333-3333-3333-333333333333',
              name: 'agents-board',
              description: '',
              path: '/Users/me/workspace/agents-board',
              createdAt: '2026-06-09T11:00:00Z',
              updatedAt: '2026-06-09T11:00:00Z',
            },
          ],
        });
      })
    );

    await user.click(screen.getByRole('button', { name: /^create project$/i }));

    // Dialog should close
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    // New project appears in the list
    await waitFor(() => {
      expect(screen.getByText('agents-board')).toBeInTheDocument();
    });
  });
});
