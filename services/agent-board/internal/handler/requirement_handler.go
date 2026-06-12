package handler

import (
	"errors"
	"log"
	"net/http"

	"agent-board/internal/repo"

	"github.com/labstack/echo/v4"
)

// requirementListItem is the per-item response shape for the requirements list endpoint.
type requirementListItem struct {
	ID          string `json:"id"`
	ProjectID   string `json:"projectId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// RequirementHandler handles HTTP requests for requirements.
type RequirementHandler struct {
	requirementRepo repo.RequirementRepository
	projectRepo     repo.ProjectRepository
}

// NewRequirementHandler creates a new RequirementHandler.
func NewRequirementHandler(requirementRepo repo.RequirementRepository, projectRepo repo.ProjectRepository) *RequirementHandler {
	return &RequirementHandler{
		requirementRepo: requirementRepo,
		projectRepo:     projectRepo,
	}
}

// ListProjectRequirements handles GET /api/v1/projects/:pid/requirements.
// It verifies the project exists (returning 404 if not), then returns all requirements
// for that project ordered by created_at ASC. The requirements array is never null.
func (h *RequirementHandler) ListProjectRequirements(c echo.Context) error {
	ctx := c.Request().Context()
	projectID := c.Param("pid")

	// Step 1: verify the project exists.
	if _, err := h.projectRepo.GetProject(ctx, projectID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "Project not found",
			})
		}
		log.Printf("Failed to verify project %s: %v", projectID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch requirements",
		})
	}

	// Step 2: fetch requirements ordered by created_at ASC.
	requirements, err := h.requirementRepo.ListByProject(ctx, projectID)
	if err != nil {
		log.Printf("Failed to list requirements for project %s: %v", projectID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch requirements",
		})
	}

	// Build response — always an array, never null.
	items := make([]requirementListItem, 0, len(requirements))
	for _, r := range requirements {
		items = append(items, requirementListItem{
			ID:          r.ID,
			ProjectID:   r.ProjectID,
			Name:        r.Name,
			Description: r.Description,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   r.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"requirements": items,
	})
}
