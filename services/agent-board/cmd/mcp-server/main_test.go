package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
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

// IT-002 — TestRun_LogsDBConfigLine_BeforePing (mcp-server)
// Verifies the startup log line "db config: using DATABASE_URL" is emitted
// before the DB ping attempt, even when the ping itself fails (no real DB).
// Architecture cite: architecture.md §5.3; §5.7 approach (b)
func TestRun_LogsDBConfigLine_BeforePing(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test_db_nonexistent")
	os.Unsetenv("DB_URL") //nolint:errcheck

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	// run() will log "db config: using DATABASE_URL" then fail on DB ping — expected.
	_ = run()

	output := buf.String()
	assert.Contains(t, output, "db config: using DATABASE_URL",
		"startup log line must appear before DB ping attempt")
}

// IT-003 — TestRun_HardFail_WhenOnlyDBURLSet (mcp-server subprocess hard-fail regression)
// Spawns the mcp-server binary with DB_URL set and DATABASE_URL absent.
// Asserts non-zero exit code and error message containing rename instructions.
// Architecture cite: architecture.md §5.7 optional hard-fail regression test; §5.4
func TestRun_HardFail_WhenOnlyDBURLSet(t *testing.T) {
	cmd := exec.Command("go", "run", ".")
	// Build env: inherit current env, add DB_URL, strip DATABASE_URL
	filteredEnv := make([]string, 0, len(os.Environ()))
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "DATABASE_URL=") {
			filteredEnv = append(filteredEnv, e)
		}
	}
	filteredEnv = append(filteredEnv, "DB_URL=postgres://x")
	cmd.Env = filteredEnv

	output, err := cmd.CombinedOutput()

	assert.Error(t, err, "process must exit non-zero when only DB_URL is set")
	combinedOut := string(output)
	assert.Contains(t, combinedOut, "DB_URL is no longer supported",
		"stderr must contain rename instruction")
	assert.Contains(t, combinedOut, "rename to DATABASE_URL",
		"stderr must contain rename target")
}
