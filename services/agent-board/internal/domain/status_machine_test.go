package domain_test

import (
	"testing"

	"agent-board/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Task transition tests

func TestTask_IsValidTransition(t *testing.T) {
	// UT-001 - Valid forward transitions
	t.Run("UT-001 - Valid forward transitions", func(t *testing.T) {
		task := domain.Task{Status: "pending"}
		assert.True(t, task.IsValidTransition("in_progress"))

		task.Status = "in_progress"
		assert.True(t, task.IsValidTransition("in_review"))

		task.Status = "in_review"
		assert.True(t, task.IsValidTransition("completed"))

		task.Status = "changes_requested"
		assert.True(t, task.IsValidTransition("completed"))
	})

	// UT-002 - Review cycle transitions
	t.Run("UT-002 - Review cycle transitions", func(t *testing.T) {
		task := domain.Task{Status: "in_review"}
		assert.True(t, task.IsValidTransition("changes_requested"))

		task.Status = "changes_requested"
		assert.True(t, task.IsValidTransition("in_progress"))
		assert.True(t, task.IsValidTransition("in_review"))
	})

	// UT-003 - Circuit breaker transition
	t.Run("UT-003 - Circuit breaker transition", func(t *testing.T) {
		task := domain.Task{Status: "changes_requested"}
		assert.True(t, task.IsValidTransition("blocked_circuit_breaker"))
	})

	// UT-004 - Invalid transitions are rejected
	t.Run("UT-004 - Invalid transitions are rejected", func(t *testing.T) {
		task := domain.Task{Status: "pending"}
		assert.False(t, task.IsValidTransition("completed"))
		assert.False(t, task.IsValidTransition("in_review"))

		task.Status = "completed"
		assert.False(t, task.IsValidTransition("in_progress"))
		assert.False(t, task.IsValidTransition("pending"))
		assert.False(t, task.IsValidTransition("changes_requested"))
	})
}

func TestNewTask_EnforceInitialState(t *testing.T) {
	// UT-005 - Enforce initial state on creation
	t.Run("UT-005 - Enforce initial state on creation", func(t *testing.T) {
		task, err := domain.NewTask("UserStory1", "Test Task", "Description", "pending")
		assert.NoError(t, err)
		assert.Equal(t, "pending", task.Status)

		_, err = domain.NewTask("UserStory1", "Test Task", "Description", "completed")
		assert.Error(t, err)

		_, err = domain.NewTask("UserStory1", "Test Task", "Description", "")
		assert.Error(t, err)
	})
}

// UserStory transition tests

func TestUserStory_IsValidTransition(t *testing.T) {
	// UT-001 - Valid forward transitions
	t.Run("UT-001 - Valid forward transitions", func(t *testing.T) {
		story := domain.UserStory{Status: "draft"}
		assert.True(t, story.IsValidTransition("in_development"))

		story.Status = "in_development"
		assert.True(t, story.IsValidTransition("in_signoff"))

		story.Status = "in_signoff"
		assert.True(t, story.IsValidTransition("done"))

		story.Status = "changes_requested"
		assert.True(t, story.IsValidTransition("done"))
	})

	// UT-002 - Sign-off cycle transitions
	t.Run("UT-002 - Sign-off cycle transitions", func(t *testing.T) {
		story := domain.UserStory{Status: "in_signoff"}
		assert.True(t, story.IsValidTransition("changes_requested"))

		story.Status = "changes_requested"
		assert.True(t, story.IsValidTransition("in_development"))
		assert.True(t, story.IsValidTransition("in_signoff"))
	})

	// UT-003 - Circuit breaker transition
	t.Run("UT-003 - Circuit breaker transition", func(t *testing.T) {
		story := domain.UserStory{Status: "changes_requested"}
		assert.True(t, story.IsValidTransition("blocked_circuit_breaker"))
	})

	// UT-004 - Invalid transitions are rejected
	t.Run("UT-004 - Invalid transitions are rejected", func(t *testing.T) {
		story := domain.UserStory{Status: "draft"}
		assert.False(t, story.IsValidTransition("done"))
		assert.False(t, story.IsValidTransition("in_signoff"))

		story.Status = "done"
		assert.False(t, story.IsValidTransition("in_development"))
		assert.False(t, story.IsValidTransition("draft"))
		assert.False(t, story.IsValidTransition("changes_requested"))
	})
}

func TestNewUserStory_EnforceInitialState(t *testing.T) {
	// UT-005 - Enforce initial state on creation
	t.Run("UT-005 - Enforce initial state on creation", func(t *testing.T) {
		story, err := domain.NewUserStory("Proj1", "Test Story", "Description", "draft")
		assert.NoError(t, err)
		assert.Equal(t, "draft", story.Status)

		_, err = domain.NewUserStory("Proj1", "Test Story", "Description", "done")
		assert.Error(t, err)

		_, err = domain.NewUserStory("Proj1", "Test Story", "Description", "")
		assert.Error(t, err)
	})
}
