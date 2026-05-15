package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agent-board/internal/domain"
	"agent-board/internal/repo"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProjectRepo is a simple mock for repo.ProjectRepository
type mockProjectRepo struct {
	ListProjectsFunc func(ctx context.Context) ([]*domain.Project, error)
	// other methods are unimplemented as they are not needed for these tests
	repo.ProjectRepository
}

func (m *mockProjectRepo) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	if m.ListProjectsFunc != nil {
		return m.ListProjectsFunc(ctx)
	}
	return nil, nil
}

// UT-001 — Successfully load project list
func TestProjectHandler_GetProjects_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	now := time.Now()
	mockRepo := &mockProjectRepo{
		ListProjectsFunc: func(ctx context.Context) ([]*domain.Project, error) {
			return []*domain.Project{
				{
					ID:          "123e4567-e89b-12d3-a456-426614174000",
					Name:        "Test Project",
					Description: "A test project",
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			}, nil
		},
	}

	h := NewProjectHandler(mockRepo)
	err := h.GetProjects(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	projects, ok := res["projects"].([]interface{})
	require.True(t, ok)
	assert.Len(t, projects, 1)

	p := projects[0].(map[string]interface{})
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", p["id"])
	assert.Equal(t, "Test Project", p["name"])
	assert.Equal(t, "A test project", p["description"])
}

// UT-001 — Empty state
func TestProjectHandler_GetProjects_Empty(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := &mockProjectRepo{
		ListProjectsFunc: func(ctx context.Context) ([]*domain.Project, error) {
			return []*domain.Project{}, nil
		},
	}

	h := NewProjectHandler(mockRepo)
	err := h.GetProjects(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	projects, ok := res["projects"].([]interface{})
	require.True(t, ok)
	assert.Len(t, projects, 0)
}

// UT-002 — Error state
func TestProjectHandler_GetProjects_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := &mockProjectRepo{
		ListProjectsFunc: func(ctx context.Context) ([]*domain.Project, error) {
			return nil, errors.New("db connection failed")
		},
	}

	h := NewProjectHandler(mockRepo)
	err := h.GetProjects(c)
	require.NoError(t, err) // Echo handler usually returns the error for centralized handling, but architecture says handler returns 500 JSON directly. We'll verify what the handler returns. Let's design handler to return the JSON response and nil error to echo, or return echo.NewHTTPError? Wait, architecture says response must be `{"code": "INTERNAL_ERROR", "message": "Failed to fetch projects"}` on 500 error.

	// If handler returns `c.JSON(500, ...)`, `err` is nil.
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Failed to fetch projects", res["message"])
}

// IT-001 — Fetch projects end-to-end (DB)
func TestProjectHandler_GetProjects_Integration(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := repo.NewProjectRepo(db)
	h := NewProjectHandler(r)

	e := echo.New()
	e.GET("/api/v1/projects", h.GetProjects)

	now := time.Now()
	mock.ExpectQuery(`^SELECT id, name, description, created_at, updated_at FROM projects ORDER BY created_at DESC$`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow("11111111-e89b-12d3-a456-426614174000", "P1", "D1", now, now).
			AddRow("22222222-e89b-12d3-a456-426614174000", "P2", "D2", now, now))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	require.NoError(t, err)

	projects, ok := res["projects"].([]interface{})
	require.True(t, ok)
	assert.Len(t, projects, 2)

	p1 := projects[0].(map[string]interface{})
	assert.Equal(t, "11111111-e89b-12d3-a456-426614174000", p1["id"])
	assert.Equal(t, "P1", p1["name"])

	assert.NoError(t, mock.ExpectationsWereMet())
}
