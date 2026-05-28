/**
 * Tests for DocumentSidebar component
 * FCT-US002-001: Documents listed in received order (server provides updatedAt DESC)
 * FCT-US002-004: Clicking an item fires onSelect with the correct id
 */
import React from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DocumentSidebar } from './DocumentSidebar';
import { DocumentListItem } from '../../lib/api/types';

// FCT-US002-001 fixture: three items where two share the same updatedAt
// The sidebar renders in the order received — no client-side re-sort
const orderedDocuments: DocumentListItem[] = [
  {
    id: 'cccc0003-0000-0000-0000-000000000003',
    projectId: 'p1',
    title: 'Doc C',
    createdAt: '2026-05-15T08:00:00Z',
    updatedAt: '2026-05-20T10:00:00Z',
  },
  {
    id: 'aaaa0001-0000-0000-0000-000000000001',
    projectId: 'p1',
    title: 'Doc A',
    createdAt: '2026-05-14T08:00:00Z',
    updatedAt: '2026-05-20T10:00:00Z',
  },
  {
    id: 'bbbb0002-0000-0000-0000-000000000002',
    projectId: 'p1',
    title: 'Doc B',
    createdAt: '2026-05-13T08:00:00Z',
    updatedAt: '2026-05-19T10:00:00Z',
  },
]

const twoDocuments: DocumentListItem[] = [
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
]

// FCT-US002-001 — Sidebar: documents listed in received order (updatedAt DESC from server)
describe('FCT-US002-001 — DocumentSidebar renders documents in received order', () => {
  it('renders three items in the exact order passed via props', () => {
    render(
      <DocumentSidebar
        documents={orderedDocuments}
        selectedId={undefined}
        onSelect={jest.fn()}
      />
    );

    const options = screen.getAllByRole('option');
    expect(options).toHaveLength(3);
    expect(options[0]).toHaveTextContent('Doc C');
    expect(options[1]).toHaveTextContent('Doc A');
    expect(options[2]).toHaveTextContent('Doc B');
  });

  it('shows header with document count', () => {
    render(
      <DocumentSidebar
        documents={orderedDocuments}
        selectedId={undefined}
        onSelect={jest.fn()}
      />
    );

    expect(screen.getByText(/Documents \(3\)/i)).toBeInTheDocument();
  });
});

// FCT-US002-004 — Clicking an item calls onSelect with the correct id
describe('FCT-US002-004 — DocumentSidebar: clicking item fires onSelect', () => {
  it('clicking the second item calls onSelect with its id', async () => {
    const onSelect = jest.fn();
    render(
      <DocumentSidebar
        documents={twoDocuments}
        selectedId="d111aaaa-1111-1111-1111-111111111111"
        onSelect={onSelect}
      />
    );

    // First item should be aria-selected=true
    const firstOption = screen.getByRole('option', { name: /Architecture overview/i });
    expect(firstOption).toHaveAttribute('aria-selected', 'true');

    // Click the second item
    await userEvent.click(screen.getByRole('option', { name: /Onboarding guide/i }));

    expect(onSelect).toHaveBeenCalledWith('d222bbbb-2222-2222-2222-222222222222');
    expect(onSelect).toHaveBeenCalledTimes(1);
  });

  it('second item has aria-selected=false when first is selected', () => {
    render(
      <DocumentSidebar
        documents={twoDocuments}
        selectedId="d111aaaa-1111-1111-1111-111111111111"
        onSelect={jest.fn()}
      />
    );

    const secondOption = screen.getByRole('option', { name: /Onboarding guide/i });
    expect(secondOption).toHaveAttribute('aria-selected', 'false');
  });

  it('no item has aria-selected=true when selectedId is undefined', () => {
    render(
      <DocumentSidebar
        documents={twoDocuments}
        selectedId={undefined}
        onSelect={jest.fn()}
      />
    );

    const options = screen.getAllByRole('option');
    options.forEach((opt) => {
      expect(opt).toHaveAttribute('aria-selected', 'false');
    });
  });
});
