import React from 'react';
import Head from 'next/head';
import { useRouter } from 'next/router';
import { ProjectHeader } from '../../components/ProjectDetail/ProjectHeader';
import { TabSwitcher, TabId } from '../../components/ProjectDetail/TabSwitcher';
import { UserStoriesTab } from '../../components/ProjectDetail/UserStoriesTab';
import { DocumentsTab } from '../../components/ProjectDetail/DocumentsTab';
import { RequirementSelector } from '../../components/ProjectDetail/RequirementSelector';
import { useProject } from '../../hooks/useProject';
import { useProjectRequirements } from '../../hooks/useProjectRequirements';

/**
 * ProjectDetailPage — CSR-only route at /projects/[id].
 *
 * Reads `id`, `tab`, and `requirement` from useRouter().query.
 * - `tab` defaults to 'documents' when absent or unrecognized (URL is source of truth per D-003).
 * - `requirement` drives which requirement is selected in the RequirementSelector.
 *   When absent and requirements are loaded, auto-selects the first requirement.
 *
 * Renders:
 * - ProjectHeader (with loading skeleton / 404 / error states)
 * - RequirementSelector (fetches the project's requirements)
 * - TabSwitcher + active tab body scoped to the selected requirementId
 *
 * NO getServerSideProps / getStaticProps / getInitialProps — CSR-only.
 */
export default function ProjectDetailPage() {
  const router = useRouter();
  const { id, tab, requirement } = router.query;

  // Normalise id to string | undefined (query values can be string | string[] | undefined)
  const projectId = typeof id === 'string' ? id : undefined;

  // Normalise tab — default to 'documents' when absent or unrecognized
  const rawTab = typeof tab === 'string' ? tab : '';
  const activeTab: TabId =
    rawTab === 'documents' || rawTab === 'user-stories'
      ? (rawTab as TabId)
      : 'documents';

  // Normalise requirement query param
  const requirementParam = typeof requirement === 'string' ? requirement : undefined;

  const { data: project, isLoading, isNotFound, error } = useProject(projectId);

  // Load requirements for this project (only when project is loaded)
  const { requirements, loading: reqLoading } = useProjectRequirements(
    project ? projectId : undefined
  );

  // Determine selected requirementId:
  // - URL param if present and valid
  // - First requirement when no param (auto-select for migrated projects)
  // - undefined when requirements are still loading or empty
  const selectedRequirementId: string | undefined = (() => {
    if (reqLoading) return undefined;
    if (requirementParam) return requirementParam;
    if (requirements.length > 0) return requirements[0].id;
    return undefined;
  })();

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

  const handleRequirementSelect = (reqId: string) => {
    void router.replace(
      {
        pathname: router.pathname,
        query: { ...router.query, requirement: reqId },
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
            {/* Requirements selector — scopes tab content */}
            <div className="mb-4">
              <RequirementSelector
                projectId={project.id}
                selectedRequirementId={selectedRequirementId}
                onSelect={handleRequirementSelect}
              />
            </div>

            {/* Tab switcher and tab body */}
            {selectedRequirementId !== undefined ? (
              <>
                <TabSwitcher activeTab={activeTab} onTabChange={handleTabChange} />

                <div className="mt-4">
                  {activeTab === 'user-stories' ? (
                    <UserStoriesTab projectId={project.id} requirementId={selectedRequirementId} />
                  ) : (
                    <DocumentsTab projectId={project.id} requirementId={selectedRequirementId} />
                  )}
                </div>
              </>
            ) : (
              /* No requirement selected: show placeholder */
              <div className="mt-4">
                <TabSwitcher activeTab={activeTab} onTabChange={handleTabChange} />
                <div className="mt-4 p-8 text-center text-gray-500 border-2 border-dashed rounded">
                  <p>Select a requirement to view user stories and documents.</p>
                </div>
              </div>
            )}
          </>
        )}
      </main>
    </div>
  );
}
