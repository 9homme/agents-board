-- Documentation-only: the migration runner does NOT execute *.down.sql.
-- Provided for correctness guarantees and manual rollback reference.
--
-- Data-loss note: the Default requirement grouping is lost on down.
-- The original user_stories and documents rows remain (core content preserved),
-- but their requirement_id associations are dropped.

-- 1. Drop child indexes first
DROP INDEX IF EXISTS idx_user_stories_requirement_id;
DROP INDEX IF EXISTS idx_documents_requirement_id;

-- 2. Drop requirement_id columns from child tables (before dropping requirements)
ALTER TABLE user_stories DROP COLUMN IF EXISTS requirement_id;
ALTER TABLE documents    DROP COLUMN IF EXISTS requirement_id;

-- 3. Drop projects.path and its unique constraint
ALTER TABLE projects DROP CONSTRAINT IF EXISTS uq_projects_path;
ALTER TABLE projects DROP COLUMN IF EXISTS path;

-- 4. Drop the requirements table (and its index)
DROP INDEX IF EXISTS idx_requirements_project_id;
DROP TABLE IF EXISTS requirements;
