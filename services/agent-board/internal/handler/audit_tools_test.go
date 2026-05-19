package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"agent-board/internal/domain"
	"agent-board/internal/mcp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAuditRepo is a mock implementation of AuditRepository.
type MockAuditRepo struct {
	GetTaskAuditTrailFunc      func(ctx context.Context, taskID string) ([]*domain.StatusAuditLog, error)
	GetUserStoryAuditTrailFunc func(ctx context.Context, userStoryID string) ([]*domain.StatusAuditLog, error)
}

func (m *MockAuditRepo) GetTaskAuditTrail(ctx context.Context, taskID string) ([]*domain.StatusAuditLog, error) {
	return m.GetTaskAuditTrailFunc(ctx, taskID)
}

func (m *MockAuditRepo) GetUserStoryAuditTrail(ctx context.Context, userStoryID string) ([]*domain.StatusAuditLog, error) {
	return m.GetUserStoryAuditTrailFunc(ctx, userStoryID)
}

// IT-004: get_task_audit_trail returns ordered audit entries with correct JSON shape.
func TestAuditTools_GetTaskAuditTrail(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockAuditRepo{}
	RegisterAuditTools(registry, mockRepo)

	taskID := "223e4567-e89b-12d3-a456-426614174000"
	auditID1 := "aaa00001-e89b-12d3-a456-426614174000"
	auditID2 := "aaa00002-e89b-12d3-a456-426614174000"

	t1 := time.Now().Add(-2 * time.Minute)
	t2 := time.Now().Add(-1 * time.Minute)

	mockRepo.GetTaskAuditTrailFunc = func(ctx context.Context, id string) ([]*domain.StatusAuditLog, error) {
		assert.Equal(t, taskID, id)
		return []*domain.StatusAuditLog{
			{ID: auditID1, EntityID: taskID, EntityType: "task", FromStatus: "pending", ToStatus: "in_progress", ChangedAt: t1},
			{ID: auditID2, EntityID: taskID, EntityType: "task", FromStatus: "in_progress", ToStatus: "in_review", ChangedAt: t2},
		}, nil
	}

	args := json.RawMessage(`{"taskId":"` + taskID + `"}`)
	toolHandler, ok := registry.GetTool("get_task_audit_trail")
	require.True(t, ok)

	res, err := toolHandler(context.Background(), args)
	require.NoError(t, err)
	require.NotNil(t, res)

	// Verify the response shape matches architecture contract.
	resMap, ok := res.(map[string]interface{})
	require.True(t, ok, "result should be map[string]interface{}")

	trail, ok := resMap["auditTrail"].([]AuditLogResponse)
	require.True(t, ok, "auditTrail should be []AuditLogResponse")
	require.Len(t, trail, 2)

	assert.Equal(t, auditID1, trail[0].ID)
	assert.Equal(t, taskID, trail[0].EntityID)
	assert.Equal(t, "task", trail[0].EntityType)
	assert.Equal(t, "pending", trail[0].FromStatus)
	assert.Equal(t, "in_progress", trail[0].ToStatus)
	assert.Equal(t, t1.Format(time.RFC3339), trail[0].ChangedAt)

	assert.Equal(t, auditID2, trail[1].ID)
	assert.Equal(t, "in_progress", trail[1].FromStatus)
	assert.Equal(t, "in_review", trail[1].ToStatus)
}

// IT-004b: get_task_audit_trail returns empty auditTrail when no entries exist.
func TestAuditTools_GetTaskAuditTrail_Empty(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockAuditRepo{}
	RegisterAuditTools(registry, mockRepo)

	taskID := "223e4567-e89b-12d3-a456-426614174000"
	mockRepo.GetTaskAuditTrailFunc = func(ctx context.Context, id string) ([]*domain.StatusAuditLog, error) {
		return []*domain.StatusAuditLog{}, nil
	}

	args := json.RawMessage(`{"taskId":"` + taskID + `"}`)
	toolHandler, ok := registry.GetTool("get_task_audit_trail")
	require.True(t, ok)

	res, err := toolHandler(context.Background(), args)
	require.NoError(t, err)

	resMap, ok := res.(map[string]interface{})
	require.True(t, ok)
	trail, ok := resMap["auditTrail"].([]AuditLogResponse)
	require.True(t, ok)
	assert.Len(t, trail, 0)
}

// IT-004c: get_task_audit_trail requires taskId field.
func TestAuditTools_GetTaskAuditTrail_MissingTaskID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockAuditRepo{}
	RegisterAuditTools(registry, mockRepo)

	args := json.RawMessage(`{}`)
	toolHandler, ok := registry.GetTool("get_task_audit_trail")
	require.True(t, ok)

	res, err := toolHandler(context.Background(), args)
	require.Error(t, err)
	assert.Nil(t, res)
}

// IT-005: get_user_story_audit_trail returns ordered audit entries with correct JSON shape.
func TestAuditTools_GetUserStoryAuditTrail(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockAuditRepo{}
	RegisterAuditTools(registry, mockRepo)

	storyID := "333e4567-e89b-12d3-a456-426614174000"
	auditID1 := "bbb00001-e89b-12d3-a456-426614174000"
	auditID2 := "bbb00002-e89b-12d3-a456-426614174000"

	t1 := time.Now().Add(-2 * time.Minute)
	t2 := time.Now().Add(-1 * time.Minute)

	mockRepo.GetUserStoryAuditTrailFunc = func(ctx context.Context, id string) ([]*domain.StatusAuditLog, error) {
		assert.Equal(t, storyID, id)
		return []*domain.StatusAuditLog{
			{ID: auditID1, EntityID: storyID, EntityType: "user_story", FromStatus: "draft", ToStatus: "in_development", ChangedAt: t1},
			{ID: auditID2, EntityID: storyID, EntityType: "user_story", FromStatus: "in_development", ToStatus: "in_signoff", ChangedAt: t2},
		}, nil
	}

	args := json.RawMessage(`{"userStoryId":"` + storyID + `"}`)
	toolHandler, ok := registry.GetTool("get_user_story_audit_trail")
	require.True(t, ok)

	res, err := toolHandler(context.Background(), args)
	require.NoError(t, err)
	require.NotNil(t, res)

	resMap, ok := res.(map[string]interface{})
	require.True(t, ok, "result should be map[string]interface{}")

	trail, ok := resMap["auditTrail"].([]AuditLogResponse)
	require.True(t, ok, "auditTrail should be []AuditLogResponse")
	require.Len(t, trail, 2)

	assert.Equal(t, auditID1, trail[0].ID)
	assert.Equal(t, storyID, trail[0].EntityID)
	assert.Equal(t, "user_story", trail[0].EntityType)
	assert.Equal(t, "draft", trail[0].FromStatus)
	assert.Equal(t, "in_development", trail[0].ToStatus)
	assert.Equal(t, t1.Format(time.RFC3339), trail[0].ChangedAt)

	assert.Equal(t, auditID2, trail[1].ID)
	assert.Equal(t, "in_development", trail[1].FromStatus)
	assert.Equal(t, "in_signoff", trail[1].ToStatus)
}

// IT-005b: get_user_story_audit_trail requires userStoryId field.
func TestAuditTools_GetUserStoryAuditTrail_MissingID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockAuditRepo{}
	RegisterAuditTools(registry, mockRepo)

	args := json.RawMessage(`{}`)
	toolHandler, ok := registry.GetTool("get_user_story_audit_trail")
	require.True(t, ok)

	res, err := toolHandler(context.Background(), args)
	require.Error(t, err)
	assert.Nil(t, res)
}

// IT-003: When update_task is called with invalid transition, no audit record is created.
// This test verifies the existing behavior via the mock: UpdateTaskStatus is never called.
func TestAuditTools_NoAuditOnInvalidTaskTransition(t *testing.T) {
	// This is verified by the existing TestTaskTools_UpdateTask_InvalidTransition test
	// using MockTaskRepo — UpdateTaskStatus is never called when the transition is invalid.
	// We add this test as a documentation / integration bridge that confirms the expectation.
	mockRepo := new(MockTaskRepo)
	registry := mcp.NewToolRegistry()
	RegisterTaskTools(registry, mockRepo)

	ctx := context.Background()
	now := time.Now()
	id := "223e4567-e89b-12d3-a456-426614174000"

	existingTask := &domain.Task{
		ID:          id,
		UserStoryID: "123e4567-e89b-12d3-a456-426614174000",
		Title:       "Test Task",
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// pending -> completed is invalid — only GetTask should be called.
	mockRepo.On("GetTask", ctx, id).Return(existingTask, nil)

	args, _ := json.Marshal(map[string]interface{}{"id": id, "status": "completed"})
	toolHandler, ok := registry.GetTool("update_task")
	require.True(t, ok)

	res, err := toolHandler(ctx, args)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "invalid transition")

	// UpdateTaskStatus must NOT have been called (no audit record written).
	mockRepo.AssertNotCalled(t, "UpdateTaskStatus")
	mockRepo.AssertExpectations(t)
}

// IT-003b: When update_user_story is called with invalid transition, no audit record is created.
func TestAuditTools_NoAuditOnInvalidUserStoryTransition(t *testing.T) {
	registry := mcp.NewToolRegistry()

	// trackingMock records whether UpdateUserStoryStatus was ever called.
	updateStatusCalled := false
	mockRepo := &auditTestUserStoryRepo{
		getFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return &domain.UserStory{
				ID:        id,
				ProjectID: "123e4567-e89b-12d3-a456-426614174000",
				Title:     "Story",
				Status:    "draft",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
		updateStatusFunc: func(ctx context.Context, id, from, to string) (*domain.UserStory, error) {
			updateStatusCalled = true
			return nil, errors.New("should not be called")
		},
	}
	RegisterUserStoryTools(registry, mockRepo)

	ctx := context.Background()
	storyID := "333e4567-e89b-12d3-a456-426614174000"

	// draft -> done is invalid.
	args, _ := json.Marshal(map[string]interface{}{"id": storyID, "status": "done"})
	toolHandler, ok := registry.GetTool("update_user_story")
	require.True(t, ok)

	res, err := toolHandler(ctx, args)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "invalid transition")

	// UpdateUserStoryStatus must NOT have been called (no audit record written).
	assert.False(t, updateStatusCalled, "UpdateUserStoryStatus should not be called on invalid transition")
}

// auditTestUserStoryRepo is a minimal repo.UserStoryRepository implementation for audit tests
// that need to track whether UpdateUserStoryStatus was called.
type auditTestUserStoryRepo struct {
	getFunc          func(ctx context.Context, id string) (*domain.UserStory, error)
	updateStatusFunc func(ctx context.Context, id, from, to string) (*domain.UserStory, error)
	updateFunc       func(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error)
	deleteFunc       func(ctx context.Context, id string) error
	listFunc         func(ctx context.Context, projectID string) ([]*domain.UserStory, error)
	createFunc       func(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error)
}

func (m *auditTestUserStoryRepo) CreateUserStory(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, u)
	}
	return nil, errors.New("not implemented")
}

func (m *auditTestUserStoryRepo) GetUserStory(ctx context.Context, id string) (*domain.UserStory, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *auditTestUserStoryRepo) UpdateUserStory(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, u)
	}
	return nil, errors.New("not implemented")
}

func (m *auditTestUserStoryRepo) UpdateUserStoryStatus(ctx context.Context, id, from, to string) (*domain.UserStory, error) {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, id, from, to)
	}
	return nil, errors.New("not implemented")
}

func (m *auditTestUserStoryRepo) DeleteUserStory(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func (m *auditTestUserStoryRepo) ListUserStories(ctx context.Context, projectID string) ([]*domain.UserStory, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, projectID)
	}
	return nil, errors.New("not implemented")
}

// IT-004-error: get_task_audit_trail propagates repo errors.
func TestAuditTools_GetTaskAuditTrail_RepoError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockAuditRepo{}
	RegisterAuditTools(registry, mockRepo)

	taskID := "223e4567-e89b-12d3-a456-426614174000"
	mockRepo.GetTaskAuditTrailFunc = func(ctx context.Context, id string) ([]*domain.StatusAuditLog, error) {
		return nil, errors.New("db connection failed")
	}

	args := json.RawMessage(`{"taskId":"` + taskID + `"}`)
	toolHandler, ok := registry.GetTool("get_task_audit_trail")
	require.True(t, ok)

	res, err := toolHandler(context.Background(), args)
	require.Error(t, err)
	assert.Nil(t, res)
}
