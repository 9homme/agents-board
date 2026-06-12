import React from 'react';
import Link from 'next/link';
import { Project } from '../../lib/api/types';

/** Props for ProjectHeader when a project has successfully loaded. */
interface ProjectHeaderLoadedProps {
  project: Project;
  isLoading?: false;
  isNotFound?: false;
  hasError?: false;
}

/** Props for ProjectHeader when the fetch is still in progress. */
interface ProjectHeaderLoadingProps {
  isLoading: true;
  project?: undefined;
  isNotFound?: false;
  hasError?: false;
}

/** Props for ProjectHeader when the project was not found (404). */
interface ProjectHeaderNotFoundProps {
  isNotFound: true;
  project?: undefined;
  isLoading?: false;
  hasError?: false;
}

/** Props for ProjectHeader when the fetch failed with a server/network error. */
interface ProjectHeaderErrorProps {
  hasError: true;
  project?: undefined;
  isLoading?: false;
  isNotFound?: false;
}

type ProjectHeaderProps =
  | ProjectHeaderLoadedProps
  | ProjectHeaderLoadingProps
  | ProjectHeaderNotFoundProps
  | ProjectHeaderErrorProps;

/**
 * ProjectHeader component.
 *
 * Renders the project header section which includes:
 * - A "Back to dashboard" link (always present).
 * - The project name as an `<h1>` heading.
 * - The project description (or "No description" placeholder when empty).
 *
 * Also handles three alternative states:
 * - **Loading:** displays a skeleton block.
 * - **Not found (404):** displays a friendly "Project not found" message.
 * - **Error (5xx/network):** displays a friendly "Failed to load project" message.
 */
export const ProjectHeader: React.FC<ProjectHeaderProps> = (props) => {
  if (props.isLoading) {
    return (
      <header className="mb-6">
        <Link href="/" className="text-sm text-gray-500 hover:text-gray-700 mb-4 inline-block">
          &larr; Back to dashboard
        </Link>
        <div data-testid="project-header-skeleton" aria-busy="true" aria-label="Loading project">
          <div className="h-8 bg-gray-200 rounded w-64 mb-2 animate-pulse" />
          <div className="h-4 bg-gray-200 rounded w-96 animate-pulse" />
        </div>
      </header>
    );
  }

  if (props.isNotFound) {
    return (
      <header className="mb-6">
        <Link href="/" className="text-sm text-gray-500 hover:text-gray-700 mb-4 inline-block">
          &larr; Back to dashboard
        </Link>
        <h1 className="text-2xl font-bold text-gray-900">Project not found</h1>
        <p className="text-gray-600">The project you are looking for does not exist.</p>
      </header>
    );
  }

  if (props.hasError) {
    return (
      <header className="mb-6">
        <Link href="/" className="text-sm text-gray-500 hover:text-gray-700 mb-4 inline-block">
          &larr; Back to dashboard
        </Link>
        <h1 className="text-2xl font-bold text-gray-900">Failed to load project</h1>
        <p className="text-gray-600">Something went wrong. Please try again later.</p>
      </header>
    );
  }

  const { project } = props;

  return (
    <header className="mb-6">
      <Link href="/" className="text-sm text-gray-500 hover:text-gray-700 mb-4 inline-block">
        &larr; Back to dashboard
      </Link>
      <h1 className="text-3xl font-bold text-gray-900">{project.name}</h1>
      {project.description ? (
        <p className="text-gray-600 mt-2">{project.description}</p>
      ) : (
        <p className="text-gray-400 mt-2 italic">No description</p>
      )}
      {/* Read-only linked path (US047 — always present per §2) */}
      <p className="text-sm text-gray-500 mt-1 font-mono truncate" title={project.path}>
        {project.path}
      </p>
    </header>
  );
};
