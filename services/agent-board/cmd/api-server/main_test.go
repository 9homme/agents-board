package main

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UT-US004-API-001: DB ping timeout cancels PingContext.
// When the context deadline has already passed, pingDB must return an error
// wrapping context.DeadlineExceeded.
func TestPingDB_TimeoutCancels(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectPing()

	// Use a context whose deadline is already in the past — sqlmock returns
	// context.DeadlineExceeded immediately when the context is already done.
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Millisecond))
	defer cancel()

	err = pingDB(ctx, db)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded), "expected DeadlineExceeded, got: %v", err)
}

// UT-US004-API-002: Signal cancellation propagates to PingContext.
// Cancelling the parent context must cause pingDB to return context.Canceled.
func TestPingDB_CancellationPropagates(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectPing()

	// Pre-cancel the context to simulate signal arriving before ping.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = pingDB(ctx, db)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled), "expected Canceled, got: %v", err)
}

// UT-US004-API-003: Happy path — ping succeeds immediately.
// pingDB must return nil when PingContext succeeds and the context is valid.
func TestPingDB_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectPing().WillReturnError(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = pingDB(ctx, db)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-US004-API-004: No context.Background() at ping call site (structural).
// The production main.go must not contain a bare context.Background() call,
// and must contain signal.NotifyContext and context.WithTimeout.
func TestMain_NoContextBackgroundAtPingSite(t *testing.T) {
	const mainGoPath = "main.go"
	src, err := os.ReadFile(mainGoPath)
	require.NoError(t, err, "must be able to read main.go")

	content := string(src)

	assert.NotContains(t, content, "db.PingContext(context.Background())",
		"main.go must not pass context.Background() directly to PingContext")
	assert.Contains(t, content, "signal.NotifyContext",
		"main.go must use signal.NotifyContext for signal-cancellable lifecycle context")
	assert.Contains(t, content, "context.WithTimeout",
		"main.go must use context.WithTimeout to bound the DB ping")
}
