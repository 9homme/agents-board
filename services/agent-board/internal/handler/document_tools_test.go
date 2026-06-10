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

// MockDocumentRepo is a mock implementation of repo.DocumentRepository
type MockDocumentRepo struct {
	repo.DocumentRepository
	CreateDocumentFunc func(ctx context.Context, d *domain.Document) (*domain.Document, error)
	GetDocumentFunc    func(ctx context.Context, id string) (*domain.Document, error)
	UpdateDocumentFunc func(ctx context.Context, d *domain.Document) (*domain.Document, error)
	DeleteDocumentFunc func(ctx context.Context, id string) error
	ListDocumentsFunc  func(ctx context.Context, projectID string) ([]*domain.Document, error)
}

func (m *MockDocumentRepo) CreateDocument(ctx context.Context, d *domain.Document) (*domain.Document, error) {
	return m.CreateDocumentFunc(ctx, d)
}
func (m *MockDocumentRepo) GetDocument(ctx context.Context, id string) (*domain.Document, error) {
	return m.GetDocumentFunc(ctx, id)
}
func (m *MockDocumentRepo) UpdateDocument(ctx context.Context, d *domain.Document) (*domain.Document, error) {
	return m.UpdateDocumentFunc(ctx, d)
}
func (m *MockDocumentRepo) DeleteDocument(ctx context.Context, id string) error {
	return m.DeleteDocumentFunc(ctx, id)
}
func (m *MockDocumentRepo) ListDocuments(ctx context.Context, projectID string) ([]*domain.Document, error) {
	return m.ListDocumentsFunc(ctx, projectID)
}

func TestDocumentTools_CreateDocument(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	now := time.Now()
	mockRepo.CreateDocumentFunc = func(ctx context.Context, d *domain.Document) (*domain.Document, error) {
		assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", d.ProjectID)
		assert.Equal(t, "My Doc", d.Title)
		assert.Equal(t, "Content here", d.Content)
		return &domain.Document{
			ID:        "223e4567-e89b-12d3-a456-426614174000",
			ProjectID: d.ProjectID,
			Title:     d.Title,
			Content:   d.Content,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	args := json.RawMessage(`{"projectId":"123e4567-e89b-12d3-a456-426614174000", "title": "My Doc", "content": "Content here"}`)
	toolHandler, ok := registry.GetTool("create_document")
	require.True(t, ok)
	res, err := toolHandler(context.Background(), args)
	require.NoError(t, err)

	resResp, ok := res.(handler.DocumentResponse)
	require.True(t, ok)
	assert.Equal(t, "223e4567-e89b-12d3-a456-426614174000", resResp.ID)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", resResp.ProjectID)
	assert.Equal(t, "My Doc", resResp.Title)
	assert.Equal(t, "Content here", resResp.Content)
}

func TestDocumentTools_GetDocument(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	now := time.Now()
	mockRepo.GetDocumentFunc = func(ctx context.Context, id string) (*domain.Document, error) {
		if id == "223e4567-e89b-12d3-a456-426614174000" {
			return &domain.Document{
				ID:        id,
				ProjectID: "123e4567-e89b-12d3-a456-426614174000",
				Title:     "My Doc",
				Content:   "Content here",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		}
		return nil, repo.ErrNotFound
	}

	args := json.RawMessage(`{"id":"223e4567-e89b-12d3-a456-426614174000"}`)
	toolHandler, ok := registry.GetTool("get_document")
	require.True(t, ok)
	res, err := toolHandler(context.Background(), args)
	require.NoError(t, err)

	resResp, ok := res.(handler.DocumentResponse)
	require.True(t, ok)
	assert.Equal(t, "223e4567-e89b-12d3-a456-426614174000", resResp.ID)
}

func TestDocumentTools_UpdateDocument(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	now := time.Now()
	mockRepo.GetDocumentFunc = func(ctx context.Context, id string) (*domain.Document, error) {
		return &domain.Document{
			ID:        id,
			ProjectID: "123e4567-e89b-12d3-a456-426614174000",
			Title:     "Old",
			Content:   "Old content",
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	mockRepo.UpdateDocumentFunc = func(ctx context.Context, d *domain.Document) (*domain.Document, error) {
		return &domain.Document{
			ID:        d.ID,
			ProjectID: d.ProjectID,
			Title:     d.Title,
			Content:   d.Content,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	args := json.RawMessage(`{"id":"223e4567-e89b-12d3-a456-426614174000", "title": "Updated", "content": "Updated content"}`)
	toolHandler, ok := registry.GetTool("update_document")
	require.True(t, ok)
	res, err := toolHandler(context.Background(), args)
	require.NoError(t, err)

	resResp, ok := res.(handler.DocumentResponse)
	require.True(t, ok)
	assert.Equal(t, "Updated", resResp.Title)
	assert.Equal(t, "Updated content", resResp.Content)
}

func TestDocumentTools_DeleteDocument(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	mockRepo.DeleteDocumentFunc = func(ctx context.Context, id string) error {
		return nil
	}

	args := json.RawMessage(`{"id":"223e4567-e89b-12d3-a456-426614174000"}`)
	toolHandler, ok := registry.GetTool("delete_document")
	require.True(t, ok)
	res, err := toolHandler(context.Background(), args)
	require.NoError(t, err)

	resMap, ok := res.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, resMap["success"])
}

func TestDocumentTools_ListDocuments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	now := time.Now()
	mockRepo.ListDocumentsFunc = func(ctx context.Context, projectID string) ([]*domain.Document, error) {
		assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", projectID)
		return []*domain.Document{
			{
				ID:        "223e4567-e89b-12d3-a456-426614174000",
				ProjectID: projectID,
				Title:     "D1",
				Content:   "C1",
				CreatedAt: now,
				UpdatedAt: now,
			},
		}, nil
	}

	args := json.RawMessage(`{"projectId":"123e4567-e89b-12d3-a456-426614174000"}`)
	toolHandler, ok := registry.GetTool("list_documents")
	require.True(t, ok)
	res, err := toolHandler(context.Background(), args)
	require.NoError(t, err)

	resMap, ok := res.(map[string]interface{})
	require.True(t, ok)
	docs, ok := resMap["documents"].([]handler.DocumentResponse)
	require.True(t, ok)
	assert.Len(t, docs, 1)
	assert.Equal(t, "223e4567-e89b-12d3-a456-426614174000", docs[0].ID)
}

// UT-001 — TestRegisterDocumentTools_RegistersAllFiveTools
func TestRegisterDocumentTools_RegistersAllFiveTools(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	for _, name := range []string{"create_document", "get_document", "update_document", "delete_document", "list_documents"} {
		tool, ok := registry.GetTool(name)
		assert.True(t, ok, "expected tool %q to be registered", name)
		assert.NotNil(t, tool, "expected tool %q to be non-nil", name)
	}

	tool, ok := registry.GetTool("nonexistent_tool")
	assert.False(t, ok, "expected nonexistent_tool to be absent")
	assert.Nil(t, tool, "expected nonexistent_tool to return nil handler")
}

// UT-002 — TestCreateDocumentTool_InvalidArguments
func TestCreateDocumentTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	tool, ok := registry.GetTool("create_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage("not-valid-json"))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-003 — TestCreateDocumentTool_MissingProjectIDOrTitle
func TestCreateDocumentTool_MissingProjectIDOrTitle(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	tool, ok := registry.GetTool("create_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"projectId":"","title":""}`))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "projectId and title are required")
}

// UT-004 — TestCreateDocumentTool_RepoError
func TestCreateDocumentTool_RepoError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	mockErr := errors.New("db down")
	mockRepo.CreateDocumentFunc = func(_ context.Context, _ *domain.Document) (*domain.Document, error) {
		return nil, mockErr
	}

	tool, ok := registry.GetTool("create_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"projectId":"proj-1","title":"My Doc"}`))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create document:")
}

// UT-005 — TestGetDocumentTool_InvalidArguments
func TestGetDocumentTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	tool, ok := registry.GetTool("get_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage("not-valid-json"))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-006 — TestGetDocumentTool_EmptyID
func TestGetDocumentTool_EmptyID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	tool, ok := registry.GetTool("get_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id":""}`))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

// UT-007 — TestGetDocumentTool_NotFound
func TestGetDocumentTool_NotFound(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	mockRepo.GetDocumentFunc = func(_ context.Context, _ string) (*domain.Document, error) {
		return nil, repo.ErrNotFound
	}

	tool, ok := registry.GetTool("get_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id":"doc-1"}`))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document not found")
	assert.False(t, errors.Is(err, repo.ErrNotFound))
}

// UT-008 — TestGetDocumentTool_GenericError
func TestGetDocumentTool_GenericError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	mockRepo.GetDocumentFunc = func(_ context.Context, _ string) (*domain.Document, error) {
		return nil, errors.New("db down")
	}

	tool, ok := registry.GetTool("get_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id":"doc-1"}`))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get document:")
}

// UT-009 — TestUpdateDocumentTool_InvalidArguments
func TestUpdateDocumentTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	tool, ok := registry.GetTool("update_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage("not-valid-json"))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-010 — TestUpdateDocumentTool_EmptyID
func TestUpdateDocumentTool_EmptyID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	tool, ok := registry.GetTool("update_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id":""}`))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

// UT-011 — TestUpdateDocumentTool_NotFoundOnInitialGet
func TestUpdateDocumentTool_NotFoundOnInitialGet(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	mockRepo.GetDocumentFunc = func(_ context.Context, _ string) (*domain.Document, error) {
		return nil, repo.ErrNotFound
	}

	tool, ok := registry.GetTool("update_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id":"doc-1"}`))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document not found")
	assert.False(t, errors.Is(err, repo.ErrNotFound))
}

// UT-012 — TestUpdateDocumentTool_GenericErrorOnInitialGet
func TestUpdateDocumentTool_GenericErrorOnInitialGet(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	mockRepo.GetDocumentFunc = func(_ context.Context, _ string) (*domain.Document, error) {
		return nil, errors.New("db down")
	}

	tool, ok := registry.GetTool("update_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id":"doc-1"}`))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get document:")
}

// UT-013 — TestUpdateDocumentTool_UpdateRepoError
func TestUpdateDocumentTool_UpdateRepoError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	now := time.Now()
	mockRepo.GetDocumentFunc = func(_ context.Context, id string) (*domain.Document, error) {
		return &domain.Document{
			ID:        id,
			ProjectID: "proj-1",
			Title:     "My Doc",
			Content:   "Some content",
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}
	mockRepo.UpdateDocumentFunc = func(_ context.Context, _ *domain.Document) (*domain.Document, error) {
		return nil, errors.New("db down")
	}

	tool, ok := registry.GetTool("update_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id":"doc-1"}`))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update document:")
}

// UT-014 — TestDeleteDocumentTool_InvalidArguments
func TestDeleteDocumentTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	tool, ok := registry.GetTool("delete_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage("not-valid-json"))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-015 — TestDeleteDocumentTool_EmptyID
func TestDeleteDocumentTool_EmptyID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	tool, ok := registry.GetTool("delete_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id":""}`))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

// UT-016 — TestDeleteDocumentTool_RepoError
func TestDeleteDocumentTool_RepoError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	mockRepo.DeleteDocumentFunc = func(_ context.Context, _ string) error {
		return errors.New("db down")
	}

	tool, ok := registry.GetTool("delete_document")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id":"doc-1"}`))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete document:")
}

// UT-017 — TestListDocumentsTool_InvalidArguments
func TestListDocumentsTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	tool, ok := registry.GetTool("list_documents")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage("not-valid-json"))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-018 — TestListDocumentsTool_MissingProjectID
func TestListDocumentsTool_MissingProjectID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	tool, ok := registry.GetTool("list_documents")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"projectId":""}`))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "projectId")
}

// UT-019 — TestListDocumentsTool_RepoError
func TestListDocumentsTool_RepoError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	mockRepo.ListDocumentsFunc = func(_ context.Context, _ string) ([]*domain.Document, error) {
		return nil, errors.New("db down")
	}

	tool, ok := registry.GetTool("list_documents")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"projectId":"proj-1"}`))

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list documents:")
}

// -----------------------------------------------------------------------
// US045 BREAKING CHANGE tests — create_document now requires requirement_id
// -----------------------------------------------------------------------

// UT-045-042 — MCP create_document now includes requirement_id in INSERT
func TestCreateDocumentTool_WithRequirementID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockReqRepo := &MockRequirementRepo{}
	mockDocRepo := &MockDocumentRepo{}
	now := time.Now()

	projectID := "11111111-1111-1111-1111-111111111111"
	requirementID := "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f"

	mockReqRepo.GetRequirementFunc = func(_ context.Context, id string) (*domain.Requirement, error) {
		return &domain.Requirement{
			ID:        id,
			ProjectID: projectID,
			Name:      "REQ",
			Status:    "draft",
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	mockDocRepo.CreateDocumentFunc = func(_ context.Context, d *domain.Document) (*domain.Document, error) {
		assert.Equal(t, requirementID, d.RequirementID, "RequirementID must be set on INSERT")
		return &domain.Document{
			ID:            "cccccccc-cccc-cccc-cccc-cccccccccccc",
			ProjectID:     d.ProjectID,
			RequirementID: d.RequirementID,
			Title:         d.Title,
			Content:       d.Content,
			CreatedAt:     now,
			UpdatedAt:     now,
		}, nil
	}

	handler.RegisterDocumentTools(registry, mockDocRepo, mockReqRepo)

	args := json.RawMessage(`{"projectId":"11111111-1111-1111-1111-111111111111","requirement_id":"b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f","title":"README","content":"# README\n..."}`)
	tool, ok := registry.GetTool("create_document")
	require.True(t, ok)

	res, err := tool(context.Background(), args)
	require.NoError(t, err)

	b, err := json.Marshal(res)
	require.NoError(t, err)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, requirementID, m["requirementId"])
	assert.Equal(t, projectID, m["projectId"])
	assert.Equal(t, "README", m["title"])
}

// UT-045-043 — MCP create_document — missing requirement_id returns tool error
func TestCreateDocumentTool_MissingRequirementID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockReqRepo := &MockRequirementRepo{}
	mockDocRepo := &MockDocumentRepo{}

	handler.RegisterDocumentTools(registry, mockDocRepo, mockReqRepo)

	tool, ok := registry.GetTool("create_document")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"projectId":"proj-1","title":"Doc"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requirement_id")
}

// UT-045-044 — MCP create_document — requirement not in project returns tool error
func TestCreateDocumentTool_RequirementNotInProject(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockReqRepo := &MockRequirementRepo{}
	mockDocRepo := &MockDocumentRepo{}
	now := time.Now()

	mockReqRepo.GetRequirementFunc = func(_ context.Context, id string) (*domain.Requirement, error) {
		return &domain.Requirement{
			ID:        id,
			ProjectID: "different-project-id",
			Name:      "REQ",
			Status:    "draft",
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	called := false
	mockDocRepo.CreateDocumentFunc = func(_ context.Context, _ *domain.Document) (*domain.Document, error) {
		called = true
		return nil, errors.New("should not be called")
	}

	handler.RegisterDocumentTools(registry, mockDocRepo, mockReqRepo)

	args := json.RawMessage(`{"projectId":"proj-1","requirement_id":"req-belongs-to-other","title":"Doc"}`)
	tool, ok := registry.GetTool("create_document")
	require.True(t, ok)

	_, err := tool(context.Background(), args)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requirement does not belong to project")
	assert.False(t, called)
}

// UT-020 — TestListDocumentsTool_EmptySliceReturnsEmptyDocumentsArray
func TestListDocumentsTool_EmptySliceReturnsEmptyDocumentsArray(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockDocumentRepo{}
	handler.RegisterDocumentTools(registry, mockRepo)

	mockRepo.ListDocumentsFunc = func(_ context.Context, _ string) ([]*domain.Document, error) {
		return nil, nil
	}

	tool, ok := registry.GetTool("list_documents")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"projectId":"proj-1"}`))

	assert.NoError(t, err)
	require.NotNil(t, result)

	resMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	docsVal, exists := resMap["documents"]
	assert.True(t, exists, "expected 'documents' key in result")
	assert.NotNil(t, docsVal, "expected 'documents' value to be non-nil")

	docs, ok := docsVal.([]handler.DocumentResponse)
	require.True(t, ok, "expected 'documents' to be []handler.DocumentResponse")
	assert.Len(t, docs, 0, "expected empty documents slice")
}
