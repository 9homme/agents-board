package repo

import (
	"context"
	"database/sql"
	"errors"
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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

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
	defer func() { _ = db.Close() }()

	repo := NewDocumentRepo(db)
	id := "223e4567-e89b-12d3-a456-426614174000"

	mock.ExpectExec(`^DELETE FROM documents WHERE id = \$1$`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DeleteDocument(context.Background(), id)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-014: List documents by Project (ordering: updated_at DESC, id DESC)
func TestDocumentRepo_ListDocuments(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewDocumentRepo(db)
	now := time.Now()
	projectID := "123e4567-e89b-12d3-a456-426614174000"
	id1 := "11111111-e89b-12d3-a456-426614174000"
	id2 := "22222222-e89b-12d3-a456-426614174000"

	mock.ExpectQuery(`^SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE project_id = \$1 ORDER BY updated_at DESC, id DESC$`).
		WithArgs(projectID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "title", "content", "created_at", "updated_at"}).
			AddRow(id1, projectID, "D1", "C1", now, now).
			AddRow(id2, projectID, "D2", "C2", now, now))

	documents, err := r.ListDocuments(context.Background(), projectID)
	assert.NoError(t, err)
	assert.Len(t, documents, 2)
	assert.Equal(t, id1, documents[0].ID)
	assert.Equal(t, id2, documents[1].ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-US005-001 — CreateDocument generic DB error (D1)
func TestDocumentRepo_CreateDocument_GenericError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewDocumentRepo(db)

	mock.ExpectQuery(`INSERT INTO documents`).
		WillReturnError(errors.New("db down"))

	created, err := repo.CreateDocument(context.Background(), &domain.Document{
		ProjectID: "123e4567-e89b-12d3-a456-426614174000",
		Title:     "Test",
		Content:   "Content",
	})

	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
	assert.Contains(t, err.Error(), "failed to create document")
	assert.Nil(t, created)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-US005-002 — GetDocument generic DB error (D2)
func TestDocumentRepo_GetDocument_GenericError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewDocumentRepo(db)

	mock.ExpectQuery(`SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE id`).
		WillReturnError(errors.New("db down"))

	d, err := repo.GetDocument(context.Background(), "any-id")

	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
	assert.Contains(t, err.Error(), "failed to get document")
	assert.Nil(t, d)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-US005-003 — UpdateDocument not found (D3)
func TestDocumentRepo_UpdateDocument_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewDocumentRepo(db)

	mock.ExpectQuery(`UPDATE documents SET`).
		WillReturnError(sql.ErrNoRows)

	updated, err := repo.UpdateDocument(context.Background(), &domain.Document{
		ID:      "any-id",
		Title:   "T",
		Content: "C",
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotFound))
	assert.Nil(t, updated)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-US005-004 — UpdateDocument generic error (D4)
func TestDocumentRepo_UpdateDocument_GenericError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewDocumentRepo(db)

	mock.ExpectQuery(`UPDATE documents SET`).
		WillReturnError(errors.New("db down"))

	updated, err := repo.UpdateDocument(context.Background(), &domain.Document{
		ID:      "any-id",
		Title:   "T",
		Content: "C",
	})

	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
	assert.Contains(t, err.Error(), "failed to update document")
	assert.Nil(t, updated)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-US005-005 — DeleteDocument generic error (D5)
func TestDocumentRepo_DeleteDocument_GenericError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewDocumentRepo(db)

	mock.ExpectExec(`DELETE FROM documents WHERE id`).
		WillReturnError(errors.New("db down"))

	err = repo.DeleteDocument(context.Background(), "any-id")

	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
	assert.Contains(t, err.Error(), "failed to delete document")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-US005-006 — ListDocuments query error (D6)
func TestDocumentRepo_ListDocuments_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewDocumentRepo(db)

	mock.ExpectQuery(`SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE project_id`).
		WillReturnError(errors.New("db down"))

	docs, err := r.ListDocuments(context.Background(), "proj-id")

	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
	assert.Contains(t, err.Error(), "failed to list documents")
	assert.Nil(t, docs)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-US005-007 — ListDocuments scan error (D7)
// A bool value is provided for the created_at (time.Time) column to trigger a
// convertAssign type-mismatch error during rows.Scan.
func TestDocumentRepo_ListDocuments_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewDocumentRepo(db)

	rows := sqlmock.NewRows([]string{"id", "project_id", "title", "content", "created_at", "updated_at"}).
		AddRow("doc-id", "proj-id", "Title", "Content", true, true) // bool → time.Time causes scan error

	mock.ExpectQuery(`SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE project_id`).
		WillReturnRows(rows)

	docs, err := r.ListDocuments(context.Background(), "proj-id")

	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrNotFound))
	assert.Contains(t, err.Error(), "failed to scan document")
	assert.Nil(t, docs)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-US005-008 — ListDocuments rows.Err() error (D8)
func TestDocumentRepo_ListDocuments_RowsErr(t *testing.T) {
	t.Skip("red: stub D8")
}

// UT-US002-010 — Repo: ListDocuments orders by updated_at DESC, id DESC (tiebreaker test)
func TestDocumentRepo_ListDocuments_OrderByUpdatedAtDescIDDesc(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	r := NewDocumentRepo(db)
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	// Three documents:
	// Doc A: updated_at = 2026-05-20T10:00:00Z, id = "aaaa0001-..."
	// Doc B: updated_at = 2026-05-19T10:00:00Z, id = "bbbb0002-..."
	// Doc C: updated_at = 2026-05-20T10:00:00Z, id = "cccc0003-..." (same updated_at as A — tiebreaker)
	//
	// Expected order from SQL: C (same updated_at as A, cccc > aaaa so id DESC puts C before A), A, B.
	tHigh := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	tLow := time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)

	idA := "aaaa0001-0000-0000-0000-000000000001"
	idB := "bbbb0002-0000-0000-0000-000000000002"
	idC := "cccc0003-0000-0000-0000-000000000003"

	// The SQL mock verifies the ORDER BY clause is correct.
	// The rows are returned already sorted (simulating what the DB would return).
	mock.ExpectQuery(`^SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE project_id = \$1 ORDER BY updated_at DESC, id DESC$`).
		WithArgs(projectID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "title", "content", "created_at", "updated_at"}).
			AddRow(idC, projectID, "C", "", tHigh, tHigh). // C first: same updated_at as A, but cccc > aaaa
			AddRow(idA, projectID, "A", "", tHigh, tHigh). // A second
			AddRow(idB, projectID, "B", "", tLow, tLow))   // B last: older updated_at

	documents, err := r.ListDocuments(context.Background(), projectID)
	require.NoError(t, err)
	require.Len(t, documents, 3)

	assert.Equal(t, idC, documents[0].ID, "C must come first (same updated_at as A, but id cccc > aaaa)")
	assert.Equal(t, idA, documents[1].ID, "A must come second")
	assert.Equal(t, idB, documents[2].ID, "B must come last (oldest updated_at)")

	assert.NoError(t, mock.ExpectationsWereMet())
}
