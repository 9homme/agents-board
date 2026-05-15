package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"agent-board/internal/domain"
)

// DocumentRepository defines the interface for document data access.
type DocumentRepository interface {
	CreateDocument(ctx context.Context, d *domain.Document) (*domain.Document, error)
	GetDocument(ctx context.Context, id string) (*domain.Document, error)
	UpdateDocument(ctx context.Context, d *domain.Document) (*domain.Document, error)
	DeleteDocument(ctx context.Context, id string) error
	ListDocuments(ctx context.Context, projectID string) ([]*domain.Document, error)
}

type documentRepo struct {
	db *sql.DB
}

// NewDocumentRepo creates a new DocumentRepository using the provided database connection.
func NewDocumentRepo(db *sql.DB) DocumentRepository {
	return &documentRepo{db: db}
}

func (r *documentRepo) CreateDocument(ctx context.Context, d *domain.Document) (*domain.Document, error) {
	query := `INSERT INTO documents (project_id, title, content) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	var created domain.Document
	created.ProjectID = d.ProjectID
	created.Title = d.Title
	created.Content = d.Content

	err := r.db.QueryRowContext(ctx, query, d.ProjectID, d.Title, d.Content).Scan(
		&created.ID,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return &created, nil
}

func (r *documentRepo) GetDocument(ctx context.Context, id string) (*domain.Document, error) {
	query := `SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE id = $1`
	var d domain.Document

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID,
		&d.ProjectID,
		&d.Title,
		&d.Content,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return &d, nil
}

func (r *documentRepo) UpdateDocument(ctx context.Context, d *domain.Document) (*domain.Document, error) {
	query := `UPDATE documents SET title = $1, content = $2, updated_at = NOW() WHERE id = $3 RETURNING id, project_id, title, content, created_at, updated_at`
	var updated domain.Document

	err := r.db.QueryRowContext(ctx, query, d.Title, d.Content, d.ID).Scan(
		&updated.ID,
		&updated.ProjectID,
		&updated.Title,
		&updated.Content,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return &updated, nil
}

func (r *documentRepo) DeleteDocument(ctx context.Context, id string) error {
	query := `DELETE FROM documents WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

func (r *documentRepo) ListDocuments(ctx context.Context, projectID string) ([]*domain.Document, error) {
	query := `SELECT id, project_id, title, content, created_at, updated_at FROM documents WHERE project_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer func() { _ = rows.Close() }()

	documents := make([]*domain.Document, 0)
	for rows.Next() {
		var d domain.Document
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Title, &d.Content, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating documents: %w", err)
	}

	return documents, nil
}
