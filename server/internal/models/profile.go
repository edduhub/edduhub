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
	PhoneNumber  *string             `json:"phone_number" validate:"omitempty,len=10"`
	Address      *string             `json:"address" validate:"omitempty,max=250"`
	DateOfBirth  *time.Time          `json:"date_of_birth" validate:"omitempty"`
	Preferences  *JSONMap            `json:"preferences"`
	SocialLinks  *JSONMap            `json:"social_links"`
}