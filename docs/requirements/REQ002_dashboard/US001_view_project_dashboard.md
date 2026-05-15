# US001 — View project dashboard

**Requirement:** REQ002 — dashboard
**Status:** in_signoff

## Story
As a user, I want to see a dashboard listing all projects, so that I can quickly view the available projects in a minimal, beautiful card UI.

## Acceptance criteria
- **Scenario: Successfully load project list**
  - Given there are existing projects in the system
  - When I navigate to the dashboard page
  - Then I should see a list of projects displayed as cards
  - And each card should display at least the project's name and description

- **Scenario: Empty state**
  - Given there are no projects in the system
  - When I navigate to the dashboard page
  - Then I should see a clear "no projects" message

- **Scenario: Loading state**
  - Given the dashboard is fetching projects
  - When I access the page
  - Then I should see a loading indicator (e.g., spinner or skeleton cards)

- **Scenario: Error state**
  - Given the system fails to retrieve the projects
  - When I access the page
  - Then I should see a friendly error message indicating the failure

## UI / UX flow expectations
- **Entry points:** The root URL (`/`) or a dedicated `/dashboard` route.
- **Happy-path flow:** User accesses the page -> sees a loading state -> sees a grid or list of minimalist project cards.
- **Empty / loading / error states:** 
  - Loading: Visual indicator that data is being fetched.
  - Empty: "No projects found" message.
  - Error: "Failed to load projects" text.
- **Validation rules visible to the user:** None.
- **Out of UI scope:** Clicking cards to view details, pagination, sorting, filtering, and any project management actions (CRUD).

## Out of scope
- Clicking on a project card to view its details.
- Editing, creating, or deleting projects.
- Search, filter, or sort functionality.
- Pagination.

## Dependencies
- Backend needs to provide a list of projects to the web frontend. (Note: REQ001 exposed projects via an MCP interface. System Architect to determine if the frontend queries MCP directly or if a standard REST API is required).

## Notes for the team
- The visual design should be "minimal beautiful". Focus on clean layouts, typography, and whitespace.
- The frontend is Next.js Pages Router (CSR-only).

## Sign-off log

### Sign-off pass 1 — 2026-05-15 — verdict: approved
- **Spec review:** All four scenarios (Success, Empty, Loading, Error) are fully covered in FE unit specs (FCT-001 to FCT-004). BE correctly covers Success and Error (UT-001, UT-002, IT-001). E2E (E2E-001) covers the full flow from MCP creation to UI display. Specs accurately map to ACs.
- **Result review:** All listed BE, FE, and E2E test cases passed. No tests were skipped.
- **Routed to:** none
