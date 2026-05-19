package repo

import (
	"context"
	"database/sql"
	"fmt"

	"agent-board/internal/domain"
)

// AuditRepository defines the interface for audit trail data access.
type AuditRepository interface {
	// GetTaskAuditTrail retrieves all status audit log entries for a task in chronological order.
	GetTaskAuditTrail(ctx context.Context, taskID string) ([]*domain.StatusAuditLog, error)
	// GetUserStoryAuditTrail retrieves all status audit log entries for a user story in chronological order.
	GetUserStoryAuditTrail(ctx context.Context, userStoryID string) ([]*domain.StatusAuditLog, error)
}

// auditRepo handles database queries for the status_audit_trail table.
type auditRepo struct {
	db *sql.DB
}

// NewAuditRepo creates a new AuditRepository backed by the provided database connection.
func NewAuditRepo(db *sql.DB) AuditRepository {
	return &auditRepo{db: db}
}

const auditTrailQuery = `SELECT id, entity_id, entity_type, from_status, to_status, changed_at FROM status_audit_trail WHERE entity_type = $1 AND entity_id = $2 ORDER BY changed_at ASC`

func (r *auditRepo) getAuditTrail(ctx context.Context, entityType, entityID string) ([]*domain.StatusAuditLog, error) {
	rows, err := r.db.QueryContext(ctx, auditTrailQuery, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit trail: %w", err)
	}
	defer func() { _ = rows.Close() }()

	entries := make([]*domain.StatusAuditLog, 0)
	for rows.Next() {
		entry := &domain.StatusAuditLog{}
		if err := rows.Scan(&entry.ID, &entry.EntityID, &entry.EntityType, &entry.FromStatus, &entry.ToStatus, &entry.ChangedAt); err != nil {
			return nil, fmt.Errorf("failed to scan audit trail entry: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit trail: %w", err)
	}
	return entries, nil
}

// GetTaskAuditTrail retrieves all status audit log entries for the given task ID,
// ordered chronologically (oldest first).
func (r *auditRepo) GetTaskAuditTrail(ctx context.Context, taskID string) ([]*domain.StatusAuditLog, error) {
	return r.getAuditTrail(ctx, "task", taskID)
}

// GetUserStoryAuditTrail retrieves all status audit log entries for the given user story ID,
// ordered chronologically (oldest first).
func (r *auditRepo) GetUserStoryAuditTrail(ctx context.Context, userStoryID string) ([]*domain.StatusAuditLog, error) {
	return r.getAuditTrail(ctx, "user_story", userStoryID)
}
