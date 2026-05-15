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

// UT-010: Create document in DB
func TestDocumentRepo_CreateDocument(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDocumentRepo(db)
	now := time.Now()
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	d := &domain.Document{
		ProjectID: projectID,
		Title:     "Test Document",
		Content:   "A test document content",
	}

	mock.ExpectQuery(`^INSERT INTO documents \(project_id, title, content\) VALUES \(\$1, \$2, \$3\) RETURNING id, created_at, updated_at$`).
		WithArgs(d.ProjectID, d.Title, d.Content).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("223e4567-e89b-12d3-a456-426614174000", now, now))

	created, err := repo.CreateDocument(context.Background(), d)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, "223e4567-e89b-12d3-a456-426614174000", created.ID)
	assert.Equal(t, d.ProjectID, created.ProjectID)
	assert.Equal(t, d.Title, created.Title)
	assert.Equal(t, d.Content, created.Content)
	assert.Equal(t, now, created.CreatedAt)
	assert.Equal(t, now, created.UpdatedAt)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-011: Get document from DB
func TestDocumentRepo_GetDocument(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDocumentRepo(db)
	now := time.Now()
	id := "223e4567-e89b-12d3-a456-426614174000"
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	// Success case
	mock.ExpectQuery(`^SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE id = \$1$`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "title", "content", "created_at", "updated_at"}).
			AddRow(id, projectID, "Test Document", "Content", now, now))

	d, err := repo.GetDocument(context.Background(), id)
	assert.NoError(t, err)
	assert.NotNil(t, d)
	assert.Equal(t, id, d.ID)

	// Not found case
	mock.ExpectQuery(`^SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE id = \$1$`).
		WithArgs("non-existent").
		WillReturnError(sql.ErrNoRows)

	d, err = repo.GetDocument(context.Background(), "non-existent")
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, d)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-012: Update document in DB
func TestDocumentRepo_UpdateDocument(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDocumentRepo(db)
	now := time.Now()
	id := "223e4567-e89b-12d3-a456-426614174000"
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	d := &domain.Document{
		ID:        id,
		ProjectID: projectID,
		Title:     "Updated Document",
		Content:   "Updated content",
	}

	mock.ExpectQuery(`^UPDATE documents SET title = \$1, content = \$2, updated_at = NOW\(\) WHERE id = \$3 RETURNING id, project_id, title, content, created_at, updated_at$`).
		WithArgs(d.Title, d.Content, d.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "title", "content", "created_at", "updated_at"}).
			AddRow(id, d.ProjectID, d.Title, d.Content, now, now))

	updated, err := repo.UpdateDocument(context.Background(), d)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, id, updated.ID)
	assert.Equal(t, d.Title, updated.Title)
	assert.Equal(t, d.Content, updated.Content)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-013: Delete document in DB
func TestDocumentRepo_DeleteDocument(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDocumentRepo(db)
	id := "223e4567-e89b-12d3-a456-426614174000"

	mock.ExpectExec(`^DELETE FROM documents WHERE id = \$1$`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DeleteDocument(context.Background(), id)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-014: List documents by Project
func TestDocumentRepo_ListDocuments(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewDocumentRepo(db)
	now := time.Now()
	projectID := "123e4567-e89b-12d3-a456-426614174000"
	id1 := "11111111-e89b-12d3-a456-426614174000"
	id2 := "22222222-e89b-12d3-a456-426614174000"

	mock.ExpectQuery(`^SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE project_id = \$1 ORDER BY created_at DESC$`).
		WithArgs(projectID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "title", "content", "created_at", "updated_at"}).
			AddRow(id1, projectID, "D1", "C1", now, now).
			AddRow(id2, projectID, "D2", "C2", now, now))

	documents, err := repo.ListDocuments(context.Background(), projectID)
	assert.NoError(t, err)
	assert.Len(t, documents, 2)
	assert.Equal(t, id1, documents[0].ID)
	assert.Equal(t, id2, documents[1].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}
