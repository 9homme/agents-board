package domain

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

var (
	// ErrInvalidStatusTransition is returned when a status transition is not allowed.
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	// ErrInvalidInitialStatus is returned when an entity is created with an invalid initial status.
	ErrInvalidInitialStatus = errors.New("invalid initial status")
)

// Task Statuses
const (
	TaskStatusPending               = "pending"
	TaskStatusInProgress            = "in_progress"
	TaskStatusInReview              = "in_review"
	TaskStatusChangesRequested      = "changes_requested"
	TaskStatusCompleted             = "completed"
	TaskStatusBlockedCircuitBreaker = "blocked_circuit_breaker"
	// TaskStatusBlockedReviewGate is a terminal state reached when the review-gate
	// tool fails. It is distinct from blocked_circuit_breaker (code-review failure).
	// Valid transitions in: in_review → blocked_review_gate and changes_requested → blocked_review_gate.
	// No transitions out.
	TaskStatusBlockedReviewGate = "blocked_review_gate"
)

// UserStory Statuses
const (
	UserStoryStatusDraft                 = "draft"
	UserStoryStatusInDevelopment         = "in_development"
	UserStoryStatusInSignoff             = "in_signoff"
	UserStoryStatusChangesRequested      = "changes_requested"
	UserStoryStatusDone                  = "done"
	UserStoryStatusBlockedCircuitBreaker = "blocked_circuit_breaker"
)

// IsValidTransition checks if the transition for a Task is valid according to the state machine.
func (t *Task) IsValidTransition(newStatus string) bool {
	switch t.Status {
	case TaskStatusPending:
		return newStatus == TaskStatusInProgress
	case TaskStatusInProgress:
		return newStatus == TaskStatusInReview
	case TaskStatusInReview:
		return newStatus == TaskStatusCompleted || newStatus == TaskStatusChangesRequested || newStatus == TaskStatusBlockedReviewGate
	case TaskStatusChangesRequested:
		return newStatus == TaskStatusInProgress || newStatus == TaskStatusInReview || newStatus == TaskStatusCompleted || newStatus == TaskStatusBlockedCircuitBreaker || newStatus == TaskStatusBlockedReviewGate
	case TaskStatusCompleted, TaskStatusBlockedCircuitBreaker, TaskStatusBlockedReviewGate:
		return false // Terminal states
	default:
		return false
	}
}

// NewTask creates a new Task entity and enforces its initial state.
func NewTask(userStoryID, title, description, status string) (*Task, error) {
	if status != TaskStatusPending {
		return nil, ErrInvalidInitialStatus
	}
	return &Task{
		ID:          uuid.New().String(),
		UserStoryID: userStoryID,
		Title:       title,
		Description: description,
		Status:      status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// IsValidTransition checks if the transition for a UserStory is valid according to the state machine.
func (s *UserStory) IsValidTransition(newStatus string) bool {
	switch s.Status {
	case UserStoryStatusDraft:
		return newStatus == UserStoryStatusInDevelopment
	case UserStoryStatusInDevelopment:
		return newStatus == UserStoryStatusInSignoff
	case UserStoryStatusInSignoff:
		return newStatus == UserStoryStatusDone || newStatus == UserStoryStatusChangesRequested
	case UserStoryStatusChangesRequested:
		return newStatus == UserStoryStatusInDevelopment || newStatus == UserStoryStatusInSignoff || newStatus == UserStoryStatusDone || newStatus == UserStoryStatusBlockedCircuitBreaker
	case UserStoryStatusDone, UserStoryStatusBlockedCircuitBreaker:
		return false
	default:
		return false
	}
}

// NewUserStory creates a new UserStory entity and enforces its initial state.
func NewUserStory(projectID, title, description, status string) (*UserStory, error) {
	if status != UserStoryStatusDraft {
		return nil, ErrInvalidInitialStatus
	}
	return &UserStory{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Status:      status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}
