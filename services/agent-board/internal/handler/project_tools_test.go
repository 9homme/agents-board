package handler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"agent-board/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProjectRepo is a mock of ProjectRepository
type MockProjectRepo struct {
	mock.Mock
}

func (m *MockProjectRepo) CreateProject(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	args := m.Called(ctx, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *MockProjectRepo) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *MockProjectRepo) UpdateProject(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	args := m.Called(ctx, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *MockProjectRepo) DeleteProject(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepo) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Project), args.Error(1)
}

// IT-002: create_project tool call
func TestProjectTools_CreateProject(t *testing.T) {
	mockRepo := new(MockProjectRepo)
	now := time.Now()
	expectedProject := &domain.Project{
		ID:          "123",
		Name:        "Test Project",
		Description: "A test project",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockRepo.On("CreateProject", mock.Anything, &domain.Project{
		Name:        "Test Project",
		Description: "A test project",
	}).Return(expectedProject, nil)

	handler := handleCreateProject(mockRepo)
	args := json.RawMessage(`{"name":"Test Project","description":"A test project"}`)

	result, err := handler(context.Background(), args)
	assert.NoError(t, err)

	resStr, err := json.Marshal(result)
	assert.NoError(t, err)

	var res map[string]interface{}
	err = json.Unmarshal(resStr, &res)
	assert.NoError(t, err)
	assert.Equal(t, "123", res["id"])
	assert.Equal(t, "Test Project", res["name"])

	mockRepo.AssertExpectations(t)
}

// IT-003: get_project tool call
func TestProjectTools_GetProject(t *testing.T) {
	mockRepo := new(MockProjectRepo)
	now := time.Now()
	expectedProject := &domain.Project{
		ID:          "123",
		Name:        "Test Project",
		Description: "A test project",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockRepo.On("GetProject", mock.Anything, "123").Return(expectedProject, nil)

	handler := handleGetProject(mockRepo)
	args := json.RawMessage(`{"id":"123"}`)

	result, err := handler(context.Background(), args)
	assert.NoError(t, err)

	resStr, err := json.Marshal(result)
	assert.NoError(t, err)

	var res map[string]interface{}
	err = json.Unmarshal(resStr, &res)
	assert.NoError(t, err)
	assert.Equal(t, "123", res["id"])
	assert.Equal(t, "Test Project", res["name"])

	mockRepo.AssertExpectations(t)
}

// IT-004: update_project tool call
func TestProjectTools_UpdateProject(t *testing.T) {
	mockRepo := new(MockProjectRepo)
	now := time.Now()
	expectedProject := &domain.Project{
		ID:          "123",
		Name:        "Updated Project",
		Description: "Updated desc",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockRepo.On("GetProject", mock.Anything, "123").Return(&domain.Project{
		ID:          "123",
		Name:        "Old Project",
		Description: "Old desc",
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil)

	mockRepo.On("UpdateProject", mock.Anything, &domain.Project{
		ID:          "123",
		Name:        "Updated Project",
		Description: "Updated desc",
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Return(expectedProject, nil)

	handler := handleUpdateProject(mockRepo)
	args := json.RawMessage(`{"id":"123","name":"Updated Project","description":"Updated desc"}`)

	result, err := handler(context.Background(), args)
	assert.NoError(t, err)

	resStr, err := json.Marshal(result)
	assert.NoError(t, err)

	var res map[string]interface{}
	err = json.Unmarshal(resStr, &res)
	assert.NoError(t, err)
	assert.Equal(t, "123", res["id"])
	assert.Equal(t, "Updated Project", res["name"])

	mockRepo.AssertExpectations(t)
}

// IT-005: delete_project tool call
func TestProjectTools_DeleteProject(t *testing.T) {
	mockRepo := new(MockProjectRepo)

	mockRepo.On("DeleteProject", mock.Anything, "123").Return(nil)

	handler := handleDeleteProject(mockRepo)
	args := json.RawMessage(`{"id":"123"}`)

	result, err := handler(context.Background(), args)
	assert.NoError(t, err)

	resStr, err := json.Marshal(result)
	assert.NoError(t, err)

	var res map[string]interface{}
	err = json.Unmarshal(resStr, &res)
	assert.NoError(t, err)
	assert.Equal(t, true, res["success"])

	mockRepo.AssertExpectations(t)
}

// IT-006: list_projects tool call
func TestProjectTools_ListProjects(t *testing.T) {
	mockRepo := new(MockProjectRepo)
	now := time.Now()
	expectedProjects := []*domain.Project{
		{
			ID:          "123",
			Name:        "Test Project",
			Description: "A test project",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	mockRepo.On("ListProjects", mock.Anything).Return(expectedProjects, nil)

	handler := handleListProjects(mockRepo)
	args := json.RawMessage(`{}`)

	result, err := handler(context.Background(), args)
	assert.NoError(t, err)

	resStr, err := json.Marshal(result)
	assert.NoError(t, err)

	var res map[string]interface{}
	err = json.Unmarshal(resStr, &res)
	assert.NoError(t, err)

	projects := res["projects"].([]interface{})
	assert.Len(t, projects, 1)
	p1 := projects[0].(map[string]interface{})
	assert.Equal(t, "123", p1["id"])

	mockRepo.AssertExpectations(t)
}
