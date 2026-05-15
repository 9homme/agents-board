import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
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
});
