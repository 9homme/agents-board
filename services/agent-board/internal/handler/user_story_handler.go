package handler

import (
	"errors"
	"log"
	"net/http"

	"agent-board/internal/repo"

	"github.com/labstack/echo/v4"
)

// userStoryListItem is the per-item response shape for the user stories list endpoint.
type userStoryListItem struct {
	ID          string `json:"id"`
	ProjectID   string `json:"projectId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	TaskCount   int    `json:"taskCount"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// UserStoryHandler handles HTTP requests for user stories.
type UserStoryHandler struct {
	userStoryRepo repo.UserStoryRepository
	projectRepo   repo.ProjectRepository
}

// NewUserStoryHandler creates a new UserStoryHandler with the given user story and project repositories.
func NewUserStoryHandler(userStoryRepo repo.UserStoryRepository, projectRepo repo.ProjectRepository) *UserStoryHandler {
	return &UserStoryHandler{
		userStoryRepo: userStoryRepo,
		projectRepo:   projectRepo,
	}
}

// GetProjectUserStories handles GET /api/v1/projects/:id/user-stories.
// It verifies the project exists (returning 404 if not),
// then returns all user stories for that project with their task counts.
// Results are ordered by created_at DESC. The userStories array is never null.
func (h *UserStoryHandler) GetProjectUserStories(c echo.Context) error {
	ctx := c.Request().Context()
	projectID := c.Param("id")

	// Step 1: verify the project exists.
	if _, err := h.projectRepo.GetProject(ctx, projectID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "Project not found",
			})
		}
		log.Printf("Failed to verify project: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Internal server error",
		})
	}

	// Step 2: fetch user stories with task counts.
	stories, err := h.userStoryRepo.ListUserStoriesWithTaskCount(ctx, projectID)
	if err != nil {
		log.Printf("Failed to list user stories: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch user stories",
		})
	}

	// Build response — always an array, never null.
	items := make([]userStoryListItem, 0, len(stories))
	for _, s := range stories {
		items = append(items, userStoryListItem{
			ID:          s.ID,
			ProjectID:   s.ProjectID,
			Title:       s.Title,
			Description: s.Description,
			Status:      s.Status,
			TaskCount:   s.TaskCount,
			CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"userStories": items,
	})
}
