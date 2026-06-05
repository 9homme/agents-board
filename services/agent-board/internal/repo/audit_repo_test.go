package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IT-004 (repo): GetTaskAuditTrail returns audit entries for a task in chronological order.
func TestAuditRepo_GetTaskAuditTrail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewAuditRepo(db)

	taskID := "223e4567-e89b-12d3-a456-426614174000"
	auditID1 := "aaa00001-e89b-12d3-a456-426614174000"
	auditID2 := "aaa00002-e89b-12d3-a456-426614174000"

	t1 := time.Now().Add(-2 * time.Minute)
	t2 := time.Now().Add(-1 * time.Minute)

	mock.ExpectQuery(`^SELECT id, entity_id, entity_type, from_status, to_status, changed_at FROM status_audit_trail WHERE entity_type = \$1 AND entity_id = \$2 ORDER BY changed_at ASC$`).
		WithArgs("task", taskID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "entity_id", "entity_type", "from_status", "to_status", "changed_at"}).
			AddRow(auditID1, taskID, "task", "pending", "in_progress", t1).
			AddRow(auditID2, taskID, "task", "in_progress", "in_review", t2))

	entries, err := r.GetTaskAuditTrail(context.Background(), taskID)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, auditID1, entries[0].ID)
	assert.Equal(t, taskID, entries[0].EntityID)
	assert.Equal(t, "task", entries[0].EntityType)
	assert.Equal(t, "pending", entries[0].FromStatus)
	assert.Equal(t, "in_progress", entries[0].ToStatus)

	assert.Equal(t, auditID2, entries[1].ID)
	assert.Equal(t, "in_progress", entries[1].FromStatus)
	assert.Equal(t, "in_review", entries[1].ToStatus)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// IT-005 (repo): GetUserStoryAuditTrail returns audit entries for a user story in chronological order.
func TestAuditRepo_GetUserStoryAuditTrail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewAuditRepo(db)

	storyID := "333e4567-e89b-12d3-a456-426614174000"
	auditID1 := "bbb00001-e89b-12d3-a456-426614174000"
	auditID2 := "bbb00002-e89b-12d3-a456-426614174000"

	t1 := time.Now().Add(-2 * time.Minute)
	t2 := time.Now().Add(-1 * time.Minute)

	mock.ExpectQuery(`^SELECT id, entity_id, entity_type, from_status, to_status, changed_at FROM status_audit_trail WHERE entity_type = \$1 AND entity_id = \$2 ORDER BY changed_at ASC$`).
		WithArgs("user_story", storyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "entity_id", "entity_type", "from_status", "to_status", "changed_at"}).
			AddRow(auditID1, storyID, "user_story", "draft", "in_development", t1).
			AddRow(auditID2, storyID, "user_story", "in_development", "in_signoff", t2))

	entries, err := r.GetUserStoryAuditTrail(context.Background(), storyID)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, auditID1, entries[0].ID)
	assert.Equal(t, storyID, entries[0].EntityID)
	assert.Equal(t, "user_story", entries[0].EntityType)
	assert.Equal(t, "draft", entries[0].FromStatus)
	assert.Equal(t, "in_development", entries[0].ToStatus)

	assert.Equal(t, auditID2, entries[1].ID)
	assert.Equal(t, "in_development", entries[1].FromStatus)
	assert.Equal(t, "in_signoff", entries[1].ToStatus)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-001: GetTaskAuditTrail propagates QueryContext errors.
func TestAuditRepo_GetTaskAuditTrail_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewAuditRepo(db)

	mock.ExpectQuery(`SELECT .* FROM status_audit_trail WHERE entity_type`).
		WithArgs("task", "task-id-1").
		WillReturnError(errors.New("db down"))

	result, err := r.GetTaskAuditTrail(context.Background(), "task-id-1")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to query audit trail")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// GetTaskAuditTrail returns empty slice (not nil) when no audit entries exist.
func TestAuditRepo_GetTaskAuditTrail_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewAuditRepo(db)
	taskID := "223e4567-e89b-12d3-a456-426614174000"

	mock.ExpectQuery(`^SELECT id, entity_id, entity_type, from_status, to_status, changed_at FROM status_audit_trail WHERE entity_type = \$1 AND entity_id = \$2 ORDER BY changed_at ASC$`).
		WithArgs("task", taskID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "entity_id", "entity_type", "from_status", "to_status", "changed_at"}))

	entries, err := r.GetTaskAuditTrail(context.Background(), taskID)
	require.NoError(t, err)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}
