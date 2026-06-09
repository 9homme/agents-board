-- REQ000_baseline.sql
-- Baseline seed fixture for e2e test runs.
-- Provides one project + two documents with deterministic UUIDs so Robot suites
-- can reference them by ID without a discovery step.
--
-- Idempotency: INSERT ... ON CONFLICT DO NOTHING — safe to run multiple times.
-- Architecture ref: REQ005 §6.5 seed fixture contract.

-- Deterministic project UUID for Robot suite references.
-- Project: "Sample Project" (baseline, no real-domain meaning).
INSERT INTO projects (id, name, description, created_at, updated_at)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'Sample Project',
  'Baseline seed project for e2e tests.',
  '2024-01-01T00:00:00Z',
  '2024-01-01T00:00:00Z'
)
ON CONFLICT DO NOTHING;

-- Document 1: "Introduction" document in the baseline project.
INSERT INTO documents (id, project_id, title, content, created_at, updated_at)
VALUES (
  '00000000-0000-0000-0000-000000000011',
  '00000000-0000-0000-0000-000000000001',
  'Introduction',
  E'# Introduction\n\nThis is the baseline e2e seed document.',
  '2024-01-01T00:00:00Z',
  '2024-01-01T00:00:00Z'
)
ON CONFLICT DO NOTHING;

-- Document 2: "Architecture" document in the baseline project.
INSERT INTO documents (id, project_id, title, content, created_at, updated_at)
VALUES (
  '00000000-0000-0000-0000-000000000012',
  '00000000-0000-0000-0000-000000000001',
  'Architecture',
  E'# Architecture\n\nBaseline seed architecture document.',
  '2024-01-01T00:00:00Z',
  '2024-01-01T00:00:00Z'
)
ON CONFLICT DO NOTHING;
