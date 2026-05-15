package repo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"agent-board/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UT-005: Create project in DB
func TestProjectRepo_CreateProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewProjectRepo(db)
	now := time.Now()

	p := &domain.Project{
		Name:        "Test Project",
		Description: "A test project",
	}

	mock.ExpectQuery(`^INSERT INTO projects \(name, description\) VALUES \(\$1, \$2\) RETURNING id, created_at, updated_at$`).
		WithArgs(p.Name, p.Description).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("123e4567-e89b-12d3-a456-426614174000", now, now))

	created, err := repo.CreateProject(context.Background(), p)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", created.ID)
	assert.Equal(t, p.Name, created.Name)
	assert.Equal(t, p.Description, created.Description)
	assert.Equal(t, now, created.CreatedAt)
	assert.Equal(t, now, created.UpdatedAt)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-006: Get project from DB
func TestProjectRepo_GetProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewProjectRepo(db)
	now := time.Now()
	id := "123e4567-e89b-12d3-a456-426614174000"

	// Success case
	mock.ExpectQuery(`^SELECT id, name, description, created_at, updated_at FROM projects WHERE id = \$1$`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow(id, "Test Project", "Desc", now, now))

	p, err := repo.GetProject(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, id, p.ID)

	// Not found case
	mock.ExpectQuery(`^SELECT id, name, description, created_at, updated_at FROM projects WHERE id = \$1$`).
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	p, err = repo.GetProject(context.Background(), "non-existent")
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, p)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-007: Update project in DB
func TestProjectRepo_UpdateProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewProjectRepo(db)
	now := time.Now()
	id := "123e4567-e89b-12d3-a456-426614174000"

	p := &domain.Project{
		ID:          id,
		Name:        "Updated Project",
		Description: "Updated desc",
	}

	mock.ExpectQuery(`^UPDATE projects SET name = \$1, description = \$2, updated_at = NOW\(\) WHERE id = \$3 RETURNING id, name, description, created_at, updated_at$`).
		WithArgs(p.Name, p.Description, p.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow(id, p.Name, p.Description, now, now))

	updated, err := repo.UpdateProject(context.Background(), p)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, id, updated.ID)
	assert.Equal(t, p.Name, updated.Name)
	assert.Equal(t, p.Description, updated.Description)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-008: Delete project in DB
func TestProjectRepo_DeleteProject(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewProjectRepo(db)
	id := "123e4567-e89b-12d3-a456-426614174000"

	mock.ExpectExec(`^DELETE FROM projects WHERE id = \$1$`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DeleteProject(context.Background(), id)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-009: List projects in DB
func TestProjectRepo_ListProjects(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewProjectRepo(db)
	now := time.Now()
	id1 := "11111111-e89b-12d3-a456-426614174000"
	id2 := "22222222-e89b-12d3-a456-426614174000"

	mock.ExpectQuery(`^SELECT id, name, description, created_at, updated_at FROM projects ORDER BY created_at DESC$`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow(id1, "P1", "D1", now, now).
			AddRow(id2, "P2", "D2", now, now))

	projects, err := repo.ListProjects(context.Background())
	assert.NoError(t, err)
	assert.Len(t, projects, 2)
	assert.Equal(t, id1, projects[0].ID)
	assert.Equal(t, id2, projects[1].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}
