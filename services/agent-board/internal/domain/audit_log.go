package domain

import "time"

// StatusAuditLog represents an entry in the status audit trail.
type StatusAuditLog struct {
	ID         string    `json:"id"`
	EntityID   string    `json:"entityId"`
	EntityType string    `json:"entityType"` // "task" or "user_story"
	FromStatus string    `json:"fromStatus"`
	ToStatus   string    `json:"toStatus"`
	ChangedAt  time.Time `json:"changedAt"`
}
