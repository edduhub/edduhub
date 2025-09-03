package models

import "time"

// Course represents an individual course/subjectpackage models

// Course represents an individual course/subject
type Course struct {
	ID           int       `db:"id" json:"id" validate:"omitempty,gte=0"`
	Name         string    `db:"name" json:"name" validate:"required,min=3,max=100"`
	CollegeID    int       `db:"college_id" json:"college_id" validate:"required,gte=1"`
	Description  string    `db:"description" json:"description" validate:"omitempty,max=200"`
	Credits      int       `db:"credits" json:"credits" validate:"required,gte=1,lte=5"`
	InstructorID int       `db:"instructor_id" json:"instructor_id" validate:"required,gte=1"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`

	// Relations - not stored in DB
	Instructor  *User         `db:"-" json:"instructor,omitempty"`
	Enrollments []*Enrollment `db:"-" json:"enrollments,omitempty"`
}

// Lecture represents an individual class session
type Lecture struct {
	ID          int       `db:"id" json:"id" validate:"omitempty,gte=0"`                               // Primary Key
	CourseID    int       `db:"course_id" json:"course_id" validate:"required,gte=1"`                  // Foreign key to courses table
	CollegeID   int       `db:"college_id" json:"college_id" validate:"required,gte=1"`                // Denormalized, Foreign key to colleges table
	Title       string    `db:"title" json:"title" validate:"required,min=3,max=100"`                  // Title of the lecture
	Description string    `db:"description" json:"description,omitempty" validate:"omitempty,max=200"` // Optional description
	StartTime   time.Time `db:"start_time" json:"start_time" validate:"required,before_end_time"`      // Start time of the lecture
	EndTime     time.Time `db:"end_time" json:"end_time" validate:"required,after_start_time"`         // End time of the lecture
	MeetingLink string    `db:"meeting_link" json:"meeting_link,omitempty" validate:"omitempty,url"`   // For online lectures
	CreatedAt   time.Time `db:"created_at" json:"created_at"`                                          // Timestamp of creation
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`                                          // Timestamp of last update
}

// QRCode represents a unique QR code for each lecture

// Courses represents a collection of courses
type Courses struct {
	Items []*Course // Fixed syntax
}

// Subjects represents the courses a student is enrolled in
type Subjects struct {
	Current  Courses `json:"current"`  // Currently enrolled courses
	Previous Courses `json:"previous"` // Previously completed courses
	Optional Courses `json:"optional"` // Optional/elective courses
}

// UpdateCourseRequest provides fields for partial updates to Course via PATCH
type UpdateCourseRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=3,max=100"`
	CollegeID   *int    `json:"college_id" validate:"omitempty,gte=1"`
	Description *string `json:"description" validate:"omitempty,max=200"`
	Credits     *int    `json:"credits" validate:"omitempty,gte=1,lte=5"`
	InstructorID *int   `json:"instructor_id" validate:"omitempty,gte=1"`
}

// UpdateLectureRequest provides fields for partial updates to Lecture via PATCH
type UpdateLectureRequest struct {
	CourseID    *int        `json:"course_id" validate:"omitempty,gte=1"`
	CollegeID   *int        `json:"college_id" validate:"omitempty,gte=1"`
	Title       *string     `json:"title" validate:"omitempty,min=3,max=100"`
	Description *string     `json:"description" validate:"omitempty,max=200"`
	StartTime   *time.Time  `json:"start_time" validate:"omitempty"`
	EndTime     *time.Time  `json:"end_time" validate:"omitempty"`
	MeetingLink *string     `json:"meeting_link" validate:"omitempty,url"`
}