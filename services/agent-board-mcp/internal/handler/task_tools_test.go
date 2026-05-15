package handler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"agent-board-mcp/internal/domain"
	"agent-board-mcp/internal/mcp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskRepo is a mock implementation of repo.TaskRepository
type MockTaskRepo struct {
	mock.Mock
}

func (m *MockTaskRepo) CreateTask(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskRepo) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskRepo) UpdateTask(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskRepo) DeleteTask(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskRepo) ListTasks(ctx context.Context, userStoryID string) ([]*domain.Task, error) {
	args := m.Called(ctx, userStoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

// IT-017: `create_task` tool call
func TestTaskTools_CreateTask(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	registry := mcp.NewToolRegistry()
	RegisterTaskTools(registry, mockRepo)

	ctx := context.Background()
	now := time.Now()
	userStoryID := "123e4567-e89b-12d3-a456-426614174000"

	req := map[string]interface{}{
		"userStoryId": userStoryID,
		"title":       "Test Task",
		"description": "Desc",
		"status":      "pending",
	}
	reqBytes, _ := json.Marshal(req)

	expectedTask := &domain.Task{
		ID:          "223e4567-e89b-12d3-a456-426614174000",
		UserStoryID: userStoryID,
		Title:       "Test Task",
		Description: "Desc",
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockRepo.On("CreateTask", ctx, mock.AnythingOfType("*domain.Task")).Return(expectedTask, nil)

	toolHandler, ok := registry.GetTool("create_task")
	assert.True(t, ok)
	result, err := toolHandler(ctx, reqBytes)
	assert.NoError(t, err)

	resp, ok := result.(TaskResponse)
	assert.True(t, ok)
	assert.Equal(t, expectedTask.ID, resp.ID)
	assert.Equal(t, expectedTask.Title, resp.Title)

	mockRepo.AssertExpectations(t)
}

// IT-018: `get_task` tool call
func TestTaskTools_GetTask(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	registry := mcp.NewToolRegistry()
	RegisterTaskTools(registry, mockRepo)

	ctx := context.Background()
	now := time.Now()
	id := "223e4567-e89b-12d3-a456-426614174000"

	req := map[string]interface{}{
		"id": id,
	}
	reqBytes, _ := json.Marshal(req)

	expectedTask := &domain.Task{
		ID:          id,
		UserStoryID: "123e4567-e89b-12d3-a456-426614174000",
		Title:       "Test Task",
		Description: "Desc",
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockRepo.On("GetTask", ctx, id).Return(expectedTask, nil)

	toolHandler, ok := registry.GetTool("get_task")
	assert.True(t, ok)
	result, err := toolHandler(ctx, reqBytes)
	assert.NoError(t, err)

	resp, ok := result.(TaskResponse)
	assert.True(t, ok)
	assert.Equal(t, expectedTask.ID, resp.ID)
	assert.Equal(t, expectedTask.Title, resp.Title)

	mockRepo.AssertExpectations(t)
}

// IT-019: `update_task` tool call
func TestTaskTools_UpdateTask(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	registry := mcp.NewToolRegistry()
	RegisterTaskTools(registry, mockRepo)

	ctx := context.Background()
	now := time.Now()
	id := "223e4567-e89b-12d3-a456-426614174000"

	req := map[string]interface{}{
		"id":     id,
		"title":  "Updated Task",
		"status": "in_progress",
	}
	reqBytes, _ := json.Marshal(req)

	existingTask := &domain.Task{
		ID:          id,
		UserStoryID: "123e4567-e89b-12d3-a456-426614174000",
		Title:       "Test Task",
		Description: "Desc",
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	updatedTask := &domain.Task{
		ID:          id,
		UserStoryID: "123e4567-e89b-12d3-a456-426614174000",
		Title:       "Updated Task",
		Description: "Desc",
		Status:      "in_progress",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockRepo.On("GetTask", ctx, id).Return(existingTask, nil)
	mockRepo.On("UpdateTask", ctx, mock.MatchedBy(func(t *domain.Task) bool {
		return t.Title == "Updated Task" && t.Status == "in_progress" && t.Description == "Desc"
	})).Return(updatedTask, nil)

	toolHandler, ok := registry.GetTool("update_task")
	assert.True(t, ok)
	result, err := toolHandler(ctx, reqBytes)
	assert.NoError(t, err)

	resp, ok := result.(TaskResponse)
	assert.True(t, ok)
	assert.Equal(t, updatedTask.ID, resp.ID)
	assert.Equal(t, updatedTask.Title, resp.Title)
	assert.Equal(t, updatedTask.Status, resp.Status)

	mockRepo.AssertExpectations(t)
}

// IT-020: `delete_task` tool call
func TestTaskTools_DeleteTask(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	registry := mcp.NewToolRegistry()
	RegisterTaskTools(registry, mockRepo)

	ctx := context.Background()
	id := "223e4567-e89b-12d3-a456-426614174000"

	req := map[string]interface{}{
		"id": id,
	}
	reqBytes, _ := json.Marshal(req)

	mockRepo.On("DeleteTask", ctx, id).Return(nil)

	toolHandler, ok := registry.GetTool("delete_task")
	assert.True(t, ok)
	result, err := toolHandler(ctx, reqBytes)
	assert.NoError(t, err)

	resp, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.True(t, resp["success"].(bool))

	mockRepo.AssertExpectations(t)
}

// IT-021: `list_tasks` tool call
func TestTaskTools_ListTasks(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	registry := mcp.NewToolRegistry()
	RegisterTaskTools(registry, mockRepo)

	ctx := context.Background()
	now := time.Now()
	userStoryID := "123e4567-e89b-12d3-a456-426614174000"

	req := map[string]interface{}{
		"userStoryId": userStoryID,
	}
	reqBytes, _ := json.Marshal(req)

	expectedTasks := []*domain.Task{
		{
			ID:          "111",
			UserStoryID: userStoryID,
			Title:       "Task 1",
			Status:      "pending",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	mockRepo.On("ListTasks", ctx, userStoryID).Return(expectedTasks, nil)

	toolHandler, ok := registry.GetTool("list_tasks")
	assert.True(t, ok)
	result, err := toolHandler(ctx, reqBytes)
	assert.NoError(t, err)

	respMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	tasksList := respMap["tasks"].([]TaskResponse)
	assert.Len(t, tasksList, 1)
	assert.Equal(t, expectedTasks[0].ID, tasksList[0].ID)

	mockRepo.AssertExpectations(t)
}
