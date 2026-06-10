package handler

import (
	"errors"
	"log"
	"net/http"

	"agent-board/internal/repo"

	"github.com/labstack/echo/v4"
)

// userStoryListItem is the per-item response shape for the user stories list endpoint.
// requirementId is included for hierarchy-scoped list endpoints (§6).
type userStoryListItem struct {
	ID            string `json:"id"`
	ProjectID     string `json:"projectId"`
	RequirementID string `json:"requirementId"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	TaskCount     int    `json:"taskCount"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// UserStoryHandler handles HTTP requests for user stories.
type UserStoryHandler struct {
	userStoryRepo   repo.UserStoryRepository
	projectRepo     repo.ProjectRepository
	taskRepo        repo.TaskRepository
	requirementRepo repo.RequirementRepository
}

// NewUserStoryHandler creates a new UserStoryHandler with the given user story and project repositories.
func NewUserStoryHandler(userStoryRepo repo.UserStoryRepository, projectRepo repo.ProjectRepository) *UserStoryHandler {
	return &UserStoryHandler{
		userStoryRepo: userStoryRepo,
		projectRepo:   projectRepo,
	}
}

// SetRequirementRepo injects a RequirementRepository so hierarchy endpoints can perform chain guards.
func (h *UserStoryHandler) SetRequirementRepo(requirementRepo repo.RequirementRepository) {
	h.requirementRepo = requirementRepo
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
			"message": "Failed to fetch user stories",
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
			ID:            s.ID,
			ProjectID:     s.ProjectID,
			RequirementID: s.RequirementID,
			Title:         s.Title,
			Description:   s.Description,
			Status:        s.Status,
			TaskCount:     s.TaskCount,
			CreatedAt:     s.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"userStories": items,
	})
}

// checkRequirementChain fetches the requirement by rid and verifies it belongs to pid.
// Returns the requirement on success or writes a 404/500 response and returns nil.
func (h *UserStoryHandler) checkRequirementChain(c echo.Context, pid, rid string) bool {
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
		log.Printf("checkRequirementChain: failed to get requirement %s: %v", rid, err)
		_ = c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch user stories",
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

// ListRequirementUserStories handles GET /api/v1/projects/:pid/requirements/:rid/user-stories (§6).
// Ownership chain: requirement must exist and belong to :pid.
// Returns all user stories under the requirement with task counts ordered by createdAt DESC.
func (h *UserStoryHandler) ListRequirementUserStories(c echo.Context) error {
	ctx := c.Request().Context()
	pid := c.Param("pid")
	rid := c.Param("rid")

	if !h.checkRequirementChain(c, pid, rid) {
		return nil
	}

	stories, err := h.userStoryRepo.ListByRequirement(ctx, rid)
	if err != nil {
		log.Printf("ListRequirementUserStories: failed to list stories for requirement %s: %v", rid, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch user stories",
		})
	}

	items := make([]userStoryListItem, 0, len(stories))
	for _, s := range stories {
		items = append(items, userStoryListItem{
			ID:            s.ID,
			ProjectID:     s.ProjectID,
			RequirementID: s.RequirementID,
			Title:         s.Title,
			Description:   s.Description,
			Status:        s.Status,
			TaskCount:     s.TaskCount,
			CreatedAt:     s.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"userStories": items,
	})
}

// GetRequirementUserStory handles GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid (§7).
// Ownership chain: requirement belongs to :pid; story belongs to :rid and :pid.
// Returns the bare story detail (no taskCount).
func (h *UserStoryHandler) GetRequirementUserStory(c echo.Context) error {
	ctx := c.Request().Context()
	pid := c.Param("pid")
	rid := c.Param("rid")
	usid := c.Param("usid")

	if !h.checkRequirementChain(c, pid, rid) {
		return nil
	}

	story, err := h.userStoryRepo.GetUserStory(ctx, usid)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "User story not found",
			})
		}
		log.Printf("GetRequirementUserStory: failed to get story %s: %v", usid, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Internal server error",
		})
	}

	if story.RequirementID != rid || story.ProjectID != pid {
		return c.JSON(http.StatusNotFound, map[string]string{
			"code":    "NOT_FOUND",
			"message": "User story not found",
		})
	}

	return c.JSON(http.StatusOK, userStoryDetailResponse{
		ID:            story.ID,
		ProjectID:     story.ProjectID,
		RequirementID: story.RequirementID,
		Title:         story.Title,
		Description:   story.Description,
		Status:        story.Status,
		CreatedAt:     story.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     story.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// GetRequirementUserStoryTasks handles GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks (§8).
// Ownership chain: requirement belongs to :pid; story belongs to :rid and :pid; then returns tasks.
func (h *UserStoryHandler) GetRequirementUserStoryTasks(c echo.Context) error {
	ctx := c.Request().Context()
	pid := c.Param("pid")
	rid := c.Param("rid")
	usid := c.Param("usid")

	if !h.checkRequirementChain(c, pid, rid) {
		return nil
	}

	story, err := h.userStoryRepo.GetUserStory(ctx, usid)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "User story not found",
			})
		}
		log.Printf("GetRequirementUserStoryTasks: failed to get story %s: %v", usid, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Internal server error",
		})
	}

	if story.RequirementID != rid || story.ProjectID != pid {
		return c.JSON(http.StatusNotFound, map[string]string{
			"code":    "NOT_FOUND",
			"message": "User story not found",
		})
	}

	tasks, err := h.taskRepo.ListTasks(ctx, usid)
	if err != nil {
		log.Printf("GetRequirementUserStoryTasks: failed to list tasks for story %s: %v", usid, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to fetch tasks",
		})
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

// GetRequirementTask handles GET /api/v1/projects/:pid/requirements/:rid/user-stories/:usid/tasks/:tid (§9).
// Ownership chain: requirement belongs to :pid; story belongs to :rid and :pid; task belongs to :usid.
func (h *UserStoryHandler) GetRequirementTask(c echo.Context) error {
	ctx := c.Request().Context()
	pid := c.Param("pid")
	rid := c.Param("rid")
	usid := c.Param("usid")
	tid := c.Param("tid")

	if !h.checkRequirementChain(c, pid, rid) {
		return nil
	}

	story, err := h.userStoryRepo.GetUserStory(ctx, usid)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "User story not found",
			})
		}
		log.Printf("GetRequirementTask: failed to get story %s: %v", usid, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Internal server error",
		})
	}

	if story.RequirementID != rid || story.ProjectID != pid {
		return c.JSON(http.StatusNotFound, map[string]string{
			"code":    "NOT_FOUND",
			"message": "User story not found",
		})
	}

	task, err := h.taskRepo.GetTask(ctx, tid)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"code":    "NOT_FOUND",
				"message": "Task not found",
			})
		}
		log.Printf("GetRequirementTask: failed to get task %s: %v", tid, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"code":    "INTERNAL_ERROR",
			"message": "Internal server error",
		})
	}

	if task.UserStoryID != usid {
		return c.JSON(http.StatusNotFound, map[string]string{
			"code":    "NOT_FOUND",
			"message": "Task not found",
		})
	}

	return c.JSON(http.StatusOK, taskResponse{
		ID:          task.ID,
		UserStoryID: task.UserStoryID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   task.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}
