import React from 'react';
import { render, screen } from '@testing-library/react';
import { ProjectHeader } from './ProjectHeader';
import { Project } from '../../lib/api/types';

const mockProject: Project = {
  id: 'p1',
  name: 'Project Alpha',
  description: 'Some description',
  path: '/tmp/project-alpha',
  createdAt: '2026-05-20T10:00:00Z',
  updatedAt: '2026-05-20T10:00:00Z',
};

describe('ProjectHeader', () => {
  it('FCT-US001-006 — empty description shows "No description" placeholder', () => {
    const emptyDescProject: Project = { ...mockProject, description: '' };
    render(<ProjectHeader project={emptyDescProject} />);

    expect(screen.getByText(/No description/i)).toBeInTheDocument();
    // Must not render empty string as an empty element
    const paragraphs = screen.queryAllByText('');
    paragraphs.forEach((el) => {
      // If an element has empty text content, it should not be a description element
      expect(el.textContent).not.toBe('');
    });
  });

  it('renders the project name in an h1', () => {
    render(<ProjectHeader project={mockProject} />);
    expect(screen.getByRole('heading', { level: 1, name: /Project Alpha/i })).toBeInTheDocument();
  });

  it('renders description when non-empty', () => {
    render(<ProjectHeader project={mockProject} />);
    expect(screen.getByText('Some description')).toBeInTheDocument();
  });

  it('renders "Back to dashboard" link to "/"', () => {
    render(<ProjectHeader project={mockProject} />);
    const link = screen.getByRole('link', { name: /Back to dashboard/i });
    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute('href', '/');
  });

  it('renders loading skeleton when isLoading is true', () => {
    render(<ProjectHeader isLoading />);
    expect(screen.getByTestId('project-header-skeleton')).toBeInTheDocument();
  });

  it('renders "Project not found" message when isNotFound is true', () => {
    render(<ProjectHeader isNotFound />);
    expect(screen.getByText(/Project not found/i)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Back to dashboard/i })).toBeInTheDocument();
  });

  it('renders "Failed to load project" message when hasError is true', () => {
    render(<ProjectHeader hasError />);
    expect(screen.getByText(/Failed to load project/i)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Back to dashboard/i })).toBeInTheDocument();
  });
});
