package repo

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"agent-board/internal/domain"
)

// UserStoryWithCount embeds a UserStory with an additional TaskCount field
// representing the number of tasks associated with the story (via LEFT JOIN).
type UserStoryWithCount struct {
	domain.UserStory
	TaskCount int
}

// UserStoryRepository defines the interface for user story data operations.
type UserStoryRepository interface {
	CreateUserStory(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error)
	GetUserStory(ctx context.Context, id string) (*domain.UserStory, error)
	UpdateUserStory(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error)
	// UpdateUserStoryStatus atomically updates the user story status and inserts an audit trail entry.
	// It uses a single DB transaction to guarantee consistency.
	UpdateUserStoryStatus(ctx context.Context, id, fromStatus, toStatus string) (*domain.UserStory, error)
	DeleteUserStory(ctx context.Context, id string) error
	ListUserStories(ctx context.Context, projectID string) ([]*domain.UserStory, error)
	// ListUserStoriesWithTaskCount retrieves all user stories for a project,
	// each enriched with the count of tasks linked via tasks.user_story_id.
	// Results are ordered by created_at DESC.
	ListUserStoriesWithTaskCount(ctx context.Context, projectID string) ([]*UserStoryWithCount, error)
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

// UpdateUserStoryStatus atomically updates the user story status and inserts an audit trail entry
// within a single DB transaction (D-003).
func (r *UserStoryRepo) UpdateUserStoryStatus(ctx context.Context, id, fromStatus, toStatus string) (*domain.UserStory, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
				log.Printf("user_story tx rollback failed after error: %v", rbErr)
			}
		}
	}()

	query := `UPDATE user_stories SET status = $1, updated_at = NOW() WHERE id = $2 RETURNING id, project_id, title, description, status, created_at, updated_at`
	var u domain.UserStory
	err = tx.QueryRowContext(ctx, query, toStatus, id).Scan(&u.ID, &u.ProjectID, &u.Title, &u.Description, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = ErrNotFound
		}
		return nil, err
	}

	auditQuery := `INSERT INTO status_audit_trail (entity_id, entity_type, from_status, to_status) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, auditQuery, id, "user_story", fromStatus, toStatus)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
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

// ListUserStoriesWithTaskCount retrieves all user stories for a project joined with their task counts.
// It uses a LEFT JOIN so stories with zero tasks are still included.
// Results are ordered by created_at DESC.
func (r *UserStoryRepo) ListUserStoriesWithTaskCount(ctx context.Context, projectID string) ([]*UserStoryWithCount, error) {
	query := `SELECT us.id, us.project_id, us.title, us.description, us.status, us.created_at, us.updated_at, COUNT(t.id) AS task_count FROM user_stories us LEFT JOIN tasks t ON t.user_story_id = us.id WHERE us.project_id = $1 GROUP BY us.id ORDER BY us.created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	stories := make([]*UserStoryWithCount, 0)
	for rows.Next() {
		var s UserStoryWithCount
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Title, &s.Description, &s.Status, &s.CreatedAt, &s.UpdatedAt, &s.TaskCount); err != nil {
			return nil, err
		}
		stories = append(stories, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stories, nil
}

// ListUserStories retrieves all user stories for a specific project.
func (r *UserStoryRepo) ListUserStories(ctx context.Context, projectID string) ([]*domain.UserStory, error) {
	query := `SELECT id, project_id, title, description, status, created_at, updated_at FROM user_stories WHERE project_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var userStories []*domain.UserStory
	for rows.Next() {
		var u domain.UserStory
		if err := rows.Scan(&u.ID, &u.ProjectID, &u.Title, &u.Description, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		userStories = append(userStories, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if userStories == nil {
		userStories = []*domain.UserStory{}
	}
	return userStories, nil
}
