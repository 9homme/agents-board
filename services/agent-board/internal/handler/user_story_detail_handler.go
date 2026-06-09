package handler

import (
	"errors"
	"log"
	"net/http"

	"agent-board/internal/repo"

	"github.com/labstack/echo/v4"
)

// userStoryDetailResponse is the JSON shape returned by GET /api/v1/user-stories/{id}.
// taskCount is intentionally omitted — the detail endpoint returns a bare story object.
type userStoryDetailResponse struct {
	ID          string `json:"id"`
	ProjectID   string `json:"projectId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// taskResponse is the per-task JSON shape in the GET /api/v1/user-stories/{id}/tasks response.
type taskResponse struct {
	ID          string `json:"id"`
	UserStoryID string `json:"userStoryId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// SetTaskRepo injects a TaskRepository into the UserStoryHandler so the
// detail and tasks endpoints can access tasks without creating a new handler type.
func (h *UserStoryHandler) SetTaskRepo(taskRepo repo.TaskRepository) {
	h.taskRepo = taskRepo
}

// GetUserStory handles GET /api/v1/user-stories/:id.
// Returns the bare story object (no taskCount) or 404/500 on error.
func (h *UserStoryHandler) GetUserStory(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	story, err := h.userStoryRepo.GetUserStory(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "User story not found",
			})
		}
		log.Printf("GetUserStory: failed to fetch story %s: %v", id, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Internal server error",
		})
	}

	return c.JSON(http.StatusOK, userStoryDetailResponse{
		ID:          story.ID,
		ProjectID:   story.ProjectID,
		Title:       story.Title,
		Description: story.Description,
		Status:      story.Status,
		CreatedAt:   story.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   story.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// GetUserStoryTasks handles GET /api/v1/user-stories/:id/tasks.
// Returns {"tasks":[...]} — empty array [] is returned when the story has no tasks.
// Returns 404 if the story is not found, 500 on other errors.
//
// Because ListTasks returns an empty slice (not ErrNotFound) when the story
// does not exist, the handler explicitly verifies story existence via
// GetUserStory before returning the task list.
func (h *UserStoryHandler) GetUserStoryTasks(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	tasks, err := h.taskRepo.ListTasks(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "User story not found",
			})
		}
		log.Printf("GetUserStoryTasks: failed to list tasks for story %s: %v", id, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Internal server error",
		})
	}

	// ListTasks returns an empty slice for both "story exists, no tasks" and
	// "story does not exist". Verify story existence to return the correct 404.
	if len(tasks) == 0 {
		if _, err := h.userStoryRepo.GetUserStory(ctx, id); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"code":    "NOT_FOUND",
					"message": "User story not found",
				})
			}
			log.Printf("GetUserStoryTasks: failed to verify story %s: %v", id, err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": "Internal server error",
			})
		}
	}

	items := make([]taskResponse, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, taskResponse{
			ID:          t.ID,
			UserStoryID: t.UserStoryID,
			Title:       t.Title,
			Description: t.Description,
			Status:      t.Status,
			CreatedAt:   t.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   t.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"tasks": items,
	})
}
