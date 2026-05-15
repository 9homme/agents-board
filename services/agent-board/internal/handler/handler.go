package handler

import (
	"agent-board/internal/mcp"
)

// Handler handles HTTP requests for the MCP server.
type Handler struct {
	sessionManager *mcp.SessionManager
	toolRegistry   *mcp.ToolRegistry
}

// NewHandler creates a new Handler.
func NewHandler(sm *mcp.SessionManager, tr *mcp.ToolRegistry) *Handler {
	return &Handler{
		sessionManager: sm,
		toolRegistry:   tr,
	}
}
