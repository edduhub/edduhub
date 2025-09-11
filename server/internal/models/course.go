// Package models contains data structures and types used throughout the application.
// This file specifically defines course-related models including Course and UpdateCourseRequest.
//
// These models are used for:
// - Database operations (via struct tags)
// - JSON serialization/deserialization (via json tags)
// - Input validation (via validate tags)
// - API documentation and responses
package models

import "time"

// Course represents an individual course/subject in the educational system.
// It contains all the essential information about a course including its metadata,
// instructor, and relationships to other entities.
type Course struct {
	// Primary identifier for the course
	ID int `db:"id" json:"id" validate:"omitempty,gte=0"`

	// Course name/title (required, 3-100 characters)
	Name string `db:"name" json:"name" validate:"required,min=3,max=100"`

	// ID of the college this course belongs to (required, must be positive)
	CollegeID int `db:"college_id" json:"college_id" validate:"required,gte=1"`

	// Optional description of the course (max 200 characters)
	Description string `db:"description" json:"description" validate:"omitempty,max=200"`

	// Number of credits for this course (required, 1-5 credits)
	Credits int `db:"credits" json:"credits" validate:"required,gte=1,lte=5"`

	// ID of the instructor teaching this course (required, must be positive)
	InstructorID int `db:"instructor_id" json:"instructor_id" validate:"required,gte=1"`

	// Timestamps for audit trail
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Relations - not stored in DB but populated for API responses
	Instructor  *User         `db:"-" json:"instructor,omitempty"`  // Instructor details
	Enrollments []*Enrollment `db:"-" json:"enrollments,omitempty"` // Student enrollments
}

// UpdateCourseRequest provides fields for partial updates to Course via PATCH.
// All fields are optional and only provided fields will be updated.
type UpdateCourseRequest struct {
	// Optional course name update (3-100 characters)
	Name *string `json:"name" validate:"omitempty,min=3,max=100"`

	// Optional college ID update (must be positive)
	CollegeID *int `json:"college_id" validate:"omitempty,gte=1"`

	// Optional description update (max 200 characters)
	Description *string `json:"description" validate:"omitempty,max=200"`

	// Optional credits update (1-5 credits)
	Credits *int `json:"credits" validate:"omitempty,gte=1,lte=5"`

	// Optional instructor ID update (must be positive)
	InstructorID *int `json:"instructor_id" validate:"omitempty,gte=1"`
}