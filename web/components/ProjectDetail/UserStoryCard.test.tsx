import React from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { UserStoryCard } from './UserStoryCard';
import { UserStoryListItem } from '../../lib/api/types';

const story: UserStoryListItem = {
  id: 'us-001',
  projectId: 'p1',
  title: 'Add item to basket',
  description: 'As a shopper I want to add an item to my basket.',
  status: 'in_development',
  taskCount: 3,
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-02T09:30:00Z',
};

// FCT-005 — Card is clickable and accessible
describe('FCT-005 — UserStoryCard is clickable and accessible', () => {
  it('has role="button" and an accessible name that includes the title', () => {
    render(<UserStoryCard story={story} onSelect={() => {}} />);

    const card = screen.getByRole('button');
    expect(card).toBeInTheDocument();
    // Accessible name (aria-label or inner text) includes the title
    expect(card).toHaveAccessibleName(/Add item to basket/i);
  });

  it('calls onSelect with the story ID and button element when clicked', async () => {
    const onSelect = jest.fn();
    render(<UserStoryCard story={story} onSelect={onSelect} />);

    const card = screen.getByRole('button');
    await userEvent.click(card);

    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(onSelect).toHaveBeenCalledWith('us-001', expect.any(HTMLButtonElement));
  });

  it('calls onSelect with the story ID and button element when Enter is pressed', async () => {
    const onSelect = jest.fn();
    render(<UserStoryCard story={story} onSelect={onSelect} />);

    const card = screen.getByRole('button');
    card.focus();
    await userEvent.keyboard('{Enter}');

    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(onSelect).toHaveBeenCalledWith('us-001', expect.any(HTMLButtonElement));
  });

  it('calls onSelect with the story ID and button element when Space is pressed', async () => {
    const onSelect = jest.fn();
    render(<UserStoryCard story={story} onSelect={onSelect} />);

    const card = screen.getByRole('button');
    card.focus();
    await userEvent.keyboard(' ');

    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(onSelect).toHaveBeenCalledWith('us-001', expect.any(HTMLButtonElement));
  });
});
