package repo

import (
	"context"
	"database/sql"
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
