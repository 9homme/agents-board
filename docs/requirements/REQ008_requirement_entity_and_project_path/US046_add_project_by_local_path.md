# US046 — Add Project from web by linking a local path

**Requirement:** REQ008 — Requirement entity + project local-path linking
**Status:** draft

## Story
As a user of the dashboard, I want to add a new project by typing the full local path in a plain text field and a name (auto-filled from the path basename, editable), so that the project is linked to a multi-agents project directory on my disk.

## Acceptance criteria
- **Scenario: Open the add-project form**
  - Given I am on the projects list/dashboard
  - When I click "Add Project"
  - Then a form (page or dialog) appears with a plain text path field and a name field, both marked required
- **Scenario: Type the full path manually**
  - Given the add-project form is open
  - When I type into the path field
  - Then the field is a plain `<input type="text">` — I enter the full directory path by hand, with no suggestions dropdown
- **Scenario: Name auto-fills from path basename**
  - Given the name field has not been manually edited
  - When the path field changes to a non-empty value
  - Then the name field is auto-filled with the basename of that path, and remains editable; if I edit the name manually, subsequent path changes do not overwrite my edit
- **Scenario: Successful create**
  - Given a valid name and a valid path are entered
  - When I submit
  - Then the project is created via the backend, the form closes, and the new project appears in the projects list (or I am navigated to its detail page)
- **Scenario: Client-side validation — empty required fields**
  - Given the add-project form is open
  - When the name or path is blank
  - Then the submit button is disabled and/or an inline validation message is shown, and no create request is sent
- **Scenario: Server validation error surfaced**
  - Given a path the server rejects (does not exist / not a directory → 400, or already linked → 409)
  - When I submit
  - Then the server's error is shown inline (e.g. "Path does not exist or is not a directory", "Path already linked to another project"), the form stays open with my input preserved, and no project is added to the list
- **Scenario: Submission in-flight**
  - Given I have submitted the form
  - When the request is in flight
  - Then the submit control shows a loading/disabled state and double-submit is prevented

## UI / UX flow expectations
- **Entry points:** An "Add Project" button on the projects list / dashboard page.
- **Happy-path flow:** Click "Add Project" → form (dialog or `/projects/new` page) opens → user types the full local path into a plain text field → name auto-fills from the basename (editable) → user submits → on success the form closes and the new project shows in the list (or navigates to the project detail page).
- **Empty / loading / error states:**
  - Empty: name + path fields blank, submit disabled until both are non-blank.
  - Submit loading: submit shows spinner/disabled while the create request is in flight.
  - Error: server validation errors (path blank / missing on disk / not a directory / duplicate) rendered inline near the path field; input is preserved.
- **Validation rules visible to the user:** name required (non-blank); path required (non-blank); submit disabled until both are non-blank. Existence/is-directory/uniqueness are server-authoritative and surfaced as inline errors after submit.
- **Out of UI scope:** styling polish, a native OS directory-picker dialog (user types the path manually), filesystem autocomplete/suggestions (removed this REQ), editing path after creation.

## Out of scope
- Filesystem autocomplete / directory suggestions dropdown (removed this REQ — plain text path entry only).
- Native filesystem browse dialog.
- Editing or deleting an existing project's path.
- Displaying anything about the linked directory's contents (excluded live-control / file-sync feature).

## Dependencies
- US045 (project-create API accepting **required** `path` + its 400/409 validation contract).

## Notes for the team
- All backend calls go through `web/lib/api/` (CSR-only). Use the project-create contract from `architecture.md` (consumed via the task's Architecture extract). There is **no** `web/lib/api/fs.ts` and **no** suggestions endpoint this REQ.
- The path field is a plain `<input type="text">` — no `PathAutocomplete` component, no `usePathSuggestions` hook, no debounced suggestion requests.
- Existence/is-directory/uniqueness checks are server-authoritative; the FE only enforces non-blank client-side and renders server 400/409 errors inline (400 `VALIDATION_ERROR` vs 409 `DUPLICATE_PATH`).
- Name auto-fill must be "sticky-off" once the user manually edits the name, so path changes don't clobber a deliberate name.

## Sign-off log
(po-ba appends here on each sign-off pass)
