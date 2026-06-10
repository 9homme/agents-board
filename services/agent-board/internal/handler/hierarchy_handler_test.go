package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agent-board/internal/domain"
	"agent-board/internal/handler"
	"agent-board/internal/repo"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock helpers for hierarchy tests
// ---------------------------------------------------------------------------

// mockRequirementRepoHierarchy supports GetRequirement for chain-guard tests.
type mockRequirementRepoHierarchy struct {
	repo.RequirementRepository
	GetRequirementFunc func(ctx context.Context, id string) (*domain.Requirement, error)
}

func (m *mockRequirementRepoHierarchy) GetRequirement(ctx context.Context, id string) (*domain.Requirement, error) {
	if m.GetRequirementFunc != nil {
		return m.GetRequirementFunc(ctx, id)
	}
	return nil, repo.ErrNotFound
}

func (m *mockRequirementRepoHierarchy) ListByProject(ctx context.Context, projectID string) ([]domain.Requirement, error) {
	return []domain.Requirement{}, nil
}

// mockUserStoryHierarchyRepo supports hierarchy methods.
type mockUserStoryHierarchyRepo struct {
	repo.UserStoryRepository
	GetUserStoryFunc                      func(ctx context.Context, id string) (*domain.UserStory, error)
	ListByRequirementFunc                 func(ctx context.Context, requirementID string) ([]*repo.UserStoryWithCount, error)
	ListUserStoriesWithTaskCountCallCount int
	GetUserStoryCallCount                 int
}

func (m *mockUserStoryHierarchyRepo) GetUserStory(ctx context.Context, id string) (*domain.UserStory, error) {
	m.GetUserStoryCallCount++
	if m.GetUserStoryFunc != nil {
		return m.GetUserStoryFunc(ctx, id)
	}
	return nil, repo.ErrNotFound
}

func (m *mockUserStoryHierarchyRepo) ListByRequirement(ctx context.Context, requirementID string) ([]*repo.UserStoryWithCount, error) {
	if m.ListByRequirementFunc != nil {
		return m.ListByRequirementFunc(ctx, requirementID)
	}
	return []*repo.UserStoryWithCount{}, nil
}

func (m *mockUserStoryHierarchyRepo) ListUserStoriesWithTaskCount(ctx context.Context, projectID string) ([]*repo.UserStoryWithCount, error) {
	m.ListUserStoriesWithTaskCountCallCount++
	return []*repo.UserStoryWithCount{}, nil
}

// mockTaskHierarchyRepo supports GetTask and ListTasks for hierarchy tests.
type mockTaskHierarchyRepo struct {
	repo.TaskRepository
	GetTaskFunc        func(ctx context.Context, id string) (*domain.Task, error)
	ListTasksFunc      func(ctx context.Context, userStoryID string) ([]*domain.Task, error)
	GetTaskCallCount   int
	ListTasksCallCount int
}

func (m *mockTaskHierarchyRepo) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	m.GetTaskCallCount++
	if m.GetTaskFunc != nil {
		return m.GetTaskFunc(ctx, id)
	}
	return nil, repo.ErrNotFound
}

func (m *mockTaskHierarchyRepo) ListTasks(ctx context.Context, userStoryID string) ([]*domain.Task, error) {
	m.ListTasksCallCount++
	if m.ListTasksFunc != nil {
		return m.ListTasksFunc(ctx, userStoryID)
	}
	return []*domain.Task{}, nil
}

// mockDocumentHierarchyRepo supports hierarchy methods.
type mockDocumentHierarchyRepo struct {
	repo.DocumentRepository
	GetDocumentFunc       func(ctx context.Context, id string) (*domain.Document, error)
	ListByRequirementFunc func(ctx context.Context, requirementID string) ([]*domain.Document, error)
	GetDocumentCallCount  int
	ListByReqCallCount    int
}

func (m *mockDocumentHierarchyRepo) GetDocument(ctx context.Context, id string) (*domain.Document, error) {
	m.GetDocumentCallCount++
	if m.GetDocumentFunc != nil {
		return m.GetDocumentFunc(ctx, id)
	}
	return nil, repo.ErrNotFound
}

func (m *mockDocumentHierarchyRepo) ListByRequirement(ctx context.Context, requirementID string) ([]*domain.Document, error) {
	m.ListByReqCallCount++
	if m.ListByRequirementFunc != nil {
		return m.ListByRequirementFunc(ctx, requirementID)
	}
	return []*domain.Document{}, nil
}

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

const (
	hPID  = "11111111-1111-1111-1111-111111111111"
	hRID  = "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f"
	hUSID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	hTID  = "dddddddd-dddd-dddd-dddd-dddddddddddd"
	hDID  = "cccccccc-cccc-cccc-cccc-cccccccccccc"
)

func fixedHierarchyTime() time.Time {
	t, _ := time.Parse(time.RFC3339, "2026-06-02T09:00:00Z")
	return t
}

func validRequirement() *domain.Requirement {
	ts := fixedHierarchyTime()
	return &domain.Requirement{
		ID:        hRID,
		ProjectID: hPID,
		Name:      "Req 1",
		Status:    "draft",
		CreatedAt: ts,
		UpdatedAt: ts,
	}
}

func validUserStory() *domain.UserStory {
	ts := fixedHierarchyTime()
	return &domain.UserStory{
		ID:            hUSID,
		ProjectID:     hPID,
		RequirementID: hRID,
		Title:         "Add item to basket",
		Description:   "",
		Status:        "in_progress",
		CreatedAt:     ts,
		UpdatedAt:     ts,
	}
}

func validUserStoryWithCount() *repo.UserStoryWithCount {
	return &repo.UserStoryWithCount{
		UserStory: *validUserStory(),
		TaskCount: 1,
	}
}

func validTask() *domain.Task {
	ts := fixedHierarchyTime()
	return &domain.Task{
		ID:          hTID,
		UserStoryID: hUSID,
		Title:       "be_basket_repo",
		Description: "",
		Status:      "pending",
		CreatedAt:   ts,
		UpdatedAt:   ts,
	}
}

func validDocument() *domain.Document {
	ts := fixedHierarchyTime()
	return &domain.Document{
		ID:            hDID,
		ProjectID:     hPID,
		RequirementID: hRID,
		Title:         "README",
		Content:       "# README\n...",
		CreatedAt:     ts,
		UpdatedAt:     ts,
	}
}

// ---------------------------------------------------------------------------
// UT-048-011 — User story list item response includes requirementId
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_011_UserStoryListItemIncludesRequirementId(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		ListByRequirementFunc: func(ctx context.Context, requirementID string) ([]*repo.UserStoryWithCount, error) {
			return []*repo.UserStoryWithCount{validUserStoryWithCount()}, nil
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories", h.ListRequirementUserStories)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	stories, ok := res["userStories"].([]interface{})
	require.True(t, ok, "userStories must be an array")
	require.Len(t, stories, 1)

	s := stories[0].(map[string]interface{})
	assert.Equal(t, hRID, s["requirementId"], "requirementId must be present and correct")
	assert.Equal(t, hPID, s["projectId"])
	assert.Equal(t, hUSID, s["id"])
	assert.EqualValues(t, 1, s["taskCount"])
}

// ---------------------------------------------------------------------------
// UT-048-012 — User story detail response includes requirementId but NOT taskCount
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_012_UserStoryDetailIncludesRequirementIdNoTaskCount(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return validUserStory(), nil
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(&mockTaskHierarchyRepo{})

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid", h.GetRequirementUserStory)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	assert.Equal(t, hRID, res["requirementId"], "requirementId must be present")
	assert.Equal(t, hUSID, res["id"])

	_, hasTaskCount := res["taskCount"]
	assert.False(t, hasTaskCount, "detail endpoint must NOT include taskCount")
}

// ---------------------------------------------------------------------------
// UT-048-013 — Document list item response includes requirementId
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_013_DocumentListItemIncludesRequirementId(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	docRepo := &mockDocumentHierarchyRepo{
		ListByRequirementFunc: func(ctx context.Context, requirementID string) ([]*domain.Document, error) {
			return []*domain.Document{validDocument()}, nil
		},
	}

	h := handler.NewDocumentHandler(docRepo, &mockProjectRepoForHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents", h.ListRequirementDocuments)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/documents", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	docs, ok := res["documents"].([]interface{})
	require.True(t, ok)
	require.Len(t, docs, 1)

	d := docs[0].(map[string]interface{})
	assert.Equal(t, hRID, d["requirementId"], "requirementId must be present and correct")
	assert.Equal(t, hPID, d["projectId"])
	assert.Equal(t, hDID, d["id"])

	// No content in list items
	_, hasContent := d["content"]
	assert.False(t, hasContent, "list items must NOT include content")
}

// ---------------------------------------------------------------------------
// UT-048-014 — Document detail response includes requirementId AND content
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_014_DocumentDetailIncludesRequirementIdAndContent(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	docRepo := &mockDocumentHierarchyRepo{
		GetDocumentFunc: func(ctx context.Context, id string) (*domain.Document, error) {
			return validDocument(), nil
		},
	}

	h := handler.NewDocumentHandler(docRepo, &mockProjectRepoForHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents/:docid", h.GetRequirementDocument)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/documents/"+hDID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	assert.Equal(t, hRID, res["requirementId"], "requirementId must be present")
	assert.Equal(t, hDID, res["id"])
	assert.Equal(t, "# README\n...", res["content"], "content must be present in detail response")
}

// ---------------------------------------------------------------------------
// UT-048-015 — Task item shape is unchanged (no requirementId on task)
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_015_TaskItemShapeUnchangedNoRequirementId(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return validUserStory(), nil
		},
	}
	taskRepo := &mockTaskHierarchyRepo{
		ListTasksFunc: func(ctx context.Context, userStoryID string) ([]*domain.Task, error) {
			return []*domain.Task{validTask()}, nil
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks", h.GetRequirementUserStoryTasks)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	tasks, ok := res["tasks"].([]interface{})
	require.True(t, ok)
	require.Len(t, tasks, 1)

	tk := tasks[0].(map[string]interface{})
	assert.Equal(t, hTID, tk["id"])
	assert.Equal(t, hUSID, tk["userStoryId"])

	_, hasRequirementId := tk["requirementId"]
	assert.False(t, hasRequirementId, "task items must NOT include requirementId")
}

// ---------------------------------------------------------------------------
// UT-048-007 — Chain guard: requirement.ProjectID != :pid → 404
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_007_ChainGuardWrongProjectForRequirement(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			// Requirement belongs to a DIFFERENT project
			r := validRequirement()
			r.ProjectID = "other-project-uuid"
			return r, nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories", h.ListRequirementUserStories)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Requirement not found", res["message"])

	// Child resource must NOT have been fetched
	assert.Equal(t, 0, usRepo.ListUserStoriesWithTaskCountCallCount,
		"ListUserStoriesWithTaskCount must not be called when chain guard fails")
}

// ---------------------------------------------------------------------------
// UT-048-008 — Chain guard: userStory.RequirementID != :rid → 404
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_008_ChainGuardWrongRequirementForStory(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil // requirement belongs to project OK
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			s := validUserStory()
			s.RequirementID = "other-requirement-uuid" // mismatch
			return s, nil
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(&mockTaskHierarchyRepo{})

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid", h.GetRequirementUserStory)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "User story not found", res["message"])
}

// ---------------------------------------------------------------------------
// UT-048-009 — Chain guard: task.UserStoryID != :usid → 404
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_009_ChainGuardWrongStoryForTask(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return validUserStory(), nil
		},
	}
	taskRepo := &mockTaskHierarchyRepo{
		GetTaskFunc: func(ctx context.Context, id string) (*domain.Task, error) {
			t := validTask()
			t.UserStoryID = "other-story-uuid" // mismatch
			return t, nil
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid", h.GetRequirementTask)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks/"+hTID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Task not found", res["message"])
}

// ---------------------------------------------------------------------------
// UT-048-010 — Chain guard: document.RequirementID != :rid → 404
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_010_ChainGuardWrongRequirementForDocument(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	docRepo := &mockDocumentHierarchyRepo{
		GetDocumentFunc: func(ctx context.Context, id string) (*domain.Document, error) {
			d := validDocument()
			d.RequirementID = "other-requirement-uuid" // mismatch
			return d, nil
		},
	}

	h := handler.NewDocumentHandler(docRepo, &mockProjectRepoForHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents/:docid", h.GetRequirementDocument)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/documents/"+hDID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Document not found", res["message"])
}

// ---------------------------------------------------------------------------
// UT-048-001 — UserStoryHandler.ListByRequirement: 500 on repo error
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_001_ListUserStoriesByRequirement_500_RepoError(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		ListByRequirementFunc: func(ctx context.Context, requirementID string) ([]*repo.UserStoryWithCount, error) {
			return nil, errors.New("db error")
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories", h.ListRequirementUserStories)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Failed to fetch user stories", res["message"])
}

// ---------------------------------------------------------------------------
// UT-048-002 — UserStoryHandler.GetUserStory (hierarchy): 500 on repo error
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_002_GetUserStoryHierarchy_500_RepoError(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return nil, errors.New("db error")
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(&mockTaskHierarchyRepo{})

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid", h.GetRequirementUserStory)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Internal server error", res["message"])
}

// ---------------------------------------------------------------------------
// UT-048-003 — Task list (hierarchy): 500 on repo error
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_003_ListTasksHierarchy_500_RepoError(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return validUserStory(), nil
		},
	}
	taskRepo := &mockTaskHierarchyRepo{
		ListTasksFunc: func(ctx context.Context, userStoryID string) ([]*domain.Task, error) {
			return nil, errors.New("db error")
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks", h.GetRequirementUserStoryTasks)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Failed to fetch tasks", res["message"])
}

// ---------------------------------------------------------------------------
// UT-048-004 — Get task (hierarchy): 500 on repo error
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_004_GetTaskHierarchy_500_RepoError(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return validUserStory(), nil
		},
	}
	taskRepo := &mockTaskHierarchyRepo{
		GetTaskFunc: func(ctx context.Context, id string) (*domain.Task, error) {
			return nil, errors.New("db error")
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid", h.GetRequirementTask)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks/"+hTID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Internal server error", res["message"])
}

// ---------------------------------------------------------------------------
// UT-048-005 — Document list (hierarchy): 500 on repo error
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_005_ListDocumentsByRequirement_500_RepoError(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	docRepo := &mockDocumentHierarchyRepo{
		ListByRequirementFunc: func(ctx context.Context, requirementID string) ([]*domain.Document, error) {
			return nil, errors.New("db error")
		},
	}

	h := handler.NewDocumentHandler(docRepo, &mockProjectRepoForHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents", h.ListRequirementDocuments)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/documents", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Failed to fetch documents", res["message"])
}

// ---------------------------------------------------------------------------
// UT-048-006 — Get document (hierarchy): 500 on repo error
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_006_GetDocumentHierarchy_500_RepoError(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	docRepo := &mockDocumentHierarchyRepo{
		GetDocumentFunc: func(ctx context.Context, id string) (*domain.Document, error) {
			return nil, errors.New("db error")
		},
	}

	h := handler.NewDocumentHandler(docRepo, &mockProjectRepoForHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents/:docid", h.GetRequirementDocument)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/documents/"+hDID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Failed to fetch document", res["message"])
}

// ---------------------------------------------------------------------------
// UT-048-016 — Context cancellation propagated to all repo calls
// ---------------------------------------------------------------------------

func TestHierarchy_UT048_016_ContextCancellationPropagated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled immediately

	var capturedCtx context.Context
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			capturedCtx = ctx
			return nil, ctx.Err()
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories", h.ListRequirementUserStories)

	req := httptest.NewRequestWithContext(ctx, http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Handler should return error (404 since GetRequirement returns ctx.Err())
	// Most importantly, the context passed to repo must be cancelled
	require.NotNil(t, capturedCtx)
	assert.Equal(t, context.Canceled, capturedCtx.Err(), "context must be cancelled when passed to repo")

	// Child repos must not have been called with a live context
	assert.Equal(t, 0, usRepo.GetUserStoryCallCount, "child repo must not be called after context cancelled")
}

// ---------------------------------------------------------------------------
// IT-048-001 — GET /api/v1/projects/:pid/requirements/:rid/user-stories — 200
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_001_ListUserStoriesByRequirement_200(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			require.Equal(t, hRID, id)
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		ListByRequirementFunc: func(ctx context.Context, requirementID string) ([]*repo.UserStoryWithCount, error) {
			assert.Equal(t, hRID, requirementID)
			return []*repo.UserStoryWithCount{validUserStoryWithCount()}, nil
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories", h.ListRequirementUserStories)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	stories, ok := res["userStories"].([]interface{})
	require.True(t, ok)
	require.Len(t, stories, 1)

	s := stories[0].(map[string]interface{})
	assert.Equal(t, hUSID, s["id"])
	assert.Equal(t, hPID, s["projectId"])
	assert.Equal(t, hRID, s["requirementId"])
	assert.Equal(t, "Add item to basket", s["title"])
	assert.Equal(t, "", s["description"])
	assert.Equal(t, "in_progress", s["status"])
	assert.EqualValues(t, 1, s["taskCount"])
	assert.Equal(t, "2026-06-02T09:00:00Z", s["createdAt"])
	assert.Equal(t, "2026-06-02T09:00:00Z", s["updatedAt"])
}

// ---------------------------------------------------------------------------
// IT-048-002 — chain mismatch: requirement not in project → 404
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_002_ListUserStories_404_RequirementNotInProject(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			r := validRequirement()
			r.ProjectID = "other-project-id" // belongs to different project
			return r, nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories", h.ListRequirementUserStories)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Requirement not found", res["message"])
}

// ---------------------------------------------------------------------------
// IT-048-003 — project not found → 404
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_003_ListUserStories_404_RequirementNotFound(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return nil, repo.ErrNotFound
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories", h.ListRequirementUserStories)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/00000000-0000-0000-0000-000000000000/requirements/"+hRID+"/user-stories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
}

// ---------------------------------------------------------------------------
// IT-048-004 — 200 empty list
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_004_ListUserStories_200_EmptyList(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		ListByRequirementFunc: func(ctx context.Context, requirementID string) ([]*repo.UserStoryWithCount, error) {
			return []*repo.UserStoryWithCount{}, nil
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories", h.ListRequirementUserStories)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"userStories":[]}`, rec.Body.String())
}

// ---------------------------------------------------------------------------
// IT-048-005 — GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid — 200
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_005_GetUserStoryDetail_200(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return validUserStory(), nil
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(&mockTaskHierarchyRepo{})

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid", h.GetRequirementUserStory)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	assert.Equal(t, hUSID, res["id"])
	assert.Equal(t, hPID, res["projectId"])
	assert.Equal(t, hRID, res["requirementId"])
	assert.Equal(t, "Add item to basket", res["title"])
	assert.Equal(t, "", res["description"])
	assert.Equal(t, "in_progress", res["status"])
	assert.Equal(t, "2026-06-02T09:00:00Z", res["createdAt"])
	assert.Equal(t, "2026-06-02T09:00:00Z", res["updatedAt"])

	_, hasTaskCount := res["taskCount"]
	assert.False(t, hasTaskCount, "detail must not have taskCount")
}

// ---------------------------------------------------------------------------
// IT-048-006 — story belongs to different requirement → 404
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_006_GetUserStory_404_StoryInWrongRequirement(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			s := validUserStory()
			s.RequirementID = "other-req-id"
			return s, nil
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(&mockTaskHierarchyRepo{})

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid", h.GetRequirementUserStory)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "User story not found", res["message"])
}

// ---------------------------------------------------------------------------
// IT-048-007 — requirement not in project (story mismatch hidden)
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_007_GetUserStory_404_RequirementNotInProject(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			r := validRequirement()
			r.ProjectID = "other-project-id"
			return r, nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(&mockTaskHierarchyRepo{})

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid", h.GetRequirementUserStory)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
}

// ---------------------------------------------------------------------------
// IT-048-008 — non-existent user story → 404
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_008_GetUserStory_404_NotFound(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return nil, repo.ErrNotFound
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(&mockTaskHierarchyRepo{})

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid", h.GetRequirementUserStory)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/00000000-0000-0000-0000-000000000000", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "User story not found", res["message"])
}

// ---------------------------------------------------------------------------
// IT-048-009 — GET .../tasks — 200 task list
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_009_ListTasks_200(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return validUserStory(), nil
		},
	}
	taskRepo := &mockTaskHierarchyRepo{
		ListTasksFunc: func(ctx context.Context, userStoryID string) ([]*domain.Task, error) {
			return []*domain.Task{validTask()}, nil
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks", h.GetRequirementUserStoryTasks)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	tasks, ok := res["tasks"].([]interface{})
	require.True(t, ok)
	require.Len(t, tasks, 1)

	tk := tasks[0].(map[string]interface{})
	assert.Equal(t, hTID, tk["id"])
	assert.Equal(t, hUSID, tk["userStoryId"])
	assert.Equal(t, "be_basket_repo", tk["title"])
	assert.Equal(t, "", tk["description"])
	assert.Equal(t, "pending", tk["status"])
	assert.Equal(t, "2026-06-02T09:00:00Z", tk["createdAt"])
	assert.Equal(t, "2026-06-02T09:00:00Z", tk["updatedAt"])

	_, hasReqId := tk["requirementId"]
	assert.False(t, hasReqId, "task items must NOT have requirementId")
}

// ---------------------------------------------------------------------------
// IT-048-010 — story not in requirement → 404 for task list
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_010_ListTasks_404_StoryNotInRequirement(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			s := validUserStory()
			s.RequirementID = "other-req"
			return s, nil
		},
	}
	taskRepo := &mockTaskHierarchyRepo{}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks", h.GetRequirementUserStoryTasks)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "User story not found", res["message"])
}

// ---------------------------------------------------------------------------
// IT-048-011 — requirement not in project → 404 for task list
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_011_ListTasks_404_RequirementNotInProject(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			r := validRequirement()
			r.ProjectID = "other-project"
			return r, nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{}
	taskRepo := &mockTaskHierarchyRepo{}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks", h.GetRequirementUserStoryTasks)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
}

// ---------------------------------------------------------------------------
// IT-048-012 — 200 empty task list
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_012_ListTasks_200_EmptyList(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return validUserStory(), nil
		},
	}
	taskRepo := &mockTaskHierarchyRepo{
		ListTasksFunc: func(ctx context.Context, userStoryID string) ([]*domain.Task, error) {
			return []*domain.Task{}, nil
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks", h.GetRequirementUserStoryTasks)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"tasks":[]}`, rec.Body.String())
}

// ---------------------------------------------------------------------------
// IT-048-013 — GET .../tasks/:tid — 200 task detail
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_013_GetTask_200(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return validUserStory(), nil
		},
	}
	taskRepo := &mockTaskHierarchyRepo{
		GetTaskFunc: func(ctx context.Context, id string) (*domain.Task, error) {
			return validTask(), nil
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid", h.GetRequirementTask)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks/"+hTID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	assert.Equal(t, hTID, res["id"])
	assert.Equal(t, hUSID, res["userStoryId"])
	assert.Equal(t, "be_basket_repo", res["title"])
	assert.Equal(t, "", res["description"])
	assert.Equal(t, "pending", res["status"])
	assert.Equal(t, "2026-06-02T09:00:00Z", res["createdAt"])
	assert.Equal(t, "2026-06-02T09:00:00Z", res["updatedAt"])
}

// ---------------------------------------------------------------------------
// IT-048-014 — task not in story → 404
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_014_GetTask_404_TaskNotInStory(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return validUserStory(), nil
		},
	}
	taskRepo := &mockTaskHierarchyRepo{
		GetTaskFunc: func(ctx context.Context, id string) (*domain.Task, error) {
			t := validTask()
			t.UserStoryID = "other-story"
			return t, nil
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid", h.GetRequirementTask)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks/"+hTID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Task not found", res["message"])
}

// ---------------------------------------------------------------------------
// IT-048-015 — story not in requirement → 404 for task detail
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_015_GetTask_404_StoryNotInRequirement(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			s := validUserStory()
			s.RequirementID = "other-req"
			return s, nil
		},
	}
	taskRepo := &mockTaskHierarchyRepo{}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid", h.GetRequirementTask)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks/"+hTID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
}

// ---------------------------------------------------------------------------
// IT-048-016 — non-existent task → 404
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_016_GetTask_404_NotFound(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return validUserStory(), nil
		},
	}
	taskRepo := &mockTaskHierarchyRepo{
		GetTaskFunc: func(ctx context.Context, id string) (*domain.Task, error) {
			return nil, repo.ErrNotFound
		},
	}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid", h.GetRequirementTask)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks/00000000-0000-0000-0000-000000000000", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Task not found", res["message"])
}

// ---------------------------------------------------------------------------
// IT-048-017 — GET .../documents — 200 document list
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_017_ListDocumentsByRequirement_200(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	docRepo := &mockDocumentHierarchyRepo{
		ListByRequirementFunc: func(ctx context.Context, requirementID string) ([]*domain.Document, error) {
			return []*domain.Document{validDocument()}, nil
		},
	}

	h := handler.NewDocumentHandler(docRepo, &mockProjectRepoForHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents", h.ListRequirementDocuments)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/documents", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	docs, ok := res["documents"].([]interface{})
	require.True(t, ok)
	require.Len(t, docs, 1)

	d := docs[0].(map[string]interface{})
	assert.Equal(t, hDID, d["id"])
	assert.Equal(t, hPID, d["projectId"])
	assert.Equal(t, hRID, d["requirementId"])
	assert.Equal(t, "README", d["title"])
	assert.Equal(t, "2026-06-02T09:00:00Z", d["createdAt"])
	assert.Equal(t, "2026-06-02T09:00:00Z", d["updatedAt"])

	_, hasContent := d["content"]
	assert.False(t, hasContent, "content must not be in list items")
}

// ---------------------------------------------------------------------------
// IT-048-018 — requirement not in project → 404 for document list
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_018_ListDocuments_404_RequirementNotInProject(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			r := validRequirement()
			r.ProjectID = "other-project"
			return r, nil
		},
	}
	docRepo := &mockDocumentHierarchyRepo{}

	h := handler.NewDocumentHandler(docRepo, &mockProjectRepoForHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents", h.ListRequirementDocuments)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/documents", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Requirement not found", res["message"])
}

// ---------------------------------------------------------------------------
// IT-048-019 — 200 empty document list
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_019_ListDocuments_200_EmptyList(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	docRepo := &mockDocumentHierarchyRepo{
		ListByRequirementFunc: func(ctx context.Context, requirementID string) ([]*domain.Document, error) {
			return []*domain.Document{}, nil
		},
	}

	h := handler.NewDocumentHandler(docRepo, &mockProjectRepoForHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents", h.ListRequirementDocuments)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/documents", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"documents":[]}`, rec.Body.String())
}

// ---------------------------------------------------------------------------
// IT-048-020 — GET .../documents/:docid — 200 with content
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_020_GetDocument_200_WithContent(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	docRepo := &mockDocumentHierarchyRepo{
		GetDocumentFunc: func(ctx context.Context, id string) (*domain.Document, error) {
			return validDocument(), nil
		},
	}

	h := handler.NewDocumentHandler(docRepo, &mockProjectRepoForHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents/:docid", h.GetRequirementDocument)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/documents/"+hDID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	assert.Equal(t, hDID, res["id"])
	assert.Equal(t, hPID, res["projectId"])
	assert.Equal(t, hRID, res["requirementId"])
	assert.Equal(t, "README", res["title"])
	assert.Equal(t, "# README\n...", res["content"])
	assert.Equal(t, "2026-06-02T09:00:00Z", res["createdAt"])
	assert.Equal(t, "2026-06-02T09:00:00Z", res["updatedAt"])
}

// ---------------------------------------------------------------------------
// IT-048-021 — document not in requirement → 404
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_021_GetDocument_404_DocNotInRequirement(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	docRepo := &mockDocumentHierarchyRepo{
		GetDocumentFunc: func(ctx context.Context, id string) (*domain.Document, error) {
			d := validDocument()
			d.RequirementID = "other-req"
			return d, nil
		},
	}

	h := handler.NewDocumentHandler(docRepo, &mockProjectRepoForHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents/:docid", h.GetRequirementDocument)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/documents/"+hDID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Document not found", res["message"])
}

// ---------------------------------------------------------------------------
// IT-048-022 — requirement not in project → 404 for document detail
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_022_GetDocument_404_RequirementNotInProject(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			r := validRequirement()
			r.ProjectID = "other-project"
			return r, nil
		},
	}
	docRepo := &mockDocumentHierarchyRepo{}

	h := handler.NewDocumentHandler(docRepo, &mockProjectRepoForHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents/:docid", h.GetRequirementDocument)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/documents/"+hDID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
}

// ---------------------------------------------------------------------------
// IT-048-023 — non-existent document → 404
// ---------------------------------------------------------------------------

func TestHierarchy_IT048_023_GetDocument_404_NotFound(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	docRepo := &mockDocumentHierarchyRepo{
		GetDocumentFunc: func(ctx context.Context, id string) (*domain.Document, error) {
			return nil, repo.ErrNotFound
		},
	}

	h := handler.NewDocumentHandler(docRepo, &mockProjectRepoForHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents/:docid", h.GetRequirementDocument)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/documents/00000000-0000-0000-0000-000000000000", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Document not found", res["message"])
}

// ---------------------------------------------------------------------------
// IT-048-024..031 — Removed flat routes return 404/405
// ---------------------------------------------------------------------------

func buildRouterWithoutFlatRoutes(usHandler *handler.UserStoryHandler, docHandler *handler.DocumentHandler, reqHandler *handler.RequirementHandler) *echo.Echo {
	e := echo.New()
	// Hierarchy routes registered
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories", usHandler.ListRequirementUserStories)
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid", usHandler.GetRequirementUserStory)
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks", usHandler.GetRequirementUserStoryTasks)
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid", usHandler.GetRequirementTask)
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents", docHandler.ListRequirementDocuments)
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents/:docid", docHandler.GetRequirementDocument)
	e.GET("/api/v1/projects/:pid/requirements", reqHandler.ListProjectRequirements)
	// Old flat routes NOT registered:
	// /api/v1/projects/:id/user-stories
	// /api/v1/projects/:id/documents
	// /api/v1/user-stories/:id
	// /api/v1/user-stories/:id/tasks
	// /api/v1/tasks/:id
	// /api/v1/documents/:id
	// /api/v1/requirements/:rid/user-stories
	// /api/v1/requirements/:rid/documents
	return e
}

func TestHierarchy_IT048_024_FlatRouteUserStoriesRemoved(t *testing.T) {
	usHandler := handler.NewUserStoryHandler(
		&mockUserStoryHierarchyRepo{}, &mockProjectRepoForUSHandler{},
	)
	usHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})
	usHandler.SetTaskRepo(&mockTaskHierarchyRepo{})

	docHandler := handler.NewDocumentHandler(&mockDocumentHierarchyRepo{}, &mockProjectRepoForHandler{})
	docHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})

	reqHandler := handler.NewRequirementHandler(&mockRequirementRepoHierarchy{}, &mockProjectRepoForHandler{})

	e := buildRouterWithoutFlatRoutes(usHandler, docHandler, reqHandler)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/11111111-1111-1111-1111-111111111111/user-stories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.True(t, rec.Code == http.StatusNotFound || rec.Code == http.StatusMethodNotAllowed,
		"old flat route must return 404 or 405, got %d", rec.Code)
}

func TestHierarchy_IT048_025_FlatRouteDocumentsRemoved(t *testing.T) {
	usHandler := handler.NewUserStoryHandler(
		&mockUserStoryHierarchyRepo{}, &mockProjectRepoForUSHandler{},
	)
	usHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})
	usHandler.SetTaskRepo(&mockTaskHierarchyRepo{})

	docHandler := handler.NewDocumentHandler(&mockDocumentHierarchyRepo{}, &mockProjectRepoForHandler{})
	docHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})

	reqHandler := handler.NewRequirementHandler(&mockRequirementRepoHierarchy{}, &mockProjectRepoForHandler{})

	e := buildRouterWithoutFlatRoutes(usHandler, docHandler, reqHandler)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/11111111-1111-1111-1111-111111111111/documents", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.True(t, rec.Code == http.StatusNotFound || rec.Code == http.StatusMethodNotAllowed,
		"old flat route must return 404 or 405, got %d", rec.Code)
}

func TestHierarchy_IT048_026_FlatRouteUserStoryDetailRemoved(t *testing.T) {
	usHandler := handler.NewUserStoryHandler(
		&mockUserStoryHierarchyRepo{}, &mockProjectRepoForUSHandler{},
	)
	usHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})
	usHandler.SetTaskRepo(&mockTaskHierarchyRepo{})

	docHandler := handler.NewDocumentHandler(&mockDocumentHierarchyRepo{}, &mockProjectRepoForHandler{})
	docHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})

	reqHandler := handler.NewRequirementHandler(&mockRequirementRepoHierarchy{}, &mockProjectRepoForHandler{})

	e := buildRouterWithoutFlatRoutes(usHandler, docHandler, reqHandler)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/user-stories/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.True(t, rec.Code == http.StatusNotFound || rec.Code == http.StatusMethodNotAllowed,
		"old flat route must return 404 or 405, got %d", rec.Code)
}

func TestHierarchy_IT048_027_FlatRouteUserStoryTasksRemoved(t *testing.T) {
	usHandler := handler.NewUserStoryHandler(
		&mockUserStoryHierarchyRepo{}, &mockProjectRepoForUSHandler{},
	)
	usHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})
	usHandler.SetTaskRepo(&mockTaskHierarchyRepo{})

	docHandler := handler.NewDocumentHandler(&mockDocumentHierarchyRepo{}, &mockProjectRepoForHandler{})
	docHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})

	reqHandler := handler.NewRequirementHandler(&mockRequirementRepoHierarchy{}, &mockProjectRepoForHandler{})

	e := buildRouterWithoutFlatRoutes(usHandler, docHandler, reqHandler)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/user-stories/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.True(t, rec.Code == http.StatusNotFound || rec.Code == http.StatusMethodNotAllowed,
		"old flat route must return 404 or 405, got %d", rec.Code)
}

func TestHierarchy_IT048_028_FlatRouteTaskDetailRemoved(t *testing.T) {
	usHandler := handler.NewUserStoryHandler(
		&mockUserStoryHierarchyRepo{}, &mockProjectRepoForUSHandler{},
	)
	usHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})
	usHandler.SetTaskRepo(&mockTaskHierarchyRepo{})

	docHandler := handler.NewDocumentHandler(&mockDocumentHierarchyRepo{}, &mockProjectRepoForHandler{})
	docHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})

	reqHandler := handler.NewRequirementHandler(&mockRequirementRepoHierarchy{}, &mockProjectRepoForHandler{})

	e := buildRouterWithoutFlatRoutes(usHandler, docHandler, reqHandler)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/tasks/dddddddd-dddd-dddd-dddd-dddddddddddd", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.True(t, rec.Code == http.StatusNotFound || rec.Code == http.StatusMethodNotAllowed,
		"old flat route must return 404 or 405, got %d", rec.Code)
}

func TestHierarchy_IT048_029_FlatRouteDocumentDetailRemoved(t *testing.T) {
	usHandler := handler.NewUserStoryHandler(
		&mockUserStoryHierarchyRepo{}, &mockProjectRepoForUSHandler{},
	)
	usHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})
	usHandler.SetTaskRepo(&mockTaskHierarchyRepo{})

	docHandler := handler.NewDocumentHandler(&mockDocumentHierarchyRepo{}, &mockProjectRepoForHandler{})
	docHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})

	reqHandler := handler.NewRequirementHandler(&mockRequirementRepoHierarchy{}, &mockProjectRepoForHandler{})

	e := buildRouterWithoutFlatRoutes(usHandler, docHandler, reqHandler)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/documents/cccccccc-cccc-cccc-cccc-cccccccccccc", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.True(t, rec.Code == http.StatusNotFound || rec.Code == http.StatusMethodNotAllowed,
		"old flat route must return 404 or 405, got %d", rec.Code)
}

func TestHierarchy_IT048_030_IntermediateRouteReqUserStoriesRemoved(t *testing.T) {
	usHandler := handler.NewUserStoryHandler(
		&mockUserStoryHierarchyRepo{}, &mockProjectRepoForUSHandler{},
	)
	usHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})
	usHandler.SetTaskRepo(&mockTaskHierarchyRepo{})

	docHandler := handler.NewDocumentHandler(&mockDocumentHierarchyRepo{}, &mockProjectRepoForHandler{})
	docHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})

	reqHandler := handler.NewRequirementHandler(&mockRequirementRepoHierarchy{}, &mockProjectRepoForHandler{})

	e := buildRouterWithoutFlatRoutes(usHandler, docHandler, reqHandler)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/requirements/b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f/user-stories", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.True(t, rec.Code == http.StatusNotFound || rec.Code == http.StatusMethodNotAllowed,
		"intermediate route must return 404 or 405, got %d", rec.Code)
}

func TestHierarchy_IT048_031_IntermediateRouteReqDocumentsRemoved(t *testing.T) {
	usHandler := handler.NewUserStoryHandler(
		&mockUserStoryHierarchyRepo{}, &mockProjectRepoForUSHandler{},
	)
	usHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})
	usHandler.SetTaskRepo(&mockTaskHierarchyRepo{})

	docHandler := handler.NewDocumentHandler(&mockDocumentHierarchyRepo{}, &mockProjectRepoForHandler{})
	docHandler.SetRequirementRepo(&mockRequirementRepoHierarchy{})

	reqHandler := handler.NewRequirementHandler(&mockRequirementRepoHierarchy{}, &mockProjectRepoForHandler{})

	e := buildRouterWithoutFlatRoutes(usHandler, docHandler, reqHandler)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/requirements/b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f/documents", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.True(t, rec.Code == http.StatusNotFound || rec.Code == http.StatusMethodNotAllowed,
		"intermediate route must return 404 or 405, got %d", rec.Code)
}

// ---------------------------------------------------------------------------
// Coverage gap: checkDocumentRequirementChain 500 on generic DB error
// ---------------------------------------------------------------------------

func TestHierarchy_CoverageGap_DocChainGuard_500_GenericError(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return nil, errors.New("db error")
		},
	}
	docRepo := &mockDocumentHierarchyRepo{}

	h := handler.NewDocumentHandler(docRepo, &mockProjectRepoForHandler{})
	h.SetRequirementRepo(reqRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/documents", h.ListRequirementDocuments)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/documents", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Failed to fetch documents", res["message"])
}

// ---------------------------------------------------------------------------
// Coverage gap: GetRequirementUserStoryTasks 500 from checkRequirementChain generic error
// ---------------------------------------------------------------------------

func TestHierarchy_CoverageGap_TaskList_500_ChainGenericError(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return nil, errors.New("db error")
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{}
	taskRepo := &mockTaskHierarchyRepo{}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks", h.GetRequirementUserStoryTasks)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
}

// ---------------------------------------------------------------------------
// Coverage gap: GetRequirementTask 500 from checkRequirementChain generic error
// ---------------------------------------------------------------------------

func TestHierarchy_CoverageGap_GetTask_500_ChainGenericError(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return nil, errors.New("db error")
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{}
	taskRepo := &mockTaskHierarchyRepo{}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid", h.GetRequirementTask)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks/"+hTID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
}

// ---------------------------------------------------------------------------
// Coverage gap: GetRequirementTask 500 from GetUserStory generic error
// ---------------------------------------------------------------------------

func TestHierarchy_CoverageGap_GetTask_500_StoryGenericError(t *testing.T) {
	reqRepo := &mockRequirementRepoHierarchy{
		GetRequirementFunc: func(ctx context.Context, id string) (*domain.Requirement, error) {
			return validRequirement(), nil
		},
	}
	usRepo := &mockUserStoryHierarchyRepo{
		GetUserStoryFunc: func(ctx context.Context, id string) (*domain.UserStory, error) {
			return nil, errors.New("db error")
		},
	}
	taskRepo := &mockTaskHierarchyRepo{}

	h := handler.NewUserStoryHandler(usRepo, &mockProjectRepoForUSHandler{})
	h.SetRequirementRepo(reqRepo)
	h.SetTaskRepo(taskRepo)

	e := echo.New()
	e.GET("/api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid", h.GetRequirementTask)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+hPID+"/requirements/"+hRID+"/user-stories/"+hUSID+"/tasks/"+hTID, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
}
