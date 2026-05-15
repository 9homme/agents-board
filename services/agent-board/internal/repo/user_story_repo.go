package repo

import (
	"context"
	"database/sql"

	"agent-board/internal/domain"
)

// UserStoryRepository defines the interface for user story data operations.
type UserStoryRepository interface {
	CreateUserStory(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error)
	GetUserStory(ctx context.Context, id string) (*domain.UserStory, error)
	UpdateUserStory(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error)
	DeleteUserStory(ctx context.Context, id string) error
	ListUserStories(ctx context.Context, projectID string) ([]*domain.UserStory, error)
}

// UserStoryRepo handles database operations for user stories.
type UserStoryRepo struct {
	db *sql.DB
}

// NewUserStoryRepo creates a new UserStoryRepo.
func NewUserStoryRepo(db *sql.DB) *UserStoryRepo {
	return &UserStoryRepo{db: db}
}

// CreateUserStory inserts a new user story into the database.
func (r *UserStoryRepo) CreateUserStory(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error) {
	query := `INSERT INTO user_stories (project_id, title, description, status) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, u.ProjectID, u.Title, u.Description, u.Status).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserStory retrieves a user story by ID.
func (r *UserStoryRepo) GetUserStory(ctx context.Context, id string) (*domain.UserStory, error) {
	query := `SELECT id, project_id, title, description, status, created_at, updated_at FROM user_stories WHERE id = $1`
	var u domain.UserStory
	err := r.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.ProjectID, &u.Title, &u.Description, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

// UpdateUserStory updates an existing user story.
func (r *UserStoryRepo) UpdateUserStory(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error) {
	query := `UPDATE user_stories SET title = $1, description = $2, status = $3, updated_at = NOW() WHERE id = $4 RETURNING id, project_id, title, description, status, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, u.Title, u.Description, u.Status, u.ID).Scan(&u.ID, &u.ProjectID, &u.Title, &u.Description, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

// DeleteUserStory deletes a user story by ID.
func (r *UserStoryRepo) DeleteUserStory(ctx context.Context, id string) error {
	query := `DELETE FROM user_stories WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListUserStories retrieves all user stories for a specific project.
func (r *UserStoryRepo) ListUserStories(ctx context.Context, projectID string) ([]*domain.UserStory, error) {
	query := `SELECT id, project_id, title, description, status, created_at, updated_at FROM user_stories WHERE project_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userStories []*domain.UserStory
	for rows.Next() {
		var u domain.UserStory
		if err := rows.Scan(&u.ID, &u.ProjectID, &u.Title, &u.Description, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		userStories = append(userStories, &u)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if userStories == nil {
		userStories = []*domain.UserStory{}
	}
	return userStories, nil
}
