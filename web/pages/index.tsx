import React from 'react';
import Head from 'next/head';
import { useProjects } from '../hooks/useProjects';
import { ProjectList } from '../components/Dashboard/ProjectList';

/**
 * Home component
 * Main route for the dashboard showing the minimal beautiful list of projects.
 */
export default function Home() {
  const { data: projects, isLoading, isError, error } = useProjects();

  return (
    <div className="min-h-screen bg-gray-50 font-sans text-gray-900">
      <Head>
        <title>Dashboard</title>
        <meta name="description" content="A minimal beautiful dashboard" />
      </Head>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <header className="mb-10">
          <h1 className="text-3xl font-bold tracking-tight text-gray-900">Projects</h1>
          <p className="mt-2 text-sm text-gray-600">View and manage your available projects.</p>
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
      </main>
    </div>
  );
}
