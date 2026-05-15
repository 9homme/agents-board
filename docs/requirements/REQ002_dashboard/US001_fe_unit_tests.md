# US001 — Frontend component test specification

**For FE Dev:** these are the tests you write FIRST (TDD red). Implement in TypeScript using **Jest + React Testing Library**. Mock the backend at the API client layer (`web/lib/api/`) using **MSW** with handlers that match the architecture's exact JSON request/response shapes.

## Coverage matrix
| AC / UI flow | Test ID | Component / hook under test | What it asserts |
|---|---|---|---|
| Successfully load project list | FCT-001 | `web/pages/index.tsx` | renders project cards based on MSW response |
| Empty state | FCT-002 | `web/pages/index.tsx` | renders empty message when MSW returns empty array |
| Loading state | FCT-003 | `web/pages/index.tsx` | renders loading indicator while MSW request is pending |
| Error state | FCT-004 | `web/pages/index.tsx` | renders error message when MSW returns 500 |

## Component tests
### FCT-001 — Successfully load project list
- **Component / hook under test:** `web/pages/index.tsx`
- **Render with:** default providers, MSW handlers
- **MSW handlers:**
  - `GET /api/v1/projects` → 200 with `{"projects": [{"id": "1", "name": "Dashboard Test Project", "description": "A minimal beautiful dashboard", "createdAt": "2023-10-25T10:00:00Z", "updatedAt": "2023-10-25T10:00:00Z"}]}`
- **User interactions (RTL):** None
- **Expect:**
  - Wait for loading state to finish.
  - `screen.findByText('Dashboard Test Project')` is visible.
  - `screen.findByText('A minimal beautiful dashboard')` is visible.
- **Architecture cite:** API contract row `GET /api/v1/projects`, FE surface `web/pages/index.tsx`.

### FCT-002 — Empty state
- **Component / hook under test:** `web/pages/index.tsx`
- **Render with:** default providers, MSW handlers
- **MSW handlers:**
  - `GET /api/v1/projects` → 200 with `{"projects": []}`
- **User interactions (RTL):** None
- **Expect:**
  - Wait for loading state to finish.
  - `screen.findByText(/no projects/i)` is visible.
- **Architecture cite:** API contract row `GET /api/v1/projects`, FE surface `web/pages/index.tsx`.

### FCT-003 — Loading state
- **Component / hook under test:** `web/pages/index.tsx`
- **Render with:** default providers, MSW handlers configured with delay
- **MSW handlers:**
  - `GET /api/v1/projects` → delayed response
- **User interactions (RTL):** None
- **Expect:**
  - A visual loading indicator (e.g. text "Loading", spinner, or skeleton) is immediately visible.
  - Loading indicator disappears after data resolves.

### FCT-004 — Error state
- **Component / hook under test:** `web/pages/index.tsx`
- **Render with:** default providers, MSW handlers
- **MSW handlers:**
  - `GET /api/v1/projects` → 500 with `{"code": "INTERNAL_ERROR", "message": "Failed to fetch projects"}`
- **User interactions (RTL):** None
- **Expect:**
  - Wait for loading state to finish.
  - `screen.findByText(/failed to load projects/i)` is visible.
- **Architecture cite:** API contract row `GET /api/v1/projects`, FE surface `web/pages/index.tsx`.