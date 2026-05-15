/**
 * ProjectList component
 * Displays a grid of project cards.
 */
import React from 'react';
import { Project } from '../../lib/api/types';
import { ProjectCard } from './ProjectCard';

interface ProjectListProps {
  projects: Project[];
}

export const ProjectList: React.FC<ProjectListProps> = ({ projects }) => {
  if (!projects || projects.length === 0) {
    return (
      <div className="text-center py-12 bg-gray-50 rounded-lg border border-gray-100">
        <p className="text-gray-500 text-lg">No projects found.</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {projects.map((project) => (
        <ProjectCard key={project.id} project={project} />
      ))}
    </div>
  );
};
