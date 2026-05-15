package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"agent-board-mcp/internal/handler"
	"agent-board-mcp/internal/mcp"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSSEEndpoint(t *testing.T) {
	// UT-002: GET /sse endpoint
	e := echo.New()
	manager := mcp.NewSessionManager()
	h := handler.NewHandler(manager, mcp.NewToolRegistry())

	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	errCh := make(chan error, 1)
	go func() {
		errCh <- h.HandleSSE(c)
	}()

	time.Sleep(50 * time.Millisecond) // Allow SSE headers to be flushed

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))

	body := rec.Body.String()
	assert.Contains(t, body, "event: endpoint\n")
	assert.Contains(t, body, "data: /message?sessionId=")

	// Extract sessionId
	lines := strings.Split(body, "\n")
	var sessionId string
	for _, line := range lines {
		if strings.HasPrefix(line, "data: /message?sessionId=") {
			sessionId = strings.TrimPrefix(line, "data: /message?sessionId=")
			break
		}
	}
	require.NotEmpty(t, sessionId)

	_, ok := manager.GetSession(sessionId)
	assert.True(t, ok)
}

func TestPostMessageInvalidJSONRPC(t *testing.T) {
	// UT-003: POST /message with invalid JSON-RPC
	e := echo.New()
	manager := mcp.NewSessionManager()
	h := handler.NewHandler(manager, mcp.NewToolRegistry())

	session := manager.CreateSession()

	invalidJSON := []byte(`{invalid}`)
	req := httptest.NewRequest(http.MethodPost, "/message?sessionId="+session.ID, bytes.NewReader(invalidJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HandleMessage(c)
	assert.NoError(t, err) // the handler itself doesn't error out, it responds with an error
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostMessageValidToolCall(t *testing.T) {
	// UT-004: POST /message with valid tool call
	e := echo.New()
	manager := mcp.NewSessionManager()
	registry := mcp.NewToolRegistry()
	h := handler.NewHandler(manager, registry)

	// Mock a tool
	registry.RegisterTool("test_tool", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		return map[string]interface{}{"success": true}, nil
	})

	session := manager.CreateSession()

	reqPayload := mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: mcp.ToolCallParams{
			Name: "test_tool",
			Arguments: json.RawMessage(`{}`),
		},
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	req := httptest.NewRequest(http.MethodPost, "/message?sessionId="+session.ID, bytes.NewReader(bodyBytes))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.HandleMessage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify the response is queued in the session
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	received, err := session.ReceiveMessage(ctx)
	require.NoError(t, err)

	var resPayload mcp.JSONRPCResponse
	err = json.Unmarshal(received, &resPayload)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resPayload.ID)
	assert.NotNil(t, resPayload.Result)
	assert.Contains(t, string(resPayload.Result.Content[0].Text), `"success":true`)
}

func TestITFullHandshake(t *testing.T) {
	// IT-001: Full handshake: SSE + POST
	e := echo.New()
	manager := mcp.NewSessionManager()
	registry := mcp.NewToolRegistry()
	h := handler.NewHandler(manager, registry)

	registry.RegisterTool("hello", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		return map[string]string{"msg": "world"}, nil
	})

	// 1. GET /sse
	reqSSE := httptest.NewRequest(http.MethodGet, "/sse", nil)
	recSSE := httptest.NewRecorder()
	cSSE := e.NewContext(reqSSE, recSSE)

	go func() {
		_ = h.HandleSSE(cSSE)
	}()

	time.Sleep(50 * time.Millisecond) // Allow endpoint event to be sent

	bodySSE := recSSE.Body.String()
	lines := strings.Split(bodySSE, "\n")
	var sessionId string
	for _, line := range lines {
		if strings.HasPrefix(line, "data: /message?sessionId=") {
			sessionId = strings.TrimPrefix(line, "data: /message?sessionId=")
			break
		}
	}
	require.NotEmpty(t, sessionId)

	// 2. POST /message
	reqPayload := mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params: mcp.ToolCallParams{
			Name: "hello",
			Arguments: json.RawMessage(`{}`),
		},
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	reqMsg := httptest.NewRequest(http.MethodPost, "/message?sessionId="+sessionId, bytes.NewReader(bodyBytes))
	reqMsg.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recMsg := httptest.NewRecorder()
	cMsg := e.NewContext(reqMsg, recMsg)

	err := h.HandleMessage(cMsg)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, recMsg.Code)

	// Wait a little for the SSE stream to receive the response
	time.Sleep(50 * time.Millisecond)

	bodySSEUpdated := recSSE.Body.String()
	assert.Contains(t, bodySSEUpdated, "event: message\n")
	assert.Contains(t, bodySSEUpdated, "world")
}
