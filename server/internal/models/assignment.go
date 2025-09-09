package models

import (
	"time"
)

// Assignment represents an assignment given in a course.
type Assignment struct {
	ID          int       `db:"id" json:"id" validate:"omitempty,gte=0"`        // Primary Key
	CourseID    int       `db:"course_id" json:"course_id" validate:"required"` // Foreign key to courses table
	CollegeID   int       `db:"college_id" json:"college_id"`                   // Denormalized, Foreign key to colleges table
	Title       string    `db:"title" json:"title"`                             // Title of the assignment
	Description string    `db:"description" json:"description,omitempty"`       // Detailed description
	DueDate     time.Time `db:"due_date" json:"due_date"`                       // Due date for the assignment
	MaxPoints   int       `db:"max_points" json:"max_points"`                   // Maximum points for the assignment
	CreatedAt   time.Time `db:"created_at" json:"created_at"`                   // Timestamp of creation
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`                   // Timestamp of last update
}

type UpdateAssignmentRequest struct {
	ID          *int
	CourseID    *int
	CollegeID   *int
	Title       *string
	Description *string
	DueDate     *time.Time
	MaxPoints   *int
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

// AssignmentSubmission represents a student's submission for an assignment.
type AssignmentSubmission struct {
	ID             int       `db:"id" json:"id" validate:"omitempty,gte=0"`    // Primary Key
	AssignmentID   int       `db:"assignment_id" json:"assignment_id"`         // Foreign key to assignments table
	StudentID      int       `db:"student_id" json:"student_id"`               // Foreign key to students table
	SubmissionTime time.Time `db:"submission_time" json:"submission_time"`     // Timestamp of submission
	ContentText    *string   `db:"content_text" json:"content_text,omitempty"` // Text content of submission, nullable
	FilePath       *string   `db:"file_path" json:"file_path,omitempty"`       // Path to submitted file, nullable
	Grade          *int      `db:"grade" json:"grade,omitempty"`               // Grade awarded, nullable
	Feedback       *string   `db:"feedback" json:"feedback,omitempty"`         // Feedback from instructor, nullable
	CreatedAt      time.Time `db:"created_at" json:"created_at"`               // Timestamp of creation
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`               // Timestamp of last update
}
