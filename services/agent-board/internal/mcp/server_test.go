package mcp_test

// server_test.go — US009 ToolRegistry + Session + SessionManager tests
// OQ-4 tech-debt note: ListTools doc-comment claims "lexicographic order" but the
// implementation does NOT sort (iterates a map). The TestToolRegistry_ListTools_ReturnsAllRegisteredNames
// test intentionally uses assert.ElementsMatch (unordered) rather than assert.Equal on a sorted
// slice. This mismatch should be resolved in a follow-up: either sort in production code or
// update the doc-comment. See architecture.md §4.4 special case + §13.1 R-5.

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"agent-board/internal/mcp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noop is a reusable no-op ToolHandler for tests that do not need real logic.
func noop(_ context.Context, _ json.RawMessage) (interface{}, error) { return nil, nil }

// UT-001
func TestNewToolRegistry_ReturnsEmptyRegistry(t *testing.T) {
	registry := mcp.NewToolRegistry()
	names := registry.ListTools()
	assert.Empty(t, names)
}

// UT-002
func TestToolRegistry_RegisterTool_AddsHandler(t *testing.T) {
	registry := mcp.NewToolRegistry()
	registry.RegisterTool("my_tool", noop)
	h, ok := registry.GetTool("my_tool")
	assert.True(t, ok)
	assert.NotNil(t, h)
}

// UT-003
func TestToolRegistry_RegisterTool_OverwritesPriorHandler(t *testing.T) {
	registry := mcp.NewToolRegistry()
	first := func(_ context.Context, _ json.RawMessage) (interface{}, error) { return "first", nil }
	second := func(_ context.Context, _ json.RawMessage) (interface{}, error) { return "second", nil }
	registry.RegisterTool("tool", first)
	registry.RegisterTool("tool", second)
	h, ok := registry.GetTool("tool")
	require.True(t, ok)
	result, _ := h(context.Background(), nil)
	assert.Equal(t, "second", result)
}

// UT-004
func TestToolRegistry_GetTool_UnknownNameReturnsFalse(t *testing.T) {
	registry := mcp.NewToolRegistry()
	h, ok := registry.GetTool("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, h)
}

// UT-005
func TestToolRegistry_GetTool_KnownNameReturnsHandler(t *testing.T) {
	registry := mcp.NewToolRegistry()
	registry.RegisterTool("greet", func(_ context.Context, _ json.RawMessage) (interface{}, error) {
		return "hello", nil
	})
	h, ok := registry.GetTool("greet")
	require.True(t, ok)
	result, err := h(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, "hello", result)
}

// UT-006
// OQ-4: ListTools doc-comment says "lexicographic order" but map iteration is unordered.
// Using ElementsMatch (unordered membership) intentionally — do NOT change to sorted assert.Equal.
func TestToolRegistry_ListTools_ReturnsAllRegisteredNames(t *testing.T) {
	registry := mcp.NewToolRegistry()
	registry.RegisterTool("alpha", noop)
	registry.RegisterTool("beta", noop)
	registry.RegisterTool("gamma", noop)
	names := registry.ListTools()
	assert.Len(t, names, 3)
	assert.ElementsMatch(t, []string{"alpha", "beta", "gamma"}, names)
}

// UT-007
func TestToolRegistry_ListTools_EmptyAfterNoRegistrations(t *testing.T) {
	registry := mcp.NewToolRegistry()
	names := registry.ListTools()
	assert.Empty(t, names)
}

// UT-008 — concurrent RegisterTool + GetTool + ListTools; must be race-clean under go test -race.
func TestToolRegistry_ConcurrentRegisterAndGet(t *testing.T) {
	registry := mcp.NewToolRegistry()
	const N = 100
	var wg sync.WaitGroup
	wg.Add(N * 3)

	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			registry.RegisterTool("tool", noop)
		}(i)
	}
	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			registry.GetTool("tool")
		}(i)
	}
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			registry.ListTools()
		}()
	}
	wg.Wait()
}

// UT-009
func TestSession_QueueMessage_HappyPath(t *testing.T) {
	sm := mcp.NewSessionManager()
	sess := sm.CreateSession()
	msg := []byte(`{"test": true}`)

	err1 := sess.QueueMessage(msg)
	require.NoError(t, err1)

	received, err2 := sess.ReceiveMessage(context.Background())
	assert.NoError(t, err2)
	assert.Equal(t, msg, received)
}

// UT-010 — fill channel to capacity (100), then assert 101st call errors.
func TestSession_QueueMessage_FullReturnsError(t *testing.T) {
	sm := mcp.NewSessionManager()
	sess := sm.CreateSession()

	for i := 0; i < 100; i++ {
		require.NoError(t, sess.QueueMessage([]byte("filler")))
	}

	err := sess.QueueMessage([]byte("overflow"))
	require.Error(t, err)
	assert.Equal(t, "message queue full", err.Error())
}

// UT-011
func TestSession_ReceiveMessage_HappyPath(t *testing.T) {
	sm := mcp.NewSessionManager()
	sess := sm.CreateSession()

	require.NoError(t, sess.QueueMessage([]byte("hello")))

	msg, err := sess.ReceiveMessage(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello"), msg)
}

// UT-012 — context cancelled before ReceiveMessage.
func TestSession_ReceiveMessage_ContextCancelled(t *testing.T) {
	sm := mcp.NewSessionManager()
	sess := sm.CreateSession()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel BEFORE calling ReceiveMessage

	msg, err := sess.ReceiveMessage(ctx)
	assert.Nil(t, msg)
	assert.True(t, errors.Is(err, context.Canceled))
}

// UT-013
func TestSessionManager_RemoveSession_RemovesSession(t *testing.T) {
	sm := mcp.NewSessionManager()
	sess := sm.CreateSession()
	sessID := sess.ID
	sm.RemoveSession(sessID)
	_, ok := sm.GetSession(sessID)
	assert.False(t, ok)
}

// UT-014
func TestSessionManager_RemoveSession_UnknownIDIsNoop(t *testing.T) {
	sm := mcp.NewSessionManager()
	sess := sm.CreateSession()

	// Must not panic on unknown ID.
	sm.RemoveSession("completely-nonexistent-id")

	// The existing session must still be present.
	_, ok := sm.GetSession(sess.ID)
	assert.True(t, ok)
}

// UT-015 — concurrent CreateSession + RemoveSession; must be race-clean under go test -race.
func TestSessionManager_RemoveSession_ConcurrentSafe(t *testing.T) {
	sm := mcp.NewSessionManager()
	const N = 100
	var wg sync.WaitGroup
	wg.Add(N * 2)

	// Pre-create N sessions so the remove goroutines have IDs to work with.
	ids := make([]string, N)
	for i := 0; i < N; i++ {
		ids[i] = sm.CreateSession().ID
	}

	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			sm.CreateSession()
		}()
	}
	for i := 0; i < N; i++ {
		go func(id string) {
			defer wg.Done()
			sm.RemoveSession(id)
		}(ids[i])
	}
	wg.Wait()
}
