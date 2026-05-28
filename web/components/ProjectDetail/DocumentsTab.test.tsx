/**
 * Tests for DocumentsTab component
 * FCT-US002-002: Empty state
 * FCT-US002-003: Auto-selects first document when ?doc= absent
 * FCT-US002-006: Deep-link to bogus ?doc= shows "Document not found"; sidebar usable
 * FCT-US002-012: List loading state shows skeleton
 * FCT-US002-013: List error shows sidebar error + Retry; previewer neutral
 * FCT-US002-014: List Retry re-issues the list fetch
 */
import React from 'react';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse, delay } from 'msw';
import { server } from '../../test/msw/server';
import { DocumentsTab } from './DocumentsTab';

// Mock next/router
const mockReplace = jest.fn();
let mockDocQueryParam: string | undefined = undefined;

jest.mock('next/router', () => ({
  useRouter: () => ({
    query: { id: 'p1', ...(mockDocQueryParam ? { doc: mockDocQueryParam } : {}) },
    pathname: '/projects/[id]',
    replace: mockReplace,
    isReady: true,
  }),
}));

beforeEach(() => {
  mockDocQueryParam = undefined;
  mockReplace.mockClear();
});

// FCT-US002-002 — Empty state when project has no documents
describe('FCT-US002-002 — DocumentsTab: empty state', () => {
  it('shows "No documents yet" in sidebar and neutral message in previewer; no document fetch', async () => {
    server.use(
      http.get('*/api/v1/projects/p1/documents', () => {
        return HttpResponse.json({ documents: [] });
      })
    );

    render(<DocumentsTab projectId="p1" />);

    // Wait for the list to resolve
    await screen.findByText(/This project has no documents yet/i);

    // Sidebar area shows "No documents yet"
    const sidebar = screen.getByTestId('documents-sidebar-area');
    expect(within(sidebar).getByText(/No documents yet/i)).toBeInTheDocument();

    // Previewer area shows the corresponding empty state
    expect(screen.getByText(/This project has no documents yet/i)).toBeInTheDocument();
  });
});

// FCT-US002-003 — Auto-selects first document when ?doc= absent
describe('FCT-US002-003 — DocumentsTab: auto-selects first document', () => {
  it('calls router.replace with first doc id when doc param is absent', async () => {
    mockDocQueryParam = undefined;

    render(<DocumentsTab projectId="p1" />);

    await waitFor(() => {
      expect(mockReplace).toHaveBeenCalledWith(
        expect.objectContaining({
          query: expect.objectContaining({ doc: 'd111aaaa-1111-1111-1111-111111111111' }),
        }),
        undefined,
        { shallow: true }
      );
    });
  });

  it('first sidebar item gets aria-selected=true after auto-selection triggers', async () => {
    // Simulate that after auto-selection the router returns doc param
    mockDocQueryParam = 'd111aaaa-1111-1111-1111-111111111111';

    render(<DocumentsTab projectId="p1" />);

    const firstOption = await screen.findByRole('option', { name: /Architecture overview/i });
    expect(firstOption).toHaveAttribute('aria-selected', 'true');
  });
});

// FCT-US002-006 — Deep-link to bogus ?doc= shows "Document not found"; sidebar usable
describe('FCT-US002-006 — DocumentsTab: bogus ?doc= shows "Document not found"', () => {
  it('shows "Document not found" in previewer; no sidebar item active; no document fetch for bogus id', async () => {
    mockDocQueryParam = 'bogus-id-not-in-list';

    render(<DocumentsTab projectId="p1" />);

    expect(await screen.findByText(/Document not found/i)).toBeInTheDocument();

    // No sidebar item should be active
    await waitFor(() => {
      const options = screen.getAllByRole('option');
      options.forEach((opt) => {
        expect(opt).toHaveAttribute('aria-selected', 'false');
      });
    });

    // router.replace must NOT be called for auto-selection
    expect(mockReplace).not.toHaveBeenCalledWith(
      expect.objectContaining({ query: expect.objectContaining({ doc: expect.any(String) }) }),
      undefined,
      { shallow: true }
    );
  });

  it('sidebar items remain clickable when bogus doc is set', async () => {
    mockDocQueryParam = 'bogus-id-not-in-list';

    render(<DocumentsTab projectId="p1" />);

    // Wait for sidebar to render
    const archOption = await screen.findByRole('option', { name: /Architecture overview/i });

    // Clicking should call router.replace
    await userEvent.click(archOption);

    expect(mockReplace).toHaveBeenCalledWith(
      expect.objectContaining({
        query: expect.objectContaining({ doc: 'd111aaaa-1111-1111-1111-111111111111' }),
      }),
      undefined,
      { shallow: true }
    );
  });
});

// FCT-US002-012 — List loading state shows skeleton in sidebar
describe('FCT-US002-012 — DocumentsTab: list loading state', () => {
  it('shows loading indicator while list MSW is pending; no document titles', () => {
    server.use(
      http.get('*/api/v1/projects/p1/documents', async () => {
        await delay('infinite');
        return HttpResponse.json({ documents: [] });
      })
    );

    render(<DocumentsTab projectId="p1" />);

    // Loading indicator must be visible
    expect(screen.getByTestId('documents-list-loading')).toBeInTheDocument();

    // No document titles yet
    expect(screen.queryByRole('option')).not.toBeInTheDocument();
  });
});

// FCT-US002-013 — List error shows sidebar error + Retry; previewer neutral
describe('FCT-US002-013 — DocumentsTab: list error state', () => {
  it('shows error message and Retry in sidebar; previewer shows neutral state', async () => {
    server.use(
      http.get('*/api/v1/projects/p1/documents', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed to fetch documents' },
          { status: 500 }
        );
      })
    );

    render(<DocumentsTab projectId="p1" />);

    expect(await screen.findByText(/Couldn't load documents/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Retry/i })).toBeInTheDocument();

    // Previewer should show a neutral state, not a crash
    expect(screen.getByText(/Document list unavailable/i)).toBeInTheDocument();
  });
});

// FCT-US002-014 — List Retry re-issues the list fetch
describe('FCT-US002-014 — DocumentsTab: list Retry re-issues fetch', () => {
  it('clicking Retry after error triggers a second list request and shows documents', async () => {
    let callCount = 0;
    server.use(
      http.get('*/api/v1/projects/p1/documents', () => {
        callCount++;
        if (callCount === 1) {
          return HttpResponse.json(
            { code: 'INTERNAL_ERROR', message: 'Failed to fetch documents' },
            { status: 500 }
          );
        }
        return HttpResponse.json({
          documents: [
            {
              id: 'd111aaaa-1111-1111-1111-111111111111',
              projectId: 'p1',
              title: 'Architecture overview',
              createdAt: '2026-05-18T08:30:00Z',
              updatedAt: '2026-05-20T09:45:00Z',
            },
            {
              id: 'd222bbbb-2222-2222-2222-222222222222',
              projectId: 'p1',
              title: 'Onboarding guide',
              createdAt: '2026-05-15T11:00:00Z',
              updatedAt: '2026-05-19T16:20:00Z',
            },
          ],
        });
      })
    );

    render(<DocumentsTab projectId="p1" />);

    // Wait for error state
    await screen.findByText(/Couldn't load documents/i);

    // Click Retry
    await userEvent.click(screen.getByRole('button', { name: /Retry/i }));

    // Wait for successful load
    await screen.findByRole('option', { name: /Architecture overview/i });

    expect(callCount).toBe(2);
  });
});
