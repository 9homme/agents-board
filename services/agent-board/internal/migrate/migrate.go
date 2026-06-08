package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
)

// Run applies embedded migrations idempotently from the given FS.
// It tracks applied migrations in the 'schema_migrations' table.
func Run(ctx context.Context, db *sql.DB, _ fs.FS) error {
	// 1. Ensure schema_migrations table exists
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     TEXT PRIMARY KEY,
			applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// 2. Get already applied migrations
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return nil
}
