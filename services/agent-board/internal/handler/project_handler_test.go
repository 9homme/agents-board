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
	GetProjectFunc   func(ctx context.Context, id string) (*domain.Project, error)
	// other methods are unimplemented as they are not needed for these tests
	repo.ProjectRepository
}

func (m *mockProjectRepo) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	if m.ListProjectsFunc != nil {
		return m.ListProjectsFunc(ctx)
	}
	return nil, nil
}

func (m *mockProjectRepo) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	if m.GetProjectFunc != nil {
		return m.GetProjectFunc(ctx, id)
	}
	return nil, repo.ErrNotFound
}

// UT-001 — Successfully load project list
func TestProjectHandler_GetProjects_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects", nil)
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
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects", nil)
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
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects", nil)
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

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects", nil)
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

// UT-US001-001 — GetProject handler: 200 happy path
func TestProjectHandler_GetProject_200(t *testing.T) {
	fixedTime, err := time.Parse(time.RFC3339, "2026-05-20T10:00:00Z")
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects/123e4567-e89b-12d3-a456-426614174000", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")

	mockRepo := &mockProjectRepo{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return &domain.Project{
				ID:          "123e4567-e89b-12d3-a456-426614174000",
				Name:        "E-commerce Website",
				Description: "A new online store for electronics",
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			}, nil
		},
	}

	h := NewProjectHandler(mockRepo)
	err = h.GetProject(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", res["id"])
	assert.Equal(t, "E-commerce Website", res["name"])
	assert.Equal(t, "A new online store for electronics", res["description"])
	assert.Equal(t, "2026-05-20T10:00:00Z", res["createdAt"])
	assert.Equal(t, "2026-05-20T10:00:00Z", res["updatedAt"])

	// Exactly five fields — no extra fields, not wrapped.
	assert.Len(t, res, 5)

	// Ensure no "project" wrapper key
	_, hasWrapper := res["project"]
	assert.False(t, hasWrapper, "response must be a bare object, not wrapped in {\"project\":...}")
}

// UT-US001-001 edge case — empty description serializes as "" not null
func TestProjectHandler_GetProject_EmptyDescription(t *testing.T) {
	fixedTime, err := time.Parse(time.RFC3339, "2026-05-20T10:00:00Z")
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects/123e4567-e89b-12d3-a456-426614174000", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("123e4567-e89b-12d3-a456-426614174000")

	mockRepo := &mockProjectRepo{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return &domain.Project{
				ID:          "123e4567-e89b-12d3-a456-426614174000",
				Name:        "E-commerce Website",
				Description: "",
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			}, nil
		},
	}

	h := NewProjectHandler(mockRepo)
	err = h.GetProject(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Raw JSON must contain `"description":""` not `"description":null`
	body := rec.Body.String()
	assert.Contains(t, body, `"description":""`)
}

// UT-US001-002 — GetProject handler: 404 not found
func TestProjectHandler_GetProject_404(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects/no-such-id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("no-such-id")

	mockRepo := &mockProjectRepo{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return nil, repo.ErrNotFound
		},
	}

	h := NewProjectHandler(mockRepo)
	err := h.GetProject(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Project not found", res["message"])
	assert.Len(t, res, 2)
}

// UT-US001-003 — GetProject handler: 500 internal error
func TestProjectHandler_GetProject_500(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects/any-id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("any-id")

	mockRepo := &mockProjectRepo{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return nil, errors.New("db connection refused")
		},
	}

	h := NewProjectHandler(mockRepo)
	err := h.GetProject(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Failed to fetch project", res["message"])
}

// IT-US001-001 — GET /api/v1/projects/{id} integration: found
func TestProjectHandler_GetProject_Integration_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	fixedTime, err := time.Parse(time.RFC3339, "2026-05-20T10:00:00Z")
	require.NoError(t, err)

	r := repo.NewProjectRepo(db)
	h := NewProjectHandler(r)

	e := echo.New()
	e.GET("/api/v1/projects/:id", h.GetProject)

	mock.ExpectQuery(`SELECT id, name, description, created_at, updated_at FROM projects WHERE id = \$1`).
		WithArgs("123e4567-e89b-12d3-a456-426614174000").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow("123e4567-e89b-12d3-a456-426614174000", "Integration Test Project", "desc", fixedTime, fixedTime))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects/123e4567-e89b-12d3-a456-426614174000", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", res["id"])
	assert.Equal(t, "Integration Test Project", res["name"])
	assert.Equal(t, "2026-05-20T10:00:00Z", res["createdAt"])
	assert.Equal(t, "2026-05-20T10:00:00Z", res["updatedAt"])
	assert.Len(t, res, 5)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// IT-US001-002 — GET /api/v1/projects/{id} integration: not found
func TestProjectHandler_GetProject_Integration_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := repo.NewProjectRepo(db)
	h := NewProjectHandler(r)

	e := echo.New()
	e.GET("/api/v1/projects/:id", h.GetProject)

	mock.ExpectQuery(`SELECT id, name, description, created_at, updated_at FROM projects WHERE id = \$1`).
		WithArgs("00000000-0000-0000-0000-000000000000").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects/00000000-0000-0000-0000-000000000000", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Project not found", res["message"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

// IT-US001-003 — Route registration smoke test
func TestProjectHandler_RouteRegistration(t *testing.T) {
	e := echo.New()

	mockRepo := &mockProjectRepo{}
	projectHandler := NewProjectHandler(mockRepo)
	e.GET("/api/v1/projects", projectHandler.GetProjects)
	e.GET("/api/v1/projects/:id", projectHandler.GetProject)

	routes := e.Routes()
	found := false
	for _, r := range routes {
		if r.Method == http.MethodGet && r.Path == "/api/v1/projects/:id" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected route GET /api/v1/projects/:id to be registered")
}
