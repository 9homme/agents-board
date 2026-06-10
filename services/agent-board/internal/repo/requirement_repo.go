package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"agent-board/internal/domain"
)

// ErrProjectNotFound is returned when a requirement references a non-existent project (FK violation 23503).
var ErrProjectNotFound = errors.New("project not found")

// RequirementRepository defines the interface for requirement data access.
type RequirementRepository interface {
	// ListByProject retrieves all requirements for a project, ordered by created_at ASC.
	// Returns an empty (non-nil) slice when no requirements exist.
	ListByProject(ctx context.Context, projectID string) ([]domain.Requirement, error)

	// Create inserts a new requirement and populates ID, CreatedAt, UpdatedAt from the DB.
	Create(ctx context.Context, req *domain.Requirement) (*domain.Requirement, error)

	// GetRequirement retrieves a single requirement by ID.
	// Returns ErrNotFound when the requirement does not exist.
	GetRequirement(ctx context.Context, id string) (*domain.Requirement, error)

	// Update applies a partial patch to the requirement identified by id.
	// Returns ErrNotFound when the requirement does not exist.
	Update(ctx context.Context, id string, patch RequirementPatch) (*domain.Requirement, error)
}

// RequirementPatch carries optional fields for a partial update of a Requirement.
// A nil pointer means "no change" for that field.
type RequirementPatch struct {
	Name        *string
	Description *string
	Status      *string
}

type requirementRepo struct {
	db *sql.DB
}

// NewRequirementRepo creates a new RequirementRepository backed by the provided database.
func NewRequirementRepo(db *sql.DB) RequirementRepository {
	return &requirementRepo{db: db}
}

// ListByProject retrieves all requirements for a project ordered by created_at ASC.
func (r *requirementRepo) ListByProject(ctx context.Context, projectID string) ([]domain.Requirement, error) {
	const query = `SELECT id, project_id, name, description, status, created_at, updated_at
		FROM requirements
		WHERE project_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("list requirements by project: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make([]domain.Requirement, 0)
	for rows.Next() {
		var req domain.Requirement
		if err := rows.Scan(
			&req.ID,
			&req.ProjectID,
			&req.Name,
			&req.Description,
			&req.Status,
			&req.CreatedAt,
			&req.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan requirement row: %w", err)
		}
		result = append(result, req)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate requirement rows: %w", err)
	}

	return result, nil
}

// Create inserts a new requirement into the database, returning the row with generated fields populated.
func (r *requirementRepo) Create(ctx context.Context, req *domain.Requirement) (*domain.Requirement, error) {
	const query = `INSERT INTO requirements (project_id, name, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	created := *req // copy so we don't mutate the caller's value
	err := r.db.QueryRowContext(ctx, query, req.ProjectID, req.Name, req.Description, req.Status).
		Scan(&created.ID, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		// FK violation: project does not exist (Postgres error code 23503)
		if isFKViolation(err) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("create requirement: %w", err)
	}

	return &created, nil
}

// GetRequirement retrieves a single requirement by its ID.
func (r *requirementRepo) GetRequirement(ctx context.Context, id string) (*domain.Requirement, error) {
	const query = `SELECT id, project_id, name, description, status, created_at, updated_at
		FROM requirements
		WHERE id = $1`

	var req domain.Requirement
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID,
		&req.ProjectID,
		&req.Name,
		&req.Description,
		&req.Status,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get requirement: %w", err)
	}

	return &req, nil
}

// Update applies a partial patch to the requirement identified by id.
// The UPDATE sets updated_at = NOW() and returns the complete updated row.
func (r *requirementRepo) Update(ctx context.Context, id string, patch RequirementPatch) (*domain.Requirement, error) {
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if patch.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *patch.Name)
		argIdx++
	}
	if patch.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *patch.Description)
		argIdx++
	}
	if patch.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *patch.Status)
		argIdx++
	}

	args = append(args, id)
	query := fmt.Sprintf(
		`UPDATE requirements SET %s WHERE id = $%d RETURNING id, project_id, name, description, status, created_at, updated_at`,
		strings.Join(setClauses, ", "),
		argIdx,
	)

	var updated domain.Requirement
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&updated.ID,
		&updated.ProjectID,
		&updated.Name,
		&updated.Description,
		&updated.Status,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update requirement: %w", err)
	}

	return &updated, nil
}

// isFKViolation checks if the error represents a Postgres FK violation (SQLSTATE 23503).
// The caller guarantees err is non-nil; string matching is used since the pq driver
// is not directly imported in this package.
func isFKViolation(err error) bool {
	return strings.Contains(err.Error(), "23503")
}
