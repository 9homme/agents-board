package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"agent-board/internal/domain"
	"agent-board/internal/repo"

	"github.com/labstack/echo/v4"
)

// projectResponse is the JSON shape for a single project in all project API responses.
type projectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// ProjectHandler handles HTTP requests for projects.
type ProjectHandler struct {
	repo      repo.ProjectRepository
	validator PathValidator
}

// NewProjectHandler creates a new ProjectHandler using the real filesystem validator.
func NewProjectHandler(r repo.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{repo: r, validator: fsutilValidator{}}
}

// newProjectHandlerWithValidator creates a ProjectHandler with an injectable validator (for testing).
func newProjectHandlerWithValidator(r repo.ProjectRepository, v PathValidator) *ProjectHandler {
	return &ProjectHandler{repo: r, validator: v}
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
			Path:        p.Path,
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
		Path:        p.Path,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// CreateProject handles POST /api/v1/projects — creates a new project.
// Returns 201 on success, 400 on validation failure, 409 on duplicate path, 500 on unexpected errors.
func (h *ProjectHandler) CreateProject(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Path        string `json:"path"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"code":    "VALIDATION_ERROR",
			"message": "invalid request body",
		})
	}

	// Blank/absent path check before calling validator (D-006).
	req.Path = strings.TrimSpace(req.Path)
	if req.Path == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"code":    "VALIDATION_ERROR",
			"message": "path is required",
		})
	}

	// Filesystem validation: must exist and be a directory.
	if err := h.validator.ValidatePath(req.Path); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"code":    "VALIDATION_ERROR",
			"message": "path does not exist or is not a directory",
		})
	}

	// Name validation (trimmed, non-blank).
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"code":    "VALIDATION_ERROR",
			"message": "name is required",
		})
	}

	p, err := h.repo.CreateProject(ctx, &domain.Project{
		Name:        req.Name,
		Description: req.Description,
		Path:        req.Path,
	})
	if err != nil {
		if errors.Is(err, repo.ErrDuplicatePath) {
			return c.JSON(http.StatusConflict, map[string]string{
				"code":    "DUPLICATE_PATH",
				"message": "path already linked to another project",
			})
		}
		log.Printf("Failed to create project: count=1, code=INTERNAL_ERROR")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create project",
		})
	}

	return c.JSON(http.StatusCreated, projectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Path:        p.Path,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}
