package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"agent-board-mcp/internal/domain"
	"agent-board-mcp/internal/mcp"
	"agent-board-mcp/internal/repo"
)

// DocumentResponse represents the exact JSON shape for a document
type DocumentResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func mapDocumentToResponse(d *domain.Document) DocumentResponse {
	return DocumentResponse{
		ID:        d.ID,
		ProjectID: d.ProjectID,
		Title:     d.Title,
		Content:   d.Content,
		CreatedAt: d.CreatedAt.Format(time.RFC3339),
		UpdatedAt: d.UpdatedAt.Format(time.RFC3339),
	}
}

// RegisterDocumentTools registers all document-related tools in the provided registry.
func RegisterDocumentTools(registry *mcp.ToolRegistry, repository repo.DocumentRepository) {
	registry.RegisterTool("create_document", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ProjectID string `json:"projectId"`
			Title     string `json:"title"`
			Content   string `json:"content"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		if req.ProjectID == "" || req.Title == "" {
			return nil, errors.New("projectId and title are required")
		}

		d := &domain.Document{
			ProjectID: req.ProjectID,
			Title:     req.Title,
			Content:   req.Content,
		}

		created, err := repository.CreateDocument(ctx, d)
		if err != nil {
			return nil, fmt.Errorf("failed to create document: %w", err)
		}

		return mapDocumentToResponse(created), nil
	})

	registry.RegisterTool("get_document", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		if req.ID == "" {
			return nil, errors.New("id is required")
		}

		d, err := repository.GetDocument(ctx, req.ID)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return nil, fmt.Errorf("document not found")
			}
			return nil, fmt.Errorf("failed to get document: %w", err)
		}

		return mapDocumentToResponse(d), nil
	})

	registry.RegisterTool("update_document", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ID      string  `json:"id"`
			Title   *string `json:"title"`
			Content *string `json:"content"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		if req.ID == "" {
			return nil, errors.New("id is required")
		}

		// First fetch the existing document
		existing, err := repository.GetDocument(ctx, req.ID)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return nil, fmt.Errorf("document not found")
			}
			return nil, fmt.Errorf("failed to get document: %w", err)
		}

		if req.Title != nil {
			existing.Title = *req.Title
		}
		if req.Content != nil {
			existing.Content = *req.Content
		}

		updated, err := repository.UpdateDocument(ctx, existing)
		if err != nil {
			return nil, fmt.Errorf("failed to update document: %w", err)
		}

		return mapDocumentToResponse(updated), nil
	})

	registry.RegisterTool("delete_document", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		if req.ID == "" {
			return nil, errors.New("id is required")
		}

		err := repository.DeleteDocument(ctx, req.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete document: %w", err)
		}

		return map[string]interface{}{"success": true}, nil
	})

	registry.RegisterTool("list_documents", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ProjectID string `json:"projectId"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		if req.ProjectID == "" {
			return nil, errors.New("projectId is required")
		}

		docs, err := repository.ListDocuments(ctx, req.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to list documents: %w", err)
		}

		docResponses := make([]DocumentResponse, len(docs))
		for i, d := range docs {
			docResponses[i] = mapDocumentToResponse(d)
		}

		// Per architecture, returns {"documents": [...]}
		return map[string]interface{}{"documents": docResponses}, nil
	})
}
