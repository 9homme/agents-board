package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"agent-board/internal/domain"
	"agent-board/internal/handler"
	"agent-board/internal/mcp"
	"agent-board/internal/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRequirementRepo is a hand-written mock for repo.RequirementRepository used in MCP tool tests.
type MockRequirementRepo struct {
	repo.RequirementRepository
	ListByProjectFunc    func(ctx context.Context, projectID string) ([]domain.Requirement, error)
	CreateFunc           func(ctx context.Context, req *domain.Requirement) (*domain.Requirement, error)
	GetRequirementFunc   func(ctx context.Context, id string) (*domain.Requirement, error)
	UpdateFunc           func(ctx context.Context, id string, patch repo.RequirementPatch) (*domain.Requirement, error)
}

func (m *MockRequirementRepo) ListByProject(ctx context.Context, projectID string) ([]domain.Requirement, error) {
	if m.ListByProjectFunc != nil {
		return m.ListByProjectFunc(ctx, projectID)
	}
	return []domain.Requirement{}, nil
}

func (m *MockRequirementRepo) Create(ctx context.Context, req *domain.Requirement) (*domain.Requirement, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, req)
	}
	return nil, errors.New("Create not implemented")
}

func (m *MockRequirementRepo) GetRequirement(ctx context.Context, id string) (*domain.Requirement, error) {
	if m.GetRequirementFunc != nil {
		return m.GetRequirementFunc(ctx, id)
	}
	return nil, repo.ErrNotFound
}

func (m *MockRequirementRepo) Update(ctx context.Context, id string, patch repo.RequirementPatch) (*domain.Requirement, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, patch)
	}
	return nil, errors.New("Update not implemented")
}

// newRequirementToolsRegistry is a test helper that wires up a ToolRegistry with requirement tools.
func newRequirementToolsRegistry(reqRepo repo.RequirementRepository) *mcp.ToolRegistry {
	registry := mcp.NewToolRegistry()
	handler.RegisterRequirementTools(registry, reqRepo)
	return registry
}

// sampleRequirement returns a representative domain.Requirement for use in mock returns.
func sampleRequirement(id, projectID, name, status string, ts time.Time) *domain.Requirement {
	return &domain.Requirement{
		ID:          id,
		ProjectID:   projectID,
		Name:        name,
		Description: "",
		Status:      status,
		CreatedAt:   ts,
		UpdatedAt:   ts,
	}
}

// UT-045-021 — MCP create_requirement happy path (status defaults to draft)
func TestCreateRequirementTool_HappyPath_DefaultsDraft(t *testing.T) {
	now := time.Now()
	projectID := "11111111-1111-1111-1111-111111111111"
	reqID := "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f"

	mockRepo := &MockRequirementRepo{}
	mockRepo.CreateFunc = func(_ context.Context, r *domain.Requirement) (*domain.Requirement, error) {
		assert.Equal(t, projectID, r.ProjectID)
		assert.Equal(t, "REQ008 Requirement entity", r.Name)
		assert.Equal(t, "draft", r.Status)
		return sampleRequirement(reqID, projectID, r.Name, "draft", now), nil
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("create_requirement")
	require.True(t, ok)

	args := json.RawMessage(`{"project_id":"11111111-1111-1111-1111-111111111111","name":"REQ008 Requirement entity"}`)
	res, err := tool(context.Background(), args)
	require.NoError(t, err)
	require.NotNil(t, res)

	b, err := json.Marshal(res)
	require.NoError(t, err)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, reqID, m["id"])
	assert.Equal(t, projectID, m["projectId"])
	assert.Equal(t, "REQ008 Requirement entity", m["name"])
	assert.Equal(t, "", m["description"])
	assert.Equal(t, "draft", m["status"])
	assert.NotEmpty(t, m["createdAt"])
	assert.NotEmpty(t, m["updatedAt"])
}

// UT-045-022 — MCP create_requirement — blank name returns tool error
func TestCreateRequirementTool_BlankName(t *testing.T) {
	mockRepo := &MockRequirementRepo{}
	called := false
	mockRepo.CreateFunc = func(_ context.Context, _ *domain.Requirement) (*domain.Requirement, error) {
		called = true
		return nil, errors.New("should not be called")
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("create_requirement")
	require.True(t, ok)

	// blank name
	_, err := tool(context.Background(), json.RawMessage(`{"project_id":"proj-1","name":""}`))
	require.Error(t, err)
	assert.False(t, called, "repo Create must NOT be called when name is blank")

	// whitespace-only
	_, err = tool(context.Background(), json.RawMessage(`{"project_id":"proj-1","name":"   "}`))
	require.Error(t, err)
	assert.False(t, called)
}

// UT-045-023 — MCP create_requirement — project not found returns tool error
func TestCreateRequirementTool_ProjectNotFound(t *testing.T) {
	mockRepo := &MockRequirementRepo{}
	mockRepo.CreateFunc = func(_ context.Context, _ *domain.Requirement) (*domain.Requirement, error) {
		return nil, repo.ErrProjectNotFound
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("create_requirement")
	require.True(t, ok)

	args := json.RawMessage(`{"project_id":"nonexistent","name":"Valid Name"}`)
	res, err := tool(context.Background(), args)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "project not found")
}

// UT-045-024 — MCP create_requirement — invalid status returns tool error
func TestCreateRequirementTool_InvalidStatus(t *testing.T) {
	mockRepo := &MockRequirementRepo{}
	called := false
	mockRepo.CreateFunc = func(_ context.Context, _ *domain.Requirement) (*domain.Requirement, error) {
		called = true
		return nil, nil
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("create_requirement")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"project_id":"proj-1","name":"Name","status":"invalid_status"}`))
	require.Error(t, err)
	assert.False(t, called, "repo Create must NOT be called on invalid status")
}

// UT-045-025 — MCP create_requirement — explicit status in_progress
func TestCreateRequirementTool_ExplicitStatus_InProgress(t *testing.T) {
	now := time.Now()
	projectID := "proj-1"
	mockRepo := &MockRequirementRepo{}
	mockRepo.CreateFunc = func(_ context.Context, r *domain.Requirement) (*domain.Requirement, error) {
		assert.Equal(t, "in_progress", r.Status)
		return sampleRequirement("req-1", projectID, r.Name, "in_progress", now), nil
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("create_requirement")
	require.True(t, ok)

	res, err := tool(context.Background(), json.RawMessage(`{"project_id":"proj-1","name":"Name","status":"in_progress"}`))
	require.NoError(t, err)

	b, _ := json.Marshal(res)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "in_progress", m["status"])
}

// UT-045-026 — MCP create_requirement — explicit status done
func TestCreateRequirementTool_ExplicitStatus_Done(t *testing.T) {
	now := time.Now()
	projectID := "proj-1"
	mockRepo := &MockRequirementRepo{}
	mockRepo.CreateFunc = func(_ context.Context, r *domain.Requirement) (*domain.Requirement, error) {
		assert.Equal(t, "done", r.Status)
		return sampleRequirement("req-1", projectID, r.Name, "done", now), nil
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("create_requirement")
	require.True(t, ok)

	res, err := tool(context.Background(), json.RawMessage(`{"project_id":"proj-1","name":"Name","status":"done"}`))
	require.NoError(t, err)

	b, _ := json.Marshal(res)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "done", m["status"])
}

// UT-045-027 — MCP create_requirement — generic repo error returns tool error
func TestCreateRequirementTool_RepoError(t *testing.T) {
	mockRepo := &MockRequirementRepo{}
	mockErr := errors.New("db error")
	mockRepo.CreateFunc = func(_ context.Context, _ *domain.Requirement) (*domain.Requirement, error) {
		return nil, mockErr
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("create_requirement")
	require.True(t, ok)

	res, err := tool(context.Background(), json.RawMessage(`{"project_id":"proj-1","name":"Valid Name"}`))
	require.Error(t, err)
	assert.Nil(t, res)
}

// UT-045-028 — MCP list_requirements happy path
func TestListRequirementsTool_HappyPath(t *testing.T) {
	now := time.Now()
	projectID := "11111111-1111-1111-1111-111111111111"

	reqs := []domain.Requirement{
		{ID: "req-001", ProjectID: projectID, Name: "First", Description: "d1", Status: "draft", CreatedAt: now, UpdatedAt: now},
		{ID: "req-002", ProjectID: projectID, Name: "Second", Description: "d2", Status: "in_progress", CreatedAt: now, UpdatedAt: now},
	}

	mockRepo := &MockRequirementRepo{}
	mockRepo.ListByProjectFunc = func(_ context.Context, pid string) ([]domain.Requirement, error) {
		assert.Equal(t, projectID, pid)
		return reqs, nil
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("list_requirements")
	require.True(t, ok)

	res, err := tool(context.Background(), json.RawMessage(`{"project_id":"11111111-1111-1111-1111-111111111111"}`))
	require.NoError(t, err)

	b, _ := json.Marshal(res)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))

	requirements, ok := m["requirements"].([]interface{})
	require.True(t, ok, "must have requirements array")
	require.Len(t, requirements, 2)

	r1 := requirements[0].(map[string]interface{})
	assert.Equal(t, "req-001", r1["id"])
	assert.Equal(t, projectID, r1["projectId"])
	assert.Equal(t, "First", r1["name"])
	assert.Equal(t, "d1", r1["description"])
	assert.Equal(t, "draft", r1["status"])
	assert.NotEmpty(t, r1["createdAt"])
	assert.NotEmpty(t, r1["updatedAt"])
}

// UT-045-029 — MCP list_requirements — unknown project_id returns tool error
func TestListRequirementsTool_UnknownProject(t *testing.T) {
	mockRepo := &MockRequirementRepo{}
	mockRepo.ListByProjectFunc = func(_ context.Context, _ string) ([]domain.Requirement, error) {
		return nil, repo.ErrProjectNotFound
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("list_requirements")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"project_id":"nonexistent"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project not found")
}

// UT-045-030 — MCP list_requirements — generic repo error
func TestListRequirementsTool_RepoError(t *testing.T) {
	mockRepo := &MockRequirementRepo{}
	mockRepo.ListByProjectFunc = func(_ context.Context, _ string) ([]domain.Requirement, error) {
		return nil, errors.New("db error")
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("list_requirements")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"project_id":"proj-1"}`))
	require.Error(t, err)
}

// UT-045-031 — MCP update_requirement happy path (status change)
func TestUpdateRequirementTool_HappyPath_StatusChange(t *testing.T) {
	createdAt := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 6, 9, 12, 30, 0, 0, time.UTC)
	projectID := "11111111-1111-1111-1111-111111111111"
	reqID := "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f"

	mockRepo := &MockRequirementRepo{}
	mockRepo.UpdateFunc = func(_ context.Context, id string, patch repo.RequirementPatch) (*domain.Requirement, error) {
		assert.Equal(t, reqID, id)
		require.NotNil(t, patch.Status)
		assert.Equal(t, "in_progress", *patch.Status)
		return &domain.Requirement{
			ID:          id,
			ProjectID:   projectID,
			Name:        "Default",
			Description: "",
			Status:      "in_progress",
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}, nil
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("update_requirement")
	require.True(t, ok)

	args := json.RawMessage(`{"requirement_id":"b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f","status":"in_progress"}`)
	res, err := tool(context.Background(), args)
	require.NoError(t, err)

	b, _ := json.Marshal(res)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, reqID, m["id"])
	assert.Equal(t, projectID, m["projectId"])
	assert.Equal(t, "in_progress", m["status"])
	assert.Equal(t, "2026-06-09T10:00:00Z", m["createdAt"])
	assert.Equal(t, "2026-06-09T12:30:00Z", m["updatedAt"])
}

// UT-045-032 — MCP update_requirement — name update only
func TestUpdateRequirementTool_NameUpdate(t *testing.T) {
	now := time.Now()
	projectID := "proj-1"
	reqID := "req-001"

	mockRepo := &MockRequirementRepo{}
	mockRepo.UpdateFunc = func(_ context.Context, id string, patch repo.RequirementPatch) (*domain.Requirement, error) {
		require.NotNil(t, patch.Name)
		assert.Equal(t, "New Name", *patch.Name)
		assert.Nil(t, patch.Status)
		assert.Nil(t, patch.Description)
		return &domain.Requirement{
			ID:          id,
			ProjectID:   projectID,
			Name:        "New Name",
			Description: "",
			Status:      "draft",
			CreatedAt:   now,
			UpdatedAt:   now,
		}, nil
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("update_requirement")
	require.True(t, ok)

	args := json.RawMessage(`{"requirement_id":"req-001","name":"New Name"}`)
	res, err := tool(context.Background(), args)
	require.NoError(t, err)

	b, _ := json.Marshal(res)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, reqID, m["id"])
	assert.Equal(t, "New Name", m["name"])
	assert.Equal(t, "draft", m["status"])
}

// UT-045-033 — MCP update_requirement — description update only
func TestUpdateRequirementTool_DescriptionUpdate(t *testing.T) {
	now := time.Now()
	projectID := "proj-1"

	mockRepo := &MockRequirementRepo{}
	mockRepo.UpdateFunc = func(_ context.Context, id string, patch repo.RequirementPatch) (*domain.Requirement, error) {
		require.NotNil(t, patch.Description)
		assert.Equal(t, "New desc", *patch.Description)
		assert.Nil(t, patch.Name)
		assert.Nil(t, patch.Status)
		return &domain.Requirement{
			ID:          id,
			ProjectID:   projectID,
			Name:        "Original",
			Description: "New desc",
			Status:      "draft",
			CreatedAt:   now,
			UpdatedAt:   now,
		}, nil
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("update_requirement")
	require.True(t, ok)

	args := json.RawMessage(`{"requirement_id":"req-001","description":"New desc"}`)
	res, err := tool(context.Background(), args)
	require.NoError(t, err)

	b, _ := json.Marshal(res)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "New desc", m["description"])
	assert.Equal(t, "Original", m["name"])
}

// UT-045-034 — MCP update_requirement — invalid status value returns tool error
func TestUpdateRequirementTool_InvalidStatus(t *testing.T) {
	mockRepo := &MockRequirementRepo{}
	called := false
	mockRepo.UpdateFunc = func(_ context.Context, _ string, _ repo.RequirementPatch) (*domain.Requirement, error) {
		called = true
		return nil, nil
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("update_requirement")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"requirement_id":"req-1","status":"not_a_valid_status"}`))
	require.Error(t, err)
	assert.False(t, called, "repo Update must NOT be called on invalid status")
}

// UT-045-035 — MCP update_requirement — blank name when provided returns tool error
func TestUpdateRequirementTool_BlankName(t *testing.T) {
	mockRepo := &MockRequirementRepo{}
	called := false
	mockRepo.UpdateFunc = func(_ context.Context, _ string, _ repo.RequirementPatch) (*domain.Requirement, error) {
		called = true
		return nil, nil
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("update_requirement")
	require.True(t, ok)

	// empty name
	_, err := tool(context.Background(), json.RawMessage(`{"requirement_id":"req-1","name":""}`))
	require.Error(t, err)
	assert.False(t, called)

	// whitespace-only name
	_, err = tool(context.Background(), json.RawMessage(`{"requirement_id":"req-1","name":"   "}`))
	require.Error(t, err)
	assert.False(t, called)
}

// UT-045-036 — MCP update_requirement — requirement not found returns tool error
func TestUpdateRequirementTool_NotFound(t *testing.T) {
	mockRepo := &MockRequirementRepo{}
	mockRepo.UpdateFunc = func(_ context.Context, _ string, _ repo.RequirementPatch) (*domain.Requirement, error) {
		return nil, repo.ErrNotFound
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("update_requirement")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"requirement_id":"nonexistent","status":"draft"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// UT-045-037 — MCP update_requirement — all-empty patch is a no-op returning current object
func TestUpdateRequirementTool_AllEmptyNoOp(t *testing.T) {
	now := time.Now()
	projectID := "proj-1"

	mockRepo := &MockRequirementRepo{}
	mockRepo.UpdateFunc = func(_ context.Context, id string, patch repo.RequirementPatch) (*domain.Requirement, error) {
		// all fields nil — no-op; repo may still be called and returns current object
		return &domain.Requirement{
			ID:          id,
			ProjectID:   projectID,
			Name:        "Original",
			Description: "desc",
			Status:      "draft",
			CreatedAt:   now,
			UpdatedAt:   now,
		}, nil
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("update_requirement")
	require.True(t, ok)

	res, err := tool(context.Background(), json.RawMessage(`{"requirement_id":"req-1"}`))
	require.NoError(t, err)
	require.NotNil(t, res)

	b, _ := json.Marshal(res)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, "Original", m["name"])
	assert.Equal(t, "draft", m["status"])
}

// UT-045-038 — MCP update_requirement — generic repo error returns tool error
func TestUpdateRequirementTool_RepoError(t *testing.T) {
	mockRepo := &MockRequirementRepo{}
	mockErr := errors.New("db error")
	mockRepo.UpdateFunc = func(_ context.Context, _ string, _ repo.RequirementPatch) (*domain.Requirement, error) {
		return nil, mockErr
	}

	registry := newRequirementToolsRegistry(mockRepo)
	tool, ok := registry.GetTool("update_requirement")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"requirement_id":"req-1","status":"draft"}`))
	require.Error(t, err)
}
