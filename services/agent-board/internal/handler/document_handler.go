package handler

import (
	"errors"
	"log"
	"net/http"

	"agent-board/internal/repo"

	"github.com/labstack/echo/v4"
)

// DocumentHandler handles HTTP requests for documents.
type DocumentHandler struct {
	documentRepo    repo.DocumentRepository
	projectRepo     repo.ProjectRepository
	requirementRepo repo.RequirementRepository
}

// NewDocumentHandler creates a new DocumentHandler with the given document and project repositories.
func NewDocumentHandler(documentRepo repo.DocumentRepository, projectRepo repo.ProjectRepository) *DocumentHandler {
	return &DocumentHandler{
		documentRepo: documentRepo,
		projectRepo:  projectRepo,
	}
}

// SetRequirementRepo injects a RequirementRepository so hierarchy endpoints can perform chain guards.
func (h *DocumentHandler) SetRequirementRepo(requirementRepo repo.RequirementRepository) {
	h.requirementRepo = requirementRepo
}

// documentListItem is the per-item response shape for the list endpoint.
// requirementId is included for hierarchy-scoped list endpoints (§10).
// It intentionally excludes the content field (architecture D-002).
type documentListItem struct {
	ID            string `json:"id"`
	ProjectID     string `json:"projectId"`
	RequirementID string `json:"requirementId"`
	Title         string `json:"title"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// ListProjectDocuments handles GET /api/v1/projects/:id/documents.
// It verifies the project exists (returning 404 if not, per D-006),
// then returns all documents for that project as metadata-only (no content field).
// The documents array is always present and never null; an empty project returns {"documents":[]}.
func (h *DocumentHandler) ListProjectDocuments(c echo.Context) error {
	ctx := c.Request().Context()
	projectID := c.Param("id")

	// Step 1: verify the project exists (D-006).
	if _, err := h.projectRepo.GetProject(ctx, projectID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "Project not found",
			})
		}
		log.Printf("Failed to list documents: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch documents",
		})
	}

	// Step 2: fetch the documents for the project.
	documents, err := h.documentRepo.ListDocuments(ctx, projectID)
	if err != nil {
		log.Printf("Failed to list documents: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch documents",
		})
	}

	// Build the response — metadata only, no content field.
	// Initialize with make so JSON serialises as [] not null.
	items := make([]documentListItem, 0, len(documents))
	for _, d := range documents {
		items = append(items, documentListItem{
			ID:        d.ID,
			ProjectID: d.ProjectID,
			Title:     d.Title,
			CreatedAt: d.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: d.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"documents": items,
	})
}

// GetDocument handles GET /api/v1/documents/:id.
// It returns the full document including its markdown content.
// Maps repo.ErrNotFound to 404; any other error to 500, following the shared error envelope.
func (h *DocumentHandler) GetDocument(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	doc, err := h.documentRepo.GetDocument(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "Document not found",
			})
		}
		log.Printf("Failed to get document: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch document",
		})
	}

	return c.JSON(http.StatusOK, mapDocumentToResponse(doc))
}

// checkDocumentRequirementChain fetches the requirement by rid and verifies it belongs to pid.
// Writes a 404/500 response and returns false on failure.
func (h *DocumentHandler) checkDocumentRequirementChain(c echo.Context, pid, rid string) bool {
	ctx := c.Request().Context()
	req, err := h.requirementRepo.GetRequirement(ctx, rid)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			_ = c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "Requirement not found",
			})
			return false
		}
		log.Printf("checkDocumentRequirementChain: failed to get requirement %s: %v", rid, err)
		_ = c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch documents",
		})
		return false
	}
	if req.ProjectID != pid {
		_ = c.JSON(http.StatusNotFound, map[string]string{
			"code":    "NOT_FOUND",
			"message": "Requirement not found",
		})
		return false
	}
	return true
}

// ListRequirementDocuments handles GET /api/v1/projects/:pid/requirements/:rid/documents (§10).
// Ownership chain: requirement must exist and belong to :pid.
// Returns metadata-only document items (no content) ordered by updatedAt DESC, id DESC.
func (h *DocumentHandler) ListRequirementDocuments(c echo.Context) error {
	ctx := c.Request().Context()
	pid := c.Param("pid")
	rid := c.Param("rid")

	if !h.checkDocumentRequirementChain(c, pid, rid) {
		return nil
	}

	documents, err := h.documentRepo.ListByRequirement(ctx, rid)
	if err != nil {
		log.Printf("ListRequirementDocuments: failed to list documents for requirement %s: %v", rid, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch documents",
		})
	}

	items := make([]documentListItem, 0, len(documents))
	for _, d := range documents {
		items = append(items, documentListItem{
			ID:            d.ID,
			ProjectID:     d.ProjectID,
			RequirementID: d.RequirementID,
			Title:         d.Title,
			CreatedAt:     d.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     d.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"documents": items,
	})
}

// GetRequirementDocument handles GET /api/v1/projects/:pid/requirements/:rid/documents/:docid (§11).
// Ownership chain: requirement belongs to :pid; document belongs to :rid and :pid.
// Returns the full document including content and requirementId.
func (h *DocumentHandler) GetRequirementDocument(c echo.Context) error {
	ctx := c.Request().Context()
	pid := c.Param("pid")
	rid := c.Param("rid")
	docid := c.Param("docid")

	if !h.checkDocumentRequirementChain(c, pid, rid) {
		return nil
	}

	doc, err := h.documentRepo.GetDocument(ctx, docid)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "Document not found",
			})
		}
		log.Printf("GetRequirementDocument: failed to get document %s: %v", docid, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch document",
		})
	}

	if doc.RequirementID != rid || doc.ProjectID != pid {
		return c.JSON(http.StatusNotFound, map[string]string{
			"code":    "NOT_FOUND",
			"message": "Document not found",
		})
	}

	return c.JSON(http.StatusOK, mapDocumentToResponse(doc))
}
