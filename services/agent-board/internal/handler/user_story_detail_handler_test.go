package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agent-board/internal/domain"
	"agent-board/internal/handler"
	"agent-board/internal/repo"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

// mockUserStoryDetailRepo is a mock implementing GetUserStory for unit tests.
type mockUserStoryDetailRepo struct {
	repo.UserStoryRepository
	GetUserStoryFunc func(ctx context.Context, id string) (*domain.UserStory, error)
}

func (m *mockUserStoryDetailRepo) GetUserStory(ctx context.Context, id string) (*domain.UserStory, error) {
	if m.GetUserStoryFunc != nil {
		return m.GetUserStoryFunc(ctx, id)
	}
	return nil, repo.ErrNotFound
}

// mockTaskRepo is a mock implementing ListTasks for unit tests.
type mockTaskRepo struct {
	repo.TaskRepository
	ListTasksFunc func(ctx context.Context, userStoryID string) ([]*domain.Task, error)
}

func (m *mockTaskRepo) ListTasks(ctx context.Context, userStoryID string) ([]*domain.Task, error) {
	if m.ListTasksFunc != nil {
		return m.ListTasksFunc(ctx, userStoryID)
	}
	return []*domain.Task{}, nil
}

// ---------------------------------------------------------------------------
// Unit tests (mock repo — no DB)
// ---------------------------------------------------------------------------

// UT-001 — Handler GET story maps ErrNotFound to 404
func TestUserStoryDetailHandler_GetUserStory_UT001_ErrNotFound(t *testing.T) {
	e := echo.New()

	usRepo := &mockUserStoryDetailRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return nil, repo.ErrNotFound
		},
	}
	taskRepo := &mockTaskRepo{}
	projectRepo := &mockProjectRepoForUSHandler{}

	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	h.SetTaskRepo(taskRepo)

	e.GET("/api/v1/user-stories/:id", h.GetUserStory)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/user-stories/00000000-0000-0000-0000-000000000000", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "User story not found", res["message"])
}

// UT-002 — Handler GET story maps generic error to 500
func TestUserStoryDetailHandler_GetUserStory_UT002_GenericError(t *testing.T) {
	e := echo.New()

	usRepo := &mockUserStoryDetailRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return nil, errors.New("db down")
		},
	}
	taskRepo := &mockTaskRepo{}
	projectRepo := &mockProjectRepoForUSHandler{}

	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	h.SetTaskRepo(taskRepo)

	e.GET("/api/v1/user-stories/:id", h.GetUserStory)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/user-stories/any-id", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Internal server error", res["message"])
}

// UT-003 — Handler GET tasks maps ErrNotFound to 404
func TestUserStoryDetailHandler_GetUserStoryTasks_UT003_ErrNotFound(t *testing.T) {
	e := echo.New()

	usRepo := &mockUserStoryDetailRepo{}
	taskRepo := &mockTaskRepo{
		ListTasksFunc: func(ctx context.Context, userStoryID string) ([]*domain.Task, error) {
			return nil, repo.ErrNotFound
		},
	}
	projectRepo := &mockProjectRepoForUSHandler{}

	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	h.SetTaskRepo(taskRepo)

	e.GET("/api/v1/user-stories/:id/tasks", h.GetUserStoryTasks)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/user-stories/00000000-0000-0000-0000-000000000000/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "User story not found", res["message"])
}

// UT-004 — Handler GET tasks maps generic error to 500
func TestUserStoryDetailHandler_GetUserStoryTasks_UT004_GenericError(t *testing.T) {
	e := echo.New()

	usRepo := &mockUserStoryDetailRepo{}
	taskRepo := &mockTaskRepo{
		ListTasksFunc: func(ctx context.Context, userStoryID string) ([]*domain.Task, error) {
			return nil, errors.New("db down")
		},
	}
	projectRepo := &mockProjectRepoForUSHandler{}

	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	h.SetTaskRepo(taskRepo)

	e.GET("/api/v1/user-stories/:id/tasks", h.GetUserStoryTasks)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/user-stories/any-id/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Internal server error", res["message"])
}

// ---------------------------------------------------------------------------
// Integration tests (sqlmock — handler ↔ repo ↔ mock DB)
// ---------------------------------------------------------------------------

// IT-001 — GET /api/v1/user-stories/{id}: 200 with correct shape (no taskCount)
func TestUserStoryDetailHandler_GetUserStory_IT001_200(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	fixedTime, err := time.Parse(time.RFC3339, "2026-06-01T10:00:00Z")
	require.NoError(t, err)

	storyID := "aaaaaaaa-0000-0000-0000-000000000001"
	projectID := "bbbbbbbb-0000-0000-0000-000000000001"

	mock.ExpectQuery(`SELECT id, project_id, requirement_id, title, description, status, created_at, updated_at FROM user_stories WHERE id = \$1`).
		WithArgs(storyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "requirement_id", "title", "description", "status", "created_at", "updated_at"}).
			AddRow(storyID, projectID, "", "Story Title", "Story description", "draft", fixedTime, fixedTime))

	usRepo := repo.NewUserStoryRepo(db)
	projectRepo := &mockProjectRepoForUSHandler{}
	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	// taskRepo not needed for this endpoint
	h.SetTaskRepo(repo.NewTaskRepo(db))

	e := echo.New()
	e.GET("/api/v1/user-stories/:id", h.GetUserStory)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/user-stories/"+storyID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	assert.Equal(t, storyID, res["id"])
	assert.Equal(t, projectID, res["projectId"])
	assert.Equal(t, "Story Title", res["title"])
	assert.Equal(t, "Story description", res["description"])
	assert.Equal(t, "draft", res["status"])
	assert.Equal(t, "2026-06-01T10:00:00Z", res["createdAt"])
	assert.Equal(t, "2026-06-01T10:00:00Z", res["updatedAt"])

	// taskCount must NOT be present
	_, hasTaskCount := res["taskCount"]
	assert.False(t, hasTaskCount, "response must NOT contain taskCount")

	// Exactly 8 fields: id, projectId, requirementId, title, description, status, createdAt, updatedAt
	assert.Len(t, res, 8)
	assert.Equal(t, "", res["requirementId"], "requirementId must be present (empty string when not set)")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// IT-002 — GET /api/v1/user-stories/{id}: 404 when story not found
func TestUserStoryDetailHandler_GetUserStory_IT002_404(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	missingID := "00000000-0000-0000-0000-000000000000"

	mock.ExpectQuery(`SELECT id, project_id, requirement_id, title, description, status, created_at, updated_at FROM user_stories WHERE id = \$1`).
		WithArgs(missingID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "requirement_id", "title", "description", "status", "created_at", "updated_at"}))

	usRepo := repo.NewUserStoryRepo(db)
	projectRepo := &mockProjectRepoForUSHandler{}
	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	h.SetTaskRepo(repo.NewTaskRepo(db))

	e := echo.New()
	e.GET("/api/v1/user-stories/:id", h.GetUserStory)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/user-stories/"+missingID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "User story not found", res["message"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

// IT-003 — GET /api/v1/user-stories/{id}/tasks: 200 with task list ordered by created_at DESC
func TestUserStoryDetailHandler_GetUserStoryTasks_IT003_200WithList(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	storyID := "aaaaaaaa-0000-0000-0000-000000000001"
	task1ID := "cccccccc-0000-0000-0000-000000000001"
	task2ID := "cccccccc-0000-0000-0000-000000000002"

	t1, err := time.Parse(time.RFC3339, "2026-06-02T10:00:00Z")
	require.NoError(t, err)
	t2, err := time.Parse(time.RFC3339, "2026-06-01T10:00:00Z")
	require.NoError(t, err)

	mock.ExpectQuery(`SELECT id, user_story_id, title, description, status, created_at, updated_at FROM tasks WHERE user_story_id = \$1 ORDER BY created_at DESC`).
		WithArgs(storyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_story_id", "title", "description", "status", "created_at", "updated_at"}).
			AddRow(task1ID, storyID, "Task One", "Description 1", "todo", t1, t1).
			AddRow(task2ID, storyID, "Task Two", "Description 2", "done", t2, t2))

	taskRepo := repo.NewTaskRepo(db)
	usRepo := &mockUserStoryDetailRepo{}
	projectRepo := &mockProjectRepoForUSHandler{}
	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/user-stories/:id/tasks", h.GetUserStoryTasks)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/user-stories/"+storyID+"/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	tasks, ok := res["tasks"].([]interface{})
	require.True(t, ok, "response must have tasks array")
	require.Len(t, tasks, 2)

	tk1 := tasks[0].(map[string]interface{})
	assert.Equal(t, task1ID, tk1["id"])
	assert.Equal(t, storyID, tk1["userStoryId"])
	assert.Equal(t, "Task One", tk1["title"])
	assert.Equal(t, "Description 1", tk1["description"])
	assert.Equal(t, "todo", tk1["status"])
	assert.Equal(t, "2026-06-02T10:00:00Z", tk1["createdAt"])
	assert.Equal(t, "2026-06-02T10:00:00Z", tk1["updatedAt"])

	tk2 := tasks[1].(map[string]interface{})
	assert.Equal(t, task2ID, tk2["id"])
	assert.Equal(t, "done", tk2["status"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

// IT-004 — GET /api/v1/user-stories/{id}/tasks: 200 with empty list (never null)
// Uses sqlmock: ListTasks returns empty rows, GetUserStory confirms story exists.
func TestUserStoryDetailHandler_GetUserStoryTasks_IT004_EmptyList(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	storyID := "aaaaaaaa-0000-0000-0000-000000000001"
	projectID := "bbbbbbbb-0000-0000-0000-000000000001"
	fixedTime, err := time.Parse(time.RFC3339, "2026-06-01T10:00:00Z")
	require.NoError(t, err)

	// ListTasks returns empty rows (story exists but has no tasks).
	mock.ExpectQuery(`SELECT id, user_story_id, title, description, status, created_at, updated_at FROM tasks WHERE user_story_id = \$1 ORDER BY created_at DESC`).
		WithArgs(storyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_story_id", "title", "description", "status", "created_at", "updated_at"}))

	// Handler must verify story exists when task list is empty.
	mock.ExpectQuery(`SELECT id, project_id, requirement_id, title, description, status, created_at, updated_at FROM user_stories WHERE id = \$1`).
		WithArgs(storyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "requirement_id", "title", "description", "status", "created_at", "updated_at"}).
			AddRow(storyID, projectID, "", "Empty Story", "No tasks here", "draft", fixedTime, fixedTime))

	taskRepo := repo.NewTaskRepo(db)
	usRepo := repo.NewUserStoryRepo(db)
	projectRepo := &mockProjectRepoForUSHandler{}
	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/user-stories/:id/tasks", h.GetUserStoryTasks)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/user-stories/"+storyID+"/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Body must be {"tasks":[]} not {"tasks":null}
	body := rec.Body.String()
	assert.Contains(t, body, `"tasks":[]`)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	tasks, ok := res["tasks"].([]interface{})
	require.True(t, ok, "tasks must be an array")
	assert.Len(t, tasks, 0)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// IT-005 — GET /api/v1/user-stories/{id}/tasks: 404 when story missing.
// Uses sqlmock: ListTasks returns empty rows (no tasks for the ID), then
// GetUserStory returns no rows (story does not exist) → 404.
func TestUserStoryDetailHandler_GetUserStoryTasks_IT005_404StoryMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	missingID := "00000000-0000-0000-0000-000000000000"

	// ListTasks finds no tasks (the story ID matches nothing).
	mock.ExpectQuery(`SELECT id, user_story_id, title, description, status, created_at, updated_at FROM tasks WHERE user_story_id = \$1 ORDER BY created_at DESC`).
		WithArgs(missingID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_story_id", "title", "description", "status", "created_at", "updated_at"}))

	// GetUserStory also finds nothing → ErrNotFound → 404.
	mock.ExpectQuery(`SELECT id, project_id, requirement_id, title, description, status, created_at, updated_at FROM user_stories WHERE id = \$1`).
		WithArgs(missingID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "requirement_id", "title", "description", "status", "created_at", "updated_at"}))

	taskRepo := repo.NewTaskRepo(db)
	usRepo := repo.NewUserStoryRepo(db)
	projectRepo := &mockProjectRepoForUSHandler{}
	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/user-stories/:id/tasks", h.GetUserStoryTasks)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/user-stories/"+missingID+"/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "User story not found", res["message"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

// IT-005b — GET /api/v1/user-stories/{id}/tasks: 500 when story-existence check fails with generic error
func TestUserStoryDetailHandler_GetUserStoryTasks_IT005b_500StoryCheckError(t *testing.T) {
	e := echo.New()

	// ListTasks returns empty; GetUserStory returns a generic (non-NotFound) error.
	taskRepo := &mockTaskRepo{
		ListTasksFunc: func(ctx context.Context, userStoryID string) ([]*domain.Task, error) {
			return []*domain.Task{}, nil
		},
	}
	usRepo := &mockUserStoryDetailRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return nil, errors.New("db down")
		},
	}
	projectRepo := &mockProjectRepoForUSHandler{}
	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	h.SetTaskRepo(taskRepo)

	e.GET("/api/v1/user-stories/:id/tasks", h.GetUserStoryTasks)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/user-stories/any-id/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Internal server error", res["message"])
}

// IT-006 — GET /api/v1/user-stories/{id}: 500 on repo error
func TestUserStoryDetailHandler_GetUserStory_IT006_500RepoError(t *testing.T) {
	e := echo.New()

	usRepo := &mockUserStoryDetailRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return nil, errors.New("db connection refused")
		},
	}
	taskRepo := &mockTaskRepo{}
	projectRepo := &mockProjectRepoForUSHandler{}
	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	h.SetTaskRepo(taskRepo)

	e.GET("/api/v1/user-stories/:id", h.GetUserStory)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/user-stories/any-id", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Internal server error", res["message"])
}

// IT-007 — GET /api/v1/user-stories/{id}/tasks: 500 on repo error
func TestUserStoryDetailHandler_GetUserStoryTasks_IT007_500RepoError(t *testing.T) {
	e := echo.New()

	taskRepo := &mockTaskRepo{
		ListTasksFunc: func(ctx context.Context, userStoryID string) ([]*domain.Task, error) {
			return nil, errors.New("db connection refused")
		},
	}
	usRepo := &mockUserStoryDetailRepo{}
	projectRepo := &mockProjectRepoForUSHandler{}
	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	h.SetTaskRepo(taskRepo)

	e.GET("/api/v1/user-stories/:id/tasks", h.GetUserStoryTasks)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/user-stories/any-id/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Internal server error", res["message"])
}
