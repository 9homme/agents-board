package migrate_test

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"agent-board/internal/migrate"
	"agent-board/migrations"
)

// pgConnURL returns a postgres connection string for a given database name.
// It reads PGHOST, PGPORT, PGUSER from environment or uses defaults matching
// the local Homebrew postgres instance (TCP on localhost:5432).
func pgConnURL(dbName string) string {
	host := envOrDefault("PGHOST", "localhost")
	port := envOrDefault("PGPORT", "5432")
	user := envOrDefault("PGUSER", os.Getenv("USER"))
	return fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=disable", user, host, port, dbName)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// skipIfNoPostgres skips the test if no PostgreSQL is reachable.
// Uses the postgres system database (always exists) as the probe.
func skipIfNoPostgres(t *testing.T) {
	t.Helper()
	db, err := sql.Open("pgx", pgConnURL("postgres"))
	if err != nil {
		t.Skipf("no postgres available: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.PingContext(context.Background()); err != nil {
		t.Skipf("postgres ping failed: %v", err)
	}
}

// newTestDB creates an isolated Postgres database for the test and returns a
// connected *sql.DB. The database is dropped when the test ends.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbName := fmt.Sprintf("test_us044_%s", sanitizeForDB(t.Name()))
	adminDB, err := sql.Open("pgx", pgConnURL("postgres"))
	require.NoError(t, err)
	defer func() { _ = adminDB.Close() }()
	require.NoError(t, adminDB.PingContext(context.Background()))

	_, err = adminDB.ExecContext(context.Background(), fmt.Sprintf("CREATE DATABASE %q", dbName))
	require.NoError(t, err)

	t.Cleanup(func() {
		// Close the test DB connection before dropping
		a, err2 := sql.Open("pgx", pgConnURL("postgres"))
		if err2 == nil {
			_, _ = a.ExecContext(context.Background(),
				fmt.Sprintf("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s'", dbName))
			_, _ = a.ExecContext(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %q", dbName))
			_ = a.Close()
		}
	})

	db, err := sql.Open("pgx", pgConnURL(dbName))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, db.PingContext(context.Background()))
	return db
}

// sanitizeForDB converts a test name to a safe DB identifier fragment.
func sanitizeForDB(name string) string {
	out := make([]byte, 0, len(name))
	for _, c := range []byte(name) {
		switch {
		case c >= 'a' && c <= 'z':
			out = append(out, c)
		case c >= 'A' && c <= 'Z':
			out = append(out, c+32) // lowercase
		case c >= '0' && c <= '9':
			out = append(out, c)
		default:
			out = append(out, '_')
		}
	}
	if len(out) > 40 {
		out = out[:40]
	}
	return string(out)
}

// runMigrationsUpTo runs all embedded *.up.sql migrations up to and including
// the named file (inclusive filter), then returns the db.
func runMigrationsUpTo(t *testing.T, db *sql.DB, upToName string) {
	t.Helper()
	filtered := fstest.MapFS{}
	entries, err := fs.ReadDir(migrations.FS, ".")
	require.NoError(t, err)
	for _, e := range entries {
		filtered[e.Name()] = &fstest.MapFile{}
		content, err2 := fs.ReadFile(migrations.FS, e.Name())
		require.NoError(t, err2)
		filtered[e.Name()].Data = content
		if e.Name() == upToName {
			break
		}
	}
	require.NoError(t, migrate.Run(context.Background(), db, filtered))
}

// runMigrationsThrough000002 runs up to and including 000002.
func runMigrationsThrough000002(t *testing.T, db *sql.DB) {
	t.Helper()
	runMigrationsUpTo(t, db, "000002_status_audit_trail.up.sql")
}

// runAllMigrations runs all migrations (including 000003).
func runAllMigrations(t *testing.T, db *sql.DB) {
	t.Helper()
	require.NoError(t, migrate.Run(context.Background(), db, migrations.FS))
}

// IT-044-001 — requirements table exists with correct columns post-migration
func TestIT044_001_RequirementsTableExistsPostMigration(t *testing.T) {
	skipIfNoPostgres(t)
	db := newTestDB(t)
	runAllMigrations(t, db)

	ctx := context.Background()

	// Table exists — check columns
	type colInfo struct {
		dataType   string
		isNullable string
	}
	cols := map[string]colInfo{}
	rows, err := db.QueryContext(ctx, `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'requirements' AND table_schema = 'public'
	`)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var name, dt, nullable string
		require.NoError(t, rows.Scan(&name, &dt, &nullable))
		cols[name] = colInfo{dataType: dt, isNullable: nullable}
	}
	require.NoError(t, rows.Err())

	assert.Contains(t, cols, "id", "requirements.id column must exist")
	assert.Contains(t, cols, "project_id")
	assert.Contains(t, cols, "name")
	assert.Contains(t, cols, "description")
	assert.Contains(t, cols, "status")
	assert.Contains(t, cols, "created_at")
	assert.Contains(t, cols, "updated_at")

	assert.Equal(t, "NO", cols["project_id"].isNullable, "project_id must be NOT NULL")
	assert.Equal(t, "NO", cols["name"].isNullable, "name must be NOT NULL")
	assert.Equal(t, "NO", cols["description"].isNullable, "description must be NOT NULL")
	assert.Equal(t, "NO", cols["status"].isNullable, "status must be NOT NULL")
	assert.Equal(t, "NO", cols["created_at"].isNullable, "created_at must be NOT NULL")
	assert.Equal(t, "NO", cols["updated_at"].isNullable, "updated_at must be NOT NULL")

	// FK on project_id references projects(id) ON DELETE CASCADE
	var fkCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.referential_constraints rc
		JOIN information_schema.key_column_usage kcu
		  ON kcu.constraint_name = rc.constraint_name
		 AND kcu.table_schema = rc.constraint_schema
		WHERE kcu.table_name = 'requirements'
		  AND kcu.column_name = 'project_id'
		  AND rc.delete_rule = 'CASCADE'
	`).Scan(&fkCount)
	require.NoError(t, err)
	assert.Equal(t, 1, fkCount, "requirements.project_id must have ON DELETE CASCADE FK")

	// CHECK constraint on status
	var checkCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.check_constraints cc
		JOIN information_schema.constraint_column_usage ccu
		  ON ccu.constraint_name = cc.constraint_name
		WHERE ccu.table_name = 'requirements'
		  AND cc.check_clause LIKE '%draft%'
		  AND cc.check_clause LIKE '%in_progress%'
		  AND cc.check_clause LIKE '%done%'
	`).Scan(&checkCount)
	require.NoError(t, err)
	assert.Greater(t, checkCount, 0, "requirements.status must have a CHECK constraint")

	// Index on project_id
	var idxExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes
			WHERE tablename = 'requirements' AND indexname = 'idx_requirements_project_id'
		)
	`).Scan(&idxExists)
	require.NoError(t, err)
	assert.True(t, idxExists, "idx_requirements_project_id must exist")
}

// IT-044-002 — user_stories.requirement_id NOT NULL after migration
func TestIT044_002_UserStoriesRequirementIDNotNull(t *testing.T) {
	skipIfNoPostgres(t)
	db := newTestDB(t)
	runAllMigrations(t, db)

	ctx := context.Background()

	var isNullable string
	err := db.QueryRowContext(ctx, `
		SELECT is_nullable FROM information_schema.columns
		WHERE table_name = 'user_stories' AND column_name = 'requirement_id'
		  AND table_schema = 'public'
	`).Scan(&isNullable)
	require.NoError(t, err)
	assert.Equal(t, "NO", isNullable, "user_stories.requirement_id must be NOT NULL")

	// FK to requirements(id) ON DELETE CASCADE
	var fkCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.referential_constraints rc
		JOIN information_schema.key_column_usage kcu
		  ON kcu.constraint_name = rc.constraint_name
		 AND kcu.table_schema = rc.constraint_schema
		WHERE kcu.table_name = 'user_stories'
		  AND kcu.column_name = 'requirement_id'
		  AND rc.delete_rule = 'CASCADE'
	`).Scan(&fkCount)
	require.NoError(t, err)
	assert.Equal(t, 1, fkCount, "user_stories.requirement_id must FK to requirements ON DELETE CASCADE")

	// Index exists
	var idxExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes
			WHERE tablename = 'user_stories' AND indexname = 'idx_user_stories_requirement_id'
		)
	`).Scan(&idxExists)
	require.NoError(t, err)
	assert.True(t, idxExists, "idx_user_stories_requirement_id must exist")

	// Edge case: INSERT with NULL requirement_id must fail
	_, err = db.ExecContext(ctx, `
		INSERT INTO user_stories (project_id, title, description, status, requirement_id)
		VALUES (gen_random_uuid(), 'test', '', 'draft', NULL)
	`)
	require.Error(t, err, "INSERT with requirement_id=NULL must be rejected")
}

// IT-044-003 — documents.requirement_id NOT NULL after migration
func TestIT044_003_DocumentsRequirementIDNotNull(t *testing.T) {
	skipIfNoPostgres(t)
	db := newTestDB(t)
	runAllMigrations(t, db)

	ctx := context.Background()

	var isNullable string
	err := db.QueryRowContext(ctx, `
		SELECT is_nullable FROM information_schema.columns
		WHERE table_name = 'documents' AND column_name = 'requirement_id'
		  AND table_schema = 'public'
	`).Scan(&isNullable)
	require.NoError(t, err)
	assert.Equal(t, "NO", isNullable, "documents.requirement_id must be NOT NULL")

	// FK to requirements(id) ON DELETE CASCADE
	var fkCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.referential_constraints rc
		JOIN information_schema.key_column_usage kcu
		  ON kcu.constraint_name = rc.constraint_name
		 AND kcu.table_schema = rc.constraint_schema
		WHERE kcu.table_name = 'documents'
		  AND kcu.column_name = 'requirement_id'
		  AND rc.delete_rule = 'CASCADE'
	`).Scan(&fkCount)
	require.NoError(t, err)
	assert.Equal(t, 1, fkCount, "documents.requirement_id must FK to requirements ON DELETE CASCADE")

	// Index exists
	var idxExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes
			WHERE tablename = 'documents' AND indexname = 'idx_documents_requirement_id'
		)
	`).Scan(&idxExists)
	require.NoError(t, err)
	assert.True(t, idxExists, "idx_documents_requirement_id must exist")

	// Edge case: INSERT with NULL requirement_id must fail
	_, err = db.ExecContext(ctx, `
		INSERT INTO documents (project_id, title, content, requirement_id)
		VALUES (gen_random_uuid(), 'test', '', NULL)
	`)
	require.Error(t, err, "INSERT with requirement_id=NULL must be rejected")
}

// IT-044-004 — existing projects get exactly one "Default" requirement per backfill
func TestIT044_004_ExistingProjectsGetDefaultRequirement(t *testing.T) {
	skipIfNoPostgres(t)
	db := newTestDB(t)

	// Run only up to 000002 first
	runMigrationsThrough000002(t, db)

	ctx := context.Background()

	// Insert 3 projects
	var p1, p2, p3 string
	err := db.QueryRowContext(ctx, `INSERT INTO projects (name, description) VALUES ('P1', '') RETURNING id`).Scan(&p1)
	require.NoError(t, err)
	err = db.QueryRowContext(ctx, `INSERT INTO projects (name, description) VALUES ('P2', '') RETURNING id`).Scan(&p2)
	require.NoError(t, err)
	err = db.QueryRowContext(ctx, `INSERT INTO projects (name, description) VALUES ('P3', '') RETURNING id`).Scan(&p3)
	require.NoError(t, err)

	// Insert 2 user stories under P1
	_, err = db.ExecContext(ctx, `INSERT INTO user_stories (project_id, title, description, status) VALUES ($1, 'US1', '', 'draft'), ($1, 'US2', '', 'draft')`, p1)
	require.NoError(t, err)

	// Insert 1 document under P2
	_, err = db.ExecContext(ctx, `INSERT INTO documents (project_id, title, content) VALUES ($1, 'Doc1', '')`, p2)
	require.NoError(t, err)

	// Now run migration 000003
	runAllMigrations(t, db)

	// 3 Default requirements (one per project)
	var defaultCount int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM requirements WHERE name = 'Default'`).Scan(&defaultCount)
	require.NoError(t, err)
	assert.Equal(t, 3, defaultCount, "must have exactly 3 Default requirements (one per project)")

	// Each has status = 'draft'
	var nonDraftCount int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM requirements WHERE name = 'Default' AND status != 'draft'`).Scan(&nonDraftCount)
	require.NoError(t, err)
	assert.Equal(t, 0, nonDraftCount, "all Default requirements must have status = 'draft'")

	// Each has correct project_id
	for _, pid := range []string{p1, p2, p3} {
		var reqCount int
		err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM requirements WHERE project_id = $1 AND name = 'Default'`, pid).Scan(&reqCount)
		require.NoError(t, err)
		assert.Equal(t, 1, reqCount, "project %s must have exactly 1 Default requirement", pid)
	}
}

// IT-044-005 — zero data loss: user stories re-parented to Default requirement
func TestIT044_005_UserStoriesReParentedToDefault(t *testing.T) {
	skipIfNoPostgres(t)
	db := newTestDB(t)

	runMigrationsThrough000002(t, db)

	ctx := context.Background()

	var p1 string
	err := db.QueryRowContext(ctx, `INSERT INTO projects (name, description) VALUES ('P1', '') RETURNING id`).Scan(&p1)
	require.NoError(t, err)

	var p2, p3 string
	err = db.QueryRowContext(ctx, `INSERT INTO projects (name, description) VALUES ('P2', '') RETURNING id`).Scan(&p2)
	require.NoError(t, err)
	err = db.QueryRowContext(ctx, `INSERT INTO projects (name, description) VALUES ('P3', '') RETURNING id`).Scan(&p3)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `INSERT INTO user_stories (project_id, title, description, status) VALUES ($1, 'US1', '', 'draft'), ($1, 'US2', '', 'draft')`, p1)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `INSERT INTO documents (project_id, title, content) VALUES ($1, 'Doc1', '')`, p2)
	require.NoError(t, err)

	// Count before migration
	var countBefore int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_stories`).Scan(&countBefore)
	require.NoError(t, err)
	assert.Equal(t, 2, countBefore)

	// Run 000003
	runAllMigrations(t, db)

	// Count after — no loss
	var countAfter int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_stories`).Scan(&countAfter)
	require.NoError(t, err)
	assert.Equal(t, 2, countAfter, "user_stories count must be unchanged after migration")

	// Get P1's Default requirement ID
	var reqID string
	err = db.QueryRowContext(ctx, `SELECT id FROM requirements WHERE project_id = $1 AND name = 'Default'`, p1).Scan(&reqID)
	require.NoError(t, err)

	// Both user stories must point to that requirement
	var linkedCount int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_stories WHERE project_id = $1 AND requirement_id = $2`, p1, reqID).Scan(&linkedCount)
	require.NoError(t, err)
	assert.Equal(t, 2, linkedCount, "both user stories must be re-parented to P1's Default requirement")

	// project_id unchanged
	var withOriginalProject int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_stories WHERE project_id = $1`, p1).Scan(&withOriginalProject)
	require.NoError(t, err)
	assert.Equal(t, 2, withOriginalProject, "user_stories must still have original project_id")

	// Zero orphans
	var orphans int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_stories WHERE project_id = $1 AND requirement_id IS NULL`, p1).Scan(&orphans)
	require.NoError(t, err)
	assert.Equal(t, 0, orphans, "no orphaned user_stories")
}

// IT-044-006 — zero data loss: documents re-parented to Default requirement
func TestIT044_006_DocumentsReParentedToDefault(t *testing.T) {
	skipIfNoPostgres(t)
	db := newTestDB(t)

	runMigrationsThrough000002(t, db)

	ctx := context.Background()

	var p1, p2, p3 string
	err := db.QueryRowContext(ctx, `INSERT INTO projects (name, description) VALUES ('P1', '') RETURNING id`).Scan(&p1)
	require.NoError(t, err)
	err = db.QueryRowContext(ctx, `INSERT INTO projects (name, description) VALUES ('P2', '') RETURNING id`).Scan(&p2)
	require.NoError(t, err)
	err = db.QueryRowContext(ctx, `INSERT INTO projects (name, description) VALUES ('P3', '') RETURNING id`).Scan(&p3)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `INSERT INTO user_stories (project_id, title, description, status) VALUES ($1, 'US1', '', 'draft'), ($1, 'US2', '', 'draft')`, p1)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `INSERT INTO documents (project_id, title, content) VALUES ($1, 'Doc1', '')`, p2)
	require.NoError(t, err)

	var countBefore int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents`).Scan(&countBefore)
	require.NoError(t, err)
	assert.Equal(t, 1, countBefore)

	runAllMigrations(t, db)

	var countAfter int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents`).Scan(&countAfter)
	require.NoError(t, err)
	assert.Equal(t, 1, countAfter, "documents count must be unchanged after migration")

	// Get P2's Default requirement ID
	var reqID string
	err = db.QueryRowContext(ctx, `SELECT id FROM requirements WHERE project_id = $1 AND name = 'Default'`, p2).Scan(&reqID)
	require.NoError(t, err)

	// Document must be re-parented
	var linkedCount int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents WHERE project_id = $1 AND requirement_id = $2`, p2, reqID).Scan(&linkedCount)
	require.NoError(t, err)
	assert.Equal(t, 1, linkedCount, "document must be re-parented to P2's Default requirement")

	// project_id unchanged
	var withOriginalProject int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents WHERE project_id = $1`, p2).Scan(&withOriginalProject)
	require.NoError(t, err)
	assert.Equal(t, 1, withOriginalProject, "document must still have original project_id")
}

// IT-044-007 — projects.path column is TEXT NOT NULL with unique constraint
func TestIT044_007_ProjectsPathColumnNotNullUnique(t *testing.T) {
	skipIfNoPostgres(t)
	db := newTestDB(t)
	runAllMigrations(t, db)

	ctx := context.Background()

	var dataType, isNullable string
	err := db.QueryRowContext(ctx, `
		SELECT data_type, is_nullable FROM information_schema.columns
		WHERE table_name = 'projects' AND column_name = 'path' AND table_schema = 'public'
	`).Scan(&dataType, &isNullable)
	require.NoError(t, err)
	assert.Equal(t, "text", dataType, "projects.path must have data_type = text")
	assert.Equal(t, "NO", isNullable, "projects.path must be NOT NULL")

	// Unique constraint exists
	var constraintExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.table_constraints
			WHERE table_name = 'projects'
			  AND constraint_name = 'uq_projects_path'
			  AND constraint_type = 'UNIQUE'
		)
	`).Scan(&constraintExists)
	require.NoError(t, err)
	assert.True(t, constraintExists, "uq_projects_path unique constraint must exist")
}

// IT-044-008 — path uniqueness constraint rejects duplicate paths
func TestIT044_008_PathUniquenessConstraintRejectsDuplicates(t *testing.T) {
	skipIfNoPostgres(t)
	db := newTestDB(t)
	runAllMigrations(t, db)

	ctx := context.Background()

	// Insert first project with a specific path
	_, err := db.ExecContext(ctx, `
		INSERT INTO projects (name, description, path) VALUES ('P1', '', '/tmp/test-project')
	`)
	require.NoError(t, err, "first INSERT must succeed")

	// Attempt to INSERT a second project with the same path
	_, err = db.ExecContext(ctx, `
		INSERT INTO projects (name, description, path) VALUES ('P2', '', '/tmp/test-project')
	`)
	require.Error(t, err, "INSERT with duplicate path must be rejected")
	// Check it's a unique violation (code 23505)
	assert.Contains(t, err.Error(), "23505", "error must be unique violation (23505)")

	// Edge case: different paths (with vs without trailing slash) are allowed
	_, err = db.ExecContext(ctx, `
		INSERT INTO projects (name, description, path) VALUES ('P3', '', '/tmp/test-project/')
	`)
	assert.NoError(t, err, "path with trailing slash is distinct and must be allowed")
}

// IT-044-009 — down migration reverses schema cleanly
// Coverage exemption: down migration is documentation-only per architecture;
// tested here for correctness guarantees only, not run in production automation.
func TestIT044_009_DownMigrationReversesSchemaCleanly(t *testing.T) {
	skipIfNoPostgres(t)
	db := newTestDB(t)
	runAllMigrations(t, db)

	ctx := context.Background()

	// Insert data that exercises all 000003 additions
	var projID string
	err := db.QueryRowContext(ctx, `INSERT INTO projects (name, description, path) VALUES ('P1', '', '/tmp/down-test') RETURNING id`).Scan(&projID)
	require.NoError(t, err)

	var reqID string
	err = db.QueryRowContext(ctx, `INSERT INTO requirements (project_id, name, status) VALUES ($1, 'Default', 'draft') RETURNING id`, projID).Scan(&reqID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `INSERT INTO user_stories (project_id, requirement_id, title, description, status) VALUES ($1, $2, 'US1', '', 'draft')`, projID, reqID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `INSERT INTO documents (project_id, requirement_id, title, content) VALUES ($1, $2, 'Doc1', '')`, projID, reqID)
	require.NoError(t, err)

	// Read and execute the down SQL directly
	downSQL, err := os.ReadFile("../../migrations/000003_requirement_entity.down.sql")
	require.NoError(t, err, "down SQL file must be readable")

	_, err = db.ExecContext(ctx, string(downSQL))
	require.NoError(t, err, "down migration must execute without error")

	// requirements table is gone
	var tableExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'requirements' AND table_schema = 'public')
	`).Scan(&tableExists)
	require.NoError(t, err)
	assert.False(t, tableExists, "requirements table must be dropped by down migration")

	// user_stories.requirement_id column is gone
	var colExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'user_stories' AND column_name = 'requirement_id')
	`).Scan(&colExists)
	require.NoError(t, err)
	assert.False(t, colExists, "user_stories.requirement_id must be dropped")

	// documents.requirement_id column is gone
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'documents' AND column_name = 'requirement_id')
	`).Scan(&colExists)
	require.NoError(t, err)
	assert.False(t, colExists, "documents.requirement_id must be dropped")

	// projects.path column is gone
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'projects' AND column_name = 'path')
	`).Scan(&colExists)
	require.NoError(t, err)
	assert.False(t, colExists, "projects.path must be dropped")

	// uq_projects_path constraint is gone
	var constraintExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'projects' AND constraint_name = 'uq_projects_path')
	`).Scan(&constraintExists)
	require.NoError(t, err)
	assert.False(t, constraintExists, "uq_projects_path constraint must be dropped")

	// Core content of user_stories and documents still present (no data loss on core fields)
	var usCount int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_stories`).Scan(&usCount)
	require.NoError(t, err)
	assert.Equal(t, 1, usCount, "user_stories rows must survive the down migration")

	var docCount int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents`).Scan(&docCount)
	require.NoError(t, err)
	assert.Equal(t, 1, docCount, "documents rows must survive the down migration")
}

// IT-044-010 — no orphaned rows after backfill (row-count invariant)
func TestIT044_010_NoOrphanedRowsAfterBackfill(t *testing.T) {
	skipIfNoPostgres(t)
	db := newTestDB(t)

	runMigrationsThrough000002(t, db)

	ctx := context.Background()

	// Setup: 3 projects, 2 user_stories, 1 document
	var p1, p2, p3 string
	err := db.QueryRowContext(ctx, `INSERT INTO projects (name, description) VALUES ('P1', '') RETURNING id`).Scan(&p1)
	require.NoError(t, err)
	err = db.QueryRowContext(ctx, `INSERT INTO projects (name, description) VALUES ('P2', '') RETURNING id`).Scan(&p2)
	require.NoError(t, err)
	err = db.QueryRowContext(ctx, `INSERT INTO projects (name, description) VALUES ('P3', '') RETURNING id`).Scan(&p3)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `INSERT INTO user_stories (project_id, title, description, status) VALUES ($1, 'US1', '', 'draft'), ($1, 'US2', '', 'draft')`, p1)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `INSERT INTO documents (project_id, title, content) VALUES ($1, 'Doc1', '')`, p2)
	require.NoError(t, err)

	// Run 000003
	runAllMigrations(t, db)

	// Row-count invariants
	var reqCount int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM requirements`).Scan(&reqCount)
	require.NoError(t, err)
	assert.Equal(t, 3, reqCount, "must have exactly 3 requirements (one per project)")

	var usCount int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_stories`).Scan(&usCount)
	require.NoError(t, err)
	assert.Equal(t, 2, usCount, "user_stories count must be unchanged")

	var docCount int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents`).Scan(&docCount)
	require.NoError(t, err)
	assert.Equal(t, 1, docCount, "documents count must be unchanged")

	// Zero orphans on user_stories (impossible after NOT NULL, but verify)
	var usOrphans int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_stories WHERE requirement_id IS NULL`).Scan(&usOrphans)
	require.NoError(t, err)
	assert.Equal(t, 0, usOrphans, "no orphaned user_stories")

	var docOrphans int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents WHERE requirement_id IS NULL`).Scan(&docOrphans)
	require.NoError(t, err)
	assert.Equal(t, 0, docOrphans, "no orphaned documents")
}
