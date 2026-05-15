import React from 'react';
import { render, screen } from '@testing-library/react';
import { ProjectList } from './ProjectList';
import { Project } from '../../lib/api/types';

describe('ProjectList', () => {
  it('renders a list of projects', () => {
    const mockProjects: Project[] = [
      {
        id: '1',
        name: 'Project 1',
        description: 'Description 1',
        createdAt: '2023-10-25T10:00:00Z',
        updatedAt: '2023-10-25T10:00:00Z',
      },
      {
        id: '2',
        name: 'Project 2',
        description: 'Description 2',
        createdAt: '2023-10-26T10:00:00Z',
        updatedAt: '2023-10-26T10:00:00Z',
      },
    ];

    render(<ProjectList projects={mockProjects} />);
    
    expect(screen.getByText('Project 1')).toBeVisible();
    expect(screen.getByText('Description 1')).toBeVisible();
    expect(screen.getByText('Project 2')).toBeVisible();
    expect(screen.getByText('Description 2')).toBeVisible();
  });

  it('renders an empty state when no projects are provided', () => {
    render(<ProjectList projects={[]} />);
    expect(screen.getByText(/no projects/i)).toBeVisible();
  });
});
