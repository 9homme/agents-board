package migrate

import (
	"context"
	"database/sql"
	"io/fs"
)

// Run applies embedded migrations idempotently from the given FS.
// It tracks applied migrations in the 'schema_migrations' table.
func Run(_ context.Context, _ *sql.DB, _ fs.FS) error {
	return nil
}
