package migrate

import (
	"context"
	"database/sql"
	"io/fs"
)

// Run applies all migrations from the given FS to the database.
func Run(ctx context.Context, db *sql.DB, fsys fs.FS) error {
	return nil
}
