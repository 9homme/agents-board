package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"agent-board/internal/domain"
)

// UT-044-001 — Requirement status constant: draft is the zero-value default
func TestRequirementStatusDraft_IsZeroValueDefault(t *testing.T) {
	r := domain.NewRequirement("proj-id-1", "My Req")
	assert.Equal(t, domain.RequirementStatusDraft, r.Status)
	assert.Equal(t, "draft", r.Status)
	assert.Equal(t, "draft", domain.RequirementStatusDraft)
}

// UT-044-002 — Requirement status enum completeness
func TestRequirementStatusConstants(t *testing.T) {
	assert.Equal(t, "draft", domain.RequirementStatusDraft)
	assert.Equal(t, "in_progress", domain.RequirementStatusInProgress)
	assert.Equal(t, "done", domain.RequirementStatusDone)
}

// UT-044-003 — Requirement domain type has all required fields
func TestRequirementStruct_HasAllRequiredFields(t *testing.T) {
	now := time.Now()
	r := domain.Requirement{
		ID:          "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f",
		ProjectID:   "11111111-1111-1111-1111-111111111111",
		Name:        "Default",
		Description: "",
		Status:      domain.RequirementStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	assert.Equal(t, "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f", r.ID)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", r.ProjectID)
	assert.Equal(t, "Default", r.Name)
	assert.Equal(t, "", r.Description)
	assert.Equal(t, "draft", r.Status)
	assert.Equal(t, now, r.CreatedAt)
	assert.Equal(t, now, r.UpdatedAt)
}

// UT-044-004 — UserStory gains RequirementID field
func TestUserStory_HasRequirementIDField(t *testing.T) {
	u := domain.UserStory{
		ID:            "us-id",
		ProjectID:     "proj-id",
		RequirementID: "req-id",
		Title:         "My Story",
		Description:   "desc",
		Status:        "draft",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	assert.Equal(t, "req-id", u.RequirementID)
	// Existing fields unchanged
	assert.Equal(t, "us-id", u.ID)
	assert.Equal(t, "proj-id", u.ProjectID)
	assert.Equal(t, "My Story", u.Title)
	assert.Equal(t, "desc", u.Description)
	assert.Equal(t, "draft", u.Status)
}

// UT-044-005 — Document gains RequirementID field
func TestDocument_HasRequirementIDField(t *testing.T) {
	d := domain.Document{
		ID:            "doc-id",
		ProjectID:     "proj-id",
		RequirementID: "req-id",
		Title:         "My Doc",
		Content:       "content",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	assert.Equal(t, "req-id", d.RequirementID)
	// Existing fields unchanged
	assert.Equal(t, "doc-id", d.ID)
	assert.Equal(t, "proj-id", d.ProjectID)
	assert.Equal(t, "My Doc", d.Title)
	assert.Equal(t, "content", d.Content)
}

// UT-044-006 — Project gains Path field (non-pointer, required)
func TestProject_HasPathField(t *testing.T) {
	p := domain.Project{
		ID:          "proj-id",
		Name:        "Test Project",
		Description: "desc",
		Path:        "/tmp/test-project",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	assert.Equal(t, "/tmp/test-project", p.Path)
	// Existing fields unchanged
	assert.Equal(t, "proj-id", p.ID)
	assert.Equal(t, "Test Project", p.Name)
	assert.Equal(t, "desc", p.Description)
}
