package domain

import "time"

// UserStory represents a user story entity belonging to a project and requirement.
type UserStory struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"projectId"`
	RequirementID string    `json:"requirementId"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
