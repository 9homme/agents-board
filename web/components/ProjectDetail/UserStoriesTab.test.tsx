import React from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { UserStoriesTab } from './UserStoriesTab';

describe('UserStoriesTab', () => {
  it('FCT-US004-001 — renders the user story card list', async () => {
    render(<UserStoriesTab projectId="proj-001" />);

    // Should show loading state initially
    expect(screen.getByText(/loading/i)).toBeInTheDocument();

    // Wait for the stories to load (mocked in handlers.ts)
    expect(await screen.findByText('Add item to basket')).toBeInTheDocument();
    expect(screen.getByText('3 tasks')).toBeInTheDocument();
  });

  it('renders the tabpanel container', () => {
    render(<UserStoriesTab projectId="proj-001" />);
    expect(screen.getByRole('tabpanel')).toBeInTheDocument();
  });

  it('onSelect stub fires silently (no-op) when a story card is clicked', async () => {
    render(<UserStoriesTab projectId="proj-001" />);

    // Wait for card to appear
    const card = await screen.findByRole('button');
    // Clicking invokes the no-op onSelect — should not throw
    await userEvent.click(card);
    // No assertion needed beyond "did not throw"; presence of card is the invariant
    expect(card).toBeInTheDocument();
  });
});
