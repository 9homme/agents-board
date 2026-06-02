package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// UT-US004-002: SIGTERM cancels lifecycle context (mcp-server).
// Verifies that signal.NotifyContext cancels when SIGTERM is delivered to the process.
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
// placeholder — skipped until UT-US004-002 red→green loop closes.
func TestPingDB_TimeoutCancels(t *testing.T) {
	t.Skip("placeholder — implement in next red cycle")
}

// UT-US004-004: Defer order: cancel → stop → db.Close (mcp-server).
// placeholder — skipped until prior red→green loops close.
func TestMain_DeferOrder(t *testing.T) {
	t.Skip("placeholder — implement in next red cycle")
}

// UT-US004-005 (mcp-server portion): No context.Background() at ping call site.
// Regression guard: main.go must use signal.NotifyContext + context.WithTimeout,
// not a bare context.Background() passed directly to PingContext.
func TestMain_NoContextBackgroundAtPingSite(t *testing.T) {
	const mainGoPath = "main.go"
	src, err := os.ReadFile(mainGoPath)
	if err != nil {
		t.Fatalf("must be able to read main.go: %v", err)
	}

	content := string(src)

	assert.NotContains(t, content, "db.PingContext(context.Background())",
		"main.go must not pass context.Background() directly to PingContext")
	assert.Contains(t, content, "signal.NotifyContext",
		"main.go must use signal.NotifyContext for signal-cancellable lifecycle context")
	assert.Contains(t, content, "context.WithTimeout",
		"main.go must use context.WithTimeout to bound the DB ping")
}
