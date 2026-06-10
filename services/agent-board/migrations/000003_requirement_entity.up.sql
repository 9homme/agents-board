-- 1. requirements table
CREATE TABLE requirements (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      VARCHAR(50) NOT NULL DEFAULT 'draft'
                CHECK (status IN ('draft', 'in_progress', 'done')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_requirements_project_id ON requirements(project_id);

-- 2. projects.path — NOT NULL, unique
-- Add column nullable first so existing rows can be backfilled with unique values
-- before NOT NULL + UNIQUE are enforced (avoids duplicate-empty-string violation).
ALTER TABLE projects ADD COLUMN path TEXT;

-- Backfill existing projects with their id as a unique placeholder path;
-- the application layer always sets a real non-blank path via the API.
UPDATE projects SET path = id::text WHERE path IS NULL;

-- Now enforce NOT NULL and uniqueness
ALTER TABLE projects ALTER COLUMN path SET NOT NULL;
ALTER TABLE projects ADD CONSTRAINT uq_projects_path UNIQUE (path);

-- 3. re-parent columns, added NULLABLE first so the backfill can populate them
ALTER TABLE user_stories ADD COLUMN requirement_id UUID
    REFERENCES requirements(id) ON DELETE CASCADE;
ALTER TABLE documents ADD COLUMN requirement_id UUID
    REFERENCES requirements(id) ON DELETE CASCADE;

-- 4. BACKFILL (zero data loss, D-2): one "Default" requirement per EXISTING project,
--    then re-parent that project's user_stories and documents under it.
INSERT INTO requirements (project_id, name, status)
SELECT id, 'Default', 'draft' FROM projects;

UPDATE user_stories us
SET requirement_id = r.id
FROM requirements r
WHERE r.project_id = us.project_id AND r.name = 'Default';

UPDATE documents d
SET requirement_id = r.id
FROM requirements r
WHERE r.project_id = d.project_id AND r.name = 'Default';

-- 5. enforce NOT NULL only AFTER backfill (no orphans by construction —
--    every project got a Default requirement; every child shares the project_id)
ALTER TABLE user_stories ALTER COLUMN requirement_id SET NOT NULL;
ALTER TABLE documents    ALTER COLUMN requirement_id SET NOT NULL;

CREATE INDEX idx_user_stories_requirement_id ON user_stories(requirement_id);
CREATE INDEX idx_documents_requirement_id    ON documents(requirement_id);
