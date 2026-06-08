package migrate_test

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"agent-board/internal/migrate"
)

// testFS returns an in-memory FS with one migration file so UT-003 through
// UT-006 can reach the transaction path.
func testFS() fstest.MapFS {
	return fstest.MapFS{
		"000001_test.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE IF NOT EXISTS test_table (id SERIAL PRIMARY KEY);"),
		},
	}
}

// TestRun_UT001_CreateTableFails verifies that Run returns an error when the
// CREATE TABLE IF NOT EXISTS schema_migrations statement fails.
func TestRun_UT001_CreateTableFails(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnError(errors.New("db error"))

	// When
	err = migrate.Run(context.Background(), db, nil)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRun_UT002_QueryAppliedVersionsFails verifies that Run returns an error
// when the SELECT version FROM schema_migrations query fails.
func TestRun_UT002_QueryAppliedVersionsFails(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT version FROM schema_migrations").
		WillReturnError(errors.New("query error"))

	// When
	err = migrate.Run(context.Background(), db, testFS())

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRun_UT003_BeginTxFails verifies that Run returns an error when
// db.BeginTx fails, proving the migration aborts before applying any file.
func TestRun_UT003_BeginTxFails(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT version FROM schema_migrations").
		WillReturnRows(sqlmock.NewRows([]string{"version"}))
	mock.ExpectBegin().WillReturnError(errors.New("begin error"))

	// When
	err = migrate.Run(context.Background(), db, testFS())

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRun_UT004_MigrationSQLExecFails verifies that Run returns an error when
// executing the embedded migration SQL fails, proving the transaction rolls back.
func TestRun_UT004_MigrationSQLExecFails(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT version FROM schema_migrations").
		WillReturnRows(sqlmock.NewRows([]string{"version"}))
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS test_table").
		WillReturnError(errors.New("exec error"))
	mock.ExpectRollback()

	// When
	err = migrate.Run(context.Background(), db, testFS())

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exec error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
