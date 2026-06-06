// Package handler_test exercises HandleMessage, sendError, and sendToolResultError
// defined in message.go (US008 — architecture.md §8 / D-010).
//
// All 13 test functions match the verbatim names from US008 AC / US008_be_unit_tests.md.
// Harness: httptest + Echo for HandleMessage; direct call via handler_internal_test.go
// re-exports for sendError / sendToolResultError (US008 AC items 10-13).
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"agent-board/internal/handler"
	"agent-board/internal/mcp"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test harness helpers (architecture.md §8.2)
// ---------------------------------------------------------------------------

// newTestHandler constructs a Handler with a real SessionManager, a real
// ToolRegistry, and a pre-created session for immediate use in tests.
func newTestHandler(t *testing.T) (*handler.Handler, *mcp.Session, *mcp.ToolRegistry) {
	t.Helper()
	sm := mcp.NewSessionManager()
	tr := mcp.NewToolRegistry()
	h := handler.NewHandler(sm, tr)
	sess := sm.CreateSession()
	return h, sess, tr
}

// postMessage submits a POST /message request via Echo's ServeHTTP so that all
// Echo middleware (error handler included) participates — mirrors production flow.
func postMessage(t *testing.T, h *handler.Handler, sessionID string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	e.POST("/message", h.HandleMessage)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/message?sessionId="+sessionID, bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// validToolsCallBody returns a minimal tools/call JSON-RPC body for the named tool.
func validToolsCallBody(toolName string) []byte {
	return validToolsCallBodyWithID(toolName, 1)
}

// validToolsCallBodyWithID returns a tools/call body with an explicit JSON-RPC ID.
func validToolsCallBodyWithID(toolName string, id int) []byte {
	req := mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "tools/call",
		Params: mcp.ToolCallParams{
			Name:      toolName,
			Arguments: json.RawMessage(`{}`),
		},
	}
	b, err := json.Marshal(req)
	if err != nil {
		panic("validToolsCallBody: json.Marshal failed: " + err.Error())
	}
	return b
}

// drainQueueMessages reads up to n messages from the session without blocking.
func drainQueueMessages(t *testing.T, sess *mcp.Session, n int) [][]byte {
	t.Helper()
	var out [][]byte
	for i := 0; i < n; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 50*1e6) // 50 ms
		msg, err := sess.ReceiveMessage(ctx)
		cancel()
		if err != nil {
			break
		}
		out = append(out, msg)
	}
	return out
}

// prefillQueue fills the session's 100-slot message channel to capacity so that
// the next QueueMessage call will fail with "message queue full" (architecture.md §8.4).
func prefillQueue(t *testing.T, sess *mcp.Session) {
	t.Helper()
	for i := 0; i < 100; i++ {
		require.NoError(t, sess.QueueMessage([]byte("filler")))
	}
}

// ---------------------------------------------------------------------------
// UT-001 — TestHandleMessage_MissingSessionID
// ---------------------------------------------------------------------------

// TestHandleMessage_MissingSessionID verifies that omitting sessionId returns 400
// with the "sessionId is required" error message.
func TestHandleMessage_MissingSessionID(t *testing.T) {
	h, _, _ := newTestHandler(t)

	validBody := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"any","arguments":{}}}`)
	rec := postMessage(t, h, "", validBody)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "sessionId is required")
}

// ---------------------------------------------------------------------------
// UT-002 — TestHandleMessage_InvalidSessionID
// ---------------------------------------------------------------------------

// TestHandleMessage_InvalidSessionID verifies that an unknown sessionId returns
// 400 with "invalid sessionId".
func TestHandleMessage_InvalidSessionID(t *testing.T) {
	h, _, _ := newTestHandler(t)

	validBody := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"any","arguments":{}}}`)
	rec := postMessage(t, h, "does-not-exist", validBody)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid sessionId")
}

// ---------------------------------------------------------------------------
// UT-003 — TestHandleMessage_InvalidJSONPayload
// ---------------------------------------------------------------------------

// TestHandleMessage_InvalidJSONPayload verifies that a non-JSON body returns 400
// with "invalid JSON-RPC payload".
func TestHandleMessage_InvalidJSONPayload(t *testing.T) {
	h, sess, _ := newTestHandler(t)

	rec := postMessage(t, h, sess.ID, []byte("not-json"))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid JSON-RPC payload")
}

// ---------------------------------------------------------------------------
// UT-004 — TestHandleMessage_NonToolsCallMethod
// ---------------------------------------------------------------------------

// TestHandleMessage_NonToolsCallMethod verifies that a valid JSON-RPC 2.0 body
// with a non-"tools/call" method routes to sendError (InvalidRequest) and the
// session queue receives the error response.
func TestHandleMessage_NonToolsCallMethod(t *testing.T) {
	h, sess, _ := newTestHandler(t)

	body := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	rec := postMessage(t, h, sess.ID, body)

	// sendError returns echo.NewHTTPError(200, ...) — Echo serialises as 200.
	assert.Equal(t, http.StatusOK, rec.Code)

	// The session queue must contain one message with the InvalidRequest error.
	msgs := drainQueueMessages(t, sess, 1)
	require.Len(t, msgs, 1)

	var resp mcp.JSONRPCResponse
	require.NoError(t, json.Unmarshal(msgs[0], &resp))
	require.NotNil(t, resp.Error)
	assert.Equal(t, mcp.InvalidRequest, resp.Error.Code)
	assert.Contains(t, resp.Error.Message, "Invalid request")
}

// ---------------------------------------------------------------------------
// UT-005 — TestHandleMessage_WrongJSONRPCVersion
// ---------------------------------------------------------------------------

// TestHandleMessage_WrongJSONRPCVersion verifies that a body with jsonrpc "1.0"
// routes to sendError (InvalidRequest) — same path as UT-004.
func TestHandleMessage_WrongJSONRPCVersion(t *testing.T) {
	h, sess, _ := newTestHandler(t)

	body := []byte(`{"jsonrpc":"1.0","id":2,"method":"tools/call","params":{"name":"any","arguments":{}}}`)
	rec := postMessage(t, h, sess.ID, body)

	assert.Equal(t, http.StatusOK, rec.Code)

	msgs := drainQueueMessages(t, sess, 1)
	require.Len(t, msgs, 1)

	var resp mcp.JSONRPCResponse
	require.NoError(t, json.Unmarshal(msgs[0], &resp))
	require.NotNil(t, resp.Error)
	assert.Equal(t, mcp.InvalidRequest, resp.Error.Code)
}

// ---------------------------------------------------------------------------
// UT-006 — TestHandleMessage_ToolNotFound
// ---------------------------------------------------------------------------

// TestHandleMessage_ToolNotFound verifies that requesting an unregistered tool
// routes to sendToolResultError and the response contains "Tool not found".
func TestHandleMessage_ToolNotFound(t *testing.T) {
	h, sess, _ := newTestHandler(t)
	// registry is empty — no tools registered

	rec := postMessage(t, h, sess.ID, validToolsCallBody("nonexistent_tool"))

	assert.Equal(t, http.StatusOK, rec.Code)

	msgs := drainQueueMessages(t, sess, 1)
	require.Len(t, msgs, 1)

	var resp mcp.JSONRPCResponse
	require.NoError(t, json.Unmarshal(msgs[0], &resp))
	require.NotNil(t, resp.Result)
	assert.True(t, resp.Result.IsError)
	require.Len(t, resp.Result.Content, 1)
	assert.Contains(t, resp.Result.Content[0].Text, "Tool not found")
}

// ---------------------------------------------------------------------------
// UT-007 — TestHandleMessage_ToolExecutionError
// ---------------------------------------------------------------------------

// TestHandleMessage_ToolExecutionError verifies that a tool returning an error
// routes to sendToolResultError and the response surfaces the error message.
func TestHandleMessage_ToolExecutionError(t *testing.T) {
	h, sess, tr := newTestHandler(t)
	tr.RegisterTool("my_tool", func(_ context.Context, _ json.RawMessage) (interface{}, error) {
		return nil, errors.New("boom")
	})

	rec := postMessage(t, h, sess.ID, validToolsCallBody("my_tool"))

	assert.Equal(t, http.StatusOK, rec.Code)

	msgs := drainQueueMessages(t, sess, 1)
	require.Len(t, msgs, 1)

	var resp mcp.JSONRPCResponse
	require.NoError(t, json.Unmarshal(msgs[0], &resp))
	require.NotNil(t, resp.Result)
	assert.True(t, resp.Result.IsError)
	require.Len(t, resp.Result.Content, 1)
	assert.Contains(t, resp.Result.Content[0].Text, "boom")
}

// ---------------------------------------------------------------------------
// UT-008 — TestHandleMessage_HappyPath
// ---------------------------------------------------------------------------

// TestHandleMessage_HappyPath verifies a successful tool call: HTTP 200, one
// message queued, result carries the tool's marshalled return value.
func TestHandleMessage_HappyPath(t *testing.T) {
	h, sess, tr := newTestHandler(t)
	tr.RegisterTool("echo_tool", func(_ context.Context, _ json.RawMessage) (interface{}, error) {
		return map[string]string{"answer": "42"}, nil
	})

	rec := postMessage(t, h, sess.ID, validToolsCallBodyWithID("echo_tool", 99))

	assert.Equal(t, http.StatusOK, rec.Code)

	msgs := drainQueueMessages(t, sess, 1)
	require.Len(t, msgs, 1, "expected exactly one message queued in the session")

	var resp mcp.JSONRPCResponse
	require.NoError(t, json.Unmarshal(msgs[0], &resp))
	assert.Equal(t, float64(99), resp.ID) // JSON numbers unmarshal to float64 via interface{}
	require.NotNil(t, resp.Result)
	require.Len(t, resp.Result.Content, 1)
	assert.Equal(t, "text", resp.Result.Content[0].Type)
	assert.Contains(t, resp.Result.Content[0].Text, `"answer":"42"`)
}

// ---------------------------------------------------------------------------
// UT-009 — TestHandleMessage_QueueMessageFails
// ---------------------------------------------------------------------------

// TestHandleMessage_QueueMessageFails verifies that when the session queue is
// already full the handler returns HTTP 500 with "failed to queue message".
func TestHandleMessage_QueueMessageFails(t *testing.T) {
	h, sess, tr := newTestHandler(t)
	prefillQueue(t, sess)

	tr.RegisterTool("any_tool", func(_ context.Context, _ json.RawMessage) (interface{}, error) {
		return map[string]string{"ok": "yes"}, nil
	})

	rec := postMessage(t, h, sess.ID, validToolsCallBody("any_tool"))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "failed to queue message")
}

// ---------------------------------------------------------------------------
// Optional coverage chase — HandleMessage json.Marshal fallback (message.go:46)
// ---------------------------------------------------------------------------

// TestHandleMessage_NonMarshalableToolResult verifies that when a registered tool
// returns a value that json.Marshal cannot serialise (e.g. chan int), HandleMessage
// falls through to sendError with mcp.InternalError. This drives message.go:46
// (architecture.md §8.5 — "optional to chase").
func TestHandleMessage_NonMarshalableToolResult(t *testing.T) {
	h, sess, tr := newTestHandler(t)
	tr.RegisterTool("bad_tool", func(_ context.Context, _ json.RawMessage) (interface{}, error) {
		return make(chan int), nil // chan int cannot be JSON-marshalled
	})

	rec := postMessage(t, h, sess.ID, validToolsCallBody("bad_tool"))

	// sendError returns echo.HTTPError(200, ...) — Echo wraps it as 200.
	assert.Equal(t, http.StatusOK, rec.Code)

	msgs := drainQueueMessages(t, sess, 1)
	require.Len(t, msgs, 1)

	var resp mcp.JSONRPCResponse
	require.NoError(t, json.Unmarshal(msgs[0], &resp))
	require.NotNil(t, resp.Error)
	assert.Equal(t, mcp.InternalError, resp.Error.Code)
}

// ---------------------------------------------------------------------------
// UT-010 — TestSendError_QueuesAndReturnsEchoError
// ---------------------------------------------------------------------------

// TestSendError_QueuesAndReturnsEchoError calls sendError directly (via the
// handler_internal_test.go re-export) and verifies that it queues a
// JSONRPCResponse with the expected error fields AND returns *echo.HTTPError
// with Code 200.
func TestSendError_QueuesAndReturnsEchoError(t *testing.T) {
	h, sess, _ := newTestHandler(t)

	result := handler.SendError(h, sess, 5, mcp.InvalidRequest, "Invalid request")

	// Must return a non-nil echo.HTTPError with status 200.
	require.NotNil(t, result)
	var echoErr *echo.HTTPError
	require.True(t, errors.As(result, &echoErr), "expected *echo.HTTPError, got %T", result)
	assert.Equal(t, http.StatusOK, echoErr.Code)

	// Session queue must contain exactly one message.
	msgs := drainQueueMessages(t, sess, 1)
	require.Len(t, msgs, 1)

	var resp mcp.JSONRPCResponse
	require.NoError(t, json.Unmarshal(msgs[0], &resp))
	assert.Equal(t, float64(5), resp.ID)
	require.NotNil(t, resp.Error)
	assert.Equal(t, mcp.InvalidRequest, resp.Error.Code)
	assert.Equal(t, "Invalid request", resp.Error.Message)
}

// ---------------------------------------------------------------------------
// UT-011 — TestSendToolResultError_QueuesAndReturnsEchoError
// ---------------------------------------------------------------------------

// TestSendToolResultError_QueuesAndReturnsEchoError calls sendToolResultError
// directly and verifies the queued message and the returned *echo.HTTPError.
func TestSendToolResultError_QueuesAndReturnsEchoError(t *testing.T) {
	h, sess, _ := newTestHandler(t)

	result := handler.SendToolResultError(h, sess, 7, "boom")

	require.NotNil(t, result)
	var echoErr *echo.HTTPError
	require.True(t, errors.As(result, &echoErr), "expected *echo.HTTPError, got %T", result)
	assert.Equal(t, http.StatusOK, echoErr.Code)

	msgs := drainQueueMessages(t, sess, 1)
	require.Len(t, msgs, 1)

	var resp mcp.JSONRPCResponse
	require.NoError(t, json.Unmarshal(msgs[0], &resp))
	assert.Equal(t, float64(7), resp.ID)
	require.NotNil(t, resp.Result)
	assert.True(t, resp.Result.IsError)
	require.Len(t, resp.Result.Content, 1)
	assert.Equal(t, "text", resp.Result.Content[0].Type)
	assert.Equal(t, "boom", resp.Result.Content[0].Text)
}

// ---------------------------------------------------------------------------
// UT-012 — TestSendError_QueueFailure_LogsButReturnsEchoError
// ---------------------------------------------------------------------------

// TestSendError_QueueFailure_LogsButReturnsEchoError verifies that when the
// session queue is full, sendError still returns a non-nil *echo.HTTPError
// (it does NOT swallow the error). The log.Printf path (message.go:89) is
// exercised — confirmed by coverage.
func TestSendError_QueueFailure_LogsButReturnsEchoError(t *testing.T) {
	h, sess, _ := newTestHandler(t)
	prefillQueue(t, sess)

	result := handler.SendError(h, sess, 1, mcp.InvalidRequest, "msg")

	require.NotNil(t, result, "sendError must return non-nil even when queue is full")
	var echoErr *echo.HTTPError
	require.True(t, errors.As(result, &echoErr), "expected *echo.HTTPError, got %T", result)
	assert.Equal(t, http.StatusOK, echoErr.Code)
}

// ---------------------------------------------------------------------------
// UT-013 — TestSendToolResultError_QueueFailure_LogsButReturnsEchoError
// ---------------------------------------------------------------------------

// TestSendToolResultError_QueueFailure_LogsButReturnsEchoError verifies that
// when the session queue is full, sendToolResultError still returns a non-nil
// *echo.HTTPError. The log.Printf path (message.go:114) is exercised —
// confirmed by coverage.
func TestSendToolResultError_QueueFailure_LogsButReturnsEchoError(t *testing.T) {
	h, sess, _ := newTestHandler(t)
	prefillQueue(t, sess)

	result := handler.SendToolResultError(h, sess, 1, "msg")

	require.NotNil(t, result, "sendToolResultError must return non-nil even when queue is full")
	var echoErr *echo.HTTPError
	require.True(t, errors.As(result, &echoErr), "expected *echo.HTTPError, got %T", result)
	assert.Equal(t, http.StatusOK, echoErr.Code)
}
