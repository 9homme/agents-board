package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"agent-board-mcp/internal/domain"
	"agent-board-mcp/internal/mcp"
	"agent-board-mcp/internal/repo"
)

// UserStoryResponse defines the JSON structure for a user story response.
type UserStoryResponse struct {
	ID          string `json:"id"`
	ProjectID   string `json:"projectId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func toUserStoryResponse(u *domain.UserStory) UserStoryResponse {
	return UserStoryResponse{
		ID:          u.ID,
		ProjectID:   u.ProjectID,
		Title:       u.Title,
		Description: u.Description,
		Status:      u.Status,
		CreatedAt:   u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   u.UpdatedAt.Format(time.RFC3339),
	}
}

// RegisterUserStoryTools registers user story MCP tools.
func RegisterUserStoryTools(registry *mcp.ToolRegistry, repository repo.UserStoryRepository) {
	registry.RegisterTool("create_user_story", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ProjectID   string `json:"projectId"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Status      string `json:"status"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		if req.ProjectID == "" || req.Title == "" || req.Status == "" {
			return nil, fmt.Errorf("missing required fields")
		}

		u := &domain.UserStory{
			ProjectID:   req.ProjectID,
			Title:       req.Title,
			Description: req.Description,
			Status:      req.Status,
		}

		created, err := repository.CreateUserStory(ctx, u)
		if err != nil {
			return nil, err
		}
		return toUserStoryResponse(created), nil
	})

	registry.RegisterTool("get_user_story", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		if req.ID == "" {
			return nil, fmt.Errorf("missing id")
		}

		u, err := repository.GetUserStory(ctx, req.ID)
		if err != nil {
			if err == repo.ErrNotFound {
				return nil, fmt.Errorf("user story not found")
			}
			return nil, err
		}
		return toUserStoryResponse(u), nil
	})

	registry.RegisterTool("update_user_story", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ID          string  `json:"id"`
			Title       *string `json:"title"`
			Description *string `json:"description"`
			Status      *string `json:"status"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		if req.ID == "" {
			return nil, fmt.Errorf("missing id")
		}

		u, err := repository.GetUserStory(ctx, req.ID)
		if err != nil {
			if err == repo.ErrNotFound {
				return nil, fmt.Errorf("user story not found")
			}
			return nil, err
		}

		if req.Title != nil {
			u.Title = *req.Title
		}
		if req.Description != nil {
			u.Description = *req.Description
		}
		if req.Status != nil {
			u.Status = *req.Status
		}

		updated, err := repository.UpdateUserStory(ctx, u)
		if err != nil {
			return nil, err
		}
		return toUserStoryResponse(updated), nil
	})

	registry.RegisterTool("delete_user_story", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		if req.ID == "" {
			return nil, fmt.Errorf("missing id")
		}

		err := repository.DeleteUserStory(ctx, req.ID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"success": true}, nil
	})

	registry.RegisterTool("list_user_stories", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ProjectID string `json:"projectId"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		if req.ProjectID == "" {
			return nil, fmt.Errorf("missing projectId")
		}

		userStories, err := repository.ListUserStories(ctx, req.ProjectID)
		if err != nil {
			return nil, err
		}

		responses := make([]UserStoryResponse, 0, len(userStories))
		for _, u := range userStories {
			responses = append(responses, toUserStoryResponse(u))
		}

		return map[string]interface{}{"userStories": responses}, nil
	})
}
