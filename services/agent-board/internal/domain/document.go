package domain

import "time"

// Document represents the core domain entity for a document belonging to a project and requirement.
type Document struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"projectId"`
	RequirementID string    `json:"requirementId"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
