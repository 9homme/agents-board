package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"agent-board/internal/handler"
	"agent-board/internal/mcp"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// safeResponseWriter is a thread-safe http.ResponseWriter used by tests that
// spawn the SSE handler in a goroutine and inspect the captured output from
// the main goroutine. It guards writes/reads with a mutex so the race
// detector does not flag overlapping access to the underlying buffer / header
// map / status code.
type safeResponseWriter struct {
	mu     sync.Mutex
	header http.Header
	body   bytes.Buffer
	code   int
}

func newSafeResponseWriter() *safeResponseWriter {
	return &safeResponseWriter{
		header: make(http.Header),
		code:   http.StatusOK,
	}
}

func (w *safeResponseWriter) Header() http.Header {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.header
}

func (w *safeResponseWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.body.Write(b)
}

func (w *safeResponseWriter) WriteHeader(statusCode int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.code = statusCode
}

// Flush implements http.Flusher so the SSE handler's c.Response().Flush()
// calls have something to dispatch to. There is no real buffering here.
func (w *safeResponseWriter) Flush() {}

func (w *safeResponseWriter) Code() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.code
}

func (w *safeResponseWriter) HeaderSnapshot() http.Header {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make(http.Header, len(w.header))
	for k, v := range w.header {
		vv := make([]string, len(v))
		copy(vv, v)
		out[k] = vv
	}
	return out
}

func (w *safeResponseWriter) BodyString() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.body.String()
}

func TestGetSSEEndpoint(t *testing.T) {
	// UT-002: GET /sse endpoint
	e := echo.New()
	manager := mcp.NewSessionManager()
	h := handler.NewHandler(manager, mcp.NewToolRegistry())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/sse", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	errCh := make(chan error, 1)
	go func() {
		errCh <- h.HandleSSE(c)
	}()

	time.Sleep(50 * time.Millisecond) // Allow SSE headers to be flushed

	// Signal the SSE goroutine to stop and wait for it to return before
	// reading the recorder. This avoids a data race between the goroutine's
	// writes to rec and the test's reads of rec.Code/Header()/Body.
	cancel()
	<-errCh

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
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/message?sessionId="+session.ID, bytes.NewReader(invalidJSON))
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
			Name:      "test_tool",
			Arguments: json.RawMessage(`{}`),
		},
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/message?sessionId="+session.ID, bytes.NewReader(bodyBytes))
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
	// Use a safeResponseWriter so the test can snapshot the SSE body while
	// the handler goroutine is still writing to it, without tripping the
	// race detector. Use a cancellable context so we can deterministically
	// join the goroutine at the end of the test.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reqSSE := httptest.NewRequestWithContext(ctx, http.MethodGet, "/sse", nil)
	wSSE := newSafeResponseWriter()
	cSSE := e.NewContext(reqSSE, wSSE)

	errCh := make(chan error, 1)
	go func() {
		errCh <- h.HandleSSE(cSSE)
	}()

	time.Sleep(50 * time.Millisecond) // Allow endpoint event to be sent

	bodySSE := wSSE.BodyString()
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
			Name:      "hello",
			Arguments: json.RawMessage(`{}`),
		},
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	reqMsg := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/message?sessionId="+sessionId, bytes.NewReader(bodyBytes))
	reqMsg.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recMsg := httptest.NewRecorder()
	cMsg := e.NewContext(reqMsg, recMsg)

	err := h.HandleMessage(cMsg)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, recMsg.Code)

	// Wait a little for the SSE stream to receive the response
	time.Sleep(50 * time.Millisecond)

	// Stop the SSE goroutine and wait for it to return before reading the
	// final body snapshot. (BodyString is mutex-protected, but joining here
	// also pins down the assertion to a stable point in the stream.)
	cancel()
	<-errCh

	bodySSEUpdated := wSSE.BodyString()
	assert.Contains(t, bodySSEUpdated, "event: message\n")
	assert.Contains(t, bodySSEUpdated, "world")
}
