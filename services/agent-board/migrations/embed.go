package migrations

import "embed"

// FS contains the embedded migration files (*.up.sql).
//
//go:embed *.up.sql
var FS embed.FS
