// Package models contains data structures for lecture-related entities.
// This file defines the Lecture model and its associated update request structure.
//
// Lectures represent individual class sessions within courses and contain
// scheduling information, content details, and meeting links for online sessions.
package models

import "time"

// Lecture represents an individual class session within a course.
// It contains scheduling information and metadata for a specific lecture.
type Lecture struct {
	// Primary identifier for the lecture
	ID int `db:"id" json:"id" validate:"omitempty,gte=0"`

	// ID of the course this lecture belongs to (required, must be positive)
	CourseID int `db:"course_id" json:"course_id" validate:"required,gte=1"`

	// ID of the college this lecture belongs to (denormalized for performance)
	CollegeID int `db:"college_id" json:"college_id" validate:"required,gte=1"`

	// Title of the lecture (required, 3-100 characters)
	Title string `db:"title" json:"title" validate:"required,min=3,max=100"`

	// Optional description of the lecture content (max 200 characters)
	Description string `db:"description" json:"description,omitempty" validate:"omitempty,max=200"`

	// Start time of the lecture (required)
	StartTime time.Time `db:"start_time" json:"start_time" validate:"required"`

	// End time of the lecture (required, must be after start time)
	EndTime time.Time `db:"end_time" json:"end_time" validate:"required,gtfield=StartTime"`

	// Optional meeting link for online lectures (must be valid URL if provided)
	MeetingLink string `db:"meeting_link" json:"meeting_link,omitempty" validate:"omitempty,url"`

	// Timestamps for audit trail
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// UpdateLectureRequest provides fields for partial updates to Lecture via PATCH
type UpdateLectureRequest struct {
	// Optional course ID to move lecture to different course
	CourseID *int `json:"course_id" validate:"omitempty,gte=1"`

	// Optional college ID (usually changed with course)
	CollegeID *int `json:"college_id" validate:"omitempty,gte=1"`

	// Optional title update
	Title *string `json:"title" validate:"omitempty,min=3,max=100"`

	// Optional description update
	Description *string `json:"description" validate:"omitempty,max=200"`

	// Optional start time update
	StartTime *time.Time `json:"start_time" validate:"omitempty"`

	// Optional end time update
	EndTime *time.Time `json:"end_time" validate:"omitempty"`

	// Optional meeting link update
	MeetingLink *string `json:"meeting_link" validate:"omitempty,url"`
}

// CreateLectureRequest provides fields for creating a new Lecture
type CreateLectureRequest struct {
	Title       string    `json:"title" validate:"required,min=3,max=100"`
	Description string    `json:"description" validate:"omitempty,max=200"`
	StartTime   time.Time `json:"start_time" validate:"required"`
	EndTime     time.Time `json:"end_time" validate:"required"`
	MeetingLink string    `json:"meeting_link" validate:"omitempty,url"`
}