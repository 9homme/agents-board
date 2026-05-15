package handler

import (
	"log"
	"net/http"

	"agent-board/internal/repo"

	"github.com/labstack/echo/v4"
)

// ProjectHandler handles HTTP requests for projects.
type ProjectHandler struct {
	repo repo.ProjectRepository
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(r repo.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{repo: r}
}

// GetProjects handles GET /api/v1/projects.
func (h *ProjectHandler) GetProjects(c echo.Context) error {
	ctx := c.Request().Context()

	projects, err := h.repo.ListProjects(ctx)
	if err != nil {
		log.Printf("Failed to list projects: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch projects",
		})
	}

	// Format response to match API contract exactly.
	// domain.Project json struct tags might not match exactly, so we map them here
	// Wait, let me check domain.Project tags. If they match, I can return them.
	// But it's safer to build the exact JSON contract shape.

	type projectResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedAt   string `json:"createdAt"`
		UpdatedAt   string `json:"updatedAt"`
	}

	res := make([]projectResponse, 0)
	for _, p := range projects {
		res = append(res, projectResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"projects": res,
	})
}
