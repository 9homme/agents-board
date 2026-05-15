# REQ002 — dashboard

## Summary
Add a web-based dashboard that displays a list of available projects using a minimal and beautiful card-based UI.

## Business Goal
Allow human users to easily view existing projects in the system through a web interface, complementing the existing agent-facing MCP capabilities.

## Confirmed Decisions
- The dashboard will exclusively feature listing projects. Other project features (create, edit, delete, or navigating to project details) are strictly out of scope for this requirement.

## User Stories
- `US001_view_project_dashboard.md` - View list of projects as cards

## Tasks
| Task | Title | Track | Status | Blocked By |
|---|---|---|---|---|
| `US001_be_service_rename_and_api_server_scaffold.md` | Scaffold API server and rename service | BE | pending | |
| `US001_be_get_projects_api.md` | Implement GET /api/v1/projects | BE | pending | `US001_be_service_rename_and_api_server_scaffold.md` |
| `US001_fe_scaffold_api_client.md` | Scaffold Next.js API client & MSW mocks | FE | pending | |
| `US001_fe_dashboard_page.md` | Implement dashboard UI | FE | pending | `US001_fe_scaffold_api_client.md` |

