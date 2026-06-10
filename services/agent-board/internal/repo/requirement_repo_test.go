package repo

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"agent-board/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UT-045-002 — RequirementRepository.ListByProject returns ordered list
func TestRequirementRepo_ListByProject_OrderedASC(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewRequirementRepo(db)
	projectID := "11111111-1111-1111-1111-111111111111"

	t1 := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 6, 9, 11, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"id", "project_id", "name", "description", "status", "created_at", "updated_at"}).
		AddRow("req-001", projectID, "First", "desc1", "draft", t1, t1).
		AddRow("req-002", projectID, "Second", "desc2", "in_progress", t2, t2)

	mock.ExpectQuery(`SELECT id, project_id, name, description, status, created_at, updated_at FROM requirements WHERE project_id = \$1 ORDER BY created_at ASC`).
		WithArgs(projectID).
		WillReturnRows(rows)

	result, err := r.ListByProject(context.Background(), projectID)
	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "req-001", result[0].ID)
	assert.Equal(t, "req-002", result[1].ID)
	assert.True(t, !result[0].CreatedAt.After(result[1].CreatedAt), "items must be ordered by createdAt ASC")
	assert.Equal(t, projectID, result[0].ProjectID)
	assert.Equal(t, "First", result[0].Name)
	assert.Equal(t, "draft", result[0].Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-045-003 — RequirementRepository.ListByProject — DB Query error
func TestRequirementRepo_ListByProject_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewRequirementRepo(db)

	mock.ExpectQuery(`SELECT id, project_id, name, description, status, created_at, updated_at FROM requirements WHERE project_id = \$1 ORDER BY created_at ASC`).
		WithArgs("any-project-id").
		WillReturnError(driver.ErrBadConn)

	result, err := r.ListByProject(context.Background(), "any-project-id")
	require.Error(t, err)
	assert.Nil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-045-004 — RequirementRepository.ListByProject — rows.Scan error
func TestRequirementRepo_ListByProject_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewRequirementRepo(db)
	projectID := "11111111-1111-1111-1111-111111111111"

	// type-incompatible value for created_at causes Scan to fail
	rows := sqlmock.NewRows([]string{"id", "project_id", "name", "description", "status", "created_at", "updated_at"}).
		AddRow("req-001", projectID, "Name", "desc", "draft", "not-a-time", "also-not-a-time")

	mock.ExpectQuery(`SELECT id, project_id, name, description, status, created_at, updated_at FROM requirements WHERE project_id = \$1 ORDER BY created_at ASC`).
		WithArgs(projectID).
		WillReturnRows(rows)

	result, err := r.ListByProject(context.Background(), projectID)
	require.Error(t, err)
	assert.Nil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-045-005 — RequirementRepository.ListByProject — rows.Err() error
func TestRequirementRepo_ListByProject_RowsErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewRequirementRepo(db)
	projectID := "11111111-1111-1111-1111-111111111111"

	t1 := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "project_id", "name", "description", "status", "created_at", "updated_at"}).
		AddRow("req-001", projectID, "Name", "desc", "draft", t1, t1).
		RowError(0, errors.New("iteration error"))

	mock.ExpectQuery(`SELECT id, project_id, name, description, status, created_at, updated_at FROM requirements WHERE project_id = \$1 ORDER BY created_at ASC`).
		WithArgs(projectID).
		WillReturnRows(rows)

	result, err := r.ListByProject(context.Background(), projectID)
	require.Error(t, err)
	assert.Nil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-045-006 — RequirementRepository.ListByProject — project not found returns empty list (not error)
func TestRequirementRepo_ListByProject_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewRequirementRepo(db)
	unknownProjectID := "00000000-0000-0000-0000-000000000000"

	rows := sqlmock.NewRows([]string{"id", "project_id", "name", "description", "status", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT id, project_id, name, description, status, created_at, updated_at FROM requirements WHERE project_id = \$1 ORDER BY created_at ASC`).
		WithArgs(unknownProjectID).
		WillReturnRows(rows)

	result, err := r.ListByProject(context.Background(), unknownProjectID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-045-007 — RequirementRepository.Create happy path
func TestRequirementRepo_Create_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewRequirementRepo(db)
	projectID := "11111111-1111-1111-1111-111111111111"
	now := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)

	req := &domain.Requirement{
		ProjectID:   projectID,
		Name:        "Default",
		Description: "",
		Status:      "draft",
	}

	mock.ExpectQuery(`INSERT INTO requirements \(project_id, name, description, status\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id, created_at, updated_at`).
		WithArgs(projectID, "Default", "", "draft").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("req-generated-001", now, now))

	result, err := r.Create(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "req-generated-001", result.ID)
	assert.Equal(t, projectID, result.ProjectID)
	assert.Equal(t, "Default", result.Name)
	assert.Equal(t, "draft", result.Status)
	assert.Equal(t, now, result.CreatedAt)
	assert.Equal(t, now, result.UpdatedAt)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-045-008 — RequirementRepository.Create — QueryRow.Scan error
func TestRequirementRepo_Create_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewRequirementRepo(db)

	mock.ExpectQuery(`INSERT INTO requirements \(project_id, name, description, status\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id, created_at, updated_at`).
		WithArgs("proj-id", "Name", "", "draft").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(42, "not-a-time", "not-a-time"))

	req := &domain.Requirement{ProjectID: "proj-id", Name: "Name", Status: "draft"}
	result, err := r.Create(context.Background(), req)
	require.Error(t, err)
	assert.Nil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-045-009 — RequirementRepository.Create — project FK violation
func TestRequirementRepo_Create_FKViolation(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewRequirementRepo(db)

	// Postgres FK violation error code 23503
	fkError := errors.New("pq: insert or update on table \"requirements\" violates foreign key constraint \"requirements_project_id_fkey\" (SQLSTATE 23503)")

	mock.ExpectQuery(`INSERT INTO requirements \(project_id, name, description, status\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id, created_at, updated_at`).
		WithArgs("nonexistent-project", "Name", "", "draft").
		WillReturnError(fkError)

	req := &domain.Requirement{ProjectID: "nonexistent-project", Name: "Name", Status: "draft"}
	result, err := r.Create(context.Background(), req)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrProjectNotFound)
	assert.Nil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-045-049 — RequirementRepository.ListByProject — context cancellation
func TestRequirementRepo_ListByProject_ContextCancelled(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewRequirementRepo(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	mock.ExpectQuery(`SELECT id, project_id, name, description, status, created_at, updated_at FROM requirements WHERE project_id = \$1 ORDER BY created_at ASC`).
		WithArgs("any-project-id").
		WillReturnError(context.Canceled)

	result, err := r.ListByProject(ctx, "any-project-id")
	require.Error(t, err)
	assert.Nil(t, result)
}

// UT-045-010 — RequirementRepository.Update happy path
func TestRequirementRepo_Update_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewRequirementRepo(db)
	reqID := "req-001"
	createdAt := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 6, 9, 12, 0, 0, 0, time.UTC)
	newStatus := "in_progress"

	mock.ExpectQuery(`UPDATE requirements SET`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "name", "description", "status", "created_at", "updated_at"}).
			AddRow(reqID, "proj-id", "Name", "desc", newStatus, createdAt, updatedAt))

	patch := RequirementPatch{Status: &newStatus}
	result, err := r.Update(context.Background(), reqID, patch)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "in_progress", result.Status)
	assert.True(t, result.UpdatedAt.After(result.CreatedAt), "updatedAt must be after createdAt")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-045-011 — RequirementRepository.Update — not found (sql.ErrNoRows)
func TestRequirementRepo_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewRequirementRepo(db)

	mock.ExpectQuery(`UPDATE requirements SET`).
		WillReturnError(sql.ErrNoRows)

	patch := RequirementPatch{}
	result, err := r.Update(context.Background(), "nonexistent-id", patch)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-045-012 — RequirementRepository.Update — QueryRow Scan error (non-ErrNoRows)
func TestRequirementRepo_Update_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewRequirementRepo(db)

	mock.ExpectQuery(`UPDATE requirements SET`).
		WillReturnError(errors.New("generic scan error"))

	patch := RequirementPatch{}
	result, err := r.Update(context.Background(), "req-id", patch)
	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound), "error must NOT be ErrNotFound")
	assert.Nil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}
