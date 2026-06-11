package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"agent-board/internal/domain"
	"agent-board/internal/mcp"
	"agent-board/internal/repo"
)

// RequirementResponse defines the JSON structure for a requirement response (§5).
type RequirementResponse struct {
	ID          string `json:"id"`
	ProjectID   string `json:"projectId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func toRequirementResponse(r *domain.Requirement) RequirementResponse {
	return RequirementResponse{
		ID:          r.ID,
		ProjectID:   r.ProjectID,
		Name:        r.Name,
		Description: r.Description,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   r.UpdatedAt.Format(time.RFC3339),
	}
}

// validRequirementStatus reports whether s is one of the allowed status values.
func validRequirementStatus(s string) bool {
	return s == domain.RequirementStatusDraft ||
		s == domain.RequirementStatusInProgress ||
		s == domain.RequirementStatusDone
}

// RegisterRequirementTools registers the create_requirement, list_requirements, and
// update_requirement MCP tools into the provided registry.
func RegisterRequirementTools(registry *mcp.ToolRegistry, requirementRepo repo.RequirementRepository) {
	registry.RegisterTool("create_requirement", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ProjectID   string `json:"project_id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Status      string `json:"status"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			return nil, fmt.Errorf("name is required and must be non-blank")
		}

		if req.Status == "" {
			req.Status = domain.RequirementStatusDraft
		}
		if !validRequirementStatus(req.Status) {
			return nil, fmt.Errorf("invalid status %q: must be one of draft, in_progress, done", req.Status)
		}

		r := &domain.Requirement{
			ProjectID:   req.ProjectID,
			Name:        req.Name,
			Description: req.Description,
			Status:      req.Status,
		}

		created, err := requirementRepo.Create(ctx, r)
		if err != nil {
			if errors.Is(err, repo.ErrProjectNotFound) {
				return nil, fmt.Errorf("project not found")
			}
			return nil, fmt.Errorf("failed to create requirement: %w", err)
		}

		resp := toRequirementResponse(created)
		return resp, nil
	})

	registry.RegisterTool("list_requirements", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			ProjectID string `json:"project_id"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		reqs, err := requirementRepo.ListByProject(ctx, req.ProjectID)
		if err != nil {
			if errors.Is(err, repo.ErrProjectNotFound) {
				return nil, fmt.Errorf("project not found")
			}
			return nil, fmt.Errorf("failed to list requirements: %w", err)
		}

		responses := make([]RequirementResponse, 0, len(reqs))
		for i := range reqs {
			responses = append(responses, toRequirementResponse(&reqs[i]))
		}

		return map[string]interface{}{"requirements": responses}, nil
	})

	registry.RegisterTool("update_requirement", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			RequirementID string  `json:"requirement_id"`
			Status        *string `json:"status"`
			Name          *string `json:"name"`
			Description   *string `json:"description"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		// Validate status if provided
		if req.Status != nil && !validRequirementStatus(*req.Status) {
			return nil, fmt.Errorf("invalid status %q: must be one of draft, in_progress, done", *req.Status)
		}

		// Validate name if provided: trim and check non-blank
		if req.Name != nil {
			trimmed := strings.TrimSpace(*req.Name)
			if trimmed == "" {
				return nil, fmt.Errorf("name must be non-blank when provided")
			}
			*req.Name = trimmed
		}

		patch := repo.RequirementPatch{
			Status:      req.Status,
			Name:        req.Name,
			Description: req.Description,
		}

		updated, err := requirementRepo.Update(ctx, req.RequirementID, patch)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return nil, fmt.Errorf("requirement not found")
			}
			return nil, fmt.Errorf("failed to update requirement: %w", err)
		}

		return toRequirementResponse(updated), nil
	})
}
