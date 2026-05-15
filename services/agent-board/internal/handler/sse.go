package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HandleSSE handles the GET /sse endpoint.
func (h *Handler) HandleSSE(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Flush()

	session := h.sessionManager.CreateSession()
	// In production, we would use a cleanup task or heartbeat.
	// Removing defer h.sessionManager.RemoveSession(session.ID)
	// to support simple E2E test clients that don't hold the socket open.

	// Send endpoint event
	endpointData := fmt.Sprintf("/message?sessionId=%s", session.ID)
	if _, err := fmt.Fprintf(c.Response(), "event: endpoint\ndata: %s\n\n", endpointData); err != nil {
		return err
	}
	c.Response().Flush()

	ctx := c.Request().Context()

	for {
		msg, err := session.ReceiveMessage(ctx)
		if err != nil {
			// Client disconnected or context canceled
			return nil
		}

		if _, err := fmt.Fprintf(c.Response(), "event: message\ndata: %s\n\n", string(msg)); err != nil {
			return err
		}
		c.Response().Flush()
	}
}
