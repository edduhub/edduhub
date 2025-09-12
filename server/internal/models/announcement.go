// Package models contains data structures and types used throughout the application.
// This file specifically defines the Announcement model.
package models

import "time"

// Announcement represents a message or notice broadcast to users within a college.
// It can be a general announcement for the entire college or targeted to specific
// groups if extended with more specific associations.
type Announcement struct {
	// Primary identifier for the announcement
	ID int `db:"id" json:"id" validate:"omitempty,gte=0"`

	// Title of the announcement (required, 5-150 characters)
	Title string `db:"title" json:"title" validate:"required,min=5,max=150"`

	// Full content of the announcement (required, min 10 characters)
	Content string `db:"content" json:"content" validate:"required,min=10"`

	// ID of the college this announcement belongs to (required, must be positive)
	CollegeID int `db:"college_id" json:"college_id" validate:"required,gte=1"`

	// ID of the user (author) who created this announcement (required, must be positive)
	UserID int `db:"user_id" json:"user_id" validate:"required,gte=1"`

	// Timestamps for audit trail
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Relations - not stored in DB but can be populated for API responses
	Author *User `db:"-" json:"author,omitempty"` // Author details
}

// CreateAnnouncementRequest defines the payload for creating a new announcement.
// It includes all the required fields to construct a valid Announcement object.
type CreateAnnouncementRequest struct {
	// Title of the announcement (required, 5-150 characters)
	Title string `json:"title" validate:"required,min=5,max=150"`

	// Full content of the announcement (required, min 10 characters)
	Content string `json:"content" json:"content" validate:"required,min=10"`
}

// UpdateAnnouncementRequest provides fields for partial updates to an Announcement.
// All fields are optional, and only provided fields will be updated.
type UpdateAnnouncementRequest struct {
	// Optional title update (5-150 characters)
	Title *string `json:"title" validate:"omitempty,min=5,max=150"`

	// Optional content update (min 10 characters)
	Content *string `json:"content" validate:"omitempty,min=10"`
}
