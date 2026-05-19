package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"agent-board/internal/domain"
	"agent-board/internal/mcp"
	"agent-board/internal/repo"
)

// AuditLogResponse is the exact JSON shape for a status audit log entry as specified
// in the architecture API contract for get_task_audit_trail and get_user_story_audit_trail.
type AuditLogResponse struct {
	ID         string `json:"id"`
	EntityID   string `json:"entityId"`
	EntityType string `json:"entityType"`
	FromStatus string `json:"fromStatus"`
	ToStatus   string `json:"toStatus"`
	ChangedAt  string `json:"changedAt"`
}

func mapAuditLogToResponse(entry *domain.StatusAuditLog) AuditLogResponse {
	return AuditLogResponse{
		ID:         entry.ID,
		EntityID:   entry.EntityID,
		EntityType: entry.EntityType,
		FromStatus: entry.FromStatus,
		ToStatus:   entry.ToStatus,
		ChangedAt:  entry.ChangedAt.Format(time.RFC3339),
	}
}

// RegisterAuditTools registers the get_task_audit_trail and get_user_story_audit_trail
// MCP tools in the provided registry using the given AuditRepository.
func RegisterAuditTools(registry *mcp.ToolRegistry, repository repo.AuditRepository) {
	registry.RegisterTool("get_task_audit_trail", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			TaskID string `json:"taskId"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		if req.TaskID == "" {
			return nil, errors.New("taskId is required")
		}

		entries, err := repository.GetTaskAuditTrail(ctx, req.TaskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task audit trail: %w", err)
		}

		responses := make([]AuditLogResponse, len(entries))
		for i, entry := range entries {
			responses[i] = mapAuditLogToResponse(entry)
		}

		return map[string]interface{}{"auditTrail": responses}, nil
	})

	registry.RegisterTool("get_user_story_audit_trail", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var req struct {
			UserStoryID string `json:"userStoryId"`
		}
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		if req.UserStoryID == "" {
			return nil, errors.New("userStoryId is required")
		}

		entries, err := repository.GetUserStoryAuditTrail(ctx, req.UserStoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user story audit trail: %w", err)
		}

		responses := make([]AuditLogResponse, len(entries))
		for i, entry := range entries {
			responses[i] = mapAuditLogToResponse(entry)
		}

		return map[string]interface{}{"auditTrail": responses}, nil
	})
}
