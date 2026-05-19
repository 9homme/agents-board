package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"agent-board/internal/domain"
)

// TaskRepository defines the interface for task data access.
type TaskRepository interface {
	CreateTask(ctx context.Context, task *domain.Task) (*domain.Task, error)
	GetTask(ctx context.Context, id string) (*domain.Task, error)
	UpdateTask(ctx context.Context, task *domain.Task) (*domain.Task, error)
	// UpdateTaskStatus atomically updates the task's status and inserts an audit log entry
	// in a single database transaction.
	UpdateTaskStatus(ctx context.Context, id, fromStatus, toStatus string) (*domain.Task, error)
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

// UpdateTaskStatus atomically updates the task's status and inserts an audit log entry
// in a single database transaction, enforcing consistency between task state and audit trail.
func (r *taskRepo) UpdateTaskStatus(ctx context.Context, id, fromStatus, toStatus string) (*domain.Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
				log.Printf("task tx rollback failed after error: %v", rbErr)
			}
		}
	}()

	updateQuery := `
		UPDATE tasks SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, user_story_id, title, description, status, created_at, updated_at
	`
	task := &domain.Task{}
	err = tx.QueryRowContext(ctx, updateQuery, toStatus, id).
		Scan(&task.ID, &task.UserStoryID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	auditQuery := `INSERT INTO status_audit_trail (entity_id, entity_type, from_status, to_status) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, auditQuery, id, "task", fromStatus, toStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to insert audit log: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return task, nil
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
	defer func() { _ = rows.Close() }()

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
