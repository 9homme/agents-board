# US046 — Frontend component test specification

**For FE Dev:** these are the tests you write FIRST (TDD red). Implement in TypeScript using **Jest + React Testing Library**. Mock the backend at the API client layer (`web/lib/api/`) using **MSW** with handlers that match the architecture's exact JSON request/response shapes.

## Coverage matrix

| AC / UI flow | Test ID | Component / hook under test | What it asserts |
|---|---|---|---|
| "Add Project" button visible on dashboard | FCT-046-001 | `web/pages/index.tsx` | button renders with accessible name |
| Dialog opens on button click | FCT-046-002 | `web/components/Dashboard/AddProjectDialog.tsx` | dialog mounts on trigger |
| Path field is plain text input | FCT-046-003 | `AddProjectDialog` | input type="text", not file/search |
| Name auto-fills from path basename | FCT-046-004 | `AddProjectDialog` + `useCreateProject` | basename extraction |
| Name auto-fill does not overwrite manual edit | FCT-046-005 | `AddProjectDialog` | sticky-off behavior |
| Submit disabled when name blank | FCT-046-006 | `AddProjectDialog` | button disabled state |
| Submit disabled when path blank | FCT-046-007 | `AddProjectDialog` | button disabled state |
| Submit disabled when both blank | FCT-046-008 | `AddProjectDialog` | button disabled state |
| Submit enabled when both non-blank | FCT-046-009 | `AddProjectDialog` | button enabled state |
| Loading state during submission | FCT-046-010 | `AddProjectDialog` + `useCreateProject` | spinner/disabled during in-flight |
| Double-submit prevented | FCT-046-011 | `AddProjectDialog` + `useCreateProject` | second click while pending has no effect |
| Happy path: 201 → dialog closes + list refreshes | FCT-046-012 | `AddProjectDialog` + `web/pages/index.tsx` | post-success UI state |
| 400 VALIDATION_ERROR surfaced inline | FCT-046-013 | `AddProjectDialog` | error message visible; form stays open |
| 409 DUPLICATE_PATH surfaced inline | FCT-046-014 | `AddProjectDialog` | different error message; form stays open |
| Input preserved on server error | FCT-046-015 | `AddProjectDialog` | path and name values unchanged after error |
| useCreateProject — idle initial state | FCT-046-016 | `web/hooks/useCreateProject.ts` | status = idle |
| useCreateProject — submitting state | FCT-046-017 | `useCreateProject` | status = submitting during fetch |
| useCreateProject — success state | FCT-046-018 | `useCreateProject` | status = success, returns project |
| useCreateProject — error state (4xx) | FCT-046-019 | `useCreateProject` | status = error, error object populated |
| useCreateProject — network error | FCT-046-020 | `useCreateProject` | status = error on fetch rejection |
| createProject API client sends correct body | FCT-046-021 | `web/lib/api/projects.ts` → `createProject` | request body shape |
| createProject API client returns typed Project | FCT-046-022 | `web/lib/api/projects.ts` → `createProject` | return type including `path` field |
| Abort on unmount (useEffect cleanup) | FCT-046-023 | `useCreateProject` | AbortController signal aborts in-flight request |
| Accessibility: error message announced | FCT-046-024 | `AddProjectDialog` | aria-live or role=alert |
| Accessibility: focus returns to trigger on close | FCT-046-025 | `AddProjectDialog` | focus management |
| Accessibility: keyboard navigation in dialog | FCT-046-026 | `AddProjectDialog` | tab/shift-tab cycles through fields |

---

## MSW handler definitions (shared across all tests)

Define these in `web/test/msw/handlers.ts` (add to existing handlers array):

```typescript
// POST /api/v1/projects — happy path
http.post('/api/v1/projects', () => {
  return HttpResponse.json({
    id: '33333333-3333-3333-3333-333333333333',
    name: 'agents-board',
    description: '',
    path: '/Users/me/workspace/agents-board',
    createdAt: '2026-06-09T11:00:00Z',
    updatedAt: '2026-06-09T11:00:00Z',
  }, { status: 201 })
})

// POST /api/v1/projects — 400 invalid path
http.post('/api/v1/projects', () => {
  return HttpResponse.json(
    { code: 'VALIDATION_ERROR', message: 'path does not exist or is not a directory' },
    { status: 400 }
  )
})

// POST /api/v1/projects — 409 duplicate path
http.post('/api/v1/projects', () => {
  return HttpResponse.json(
    { code: 'DUPLICATE_PATH', message: 'path already linked to another project' },
    { status: 409 }
  )
})
```

Switch between handler variants per-test using MSW `server.use(...)` overrides.

---

## Component tests

### FCT-046-001 — "Add Project" button is visible on the dashboard
- **Component / hook under test:** `web/pages/index.tsx`
- **Render with:** Default MSW handlers (projects list 200 with empty array); no providers except query client.
- **MSW handlers:** `GET /api/v1/projects` → 200 `{"projects": []}`
- **User interactions:** None (render assertion only).
- **Expect:**
  - `screen.getByRole('button', { name: /add project/i })` is in the document.
- **Architecture cite:** US046 AC "Open the add-project form"; FE surface `web/pages/index.tsx`

### FCT-046-002 — Clicking "Add Project" opens the dialog
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx` (rendered within `index.tsx`)
- **Render with:** Dashboard page; MSW projects list handler.
- **User interactions:**
  1. `userEvent.click(screen.getByRole('button', { name: /add project/i }))`
- **Expect:**
  - `screen.getByRole('dialog')` is visible.
  - Dialog contains a path text input and a name text input.
- **Architecture cite:** US046 AC "Open the add-project form"

### FCT-046-003 — Path field is a plain text input (not file/search/etc.)
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx`
- **Render with:** Dialog open.
- **Expect:**
  - The path input has `type="text"` (not `type="file"`, not combobox, not search).
  - It has a label associated to it (via `htmlFor`/`aria-label`).
- **Architecture cite:** US046 AC "Type the full path manually — plain `<input type="text">`"; D-005

### FCT-046-004 — Name auto-fills with basename when path changes
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx`
- **Render with:** Dialog open, name field initially empty.
- **User interactions:**
  1. `userEvent.type(screen.getByLabelText(/path/i), '/Users/me/workspace/my-project')`
- **Expect:**
  - `screen.getByLabelText(/name/i)` has value `"my-project"` (the basename of the path).
- **Edge cases:**
  - Path with trailing slash: `/Users/me/workspace/my-project/` → basename = `"my-project"`.
  - Path that is just `"/"` → basename = `""` or `"/"` (implementation choice; name field should NOT auto-fill with an empty string as a meaningful value — the submit button should remain disabled).
- **Architecture cite:** US046 AC "Name auto-fills from path basename"

### FCT-046-005 — Name auto-fill is "sticky-off" once manually edited
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx`
- **User interactions:**
  1. Type `/some/path/project-a` into path field → name auto-fills to `"project-a"`.
  2. Clear the name field and type `"my-custom-name"` manually.
  3. Change the path field to `/some/other/project-b`.
- **Expect:**
  - Name field retains `"my-custom-name"` — path change did NOT overwrite the manually-set name.
- **Architecture cite:** US046 AC "if I edit the name manually, subsequent path changes do not overwrite my edit"

### FCT-046-006 — Submit button is disabled when name is blank
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx`
- **Render with:** Dialog open; path field has a value.
- **User interactions:**
  1. Type a valid path but leave name empty (after clearing auto-fill).
- **Expect:**
  - `screen.getByRole('button', { name: /create|submit|add/i })` has `disabled` attribute (or `aria-disabled="true"`).
  - No `fetch` call is made on click.
- **Architecture cite:** US046 AC "Client-side validation — empty required fields"

### FCT-046-007 — Submit button is disabled when path is blank
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx`
- **User interactions:** Fill name, leave path empty.
- **Expect:** Submit button disabled; no fetch call.
- **Architecture cite:** US046 AC "Client-side validation"

### FCT-046-008 — Submit button is disabled when both fields are blank
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx`
- **Render with:** Dialog freshly opened (no user input).
- **Expect:** Submit button disabled.
- **Architecture cite:** US046 AC "Client-side validation"

### FCT-046-009 — Submit button is enabled when both name and path are non-blank
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx`
- **User interactions:**
  1. Type `/some/valid/path` into path.
  2. Verify name auto-fills.
- **Expect:** Submit button is NOT disabled (`disabled` attribute absent).

### FCT-046-010 — Loading state: submit shows spinner/disabled while request is in flight
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx` + `useCreateProject`
- **MSW handler:** Override `POST /api/v1/projects` with a delayed response (use `http.post` + `async` handler that awaits a deferred promise, then resolve in the assertion step).
- **User interactions:**
  1. Fill path and name; click Submit.
- **Expect (immediately after click, before response):**
  - Submit button is disabled.
  - A loading indicator (spinner, `aria-busy`, or text "Creating...") is visible.
- **Architecture cite:** US046 AC "Submission in-flight" — "the submit control shows a loading/disabled state"

### FCT-046-011 — Double-submit is prevented
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx` + `useCreateProject`
- **MSW handler:** Delayed response same as FCT-046-010. Count number of times the handler is called.
- **User interactions:**
  1. Fill path and name; click Submit.
  2. Click Submit again while the first request is in flight.
- **Expect:**
  - The MSW handler is called exactly once (second click did nothing because the button is disabled or the hook guards against re-entry).
- **Architecture cite:** US046 AC "double-submit is prevented"

### FCT-046-012 — Happy path: 201 → dialog closes + projects list refreshes
- **Component / hook under test:** `web/pages/index.tsx` (includes `AddProjectDialog` + project list)
- **MSW handlers:**
  - `GET /api/v1/projects` → 200 `{"projects": []}`
  - `POST /api/v1/projects` → 201 with the exact shape from the architecture contract
  - After POST, override `GET /api/v1/projects` to return the newly created project
- **User interactions:**
  1. Click "Add Project".
  2. Fill path + name.
  3. Click Submit.
- **Expect (after response resolves):**
  - `screen.queryByRole('dialog')` is null (dialog closed).
  - The new project's name appears in the projects list.
- **Architecture cite:** §3 201 response; US046 AC "Successful create"

### FCT-046-013 — 400 VALIDATION_ERROR shown inline; form stays open
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx`
- **MSW handler:** Override `POST /api/v1/projects` → 400 `{"code":"VALIDATION_ERROR","message":"path does not exist or is not a directory"}`.
- **User interactions:**
  1. Fill path + name; click Submit.
- **Expect:**
  - `screen.getByText(/path does not exist or is not a directory/i)` is visible (inline near path field).
  - `screen.getByRole('dialog')` is still present (form NOT closed).
- **Architecture cite:** §3 400; US046 AC "Server validation error surfaced"

### FCT-046-014 — 409 DUPLICATE_PATH shown inline; form stays open
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx`
- **MSW handler:** Override `POST /api/v1/projects` → 409 `{"code":"DUPLICATE_PATH","message":"path already linked to another project"}`.
- **User interactions:**
  1. Fill path + name; click Submit.
- **Expect:**
  - `screen.getByText(/path already linked to another project/i)` is visible inline.
  - Dialog still open.
- **Architecture cite:** §3 409; US046 AC "Server validation error surfaced"

### FCT-046-015 — Input values are preserved after a server error
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx`
- **MSW handler:** 400 or 409.
- **User interactions:**
  1. Type `"/my/custom/path"` in path; name auto-fills.
  2. Submit → server returns 400.
- **Expect:**
  - Path field still shows `"/my/custom/path"`.
  - Name field retains its value.
- **Architecture cite:** US046 AC "the form stays open with my input preserved"

### FCT-046-016 — useCreateProject: initial state is idle
- **Component / hook under test:** `web/hooks/useCreateProject.ts`
- **Render with:** Isolated hook test (`renderHook`).
- **Expect:**
  - `result.current.status === 'idle'`
  - `result.current.error` is null/undefined.
- **Architecture cite:** US046 UX — "Empty: submit loading state doesn't show initially"

### FCT-046-017 — useCreateProject: status is submitting while request is in flight
- **Component / hook under test:** `web/hooks/useCreateProject.ts`
- **MSW handler:** Delayed POST handler.
- **When:** `result.current.createProject({ name: '...', path: '...' })` is called.
- **Expect (before response):**
  - `result.current.status === 'submitting'`
- **Architecture cite:** US046 UX — submit loading state

### FCT-046-018 — useCreateProject: status is success on 201
- **Component / hook under test:** `web/hooks/useCreateProject.ts`
- **MSW handler:** `POST /api/v1/projects` → 201 (exact architecture shape).
- **When:** `createProject` called; await response.
- **Expect:**
  - `result.current.status === 'success'` (or `'idle'` reset — implementation choice, but the returned project matches the shape)
  - The promise resolves with a project object including `path` field.
- **Architecture cite:** §3 201 response shape

### FCT-046-019 — useCreateProject: status is error on 4xx
- **Component / hook under test:** `web/hooks/useCreateProject.ts`
- **MSW handler:** 400.
- **When:** `createProject` called; await response.
- **Expect:**
  - `result.current.status === 'error'`
  - `result.current.error` has `code = 'VALIDATION_ERROR'` and `message` set.
- **Architecture cite:** §3 400 error body shape

### FCT-046-020 — useCreateProject: status is error on network failure
- **Component / hook under test:** `web/hooks/useCreateProject.ts`
- **MSW handler:** `http.post('/api/v1/projects', () => { return HttpResponse.error() })`
- **When:** `createProject` called.
- **Expect:**
  - `result.current.status === 'error'`
  - `result.current.error` is non-null.

### FCT-046-021 — createProject API client sends exact request body
- **Component / hook under test:** `web/lib/api/projects.ts` → `createProject`
- **MSW handler:** `POST /api/v1/projects` — capture and inspect the request body.
- **When:** `createProject({ name: 'Test', description: 'Desc', path: '/tmp/test' })` called.
- **Expect (MSW request inspection):**
  - Request body equals `{"name":"Test","description":"Desc","path":"/tmp/test"}`.
  - `Content-Type: application/json` header present.
- **Architecture cite:** §3 request body — `{ name, description, path }`

### FCT-046-022 — createProject API client returns typed Project including `path`
- **Component / hook under test:** `web/lib/api/projects.ts` → `createProject`
- **MSW handler:** 201 with exact architecture shape.
- **When:** `createProject(...)` called and awaited.
- **Expect:**
  - Returns an object with `id`, `name`, `description`, `path`, `createdAt`, `updatedAt`.
  - `path` is a non-empty string (not null, not absent).
  - TypeScript type `Project` has a `path: string` field (verified via `tsc --noEmit`).
- **Architecture cite:** §3 201 response; `web/lib/api/types.ts` — `Project.path`

### FCT-046-023 — useCreateProject: in-flight request is aborted on unmount
- **Component / hook under test:** `web/hooks/useCreateProject.ts`
- **MSW handler:** Delayed POST handler.
- **When:**
  1. Render hook; call `createProject(...)`.
  2. Before response resolves, call `unmount()`.
- **Expect:**
  - No state update happens after unmount (no "Can't perform a React state update on an unmounted component" warning).
  - The AbortController signal triggered the request cancellation (MSW receives the abort signal and the promise rejects with `AbortError`).
- **Architecture cite:** Tester policy — "one FCT-* per `useEffect` cleanup / abort path"

### FCT-046-024 — Error messages are announced to screen readers
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx`
- **When:** Server returns 400 and the error message is rendered.
- **Expect:**
  - The error container has `role="alert"` or is inside a live region (`aria-live="polite"` or `"assertive"`).
  - The error text is programmatically associated with or adjacent to the path input (via `aria-describedby` or label proximity).
- **Architecture cite:** Tester policy — accessibility surface: `aria-*` on error

### FCT-046-025 — Focus returns to "Add Project" button when dialog closes
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx`
- **When:**
  1. Click "Add Project" to open dialog.
  2. Close dialog (e.g. via Escape or a Cancel button).
- **Expect:**
  - Focus returns to the "Add Project" trigger button.
- **Architecture cite:** Tester policy — "one FCT-* per user-input edge: focus management"

### FCT-046-026 — Keyboard navigation cycles through dialog fields
- **Component / hook under test:** `web/components/Dashboard/AddProjectDialog.tsx`
- **When:**
  1. Dialog is open; focus is on the first focusable element.
  2. Press Tab repeatedly.
- **Expect:**
  - Tab moves focus through: path input → name input → submit button → (trap/close button) and cycles within the dialog (focus trap).
  - Shift+Tab moves in reverse.
- **Architecture cite:** Tester policy — "keyboard nav"
