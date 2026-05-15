import { http, HttpResponse } from 'msw'

export const handlers = [
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
]
