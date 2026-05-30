import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { TabSwitcher } from './TabSwitcher';

describe('TabSwitcher', () => {
  it('FCT-US001-008 — two tabs, Documents active by default', () => {
    const onTabChange = jest.fn();
    render(<TabSwitcher activeTab="documents" onTabChange={onTabChange} />);

    const tabs = screen.getAllByRole('tab');
    expect(tabs).toHaveLength(2);

    const documentsTab = screen.getByRole('tab', { name: /Documents/i });
    const userStoriesTab = screen.getByRole('tab', { name: /User Stories/i });

    expect(documentsTab).toHaveAttribute('aria-selected', 'true');
    expect(userStoriesTab).toHaveAttribute('aria-selected', 'false');

    // Tab container must have role="tablist"
    expect(screen.getByRole('tablist')).toBeInTheDocument();
  });

  it('FCT-US001-009 — clicking "User Stories" calls onTabChange with "user-stories"', () => {
    const onTabChange = jest.fn();
    render(<TabSwitcher activeTab="documents" onTabChange={onTabChange} />);

    fireEvent.click(screen.getByRole('tab', { name: /User Stories/i }));

    expect(onTabChange).toHaveBeenCalledWith('user-stories');
  });

  it('FCT-US001-010 — clicking "Documents" from User Stories state calls onTabChange with "documents"', () => {
    const onTabChange = jest.fn();
    render(<TabSwitcher activeTab="user-stories" onTabChange={onTabChange} />);

    const documentsTab = screen.getByRole('tab', { name: /Documents/i });
    const userStoriesTab = screen.getByRole('tab', { name: /User Stories/i });

    expect(userStoriesTab).toHaveAttribute('aria-selected', 'true');
    expect(documentsTab).toHaveAttribute('aria-selected', 'false');

    fireEvent.click(documentsTab);

    expect(onTabChange).toHaveBeenCalledWith('documents');
  });

  it('tab panels have correct role and aria-labelledby wiring', () => {
    const onTabChange = jest.fn();
    render(<TabSwitcher activeTab="documents" onTabChange={onTabChange} />);

    // Each tab must have an id for aria-labelledby
    const documentsTab = screen.getByRole('tab', { name: /Documents/i });
    expect(documentsTab).toHaveAttribute('id', 'tab-documents');

    const userStoriesTab = screen.getByRole('tab', { name: /User Stories/i });
    expect(userStoriesTab).toHaveAttribute('id', 'tab-user-stories');
  });
});
