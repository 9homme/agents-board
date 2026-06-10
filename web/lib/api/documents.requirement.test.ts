/**
 * Tests for web/lib/api/documents.ts — requirement-scoped fetcher
 * FCT-047-024 — fetchRequirementDocuments sends correct URL
 */
import { server } from '../../test/msw/server';
import { http, HttpResponse, delay } from 'msw';
import { ApiError } from './client';

const PROJECT_ID = '11111111-1111-1111-1111-111111111111';
const REQ_ID = 'b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f';

describe('fetchRequirementDocuments', () => {
  // FCT-047-024 — sends GET /api/v1/projects/:pid/requirements/:rid/documents
  it('FCT-047-024: sends correct canonical URL and parses §10 shape', async () => {
    const { fetchRequirementDocuments } = await import('./documents');

    let capturedUrl: string | undefined;
    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/documents', ({ request, params }) => {
        capturedUrl = request.url;
        const pid = typeof params.pid === 'string' ? params.pid : String(params.pid);
        const rid = typeof params.rid === 'string' ? params.rid : String(params.rid);
        return HttpResponse.json({
          documents: [
            {
              id: 'cccccccc-cccc-cccc-cccc-cccccccccccc',
              projectId: pid,
              requirementId: rid,
              title: 'README',
              createdAt: '2026-06-02T09:00:00Z',
              updatedAt: '2026-06-02T09:00:00Z',
            },
          ],
        });
      })
    );

    const result = await fetchRequirementDocuments(PROJECT_ID, REQ_ID);

    expect(capturedUrl).toContain(
      `/api/v1/projects/${PROJECT_ID}/requirements/${REQ_ID}/documents`
    );
    expect(result.documents).toHaveLength(1);
    const doc = result.documents[0];
    expect(doc.id).toBe('cccccccc-cccc-cccc-cccc-cccccccccccc');
    expect(doc.projectId).toBe(PROJECT_ID);
    expect(doc.requirementId).toBe(REQ_ID);
    expect(doc.title).toBe('README');
    expect(doc.createdAt).toBe('2026-06-02T09:00:00Z');
    expect(doc.updatedAt).toBe('2026-06-02T09:00:00Z');
  });

  it('does NOT call the old flat route /api/v1/projects/:pid/documents', async () => {
    const { fetchRequirementDocuments } = await import('./documents');

    let flatRouteHit = false;
    server.use(
      http.get('*/api/v1/projects/:pid/documents', () => {
        flatRouteHit = true;
        return HttpResponse.json({ documents: [] });
      })
    );

    await fetchRequirementDocuments(PROJECT_ID, REQ_ID);
    expect(flatRouteHit).toBe(false);
  });

  it('returns empty array on 200 with no documents', async () => {
    const { fetchRequirementDocuments } = await import('./documents');

    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/documents', () => {
        return HttpResponse.json({ documents: [] });
      })
    );

    const result = await fetchRequirementDocuments(PROJECT_ID, REQ_ID);
    expect(result.documents).toHaveLength(0);
  });

  it('throws ApiError with NOT_FOUND on 404', async () => {
    const { fetchRequirementDocuments } = await import('./documents');

    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/documents', () => {
        return HttpResponse.json(
          { code: 'NOT_FOUND', message: 'Requirement not found' },
          { status: 404 }
        );
      })
    );

    await expect(fetchRequirementDocuments(PROJECT_ID, REQ_ID)).rejects.toBeInstanceOf(ApiError);
    try {
      await fetchRequirementDocuments(PROJECT_ID, REQ_ID);
    } catch (err) {
      expect((err as ApiError).code).toBe('NOT_FOUND');
    }
  });

  it('throws ApiError on 500', async () => {
    const { fetchRequirementDocuments } = await import('./documents');

    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/documents', () => {
        return HttpResponse.json(
          { code: 'INTERNAL_ERROR', message: 'Failed to fetch documents' },
          { status: 500 }
        );
      })
    );

    await expect(fetchRequirementDocuments(PROJECT_ID, REQ_ID)).rejects.toBeInstanceOf(ApiError);
  });

  it('passes AbortSignal to fetch', async () => {
    const { fetchRequirementDocuments } = await import('./documents');

    const controller = new AbortController();
    server.use(
      http.get('*/api/v1/projects/:pid/requirements/:rid/documents', async () => {
        await delay('infinite');
        return HttpResponse.json({ documents: [] });
      })
    );

    const promise = fetchRequirementDocuments(PROJECT_ID, REQ_ID, controller.signal);
    controller.abort();
    await expect(promise).rejects.toThrow();
  });
});
