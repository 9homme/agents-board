package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"agent-board/internal/mcp"

	"github.com/labstack/echo/v4"
)

// HandleMessage handles the POST /message endpoint.
func (h *Handler) HandleMessage(c echo.Context) error {
	sessionId := c.QueryParam("sessionId")
	if sessionId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "sessionId is required"})
	}

	session, ok := h.sessionManager.GetSession(sessionId)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid sessionId"})
	}

	var req mcp.JSONRPCRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON-RPC payload"})
	}

	if req.JSONRPC != "2.0" || req.Method != "tools/call" {
		return h.sendError(session, req.ID, mcp.InvalidRequest, "Invalid request")
	}

	tool, ok := h.toolRegistry.GetTool(req.Params.Name)
	if !ok {
		return h.sendToolResultError(session, req.ID, "Tool not found")
	}

	result, err := tool(c.Request().Context(), req.Params.Arguments)
	if err != nil {
		return h.sendToolResultError(session, req.ID, err.Error())
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return h.sendError(session, req.ID, mcp.InternalError, "Internal error")
	}

	resp := mcp.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: &mcp.ToolResult{
			Content: []mcp.ToolContent{
				{
					Type: "text",
					Text: string(resultBytes),
				},
			},
		},
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return h.sendError(session, req.ID, mcp.InternalError, "Internal error")
	}

	if err := session.QueueMessage(respBytes); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to queue message"})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) sendError(session *mcp.Session, id interface{}, code int, message string) error {
	resp := mcp.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &mcp.JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
	respBytes, _ := json.Marshal(resp)
	if err := session.QueueMessage(respBytes); err != nil {
		log.Printf("failed to queue error message: %v\n", err)
	}
	// Return the error in the body as well
	return echo.NewHTTPError(http.StatusOK, resp)
}

func (h *Handler) sendToolResultError(session *mcp.Session, id interface{}, message string) error {
	resp := mcp.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: &mcp.ToolResult{
			IsError: true,
			Content: []mcp.ToolContent{
				{
					Type: "text",
					Text: message,
				},
			},
		},
	}
	respBytes, _ := json.Marshal(resp)
	if err := session.QueueMessage(respBytes); err != nil {
		log.Printf("failed to queue tool result error: %v\n", err)
	}
	// Return the result (which contains the tool error) in the body
	return echo.NewHTTPError(http.StatusOK, resp)
}
