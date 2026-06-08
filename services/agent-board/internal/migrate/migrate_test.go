package migrate_test

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"agent-board/internal/migrate"
)

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
