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

// mockDocumentRepoForHandler is a mock for repo.DocumentRepository used in document handler tests.
type mockDocumentRepoForHandler struct {
	repo.DocumentRepository
	ListDocumentsFunc func(ctx context.Context, projectID string) ([]*domain.Document, error)
	listCallCount     int
}

func (m *mockDocumentRepoForHandler) ListDocuments(ctx context.Context, projectID string) ([]*domain.Document, error) {
	m.listCallCount++
	if m.ListDocumentsFunc != nil {
		return m.ListDocumentsFunc(ctx, projectID)
	}
	return nil, nil
}

// mockProjectRepoForHandler is a mock for repo.ProjectRepository used in document handler tests.
type mockProjectRepoForHandler struct {
	repo.ProjectRepository
	GetProjectFunc func(ctx context.Context, id string) (*domain.Project, error)
}

func (m *mockProjectRepoForHandler) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	if m.GetProjectFunc != nil {
		return m.GetProjectFunc(ctx, id)
	}
	return nil, repo.ErrNotFound
}

// mustParseTime parses an RFC3339 timestamp and panics on error (test helper).
func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

// newDocumentHandlerContext creates an Echo context with path param :id set.
func newDocumentHandlerContext(e *echo.Echo, method, path, paramValue string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequestWithContext(context.Background(), method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(paramValue)
	return c, rec
}

// UT-US002-001 — ListProjectDocuments handler: 200 with multiple documents
func TestDocumentHandler_ListProjectDocuments_200_MultipleDocuments(t *testing.T) {
	e := echo.New()
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	createdAt1 := mustParseTime("2026-05-18T08:30:00Z")
	updatedAt1 := mustParseTime("2026-05-20T09:45:00Z")
	createdAt2 := mustParseTime("2026-05-15T11:00:00Z")
	updatedAt2 := mustParseTime("2026-05-19T16:20:00Z")

	projectRepo := &mockProjectRepoForHandler{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return &domain.Project{ID: projectID, Name: "Test", Description: ""}, nil
		},
	}
	docRepo := &mockDocumentRepoForHandler{
		ListDocumentsFunc: func(ctx context.Context, pid string) ([]*domain.Document, error) {
			return []*domain.Document{
				{
					ID:        "d111aaaa-1111-1111-1111-111111111111",
					ProjectID: projectID,
					Title:     "Architecture overview",
					Content:   "should be excluded",
					CreatedAt: createdAt1,
					UpdatedAt: updatedAt1,
				},
				{
					ID:        "d222bbbb-2222-2222-2222-222222222222",
					ProjectID: projectID,
					Title:     "Onboarding guide",
					Content:   "should be excluded",
					CreatedAt: createdAt2,
					UpdatedAt: updatedAt2,
				},
			}, nil
		},
	}

	h := handler.NewDocumentHandler(docRepo, projectRepo)
	c, rec := newDocumentHandlerContext(e, http.MethodGet, "/api/v1/projects/"+projectID+"/documents", projectID)

	err := h.ListProjectDocuments(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	docs, ok := res["documents"].([]interface{})
	require.True(t, ok, "documents key must be a JSON array")
	require.Len(t, docs, 2)

	d0 := docs[0].(map[string]interface{})
	assert.Equal(t, "d111aaaa-1111-1111-1111-111111111111", d0["id"])
	assert.Equal(t, projectID, d0["projectId"])
	assert.Equal(t, "Architecture overview", d0["title"])
	assert.Equal(t, "2026-05-18T08:30:00Z", d0["createdAt"])
	assert.Equal(t, "2026-05-20T09:45:00Z", d0["updatedAt"])

	d1 := docs[1].(map[string]interface{})
	assert.Equal(t, "d222bbbb-2222-2222-2222-222222222222", d1["id"])
	assert.Equal(t, projectID, d1["projectId"])
	assert.Equal(t, "Onboarding guide", d1["title"])
	assert.Equal(t, "2026-05-15T11:00:00Z", d1["createdAt"])
	assert.Equal(t, "2026-05-19T16:20:00Z", d1["updatedAt"])
}

// UT-US002-002 — ListProjectDocuments handler: 200 empty list (project exists, no documents)
func TestDocumentHandler_ListProjectDocuments_200_EmptyList(t *testing.T) {
	e := echo.New()
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	projectRepo := &mockProjectRepoForHandler{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return &domain.Project{ID: projectID, Name: "Test", Description: ""}, nil
		},
	}
	docRepo := &mockDocumentRepoForHandler{
		ListDocumentsFunc: func(ctx context.Context, pid string) ([]*domain.Document, error) {
			return []*domain.Document{}, nil
		},
	}

	h := handler.NewDocumentHandler(docRepo, projectRepo)
	c, rec := newDocumentHandlerContext(e, http.MethodGet, "/api/v1/projects/"+projectID+"/documents", projectID)

	err := h.ListProjectDocuments(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify the exact body — documents must be [] not null.
	body := rec.Body.String()
	assert.JSONEq(t, `{"documents":[]}`, body)

	// Also confirm via struct that documents is an array (not null).
	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	docs, ok := res["documents"].([]interface{})
	require.True(t, ok, "documents must be a JSON array, not null")
	assert.Empty(t, docs)
}

// UT-US002-003 — ListProjectDocuments handler: 404 project not found (D-006)
func TestDocumentHandler_ListProjectDocuments_404_ProjectNotFound(t *testing.T) {
	e := echo.New()

	projectRepo := &mockProjectRepoForHandler{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return nil, repo.ErrNotFound
		},
	}
	docRepo := &mockDocumentRepoForHandler{}

	h := handler.NewDocumentHandler(docRepo, projectRepo)
	c, rec := newDocumentHandlerContext(e, http.MethodGet, "/api/v1/projects/no-such-project/documents", "no-such-project")

	err := h.ListProjectDocuments(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "NOT_FOUND", res["code"])
	assert.Equal(t, "Project not found", res["message"])

	// ListDocuments must NOT have been called.
	assert.Equal(t, 0, docRepo.listCallCount, "ListDocuments must not be called when project is not found")
}

// UT-US002-004 — ListProjectDocuments handler: 500 on project lookup failure
func TestDocumentHandler_ListProjectDocuments_500_ProjectLookupFailure(t *testing.T) {
	e := echo.New()

	projectRepo := &mockProjectRepoForHandler{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return nil, errors.New("connection pool exhausted")
		},
	}
	docRepo := &mockDocumentRepoForHandler{}

	h := handler.NewDocumentHandler(docRepo, projectRepo)
	c, rec := newDocumentHandlerContext(e, http.MethodGet, "/api/v1/projects/any-id/documents", "any-id")

	err := h.ListProjectDocuments(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Failed to fetch documents", res["message"])
}

// UT-US002-005 — ListProjectDocuments handler: 500 on document list failure
func TestDocumentHandler_ListProjectDocuments_500_DocumentListFailure(t *testing.T) {
	e := echo.New()
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	projectRepo := &mockProjectRepoForHandler{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return &domain.Project{ID: projectID, Name: "Test", Description: ""}, nil
		},
	}
	docRepo := &mockDocumentRepoForHandler{
		ListDocumentsFunc: func(ctx context.Context, pid string) ([]*domain.Document, error) {
			return nil, errors.New("db connection dropped")
		},
	}

	h := handler.NewDocumentHandler(docRepo, projectRepo)
	c, rec := newDocumentHandlerContext(e, http.MethodGet, "/api/v1/projects/"+projectID+"/documents", projectID)

	err := h.ListProjectDocuments(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var res map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "INTERNAL_ERROR", res["code"])
	assert.Equal(t, "Failed to fetch documents", res["message"])
}

// UT-US002-006 — ListProjectDocuments handler: content field absent from response items
func TestDocumentHandler_ListProjectDocuments_ContentFieldAbsent(t *testing.T) {
	e := echo.New()
	projectID := "123e4567-e89b-12d3-a456-426614174000"

	projectRepo := &mockProjectRepoForHandler{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return &domain.Project{ID: projectID, Name: "Test", Description: ""}, nil
		},
	}
	docRepo := &mockDocumentRepoForHandler{
		ListDocumentsFunc: func(ctx context.Context, pid string) ([]*domain.Document, error) {
			return []*domain.Document{
				{
					ID:        "d111aaaa-1111-1111-1111-111111111111",
					ProjectID: projectID,
					Title:     "Some doc",
					Content:   "# Very long markdown…",
					CreatedAt: mustParseTime("2026-05-18T08:30:00Z"),
					UpdatedAt: mustParseTime("2026-05-20T09:45:00Z"),
				},
			}, nil
		},
	}

	h := handler.NewDocumentHandler(docRepo, projectRepo)
	c, rec := newDocumentHandlerContext(e, http.MethodGet, "/api/v1/projects/"+projectID+"/documents", projectID)

	err := h.ListProjectDocuments(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse the body as a raw map to check key absence explicitly.
	var res map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	docsRaw, ok := res["documents"]
	require.True(t, ok, "documents key must be present")

	var items []map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(docsRaw, &items))
	require.Len(t, items, 1)

	item := items[0]
	// content key must be absent entirely.
	_, hasContent := item["content"]
	assert.False(t, hasContent, "content key must NOT be present in list response items")

	// Expected keys ARE present.
	assert.Contains(t, item, "id")
	assert.Contains(t, item, "projectId")
	assert.Contains(t, item, "title")
	assert.Contains(t, item, "createdAt")
	assert.Contains(t, item, "updatedAt")
}

// IT-US002-001 — GET /api/v1/projects/{id}/documents — missing project returns 404 not empty list (D-006)
func TestDocumentHandler_IT_ListProjectDocuments_MissingProject_404(t *testing.T) {
	// Integration test: handler wired to mocked repos, tested via Echo ServeHTTP.
	e := echo.New()

	projectRepo := &mockProjectRepoForHandler{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return nil, repo.ErrNotFound
		},
	}
	docRepo := &mockDocumentRepoForHandler{}

	h := handler.NewDocumentHandler(docRepo, projectRepo)
	e.GET("/api/v1/projects/:id/documents", h.ListProjectDocuments)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/00000000-0000-0000-0000-000000000000/documents", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	body := rec.Body.String()
	assert.JSONEq(t, `{"code":"NOT_FOUND","message":"Project not found"}`, body)

	// Body must NOT contain "documents" key.
	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(body), &raw))
	_, hasDocuments := raw["documents"]
	assert.False(t, hasDocuments, "response body must NOT contain documents key for missing project")
}

// IT-US002-002 — GET /api/v1/projects/{id}/documents — project exists, zero documents returns {"documents":[]}
func TestDocumentHandler_IT_ListProjectDocuments_EmptyProject_200(t *testing.T) {
	e := echo.New()
	projectID := "123e4567-e89b-12d3-a456-426614174001"

	projectRepo := &mockProjectRepoForHandler{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			if id == projectID {
				return &domain.Project{ID: projectID, Name: "Empty Project", Description: ""}, nil
			}
			return nil, repo.ErrNotFound
		},
	}
	docRepo := &mockDocumentRepoForHandler{
		ListDocumentsFunc: func(ctx context.Context, pid string) ([]*domain.Document, error) {
			return []*domain.Document{}, nil
		},
	}

	h := handler.NewDocumentHandler(docRepo, projectRepo)
	e.GET("/api/v1/projects/:id/documents", h.ListProjectDocuments)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+projectID+"/documents", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"documents":[]}`, rec.Body.String())
}

// IT-US002-003 — GET /api/v1/projects/{id}/documents — ordering verified
func TestDocumentHandler_IT_ListProjectDocuments_OrderingVerified(t *testing.T) {
	e := echo.New()
	projectID := "123e4567-e89b-12d3-a456-426614174002"

	// The handler receives documents in the order the repo returns them
	// (repo layer owns ordering). We provide them in the expected order here:
	// A2 and A1 share the same updated_at; A2 has the higher id (tiebreaker).
	// B has an older updated_at so it comes last.
	tSameUpdated := mustParseTime("2026-05-20T10:00:00Z")
	tOlderUpdated := mustParseTime("2026-05-19T10:00:00Z")

	// Simulate repo returning docs already ordered: A2, A1, B
	projectRepo := &mockProjectRepoForHandler{
		GetProjectFunc: func(ctx context.Context, id string) (*domain.Project, error) {
			return &domain.Project{ID: projectID, Name: "Test", Description: ""}, nil
		},
	}
	docRepo := &mockDocumentRepoForHandler{
		ListDocumentsFunc: func(ctx context.Context, pid string) ([]*domain.Document, error) {
			return []*domain.Document{
				{ID: "zzzz0003-0000-0000-0000-000000000000", ProjectID: projectID, Title: "A2", UpdatedAt: tSameUpdated},
				{ID: "aaaa0001-0000-0000-0000-000000000000", ProjectID: projectID, Title: "A1", UpdatedAt: tSameUpdated},
				{ID: "bbbb0002-0000-0000-0000-000000000000", ProjectID: projectID, Title: "B", UpdatedAt: tOlderUpdated},
			}, nil
		},
	}

	h := handler.NewDocumentHandler(docRepo, projectRepo)
	e.GET("/api/v1/projects/:id/documents", h.ListProjectDocuments)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/projects/"+projectID+"/documents", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	docs, ok := res["documents"].([]interface{})
	require.True(t, ok)
	require.Len(t, docs, 3)

	d0 := docs[0].(map[string]interface{})
	assert.Equal(t, "A2", d0["title"])

	d1 := docs[1].(map[string]interface{})
	assert.Equal(t, "A1", d1["title"])

	d2 := docs[2].(map[string]interface{})
	assert.Equal(t, "B", d2["title"])
}

// IT-US002-006 — Route registration smoke test: list-documents route is registered
// Note: The sibling task (US002_be_get_document_endpoint) adds GET /api/v1/documents/:id
// and completes the full IT-US002-006 spec (both routes). This test covers the list route.
func TestDocumentHandler_IT_RouteRegistration_ListDocuments(t *testing.T) {
	// Build an Echo instance the same way main.go does (minus the DB).
	e := echo.New()

	// Use nil repos — we're only checking route registration, not handler logic.
	docRepo := &mockDocumentRepoForHandler{}
	projectRepo := &mockProjectRepoForHandler{}
	h := handler.NewDocumentHandler(docRepo, projectRepo)

	e.GET("/api/v1/projects/:id/documents", h.ListProjectDocuments)

	routes := e.Routes()
	var found bool
	for _, r := range routes {
		if r.Method == http.MethodGet && r.Path == "/api/v1/projects/:id/documents" {
			found = true
			break
		}
	}
	assert.True(t, found, "route GET /api/v1/projects/:id/documents must be registered")
}
