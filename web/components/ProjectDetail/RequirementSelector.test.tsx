/**
 * Tests for RequirementSelector component
 * FCT-047-003 — loading state
 * FCT-047-004 — populated list
 * FCT-047-005 — empty state
 * FCT-047-006 — error state
 * FCT-047-019 — name and status rendered per item
 * FCT-047-020 — "Default" requirement for migrated project
 * FCT-047-027 — keyboard navigation
 * FCT-047-028 — error state has role="alert"
 */
import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../../test/msw/server';
import { RequirementSelector } from './RequirementSelector';

const PROJECT_ID = '11111111-1111-1111-1111-111111111111';
const REQ_ID = 'b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f';

const mockOnSelect = jest.fn();

beforeEach(() => {
  mockOnSelect.mockClear();
});

// FCT-047-003 — loading state
describe('FCT-047-003 — RequirementSelector: loading state', () => {
  it('shows a loading indicator while fetching, no requirement items yet', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements', async () => {
        await delay('infinite');
        return HttpResponse.json({ requirements: [] });
      })
    );

    render(
      <RequirementSelector
        projectId={PROJECT_ID}
        selectedRequirementId={undefined}
        onSelect={mockOnSelect}
      />
    );

    // Loading indicator must be present before response resolves
    const loadingEl = screen.queryByText(/loading/i) ?? document.querySelector('[aria-busy="true"]');
    expect(loadingEl).toBeInTheDocument();

    // No requirement items yet
    expect(screen.queryByText('Default')).not.toBeInTheDocument();
  });
});

// FCT-047-004 — populated list
describe('FCT-047-004 — RequirementSelector: populated list', () => {
  it('renders requirement names on success', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements', ({ params }) => {
        const pid = typeof params.pid === 'string' ? params.pid : String(params.pid);
        return HttpResponse.json({
          requirements: [
            {
              id: REQ_ID,
              projectId: pid,
              name: 'Default',
              description: '',
              status: 'draft',
              createdAt: '2026-06-09T10:00:00Z',
              updatedAt: '2026-06-09T10:00:00Z',
            },
          ],
        });
      })
    );

    render(
      <RequirementSelector
        projectId={PROJECT_ID}
        selectedRequirementId={undefined}
        onSelect={mockOnSelect}
      />
    );

    expect(await screen.findByText('Default')).toBeInTheDocument();
  });
});

// FCT-047-005 — empty state
describe('FCT-047-005 — RequirementSelector: empty state', () => {
  it('shows "No requirements yet" when project has no requirements', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements', () => {
        return HttpResponse.json({ requirements: [] });
      })
    );

    render(
      <RequirementSelector
        projectId={PROJECT_ID}
        selectedRequirementId={undefined}
        onSelect={mockOnSelect}
      />
    );

    expect(await screen.findByText(/no requirements yet/i)).toBeInTheDocument();
    expect(screen.queryByText('Default')).not.toBeInTheDocument();
  });
});

// FCT-047-006 — error state
describe('FCT-047-006 — RequirementSelector: error state', () => {
  it('shows inline error message on fetch failure', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed to fetch requirements' },
          { status: 500 }
        );
      })
    );

    render(
      <RequirementSelector
        projectId={PROJECT_ID}
        selectedRequirementId={undefined}
        onSelect={mockOnSelect}
      />
    );

    // Error message is visible in the requirements area
    const errorEl = await screen.findByRole('alert');
    expect(errorEl).toBeInTheDocument();
    expect(errorEl.textContent).toBeTruthy();
  });
});

// FCT-047-019 — name and status per item
describe('FCT-047-019 — RequirementSelector: name and status rendered per item', () => {
  it('renders both requirement names and status labels', async () => {
    const REQ_ID_2 = 'deadbeef-dead-dead-dead-deadbeefcafe';

    server.use(
      http.get('*/api/v1/projects/:pid/requirements', ({ params }) => {
        const pid = typeof params.pid === 'string' ? params.pid : String(params.pid);
        return HttpResponse.json({
          requirements: [
            {
              id: REQ_ID,
              projectId: pid,
              name: 'Default',
              description: '',
              status: 'draft',
              createdAt: '2026-06-09T10:00:00Z',
              updatedAt: '2026-06-09T10:00:00Z',
            },
            {
              id: REQ_ID_2,
              projectId: pid,
              name: 'Authentication',
              description: '',
              status: 'in_progress',
              createdAt: '2026-06-09T11:00:00Z',
              updatedAt: '2026-06-09T11:00:00Z',
            },
          ],
        });
      })
    );

    render(
      <RequirementSelector
        projectId={PROJECT_ID}
        selectedRequirementId={undefined}
        onSelect={mockOnSelect}
      />
    );

    expect(await screen.findByText('Default')).toBeInTheDocument();
    expect(await screen.findByText('Authentication')).toBeInTheDocument();

    // Status labels (draft and in_progress) must be visible
    expect(screen.getByText(/draft/i)).toBeInTheDocument();
    expect(screen.getByText(/in.?progress/i)).toBeInTheDocument();
  });
});

// FCT-047-020 — "Default" requirement for migrated project
describe('FCT-047-020 — RequirementSelector: Default requirement for migrated project', () => {
  it('renders single "Default" requirement without crash', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements', ({ params }) => {
        const pid = typeof params.pid === 'string' ? params.pid : String(params.pid);
        return HttpResponse.json({
          requirements: [
            {
              id: REQ_ID,
              projectId: pid,
              name: 'Default',
              description: '',
              status: 'draft',
              createdAt: '2026-06-09T10:00:00Z',
              updatedAt: '2026-06-09T10:00:00Z',
            },
          ],
        });
      })
    );

    render(
      <RequirementSelector
        projectId={PROJECT_ID}
        selectedRequirementId={undefined}
        onSelect={mockOnSelect}
      />
    );

    expect(await screen.findByText('Default')).toBeInTheDocument();
  });
});

// FCT-047-027 — keyboard navigation
describe('FCT-047-027 — RequirementSelector: keyboard navigation', () => {
  it('requirement item is selectable by keyboard (Enter/Space calls onSelect)', async () => {
    const user = userEvent.setup();

    server.use(
      http.get('*/api/v1/projects/:pid/requirements', ({ params }) => {
        const pid = typeof params.pid === 'string' ? params.pid : String(params.pid);
        return HttpResponse.json({
          requirements: [
            {
              id: REQ_ID,
              projectId: pid,
              name: 'Default',
              description: '',
              status: 'draft',
              createdAt: '2026-06-09T10:00:00Z',
              updatedAt: '2026-06-09T10:00:00Z',
            },
          ],
        });
      })
    );

    render(
      <RequirementSelector
        projectId={PROJECT_ID}
        selectedRequirementId={undefined}
        onSelect={mockOnSelect}
      />
    );

    // Wait for the item to render
    const item = await screen.findByText('Default');
    // Tab to the item and press Enter/Space
    const button = item.closest('button') ?? item;
    button.focus();
    await user.keyboard('{Enter}');

    expect(mockOnSelect).toHaveBeenCalledWith(REQ_ID);
  });
});

// FCT-047-028 — error state has role="alert"
describe('FCT-047-028 — RequirementSelector: error state is announced', () => {
  it('error container has role="alert" or live region', async () => {
    server.use(
      http.get('*/api/v1/projects/:pid/requirements', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed to fetch requirements' },
          { status: 500 }
        );
      })
    );

    render(
      <RequirementSelector
        projectId={PROJECT_ID}
        selectedRequirementId={undefined}
        onSelect={mockOnSelect}
      />
    );

    await waitFor(() => {
      expect(screen.queryByRole('alert')).toBeInTheDocument();
    });

    const alertEl = screen.getByRole('alert');
    expect(alertEl).toBeInTheDocument();
  });
});
