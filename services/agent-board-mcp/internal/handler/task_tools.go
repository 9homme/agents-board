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

// TaskResponse represents the exact JSON shape for a task
type TaskResponse struct {
	ID          string `json:"id"`
	UserStoryID string `json:"userStoryId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func mapTaskToResponse(t *domain.Task) TaskResponse {
	return TaskResponse{
		ID:          t.ID,
		UserStoryID: t.UserStoryID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
	}
}

// RegisterTaskTools registers all task-related tools in the provided registry.
func RegisterTaskTools(registry *mcp.ToolRegistry, repository repo.TaskRepository) {
	registry.RegisterTool("create_task", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			UserStoryID string `json:"userStoryId"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Status      string `json:"status"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		if req.UserStoryID == "" || req.Title == "" {
			return nil, errors.New("userStoryId and title are required")
		}

		if req.Status == "" {
			req.Status = "pending"
		}

		task := &domain.Task{
			UserStoryID: req.UserStoryID,
			Title:       req.Title,
			Description: req.Description,
			Status:      req.Status,
		}

		created, err := repository.CreateTask(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("failed to create task: %w", err)
		}

		return mapTaskToResponse(created), nil
	})

	registry.RegisterTool("get_task", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		if req.ID == "" {
			return nil, errors.New("id is required")
		}

		task, err := repository.GetTask(ctx, req.ID)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return nil, fmt.Errorf("task not found")
			}
			return nil, fmt.Errorf("failed to get task: %w", err)
		}

		return mapTaskToResponse(task), nil
	})

	registry.RegisterTool("update_task", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
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
			return nil, errors.New("id is required")
		}

		// First fetch the existing task
		existing, err := repository.GetTask(ctx, req.ID)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return nil, fmt.Errorf("task not found")
			}
			return nil, fmt.Errorf("failed to get task: %w", err)
		}

		if req.Title != nil {
			existing.Title = *req.Title
		}
		if req.Description != nil {
			existing.Description = *req.Description
		}
		if req.Status != nil {
			existing.Status = *req.Status
		}

		updated, err := repository.UpdateTask(ctx, existing)
		if err != nil {
			return nil, fmt.Errorf("failed to update task: %w", err)
		}

		return mapTaskToResponse(updated), nil
	})

	registry.RegisterTool("delete_task", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		if req.ID == "" {
			return nil, errors.New("id is required")
		}

		err := repository.DeleteTask(ctx, req.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete task: %w", err)
		}

		return map[string]interface{}{"success": true}, nil
	})

	registry.RegisterTool("list_tasks", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			UserStoryID string `json:"userStoryId"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		if req.UserStoryID == "" {
			return nil, errors.New("userStoryId is required")
		}

		tasks, err := repository.ListTasks(ctx, req.UserStoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to list tasks: %w", err)
		}

		taskResponses := make([]TaskResponse, len(tasks))
		for i, t := range tasks {
			taskResponses[i] = mapTaskToResponse(t)
		}

		// Per architecture, returns {"tasks": [...]}
		return map[string]interface{}{"tasks": taskResponses}, nil
	})
}
