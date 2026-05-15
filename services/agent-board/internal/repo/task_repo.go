package repo

import (
	"context"
	"database/sql"
	"errors"

	"agent-board/internal/domain"
)

// TaskRepository defines the interface for task data access.
type TaskRepository interface {
	CreateTask(ctx context.Context, task *domain.Task) (*domain.Task, error)
	GetTask(ctx context.Context, id string) (*domain.Task, error)
	UpdateTask(ctx context.Context, task *domain.Task) (*domain.Task, error)
	DeleteTask(ctx context.Context, id string) error
	ListTasks(ctx context.Context, userStoryID string) ([]*domain.Task, error)
}

// taskRepo handles database operations for tasks.
type taskRepo struct {
	db *sql.DB
}

// NewTaskRepo creates a new TaskRepository.
func NewTaskRepo(db *sql.DB) TaskRepository {
	return &taskRepo{db: db}
}

// CreateTask creates a new task in the database.
func (r *taskRepo) CreateTask(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	query := `
		INSERT INTO tasks (user_story_id, title, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, task.UserStoryID, task.Title, task.Description, task.Status).
		Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// GetTask retrieves a task by ID.
func (r *taskRepo) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	query := `
		SELECT id, user_story_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`
	task := &domain.Task{}
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&task.ID, &task.UserStoryID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return task, nil
}

// UpdateTask updates an existing task.
func (r *taskRepo) UpdateTask(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, user_story_id, title, description, status, created_at, updated_at
	`
	updated := &domain.Task{}
	err := r.db.QueryRowContext(ctx, query, task.Title, task.Description, task.Status, task.ID).
		Scan(&updated.ID, &updated.UserStoryID, &updated.Title, &updated.Description, &updated.Status, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return updated, nil
}

// DeleteTask deletes a task by ID.
func (r *taskRepo) DeleteTask(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListTasks retrieves all tasks for a specific user story, ordered by created_at desc.
func (r *taskRepo) ListTasks(ctx context.Context, userStoryID string) ([]*domain.Task, error) {
	query := `
		SELECT id, user_story_id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE user_story_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userStoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task := &domain.Task{}
		err := rows.Scan(&task.ID, &task.UserStoryID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}
