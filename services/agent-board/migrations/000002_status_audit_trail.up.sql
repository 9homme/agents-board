CREATE TABLE status_audit_trail (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_id UUID NOT NULL,
    entity_type VARCHAR(50) NOT NULL, -- 'task' or 'user_story'
    from_status VARCHAR(50) NOT NULL,
    to_status VARCHAR(50) NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for efficient chronological querying of audit trails per entity
CREATE INDEX idx_status_audit_trail_entity ON status_audit_trail(entity_type, entity_id, changed_at ASC);
