package handler_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"agent-board/internal/domain"
	"agent-board/internal/handler"
	"agent-board/internal/mcp"
	"agent-board/internal/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockUserStoryRepo is a mock implementation of repo.UserStoryRepository
type MockUserStoryRepo struct {
	repo.UserStoryRepository
	CreateUserStoryFunc func(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error)
	GetUserStoryFunc    func(ctx context.Context, id string) (*domain.UserStory, error)
	UpdateUserStoryFunc func(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error)
	DeleteUserStoryFunc func(ctx context.Context, id string) error
	ListUserStoriesFunc func(ctx context.Context, projectID string) ([]*domain.UserStory, error)
}

func (m *MockUserStoryRepo) CreateUserStory(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error) {
	return m.CreateUserStoryFunc(ctx, u)
}
func (m *MockUserStoryRepo) GetUserStory(ctx context.Context, id string) (*domain.UserStory, error) {
	return m.GetUserStoryFunc(ctx, id)
}
func (m *MockUserStoryRepo) UpdateUserStory(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error) {
	return m.UpdateUserStoryFunc(ctx, u)
}
func (m *MockUserStoryRepo) DeleteUserStory(ctx context.Context, id string) error {
	return m.DeleteUserStoryFunc(ctx, id)
}
func (m *MockUserStoryRepo) ListUserStories(ctx context.Context, projectID string) ([]*domain.UserStory, error) {
	return m.ListUserStoriesFunc(ctx, projectID)
}

func TestUserStoryTools_CreateUserStory(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	now := time.Now()
	mockRepo.CreateUserStoryFunc = func(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error) {
		assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", u.ProjectID)
		assert.Equal(t, "My Story", u.Title)
		assert.Equal(t, "Desc", u.Description)
		assert.Equal(t, "draft", u.Status)
		return &domain.UserStory{
			ID:          "223e4567-e89b-12d3-a456-426614174000",
			ProjectID:   u.ProjectID,
			Title:       u.Title,
			Description: u.Description,
			Status:      u.Status,
			CreatedAt:   now,
			UpdatedAt:   now,
		}, nil
	}

	args := json.RawMessage(`{"projectId":"123e4567-e89b-12d3-a456-426614174000", "title": "My Story", "description": "Desc", "status": "draft"}`)
	toolHandler, ok := registry.GetTool("create_user_story")
	require.True(t, ok)
	res, err := toolHandler(context.Background(), args)
	require.NoError(t, err)

	resResp, ok := res.(handler.UserStoryResponse)
	require.True(t, ok)
	assert.Equal(t, "223e4567-e89b-12d3-a456-426614174000", resResp.ID)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", resResp.ProjectID)
	assert.Equal(t, "My Story", resResp.Title)
	assert.Equal(t, "Desc", resResp.Description)
	assert.Equal(t, "draft", resResp.Status)
}

func TestUserStoryTools_GetUserStory(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	now := time.Now()
	mockRepo.GetUserStoryFunc = func(ctx context.Context, id string) (*domain.UserStory, error) {
		if id == "223e4567-e89b-12d3-a456-426614174000" {
			return &domain.UserStory{
				ID:          id,
				ProjectID:   "123e4567-e89b-12d3-a456-426614174000",
				Title:       "My Story",
				Description: "Desc",
				Status:      "draft",
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		}
		return nil, repo.ErrNotFound
	}

	args := json.RawMessage(`{"id":"223e4567-e89b-12d3-a456-426614174000"}`)
	toolHandler, ok := registry.GetTool("get_user_story")
	require.True(t, ok)
	res, err := toolHandler(context.Background(), args)
	require.NoError(t, err)

	resResp, ok := res.(handler.UserStoryResponse)
	require.True(t, ok)
	assert.Equal(t, "223e4567-e89b-12d3-a456-426614174000", resResp.ID)
}

func TestUserStoryTools_UpdateUserStory(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	now := time.Now()
	mockRepo.GetUserStoryFunc = func(ctx context.Context, id string) (*domain.UserStory, error) {
		return &domain.UserStory{
			ID:          id,
			ProjectID:   "123e4567-e89b-12d3-a456-426614174000",
			Title:       "Old",
			Description: "Old desc",
			Status:      "draft",
			CreatedAt:   now,
			UpdatedAt:   now,
		}, nil
	}

	mockRepo.UpdateUserStoryFunc = func(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error) {
		return &domain.UserStory{
			ID:          u.ID,
			ProjectID:   u.ProjectID,
			Title:       u.Title,
			Description: u.Description,
			Status:      u.Status,
			CreatedAt:   now,
			UpdatedAt:   now,
		}, nil
	}

	args := json.RawMessage(`{"id":"223e4567-e89b-12d3-a456-426614174000", "title": "Updated", "description": "Updated desc", "status": "in_progress"}`)
	toolHandler, ok := registry.GetTool("update_user_story")
	require.True(t, ok)
	res, err := toolHandler(context.Background(), args)
	require.NoError(t, err)

	resResp, ok := res.(handler.UserStoryResponse)
	require.True(t, ok)
	assert.Equal(t, "Updated", resResp.Title)
	assert.Equal(t, "Updated desc", resResp.Description)
	assert.Equal(t, "in_progress", resResp.Status)
}

func TestUserStoryTools_DeleteUserStory(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	mockRepo.DeleteUserStoryFunc = func(ctx context.Context, id string) error {
		return nil
	}

	args := json.RawMessage(`{"id":"223e4567-e89b-12d3-a456-426614174000"}`)
	toolHandler, ok := registry.GetTool("delete_user_story")
	require.True(t, ok)
	res, err := toolHandler(context.Background(), args)
	require.NoError(t, err)

	resMap, ok := res.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, resMap["success"])
}

func TestUserStoryTools_ListUserStories(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	now := time.Now()
	mockRepo.ListUserStoriesFunc = func(ctx context.Context, projectID string) ([]*domain.UserStory, error) {
		assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", projectID)
		return []*domain.UserStory{
			{
				ID:          "223e4567-e89b-12d3-a456-426614174000",
				ProjectID:   projectID,
				Title:       "US1",
				Description: "D1",
				Status:      "draft",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}, nil
	}

	args := json.RawMessage(`{"projectId":"123e4567-e89b-12d3-a456-426614174000"}`)
	toolHandler, ok := registry.GetTool("list_user_stories")
	require.True(t, ok)
	res, err := toolHandler(context.Background(), args)
	require.NoError(t, err)

	resMap, ok := res.(map[string]interface{})
	require.True(t, ok)
	stories, ok := resMap["userStories"].([]handler.UserStoryResponse)
	require.True(t, ok)
	assert.Len(t, stories, 1)
	assert.Equal(t, "223e4567-e89b-12d3-a456-426614174000", stories[0].ID)
}
