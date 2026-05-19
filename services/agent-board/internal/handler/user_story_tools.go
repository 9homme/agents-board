package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"agent-board/internal/domain"
	"agent-board/internal/mcp"
	"agent-board/internal/repo"
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
		if req.ProjectID == "" || req.Title == "" {
			return nil, fmt.Errorf("missing required fields")
		}

		// Default to "draft" if no status provided; reject any non-draft initial status (UT-005, D-001).
		if req.Status == "" {
			req.Status = domain.UserStoryStatusDraft
		}
		u := &domain.UserStory{
			ProjectID:   req.ProjectID,
			Title:       req.Title,
			Description: req.Description,
			Status:      req.Status,
		}
		// Enforce initial state via domain constructor.
		if _, err := domain.NewUserStory(req.ProjectID, req.Title, req.Description, req.Status); err != nil {
			return nil, fmt.Errorf("invalid initial status: %w", err)
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

		existing, err := repository.GetUserStory(ctx, req.ID)
		if err != nil {
			if err == repo.ErrNotFound {
				return nil, fmt.Errorf("user story not found")
			}
			return nil, err
		}

		// If a status change is requested, validate the transition (IT-001, D-001, D-003).
		if req.Status != nil && *req.Status != existing.Status {
			if !existing.IsValidTransition(*req.Status) {
				return nil, fmt.Errorf("invalid transition from %s to %s", existing.Status, *req.Status)
			}
			// Perform the transactional status update + audit log insertion.
			updated, err := repository.UpdateUserStoryStatus(ctx, existing.ID, existing.Status, *req.Status)
			if err != nil {
				return nil, err
			}
			// Apply any non-status field updates on top of the status-updated entity.
			if req.Title != nil {
				updated.Title = *req.Title
			}
			if req.Description != nil {
				updated.Description = *req.Description
			}
			// If there are additional field changes, persist them.
			if req.Title != nil || req.Description != nil {
				saved, err := repository.UpdateUserStory(ctx, updated)
				if err != nil {
					return nil, err
				}
				return toUserStoryResponse(saved), nil
			}
			return toUserStoryResponse(updated), nil
		}

		// No status change — apply non-status field updates normally.
		if req.Title != nil {
			existing.Title = *req.Title
		}
		if req.Description != nil {
			existing.Description = *req.Description
		}

		updated, err := repository.UpdateUserStory(ctx, existing)
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
