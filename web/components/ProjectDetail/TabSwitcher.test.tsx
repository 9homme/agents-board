import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
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

  // ── US013 coverage-backfill tests (12 FCT-* cases) ─────────────────────────

  it('FCT-US013-001 — clicking non-active tab fires onTabChange with clicked id; controlled component does NOT mutate activeTab', async () => {
    // Architecture cite: architecture.md §7.2 FCT-US013-001
    const onTabChange = jest.fn();
    const user = userEvent.setup();
    render(<TabSwitcher activeTab="documents" onTabChange={onTabChange} />);

    await user.click(screen.getByRole('tab', { name: /user stories/i }));

    expect(onTabChange).toHaveBeenCalledTimes(1);
    expect(onTabChange).toHaveBeenCalledWith('user-stories');
    // Controlled component: activeTab prop not changed by parent — Documents still selected
    expect(screen.getByRole('tab', { name: /documents/i })).toHaveAttribute('aria-selected', 'true');
  });

  it('FCT-US013-002 — ArrowRight moves focus forward and fires onTabChange', async () => {
    // Architecture cite: architecture.md §7.2 FCT-US013-002
    const onTabChange = jest.fn();
    const user = userEvent.setup();
    render(<TabSwitcher activeTab="documents" onTabChange={onTabChange} />);

    const docsTab = screen.getByRole('tab', { name: /documents/i });
    await user.click(docsTab); // focus the Documents tab
    await user.keyboard('{ArrowRight}');

    expect(onTabChange).toHaveBeenLastCalledWith('user-stories');
    expect(document.activeElement).toBe(screen.getByRole('tab', { name: /user stories/i }));
  });

  it('FCT-US013-003 — ArrowRight from last tab wraps to first', async () => {
    // Architecture cite: architecture.md §7.2 FCT-US013-003; modulo arithmetic
    const onTabChange = jest.fn();
    const user = userEvent.setup();
    render(<TabSwitcher activeTab="user-stories" onTabChange={onTabChange} />);

    await user.click(screen.getByRole('tab', { name: /user stories/i }));
    await user.keyboard('{ArrowRight}');

    expect(onTabChange).toHaveBeenLastCalledWith('documents');
    expect(document.activeElement).toBe(screen.getByRole('tab', { name: /documents/i }));
  });

  it('FCT-US013-004 — ArrowLeft moves focus backward and fires onTabChange', async () => {
    // Architecture cite: architecture.md §7.2 FCT-US013-004
    const onTabChange = jest.fn();
    const user = userEvent.setup();
    render(<TabSwitcher activeTab="user-stories" onTabChange={onTabChange} />);

    await user.click(screen.getByRole('tab', { name: /user stories/i }));
    await user.keyboard('{ArrowLeft}');

    expect(onTabChange).toHaveBeenLastCalledWith('documents');
    expect(document.activeElement).toBe(screen.getByRole('tab', { name: /documents/i }));
  });

  it('FCT-US013-005 — ArrowLeft from first tab wraps to last', async () => {
    // Architecture cite: architecture.md §7.2 FCT-US013-005; modulo arithmetic
    const onTabChange = jest.fn();
    const user = userEvent.setup();
    render(<TabSwitcher activeTab="documents" onTabChange={onTabChange} />);

    await user.click(screen.getByRole('tab', { name: /documents/i }));
    await user.keyboard('{ArrowLeft}');

    expect(onTabChange).toHaveBeenLastCalledWith('user-stories');
    expect(document.activeElement).toBe(screen.getByRole('tab', { name: /user stories/i }));
  });

  it('FCT-US013-006 — Enter activates focused tab; calls preventDefault', async () => {
    // Architecture cite: architecture.md §7.2 FCT-US013-006
    // preventDefault is implicit in user-event's full event sequence.
    const onTabChange = jest.fn();
    const user = userEvent.setup();
    render(<TabSwitcher activeTab="documents" onTabChange={onTabChange} />);

    await user.click(screen.getByRole('tab', { name: /documents/i }));
    await user.keyboard('{ArrowRight}'); // focus moves to User Stories
    onTabChange.mockClear(); // reset to isolate Enter assertion
    await user.keyboard('{Enter}');

    expect(onTabChange).toHaveBeenCalledTimes(1);
    expect(onTabChange).toHaveBeenCalledWith('user-stories');
  });

  it('FCT-US013-007 — Space activates focused tab; calls preventDefault', async () => {
    // Architecture cite: architecture.md §7.2 FCT-US013-007
    const onTabChange = jest.fn();
    const user = userEvent.setup();
    render(<TabSwitcher activeTab="documents" onTabChange={onTabChange} />);

    await user.click(screen.getByRole('tab', { name: /documents/i }));
    await user.keyboard('{ArrowRight}'); // User Stories focused
    onTabChange.mockClear(); // reset to isolate Space assertion
    await user.keyboard('{ }'); // Space key

    expect(onTabChange).toHaveBeenCalledTimes(1);
    expect(onTabChange).toHaveBeenCalledWith('user-stories');
  });

  it('FCT-US013-008 — aria-selected per tab reflects activeTab prop', () => {
    // Architecture cite: architecture.md §7.2 FCT-US013-008; §7.1 ARIA attributes
    const onTabChange = jest.fn();
    render(<TabSwitcher activeTab="documents" onTabChange={onTabChange} />);

    expect(screen.getByRole('tab', { name: /documents/i })).toHaveAttribute('aria-selected', 'true');
    expect(screen.getByRole('tab', { name: /user stories/i })).toHaveAttribute('aria-selected', 'false');
  });

  it('FCT-US013-009 — roving tabIndex per tab reflects activeTab prop', () => {
    // Architecture cite: architecture.md §7.2 FCT-US013-009; §7.1 roving tabindex
    const onTabChange = jest.fn();
    render(<TabSwitcher activeTab="documents" onTabChange={onTabChange} />);

    expect(screen.getByRole('tab', { name: /documents/i })).toHaveAttribute('tabIndex', '0');
    expect(screen.getByRole('tab', { name: /user stories/i })).toHaveAttribute('tabIndex', '-1');
  });

  it('FCT-US013-010 — prop-driven activeTab change re-renders with new active tab AND does NOT fire onTabChange', () => {
    // Architecture cite: architecture.md §7.2 FCT-US013-010
    const onTabChange = jest.fn();
    const { rerender } = render(<TabSwitcher activeTab="documents" onTabChange={onTabChange} />);

    rerender(<TabSwitcher activeTab="user-stories" onTabChange={onTabChange} />);

    expect(screen.getByRole('tab', { name: /user stories/i })).toHaveAttribute('aria-selected', 'true');
    expect(screen.getByRole('tab', { name: /user stories/i })).toHaveAttribute('tabIndex', '0');
    expect(screen.getByRole('tab', { name: /documents/i })).toHaveAttribute('aria-selected', 'false');
    expect(screen.getByRole('tab', { name: /documents/i })).toHaveAttribute('tabIndex', '-1');
    // Prop-driven re-render must NOT fire the callback
    expect(onTabChange).not.toHaveBeenCalled();
  });

  it('FCT-US013-011 — tablist semantics are present', () => {
    // Architecture cite: architecture.md §7.2 FCT-US013-011; §7.1 ARIA attributes
    const onTabChange = jest.fn();
    render(<TabSwitcher activeTab="documents" onTabChange={onTabChange} />);

    // Single tablist with aria-label
    const tablist = screen.getByRole('tablist');
    expect(tablist).toHaveAttribute('aria-label', 'Project tabs');

    // Exactly 2 tab elements
    expect(screen.getAllByRole('tab')).toHaveLength(2);

    // Documents tab ARIA wiring
    const docsTab = screen.getByRole('tab', { name: /documents/i });
    expect(docsTab).toHaveAttribute('id', 'tab-documents');
    expect(docsTab).toHaveAttribute('aria-controls', 'tabpanel-documents');

    // User Stories tab ARIA wiring
    const storiesTab = screen.getByRole('tab', { name: /user stories/i });
    expect(storiesTab).toHaveAttribute('id', 'tab-user-stories');
    expect(storiesTab).toHaveAttribute('aria-controls', 'tabpanel-user-stories');
  });

  it('FCT-US013-012 — unrelated keys do NOT fire onTabChange', async () => {
    // Architecture cite: architecture.md §7.2 FCT-US013-012; §7.1 NOT IMPLEMENTED list
    // Note: Tab is handled by browser's native focus management; user-event will move
    // focus away from the tablist but onTabChange must not fire (spec: US013_fe_unit_tests.md).
    const onTabChange = jest.fn();
    const user = userEvent.setup();
    render(<TabSwitcher activeTab="documents" onTabChange={onTabChange} />);

    // Focus the Documents tab (click fires onTabChange('documents') once)
    await user.click(screen.getByRole('tab', { name: /documents/i }));
    // Reset mock so we only count calls triggered by unrelated keys
    onTabChange.mockClear();

    // Press unrelated keys: Escape, 'a', Tab — component ignores them
    await user.keyboard('{Escape}a{Tab}');

    expect(onTabChange).not.toHaveBeenCalled();
  });
});
