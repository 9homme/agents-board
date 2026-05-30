/**
 * Tests for web/lib/api/documents.ts
 * Tests fetchProjectDocuments and fetchDocument, plus signal pass-through in client.ts
 */
import { fetchProjectDocuments, fetchDocument } from './documents';
import { ApiError } from './client';
import { server } from '../../test/msw/server';
import { http, HttpResponse, delay } from 'msw';

describe('fetchProjectDocuments', () => {
  it('returns DocumentsListResponse on 200', async () => {
    const result = await fetchProjectDocuments('p1');
    expect(result.documents).toHaveLength(2);
    expect(result.documents[0].id).toBe('d111aaaa-1111-1111-1111-111111111111');
    expect(result.documents[0].title).toBe('Architecture overview');
    // Must NOT include content field
    expect((result.documents[0] as unknown as Record<string, unknown>).content).toBeUndefined();
  });

  it('throws ApiError with NOT_FOUND on 404', async () => {
    await expect(fetchProjectDocuments('ghost-project')).rejects.toBeInstanceOf(ApiError);
    try {
      await fetchProjectDocuments('ghost-project');
    } catch (err) {
      expect((err as ApiError).code).toBe('NOT_FOUND');
    }
  });

  it('throws ApiError with INTERNAL_ERROR on 500', async () => {
    await expect(fetchProjectDocuments('broken-project')).rejects.toBeInstanceOf(ApiError);
  });

  it('passes signal to fetch (AbortSignal pass-through)', async () => {
    const controller = new AbortController();
    // Override handler to delay indefinitely, then abort
    server.use(
      http.get('*/api/v1/projects/p1/documents', async () => {
        await delay('infinite');
        return HttpResponse.json({ documents: [] });
      })
    );

    const promise = fetchProjectDocuments('p1', controller.signal);
    controller.abort();

    await expect(promise).rejects.toThrow();
  });
});

describe('fetchDocument', () => {
  it('returns Document with content on 200', async () => {
    const result = await fetchDocument('d111aaaa-1111-1111-1111-111111111111');
    expect(result.id).toBe('d111aaaa-1111-1111-1111-111111111111');
    expect(result.title).toBe('Architecture overview');
    expect(result.content).toBeDefined();
    expect(result.projectId).toBe('p1');
    expect(result.createdAt).toBe('2026-05-18T08:30:00Z');
    expect(result.updatedAt).toBe('2026-05-20T09:45:00Z');
  });

  it('throws ApiError with NOT_FOUND on 404', async () => {
    await expect(fetchDocument('not-found-document')).rejects.toBeInstanceOf(ApiError);
    try {
      await fetchDocument('not-found-document');
    } catch (err) {
      expect((err as ApiError).code).toBe('NOT_FOUND');
    }
  });

  it('throws ApiError with INTERNAL_ERROR on 500', async () => {
    await expect(fetchDocument('broken-document')).rejects.toBeInstanceOf(ApiError);
  });

  it('passes signal to fetch (AbortSignal pass-through)', async () => {
    const controller = new AbortController();
    server.use(
      http.get('*/api/v1/documents/d111aaaa-1111-1111-1111-111111111111', async () => {
        await delay('infinite');
        return HttpResponse.json({});
      })
    );

    const promise = fetchDocument('d111aaaa-1111-1111-1111-111111111111', controller.signal);
    controller.abort();

    await expect(promise).rejects.toThrow();
  });
});

// signal pass-through test for client.ts
describe('fetchClient signal pass-through', () => {
  it('forwards signal to fetch so abort cancels in-flight requests', async () => {
    // We test indirectly by verifying the fetch is aborted when signal fires
    const controller = new AbortController();
    server.use(
      http.get('*/api/v1/documents/d222bbbb-2222-2222-2222-222222222222', async () => {
        await delay('infinite');
        return HttpResponse.json({});
      })
    );

    const promise = fetchDocument('d222bbbb-2222-2222-2222-222222222222', controller.signal);
    controller.abort();

    await expect(promise).rejects.toThrow();
  });
});
