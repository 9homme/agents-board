package handler

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"agent-board/internal/domain"
	"agent-board/internal/fsutil"
	"agent-board/internal/mcp"
	"agent-board/internal/repo"
)

// PathValidator abstracts filesystem path validation so it can be mocked in tests.
type PathValidator interface {
	// ValidatePath checks that path is non-blank, exists on disk, and is a directory.
	ValidatePath(path string) error
}

// fsutilValidator is the production PathValidator backed by the fsutil package.
type fsutilValidator struct{}

func (fsutilValidator) ValidatePath(path string) error {
	return fsutil.ValidatePath(path)
}

// RegisterProjectTools registers project-related tools to the given registry.
// validator is used to validate the path field on create_project (D-008).
// Passing nil uses the real filesystem validator backed by the fsutil package.
func RegisterProjectTools(registry *mcp.ToolRegistry, projectRepo repo.ProjectRepository, validator PathValidator) {
	if validator == nil {
		validator = fsutilValidator{}
	}
	registry.RegisterTool("create_project", handleCreateProject(projectRepo, validator))
	registry.RegisterTool("get_project", handleGetProject(projectRepo))
	registry.RegisterTool("update_project", handleUpdateProject(projectRepo))
	registry.RegisterTool("delete_project", handleDeleteProject(projectRepo))
	registry.RegisterTool("list_projects", handleListProjects(projectRepo))
}

func handleCreateProject(projectRepo repo.ProjectRepository, validator PathValidator) mcp.ToolHandler {
	return func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Path        string `json:"path"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, errors.New("invalid arguments")
		}

		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			return nil, errors.New("name is required and cannot be empty")
		}

		req.Path = strings.TrimSpace(req.Path)
		if req.Path == "" {
			return nil, errors.New("path is required")
		}

		if err := validator.ValidatePath(req.Path); err != nil {
			return nil, errors.New("path does not exist or is not a directory")
		}

		p := &domain.Project{
			Name:        req.Name,
			Description: req.Description,
			Path:        req.Path,
		}

		created, err := projectRepo.CreateProject(ctx, p)
		if err != nil {
			if errors.Is(err, repo.ErrDuplicatePath) {
				return nil, errors.New("path already linked to another project")
			}
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
