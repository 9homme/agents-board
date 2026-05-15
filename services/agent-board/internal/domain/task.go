package domain

import "time"

// Task represents the core domain entity for a task.
type Task struct {
	ID          string    `json:"id"`
	UserStoryID string    `json:"userStoryId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
