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

// --- US049 tests ---

// UT-049-001 — TaskStatusBlockedReviewGate constant has value "blocked_review_gate"
func TestTaskStatusBlockedReviewGate_ConstantValue(t *testing.T) {
	assert.Equal(t, "blocked_review_gate", domain.TaskStatusBlockedReviewGate)
}

// UT-049-002 — IsValidTransition("blocked_review_gate") from in_review returns true
func TestTask_IsValidTransition_InReview_To_BlockedReviewGate(t *testing.T) {
	task := &domain.Task{Status: domain.TaskStatusInReview}
	assert.True(t, task.IsValidTransition(domain.TaskStatusBlockedReviewGate))
}

// UT-049-003 — IsValidTransition("blocked_review_gate") from changes_requested returns true
func TestTask_IsValidTransition_ChangesRequested_To_BlockedReviewGate(t *testing.T) {
	task := &domain.Task{Status: domain.TaskStatusChangesRequested}
	assert.True(t, task.IsValidTransition(domain.TaskStatusBlockedReviewGate))
}

// UT-049-004 — IsValidTransition from blocked_review_gate to any status returns false (terminal)
func TestTask_IsValidTransition_BlockedReviewGate_IsTerminal(t *testing.T) {
	task := &domain.Task{Status: domain.TaskStatusBlockedReviewGate}
	targets := []string{
		"pending",
		"in_progress",
		"in_review",
		"completed",
		"changes_requested",
		"blocked_review_gate", // self
		"blocked_circuit_breaker",
		"",
		"unknown_status",
	}
	for _, target := range targets {
		assert.False(t, task.IsValidTransition(target), "expected false for blocked_review_gate → %q", target)
	}
}

// UT-049-005 — Cannot reach blocked_review_gate from pending
func TestTask_IsValidTransition_Pending_To_BlockedReviewGate_False(t *testing.T) {
	task := &domain.Task{Status: domain.TaskStatusPending}
	assert.False(t, task.IsValidTransition(domain.TaskStatusBlockedReviewGate))
}

// UT-049-006 — Cannot reach blocked_review_gate from in_progress
func TestTask_IsValidTransition_InProgress_To_BlockedReviewGate_False(t *testing.T) {
	task := &domain.Task{Status: domain.TaskStatusInProgress}
	assert.False(t, task.IsValidTransition(domain.TaskStatusBlockedReviewGate))
}

// UT-049-007 — Cannot reach blocked_review_gate from completed
func TestTask_IsValidTransition_Completed_To_BlockedReviewGate_False(t *testing.T) {
	task := &domain.Task{Status: domain.TaskStatusCompleted}
	assert.False(t, task.IsValidTransition(domain.TaskStatusBlockedReviewGate))
}

// UT-049-008 — Cannot reach blocked_review_gate from blocked_circuit_breaker
func TestTask_IsValidTransition_BlockedCircuitBreaker_To_BlockedReviewGate_False(t *testing.T) {
	task := &domain.Task{Status: domain.TaskStatusBlockedCircuitBreaker}
	assert.False(t, task.IsValidTransition(domain.TaskStatusBlockedReviewGate))
}

// UT-049-009 — Full transition matrix table-driven test (no regression)
func TestTask_IsValidTransition_FullMatrix(t *testing.T) {
	type testCase struct {
		from     string
		to       string
		expected bool
	}
	cases := []testCase{
		// pending
		{domain.TaskStatusPending, domain.TaskStatusInProgress, true},
		{domain.TaskStatusPending, domain.TaskStatusInReview, false},
		{domain.TaskStatusPending, domain.TaskStatusCompleted, false},
		{domain.TaskStatusPending, domain.TaskStatusChangesRequested, false},
		{domain.TaskStatusPending, domain.TaskStatusBlockedCircuitBreaker, false},
		{domain.TaskStatusPending, domain.TaskStatusBlockedReviewGate, false},
		// in_progress
		{domain.TaskStatusInProgress, domain.TaskStatusInReview, true},
		{domain.TaskStatusInProgress, domain.TaskStatusPending, false},
		{domain.TaskStatusInProgress, domain.TaskStatusCompleted, false},
		{domain.TaskStatusInProgress, domain.TaskStatusBlockedReviewGate, false},
		// in_review
		{domain.TaskStatusInReview, domain.TaskStatusCompleted, true},
		{domain.TaskStatusInReview, domain.TaskStatusChangesRequested, true},
		{domain.TaskStatusInReview, domain.TaskStatusBlockedReviewGate, true},
		{domain.TaskStatusInReview, domain.TaskStatusInProgress, false},
		{domain.TaskStatusInReview, domain.TaskStatusPending, false},
		{domain.TaskStatusInReview, domain.TaskStatusBlockedCircuitBreaker, false},
		// changes_requested
		{domain.TaskStatusChangesRequested, domain.TaskStatusInProgress, true},
		{domain.TaskStatusChangesRequested, domain.TaskStatusInReview, true},
		{domain.TaskStatusChangesRequested, domain.TaskStatusCompleted, true},
		{domain.TaskStatusChangesRequested, domain.TaskStatusBlockedCircuitBreaker, true},
		{domain.TaskStatusChangesRequested, domain.TaskStatusBlockedReviewGate, true},
		{domain.TaskStatusChangesRequested, domain.TaskStatusPending, false},
		// completed (terminal)
		{domain.TaskStatusCompleted, domain.TaskStatusInProgress, false},
		{domain.TaskStatusCompleted, domain.TaskStatusInReview, false},
		{domain.TaskStatusCompleted, domain.TaskStatusBlockedReviewGate, false},
		{domain.TaskStatusCompleted, domain.TaskStatusPending, false},
		{domain.TaskStatusCompleted, domain.TaskStatusChangesRequested, false},
		{domain.TaskStatusCompleted, domain.TaskStatusBlockedCircuitBreaker, false},
		// blocked_circuit_breaker (terminal)
		{domain.TaskStatusBlockedCircuitBreaker, domain.TaskStatusInProgress, false},
		{domain.TaskStatusBlockedCircuitBreaker, domain.TaskStatusInReview, false},
		{domain.TaskStatusBlockedCircuitBreaker, domain.TaskStatusCompleted, false},
		{domain.TaskStatusBlockedCircuitBreaker, domain.TaskStatusChangesRequested, false},
		{domain.TaskStatusBlockedCircuitBreaker, domain.TaskStatusPending, false},
		{domain.TaskStatusBlockedCircuitBreaker, domain.TaskStatusBlockedReviewGate, false},
		// blocked_review_gate (terminal)
		{domain.TaskStatusBlockedReviewGate, domain.TaskStatusInProgress, false},
		{domain.TaskStatusBlockedReviewGate, domain.TaskStatusInReview, false},
		{domain.TaskStatusBlockedReviewGate, domain.TaskStatusCompleted, false},
		{domain.TaskStatusBlockedReviewGate, domain.TaskStatusChangesRequested, false},
		{domain.TaskStatusBlockedReviewGate, domain.TaskStatusPending, false},
		{domain.TaskStatusBlockedReviewGate, domain.TaskStatusBlockedCircuitBreaker, false},
		{domain.TaskStatusBlockedReviewGate, domain.TaskStatusBlockedReviewGate, false},
	}

	for _, tc := range cases {
		tc := tc // capture range variable
		t.Run(tc.from+"→"+tc.to, func(t *testing.T) {
			task := &domain.Task{Status: tc.from}
			got := task.IsValidTransition(tc.to)
			assert.Equal(t, tc.expected, got, "from=%q to=%q", tc.from, tc.to)
		})
	}
}
