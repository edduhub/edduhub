package repository

import (
	"context"
	"eduhub/server/internal/models"
	"regexp"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAnnouncementTest(t *testing.T) (pgxmock.PgxPoolIface, AnnouncementRepository, context.Context) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	repo := &announcementRepository{pool: mock}
	ctx := context.Background()

	return mock, repo, ctx
}

func TestCreateAnnouncement(t *testing.T) {
	mock, repo, ctx := setupAnnouncementTest(t)
	defer mock.Close()

	announcement := &models.Announcement{
		Title:     "Test Announcement",
		Content:   "This is a test announcement.",
		CollegeID: 1,
		UserID:    1,
	}
	expectedID := 1

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO announcements (title, content, college_id, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`)).
		WithArgs(announcement.Title, announcement.Content, announcement.CollegeID, announcement.UserID, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(expectedID))

	id, err := repo.Create(ctx, announcement)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAnnouncementByID(t *testing.T) {
	mock, repo, ctx := setupAnnouncementTest(t)
	defer mock.Close()

	expectedAnnouncement := &models.Announcement{
		ID:        1,
		Title:     "Test Announcement",
		Content:   "This is a test announcement.",
		CollegeID: 1,
		UserID:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := pgxmock.NewRows([]string{"id", "title", "content", "college_id", "user_id", "created_at", "updated_at"}).
		AddRow(expectedAnnouncement.ID, expectedAnnouncement.Title, expectedAnnouncement.Content, expectedAnnouncement.CollegeID, expectedAnnouncement.UserID, expectedAnnouncement.CreatedAt, expectedAnnouncement.UpdatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, title, content, college_id, user_id, created_at, updated_at
		FROM announcements
		WHERE id = $1`)).
		WithArgs(expectedAnnouncement.ID).
		WillReturnRows(rows)

	announcement, err := repo.GetByID(ctx, expectedAnnouncement.ID)

	assert.NoError(t, err)
	assert.Equal(t, expectedAnnouncement, announcement)
	assert.NoError(t, mock.ExpectationsWereMet())
}
