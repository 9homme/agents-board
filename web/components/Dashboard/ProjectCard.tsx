/**
 * ProjectCard component
 * Displays a single project in a card UI with its name, description, and dates.
 */
import React from 'react';
import { Project } from '../../lib/api/types';

interface ProjectCardProps {
  project: Project;
}

export const ProjectCard: React.FC<ProjectCardProps> = ({ project }) => {
  return (
    <article className="border border-gray-200 rounded-lg p-6 shadow-sm hover:shadow-md transition-shadow bg-white flex flex-col h-full">
      <h3 className="text-xl font-semibold mb-2 text-gray-900">{project.name}</h3>
      <p className="text-gray-600 mb-4 flex-grow">{project.description}</p>
      <div className="text-xs text-gray-500 mt-auto flex flex-col gap-1 border-t pt-4">
        <span>Created: {new Date(project.createdAt).toLocaleDateString()}</span>
        <span>Updated: {new Date(project.updatedAt).toLocaleDateString()}</span>
      </div>
    </article>
  );
};
