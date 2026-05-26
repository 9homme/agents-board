package handler

import (
	"errors"
	"log"
	"net/http"

	"agent-board/internal/repo"

	"github.com/labstack/echo/v4"
)

// projectResponse is the JSON shape for a single project in all project API responses.
type projectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// ProjectHandler handles HTTP requests for projects.
type ProjectHandler struct {
	repo repo.ProjectRepository
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(r repo.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{repo: r}
}

// GetProjects handles GET /api/v1/projects — returns the full list of projects.
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

// GetProject handles GET /api/v1/projects/:id — returns a single project by id.
// Returns 200 with the bare project object on success, 404 when the project does not exist,
// and 500 on unexpected errors.
func (h *ProjectHandler) GetProject(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	p, err := h.repo.GetProject(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "Project not found",
			})
		}
		log.Printf("Failed to get project: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch project",
		})
	}

	return c.JSON(http.StatusOK, projectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}
