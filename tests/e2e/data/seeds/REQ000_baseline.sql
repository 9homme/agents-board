-- REQ000_baseline.sql
-- Baseline seed fixture for e2e test runs.
-- Provides one project + one requirement + two documents with deterministic UUIDs so Robot suites
-- can reference them by ID without a discovery step.
--
-- TRUNCATE first so repeated `make e2e-seed` calls always produce a clean slate.
-- Architecture ref: REQ005 §6.5 seed fixture contract.
-- Updated REQ008: added path (NOT NULL), requirements row, requirement_id on documents.

TRUNCATE TABLE tasks, user_stories, documents, requirements, projects RESTART IDENTITY CASCADE;

-- Deterministic project UUID for Robot suite references.
-- Project: "Sample Project" (baseline, no real-domain meaning).
INSERT INTO projects (id, name, description, path, created_at, updated_at)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'Sample Project',
  'Baseline seed project for e2e tests.',
  '/e2e-baseline-project',
  '2024-01-01T00:00:00Z',
  '2024-01-01T00:00:00Z'
)
ON CONFLICT DO NOTHING;

-- Baseline requirement for the seed project.
INSERT INTO requirements (id, project_id, name, description, status, created_at, updated_at)
VALUES (
  '00000000-0000-0000-0000-000000000002',
  '00000000-0000-0000-0000-000000000001',
  'Default',
  '',
  'draft',
  '2024-01-01T00:00:00Z',
  '2024-01-01T00:00:00Z'
)
ON CONFLICT DO NOTHING;

-- Document 1: "Introduction" document in the baseline project.
INSERT INTO documents (id, project_id, requirement_id, title, content, created_at, updated_at)
VALUES (
  '00000000-0000-0000-0000-000000000011',
  '00000000-0000-0000-0000-000000000001',
  '00000000-0000-0000-0000-000000000002',
  'Introduction',
  E'# Introduction\n\nThis is the baseline e2e seed document.',
  '2024-01-01T00:00:00Z',
  '2024-01-01T00:00:00Z'
)
ON CONFLICT DO NOTHING;

-- Document 2: "Architecture" document in the baseline project.
INSERT INTO documents (id, project_id, requirement_id, title, content, created_at, updated_at)
VALUES (
  '00000000-0000-0000-0000-000000000012',
  '00000000-0000-0000-0000-000000000001',
  '00000000-0000-0000-0000-000000000002',
  'Architecture',
  E'# Architecture\n\nBaseline seed architecture document.',
  '2024-01-01T00:00:00Z',
  '2024-01-01T00:00:00Z'
)
ON CONFLICT DO NOTHING;
