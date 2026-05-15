package repo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"agent-board-mcp/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UT-020: Create task in DB
func TestTaskRepo_CreateTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTaskRepo(db)
	now := time.Now()
	userStoryID := "123e4567-e89b-12d3-a456-426614174000"

	task := &domain.Task{
		UserStoryID: userStoryID,
		Title:       "Test Task",
		Description: "A test task description",
		Status:      "pending",
	}

	mock.ExpectQuery(`^INSERT INTO tasks \(user_story_id, title, description, status\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id, created_at, updated_at$`).
		WithArgs(task.UserStoryID, task.Title, task.Description, task.Status).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("223e4567-e89b-12d3-a456-426614174000", now, now))

	created, err := repo.CreateTask(context.Background(), task)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, "223e4567-e89b-12d3-a456-426614174000", created.ID)
	assert.Equal(t, task.UserStoryID, created.UserStoryID)
	assert.Equal(t, task.Title, created.Title)
	assert.Equal(t, task.Description, created.Description)
	assert.Equal(t, task.Status, created.Status)
	assert.Equal(t, now, created.CreatedAt)
	assert.Equal(t, now, created.UpdatedAt)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-021: Get task from DB
func TestTaskRepo_GetTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTaskRepo(db)
	now := time.Now()
	id := "223e4567-e89b-12d3-a456-426614174000"
	userStoryID := "123e4567-e89b-12d3-a456-426614174000"

	// Success case
	mock.ExpectQuery(`^SELECT id, user_story_id, title, description, status, created_at, updated_at FROM tasks WHERE id = \$1$`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_story_id", "title", "description", "status", "created_at", "updated_at"}).
			AddRow(id, userStoryID, "Test Task", "Description", "pending", now, now))

	task, err := repo.GetTask(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, id, task.ID)

	// Not found case
	mock.ExpectQuery(`^SELECT id, user_story_id, title, description, status, created_at, updated_at FROM tasks WHERE id = \$1$`).
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	task, err = repo.GetTask(context.Background(), "non-existent")
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, task)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-022: Update task in DB
func TestTaskRepo_UpdateTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTaskRepo(db)
	now := time.Now()
	id := "223e4567-e89b-12d3-a456-426614174000"
	userStoryID := "123e4567-e89b-12d3-a456-426614174000"

	task := &domain.Task{
		ID:          id,
		UserStoryID: userStoryID,
		Title:       "Updated Task",
		Description: "Updated desc",
		Status:      "in_progress",
	}

	mock.ExpectQuery(`^UPDATE tasks SET title = \$1, description = \$2, status = \$3, updated_at = NOW\(\) WHERE id = \$4 RETURNING id, user_story_id, title, description, status, created_at, updated_at$`).
		WithArgs(task.Title, task.Description, task.Status, task.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_story_id", "title", "description", "status", "created_at", "updated_at"}).
			AddRow(id, task.UserStoryID, task.Title, task.Description, task.Status, now, now))

	updated, err := repo.UpdateTask(context.Background(), task)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, id, updated.ID)
	assert.Equal(t, task.Title, updated.Title)
	assert.Equal(t, task.Description, updated.Description)
	assert.Equal(t, task.Status, updated.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-023: Delete task in DB
func TestTaskRepo_DeleteTask(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTaskRepo(db)
	id := "223e4567-e89b-12d3-a456-426614174000"

	mock.ExpectExec(`^DELETE FROM tasks WHERE id = \$1$`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DeleteTask(context.Background(), id)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-024: List tasks by User Story
func TestTaskRepo_ListTasks(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewTaskRepo(db)
	now := time.Now()
	userStoryID := "123e4567-e89b-12d3-a456-426614174000"
	id1 := "11111111-e89b-12d3-a456-426614174000"
	id2 := "22222222-e89b-12d3-a456-426614174000"

	mock.ExpectQuery(`^SELECT id, user_story_id, title, description, status, created_at, updated_at FROM tasks WHERE user_story_id = \$1 ORDER BY created_at DESC$`).
		WithArgs(userStoryID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_story_id", "title", "description", "status", "created_at", "updated_at"}).
			AddRow(id1, userStoryID, "T1", "D1", "pending", now, now).
			AddRow(id2, userStoryID, "T2", "D2", "in_progress", now, now))

	tasks, err := repo.ListTasks(context.Background(), userStoryID)
	assert.NoError(t, err)
	assert.Len(t, tasks, 2)
	assert.Equal(t, id1, tasks[0].ID)
	assert.Equal(t, id2, tasks[1].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}
