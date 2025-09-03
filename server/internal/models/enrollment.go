package models

import "time"

type EnrollmentStatus = string

const (
	Active    EnrollmentStatus = "active"
	Inactive  EnrollmentStatus = "inactive"
	Completed EnrollmentStatus = "completed"
)

type Enrollment struct {
	ID             int              `db:"id" json:"id"`
	StudentID      int              `db:"student_id" json:"student_id"`
	CourseID       int              `db:"course_id" json:"course_id"`
	CollegeID      int              `db:"college_id" json:"college_id"`
	EnrollmentDate time.Time        `db:"enrollment_date" json:"enrollment_date"`
	Status         EnrollmentStatus `db:"status" json:"status"` // Active, Completed, Dropped
	Grade          string           `db:"grade" json:"grade,omitempty"`
	CreatedAt      time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time        `db:"updated_at" json:"updated_at"`

	// Relations - not stored in DB
	Student *Student `db:"-" json:"student,omitempty"`
	Course  *Course  `db:"-" json:"course,omitempty"`
}

// UpdateEnrollmentRequest provides fields for partial updates to Enrollment via PATCH
type UpdateEnrollmentRequest struct {
	StudentID      *int              `json:"student_id" validate:"omitempty,gte=1"`
	CourseID       *int              `json:"course_id" validate:"omitempty,gte=1"`
	CollegeID      *int              `json:"college_id" validate:"omitempty,gte=1"`
	EnrollmentDate *time.Time        `json:"enrollment_date" validate:"omitempty"`
	Status         *EnrollmentStatus `json:"status" validate:"omitempty,oneof=active inactive completed"`
	Grade          *string           `json:"grade" validate:"omitempty,max=5"`
}