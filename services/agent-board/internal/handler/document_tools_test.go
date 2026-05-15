package handler_test

import (
	"context"
	"encoding/json"
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
