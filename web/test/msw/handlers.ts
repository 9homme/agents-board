import { http, HttpResponse } from 'msw'

// ---------------------------------------------------------------------------
// Shared fixtures
// ---------------------------------------------------------------------------

const TWO_DOCUMENT_LIST = [
  {
    id: 'd111aaaa-1111-1111-1111-111111111111',
    projectId: 'p1',
    title: 'Architecture overview',
    createdAt: '2026-05-18T08:30:00Z',
    updatedAt: '2026-05-20T09:45:00Z',
  },
  {
    id: 'd222bbbb-2222-2222-2222-222222222222',
    projectId: 'p1',
    title: 'Onboarding guide',
    createdAt: '2026-05-15T11:00:00Z',
    updatedAt: '2026-05-19T16:20:00Z',
  },
]

export const handlers = [
  // ---------------------------------------------------------------------------
  // POST /api/v1/projects — create a project (§3 happy path)
  // ---------------------------------------------------------------------------
  http.post('*/api/v1/projects', () => {
    return HttpResponse.json(
      {
        id: '33333333-3333-3333-3333-333333333333',
        name: 'agents-board',
        description: '',
        path: '/Users/me/workspace/agents-board',
        createdAt: '2026-06-09T11:00:00Z',
        updatedAt: '2026-06-09T11:00:00Z',
      },
      { status: 201 }
    )
  }),

  // ---------------------------------------------------------------------------
  // GET /api/v1/projects — list all projects
  // ---------------------------------------------------------------------------
  http.get('*/api/v1/projects', () => {
    return HttpResponse.json({
      projects: [
        {
          id: '1',
          name: 'Dashboard Test Project',
          description: 'A minimal beautiful dashboard',
          path: '/tmp/dashboard-test-project',
          createdAt: '2023-10-25T10:00:00Z',
          updatedAt: '2023-10-25T10:00:00Z',
        },
      ],
    })
  }),

  // ---------------------------------------------------------------------------
  // GET /api/v1/projects/:id — single project (bare object per architecture §1)
  // ---------------------------------------------------------------------------

  // Happy path: proj-001
  http.get('*/api/v1/projects/proj-001', () => {
    return HttpResponse.json({
      id: 'proj-001',
      name: 'E-commerce Website',
      description: 'A new online store for electronics',
      path: '/tmp/ecommerce-website',
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
      path: '/tmp/project-alpha',
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

  // 404 variant: ghost-project (for FCT-US002-015)
  http.get('*/api/v1/projects/ghost-project', () => {
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

  // ---------------------------------------------------------------------------
  // GET /api/v1/projects/:id/documents — list documents for a project
  // Architecture §2: 200 with documents array; 404 when project not found; 500
  // ---------------------------------------------------------------------------

  // p1 happy path — two documents (updatedAt DESC order)
  http.get('*/api/v1/projects/p1/documents', () => {
    return HttpResponse.json({ documents: TWO_DOCUMENT_LIST })
  }),

  // proj-001 happy path — same two documents (re-used fixture)
  http.get('*/api/v1/projects/proj-001/documents', () => {
    return HttpResponse.json({ documents: TWO_DOCUMENT_LIST })
  }),

  // ghost-project — 404 (project not found per D-006)
  http.get('*/api/v1/projects/ghost-project/documents', () => {
    return HttpResponse.json(
      { code: 'NOT_FOUND', message: 'Project not found' },
      { status: 404 }
    )
  }),

  // broken-project — 500
  http.get('*/api/v1/projects/broken-project/documents', () => {
    return HttpResponse.json(
      { code: 'INTERNAL_ERROR', message: 'Failed to fetch documents' },
      { status: 500 }
    )
  }),

  // ---------------------------------------------------------------------------
  // GET /api/v1/documents/:id — single document with content (architecture §3)
  // ---------------------------------------------------------------------------

  // d111aaaa — Architecture overview (full content)
  http.get('*/api/v1/documents/d111aaaa-1111-1111-1111-111111111111', () => {
    return HttpResponse.json({
      id: 'd111aaaa-1111-1111-1111-111111111111',
      projectId: 'p1',
      title: 'Architecture overview',
      content: '# Architecture\n\nThis project uses…\n\n```mermaid\ngraph TD; A-->B;\n```\n',
      createdAt: '2026-05-18T08:30:00Z',
      updatedAt: '2026-05-20T09:45:00Z',
    })
  }),

  // d222bbbb — Onboarding guide (full content)
  http.get('*/api/v1/documents/d222bbbb-2222-2222-2222-222222222222', () => {
    return HttpResponse.json({
      id: 'd222bbbb-2222-2222-2222-222222222222',
      projectId: 'p1',
      title: 'Onboarding guide',
      content: '# Onboarding\n\nWelcome.',
      createdAt: '2026-05-15T11:00:00Z',
      updatedAt: '2026-05-19T16:20:00Z',
    })
  }),

  // doc-B — for FCT-US002-007 race-cancellation test
  http.get('*/api/v1/documents/doc-B', () => {
    return HttpResponse.json({
      id: 'doc-B',
      projectId: 'p1',
      title: 'Doc B',
      content: 'Doc B content',
      createdAt: '2026-05-20T10:00:00Z',
      updatedAt: '2026-05-20T10:00:00Z',
    })
  }),

  // broken-document — 500
  http.get('*/api/v1/documents/broken-document', () => {
    return HttpResponse.json(
      { code: 'INTERNAL_ERROR', message: 'Failed to fetch document' },
      { status: 500 }
    )
  }),

  // not-found-document — 404
  http.get('*/api/v1/documents/not-found-document', () => {
    return HttpResponse.json(
      { code: 'NOT_FOUND', message: 'Document not found' },
      { status: 404 }
    )
  }),

  // ---------------------------------------------------------------------------
  // GET /api/v1/projects/:id/user-stories — list user stories (US004)
  // Architecture: 200 with userStories array; 404 when project not found; 500
  // ---------------------------------------------------------------------------

  // broken-project — 500
  http.get('*/api/v1/projects/broken-project/user-stories', () => {
    return HttpResponse.json(
      { code: 'INTERNAL_ERROR', message: 'Failed to fetch user stories' },
      { status: 500 }
    )
  }),

  // ghost-project — 404 (project not found)
  http.get('*/api/v1/projects/ghost-project/user-stories', () => {
    return HttpResponse.json(
      { code: 'NOT_FOUND', message: 'Project not found' },
      { status: 404 }
    )
  }),

  // Wildcard happy path — returns one story for any other project id
  http.get('*/api/v1/projects/:id/user-stories', ({ params }) => {
    const id = typeof params.id === 'string' ? params.id : String(params.id);
    return HttpResponse.json({
      userStories: [
        {
          id: 'us1',
          projectId: id,
          title: 'Add item to basket',
          description: 'As a shopper I want to add an item to my basket.',
          status: 'in_development',
          taskCount: 3,
          createdAt: '2024-01-01T00:00:00Z',
          updatedAt: '2024-01-02T09:30:00Z',
        },
      ],
    })
  }),

  // ---------------------------------------------------------------------------
  // GET /api/v1/user-stories/:id/tasks — tasks for a user story (US005)
  // Architecture §3: wrapped tasks array; 404 / 500
  // NOTE: must be registered BEFORE the bare /:id handler to avoid shadowing.
  // ---------------------------------------------------------------------------

  // Wildcard happy path — returns two tasks for any story id
  http.get('*/api/v1/user-stories/:id/tasks', ({ params }) => {
    const id = typeof params.id === 'string' ? params.id : String(params.id);
    return HttpResponse.json({
      tasks: [
        {
          id: 'task-001',
          userStoryId: id,
          title: 'be_basket_repo',
          description: 'Implement the basket repository layer.',
          status: 'completed',
          createdAt: '2024-01-01T10:00:00Z',
          updatedAt: '2024-01-02T11:00:00Z',
        },
        {
          id: 'task-002',
          userStoryId: id,
          title: 'fe_basket_button',
          description: 'Add the add-to-basket button.',
          status: 'in_review',
          createdAt: '2024-01-02T10:00:00Z',
          updatedAt: '2024-01-03T11:00:00Z',
        },
      ],
    })
  }),

  // ---------------------------------------------------------------------------
  // GET /api/v1/user-stories/:id — single user story detail (US005)
  // Architecture §2: bare UserStory object (no taskCount); 404 / 500
  // NOTE: registered AFTER /:id/tasks to avoid shadowing.
  // ---------------------------------------------------------------------------

  // Wildcard happy path — returns story detail for any id
  http.get('*/api/v1/user-stories/:id', ({ params }) => {
    const id = typeof params.id === 'string' ? params.id : String(params.id);
    return HttpResponse.json({
      id,
      projectId: 'proj-001',
      title: 'Add item to basket',
      description: 'As a shopper I want to add an item to my basket so that I can purchase it later.',
      status: 'in_development',
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-02T09:30:00Z',
    })
  }),

  // ---------------------------------------------------------------------------
  // US047 — Requirements hierarchy endpoints
  // ---------------------------------------------------------------------------

  // §4 GET /api/v1/projects/:pid/requirements — requirements list
  // Default happy path: one "Default" requirement
  http.get('*/api/v1/projects/:pid/requirements', ({ params }) => {
    const pid = typeof params.pid === 'string' ? params.pid : String(params.pid)
    return HttpResponse.json({
      requirements: [
        {
          id: 'b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f',
          projectId: pid,
          name: 'Default',
          description: '',
          status: 'draft',
          createdAt: '2026-06-09T10:00:00Z',
          updatedAt: '2026-06-09T10:00:00Z',
        },
      ],
    })
  }),

  // §6 GET /api/v1/projects/:pid/requirements/:rid/user-stories
  http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories', ({ params }) => {
    const pid = typeof params.pid === 'string' ? params.pid : String(params.pid)
    const rid = typeof params.rid === 'string' ? params.rid : String(params.rid)
    return HttpResponse.json({
      userStories: [
        {
          id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
          projectId: pid,
          requirementId: rid,
          title: 'Add item to basket',
          description: '',
          status: 'in_progress',
          taskCount: 3,
          createdAt: '2026-06-02T09:00:00Z',
          updatedAt: '2026-06-02T09:00:00Z',
        },
      ],
    })
  }),

  // §10 GET /api/v1/projects/:pid/requirements/:rid/documents
  // Returns the two-document fixture so existing page tests still see Architecture/Onboarding.
  http.get('*/api/v1/projects/:pid/requirements/:rid/documents', ({ params }) => {
    const pid = typeof params.pid === 'string' ? params.pid : String(params.pid)
    const rid = typeof params.rid === 'string' ? params.rid : String(params.rid)
    return HttpResponse.json({
      documents: [
        {
          id: 'd111aaaa-1111-1111-1111-111111111111',
          projectId: pid,
          requirementId: rid,
          title: 'Architecture overview',
          createdAt: '2026-05-18T08:30:00Z',
          updatedAt: '2026-05-20T09:45:00Z',
        },
        {
          id: 'd222bbbb-2222-2222-2222-222222222222',
          projectId: pid,
          requirementId: rid,
          title: 'Onboarding guide',
          createdAt: '2026-05-15T11:00:00Z',
          updatedAt: '2026-05-19T16:20:00Z',
        },
      ],
    })
  }),

  // §10 GET /api/v1/projects/:pid/requirements/:rid/documents/:did — single document with content
  http.get('*/api/v1/projects/:pid/requirements/:rid/documents/:did', ({ params }) => {
    const pid = typeof params.pid === 'string' ? params.pid : String(params.pid)
    const rid = typeof params.rid === 'string' ? params.rid : String(params.rid)
    const did = typeof params.did === 'string' ? params.did : String(params.did)
    if (did === 'broken-document' || did === 'doc-bad') {
      return HttpResponse.json({ code: 'INTERNAL_ERROR', message: 'Failed to fetch document' }, { status: 500 })
    }
    if (did === 'not-found-document') {
      return HttpResponse.json({ code: 'NOT_FOUND', message: 'Document not found' }, { status: 404 })
    }
    const knownContent: Record<string, string> = {
      'd111aaaa-1111-1111-1111-111111111111': '# Architecture\n\nThis project uses…\n\n```mermaid\ngraph TD; A-->B;\n```\n',
      'd222bbbb-2222-2222-2222-222222222222': '# Onboarding\n\nWelcome.',
    }
    const knownMeta: Record<string, { title: string; createdAt: string; updatedAt: string }> = {
      'd111aaaa-1111-1111-1111-111111111111': { title: 'Architecture overview', createdAt: '2026-05-18T08:30:00Z', updatedAt: '2026-05-20T09:45:00Z' },
      'd222bbbb-2222-2222-2222-222222222222': { title: 'Onboarding guide', createdAt: '2026-05-15T11:00:00Z', updatedAt: '2026-05-19T16:20:00Z' },
    }
    const meta = knownMeta[did] ?? { title: 'Test Document', createdAt: '2026-05-20T10:00:00Z', updatedAt: '2026-05-20T10:00:00Z' }
    return HttpResponse.json({
      id: did,
      projectId: pid,
      requirementId: rid,
      title: meta.title,
      content: knownContent[did] ?? '# Test',
      createdAt: meta.createdAt,
      updatedAt: meta.updatedAt,
    })
  }),

  // §6 GET /api/v1/projects/:pid/requirements/:rid/user-stories/:sid/tasks — must be before detail
  http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories/:sid/tasks', ({ params }) => {
    const sid = typeof params.sid === 'string' ? params.sid : String(params.sid)
    if (sid === 'broken-story') {
      return HttpResponse.json({ code: 'INTERNAL_ERROR', message: 'Failed to fetch tasks' }, { status: 500 })
    }
    if (sid === 'not-found-story') {
      return HttpResponse.json({ code: 'NOT_FOUND', message: 'User story not found' }, { status: 404 })
    }
    if (sid === 'us-empty') {
      return HttpResponse.json({ tasks: [] })
    }
    return HttpResponse.json({
      tasks: [
        { id: 'task-001', userStoryId: sid, title: 'be_basket_repo', description: 'Implement the basket repository layer.', status: 'completed', createdAt: '2024-01-01T10:00:00Z', updatedAt: '2024-01-02T11:00:00Z' },
        { id: 'task-002', userStoryId: sid, title: 'fe_basket_button', description: 'Add the add-to-basket button.', status: 'in_review', createdAt: '2024-01-02T10:00:00Z', updatedAt: '2024-01-03T11:00:00Z' },
      ],
    })
  }),

  // §6 GET /api/v1/projects/:pid/requirements/:rid/user-stories/:sid — single story detail
  http.get('*/api/v1/projects/:pid/requirements/:rid/user-stories/:sid', ({ params }) => {
    const pid = typeof params.pid === 'string' ? params.pid : String(params.pid)
    const rid = typeof params.rid === 'string' ? params.rid : String(params.rid)
    const sid = typeof params.sid === 'string' ? params.sid : String(params.sid)
    if (sid === 'broken-story') {
      return HttpResponse.json({ code: 'INTERNAL_ERROR', message: 'Failed to fetch user story' }, { status: 500 })
    }
    if (sid === 'not-found-story') {
      return HttpResponse.json({ code: 'NOT_FOUND', message: 'User story not found' }, { status: 404 })
    }
    return HttpResponse.json({
      id: sid,
      projectId: pid,
      requirementId: rid,
      title: 'Add item to basket',
      description: 'As a shopper I want to add an item to my basket so that I can purchase it later.',
      status: 'in_development',
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-02T09:30:00Z',
    })
  }),
]
