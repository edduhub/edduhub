// Package repository provides a data access layer for announcement-related operations.
// It abstracts database interactions, enabling clean, testable, and maintainable code
// for managing announcements within the system.
package repository

import (
	"context"
	"eduhub/server/internal/models"
	"errors"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AnnouncementRepository defines the interface for announcement data access operations.
// This interface ensures a consistent set of methods for interacting with announcement
// data, scoped appropriately by college.
type AnnouncementRepository interface {
	Create(ctx context.Context, announcement *models.Announcement) (int, error)
	GetByID(ctx context.Context, id int) (*models.Announcement, error)
	GetByCollegeID(ctx context.Context, collegeID int, limit, offset int) ([]*models.Announcement, error)
	Update(ctx context.Context, announcement *models.Announcement) error
	DeleteByID(ctx context.Context, id int) error
}

// announcementRepository implements the AnnouncementRepository interface using a PostgreSQL database.
type announcementRepository struct {
	pool *pgxpool.Pool
}

// NewAnnouncementRepository creates a new instance of AnnouncementRepository.
// It requires a database connection pool to be provided.
func NewAnnouncementRepository(db *DB) AnnouncementRepository {
	return &announcementRepository{
		pool: db.Pool,
	}
}

// Create inserts a new announcement record into the database.
// It returns the ID of the newly created announcement.
func (r *announcementRepository) Create(ctx context.Context, announcement *models.Announcement) (int, error) {
	if announcement == nil {
		return 0, errors.New("announcement cannot be nil")
	}

	now := time.Now()
	announcement.CreatedAt = now
	announcement.UpdatedAt = now

	query := `
		INSERT INTO announcements (title, content, college_id, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int
	err := r.pool.QueryRow(ctx, query,
		announcement.Title,
		announcement.Content,
		announcement.CollegeID,
		announcement.UserID,
		announcement.CreatedAt,
		announcement.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create announcement: %w", err)
	}

	return id, nil
}

// GetByID retrieves a single announcement by its primary key ID.
// It returns the announcement if found, or an error otherwise.
func (r *announcementRepository) GetByID(ctx context.Context, id int) (*models.Announcement, error) {
	query := `
		SELECT id, title, content, college_id, user_id, created_at, updated_at
		FROM announcements
		WHERE id = $1
	`

	var announcement models.Announcement
	err := pgxscan.Get(ctx, r.pool, &announcement, query, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("announcement with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get announcement by ID: %w", err)
	}

	return &announcement, nil
}

// GetByCollegeID retrieves a list of announcements for a specific college, with pagination.
// It allows for efficient retrieval of a subset of announcements.
func (r *announcementRepository) GetByCollegeID(ctx context.Context, collegeID int, limit, offset int) ([]*models.Announcement, error) {
	query := `
		SELECT id, title, content, college_id, user_id, created_at, updated_at
		FROM announcements
		WHERE college_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var announcements []*models.Announcement
	err := pgxscan.Select(ctx, r.pool, &announcements, query, collegeID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to get announcements by college ID: %w", err)
	}

	return announcements, nil
}

// Update modifies an existing announcement in the database.
// It updates the title, content, and the 'updated_at' timestamp.
func (r *announcementRepository) Update(ctx context.Context, announcement *models.Announcement) error {
	if announcement == nil {
		return errors.New("announcement cannot be nil")
	}

	announcement.UpdatedAt = time.Now()

	query := `
		UPDATE announcements
		SET title = $1, content = $2, updated_at = $3
		WHERE id = $4
	`

	cmdTag, err := r.pool.Exec(ctx, query,
		announcement.Title,
		announcement.Content,
		announcement.UpdatedAt,
		announcement.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update announcement: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("announcement with ID %d not found", announcement.ID)
	}

	return nil
}

// DeleteByID removes an announcement from the database using its ID.
func (r *announcementRepository) DeleteByID(ctx context.Context, id int) error {
	query := "DELETE FROM announcements WHERE id = $1"

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete announcement: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("announcement with ID %d not found for deletion", id)
	}

	return nil
}
