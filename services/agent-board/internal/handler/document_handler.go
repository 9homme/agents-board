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
	documentRepo repo.DocumentRepository
	projectRepo  repo.ProjectRepository
}

// NewDocumentHandler creates a new DocumentHandler with the given document and project repositories.
func NewDocumentHandler(documentRepo repo.DocumentRepository, projectRepo repo.ProjectRepository) *DocumentHandler {
	return &DocumentHandler{
		documentRepo: documentRepo,
		projectRepo:  projectRepo,
	}
}

// documentListItem is the per-item response shape for the list endpoint.
// It intentionally excludes the content field (architecture D-002).
type documentListItem struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Title     string `json:"title"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
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
