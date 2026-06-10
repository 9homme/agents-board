package repo_test

import (
	"context"
	"testing"
	"time"

	"agent-board/internal/domain"
	"agent-board/internal/repo"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// UserStoryRepo.ListByRequirement
// ---------------------------------------------------------------------------

// UT-US-ListByReq-001 — ListByRequirement returns stories with task counts ordered by createdAt DESC
func TestUserStoryRepo_ListByRequirement_200(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	requirementID := "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f"
	projectID := "11111111-1111-1111-1111-111111111111"
	storyID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	ts := time.Date(2026, 6, 2, 9, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT us.id, us.project_id, us.requirement_id, us.title, us.description, us.status, us.created_at, us.updated_at, COUNT\(t.id\) AS task_count FROM user_stories us LEFT JOIN tasks t ON t.user_story_id = us.id WHERE us.requirement_id = \$1 GROUP BY us.id ORDER BY us.created_at DESC`).
		WithArgs(requirementID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "project_id", "requirement_id", "title", "description", "status", "created_at", "updated_at", "task_count",
		}).AddRow(storyID, projectID, requirementID, "Add item to basket", "", "in_progress", ts, ts, 1))

	r := repo.NewUserStoryRepo(db)
	stories, err := r.ListByRequirement(context.Background(), requirementID)

	require.NoError(t, err)
	require.Len(t, stories, 1)
	assert.Equal(t, storyID, stories[0].ID)
	assert.Equal(t, projectID, stories[0].ProjectID)
	assert.Equal(t, requirementID, stories[0].RequirementID)
	assert.Equal(t, "Add item to basket", stories[0].Title)
	assert.Equal(t, "in_progress", stories[0].Status)
	assert.Equal(t, 1, stories[0].TaskCount)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-US-ListByReq-002 — ListByRequirement returns empty slice (not nil)
func TestUserStoryRepo_ListByRequirement_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	requirementID := "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f"

	mock.ExpectQuery(`SELECT us.id, us.project_id, us.requirement_id, us.title, us.description, us.status, us.created_at, us.updated_at, COUNT\(t.id\) AS task_count FROM user_stories us LEFT JOIN tasks t ON t.user_story_id = us.id WHERE us.requirement_id = \$1 GROUP BY us.id ORDER BY us.created_at DESC`).
		WithArgs(requirementID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "project_id", "requirement_id", "title", "description", "status", "created_at", "updated_at", "task_count",
		}))

	r := repo.NewUserStoryRepo(db)
	stories, err := r.ListByRequirement(context.Background(), requirementID)

	require.NoError(t, err)
	assert.NotNil(t, stories, "result must not be nil")
	assert.Empty(t, stories)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-US-GetUserStory-ReqID — GetUserStory returns RequirementID
func TestUserStoryRepo_GetUserStory_ReturnsRequirementID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	storyID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	projectID := "11111111-1111-1111-1111-111111111111"
	requirementID := "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f"
	ts := time.Date(2026, 6, 2, 9, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT id, project_id, requirement_id, title, description, status, created_at, updated_at FROM user_stories WHERE id = \$1`).
		WithArgs(storyID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "project_id", "requirement_id", "title", "description", "status", "created_at", "updated_at",
		}).AddRow(storyID, projectID, requirementID, "Add item to basket", "", "in_progress", ts, ts))

	r := repo.NewUserStoryRepo(db)
	story, err := r.GetUserStory(context.Background(), storyID)

	require.NoError(t, err)
	assert.Equal(t, requirementID, story.RequirementID, "GetUserStory must return RequirementID")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// DocumentRepo.ListByRequirement
// ---------------------------------------------------------------------------

// UT-Doc-ListByReq-001 — ListByRequirement returns documents ordered by updatedAt DESC, id DESC
func TestDocumentRepo_ListByRequirement_200(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	requirementID := "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f"
	projectID := "11111111-1111-1111-1111-111111111111"
	docID := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	ts := time.Date(2026, 6, 2, 9, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT id, project_id, requirement_id, title, content, created_at, updated_at FROM documents WHERE requirement_id = \$1 ORDER BY updated_at DESC, id DESC`).
		WithArgs(requirementID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "project_id", "requirement_id", "title", "content", "created_at", "updated_at",
		}).AddRow(docID, projectID, requirementID, "README", "# README\n...", ts, ts))

	r := repo.NewDocumentRepo(db)
	docs, err := r.ListByRequirement(context.Background(), requirementID)

	require.NoError(t, err)
	require.Len(t, docs, 1)
	assert.Equal(t, docID, docs[0].ID)
	assert.Equal(t, projectID, docs[0].ProjectID)
	assert.Equal(t, requirementID, docs[0].RequirementID)
	assert.Equal(t, "README", docs[0].Title)
	assert.Equal(t, "# README\n...", docs[0].Content)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-Doc-ListByReq-002 — ListByRequirement returns empty slice (not nil)
func TestDocumentRepo_ListByRequirement_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	requirementID := "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f"

	mock.ExpectQuery(`SELECT id, project_id, requirement_id, title, content, created_at, updated_at FROM documents WHERE requirement_id = \$1 ORDER BY updated_at DESC, id DESC`).
		WithArgs(requirementID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "project_id", "requirement_id", "title", "content", "created_at", "updated_at",
		}))

	r := repo.NewDocumentRepo(db)
	docs, err := r.ListByRequirement(context.Background(), requirementID)

	require.NoError(t, err)
	assert.NotNil(t, docs, "result must not be nil")
	assert.Empty(t, docs)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UT-Doc-GetDocument-ReqID — GetDocument returns RequirementID
func TestDocumentRepo_GetDocument_ReturnsRequirementID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	docID := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	projectID := "11111111-1111-1111-1111-111111111111"
	requirementID := "b2e9d0c1-2f3a-4b5c-8d7e-1a2b3c4d5e6f"
	ts := time.Date(2026, 6, 2, 9, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT id, project_id, requirement_id, title, content, created_at, updated_at FROM documents WHERE id = \$1`).
		WithArgs(docID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "project_id", "requirement_id", "title", "content", "created_at", "updated_at",
		}).AddRow(docID, projectID, requirementID, "README", "# README\n...", ts, ts))

	r := repo.NewDocumentRepo(db)
	doc, err := r.GetDocument(context.Background(), docID)

	require.NoError(t, err)
	assert.Equal(t, requirementID, doc.RequirementID, "GetDocument must return RequirementID")

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Ensure domain.Document has RequirementID (compile-time check)
var _ = domain.Document{RequirementID: "test"}
