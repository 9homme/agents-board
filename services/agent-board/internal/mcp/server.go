package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/google/uuid"
)

// Session represents an active SSE connection.
type Session struct {
	ID       string
	messages chan []byte
}

// QueueMessage adds a message to the session's send queue.
func (s *Session) QueueMessage(msg []byte) error {
	select {
	case s.messages <- msg:
		return nil
	default:
		return errors.New("message queue full")
	}
}

// ReceiveMessage waits for a message from the queue.
func (s *Session) ReceiveMessage(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-s.messages:
		return msg, nil
	}
}

// SessionManager manages active SSE sessions.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates and registers a new session.
func (m *SessionManager) CreateSession() *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.New().String()
	session := &Session{
		ID:       id,
		messages: make(chan []byte, 100), // Buffer size of 100 messages
	}
	m.sessions[id] = session
	return session
}

// GetSession retrieves a session by ID.
func (m *SessionManager) GetSession(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[id]
	return session, ok
}

// RemoveSession removes a session by ID.
func (m *SessionManager) RemoveSession(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, id)
}

// ToolHandler is a function that implements a specific tool logic.
type ToolHandler func(ctx context.Context, args json.RawMessage) (interface{}, error)

// ToolRegistry manages registered MCP tools.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]ToolHandler
}

// NewToolRegistry creates a new ToolRegistry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]ToolHandler),
	}
}

// RegisterTool adds a new tool to the registry.
func (r *ToolRegistry) RegisterTool(name string, handler ToolHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools[name] = handler
}

// GetTool retrieves a tool handler by name.
func (r *ToolRegistry) GetTool(name string) (ToolHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, ok := r.tools[name]
	return handler, ok
}
