package mcp_test

import (
	"context"
	"testing"
	"time"

	"agent-board/internal/mcp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionCreationAndMessageQueuing(t *testing.T) {
	// UT-001: Session creation and message queuing
	manager := mcp.NewSessionManager()

	session := manager.CreateSession()
	require.NotEmpty(t, session.ID, "Session should have a unique ID")

	retrievedSession, ok := manager.GetSession(session.ID)
	require.True(t, ok, "Session should be retrievable by ID")
	assert.Equal(t, session.ID, retrievedSession.ID)

	message := []byte(`{"jsonrpc": "2.0", "id": 1, "result": {"success": true}}`)
	err := session.QueueMessage(message)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	received, err := session.ReceiveMessage(ctx)
	require.NoError(t, err)
	assert.Equal(t, message, received)
}
