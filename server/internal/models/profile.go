package models

import (
	"time"
)

type Profile struct {
	ID           int        `json:"id" db:"id"`
	UserID       int        `json:"user_id" db:"user_id"`
	CollegeID    int        `json:"college_id" db:"college_id"`
	FirstName    string     `json:"first_name" db:"first_name"`
	LastName     string     `json:"last_name" db:"last_name"`
	Bio          string     `json:"bio" db:"bio"`
	ProfileImage string     `json:"profile_image" db:"profile_image"`
	PhoneNumber  string     `json:"phone_number" db:"phone_number"`
	Address      string     `json:"address" db:"address"`
	DateOfBirth  *time.Time `json:"date_of_birth" db:"date_of_birth"`
	JoinedAt     time.Time  `json:"joined_at" db:"joined_at"`
	LastActive   time.Time  `json:"last_active" db:"last_active"`
	Preferences  JSONMap    `json:"preferences" db:"preferences"`
	SocialLinks  JSONMap    `json:"social_links" db:"social_links"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// JSONMap is a helper type for storing JSON data
type JSONMap map[string]any

// UpdateProfileRequest provides fields for partial updates to Profile via PATCH
type UpdateProfileRequest struct {
	FirstName    *string    `json:"first_name" validate:"omitempty,max=100"`
	LastName     *string    `json:"last_name" validate:"omitempty,max=100"`
	Bio          *string    `json:"bio" validate:"omitempty,max=500"`
	ProfileImage *string    `json:"profile_image" validate:"omitempty,url"`
	PhoneNumber  *string    `json:"phone_number" validate:"omitempty,max=20"`
	Address      *string    `json:"address" validate:"omitempty,max=250"`
	DateOfBirth  *time.Time `json:"date_of_birth" validate:"omitempty"`
	Preferences  *JSONMap   `json:"preferences"`
	SocialLinks  *JSONMap   `json:"social_links"`
}

// ProfileHistory represents a change log entry for profile updates
type ProfileHistory struct {
	ID            int       `json:"id" db:"id"`
	ProfileID     int       `json:"profile_id" db:"profile_id"`
	UserID        int       `json:"user_id" db:"user_id"`
	ChangedFields JSONMap   `json:"changed_fields" db:"changed_fields"`
	OldValues     *JSONMap  `json:"old_values" db:"old_values"`
	NewValues     *JSONMap  `json:"new_values" db:"new_values"`
	ChangedBy     *int      `json:"changed_by" db:"changed_by"`
	ChangeReason  *string   `json:"change_reason" db:"change_reason"`
	ChangedAt     time.Time `json:"changed_at" db:"changed_at"`
}

// UploadProfileImageRequest for profile picture upload
type UploadProfileImageRequest struct {
	FileType string `json:"file_type" validate:"required,oneof=jpeg png jpg gif"`
}
