# US047 — Requirement-level navigation on the project detail page

**Requirement:** REQ008 — Requirement entity + project local-path linking
**Status:** draft

## Story
As a user viewing a project, I want to see the project's Requirements and drill into one to view its User Stories and Documents, so that the dashboard reflects the corrected `Project → Requirement → User Story / Document` hierarchy.

## Acceptance criteria
- **Scenario: Project shows its requirements**
  - Given a project that has one or more requirements
  - When I open the project detail page
  - Then I see a list of the project's requirements (name, plus its status; description optional)
- **Scenario: Drill into a requirement**
  - Given the project detail page shows requirements
  - When I select a requirement
  - Then I see that requirement's User Stories and Documents (the existing tabs/lists now scoped to the selected requirement)
- **Scenario: Linked path visible**
  - Given a project linked to a local path (every project has a non-blank `path`)
  - When I view the project detail page
  - Then the linked local path is displayed (read-only) in the project header
- **Scenario: Empty requirements**
  - Given a project with no requirements
  - When I open the project detail page
  - Then I see an empty state for requirements (e.g. "No requirements yet")
- **Scenario: Loading state**
  - Given the requirements are being fetched
  - When the request is in flight
  - Then a loading indicator/skeleton is shown for the requirements area
- **Scenario: Error state**
  - Given the requirements request fails
  - When I open the project detail page
  - Then an error message is shown for the requirements area (the rest of the page still renders)
- **Scenario: Deep-link to a requirement**
  - Given a URL that selects a project and a requirement
  - When I load that URL
  - Then the matching requirement's stories/documents are shown (URL is the source of truth, consistent with existing tab behaviour)

## UI / UX flow expectations
- **Entry points:** Clicking a project from the projects list opens `/projects/[id]`; the requirement selection is reachable from there (and deep-linkable, e.g. via a query param like the existing `tab`).
- **Happy-path flow:** Open project detail → see project header (now including the linked local path, always present) + a Requirements list/selector → select a requirement → the User Stories and Documents tabs/lists render scoped to that requirement.
- **Empty / loading / error states:**
  - Empty: "No requirements yet" in the requirements area.
  - Loading: skeleton/spinner for requirements while fetching.
  - Error: inline error in the requirements area; project header still renders.
- **Validation rules visible to the user:** none (read/navigate only).
- **Out of UI scope:** creating/editing requirements from this page, styling polish, animations.

## Out of scope
- Creating/editing/deleting requirements from the web (requirement creation is MCP-only per US045; the web is read-only for requirements — only the HTTP list endpoint exists, and no create-requirement UI is part of this story).
- Editing the project's path.
- Changing the internal structure of the existing User Stories / Documents tab bodies beyond re-scoping them to a selected requirement.

## Dependencies
- US045 (`GET .../requirements` listing + requirement-scoped story/document listing contract).
- Existing project detail page (`web/pages/projects/[id].tsx`) and its Documents / User Stories tabs.

## Notes for the team
- The existing detail page already drives `tab` from the URL (`shallow` routing); follow the same source-of-truth pattern for the selected requirement (e.g. a `requirement` query param).
- The existing UserStoriesTab / DocumentsTab fetch by `projectId`; they will need to fetch by `requirementId` once the model is re-parented — confirm the contract with the architect's API shapes (US045). Both User Stories and Documents are now requirement-scoped (confirmed D-1).
- Migrated projects each have one "Default" requirement holding all their pre-existing stories and documents — the navigation must render that gracefully (a single requirement in the list).

## Sign-off log
(po-ba appends here on each sign-off pass)
