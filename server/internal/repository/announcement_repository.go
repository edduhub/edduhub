package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type AnnouncementRepository interface {
	CreateAnnouncement(ctx context.Context, announcement *models.Announcement) error
	GetAnnouncementByID(ctx context.Context, collegeID int, announcementID int) (*models.Announcement, error)
	UpdateAnnouncement(ctx context.Context, announcement *models.Announcement) error
	UpdateAnnouncementPartial(ctx context.Context, collegeID int, announcementID int, req *models.UpdateAnnouncementRequest) error
	DeleteAnnouncement(ctx context.Context, collegeID int, announcementID int) error
	GetAnnouncements(ctx context.Context, filter models.AnnouncementFilter) ([]*models.Announcement, error)
	CountAnnouncements(ctx context.Context, filter models.AnnouncementFilter) (int, error)
}

type announcementRepository struct {
	DB *DB
}

func NewAnnouncementRepository(db *DB) AnnouncementRepository {
	return &announcementRepository{DB: db}
}

func (r *announcementRepository) CreateAnnouncement(ctx context.Context, announcement *models.Announcement) error {
	now := time.Now()
	announcement.CreatedAt = now
	announcement.UpdatedAt = now

	if announcement.PublishedAt == nil && announcement.IsPublished {
		announcement.PublishedAt = &now
	}

	sql := `INSERT INTO announcements (college_id, course_id, title, content, priority, is_published, published_at, expires_at, created_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id`

	temp := struct {
		ID int `db:"id"`
	}{}

	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql,
		announcement.CollegeID, announcement.CourseID, announcement.Title, announcement.Content,
		announcement.Priority, announcement.IsPublished, announcement.PublishedAt, announcement.ExpiresAt,
		announcement.CreatedBy, announcement.CreatedAt, announcement.UpdatedAt)

	if err != nil {
		return fmt.Errorf("CreateAnnouncement: failed to execute query or scan ID: %w", err)
	}
	announcement.ID = temp.ID
	return nil
}

func (r *announcementRepository) GetAnnouncementByID(ctx context.Context, collegeID int, announcementID int) (*models.Announcement, error) {
	announcement := &models.Announcement{}
	sql := `SELECT id, college_id, course_id, title, content, priority, is_published, published_at, expires_at, created_by, created_at, updated_at
			FROM announcements
			WHERE id = $1 AND college_id = $2`

	err := pgxscan.Get(ctx, r.DB.Pool, announcement, sql, announcementID, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetAnnouncementByID: announcement with ID %d not found for college ID %d", announcementID, collegeID)
		}
		return nil, fmt.Errorf("GetAnnouncementByID: failed to execute query or scan: %w", err)
	}
	return announcement, nil
}

func (r *announcementRepository) UpdateAnnouncement(ctx context.Context, announcement *models.Announcement) error {
	announcement.UpdatedAt = time.Now()

	sql := `UPDATE announcements
			SET title = $1, content = $2, priority = $3, is_published = $4, published_at = $5, expires_at = $6, created_by = $7, updated_at = $8
			WHERE id = $9 AND college_id = $10`

	cmdTag, err := r.DB.Pool.Exec(ctx, sql,
		announcement.Title, announcement.Content, announcement.Priority, announcement.IsPublished,
		announcement.PublishedAt, announcement.ExpiresAt, announcement.CreatedBy, announcement.UpdatedAt,
		announcement.ID, announcement.CollegeID)

	if err != nil {
		return fmt.Errorf("UpdateAnnouncement: failed to execute query: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateAnnouncement: no announcement found with ID %d for college ID %d", announcement.ID, announcement.CollegeID)
	}
	return nil
}

func (r *announcementRepository) UpdateAnnouncementPartial(ctx context.Context, collegeID int, announcementID int, req *models.UpdateAnnouncementRequest) error {
	sql := `UPDATE announcements SET updated_at = NOW()`
	args := []interface{}{}
	argIndex := 1

	if req.Title != nil {
		sql += fmt.Sprintf(`, title = $%d`, argIndex)
		args = append(args, *req.Title)
		argIndex++
	}
	if req.Content != nil {
		sql += fmt.Sprintf(`, content = $%d`, argIndex)
		args = append(args, *req.Content)
		argIndex++
	}
	if req.Priority != nil {
		sql += fmt.Sprintf(`, priority = $%d`, argIndex)
		args = append(args, *req.Priority)
		argIndex++
	}
	if req.IsPublished != nil {
		sql += fmt.Sprintf(`, is_published = $%d`, argIndex)
		args = append(args, *req.IsPublished)
		argIndex++
	}
	if req.PublishedAt != nil {
		sql += fmt.Sprintf(`, published_at = $%d`, argIndex)
		args = append(args, *req.PublishedAt)
		argIndex++
	}
	if req.ExpiresAt != nil {
		sql += fmt.Sprintf(`, expires_at = $%d`, argIndex)
		args = append(args, *req.ExpiresAt)
		argIndex++
	}
	if req.CreatedBy != nil {
		sql += fmt.Sprintf(`, created_by = $%d`, argIndex)
		args = append(args, *req.CreatedBy)
		argIndex++
	}

	if len(args) == 0 {
		return errors.New("UpdateAnnouncementPartial: no fields to update")
	}

	sql += fmt.Sprintf(` WHERE id = $%d AND college_id = $%d`, argIndex, argIndex+1)
	args = append(args, announcementID, collegeID)

	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateAnnouncementPartial: failed to execute query: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateAnnouncementPartial: announcement with ID %d not found for college ID %d", announcementID, collegeID)
	}

	return nil
}

func (r *announcementRepository) DeleteAnnouncement(ctx context.Context, collegeID int, announcementID int) error {
	sql := `DELETE FROM announcements WHERE id = $1 AND college_id = $2`

	cmdTag, err := r.DB.Pool.Exec(ctx, sql, announcementID, collegeID)
	if err != nil {
		return fmt.Errorf("DeleteAnnouncement: failed to execute query: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteAnnouncement: no announcement found with ID %d for college ID %d", announcementID, collegeID)
	}
	return nil
}

func (r *announcementRepository) GetAnnouncements(ctx context.Context, filter models.AnnouncementFilter) ([]*models.Announcement, error) {
	if filter.CollegeID == nil {
		return nil, errors.New("GetAnnouncements: CollegeID filter is required")
	}

	sql := `SELECT id, college_id, course_id, title, content, priority, is_published, published_at, expires_at, created_by, created_at, updated_at
			FROM announcements WHERE college_id = $1`
	args := []interface{}{*filter.CollegeID}
	argIndex := 2

	if filter.CourseID != nil {
		sql += fmt.Sprintf(` AND course_id = $%d`, argIndex)
		args = append(args, *filter.CourseID)
		argIndex++
	}
	if filter.Priority != nil {
		sql += fmt.Sprintf(` AND priority = $%d`, argIndex)
		args = append(args, *filter.Priority)
		argIndex++
	}
	if filter.IsPublished != nil {
		sql += fmt.Sprintf(` AND is_published = $%d`, argIndex)
		args = append(args, *filter.IsPublished)
		argIndex++
	}

	sql += ` ORDER BY created_at DESC`

	if filter.Limit > 0 {
		sql += fmt.Sprintf(` LIMIT %d`, filter.Limit)
	}
	if filter.Offset > 0 {
		sql += fmt.Sprintf(` OFFSET %d`, filter.Offset)
	}

	var announcements []*models.Announcement
	err := pgxscan.Select(ctx, r.DB.Pool, &announcements, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*models.Announcement{}, nil
		}
		return nil, fmt.Errorf("GetAnnouncements: failed to execute query or scan: %w", err)
	}
	return announcements, nil
}

func (r *announcementRepository) CountAnnouncements(ctx context.Context, filter models.AnnouncementFilter) (int, error) {
	if filter.CollegeID == nil {
		return 0, errors.New("CountAnnouncements: CollegeID filter is required")
	}

	sql := `SELECT COUNT(*) FROM announcements WHERE college_id = $1`
	args := []interface{}{*filter.CollegeID}
	argIndex := 2

	if filter.CourseID != nil {
		sql += fmt.Sprintf(` AND course_id = $%d`, argIndex)
		args = append(args, *filter.CourseID)
		argIndex++
	}
	if filter.Priority != nil {
		sql += fmt.Sprintf(` AND priority = $%d`, argIndex)
		args = append(args, *filter.Priority)
		argIndex++
	}
	if filter.IsPublished != nil {
		sql += fmt.Sprintf(` AND is_published = $%d`, argIndex)
		args = append(args, *filter.IsPublished)
		argIndex++
	}

	temp := struct {
		Count int `db:"count"`
	}{}
	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("CountAnnouncements: failed to execute query or scan: %w", err)
	}
	return temp.Count, nil
}
