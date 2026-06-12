package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"agent-board/internal/domain"
	"agent-board/internal/mcp"
	"agent-board/internal/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func (m *MockTaskRepo) UpdateTaskStatus(ctx context.Context, id, fromStatus, toStatus string) (*domain.Task, error) {
	args := m.Called(ctx, id, fromStatus, toStatus)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
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

// IT-019: `update_task` tool call -- title-only update (no status change)
func TestTaskTools_UpdateTask(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	registry := mcp.NewToolRegistry()
	RegisterTaskTools(registry, mockRepo)

	ctx := context.Background()
	now := time.Now()
	id := "223e4567-e89b-12d3-a456-426614174000"

	req := map[string]interface{}{
		"id":    id,
		"title": "Updated Task",
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
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockRepo.On("GetTask", ctx, id).Return(existingTask, nil)
	mockRepo.On("UpdateTask", ctx, mock.MatchedBy(func(t *domain.Task) bool {
		return t.Title == "Updated Task" && t.Status == "pending" && t.Description == "Desc"
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

// IT-001: Reject invalid state transition in update_task
func TestTaskTools_UpdateTask_InvalidTransition(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	registry := mcp.NewToolRegistry()
	RegisterTaskTools(registry, mockRepo)

	ctx := context.Background()
	now := time.Now()
	id := "223e4567-e89b-12d3-a456-426614174000"

	// Request: pending -> completed (invalid)
	req := map[string]interface{}{
		"id":     id,
		"status": "completed",
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

	mockRepo.On("GetTask", ctx, id).Return(existingTask, nil)
	// UpdateTask and UpdateTaskStatus should NOT be called because the transition is invalid

	toolHandler, ok := registry.GetTool("update_task")
	assert.True(t, ok)
	result, err := toolHandler(ctx, reqBytes)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid transition")

	mockRepo.AssertExpectations(t)
}

// IT-001b: Valid transition is accepted in update_task and uses UpdateTaskStatus
func TestTaskTools_UpdateTask_ValidTransition(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	registry := mcp.NewToolRegistry()
	RegisterTaskTools(registry, mockRepo)

	ctx := context.Background()
	now := time.Now()
	id := "223e4567-e89b-12d3-a456-426614174000"

	// Request: pending -> in_progress (valid)
	req := map[string]interface{}{
		"id":     id,
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
		Title:       "Test Task",
		Description: "Desc",
		Status:      "in_progress",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockRepo.On("GetTask", ctx, id).Return(existingTask, nil)
	mockRepo.On("UpdateTaskStatus", ctx, id, "pending", "in_progress").Return(updatedTask, nil)

	toolHandler, ok := registry.GetTool("update_task")
	assert.True(t, ok)
	result, err := toolHandler(ctx, reqBytes)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resp, ok := result.(TaskResponse)
	assert.True(t, ok)
	assert.Equal(t, "in_progress", resp.Status)

	mockRepo.AssertExpectations(t)
}

// UT-005b: create_task enforces pending initial status (handler layer)
func TestTaskTools_CreateTask_EnforcesInitialStatus(t *testing.T) {
	mockRepo := new(MockTaskRepo)
	registry := mcp.NewToolRegistry()
	RegisterTaskTools(registry, mockRepo)

	ctx := context.Background()

	// Providing non-pending status should fail
	req := map[string]interface{}{
		"userStoryId": "123e4567-e89b-12d3-a456-426614174000",
		"title":       "Test Task",
		"description": "Desc",
		"status":      "in_progress",
	}
	reqBytes, _ := json.Marshal(req)

	toolHandler, ok := registry.GetTool("create_task")
	assert.True(t, ok)
	result, err := toolHandler(ctx, reqBytes)
	assert.Error(t, err)
	assert.Nil(t, result)

	// Repo should NOT be called
	mockRepo.AssertNotCalled(t, "CreateTask")
}

// --- US006 verbatim test functions (UT-001..UT-025) ---

// UT-001 — TestRegisterTaskTools_RegistersAllFiveTools
func TestRegisterTaskTools_RegistersAllFiveTools(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	for _, name := range []string{"create_task", "get_task", "update_task", "delete_task", "list_tasks"} {
		handler, ok := registry.GetTool(name)
		assert.True(t, ok, "expected tool %q to be registered", name)
		assert.NotNil(t, handler)
	}

	h, ok := registry.GetTool("nonexistent_tool")
	assert.False(t, ok)
	assert.Nil(t, h)
}

// UT-002 — TestCreateTaskTool_InvalidArguments
func TestCreateTaskTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	tool, ok := registry.GetTool("create_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage("not-valid-json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-003 — TestCreateTaskTool_MissingUserStoryIDOrTitle
func TestCreateTaskTool_MissingUserStoryIDOrTitle(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	tool, ok := registry.GetTool("create_task")
	require.True(t, ok)

	// missing title
	_, err := tool(context.Background(), json.RawMessage(`{"userStoryId": "us-1", "title": ""}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "userStoryId and title are required")

	// missing userStoryId
	_, err = tool(context.Background(), json.RawMessage(`{"userStoryId": "", "title": "T"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "userStoryId and title are required")
}

// UT-004 — TestCreateTaskTool_DefaultStatusWhenOmitted
func TestCreateTaskTool_DefaultStatusWhenOmitted(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	var capturedTask *domain.Task
	mockRepo.On("CreateTask", mock.Anything, mock.AnythingOfType("*domain.Task")).
		Run(func(args mock.Arguments) {
			capturedTask = args.Get(1).(*domain.Task)
		}).
		Return(&domain.Task{
			ID:          "t-1",
			UserStoryID: "us-1",
			Title:       "Do thing",
			Status:      domain.TaskStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}, nil)

	tool, ok := registry.GetTool("create_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"userStoryId": "us-1", "title": "Do thing"}`))
	require.NoError(t, err)
	require.NotNil(t, capturedTask)
	assert.Equal(t, domain.TaskStatusPending, capturedTask.Status)

	mockRepo.AssertExpectations(t)
}

// UT-005 — TestCreateTaskTool_InvalidInitialStatus
func TestCreateTaskTool_InvalidInitialStatus(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	tool, ok := registry.GetTool("create_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"userStoryId": "us-1", "title": "T", "status": "in_progress"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid initial status:")
}

// UT-006 — TestCreateTaskTool_RepoError
func TestCreateTaskTool_RepoError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	mockRepo.On("CreateTask", mock.Anything, mock.AnythingOfType("*domain.Task")).
		Return(nil, errors.New("db down"))

	tool, ok := registry.GetTool("create_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"userStoryId": "us-1", "title": "T"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create task:")

	mockRepo.AssertExpectations(t)
}

// UT-007 — TestGetTaskTool_InvalidArguments
func TestGetTaskTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	tool, ok := registry.GetTool("get_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage("not-valid-json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-008 — TestGetTaskTool_EmptyID
func TestGetTaskTool_EmptyID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	tool, ok := registry.GetTool("get_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"id": ""}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

// UT-009 — TestGetTaskTool_NotFound
func TestGetTaskTool_NotFound(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	mockRepo.On("GetTask", mock.Anything, "task-1").Return(nil, repo.ErrNotFound)

	tool, ok := registry.GetTool("get_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"id": "task-1"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
	assert.False(t, errors.Is(err, repo.ErrNotFound))

	mockRepo.AssertExpectations(t)
}

// UT-010 — TestGetTaskTool_GenericError
func TestGetTaskTool_GenericError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	mockRepo.On("GetTask", mock.Anything, "task-1").Return(nil, errors.New("db down"))

	tool, ok := registry.GetTool("get_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"id": "task-1"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get task:")

	mockRepo.AssertExpectations(t)
}

// UT-011 — TestUpdateTaskTool_InvalidArguments
func TestUpdateTaskTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage("not-valid-json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-012 — TestUpdateTaskTool_EmptyID
func TestUpdateTaskTool_EmptyID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"id": ""}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

// UT-013 — TestUpdateTaskTool_NotFoundOnInitialGet
func TestUpdateTaskTool_NotFoundOnInitialGet(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	mockRepo.On("GetTask", mock.Anything, "task-1").Return(nil, repo.ErrNotFound)

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"id": "task-1"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "task not found")
	assert.False(t, errors.Is(err, repo.ErrNotFound))

	mockRepo.AssertExpectations(t)
}

// UT-014 — TestUpdateTaskTool_GenericErrorOnInitialGet
func TestUpdateTaskTool_GenericErrorOnInitialGet(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	mockRepo.On("GetTask", mock.Anything, "task-1").Return(nil, errors.New("db down"))

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"id": "task-1"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get task:")

	mockRepo.AssertExpectations(t)
}

// UT-015 — TestUpdateTaskTool_InvalidStatusTransition
// pending → done is invalid per domain state machine (pending only allows → in_progress)
func TestUpdateTaskTool_InvalidStatusTransition(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	existing := &domain.Task{
		ID:          "task-1",
		UserStoryID: "us-1",
		Title:       "T",
		Status:      domain.TaskStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	mockRepo.On("GetTask", mock.Anything, "task-1").Return(existing, nil)

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"id": "task-1", "status": "done"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transition from pending to done")

	mockRepo.AssertExpectations(t)
}

// UT-016 — TestUpdateTaskTool_StatusChange_FieldUpdateError
// valid transition AND title update; UpdateTask (field update) returns error
func TestUpdateTaskTool_StatusChange_FieldUpdateError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	existing := &domain.Task{
		ID:          "task-1",
		UserStoryID: "us-1",
		Title:       "T",
		Status:      domain.TaskStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	mockRepo.On("GetTask", mock.Anything, "task-1").Return(existing, nil)
	mockRepo.On("UpdateTask", mock.Anything, mock.AnythingOfType("*domain.Task")).
		Return(nil, errors.New("db down"))

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"id": "task-1", "status": "in_progress", "title": "Updated"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update task fields:")

	mockRepo.AssertExpectations(t)
}

// UT-017 — TestUpdateTaskTool_StatusChange_UpdateTaskStatusError
// valid transition, no field changes; UpdateTaskStatus returns error
func TestUpdateTaskTool_StatusChange_UpdateTaskStatusError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	existing := &domain.Task{
		ID:          "task-1",
		UserStoryID: "us-1",
		Title:       "T",
		Status:      domain.TaskStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	mockRepo.On("GetTask", mock.Anything, "task-1").Return(existing, nil)
	mockRepo.On("UpdateTaskStatus", mock.Anything, "task-1", domain.TaskStatusPending, domain.TaskStatusInProgress).
		Return(nil, errors.New("db down"))

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"id": "task-1", "status": "in_progress"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update task status:")

	mockRepo.AssertExpectations(t)
}

// UT-018 — TestUpdateTaskTool_NoStatusChange_RepoUpdateError
// no status change, field update; UpdateTask returns error
func TestUpdateTaskTool_NoStatusChange_RepoUpdateError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	existing := &domain.Task{
		ID:          "task-1",
		UserStoryID: "us-1",
		Title:       "T",
		Status:      domain.TaskStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	mockRepo.On("GetTask", mock.Anything, "task-1").Return(existing, nil)
	mockRepo.On("UpdateTask", mock.Anything, mock.AnythingOfType("*domain.Task")).
		Return(nil, errors.New("db down"))

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"id": "task-1", "title": "New title"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update task:")

	mockRepo.AssertExpectations(t)
}

// UT-019 — TestUpdateTaskTool_StatusChange_HappyPath
// pending → in_progress, no field changes; full success
func TestUpdateTaskTool_StatusChange_HappyPath(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	now := time.Now()
	existing := &domain.Task{
		ID:          "task-1",
		UserStoryID: "us-1",
		Title:       "T",
		Status:      domain.TaskStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	updated := &domain.Task{
		ID:          "task-1",
		UserStoryID: "us-1",
		Title:       "T",
		Status:      domain.TaskStatusInProgress,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	mockRepo.On("GetTask", mock.Anything, "task-1").Return(existing, nil)
	mockRepo.On("UpdateTaskStatus", mock.Anything, "task-1", domain.TaskStatusPending, domain.TaskStatusInProgress).
		Return(updated, nil)

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "task-1", "status": "in_progress"}`))
	require.NoError(t, err)
	require.NotNil(t, result)

	resp, ok := result.(TaskResponse)
	require.True(t, ok)
	assert.Equal(t, domain.TaskStatusInProgress, resp.Status)

	mockRepo.AssertExpectations(t)
}

// UT-020 — TestDeleteTaskTool_InvalidArguments
func TestDeleteTaskTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	tool, ok := registry.GetTool("delete_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage("not-valid-json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-021 — TestDeleteTaskTool_EmptyID
func TestDeleteTaskTool_EmptyID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	tool, ok := registry.GetTool("delete_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"id": ""}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

// UT-022 — TestDeleteTaskTool_RepoError
func TestDeleteTaskTool_RepoError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	mockRepo.On("DeleteTask", mock.Anything, "task-1").Return(errors.New("db down"))

	tool, ok := registry.GetTool("delete_task")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"id": "task-1"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete task:")

	mockRepo.AssertExpectations(t)
}

// UT-023 — TestListTasksTool_InvalidArguments
func TestListTasksTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	tool, ok := registry.GetTool("list_tasks")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage("not-valid-json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-024 — TestListTasksTool_MissingUserStoryID
func TestListTasksTool_MissingUserStoryID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	tool, ok := registry.GetTool("list_tasks")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"userStoryId": ""}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "userStoryId is required")
}

// UT-025 — TestListTasksTool_RepoError
func TestListTasksTool_RepoError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	mockRepo.On("ListTasks", mock.Anything, "us-1").Return(nil, errors.New("db down"))

	tool, ok := registry.GetTool("list_tasks")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"userStoryId": "us-1"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list tasks:")

	mockRepo.AssertExpectations(t)
}

// TestUpdateTaskTool_NoStatusChange_HappyPath exercises the update path where no
// status field is sent and UpdateTask succeeds — covers the success return on the
// no-status-change branch (required for ≥95% statement coverage per IT-001).
func TestUpdateTaskTool_NoStatusChange_HappyPath(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	now := time.Now()
	existing := &domain.Task{
		ID:          "task-1",
		UserStoryID: "us-1",
		Title:       "T",
		Status:      domain.TaskStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	updated := &domain.Task{
		ID:          "task-1",
		UserStoryID: "us-1",
		Title:       "New title",
		Status:      domain.TaskStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	mockRepo.On("GetTask", mock.Anything, "task-1").Return(existing, nil)
	mockRepo.On("UpdateTask", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(updated, nil)

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "task-1", "title": "New title"}`))
	require.NoError(t, err)
	require.NotNil(t, result)

	resp, ok := result.(TaskResponse)
	require.True(t, ok)
	assert.Equal(t, "New title", resp.Title)
	assert.Equal(t, domain.TaskStatusPending, resp.Status)

	mockRepo.AssertExpectations(t)
}

// TestListTasksTool_HappyPath exercises the list_tasks success path including the
// for loop over task results — required for ≥95% statement coverage per IT-001.
func TestListTasksTool_HappyPath(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	now := time.Now()
	tasks := []*domain.Task{
		{
			ID:          "t-1",
			UserStoryID: "us-1",
			Title:       "Task 1",
			Status:      domain.TaskStatusPending,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "t-2",
			UserStoryID: "us-1",
			Title:       "Task 2",
			Status:      domain.TaskStatusInProgress,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	mockRepo.On("ListTasks", mock.Anything, "us-1").Return(tasks, nil)

	tool, ok := registry.GetTool("list_tasks")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"userStoryId": "us-1"}`))
	require.NoError(t, err)
	require.NotNil(t, result)

	respMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	taskList, ok := respMap["tasks"].([]TaskResponse)
	require.True(t, ok)
	assert.Len(t, taskList, 2)
	assert.Equal(t, "t-1", taskList[0].ID)
	assert.Equal(t, "t-2", taskList[1].ID)

	mockRepo.AssertExpectations(t)
}

// --- end US006 verbatim test functions ---

// --- US049 tests: blocked_review_gate MCP handler ---

// UT-049-010 — MCP update_task accepts blocked_review_gate from in_review
func TestUpdateTaskTool_BlockedReviewGate_FromInReview(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	now := time.Now()
	existing := &domain.Task{
		ID:          "task-1",
		UserStoryID: "us-1",
		Title:       "T",
		Status:      domain.TaskStatusInReview,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	updated := &domain.Task{
		ID:          "task-1",
		UserStoryID: "us-1",
		Title:       "T",
		Status:      domain.TaskStatusBlockedReviewGate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	mockRepo.On("GetTask", mock.Anything, "task-1").Return(existing, nil)
	mockRepo.On("UpdateTaskStatus", mock.Anything, "task-1", domain.TaskStatusInReview, domain.TaskStatusBlockedReviewGate).
		Return(updated, nil)

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "task-1", "status": "blocked_review_gate"}`))
	require.NoError(t, err)
	require.NotNil(t, result)

	resp, ok := result.(TaskResponse)
	require.True(t, ok)
	assert.Equal(t, domain.TaskStatusBlockedReviewGate, resp.Status)

	mockRepo.AssertExpectations(t)
}

// UT-049-011 — MCP update_task accepts blocked_review_gate from changes_requested
func TestUpdateTaskTool_BlockedReviewGate_FromChangesRequested(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	now := time.Now()
	existing := &domain.Task{
		ID:          "task-1",
		UserStoryID: "us-1",
		Title:       "T",
		Status:      domain.TaskStatusChangesRequested,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	updated := &domain.Task{
		ID:          "task-1",
		UserStoryID: "us-1",
		Title:       "T",
		Status:      domain.TaskStatusBlockedReviewGate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	mockRepo.On("GetTask", mock.Anything, "task-1").Return(existing, nil)
	mockRepo.On("UpdateTaskStatus", mock.Anything, "task-1", domain.TaskStatusChangesRequested, domain.TaskStatusBlockedReviewGate).
		Return(updated, nil)

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "task-1", "status": "blocked_review_gate"}`))
	require.NoError(t, err)
	require.NotNil(t, result)

	resp, ok := result.(TaskResponse)
	require.True(t, ok)
	assert.Equal(t, domain.TaskStatusBlockedReviewGate, resp.Status)

	mockRepo.AssertExpectations(t)
}

// UT-049-012 — MCP update_task rejects blocked_review_gate from pending
func TestUpdateTaskTool_BlockedReviewGate_FromPending_Rejected(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	now := time.Now()
	existing := &domain.Task{
		ID:          "task-1",
		UserStoryID: "us-1",
		Title:       "T",
		Status:      domain.TaskStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	mockRepo.On("GetTask", mock.Anything, "task-1").Return(existing, nil)
	// UpdateTask and UpdateTaskStatus must NOT be called

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "task-1", "status": "blocked_review_gate"}`))
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid transition")

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "UpdateTask")
	mockRepo.AssertNotCalled(t, "UpdateTaskStatus")
}

// IT-049-001 — MCP update_task persists blocked_review_gate status (via mock repo, testing handler→repo boundary)
func TestUpdateTaskTool_BlockedReviewGate_Persisted(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	now := time.Now()
	taskID := "task-it049-001"
	existing := &domain.Task{
		ID:          taskID,
		UserStoryID: "us-1",
		Title:       "Gate blocked task",
		Status:      domain.TaskStatusInReview,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	persisted := &domain.Task{
		ID:          taskID,
		UserStoryID: "us-1",
		Title:       "Gate blocked task",
		Status:      domain.TaskStatusBlockedReviewGate,
		CreatedAt:   now,
		UpdatedAt:   now.Add(time.Second),
	}
	mockRepo.On("GetTask", mock.Anything, taskID).Return(existing, nil)
	mockRepo.On("UpdateTaskStatus", mock.Anything, taskID, domain.TaskStatusInReview, domain.TaskStatusBlockedReviewGate).
		Return(persisted, nil)

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "`+taskID+`", "status": "blocked_review_gate"}`))
	require.NoError(t, err)
	require.NotNil(t, result)

	resp, ok := result.(TaskResponse)
	require.True(t, ok)
	assert.Equal(t, domain.TaskStatusBlockedReviewGate, resp.Status)
	// updated_at is later than created_at (both serialised through RFC3339)
	assert.NotEqual(t, resp.CreatedAt, resp.UpdatedAt, "UpdatedAt should be later than CreatedAt after status update")

	mockRepo.AssertExpectations(t)
}

// IT-049-002 — MCP update_task rejects further transitions out of blocked_review_gate (terminal)
func TestUpdateTaskTool_BlockedReviewGate_IsTerminal(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := new(MockTaskRepo)
	RegisterTaskTools(registry, mockRepo)

	now := time.Now()
	taskID := "task-it049-002"
	// Task is already in blocked_review_gate
	existing := &domain.Task{
		ID:          taskID,
		UserStoryID: "us-1",
		Title:       "Terminal task",
		Status:      domain.TaskStatusBlockedReviewGate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	mockRepo.On("GetTask", mock.Anything, taskID).Return(existing, nil)
	// UpdateTask and UpdateTaskStatus must NOT be called

	tool, ok := registry.GetTool("update_task")
	require.True(t, ok)

	result, err := tool(context.Background(), json.RawMessage(`{"id": "`+taskID+`", "status": "in_progress"}`))
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid transition")

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "UpdateTaskStatus")
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
