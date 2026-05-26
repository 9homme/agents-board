import React from 'react';
import { render, screen } from '@testing-library/react';
import { UserStoriesTab } from './UserStoriesTab';

describe('UserStoriesTab', () => {
  it('FCT-US001-012 — renders exact verbatim placeholder text (em dash, not hyphen)', () => {
    render(<UserStoriesTab />);
    // Must match character-for-character including the em dash —
    expect(
      screen.getByText('Coming soon — user stories will appear here in a future release.')
    ).toBeInTheDocument();
  });

  it('FCT-US001-013 — no network requests made on render', async () => {
    // MSW is in error mode for unhandled requests — if any request is made this test will fail
    render(<UserStoriesTab />);
    // Wait one event loop tick
    await new Promise((resolve) => setTimeout(resolve, 0));
    // If we reach here with no unhandled request error from MSW, test passes
    expect(
      screen.getByText('Coming soon — user stories will appear here in a future release.')
    ).toBeInTheDocument();
  });
});
