package fsutil_test

import (
	"os"
	"testing"

	"agent-board/internal/fsutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UT-045-017 — ValidatePath: exists and is a directory (happy path)
func TestValidatePath_ExistsAndIsDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "fsutil-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(dir) }()

	err = fsutil.ValidatePath(dir)
	assert.NoError(t, err)
}

// UT-045-018 — ValidatePath: path does not exist on disk
func TestValidatePath_DoesNotExist(t *testing.T) {
	err := fsutil.ValidatePath("/tmp/this-does-not-exist-xxxxxxxxxxx-fsutil-test")
	assert.ErrorIs(t, err, fsutil.ErrInvalidPath)
}

// UT-045-019 — ValidatePath: path exists but is a regular file
func TestValidatePath_ExistsButIsFile(t *testing.T) {
	f, err := os.CreateTemp("", "fsutil-test-file-*")
	require.NoError(t, err)
	_ = f.Close()
	defer func() { _ = os.Remove(f.Name()) }()

	err = fsutil.ValidatePath(f.Name())
	assert.ErrorIs(t, err, fsutil.ErrInvalidPath)
}

// UT-045-020 — ValidatePath: empty path returns ErrInvalidPath immediately (no syscall)
func TestValidatePath_EmptyPath(t *testing.T) {
	err := fsutil.ValidatePath("")
	assert.ErrorIs(t, err, fsutil.ErrInvalidPath)
}
