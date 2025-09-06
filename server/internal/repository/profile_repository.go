package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5" // For pgx.ErrNoRows
)

const profileTable = "profiles"

type ProfileRepository interface {
	CreateProfile(ctx context.Context, profile *models.Profile) error
	GetProfileByUserID(ctx context.Context, userID string) (*models.Profile, error)
	GetProfileByID(ctx context.Context, profileID int) (*models.Profile, error)
	UpdateProfile(ctx context.Context, profile *models.Profile) error
	DeleteProfile(ctx context.Context, profile *models.Profile) error
}

type profileRepository struct {
	DB *DB
}

func NewProfileRepository(db *DB) ProfileRepository {
	return &profileRepository{DB: db}
}

func (r *profileRepository) CreateProfile(ctx context.Context, profile *models.Profile) error {
	now := time.Now()
	if profile.JoinedAt.IsZero() {
		profile.JoinedAt = now
	}
	profile.LastActive = now
	profile.CreatedAt = now
	profile.UpdatedAt = now

	if profile.Preferences == nil {
		profile.Preferences = make(models.JSONMap)
	}
	if profile.SocialLinks == nil {
		profile.SocialLinks = make(models.JSONMap)
	}

	sql := `INSERT INTO profiles (user_id, college_id, bio, profile_image, phone_number, address, date_of_birth, joined_at, last_active, preferences, social_links, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`
	temp := struct {
		ID int `db:"id"`
	}{}
	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql, profile.UserID, profile.CollegeID, profile.Bio, profile.ProfileImage, profile.PhoneNumber, profile.Address, profile.DateOfBirth, profile.JoinedAt, profile.LastActive, profile.Preferences, profile.SocialLinks, profile.CreatedAt, profile.UpdatedAt)
	if err != nil {
		return fmt.Errorf("CreateProfile: failed to execute query or scan ID: %w", err)
	}
	profile.ID = temp.ID
	return nil
}

func (r *profileRepository) GetProfileByUserID(ctx context.Context, userID string) (*models.Profile, error) {
	profile := &models.Profile{}
	sql := `SELECT id, user_id, college_id, bio, profile_image, phone_number, address, date_of_birth, joined_at, last_active, preferences, social_links, created_at, updated_at FROM profiles WHERE user_id = $1`
	err := pgxscan.Get(ctx, r.DB.Pool, profile, sql, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetProfileByUserID: profile for user ID %s not found", userID)
		}
		return nil, fmt.Errorf("GetProfileByUserID: failed to execute query or scan: %w", err)
	}
	return profile, nil
}

func (r *profileRepository) GetProfileByID(ctx context.Context, profileID int) (*models.Profile, error) {
	profile := &models.Profile{}
	sql := `SELECT id, user_id, college_id, bio, profile_image, phone_number, address, date_of_birth, joined_at, last_active, preferences, social_links, created_at, updated_at FROM profiles WHERE id = $1`
	err := pgxscan.Get(ctx, r.DB.Pool, profile, sql, profileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetProfileByID: profile with ID %d not found", profileID)
		}
		return nil, fmt.Errorf("GetProfileByID: failed to execute query or scan: %w", err)
	}
	return profile, nil
}

func (r *profileRepository) UpdateProfile(ctx context.Context, profile *models.Profile) error {
	now := time.Now()
	profile.LastActive = now
	profile.UpdatedAt = now

	if profile.Preferences == nil {
		profile.Preferences = make(models.JSONMap)
	}
	if profile.SocialLinks == nil {
		profile.SocialLinks = make(models.JSONMap)
	}

	sql := `UPDATE profiles SET college_id = $1, bio = $2, profile_image = $3, phone_number = $4, address = $5, date_of_birth = $6, last_active = $7, preferences = $8, social_links = $9, updated_at = $10 WHERE id = $11`
	commandTag, err := r.DB.Pool.Exec(ctx, sql, profile.CollegeID, profile.Bio, profile.ProfileImage, profile.PhoneNumber, profile.Address, profile.DateOfBirth, profile.LastActive, profile.Preferences, profile.SocialLinks, profile.UpdatedAt, profile.ID)
	if err != nil {
		return fmt.Errorf("UpdateProfile: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateProfile: no profile found with ID %d, or no changes made", profile.ID)
	}
	return nil
}


func (r *profileRepository) DeleteProfile(ctx context.Context, profile *models.Profile) error {
	sql := `DELETE FROM profiles WHERE id = $1`
	commandTag, err := r.DB.Pool.Exec(ctx, sql, profile.ID)
	if err != nil {
		return fmt.Errorf("DeleteProfile: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteProfile: no profile found with ID %d", profile.ID)
	}

	return nil
}