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

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockUserStoryListRepo is a mock implementing only ListUserStoriesWithTaskCount.
type mockUserStoryListRepo struct {
	repo.UserStoryRepository
	ListWithTaskCountFunc func(ctx context.Context, projectID string) ([]*repo.UserStoryWithCount, error)
}

func (m *mockUserStoryListRepo) ListUserStoriesWithTaskCount(ctx context.Context, projectID string) ([]*repo.UserStoryWithCount, error) {
	if m.ListWithTaskCountFunc != nil {
		return m.ListWithTaskCountFunc(ctx, projectID)
	}
	return []*repo.UserStoryWithCount{}, nil
}

// mockProjectRepoForUSHandler is a mock for repo.ProjectRepository used in user story handler tests.
type mockProjectRepoForUSHandler struct {
	repo.ProjectRepository
	GetProjectFunc func(ctx context.Context, id string) (*domain.Project, error)
}

func (m *mockProjectRepoForUSHandler) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	if m.GetProjectFunc != nil {
		return m.GetProjectFunc(ctx, id)
	}
	return nil, repo.ErrNotFound
}

// IT-001 — GET /api/v1/projects/{id}/user-stories: 200 with list and task counts
func TestUserStoryHandler_GetProjectUserStories_200(t *testing.T) {
	e := echo.New()
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	fixedTime, err := time.Parse(time.RFC3339, "2026-06-01T10:00:00Z")
	require.NoError(t, err)

	projectRepo := &mockProjectRepoForUSHandler{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return &domain.Project{ID: id, Name: "Test Project", Description: ""}, nil
		},
	}
	usRepo := &mockUserStoryListRepo{
		ListWithTaskCountFunc: func(ctx context.Context, pid string) ([]*repo.UserStoryWithCount, error) {
			return []*repo.UserStoryWithCount{
				{
					UserStory: domain.UserStory{
						ID:          "aaaaaaaa-0000-0000-0000-000000000001",
						ProjectID:   pid,
						Title:       "Story One",
						Description: "Description one",
						Status:      "draft",
						CreatedAt:   fixedTime,
						UpdatedAt:   fixedTime,
					},
					TaskCount: 3,
				},
				{
					UserStory: domain.UserStory{
						ID:          "aaaaaaaa-0000-0000-0000-000000000002",
						ProjectID:   pid,
						Title:       "Story Two",
						Description: "Description two",
						Status:      "in_progress",
						CreatedAt:   fixedTime,
						UpdatedAt:   fixedTime,
					},
					TaskCount: 0,
				},
			}, nil
		},
	}

	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	e.GET("/api/v1/projects/:id/user-stories", h.GetProjectUserStories)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects/"+projectID+"/user-stories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	stories, ok := res["userStories"].([]interface{})
	require.True(t, ok, "response must have userStories array")
	require.Len(t, stories, 2)

	s1 := stories[0].(map[string]interface{})
	assert.Equal(t, "aaaaaaaa-0000-0000-0000-000000000001", s1["id"])
	assert.Equal(t, projectID, s1["projectId"])
	assert.Equal(t, "Story One", s1["title"])
	assert.Equal(t, "Description one", s1["description"])
	assert.Equal(t, "draft", s1["status"])
	assert.EqualValues(t, 3, s1["taskCount"])
	assert.Equal(t, "2026-06-01T10:00:00Z", s1["createdAt"])
	assert.Equal(t, "2026-06-01T10:00:00Z", s1["updatedAt"])

	s2 := stories[1].(map[string]interface{})
	assert.Equal(t, "aaaaaaaa-0000-0000-0000-000000000002", s2["id"])
	assert.EqualValues(t, 0, s2["taskCount"])
}

// IT-002 — GET /api/v1/projects/{id}/user-stories: 404 for missing project
func TestUserStoryHandler_GetProjectUserStories_404_MissingProject(t *testing.T) {
	e := echo.New()
	missingID := "00000000-0000-0000-0000-000000000000"

	projectRepo := &mockProjectRepoForUSHandler{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return nil, repo.ErrNotFound
		},
	}
	usRepo := &mockUserStoryListRepo{}

	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	e.GET("/api/v1/projects/:id/user-stories", h.GetProjectUserStories)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects/"+missingID+"/user-stories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Project not found", res["message"])
}

// IT-003 — GET /api/v1/projects/{id}/user-stories: 404 for invalid project ID format
func TestUserStoryHandler_GetProjectUserStories_404_InvalidUUID(t *testing.T) {
	e := echo.New()

	// projectRepo returns ErrNotFound for any ID that's not a valid uuid (mimicking DB behaviour)
	projectRepo := &mockProjectRepoForUSHandler{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return nil, repo.ErrNotFound
		},
	}
	usRepo := &mockUserStoryListRepo{}

	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	e.GET("/api/v1/projects/:id/user-stories", h.GetProjectUserStories)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects/invalid-uuid/user-stories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Project not found", res["message"])
}

// IT-004 — GET /api/v1/projects/{id}/user-stories: 500 on repository failure
func TestUserStoryHandler_GetProjectUserStories_500_RepoFailure(t *testing.T) {
	e := echo.New()
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	projectRepo := &mockProjectRepoForUSHandler{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return &domain.Project{ID: id, Name: "Test Project", Description: ""}, nil
		},
	}
	usRepo := &mockUserStoryListRepo{
		ListWithTaskCountFunc: func(ctx context.Context, pid string) ([]*repo.UserStoryWithCount, error) {
			return nil, errors.New("db connection refused")
		},
	}

	h := handler.NewUserStoryHandler(usRepo, projectRepo)
	e.GET("/api/v1/projects/:id/user-stories", h.GetProjectUserStories)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects/"+projectID+"/user-stories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Failed to fetch user stories", res["message"])
}
