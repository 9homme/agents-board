package domain

import "time"

// Requirement status constants.
// Architecture: §4 — status enum "draft"|"in_progress"|"done". No state-machine
// enforcement on Requirement.status this REQ (architecture scope note).
const (
	RequirementStatusDraft      = "draft"
	RequirementStatusInProgress = "in_progress"
	RequirementStatusDone       = "done"
)

// Requirement represents the core domain entity for a requirement,
// grouping UserStories and Documents under a Project.
type Requirement struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"projectId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// NewRequirement constructs a Requirement with Status defaulting to "draft".
func NewRequirement(projectID, name string) Requirement {
	return Requirement{
		ProjectID: projectID,
		Name:      name,
		Status:    RequirementStatusDraft,
	}
}
