package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5" // For pgx.ErrNoRows
)

const profileTable = "profiles"

type ProfileRepository interface {
	CreateProfile(ctx context.Context, profile *models.Profile) error
	GetProfileByUserID(ctx context.Context, userID int) (*models.Profile, error)
	GetProfileByID(ctx context.Context, profileID int) (*models.Profile, error)
	UpdateProfile(ctx context.Context, profile *models.Profile) error
	UpdateProfilePartial(ctx context.Context, profileID int, req *models.UpdateProfileRequest) error
	DeleteProfile(ctx context.Context, profile *models.Profile) error
	CreateProfileHistory(ctx context.Context, history *models.ProfileHistory) error
	GetProfileHistory(ctx context.Context, profileID int, limit, offset int) ([]*models.ProfileHistory, error)
	GetProfileByKratosID(ctx context.Context, kratosID string) (*models.Profile, error)
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

	sql := `INSERT INTO profiles (user_id, college_id, first_name, last_name, bio, profile_image, phone_number, address, date_of_birth, joined_at, last_active, preferences, social_links, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id`
	temp := struct {
		ID int `db:"id"`
	}{}
	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql, profile.UserID, profile.CollegeID, profile.FirstName, profile.LastName, profile.Bio, profile.ProfileImage, profile.PhoneNumber, profile.Address, profile.DateOfBirth, profile.JoinedAt, profile.LastActive, profile.Preferences, profile.SocialLinks, profile.CreatedAt, profile.UpdatedAt)
	if err != nil {
		return fmt.Errorf("CreateProfile: failed to execute query or scan ID: %w", err)
	}
	profile.ID = temp.ID
	return nil
}

func (r *profileRepository) GetProfileByUserID(ctx context.Context, userID int) (*models.Profile, error) {
	profile := &models.Profile{}
	sql := `SELECT id, user_id, college_id, first_name, last_name, bio, profile_image, phone_number, address, date_of_birth, joined_at, last_active, preferences, social_links, created_at, updated_at FROM profiles WHERE user_id = $1`
	err := pgxscan.Get(ctx, r.DB.Pool, profile, sql, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetProfileByUserID: profile for user ID %d not found", userID)
		}
		return nil, fmt.Errorf("GetProfileByUserID: failed to execute query or scan: %w", err)
	}
	return profile, nil
}

func (r *profileRepository) GetProfileByID(ctx context.Context, profileID int) (*models.Profile, error) {
	profile := &models.Profile{}
	sql := `SELECT id, user_id, college_id, first_name, last_name, bio, profile_image, phone_number, address, date_of_birth, joined_at, last_active, preferences, social_links, created_at, updated_at FROM profiles WHERE id = $1`
	err := pgxscan.Get(ctx, r.DB.Pool, profile, sql, profileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetProfileByID: profile with ID %d not found", profileID)
		}
		return nil, fmt.Errorf("GetProfileByID: failed to execute query or scan: %w", err)
	}
	return profile, nil
}

func (r *profileRepository) GetProfileByKratosID(ctx context.Context, kratosID string) (*models.Profile, error) {
	profile := &models.Profile{}
	sql := `SELECT p.id, p.user_id, p.college_id, p.first_name, p.last_name, p.bio, p.profile_image, p.phone_number, p.address, p.date_of_birth, p.joined_at, p.last_active, p.preferences, p.social_links, p.created_at, p.updated_at 
			FROM profiles p 
			JOIN users u ON p.user_id = u.id 
			WHERE u.kratos_identity_id = $1`
	err := pgxscan.Get(ctx, r.DB.Pool, profile, sql, kratosID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetProfileByKratosID: profile for kratos ID %s not found", kratosID)
		}
		return nil, fmt.Errorf("GetProfileByKratosID: failed to execute query or scan: %w", err)
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

	sql := `UPDATE profiles SET college_id = $1, first_name = $2, last_name = $3, bio = $4, profile_image = $5, phone_number = $6, address = $7, date_of_birth = $8, last_active = $9, preferences = $10, social_links = $11, updated_at = $12 WHERE id = $13`
	commandTag, err := r.DB.Pool.Exec(ctx, sql, profile.CollegeID, profile.FirstName, profile.LastName, profile.Bio, profile.ProfileImage, profile.PhoneNumber, profile.Address, profile.DateOfBirth, profile.LastActive, profile.Preferences, profile.SocialLinks, profile.UpdatedAt, profile.ID)
	if err != nil {
		return fmt.Errorf("UpdateProfile: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateProfile: no profile found with ID %d, or no changes made", profile.ID)
	}
	return nil
}

func (r *profileRepository) CreateProfileHistory(ctx context.Context, history *models.ProfileHistory) error {
	sql := `INSERT INTO profile_history (profile_id, user_id, changed_fields, old_values, new_values, changed_by, change_reason, changed_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	var id int
	err := r.DB.Pool.QueryRow(ctx, sql,
		history.ProfileID,
		history.UserID,
		history.ChangedFields,
		history.OldValues,
		history.NewValues,
		history.ChangedBy,
		history.ChangeReason,
		history.ChangedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("CreateProfileHistory: failed to execute query: %w", err)
	}

	history.ID = id
	return nil
}

func (r *profileRepository) GetProfileHistory(ctx context.Context, profileID int, limit, offset int) ([]*models.ProfileHistory, error) {
	sql := `SELECT id, profile_id, user_id, changed_fields, old_values, new_values, changed_by, change_reason, changed_at
			FROM profile_history
			WHERE profile_id = $1
			ORDER BY changed_at DESC
			LIMIT $2 OFFSET $3`

	var history []*models.ProfileHistory
	err := pgxscan.Select(ctx, r.DB.Pool, &history, sql, profileID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetProfileHistory: failed to execute query: %w", err)
	}

	return history, nil
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

func (r *profileRepository) UpdateProfilePartial(ctx context.Context, profileID int, req *models.UpdateProfileRequest) error {
	if profileID == 0 {
		return fmt.Errorf("UpdateProfilePartial: profileID is required")
	}

	if req == nil {
		return fmt.Errorf("UpdateProfilePartial: request cannot be nil")
	}

	hasUpdates := false
	if req.FirstName != nil || req.LastName != nil || req.Bio != nil || req.ProfileImage != nil || req.PhoneNumber != nil || req.Address != nil || req.DateOfBirth != nil || req.Preferences != nil || req.SocialLinks != nil {
		hasUpdates = true
	}

	if !hasUpdates {
		return fmt.Errorf("UpdateProfilePartial: at least one field must be provided for update")
	}

	var fields []string
	args := []interface{}{}

	if req.FirstName != nil {
		fields = append(fields, fmt.Sprintf("first_name = $%d", len(args)+1))
		args = append(args, *req.FirstName)
	}
	if req.LastName != nil {
		fields = append(fields, fmt.Sprintf("last_name = $%d", len(args)+1))
		args = append(args, *req.LastName)
	}
	if req.Bio != nil {
		fields = append(fields, fmt.Sprintf("bio = $%d", len(args)+1))
		args = append(args, *req.Bio)
	}
	if req.ProfileImage != nil {
		fields = append(fields, fmt.Sprintf("profile_image = $%d", len(args)+1))
		args = append(args, *req.ProfileImage)
	}
	if req.PhoneNumber != nil {
		fields = append(fields, fmt.Sprintf("phone_number = $%d", len(args)+1))
		args = append(args, *req.PhoneNumber)
	}
	if req.Address != nil {
		fields = append(fields, fmt.Sprintf("address = $%d", len(args)+1))
		args = append(args, *req.Address)
	}
	if req.DateOfBirth != nil {
		fields = append(fields, fmt.Sprintf("date_of_birth = $%d", len(args)+1))
		args = append(args, *req.DateOfBirth)
	}
	if req.Preferences != nil {
		fields = append(fields, fmt.Sprintf("preferences = $%d", len(args)+1))
		args = append(args, *req.Preferences)
	}
	if req.SocialLinks != nil {
		fields = append(fields, fmt.Sprintf("social_links = $%d", len(args)+1))
		args = append(args, *req.SocialLinks)
	}

	wherePlaceholder := fmt.Sprintf("$%d", len(args)+1)
	args = append(args, profileID)

	sql := fmt.Sprintf("UPDATE profiles SET %s, updated_at = NOW() WHERE id = %s", strings.Join(fields, ", "), wherePlaceholder)

	commandTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateProfilePartial: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateProfilePartial: no profile found with ID %d", profileID)
	}

	return nil
}
