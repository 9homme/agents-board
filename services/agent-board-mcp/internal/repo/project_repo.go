package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"agent-board-mcp/internal/domain"
)

// ErrNotFound is returned when a requested record is not found in the database.
var ErrNotFound = errors.New("record not found")

// ProjectRepository defines the interface for project data access.
type ProjectRepository interface {
	CreateProject(ctx context.Context, p *domain.Project) (*domain.Project, error)
	GetProject(ctx context.Context, id string) (*domain.Project, error)
	UpdateProject(ctx context.Context, p *domain.Project) (*domain.Project, error)
	DeleteProject(ctx context.Context, id string) error
	ListProjects(ctx context.Context) ([]*domain.Project, error)
}

type projectRepo struct {
	db *sql.DB
}

// NewProjectRepo creates a new ProjectRepository using the provided database connection.
func NewProjectRepo(db *sql.DB) ProjectRepository {
	return &projectRepo{db: db}
}

func (r *projectRepo) CreateProject(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	query := `INSERT INTO projects (name, description) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	var created domain.Project
	created.Name = p.Name
	created.Description = p.Description

	err := r.db.QueryRowContext(ctx, query, p.Name, p.Description).Scan(
		&created.ID,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &created, nil
}

func (r *projectRepo) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM projects WHERE id = $1`
	var p domain.Project

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &p, nil
}

func (r *projectRepo) UpdateProject(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	query := `UPDATE projects SET name = $1, description = $2, updated_at = NOW() WHERE id = $3 RETURNING id, name, description, created_at, updated_at`
	var updated domain.Project

	err := r.db.QueryRowContext(ctx, query, p.Name, p.Description, p.ID).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Description,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return &updated, nil
}

func (r *projectRepo) DeleteProject(ctx context.Context, id string) error {
	query := `DELETE FROM projects WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

func (r *projectRepo) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM projects ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	projects := make([]*domain.Project, 0)
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating projects: %w", err)
	}

	return projects, nil
}
