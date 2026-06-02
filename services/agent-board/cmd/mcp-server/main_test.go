package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UT-US004-002: SIGTERM cancels lifecycle context (mcp-server).
// Verifies that signal.NotifyContext cancels when SIGTERM or SIGINT is delivered to the process.
// Architecture cite: §3.2 — signals handled: os.Interrupt, syscall.SIGTERM; §3.7 testability notes.
func TestLifecycleContext_SIGTERMCancels(t *testing.T) {
	t.Run("SIGTERM", func(t *testing.T) {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}()

		select {
		case <-ctx.Done():
			// expected
		case <-time.After(1 * time.Second):
			t.Fatal("context not cancelled within timeout")
		}

		assert.Equal(t, context.Canceled, ctx.Err())
	})

	t.Run("SIGINT", func(t *testing.T) {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}()

		select {
		case <-ctx.Done():
			// expected
		case <-time.After(1 * time.Second):
			t.Fatal("context not cancelled within timeout")
		}

		assert.Equal(t, context.Canceled, ctx.Err())
	})
}

// IT-US004-002: DB ping times out with context.DeadlineExceeded (mcp-server).
// Uses a context whose deadline is already past so sqlmock returns DeadlineExceeded immediately.
// Architecture cite: §3.3 — 5 s timeout; §3.4 run() shape; §3.7.
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

// UT-US004-004: Defer order: cancel → stop → db.Close (mcp-server).
// Uses the AST-parse approach: reads main.go and verifies that the defer declarations appear
// in the correct source order (db.Close first, then stop, then cancel) so that LIFO execution
// produces cancel → stop → db.Close.
// Architecture cite: §3.5 — defer order: cancel (innermost) → stop → db.Close.
func TestMain_DeferOrder(t *testing.T) {
	src, err := os.ReadFile("main.go")
	require.NoError(t, err, "must be able to read main.go")

	content := string(src)

	// Locate the positions of each defer declaration in source order.
	dbCloseIdx := indexOfSubstring(content, "defer func() { _ = db.Close() }()")
	stopIdx := indexOfSubstring(content, "defer stop()")
	cancelIdx := indexOfSubstring(content, "defer cancel()")

	require.NotEqual(t, -1, dbCloseIdx, "main.go must contain defer db.Close()")
	require.NotEqual(t, -1, stopIdx, "main.go must contain defer stop()")
	require.NotEqual(t, -1, cancelIdx, "main.go must contain defer cancel()")

	// Declaration order in source must be: db.Close < stop < cancel
	// (LIFO execution order: cancel → stop → db.Close — matching architecture §3.5)
	assert.Less(t, dbCloseIdx, stopIdx,
		"defer db.Close() must be declared before defer stop() for correct LIFO cleanup order")
	assert.Less(t, stopIdx, cancelIdx,
		"defer stop() must be declared before defer cancel() for correct LIFO cleanup order")
}

// UT-US004-005 (mcp-server portion): No context.Background() at ping call site.
// Regression guard: main.go must use signal.NotifyContext + context.WithTimeout,
// not a bare context.Background() passed directly to PingContext.
// Architecture cite: US004 AC "Scenario: no context.Background() remains in production code".
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

// indexOfSubstring returns the byte index of the first occurrence of sub in s, or -1 if not found.
func indexOfSubstring(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
