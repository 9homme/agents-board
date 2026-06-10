package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"agent-board/internal/domain"
	"agent-board/internal/handler"
	"agent-board/internal/mcp"
	"agent-board/internal/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockUserStoryRepo is a mock implementation of repo.UserStoryRepository.
type MockUserStoryRepo struct {
	repo.UserStoryRepository
	CreateUserStoryFunc       func(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error)
	GetUserStoryFunc          func(ctx context.Context, id string) (*domain.UserStory, error)
	UpdateUserStoryFunc       func(ctx context.Context, u *domain.UserStory) (*domain.UserStory, error)
	UpdateUserStoryStatusFunc func(ctx context.Context, id, fromStatus, toStatus string) (*domain.UserStory, error)
	DeleteUserStoryFunc       func(ctx context.Context, id string) error
	ListUserStoriesFunc       func(ctx context.Context, projectID string) ([]*domain.UserStory, error)
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
func (m *MockUserStoryRepo) UpdateUserStoryStatus(ctx context.Context, id, fromStatus, toStatus string) (*domain.UserStory, error) {
	return m.UpdateUserStoryStatusFunc(ctx, id, fromStatus, toStatus)
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

// UT-005 — create_user_story rejects non-draft initial status
func TestUserStoryTools_CreateUserStory_InvalidInitialStatus(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	// Providing non-draft status should be rejected
	args := json.RawMessage(`{"projectId":"123e4567-e89b-12d3-a456-426614174000", "title": "My Story", "description": "Desc", "status": "done"}`)
	toolHandler, ok := registry.GetTool("create_user_story")
	require.True(t, ok)
	res, err := toolHandler(context.Background(), args)
	require.Error(t, err)
	assert.Nil(t, res)
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

// TestUserStoryTools_UpdateUserStory tests a valid status transition (draft -> in_development).
// When status changes, UpdateUserStoryStatus (transactional) is called.
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

	// Status changes use UpdateUserStoryStatus (transactional audit trail).
	mockRepo.UpdateUserStoryStatusFunc = func(ctx context.Context, id, fromStatus, toStatus string) (*domain.UserStory, error) {
		return &domain.UserStory{
			ID:          id,
			ProjectID:   "123e4567-e89b-12d3-a456-426614174000",
			Title:       "Old",
			Description: "Old desc",
			Status:      toStatus,
			CreatedAt:   now,
			UpdatedAt:   now,
		}, nil
	}

	// Non-status field changes use UpdateUserStory.
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

	// Valid transition: draft -> in_development
	args := json.RawMessage(`{"id":"223e4567-e89b-12d3-a456-426614174000", "title": "Updated", "description": "Updated desc", "status": "in_development"}`)
	toolHandler, ok := registry.GetTool("update_user_story")
	require.True(t, ok)
	res, err := toolHandler(context.Background(), args)
	require.NoError(t, err)

	resResp, ok := res.(handler.UserStoryResponse)
	require.True(t, ok)
	assert.Equal(t, "in_development", resResp.Status)
}

// IT-001 — Reject invalid transitions at MCP layer
func TestUserStoryTools_UpdateUserStory_InvalidTransition(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	now := time.Now()
	mockRepo.GetUserStoryFunc = func(ctx context.Context, id string) (*domain.UserStory, error) {
		return &domain.UserStory{
			ID:          id,
			ProjectID:   "123e4567-e89b-12d3-a456-426614174000",
			Title:       "Story",
			Description: "Desc",
			Status:      "draft",
			CreatedAt:   now,
			UpdatedAt:   now,
		}, nil
	}

	// draft -> done is not a valid transition for user stories
	args := json.RawMessage(`{"id":"223e4567-e89b-12d3-a456-426614174000", "status": "done"}`)
	toolHandler, ok := registry.GetTool("update_user_story")
	require.True(t, ok)
	res, err := toolHandler(context.Background(), args)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "invalid transition")
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

// -----------------------------------------------------------------------
// US007 — 27 verbatim test functions (error-mapping + edge-case backfill)
// -----------------------------------------------------------------------

// UT-001
func TestRegisterUserStoryTools_RegistersAllFiveTools(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	for _, name := range []string{
		"create_user_story",
		"get_user_story",
		"update_user_story",
		"delete_user_story",
		"list_user_stories",
	} {
		h, ok := registry.GetTool(name)
		assert.True(t, ok, "expected tool %q to be registered", name)
		assert.NotNil(t, h, "expected non-nil handler for tool %q", name)
	}

	_, ok := registry.GetTool("nonexistent_tool")
	assert.False(t, ok)
}

// UT-002
func TestCreateUserStoryTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	toolHandler, ok := registry.GetTool("create_user_story")
	require.True(t, ok)

	_, err := toolHandler(context.Background(), json.RawMessage("not-valid-json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-003
func TestCreateUserStoryTool_MissingProjectIDOrTitle(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	toolHandler, ok := registry.GetTool("create_user_story")
	require.True(t, ok)

	// Missing projectId
	_, err := toolHandler(context.Background(), json.RawMessage(`{"title":"Some Title"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required fields")

	// Missing title
	_, err = toolHandler(context.Background(), json.RawMessage(`{"projectId":"123e4567-e89b-12d3-a456-426614174000"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required fields")
}

// UT-004
func TestCreateUserStoryTool_DefaultStatusWhenOmitted(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	var capturedStory *domain.UserStory
	mockRepo.CreateUserStoryFunc = func(_ context.Context, s *domain.UserStory) (*domain.UserStory, error) {
		capturedStory = s
		return &domain.UserStory{
			ID:          "aaa",
			ProjectID:   s.ProjectID,
			Title:       s.Title,
			Description: s.Description,
			Status:      s.Status,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}, nil
	}

	toolHandler, ok := registry.GetTool("create_user_story")
	require.True(t, ok)

	_, err := toolHandler(context.Background(), json.RawMessage(`{"projectId":"pid-1","title":"My Story"}`))
	require.NoError(t, err)
	require.NotNil(t, capturedStory)
	assert.Equal(t, domain.UserStoryStatusDraft, capturedStory.Status)
}

// UT-005
func TestCreateUserStoryTool_InvalidInitialStatus(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	toolHandler, ok := registry.GetTool("create_user_story")
	require.True(t, ok)

	_, err := toolHandler(context.Background(), json.RawMessage(`{"projectId":"pid-1","title":"Story","status":"in_signoff"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid initial status:")
}

// UT-006
func TestCreateUserStoryTool_RepoError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	mockErr := errors.New("db down")
	mockRepo.CreateUserStoryFunc = func(_ context.Context, _ *domain.UserStory) (*domain.UserStory, error) {
		return nil, mockErr
	}

	toolHandler, ok := registry.GetTool("create_user_story")
	require.True(t, ok)

	_, returnedErr := toolHandler(context.Background(), json.RawMessage(`{"projectId":"pid-1","title":"Story"}`))
	require.Error(t, returnedErr)
	assert.True(t, errors.Is(returnedErr, mockErr), "expected passthrough error, got: %v", returnedErr)
}

// UT-007 (happy-path companion — not in the 27 error-case names but required for ≥95% coverage of line 94)
func TestGetUserStoryTool_HappyPath(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	now := time.Now()
	mockRepo.GetUserStoryFunc = func(_ context.Context, id string) (*domain.UserStory, error) {
		return &domain.UserStory{
			ID:          id,
			ProjectID:   "pid-1",
			Title:       "Story",
			Description: "Desc",
			Status:      domain.UserStoryStatusDraft,
			CreatedAt:   now,
			UpdatedAt:   now,
		}, nil
	}

	toolHandler, ok := registry.GetTool("get_user_story")
	require.True(t, ok)

	res, err := toolHandler(context.Background(), json.RawMessage(`{"id":"some-id"}`))
	require.NoError(t, err)

	resp, ok := res.(handler.UserStoryResponse)
	require.True(t, ok)
	assert.Equal(t, "some-id", resp.ID)
	assert.Equal(t, domain.UserStoryStatusDraft, resp.Status)
}

// UT-007
func TestGetUserStoryTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	toolHandler, ok := registry.GetTool("get_user_story")
	require.True(t, ok)

	_, err := toolHandler(context.Background(), json.RawMessage("not-valid-json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-008
func TestGetUserStoryTool_MissingID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	toolHandler, ok := registry.GetTool("get_user_story")
	require.True(t, ok)

	_, err := toolHandler(context.Background(), json.RawMessage(`{"id":""}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing id")
}

// UT-009
func TestGetUserStoryTool_NotFound(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	mockRepo.GetUserStoryFunc = func(_ context.Context, _ string) (*domain.UserStory, error) {
		return nil, repo.ErrNotFound
	}

	toolHandler, ok := registry.GetTool("get_user_story")
	require.True(t, ok)

	_, err := toolHandler(context.Background(), json.RawMessage(`{"id":"some-id"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user story not found")
	assert.False(t, errors.Is(err, repo.ErrNotFound), "handler should NOT passthrough ErrNotFound sentinel")
}

// UT-010
func TestGetUserStoryTool_GenericError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	mockErr := errors.New("db down")
	mockRepo.GetUserStoryFunc = func(_ context.Context, _ string) (*domain.UserStory, error) {
		return nil, mockErr
	}

	toolHandler, ok := registry.GetTool("get_user_story")
	require.True(t, ok)

	_, returnedErr := toolHandler(context.Background(), json.RawMessage(`{"id":"some-id"}`))
	require.Error(t, returnedErr)
	assert.True(t, errors.Is(returnedErr, mockErr), "expected passthrough error, got: %v", returnedErr)
}

// UT-011
func TestUpdateUserStoryTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	toolHandler, ok := registry.GetTool("update_user_story")
	require.True(t, ok)

	_, err := toolHandler(context.Background(), json.RawMessage("not-valid-json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-012
func TestUpdateUserStoryTool_MissingID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	toolHandler, ok := registry.GetTool("update_user_story")
	require.True(t, ok)

	_, err := toolHandler(context.Background(), json.RawMessage(`{"id":""}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing id")
}

// UT-013
func TestUpdateUserStoryTool_NotFoundOnInitialGet(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	mockRepo.GetUserStoryFunc = func(_ context.Context, _ string) (*domain.UserStory, error) {
		return nil, repo.ErrNotFound
	}

	toolHandler, ok := registry.GetTool("update_user_story")
	require.True(t, ok)

	_, err := toolHandler(context.Background(), json.RawMessage(`{"id":"some-id"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user story not found")
	assert.False(t, errors.Is(err, repo.ErrNotFound), "handler should NOT passthrough ErrNotFound sentinel")
}

// UT-014
func TestUpdateUserStoryTool_GenericErrorOnInitialGet(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	mockErr := errors.New("db down")
	mockRepo.GetUserStoryFunc = func(_ context.Context, _ string) (*domain.UserStory, error) {
		return nil, mockErr
	}

	toolHandler, ok := registry.GetTool("update_user_story")
	require.True(t, ok)

	_, returnedErr := toolHandler(context.Background(), json.RawMessage(`{"id":"some-id"}`))
	require.Error(t, returnedErr)
	assert.True(t, errors.Is(returnedErr, mockErr), "expected passthrough error, got: %v", returnedErr)
}

// UT-015
func TestUpdateUserStoryTool_InvalidStatusTransition(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	now := time.Now()
	mockRepo.GetUserStoryFunc = func(_ context.Context, id string) (*domain.UserStory, error) {
		return &domain.UserStory{
			ID:        id,
			ProjectID: "pid-1",
			Title:     "Story",
			Status:    domain.UserStoryStatusDraft,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	toolHandler, ok := registry.GetTool("update_user_story")
	require.True(t, ok)

	// draft -> done is not a valid direct transition
	_, err := toolHandler(context.Background(), json.RawMessage(`{"id":"some-id","status":"done"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid transition from draft to done")
}

// UT-016
func TestUpdateUserStoryTool_StatusChange_UpdateUserStoryStatusError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	now := time.Now()
	mockRepo.GetUserStoryFunc = func(_ context.Context, id string) (*domain.UserStory, error) {
		return &domain.UserStory{
			ID:        id,
			ProjectID: "pid-1",
			Title:     "Story",
			Status:    domain.UserStoryStatusDraft,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	mockErr := errors.New("db down")
	mockRepo.UpdateUserStoryStatusFunc = func(_ context.Context, _, _, _ string) (*domain.UserStory, error) {
		return nil, mockErr
	}

	toolHandler, ok := registry.GetTool("update_user_story")
	require.True(t, ok)

	// draft -> in_development is a valid transition
	_, returnedErr := toolHandler(context.Background(), json.RawMessage(`{"id":"some-id","status":"in_development"}`))
	require.Error(t, returnedErr)
	assert.True(t, errors.Is(returnedErr, mockErr), "expected passthrough error, got: %v", returnedErr)
}

// UT-017
func TestUpdateUserStoryTool_StatusChange_PostStatusFieldUpdateError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	now := time.Now()
	mockRepo.GetUserStoryFunc = func(_ context.Context, id string) (*domain.UserStory, error) {
		return &domain.UserStory{
			ID:        id,
			ProjectID: "pid-1",
			Title:     "Old Title",
			Status:    domain.UserStoryStatusDraft,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	mockRepo.UpdateUserStoryStatusFunc = func(_ context.Context, id, _, toStatus string) (*domain.UserStory, error) {
		return &domain.UserStory{
			ID:        id,
			ProjectID: "pid-1",
			Title:     "Old Title",
			Status:    toStatus,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	mockErr := errors.New("db down")
	mockRepo.UpdateUserStoryFunc = func(_ context.Context, _ *domain.UserStory) (*domain.UserStory, error) {
		return nil, mockErr
	}

	toolHandler, ok := registry.GetTool("update_user_story")
	require.True(t, ok)

	// Status change + title change triggers UpdateUserStoryStatus then UpdateUserStory
	_, returnedErr := toolHandler(context.Background(), json.RawMessage(`{"id":"some-id","status":"in_development","title":"New Title"}`))
	require.Error(t, returnedErr)
	assert.True(t, errors.Is(returnedErr, mockErr), "expected passthrough error, got: %v", returnedErr)
}

// UT-018
func TestUpdateUserStoryTool_StatusChange_HappyPath_NoExtraFields(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	now := time.Now()
	mockRepo.GetUserStoryFunc = func(_ context.Context, id string) (*domain.UserStory, error) {
		return &domain.UserStory{
			ID:        id,
			ProjectID: "pid-1",
			Title:     "Story",
			Status:    domain.UserStoryStatusDraft,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	mockRepo.UpdateUserStoryStatusFunc = func(_ context.Context, id, _, toStatus string) (*domain.UserStory, error) {
		return &domain.UserStory{
			ID:        id,
			ProjectID: "pid-1",
			Title:     "Story",
			Status:    toStatus,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	toolHandler, ok := registry.GetTool("update_user_story")
	require.True(t, ok)

	res, err := toolHandler(context.Background(), json.RawMessage(`{"id":"some-id","status":"in_development"}`))
	require.NoError(t, err)

	resp, ok := res.(handler.UserStoryResponse)
	require.True(t, ok)
	assert.Equal(t, domain.UserStoryStatusInDevelopment, resp.Status)
}

// UT-019
func TestUpdateUserStoryTool_StatusChange_HappyPath_WithExtraFields(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	now := time.Now()
	mockRepo.GetUserStoryFunc = func(_ context.Context, id string) (*domain.UserStory, error) {
		return &domain.UserStory{
			ID:          id,
			ProjectID:   "pid-1",
			Title:       "Old Title",
			Description: "Old Desc",
			Status:      domain.UserStoryStatusDraft,
			CreatedAt:   now,
			UpdatedAt:   now,
		}, nil
	}

	mockRepo.UpdateUserStoryStatusFunc = func(_ context.Context, id, _, toStatus string) (*domain.UserStory, error) {
		return &domain.UserStory{
			ID:          id,
			ProjectID:   "pid-1",
			Title:       "Old Title",
			Description: "Old Desc",
			Status:      toStatus,
			CreatedAt:   now,
			UpdatedAt:   now,
		}, nil
	}

	mockRepo.UpdateUserStoryFunc = func(_ context.Context, u *domain.UserStory) (*domain.UserStory, error) {
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

	toolHandler, ok := registry.GetTool("update_user_story")
	require.True(t, ok)

	res, err := toolHandler(context.Background(), json.RawMessage(`{"id":"some-id","status":"in_development","title":"New Title","description":"New Desc"}`))
	require.NoError(t, err)

	resp, ok := res.(handler.UserStoryResponse)
	require.True(t, ok)
	assert.Equal(t, domain.UserStoryStatusInDevelopment, resp.Status)
	assert.Equal(t, "New Title", resp.Title)
	assert.Equal(t, "New Desc", resp.Description)
}

// UT-020
func TestUpdateUserStoryTool_NoStatusChange_RepoUpdateError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	now := time.Now()
	mockRepo.GetUserStoryFunc = func(_ context.Context, id string) (*domain.UserStory, error) {
		return &domain.UserStory{
			ID:        id,
			ProjectID: "pid-1",
			Title:     "Story",
			Status:    domain.UserStoryStatusDraft,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	mockErr := errors.New("db down")
	mockRepo.UpdateUserStoryFunc = func(_ context.Context, _ *domain.UserStory) (*domain.UserStory, error) {
		return nil, mockErr
	}

	toolHandler, ok := registry.GetTool("update_user_story")
	require.True(t, ok)

	// No status field — only title change; goes to the no-status-change branch
	_, returnedErr := toolHandler(context.Background(), json.RawMessage(`{"id":"some-id","title":"New Title"}`))
	require.Error(t, returnedErr)
	assert.True(t, errors.Is(returnedErr, mockErr), "expected passthrough error, got: %v", returnedErr)
}

// UT-021
func TestDeleteUserStoryTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	toolHandler, ok := registry.GetTool("delete_user_story")
	require.True(t, ok)

	_, err := toolHandler(context.Background(), json.RawMessage("not-valid-json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-022
func TestDeleteUserStoryTool_MissingID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	toolHandler, ok := registry.GetTool("delete_user_story")
	require.True(t, ok)

	_, err := toolHandler(context.Background(), json.RawMessage(`{"id":""}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing id")
}

// UT-023
func TestDeleteUserStoryTool_RepoError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	mockErr := errors.New("db down")
	mockRepo.DeleteUserStoryFunc = func(_ context.Context, _ string) error {
		return mockErr
	}

	toolHandler, ok := registry.GetTool("delete_user_story")
	require.True(t, ok)

	_, returnedErr := toolHandler(context.Background(), json.RawMessage(`{"id":"some-id"}`))
	require.Error(t, returnedErr)
	assert.True(t, errors.Is(returnedErr, mockErr), "expected passthrough error, got: %v", returnedErr)
}

// UT-024
func TestListUserStoriesTool_InvalidArguments(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	toolHandler, ok := registry.GetTool("list_user_stories")
	require.True(t, ok)

	_, err := toolHandler(context.Background(), json.RawMessage("not-valid-json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid arguments")
}

// UT-025
func TestListUserStoriesTool_MissingProjectID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	toolHandler, ok := registry.GetTool("list_user_stories")
	require.True(t, ok)

	_, err := toolHandler(context.Background(), json.RawMessage(`{"projectId":""}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing projectId")
}

// UT-026
func TestListUserStoriesTool_RepoError(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	mockErr := errors.New("db down")
	mockRepo.ListUserStoriesFunc = func(_ context.Context, _ string) ([]*domain.UserStory, error) {
		return nil, mockErr
	}

	toolHandler, ok := registry.GetTool("list_user_stories")
	require.True(t, ok)

	_, returnedErr := toolHandler(context.Background(), json.RawMessage(`{"projectId":"pid-1"}`))
	require.Error(t, returnedErr)
	assert.True(t, errors.Is(returnedErr, mockErr), "expected passthrough error, got: %v", returnedErr)
}

// -----------------------------------------------------------------------
// US045 BREAKING CHANGE tests — create_user_story now requires requirement_id
// -----------------------------------------------------------------------

// UT-045-039 — MCP create_user_story now includes requirement_id in INSERT
func TestCreateUserStoryTool_WithRequirementID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockReqRepo := &MockRequirementRepo{} // used for membership check
	mockUSRepo := &MockUserStoryRepo{}
	now := time.Now()

	projectID := "11111111-1111-1111-1111-111111111111"
	requirementID := "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f"

	// GetRequirement returns a requirement belonging to the same project
	mockReqRepo.GetRequirementFunc = func(_ context.Context, id string) (*domain.Requirement, error) {
		return &domain.Requirement{
			ID:        id,
			ProjectID: projectID,
			Name:      "REQ",
			Status:    "draft",
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	mockUSRepo.CreateUserStoryFunc = func(_ context.Context, u *domain.UserStory) (*domain.UserStory, error) {
		assert.Equal(t, requirementID, u.RequirementID, "RequirementID must be set on INSERT")
		return &domain.UserStory{
			ID:            "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			ProjectID:     u.ProjectID,
			RequirementID: u.RequirementID,
			Title:         u.Title,
			Description:   u.Description,
			Status:        u.Status,
			CreatedAt:     now,
			UpdatedAt:     now,
		}, nil
	}

	handler.RegisterUserStoryTools(registry, mockUSRepo, mockReqRepo)

	args := json.RawMessage(`{"projectId":"11111111-1111-1111-1111-111111111111","requirement_id":"b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f","title":"Add item to basket"}`)
	tool, ok := registry.GetTool("create_user_story")
	require.True(t, ok)

	res, err := tool(context.Background(), args)
	require.NoError(t, err)

	b, err := json.Marshal(res)
	require.NoError(t, err)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &m))
	assert.Equal(t, requirementID, m["requirementId"])
	assert.Equal(t, projectID, m["projectId"])
	assert.Equal(t, "Add item to basket", m["title"])
}

// UT-045-040 — MCP create_user_story — missing requirement_id returns tool error
func TestCreateUserStoryTool_MissingRequirementID(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockReqRepo := &MockRequirementRepo{}
	mockUSRepo := &MockUserStoryRepo{}

	handler.RegisterUserStoryTools(registry, mockUSRepo, mockReqRepo)

	tool, ok := registry.GetTool("create_user_story")
	require.True(t, ok)

	_, err := tool(context.Background(), json.RawMessage(`{"projectId":"proj-1","title":"Story"}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requirement_id")
}

// UT-045-041 — MCP create_user_story — requirement does not belong to project returns tool error
func TestCreateUserStoryTool_RequirementNotInProject(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockReqRepo := &MockRequirementRepo{}
	mockUSRepo := &MockUserStoryRepo{}
	now := time.Now()

	// requirement belongs to a DIFFERENT project
	mockReqRepo.GetRequirementFunc = func(_ context.Context, id string) (*domain.Requirement, error) {
		return &domain.Requirement{
			ID:        id,
			ProjectID: "different-project-id",
			Name:      "REQ",
			Status:    "draft",
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	called := false
	mockUSRepo.CreateUserStoryFunc = func(_ context.Context, _ *domain.UserStory) (*domain.UserStory, error) {
		called = true
		return nil, errors.New("should not be called")
	}

	handler.RegisterUserStoryTools(registry, mockUSRepo, mockReqRepo)

	args := json.RawMessage(`{"projectId":"proj-1","requirement_id":"req-belongs-to-other","title":"Story"}`)
	tool, ok := registry.GetTool("create_user_story")
	require.True(t, ok)

	_, err := tool(context.Background(), args)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requirement does not belong to project")
	assert.False(t, called, "repo must NOT be called when requirement doesn't belong to project")
}

// UT-027
func TestListUserStoriesTool_EmptySliceReturnsEmptyArray(t *testing.T) {
	registry := mcp.NewToolRegistry()
	mockRepo := &MockUserStoryRepo{}
	handler.RegisterUserStoryTools(registry, mockRepo)

	mockRepo.ListUserStoriesFunc = func(_ context.Context, _ string) ([]*domain.UserStory, error) {
		return nil, nil
	}

	toolHandler, ok := registry.GetTool("list_user_stories")
	require.True(t, ok)

	res, err := toolHandler(context.Background(), json.RawMessage(`{"projectId":"pid-1"}`))
	require.NoError(t, err)

	resMap, ok := res.(map[string]interface{})
	require.True(t, ok)

	storiesRaw, exists := resMap["userStories"]
	require.True(t, exists, "expected 'userStories' key in result")

	stories, ok := storiesRaw.([]handler.UserStoryResponse)
	require.True(t, ok, "expected []handler.UserStoryResponse, got %T", storiesRaw)
	assert.NotNil(t, stories, "userStories must not be nil")
	assert.Len(t, stories, 0, "expected empty slice, not nil")
}
