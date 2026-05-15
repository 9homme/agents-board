package handler

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"agent-board/internal/domain"
	"agent-board/internal/mcp"
	"agent-board/internal/repo"
)

// RegisterProjectTools registers project-related tools to the given registry.
func RegisterProjectTools(registry *mcp.ToolRegistry, projectRepo repo.ProjectRepository) {
	registry.RegisterTool("create_project", handleCreateProject(projectRepo))
	registry.RegisterTool("get_project", handleGetProject(projectRepo))
	registry.RegisterTool("update_project", handleUpdateProject(projectRepo))
	registry.RegisterTool("delete_project", handleDeleteProject(projectRepo))
	registry.RegisterTool("list_projects", handleListProjects(projectRepo))
}

func handleCreateProject(projectRepo repo.ProjectRepository) mcp.ToolHandler {
	return func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, errors.New("invalid arguments")
		}

		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			return nil, errors.New("name is required and cannot be empty")
		}

		p := &domain.Project{
			Name:        req.Name,
			Description: req.Description,
		}

		created, err := projectRepo.CreateProject(ctx, p)
		if err != nil {
			return nil, err
		}

		return created, nil
	}
}

func handleGetProject(projectRepo repo.ProjectRepository) mcp.ToolHandler {
	return func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, errors.New("invalid arguments")
		}
		if req.ID == "" {
			return nil, errors.New("id is required")
		}

		p, err := projectRepo.GetProject(ctx, req.ID)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return nil, errors.New("project not found")
			}
			return nil, err
		}

		return p, nil
	}
}

func handleUpdateProject(projectRepo repo.ProjectRepository) mcp.ToolHandler {
	return func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ID          string  `json:"id"`
			Name        *string `json:"name"`
			Description *string `json:"description"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, errors.New("invalid arguments")
		}
		if req.ID == "" {
			return nil, errors.New("id is required")
		}

		p, err := projectRepo.GetProject(ctx, req.ID)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return nil, errors.New("project not found")
			}
			return nil, err
		}

		if req.Name != nil {
			name := strings.TrimSpace(*req.Name)
			if name == "" {
				return nil, errors.New("name cannot be empty if provided")
			}
			p.Name = name
		}
		if req.Description != nil {
			p.Description = *req.Description
		}

		updated, err := projectRepo.UpdateProject(ctx, p)
		if err != nil {
			return nil, err
		}

		return updated, nil
	}
}

func handleDeleteProject(projectRepo repo.ProjectRepository) mcp.ToolHandler {
	return func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, errors.New("invalid arguments")
		}
		if req.ID == "" {
			return nil, errors.New("id is required")
		}

		err := projectRepo.DeleteProject(ctx, req.ID)
		if err != nil {
			return nil, err
		}

		return map[string]bool{"success": true}, nil
	}
}

func handleListProjects(projectRepo repo.ProjectRepository) mcp.ToolHandler {
	return func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		projects, err := projectRepo.ListProjects(ctx)
		if err != nil {
			return nil, err
		}

		// Ensure we don't return null for empty lists in JSON
		if projects == nil {
			projects = make([]*domain.Project, 0)
		}

		return map[string]interface{}{
			"projects": projects,
		}, nil
	}
}
