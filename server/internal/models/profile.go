package models

import (
	"time"
)

type Profile struct {
	ID           int       `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`             // Reference to Kratos identity ID
	CollegeID    string    `json:"college_id" db:"college_id"`       // College affiliation
	Bio          string    `json:"bio" db:"bio"`                     // Short biography
	ProfileImage string    `json:"profile_image" db:"profile_image"` // URL to profile picture
	PhoneNumber  string    `json:"phone_number" db:"phone_number"`   // Contact information
	Address      string    `json:"address" db:"address"`             // Physical address
	DateOfBirth  time.Time `json:"date_of_birth" db:"date_of_birth"` // DOB for age verification and birthday notifications
	JoinedAt     time.Time `json:"joined_at" db:"joined_at"`         // When they joined the platform
	LastActive   time.Time `json:"last_active" db:"last_active"`     // Last activity timestamp
	Preferences  JSONMap   `json:"preferences" db:"preferences"`     // User preferences (notifications, UI settings, etc.)
	SocialLinks  JSONMap   `json:"social_links" db:"social_links"`   // LinkedIn, GitHub, etc.
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// JSONMap is a helper type for storing JSON data
type JSONMap map[string]interface{}

// UpdateProfileRequest provides fields for partial updates to Profile via PATCH
type UpdateProfileRequest struct {
	UserID       *string             `json:"user_id" validate:"omitempty,len=36"`
	CollegeID    *string             `json:"college_id" validate:"omitempty,len=36"`
	Bio          *string             `json:"bio" validate:"omitempty,max=500"`
	ProfileImage *string             `json:"profile_image" validate:"omitempty,url"`
	PhoneNumber  *string             `json:"phone_number" validate:"omitempty,e164"`
	Address      *string             `json:"address" validate:"omitempty,max=250"`
	DateOfBirth  *time.Time          `json:"date_of_birth" validate:"omitempty"`
	Preferences  *JSONMap            `json:"preferences"`
	SocialLinks  *JSONMap            `json:"social_links"`
}

// ProfileHistory represents a change log entry for profile updates
type ProfileHistory struct {
	ID         int       `json:"id" db:"id"`
	ProfileID  int       `json:"profile_id" db:"profile_id"`
	UserID     int       `json:"user_id" db:"user_id"`
	Action     string    `json:"action" db:"action"` // UPDATE, UPLOAD_IMAGE, etc.
	Field      string    `json:"field" db:"field"`   // field that was changed
	OldValue   string    `json:"old_value" db:"old_value"`
	NewValue   string    `json:"new_value" db:"new_value"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	UserAgent  string    `json:"user_agent" db:"user_agent"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// UploadProfileImageRequest for profile picture upload
type UploadProfileImageRequest struct {
	FileType string `json:"file_type" validate:"required,oneof=jpeg png jpg gif"`
}