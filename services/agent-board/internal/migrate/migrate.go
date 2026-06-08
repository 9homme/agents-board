package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
)

// Run applies embedded migrations idempotently from the given FS.
// It tracks applied migrations in the 'schema_migrations' table.
func Run(ctx context.Context, db *sql.DB, fsys fs.FS) error {
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

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = true
	}

	// 3. Read migration files from FS
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != "" {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	// 4. Apply pending migrations in lexical order
	for _, file := range files {
		if applied[file] {
			continue
		}

		content, err := fs.ReadFile(fsys, file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		err = func() error {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return err
			}
			defer func() { _ = tx.Rollback() }()

			if _, err := tx.ExecContext(ctx, string(content)); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", file, err)
			}

			return tx.Commit()
		}()

		if err != nil {
			return err
		}
	}

	return nil
}
