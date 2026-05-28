import React from 'react';
import Head from 'next/head';
import { useRouter } from 'next/router';
import { ProjectHeader } from '../../components/ProjectDetail/ProjectHeader';
import { TabSwitcher, TabId } from '../../components/ProjectDetail/TabSwitcher';
import { UserStoriesTab } from '../../components/ProjectDetail/UserStoriesTab';
import { DocumentsTab } from '../../components/ProjectDetail/DocumentsTab';
import { useProject } from '../../hooks/useProject';

/**
 * ProjectDetailPage — CSR-only route at /projects/[id].
 *
 * Reads `id` and `tab` from useRouter().query. The `tab` defaults to
 * 'documents' when absent or unrecognized (URL is source of truth per D-003).
 *
 * Renders:
 * - ProjectHeader (with loading skeleton / 404 / error states)
 * - TabSwitcher + active tab body (only when project is found)
 *
 * NO getServerSideProps / getStaticProps / getInitialProps — CSR-only.
 */
export default function ProjectDetailPage() {
  const router = useRouter();
  const { id, tab } = router.query;

  // Normalise id to string | undefined (query values can be string | string[] | undefined)
  const projectId = typeof id === 'string' ? id : undefined;

  // Normalise tab — default to 'documents' when absent or unrecognized
  const rawTab = typeof tab === 'string' ? tab : '';
  const activeTab: TabId =
    rawTab === 'documents' || rawTab === 'user-stories'
      ? (rawTab as TabId)
      : 'documents';

  const { data: project, isLoading, isNotFound, error } = useProject(projectId);

  const hasError = !isLoading && !isNotFound && error !== null;

  const handleTabChange = (nextTab: TabId) => {
    void router.replace(
      {
        pathname: router.pathname,
        query: { ...router.query, tab: nextTab },
      },
      undefined,
      { shallow: true }
    );
  };

  const title = project?.name ?? 'Project';

  return (
    <div className="min-h-screen bg-gray-50 font-sans text-gray-900">
      <Head>
        <title>{title}</title>
      </Head>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* Project header — handles all four states: loading, not-found, error, loaded */}
        {isLoading ? (
          <ProjectHeader isLoading />
        ) : isNotFound ? (
          <ProjectHeader isNotFound />
        ) : hasError ? (
          <ProjectHeader hasError />
        ) : project ? (
          <ProjectHeader project={project} />
        ) : null}

        {/* Tab switcher and tab body — only visible when project loaded successfully */}
        {!isLoading && !isNotFound && !hasError && project && (
          <>
            <TabSwitcher activeTab={activeTab} onTabChange={handleTabChange} />

            <div className="mt-4">
              {activeTab === 'user-stories' ? (
                <UserStoriesTab />
              ) : (
                <DocumentsTab projectId={project.id} />
              )}
            </div>
          </>
        )}
      </main>
    </div>
  );
}
