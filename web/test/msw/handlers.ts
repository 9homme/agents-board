import { http, HttpResponse } from 'msw'

export const handlers = [
  // GET /api/v1/projects — list all projects
  http.get('*/api/v1/projects', () => {
    return HttpResponse.json({
      projects: [
        {
          id: '1',
          name: 'Dashboard Test Project',
          description: 'A minimal beautiful dashboard',
          createdAt: '2023-10-25T10:00:00Z',
          updatedAt: '2023-10-25T10:00:00Z',
        },
      ],
    })
  }),

  // GET /api/v1/projects/:id — single project (bare object per architecture §1)
  // Happy path: proj-001
  http.get('*/api/v1/projects/proj-001', () => {
    return HttpResponse.json({
      id: 'proj-001',
      name: 'E-commerce Website',
      description: 'A new online store for electronics',
      createdAt: '2026-05-20T10:00:00Z',
      updatedAt: '2026-05-20T10:00:00Z',
    })
  }),

  // Happy path: p1
  http.get('*/api/v1/projects/p1', () => {
    return HttpResponse.json({
      id: 'p1',
      name: 'Project Alpha',
      description: 'A valid project',
      createdAt: '2026-05-20T10:00:00Z',
      updatedAt: '2026-05-20T10:00:00Z',
    })
  }),

  // 404 variant: no-such-project
  http.get('*/api/v1/projects/no-such-project', () => {
    return HttpResponse.json(
      { code: 'NOT_FOUND', message: 'Project not found' },
      { status: 404 }
    )
  }),

  // 500 variant: broken-project
  http.get('*/api/v1/projects/broken-project', () => {
    return HttpResponse.json(
      { code: 'INTERNAL_ERROR', message: 'Failed to fetch project' },
      { status: 500 }
    )
  }),
]
