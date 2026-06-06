package config_test

import (
	"os"
	"testing"

	"agent-board/internal/config"

	"github.com/stretchr/testify/assert"
)

// UT-001 — TestResolveDBURL_OnlyDatabaseURLSet_Happy
// DATABASE_URL set, DB_URL unset → (value, nil)
// Architecture cite: architecture.md §5.6 case 1; §5.4 happy-path row
func TestResolveDBURL_OnlyDatabaseURLSet_Happy(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	os.Unsetenv("DB_URL") //nolint:errcheck

	url, err := config.ResolveDBURL()

	assert.NoError(t, err)
	assert.Equal(t, "postgres://x", url)
}

// UT-002 — TestResolveDBURL_OnlyDBURLSet_RejectsWithRenameError
// DB_URL set, DATABASE_URL unset → ("", err) with exact wording
// Architecture cite: architecture.md §5.6 case 2; §5.4 Only DB_URL set row
func TestResolveDBURL_OnlyDBURLSet_RejectsWithRenameError(t *testing.T) {
	t.Setenv("DB_URL", "postgres://legacy")
	os.Unsetenv("DATABASE_URL") //nolint:errcheck

	url, err := config.ResolveDBURL()

	assert.Equal(t, "", url)
	assert.EqualError(t, err, "DB_URL is no longer supported; rename to DATABASE_URL (REQ006/US010)")
}

// UT-003 — TestResolveDBURL_BothSet_RejectsWithDisambiguateError
// Both DB_URL and DATABASE_URL set → ("", err) with exact wording
// Architecture cite: architecture.md §5.6 case 3; §5.4 Both set row
func TestResolveDBURL_BothSet_RejectsWithDisambiguateError(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://new")
	t.Setenv("DB_URL", "postgres://old")

	url, err := config.ResolveDBURL()

	assert.Equal(t, "", url)
	assert.EqualError(t, err, "DB_URL is set but no longer supported; remove DB_URL from your environment to disambiguate (DATABASE_URL is the sole accepted name as of REQ006/US010)")
}

// UT-004 — TestResolveDBURL_NeitherSet_RejectsWithRequiredError
// Both DB_URL and DATABASE_URL unset → ("", err) with exact wording
// Architecture cite: architecture.md §5.6 case 4; §5.4 Neither set row
func TestResolveDBURL_NeitherSet_RejectsWithRequiredError(t *testing.T) {
	os.Unsetenv("DATABASE_URL") //nolint:errcheck
	os.Unsetenv("DB_URL")       //nolint:errcheck

	url, err := config.ResolveDBURL()

	assert.Equal(t, "", url)
	assert.EqualError(t, err, "DATABASE_URL environment variable is required")
}
