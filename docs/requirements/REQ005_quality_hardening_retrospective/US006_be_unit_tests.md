# US006 — Backend unit & integration test specification (stub)

**No BE changes.** US006 is a frontend-only story (harmonising `useProject`, `useProjectDocuments`, and `useDocument` on the AbortController pattern, plus adding `signal?: AbortSignal` to `fetchProject` and `fetchProjects` in `web/lib/api/projects.ts`).

Per architecture §1.2 and §8: no endpoints are added, removed, or changed. No Go files are touched. The existing API surface (`GET /api/v1/projects`, `GET /api/v1/projects/{id}`, `GET /api/v1/projects/{id}/documents`, `GET /api/v1/documents/{id}`) is preserved byte-for-byte.

**BE Dev: there is nothing for you to implement or test for US006. This stub exists so the test-coverage matrix is complete and the absence of BE work is explicit — not an oversight.**
