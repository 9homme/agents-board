package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"agent-board/internal/domain"
	"agent-board/internal/mcp"
	"agent-board/internal/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errInvalidPathForTest is a sentinel used in UT-045-047 to simulate ErrInvalidPath.
var errInvalidPathForTest = errors.New("path does not exist or is not a directory")

// MockPathValidator is a hand-written mock for PathValidator.
type MockPathValidator struct {
	ValidatePathFunc func(path string) error
}

func (m *MockPathValidator) ValidatePath(path string) error {
	if m.ValidatePathFunc != nil {
		return m.ValidatePathFunc(path)
	}
	return nil
}

// MockProjectRepo is a hand-written mock for repo.ProjectRepository.
// Each method delegates to its corresponding Func field; if the field is nil the method no-ops.
type MockProjectRepo struct {
	repo.ProjectRepository // embed for forward-compat
	CreateProjectFunc      func(ctx context.Context, p *domain.Project) (*domain.Project, error)
	GetProjectFunc         func(ctx context.Context, id string) (*domain.Project, error)
	UpdateProjectFunc      func(ctx context.Context, p *domain.Project) (*domain.Project, error)
	DeleteProjectFunc      func(ctx context.Context, id string) error
	ListProjectsFunc       func(ctx context.Context) ([]*domain.Project, error)
}

// CreateProject delegates to CreateProjectFunc if set, otherwise returns nil.
func (m *MockProjectRepo) CreateProject(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	if m.CreateProjectFunc != nil {
		return m.CreateProjectFunc(ctx, p)
	}
	return nil, nil
}

// GetProject delegates to GetProjectFunc if set, otherwise returns nil.
func (m *MockProjectRepo) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	if m.GetProjectFunc != nil {
		return m.GetProjectFunc(ctx, id)
	}
	return nil, nil
}

// UpdateProject delegates to UpdateProjectFunc if set, otherwise returns nil.
func (m *MockProjectRepo) UpdateProject(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	if m.UpdateProjectFunc != nil {
		return m.UpdateProjectFunc(ctx, p)
	}
	return nil, nil
}

// DeleteProject delegates to DeleteProjectFunc if set, otherwise returns nil.
func (m *MockProjectRepo) DeleteProject(ctx context.Context, id string) error {
	if m.DeleteProjectFunc != nil {
		return m.DeleteProjectFunc(ctx, id)
	}
	return nil
}

// ListProjects delegates to ListProjectsFunc if set, otherwise returns nil.
func (m *MockProjectRepo) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	if m.ListProjectsFunc != nil {
		return m.ListProjectsFunc(ctx)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// UT-001: TestRegisterProjectTools_RegistersAllFiveTools
// ---------------------------------------------------------------------------

// TestRegisterProjectTools_RegistersAllFiveTools verifies that RegisterProjectTools
// registers exactly the five expected tools into a fresh ToolRegistry.
func TestRegisterProjectTools_RegistersAllFiveTools(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{}
	RegisterProjectTools(registry, mockRepo, nil)

	names := []string{"create_project", "get_project", "update_project", "delete_project", "list_projects"}
	for _, name := range names {
		tool, ok := registry.GetTool(name)
		assert.True(t, ok, "expected tool %q to be registered", name)
		assert.NotNil(t, tool, "expected tool %q handler to be non-nil", name)
	}

	_, ok := registry.GetTool("nonexistent_tool")
	assert.False(t, ok, "expected nonexistent_tool to be absent from registry")
}

// ---------------------------------------------------------------------------
// UT-002: TestHandleCreateProject_InvalidArguments
// ---------------------------------------------------------------------------

// TestHandleCreateProject_InvalidArguments verifies that malformed JSON triggers
// the "invalid arguments" error path before any repo call.
func TestHandleCreateProject_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("create_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage("not-valid-json"))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// ---------------------------------------------------------------------------
// UT-003: TestHandleCreateProject_EmptyName
// ---------------------------------------------------------------------------

// TestHandleCreateProject_EmptyName verifies that an empty/whitespace project name
// is rejected before any repo call.
func TestHandleCreateProject_EmptyName(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("create_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"name": ""}`))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required and cannot be empty")
}

// ---------------------------------------------------------------------------
// UT-004: TestHandleCreateProject_RepoError
// ---------------------------------------------------------------------------

// TestHandleCreateProject_RepoError verifies that a repo error on CreateProject
// is passed through without wrapping.
func TestHandleCreateProject_RepoError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockErr := errors.New("db down")
	mockRepo := &MockProjectRepo{
		CreateProjectFunc: func(_ context.Context, _ *domain.Project) (*domain.Project, error) {
			return nil, mockErr
		},
	}
	mockValidator := &MockPathValidator{ValidatePathFunc: func(_ string) error { return nil }}
	RegisterProjectTools(registry, mockRepo, mockValidator)

	tool, ok := registry.GetTool("create_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"name": "My Project", "path": "/tmp"}`))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.True(t, errors.Is(err, mockErr), "expected passthrough: errors.Is(err, mockErr)")
}

// ---------------------------------------------------------------------------
// UT-005: TestHandleGetProject_InvalidArguments
// ---------------------------------------------------------------------------

// TestHandleGetProject_InvalidArguments verifies that malformed JSON returns
// "invalid arguments".
func TestHandleGetProject_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("get_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage("not-valid-json"))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// ---------------------------------------------------------------------------
// UT-006: TestHandleGetProject_EmptyID
// ---------------------------------------------------------------------------

// TestHandleGetProject_EmptyID verifies that an empty id field is rejected.
func TestHandleGetProject_EmptyID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("get_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": ""}`))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

// ---------------------------------------------------------------------------
// UT-007: TestHandleGetProject_NotFound
// ---------------------------------------------------------------------------

// TestHandleGetProject_NotFound verifies that repo.ErrNotFound is converted to a
// fresh "project not found" error (sentinel NOT preserved).
func TestHandleGetProject_NotFound(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{
		GetProjectFunc: func(_ context.Context, _ string) (*domain.Project, error) {
			return nil, repo.ErrNotFound
		},
	}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("get_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "proj-1"}`))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
	assert.False(t, errors.Is(err, repo.ErrNotFound), "ErrNotFound sentinel must NOT be preserved")
}

// ---------------------------------------------------------------------------
// UT-008: TestHandleGetProject_GenericError
// ---------------------------------------------------------------------------

// TestHandleGetProject_GenericError verifies that a non-NotFound repo error is
// passed through without wrapping.
func TestHandleGetProject_GenericError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockErr := errors.New("db down")
	mockRepo := &MockProjectRepo{
		GetProjectFunc: func(_ context.Context, _ string) (*domain.Project, error) {
			return nil, mockErr
		},
	}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("get_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "proj-1"}`))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.True(t, errors.Is(err, mockErr), "expected passthrough: errors.Is(err, mockErr)")
}

// ---------------------------------------------------------------------------
// UT-009: TestHandleUpdateProject_InvalidArguments
// ---------------------------------------------------------------------------

// TestHandleUpdateProject_InvalidArguments verifies that malformed JSON returns
// "invalid arguments".
func TestHandleUpdateProject_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("update_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage("not-valid-json"))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// ---------------------------------------------------------------------------
// UT-010: TestHandleUpdateProject_EmptyID
// ---------------------------------------------------------------------------

// TestHandleUpdateProject_EmptyID verifies that an empty id field is rejected
// before any repo call.
func TestHandleUpdateProject_EmptyID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("update_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": ""}`))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

// ---------------------------------------------------------------------------
// UT-011: TestHandleUpdateProject_NotFoundOnInitialGet
// ---------------------------------------------------------------------------

// TestHandleUpdateProject_NotFoundOnInitialGet verifies that ErrNotFound on the
// initial GetProject call is converted to a fresh "project not found" error.
func TestHandleUpdateProject_NotFoundOnInitialGet(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{
		GetProjectFunc: func(_ context.Context, _ string) (*domain.Project, error) {
			return nil, repo.ErrNotFound
		},
	}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("update_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "proj-1"}`))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
	assert.False(t, errors.Is(err, repo.ErrNotFound), "ErrNotFound sentinel must NOT be preserved")
}

// ---------------------------------------------------------------------------
// UT-012: TestHandleUpdateProject_GenericErrorOnInitialGet
// ---------------------------------------------------------------------------

// TestHandleUpdateProject_GenericErrorOnInitialGet verifies that a non-NotFound
// error on the initial GetProject call is passed through without wrapping.
func TestHandleUpdateProject_GenericErrorOnInitialGet(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockErr := errors.New("db down")
	mockRepo := &MockProjectRepo{
		GetProjectFunc: func(_ context.Context, _ string) (*domain.Project, error) {
			return nil, mockErr
		},
	}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("update_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "proj-1"}`))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.True(t, errors.Is(err, mockErr), "expected passthrough: errors.Is(err, mockErr)")
}

// ---------------------------------------------------------------------------
// UT-013: TestHandleUpdateProject_EmptyNameWhenProvided
// ---------------------------------------------------------------------------

// TestHandleUpdateProject_EmptyNameWhenProvided verifies that an explicitly
// provided but whitespace-only name field is rejected.
func TestHandleUpdateProject_EmptyNameWhenProvided(t *testing.T) {
	now := time.Now()
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{
		GetProjectFunc: func(_ context.Context, _ string) (*domain.Project, error) {
			return &domain.Project{
				ID:        "proj-1",
				Name:      "Old Name",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("update_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "proj-1", "name": " "}`))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty if provided")
}

// ---------------------------------------------------------------------------
// UT-014: TestHandleUpdateProject_RepoUpdateError
// ---------------------------------------------------------------------------

// TestHandleUpdateProject_RepoUpdateError verifies that an error from UpdateProject
// is passed through without wrapping.
func TestHandleUpdateProject_RepoUpdateError(t *testing.T) {
	now := time.Now()
	registry := mcp.NewToolRegistry()
	mockErr := errors.New("db down")
	mockRepo := &MockProjectRepo{
		GetProjectFunc: func(_ context.Context, _ string) (*domain.Project, error) {
			return &domain.Project{
				ID:        "proj-1",
				Name:      "Old Name",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
		UpdateProjectFunc: func(_ context.Context, _ *domain.Project) (*domain.Project, error) {
			return nil, mockErr
		},
	}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("update_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "proj-1", "name": "New Name"}`))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.True(t, errors.Is(err, mockErr), "expected passthrough: errors.Is(err, mockErr)")
}

// ---------------------------------------------------------------------------
// UT-015: TestHandleDeleteProject_InvalidArguments
// ---------------------------------------------------------------------------

// TestHandleDeleteProject_InvalidArguments verifies that malformed JSON returns
// "invalid arguments".
func TestHandleDeleteProject_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("delete_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage("not-valid-json"))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// ---------------------------------------------------------------------------
// UT-016: TestHandleDeleteProject_EmptyID
// ---------------------------------------------------------------------------

// TestHandleDeleteProject_EmptyID verifies that an empty id field is rejected.
func TestHandleDeleteProject_EmptyID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("delete_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": ""}`))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

// ---------------------------------------------------------------------------
// UT-017: TestHandleDeleteProject_RepoError
// ---------------------------------------------------------------------------

// TestHandleDeleteProject_RepoError verifies that a repo error on DeleteProject
// is passed through without wrapping.
func TestHandleDeleteProject_RepoError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockErr := errors.New("db down")
	mockRepo := &MockProjectRepo{
		DeleteProjectFunc: func(_ context.Context, _ string) error {
			return mockErr
		},
	}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("delete_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "proj-1"}`))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.True(t, errors.Is(err, mockErr), "expected passthrough: errors.Is(err, mockErr)")
}

// ---------------------------------------------------------------------------
// UT-018: TestHandleListProjects_RepoError
// ---------------------------------------------------------------------------

// TestHandleListProjects_RepoError verifies that a repo error on ListProjects
// is passed through without wrapping.
func TestHandleListProjects_RepoError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockErr := errors.New("db down")
	mockRepo := &MockProjectRepo{
		ListProjectsFunc: func(_ context.Context) ([]*domain.Project, error) {
			return nil, mockErr
		},
	}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("list_projects")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{}`))

	assert.Nil(t, result)
	require.Error(t, err)
	assert.True(t, errors.Is(err, mockErr), "expected passthrough: errors.Is(err, mockErr)")
}

// ---------------------------------------------------------------------------
// Happy-path tests — named to match the IT-001 coverage-filter pattern
// so that the architecture §4.6 command achieves ≥95% per-file coverage.
// ---------------------------------------------------------------------------

// TestHandleCreateProject_Success tests the successful creation path.
func TestHandleCreateProject_Success(t *testing.T) {
	now := time.Now()
	expected := &domain.Project{
		ID:          "proj-1",
		Name:        "My Project",
		Description: "",
		Path:        "/tmp",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{
		CreateProjectFunc: func(_ context.Context, _ *domain.Project) (*domain.Project, error) {
			return expected, nil
		},
	}
	mockValidator := &MockPathValidator{ValidatePathFunc: func(_ string) error { return nil }}
	RegisterProjectTools(registry, mockRepo, mockValidator)

	tool, ok := registry.GetTool("create_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"name": "My Project", "path": "/tmp"}`))
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

// TestHandleGetProject_Success tests the successful retrieval path.
func TestHandleGetProject_Success(t *testing.T) {
	now := time.Now()
	expected := &domain.Project{
		ID:        "proj-1",
		Name:      "My Project",
		CreatedAt: now,
		UpdatedAt: now,
	}
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{
		GetProjectFunc: func(_ context.Context, _ string) (*domain.Project, error) {
			return expected, nil
		},
	}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("get_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "proj-1"}`))
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

// TestHandleUpdateProject_Success tests the successful update path.
func TestHandleUpdateProject_Success(t *testing.T) {
	now := time.Now()
	existing := &domain.Project{ID: "proj-1", Name: "Old", CreatedAt: now, UpdatedAt: now}
	updated := &domain.Project{ID: "proj-1", Name: "New Name", CreatedAt: now, UpdatedAt: now}
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{
		GetProjectFunc: func(_ context.Context, _ string) (*domain.Project, error) {
			return existing, nil
		},
		UpdateProjectFunc: func(_ context.Context, _ *domain.Project) (*domain.Project, error) {
			return updated, nil
		},
	}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("update_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "proj-1", "name": "New Name"}`))
	require.NoError(t, err)
	assert.Equal(t, updated, result)
}

// TestHandleDeleteProject_Success tests the successful delete path.
func TestHandleDeleteProject_Success(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{
		DeleteProjectFunc: func(_ context.Context, _ string) error { return nil },
	}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("delete_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "proj-1"}`))
	require.NoError(t, err)
	assert.Equal(t, map[string]bool{"success": true}, result)
}

// TestHandleListProjects_Success tests the successful list path (non-nil result).
func TestHandleListProjects_Success(t *testing.T) {
	now := time.Now()
	projects := []*domain.Project{{ID: "proj-1", Name: "P1", CreatedAt: now, UpdatedAt: now}}
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{
		ListProjectsFunc: func(_ context.Context) ([]*domain.Project, error) {
			return projects, nil
		},
	}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("list_projects")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{}`))
	require.NoError(t, err)
	assert.NotNil(t, result)
}

// TestHandleListProjects_NilReturnsEmpty verifies that a nil slice from the repo
// is normalised to an empty (non-null) list.
func TestHandleListProjects_NilReturnsEmpty(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockProjectRepo{
		ListProjectsFunc: func(_ context.Context) ([]*domain.Project, error) {
			return nil, nil
		},
	}
	RegisterProjectTools(registry, mockRepo, nil)

	tool, ok := registry.GetTool("list_projects")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{}`))
	require.NoError(t, err)
	m, ok := result.(map[string]interface{})
	require.True(t, ok)
	projects, ok := m["projects"].([]*domain.Project)
	require.True(t, ok)
	assert.NotNil(t, projects)
	assert.Empty(t, projects)
}

// ---------------------------------------------------------------------------
// Legacy happy-path tests (pre-existing — kept for regression coverage)
// ---------------------------------------------------------------------------

// TestProjectTools_CreateProject tests the happy path of create_project.
func TestProjectTools_CreateProject(t *testing.T) {
	now := time.Now()
	expectedProject := &domain.Project{
		ID:          "123",
		Name:        "Test Project",
		Description: "A test project",
		Path:        "/tmp/testdir",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	mockRepo := &MockProjectRepo{
		CreateProjectFunc: func(_ context.Context, p *domain.Project) (*domain.Project, error) {
			return expectedProject, nil
		},
	}
	mockValidator := &MockPathValidator{ValidatePathFunc: func(_ string) error { return nil }}

	h := handleCreateProject(mockRepo, mockValidator)
	args := json.RawMessage(`{"name":"Test Project","description":"A test project","path":"/tmp/testdir"}`)

	result, err := h(context.Background(), args)
	assert.NoError(t, err)

	resStr, err := json.Marshal(result)
	assert.NoError(t, err)

	var res map[string]interface{}
	err = json.Unmarshal(resStr, &res)
	assert.NoError(t, err)
	assert.Equal(t, "123", res["id"])
	assert.Equal(t, "Test Project", res["name"])
}

// TestProjectTools_GetProject tests the happy path of get_project.
func TestProjectTools_GetProject(t *testing.T) {
	now := time.Now()
	expectedProject := &domain.Project{
		ID:          "123",
		Name:        "Test Project",
		Description: "A test project",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	mockRepo := &MockProjectRepo{
		GetProjectFunc: func(_ context.Context, _ string) (*domain.Project, error) {
			return expectedProject, nil
		},
	}

	h := handleGetProject(mockRepo)
	args := json.RawMessage(`{"id":"123"}`)

	result, err := h(context.Background(), args)
	assert.NoError(t, err)

	resStr, err := json.Marshal(result)
	assert.NoError(t, err)

	var res map[string]interface{}
	err = json.Unmarshal(resStr, &res)
	assert.NoError(t, err)
	assert.Equal(t, "123", res["id"])
	assert.Equal(t, "Test Project", res["name"])
}

// TestProjectTools_UpdateProject tests the happy path of update_project.
func TestProjectTools_UpdateProject(t *testing.T) {
	now := time.Now()
	existing := &domain.Project{
		ID:          "123",
		Name:        "Old Project",
		Description: "Old desc",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	updated := &domain.Project{
		ID:          "123",
		Name:        "Updated Project",
		Description: "Updated desc",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	mockRepo := &MockProjectRepo{
		GetProjectFunc: func(_ context.Context, _ string) (*domain.Project, error) {
			return existing, nil
		},
		UpdateProjectFunc: func(_ context.Context, _ *domain.Project) (*domain.Project, error) {
			return updated, nil
		},
	}

	h := handleUpdateProject(mockRepo)
	args := json.RawMessage(`{"id":"123","name":"Updated Project","description":"Updated desc"}`)

	result, err := h(context.Background(), args)
	assert.NoError(t, err)

	resStr, err := json.Marshal(result)
	assert.NoError(t, err)

	var res map[string]interface{}
	err = json.Unmarshal(resStr, &res)
	assert.NoError(t, err)
	assert.Equal(t, "123", res["id"])
	assert.Equal(t, "Updated Project", res["name"])
}

// TestProjectTools_DeleteProject tests the happy path of delete_project.
func TestProjectTools_DeleteProject(t *testing.T) {
	mockRepo := &MockProjectRepo{
		DeleteProjectFunc: func(_ context.Context, _ string) error {
			return nil
		},
	}

	h := handleDeleteProject(mockRepo)
	args := json.RawMessage(`{"id":"123"}`)

	result, err := h(context.Background(), args)
	assert.NoError(t, err)

	resStr, err := json.Marshal(result)
	assert.NoError(t, err)

	var res map[string]interface{}
	err = json.Unmarshal(resStr, &res)
	assert.NoError(t, err)
	assert.Equal(t, true, res["success"])
}

// -----------------------------------------------------------------------
// US045 tests — create_project now requires path (D-008)
// -----------------------------------------------------------------------

// UT-045-045 — MCP create_project now requires path
func TestHandleCreateProject_WithPath(t *testing.T) {
	now := time.Now()
	registry := mcp.NewToolRegistry()
	mockValidator := &MockPathValidator{}
	mockRepo := &MockProjectRepo{}

	mockValidator.ValidatePathFunc = func(path string) error {
		assert.NotEmpty(t, path)
		return nil
	}

	mockRepo.CreateProjectFunc = func(_ context.Context, p *domain.Project) (*domain.Project, error) {
		assert.NotEmpty(t, p.Path, "Path must be passed to repo")
		return &domain.Project{
			ID:        "proj-1",
			Name:      p.Name,
			Path:      p.Path,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	RegisterProjectTools(registry, mockRepo, mockValidator)

	tool, ok := registry.GetTool("create_project")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"name":"Test Project","path":"/tmp/testdir"}`))
	require.NoError(t, err)

	b, _ := json.Marshal(result)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "/tmp/testdir", m["path"])
}

// UT-045-046 — MCP create_project — missing path returns tool error
func TestHandleCreateProject_MissingPath(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockValidator := &MockPathValidator{}
	mockRepo := &MockProjectRepo{}

	validatorCalled := false
	mockValidator.ValidatePathFunc = func(_ string) error {
		validatorCalled = true
		return nil
	}

	RegisterProjectTools(registry, mockRepo, mockValidator)

	tool, ok := registry.GetTool("create_project")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"name":"Test Project"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path")
	assert.False(t, validatorCalled, "ValidatePath must NOT be called when path is missing")
}

// UT-045-047 — MCP create_project — invalid path (not a directory) returns tool error
func TestHandleCreateProject_InvalidPath(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockValidator := &MockPathValidator{}
	mockRepo := &MockProjectRepo{}

	mockValidator.ValidatePathFunc = func(_ string) error {
		return errInvalidPathForTest
	}

	repoCalled := false
	mockRepo.CreateProjectFunc = func(_ context.Context, _ *domain.Project) (*domain.Project, error) {
		repoCalled = true
		return nil, nil
	}

	RegisterProjectTools(registry, mockRepo, mockValidator)

	tool, ok := registry.GetTool("create_project")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"name":"Test Project","path":"/not/a/dir"}`))
	require.Error(t, err)
	assert.False(t, repoCalled, "repo must NOT be called on invalid path")
}

// UT-045-048 — MCP create_project — duplicate path returns tool error
func TestHandleCreateProject_DuplicatePath(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockValidator := &MockPathValidator{}
	mockRepo := &MockProjectRepo{}

	mockValidator.ValidatePathFunc = func(_ string) error { return nil }
	mockRepo.CreateProjectFunc = func(_ context.Context, _ *domain.Project) (*domain.Project, error) {
		return nil, repo.ErrDuplicatePath
	}

	RegisterProjectTools(registry, mockRepo, mockValidator)

	tool, ok := registry.GetTool("create_project")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"name":"Test Project","path":"/already/taken"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path already")
}

// TestProjectTools_ListProjects tests the happy path of list_projects.
func TestProjectTools_ListProjects(t *testing.T) {
	now := time.Now()
	expected := []*domain.Project{
		{
			ID:          "123",
			Name:        "Test Project",
			Description: "A test project",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	mockRepo := &MockProjectRepo{
		ListProjectsFunc: func(_ context.Context) ([]*domain.Project, error) {
			return expected, nil
		},
	}

	h := handleListProjects(mockRepo)
	args := json.RawMessage(`{}`)

	result, err := h(context.Background(), args)
	assert.NoError(t, err)

	resStr, err := json.Marshal(result)
	assert.NoError(t, err)

	var res map[string]interface{}
	err = json.Unmarshal(resStr, &res)
	assert.NoError(t, err)

	projects := res["projects"].([]interface{})
	assert.Len(t, projects, 1)
	p1 := projects[0].(map[string]interface{})
	assert.Equal(t, "123", p1["id"])
}
