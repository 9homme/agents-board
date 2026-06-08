package repo

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"agent-board/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UT-015: Create user story in DB
func TestUserStoryRepo_CreateUserStory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewUserStoryRepo(db)
	now := time.Now()
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	u := &domain.UserStory{
		ProjectID:   projectID,
		Title:       "Test User Story",
		Description: "A test description",
		Status:      "draft",
	}

	mock.ExpectQuery(`^INSERT INTO user_stories \(project_id, title, description, status\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id, created_at, updated_at$`).
		WithArgs(u.ProjectID, u.Title, u.Description, u.Status).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("223e4567-e89b-12d3-a456-426614174000", now, now))

	created, err := repo.CreateUserStory(context.Background(), u)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, "223e4567-e89b-12d3-a456-426614174000", created.ID)
	assert.Equal(t, u.ProjectID, created.ProjectID)
	assert.Equal(t, u.Title, created.Title)
	assert.Equal(t, u.Description, created.Description)
	assert.Equal(t, u.Status, created.Status)
	assert.Equal(t, now, created.CreatedAt)
	assert.Equal(t, now, created.UpdatedAt)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-016: Get user story from DB
func TestUserStoryRepo_GetUserStory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewUserStoryRepo(db)
	now := time.Now()
	id := "223e4567-e89b-12d3-a456-426614174000"
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	// Success case
	mock.ExpectQuery(`^SELECT id, project_id, title, description, status, created_at, updated_at FROM user_stories WHERE id = \$1$`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "title", "description", "status", "created_at", "updated_at"}).
			AddRow(id, projectID, "Test User Story", "Desc", "draft", now, now))

	u, err := repo.GetUserStory(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, id, u.ID)

	// Not found case
	mock.ExpectQuery(`^SELECT id, project_id, title, description, status, created_at, updated_at FROM user_stories WHERE id = \$1$`).
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	u, err = repo.GetUserStory(context.Background(), "non-existent")
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, u)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-017: Update user story in DB
func TestUserStoryRepo_UpdateUserStory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewUserStoryRepo(db)
	now := time.Now()
	id := "223e4567-e89b-12d3-a456-426614174000"
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	u := &domain.UserStory{
		ID:          id,
		ProjectID:   projectID,
		Title:       "Updated User Story",
		Description: "Updated desc",
		Status:      "in_progress",
	}

	mock.ExpectQuery(`^UPDATE user_stories SET title = \$1, description = \$2, status = \$3, updated_at = NOW\(\) WHERE id = \$4 RETURNING id, project_id, title, description, status, created_at, updated_at$`).
		WithArgs(u.Title, u.Description, u.Status, u.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "title", "description", "status", "created_at", "updated_at"}).
			AddRow(id, u.ProjectID, u.Title, u.Description, u.Status, now, now))

	updated, err := repo.UpdateUserStory(context.Background(), u)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, id, updated.ID)
	assert.Equal(t, u.Title, updated.Title)
	assert.Equal(t, u.Description, updated.Description)
	assert.Equal(t, u.Status, updated.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-018: Delete user story in DB
func TestUserStoryRepo_DeleteUserStory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewUserStoryRepo(db)
	id := "223e4567-e89b-12d3-a456-426614174000"

	mock.ExpectExec(`^DELETE FROM user_stories WHERE id = \$1$`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DeleteUserStory(context.Background(), id)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// IT-001 (repo layer): UpdateUserStoryStatus transactionally updates status + audit trail
func TestUserStoryRepo_UpdateUserStoryStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)
	now := time.Now()
	id := "223e4567-e89b-12d3-a456-426614174000"
	fromStatus := "draft"
	toStatus := "in_development"

	mock.ExpectBegin()
	mock.ExpectQuery(`^UPDATE user_stories SET status = \$1, updated_at = NOW\(\) WHERE id = \$2 RETURNING id, project_id, title, description, status, created_at, updated_at$`).
		WithArgs(toStatus, id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "title", "description", "status", "created_at", "updated_at"}).
			AddRow(id, "proj-id", "Title", "Desc", toStatus, now, now))
	mock.ExpectExec(`^INSERT INTO status_audit_trail \(entity_id, entity_type, from_status, to_status\) VALUES \(\$1, \$2, \$3, \$4\)$`).
		WithArgs(id, "user_story", fromStatus, toStatus).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	updated, err := r.UpdateUserStoryStatus(context.Background(), id, fromStatus, toStatus)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, id, updated.ID)
	assert.Equal(t, toStatus, updated.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// IT-001 (repo layer): UpdateUserStoryStatus rolls back on audit trail failure
func TestUserStoryRepo_UpdateUserStoryStatus_RollbackOnAuditFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)
	id := "223e4567-e89b-12d3-a456-426614174000"
	fromStatus := "draft"
	toStatus := "in_development"

	mock.ExpectBegin()
	mock.ExpectQuery(`^UPDATE user_stories SET status = \$1, updated_at = NOW\(\) WHERE id = \$2 RETURNING id, project_id, title, description, status, created_at, updated_at$`).
		WithArgs(toStatus, id).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	updated, err := r.UpdateUserStoryStatus(context.Background(), id, fromStatus, toStatus)
	require.Error(t, err)
	assert.Nil(t, updated)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-001: CreateUserStory returns error when DB fails
func TestUserStoryRepo_CreateUserStory_GenericError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)

	mock.ExpectQuery(`INSERT INTO user_stories`).
		WillReturnError(errors.New("db down"))

	result, err := r.CreateUserStory(context.Background(), &domain.UserStory{
		ProjectID:   "proj-id-1",
		Title:       "Test Story",
		Description: "desc",
		Status:      "draft",
	})

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-002: GetUserStory returns error when DB fails (non-NotFound)
func TestUserStoryRepo_GetUserStory_GenericError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)

	mock.ExpectQuery(`SELECT .* FROM user_stories WHERE`).
		WithArgs("us-id-1").
		WillReturnError(errors.New("db down"))

	result, err := r.GetUserStory(context.Background(), "us-id-1")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-003: GetUserStory maps sql.ErrNoRows to ErrNotFound
func TestUserStoryRepo_GetUserStory_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)

	mock.ExpectQuery(`SELECT .* FROM user_stories WHERE`).
		WithArgs("us-id-1").
		WillReturnError(sql.ErrNoRows)

	result, err := r.GetUserStory(context.Background(), "us-id-1")

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrNotFound))
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-004: UpdateUserStory maps sql.ErrNoRows to ErrNotFound
func TestUserStoryRepo_UpdateUserStory_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)

	mock.ExpectQuery(`UPDATE user_stories SET`).
		WillReturnError(sql.ErrNoRows)

	result, err := r.UpdateUserStory(context.Background(), &domain.UserStory{
		ID:          "us-id-1",
		ProjectID:   "proj-id-1",
		Title:       "Title",
		Description: "desc",
		Status:      "draft",
	})

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrNotFound))
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-005: UpdateUserStory returns error when DB fails (non-NotFound)
func TestUserStoryRepo_UpdateUserStory_GenericError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)

	mock.ExpectQuery(`UPDATE user_stories SET`).
		WillReturnError(errors.New("db down"))

	result, err := r.UpdateUserStory(context.Background(), &domain.UserStory{
		ID:          "us-id-1",
		ProjectID:   "proj-id-1",
		Title:       "Title",
		Description: "desc",
		Status:      "draft",
	})

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-006: UpdateUserStoryStatus returns error when BeginTx fails
func TestUserStoryRepo_UpdateUserStoryStatus_BeginTxError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)

	mock.ExpectBegin().WillReturnError(errors.New("begin fail"))

	result, err := r.UpdateUserStoryStatus(context.Background(), "us-id-1", "in_development", "user-1")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-007: UpdateUserStoryStatus maps sql.ErrNoRows to ErrNotFound within transaction
func TestUserStoryRepo_UpdateUserStoryStatus_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE user_stories SET status`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	result, err := r.UpdateUserStoryStatus(context.Background(), "us-id-1", "done", "user-1")

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrNotFound))
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-008: UpdateUserStoryStatus returns error when QueryRowContext fails (non-NotFound)
func TestUserStoryRepo_UpdateUserStoryStatus_UpdateGenericError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE user_stories SET status`).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	result, err := r.UpdateUserStoryStatus(context.Background(), "us-id-1", "done", "user-1")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
	assert.NoError(t, mock.ExpectationsWereMet())
}

// userStoryCols defines the column names returned by user_story SELECT/UPDATE queries.
var userStoryCols = []string{"id", "project_id", "title", "description", "status", "created_at", "updated_at"}

// UT-009: UpdateUserStoryStatus returns error and rolls back when audit insert fails
func TestUserStoryRepo_UpdateUserStoryStatus_AuditInsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE user_stories SET status`).
		WillReturnRows(sqlmock.NewRows(userStoryCols).AddRow(
			"us-id-1", "proj-id-1", "Title", "desc", "done", now, now,
		))
	mock.ExpectExec(`INSERT INTO status_audit_trail`).
		WillReturnError(errors.New("audit fail"))
	mock.ExpectRollback()

	result, err := r.UpdateUserStoryStatus(context.Background(), "us-id-1", "done", "user-1")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-010: UpdateUserStoryStatus returns error when Commit fails
func TestUserStoryRepo_UpdateUserStoryStatus_CommitError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(`UPDATE user_stories SET status`).
		WillReturnRows(sqlmock.NewRows(userStoryCols).AddRow(
			"us-id-1", "proj-id-1", "Title", "desc", "done", now, now,
		))
	mock.ExpectExec(`INSERT INTO status_audit_trail`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit fail"))

	result, err := r.UpdateUserStoryStatus(context.Background(), "us-id-1", "done", "user-1")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-011: ListUserStories returns error when QueryContext fails
func TestUserStoryRepo_ListUserStories_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)

	mock.ExpectQuery(`SELECT .* FROM user_stories WHERE`).
		WillReturnError(errors.New("db down"))

	result, err := r.ListUserStories(context.Background(), "project-id-1")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-012: ListUserStories returns error when Scan fails due to type mismatch
func TestUserStoryRepo_ListUserStories_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)

	// Pass a non-time string for created_at to force Scan failure (time.Time cannot unmarshal arbitrary string).
	cols := []string{"id", "project_id", "title", "description", "status", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT .* FROM user_stories WHERE`).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(
			"us-id-1", "proj-id-1", "Title", "desc", "draft",
			"not-a-time", /* wrong type for time.Time created_at */
			"not-a-time",
		))

	result, err := r.ListUserStories(context.Background(), "project-id-1")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-013: ListUserStories returns error when rows.Err() is set after iteration
func TestUserStoryRepo_ListUserStories_RowsErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)
	now := time.Now()

	cols := []string{"id", "project_id", "title", "description", "status", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT .* FROM user_stories WHERE`).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("us-id-1", "proj-id-1", "Title", "desc", "draft", now, now).
			RowError(0, errors.New("rows err")))

	result, err := r.ListUserStories(context.Background(), "project-id-1")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// IT-001 coverage: ListUserStories returns an empty (non-nil) slice when no rows are returned.
func TestUserStoryRepo_ListUserStories_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)

	cols := []string{"id", "project_id", "title", "description", "status", "created_at", "updated_at"}
	mock.ExpectQuery(`SELECT .* FROM user_stories WHERE`).
		WillReturnRows(sqlmock.NewRows(cols))

	result, err := r.ListUserStories(context.Background(), "project-id-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// US004-UT-001: ListUserStoriesWithTaskCount returns stories with correct taskCount aggregate.
func TestUserStoryRepo_ListUserStoriesWithTaskCount_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)
	now := time.Now()
	projectID := "123e4567-e89b-12d3-a456-426614174000"
	id1 := "11111111-e89b-12d3-a456-426614174000"
	id2 := "22222222-e89b-12d3-a456-426614174000"

	cols := []string{"id", "project_id", "title", "description", "status", "created_at", "updated_at", "task_count"}
	mock.ExpectQuery(`SELECT us.id, us.project_id, us.title, us.description, us.status, us.created_at, us.updated_at, COUNT\(t.id\) AS task_count FROM user_stories us LEFT JOIN tasks t ON t.user_story_id = us.id WHERE us.project_id = \$1 GROUP BY us.id ORDER BY us.created_at DESC`).
		WithArgs(projectID).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(id1, projectID, "US One", "Desc one", "draft", now, now, 2).
			AddRow(id2, projectID, "US Two", "Desc two", "in_progress", now, now, 0))

	stories, err := r.ListUserStoriesWithTaskCount(context.Background(), projectID)
	require.NoError(t, err)
	require.Len(t, stories, 2)
	assert.Equal(t, id1, stories[0].ID)
	assert.Equal(t, 2, stories[0].TaskCount)
	assert.Equal(t, id2, stories[1].ID)
	assert.Equal(t, 0, stories[1].TaskCount)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// US004-UT-002: ListUserStoriesWithTaskCount returns empty slice when no stories exist.
func TestUserStoryRepo_ListUserStoriesWithTaskCount_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	cols := []string{"id", "project_id", "title", "description", "status", "created_at", "updated_at", "task_count"}
	mock.ExpectQuery(`SELECT us.id`).
		WithArgs(projectID).
		WillReturnRows(sqlmock.NewRows(cols))

	stories, err := r.ListUserStoriesWithTaskCount(context.Background(), projectID)
	require.NoError(t, err)
	assert.NotNil(t, stories)
	assert.Len(t, stories, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// US004-UT-003: ListUserStoriesWithTaskCount returns error on query execution failure.
func TestUserStoryRepo_ListUserStoriesWithTaskCount_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)
	projectID := "123e4567-e89b-12d3-a456-426614174000"
	queryErr := errors.New("connection refused")

	mock.ExpectQuery(`SELECT us.id`).
		WithArgs(projectID).
		WillReturnError(queryErr)

	stories, err := r.ListUserStoriesWithTaskCount(context.Background(), projectID)
	assert.Nil(t, stories)
	assert.ErrorIs(t, err, queryErr)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// US004-UT-004: ListUserStoriesWithTaskCount returns error on row scan failure (type mismatch).
func TestUserStoryRepo_ListUserStoriesWithTaskCount_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	// task_count column has wrong type (string instead of int) to force scan failure.
	cols := []string{"id", "project_id", "title", "description", "status", "created_at", "updated_at", "task_count"}
	mock.ExpectQuery(`SELECT us.id`).
		WithArgs(projectID).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("us-id-1", projectID, "Title", "Desc", "draft", "not-a-time", "not-a-time", "not-an-int"))

	stories, err := r.ListUserStoriesWithTaskCount(context.Background(), projectID)
	assert.Nil(t, stories)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// US004-UT-005: ListUserStoriesWithTaskCount returns error on rows iteration failure (rows.Err()).
func TestUserStoryRepo_ListUserStoriesWithTaskCount_RowsErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewUserStoryRepo(db)
	projectID := "123e4567-e89b-12d3-a456-426614174000"
	now := time.Now()

	cols := []string{"id", "project_id", "title", "description", "status", "created_at", "updated_at", "task_count"}
	iterErr := errors.New("rows iteration failed")
	mock.ExpectQuery(`SELECT us.id`).
		WithArgs(projectID).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("us-id-1", projectID, "Title", "Desc", "draft", now, now, 1).
			RowError(0, iterErr))

	stories, err := r.ListUserStoriesWithTaskCount(context.Background(), projectID)
	assert.Nil(t, stories)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-019: List user stories by Project
func TestUserStoryRepo_ListUserStories(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewUserStoryRepo(db)
	now := time.Now()
	projectID := "123e4567-e89b-12d3-a456-426614174000"
	id1 := "11111111-e89b-12d3-a456-426614174000"
	id2 := "22222222-e89b-12d3-a456-426614174000"

	mock.ExpectQuery(`^SELECT id, project_id, title, description, status, created_at, updated_at FROM user_stories WHERE project_id = \$1 ORDER BY created_at DESC$`).
		WithArgs(projectID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "title", "description", "status", "created_at", "updated_at"}).
			AddRow(id1, projectID, "US1", "D1", "draft", now, now).
			AddRow(id2, projectID, "US2", "D2", "in_progress", now, now))

	userStories, err := repo.ListUserStories(context.Background(), projectID)
	assert.NoError(t, err)
	assert.Len(t, userStories, 2)
	assert.Equal(t, id1, userStories[0].ID)
	assert.Equal(t, id2, userStories[1].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}
