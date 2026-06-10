// Package fsutil provides filesystem utility helpers for the agent-board service.
package fsutil

import (
	"errors"
	"os"
)

// ErrInvalidPath is returned when a path does not exist on disk or is not a directory.
var ErrInvalidPath = errors.New("path does not exist or is not a directory")

// ValidatePath verifies that path is non-blank, exists on disk, and is a directory.
// Returns ErrInvalidPath if any of those conditions fail.
func ValidatePath(path string) error {
	if path == "" {
		return ErrInvalidPath
	}

	info, err := os.Stat(path)
	if err != nil {
		return ErrInvalidPath
	}

	if !info.IsDir() {
		return ErrInvalidPath
	}

	return nil
}
