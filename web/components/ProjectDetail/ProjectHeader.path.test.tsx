/**
 * Tests for ProjectHeader — path display
 * FCT-047-001 — renders path as read-only text (not editable)
 * FCT-047-002 — renders exact path string from project object
 */
import React from 'react';
import { render, screen } from '@testing-library/react';
import { ProjectHeader } from './ProjectHeader';
import { Project } from '../../lib/api/types';

const baseProject: Project = {
  id: '11111111-1111-1111-1111-111111111111',
  name: 'agents-board',
  description: '',
  path: '/Users/me/workspace/agents-board',
  createdAt: '2026-06-01T09:00:00Z',
  updatedAt: '2026-06-01T09:00:00Z',
};

describe('ProjectHeader — path display (FCT-047)', () => {
  // FCT-047-001 — path visible, read-only
  it('FCT-047-001: renders path text and is not in an input field', () => {
    render(<ProjectHeader project={baseProject} />);

    // Path text is visible
    expect(screen.getByText('/Users/me/workspace/agents-board')).toBeInTheDocument();

    // Must NOT be in an input
    const inputs = document.querySelectorAll('input');
    inputs.forEach((input) => {
      expect(input.value).not.toBe('/Users/me/workspace/agents-board');
    });
  });

  // FCT-047-002 — various path values rendered exactly
  it('FCT-047-002: renders exact path string from project object', () => {
    const paths = [
      '/Users/me/workspace/agents-board',
      '/Users/John Doe/my projects/cool project',
      '/tmp/пproject/テスト',
      '/very/long/path/that/exceeds/eighty/characters/and/then/some/more/path/segments/here',
    ];

    paths.forEach((path) => {
      const { unmount } = render(<ProjectHeader project={{ ...baseProject, path }} />);
      expect(screen.getByText(path)).toBeInTheDocument();
      unmount();
    });
  });

  it('path element is not a textarea or contentEditable element', () => {
    render(<ProjectHeader project={baseProject} />);

    const pathEl = screen.getByText('/Users/me/workspace/agents-board');

    // Not a textarea
    expect(pathEl.tagName.toLowerCase()).not.toBe('textarea');
    // Not editable
    expect(pathEl).not.toHaveAttribute('contenteditable', 'true');
  });
});
