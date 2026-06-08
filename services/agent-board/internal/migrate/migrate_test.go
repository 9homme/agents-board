package migrate

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_UT001_CreateTableFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	expectedErr := errors.New("failed to create table")
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").WillReturnError(expectedErr)

	fs := fstest.MapFS{}
	err = Run(context.Background(), db, fs)

	assert.ErrorIs(t, err, expectedErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}
