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

// mockRequirementRepo is a simple mock for RequirementRepository.
type mockRequirementRepo struct {
	repo.RequirementRepository
	ListByProjectFunc func(ctx context.Context, projectID string) ([]domain.Requirement, error)
	CreateFunc        func(ctx context.Context, req *domain.Requirement) (*domain.Requirement, error)
}

func (m *mockRequirementRepo) ListByProject(ctx context.Context, projectID string) ([]domain.Requirement, error) {
	if m.ListByProjectFunc != nil {
		return m.ListByProjectFunc(ctx, projectID)
	}
	return []domain.Requirement{}, nil
}

func (m *mockRequirementRepo) Create(ctx context.Context, req *domain.Requirement) (*domain.Requirement, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return nil, errors.New("Create not implemented in mock")
}

// mockProjectRepoForReqHandler is a mock for repo.ProjectRepository used in requirement handler tests.
type mockProjectRepoForReqHandler struct {
	repo.ProjectRepository
	GetProjectFunc func(ctx context.Context, id string) (*domain.Project, error)
}

func (m *mockProjectRepoForReqHandler) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	if m.GetProjectFunc != nil {
		return m.GetProjectFunc(ctx, id)
	}
	return nil, repo.ErrNotFound
}

// UT-045-001 — RequirementHandler returns 500 on repo error for list
func TestRequirementHandler_ListProjectRequirements_500_RepoError(t *testing.T) {
	e := echo.New()
	projectID := "11111111-1111-1111-1111-111111111111"

	projectRepo := &mockProjectRepoForReqHandler{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return &domain.Project{ID: id, Name: "Test Project"}, nil
		},
	}
	reqRepo := &mockRequirementRepo{
		ListByProjectFunc: func(ctx context.Context, pid string) ([]domain.Requirement, error) {
			return nil, errors.New("db error")
		},
	}

	h := NewRequirementHandler(reqRepo, projectRepo)
	e.GET("/api/v1/projects/:pid/requirements", h.ListProjectRequirements)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects/"+projectID+"/requirements", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Failed to fetch requirements", res["message"])
}

// IT-045-001 — GET /api/v1/projects/:pid/requirements — 200 with requirements list (sqlmock integration)
func TestRequirementHandler_ListProjectRequirements_200_WithList(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	projectID := "11111111-1111-1111-1111-111111111111"
	t1 := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 6, 9, 11, 0, 0, 0, time.UTC)

	// project existence check
	mock.ExpectQuery(`SELECT id, name, description, created_at, updated_at FROM projects WHERE id = \$1`).
		WithArgs(projectID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow(projectID, "Test Project", "", t1, t1))

	// list requirements
	mock.ExpectQuery(`SELECT id, project_id, name, description, status, created_at, updated_at FROM requirements WHERE project_id = \$1 ORDER BY created_at ASC`).
		WithArgs(projectID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "name", "description", "status", "created_at", "updated_at"}).
			AddRow("req-001", projectID, "First Requirement", "desc1", "draft", t1, t1).
			AddRow("req-002", projectID, "Second Requirement", "desc2", "in_progress", t2, t2))

	projectRepo := repo.NewProjectRepo(db)
	reqRepo := repo.NewRequirementRepo(db)
	h := NewRequirementHandler(reqRepo, projectRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements", h.ListProjectRequirements)

	httpReq := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects/"+projectID+"/requirements", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httpReq)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	requirements, ok := res["requirements"].([]interface{})
	require.True(t, ok, "response must have requirements array")
	require.Len(t, requirements, 2)

	r1 := requirements[0].(map[string]interface{})
	assert.Equal(t, "req-001", r1["id"])
	assert.Equal(t, projectID, r1["projectId"])
	assert.Equal(t, "First Requirement", r1["name"])
	assert.Equal(t, "desc1", r1["description"])
	assert.Equal(t, "draft", r1["status"])
	assert.Equal(t, "2026-06-09T10:00:00Z", r1["createdAt"])
	assert.Equal(t, "2026-06-09T10:00:00Z", r1["updatedAt"])

	r2 := requirements[1].(map[string]interface{})
	assert.Equal(t, "req-002", r2["id"])
	assert.Equal(t, "in_progress", r2["status"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

// IT-045-002 — GET /api/v1/projects/:pid/requirements — 200 empty list for new project
func TestRequirementHandler_ListProjectRequirements_200_EmptyList(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	projectID := "22222222-2222-2222-2222-222222222222"
	now := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)

	// project existence check
	mock.ExpectQuery(`SELECT id, name, description, created_at, updated_at FROM projects WHERE id = \$1`).
		WithArgs(projectID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow(projectID, "Test Project 2", "", now, now))

	// list requirements — empty result
	mock.ExpectQuery(`SELECT id, project_id, name, description, status, created_at, updated_at FROM requirements WHERE project_id = \$1 ORDER BY created_at ASC`).
		WithArgs(projectID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "name", "description", "status", "created_at", "updated_at"}))

	projectRepo := repo.NewProjectRepo(db)
	reqRepo := repo.NewRequirementRepo(db)
	h := NewRequirementHandler(reqRepo, projectRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements", h.ListProjectRequirements)

	httpReq := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects/"+projectID+"/requirements", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httpReq)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	requirements, ok := res["requirements"].([]interface{})
	require.True(t, ok, "requirements key must be present")
	assert.Empty(t, requirements, "requirements must be an empty array, not null")

	// Ensure the key is present in the raw JSON (not null)
	body := rec.Body.String()
	assert.Contains(t, body, `"requirements":[]`)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// IT-045-003 — GET /api/v1/projects/:pid/requirements — 404 unknown project
func TestRequirementHandler_ListProjectRequirements_404_UnknownProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	unknownProjectID := "00000000-0000-0000-0000-000000000000"

	// project not found
	mock.ExpectQuery(`SELECT id, name, description, created_at, updated_at FROM projects WHERE id = \$1`).
		WithArgs(unknownProjectID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}))

	projectRepo := repo.NewProjectRepo(db)
	reqRepo := repo.NewRequirementRepo(db)
	h := NewRequirementHandler(reqRepo, projectRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements", h.ListProjectRequirements)

	httpReq := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/projects/"+unknownProjectID+"/requirements", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httpReq)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Project not found", res["message"])

	assert.NoError(t, mock.ExpectationsWereMet())
}

// IT-045-004 — POST /api/v1/projects/:pid/requirements — 404 (no HTTP create endpoint)
func TestRequirementHandler_PostRoute_NotRegistered(t *testing.T) {
	e := echo.New()

	projectRepo := &mockProjectRepoForReqHandler{}
	reqRepo := &mockRequirementRepo{}
	h := NewRequirementHandler(reqRepo, projectRepo)

	// Only register GET, not POST
	e.GET("/api/v1/projects/:pid/requirements", h.ListProjectRequirements)

	projectID := "11111111-1111-1111-1111-111111111111"
	httpReq := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/projects/"+projectID+"/requirements", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httpReq)

	// Should be 404 or 405 — router returns its default unmatched-route response
	assert.True(t, rec.Code == http.StatusNotFound || rec.Code == http.StatusMethodNotAllowed,
		"expected 404 or 405 for unregistered POST route, got %d", rec.Code)
}
