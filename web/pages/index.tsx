import React, { useRef, useState } from 'react';
import Head from 'next/head';
import { useProjects } from '../hooks/useProjects';
import { ProjectList } from '../components/Dashboard/ProjectList';
import { AddProjectDialog } from '../components/Dashboard/AddProjectDialog';

/**
 * Home component
 * Main route for the dashboard showing the minimal beautiful list of projects.
 * Includes an "Add Project" button that opens a dialog for creating a new project.
 */
export default function Home() {
  const { data: projects, isLoading, isError, error, refetch } = useProjects();
  const [dialogOpen, setDialogOpen] = useState(false);
  const addProjectBtnRef = useRef<HTMLButtonElement>(null);

  function handleSuccess() {
    refetch();
  }

  return (
    <div className="min-h-screen bg-gray-50 font-sans text-gray-900">
      <Head>
        <title>Dashboard</title>
        <meta name="description" content="A minimal beautiful dashboard" />
      </Head>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <header className="mb-10 flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-gray-900">Projects</h1>
            <p className="mt-2 text-sm text-gray-600">View and manage your available projects.</p>
          </div>
          <button
            ref={addProjectBtnRef}
            onClick={() => setDialogOpen(true)}
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            Add Project
          </button>
        </header>

        <section aria-label="Projects">
          {isLoading ? (
            <div className="flex justify-center py-20">
              <div className="animate-pulse text-lg text-gray-500 font-medium">Loading projects...</div>
            </div>
          ) : isError ? (
            <div className="bg-red-50 border border-red-200 rounded-lg p-6 text-center text-red-600">
              <h2 className="text-lg font-semibold mb-2">Error</h2>
              <p>Failed to load projects: {error?.message}</p>
            </div>
          ) : (
            <ProjectList projects={projects} />
          )}
        </section>

        <AddProjectDialog
          open={dialogOpen}
          onClose={() => setDialogOpen(false)}
          onSuccess={handleSuccess}
          triggerRef={addProjectBtnRef}
        />
      </main>
    </div>
  );
}
