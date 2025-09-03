package models

import "time"

type Student struct {
	StudentID        int       `db:"id" json:"student_id"`
	UserID           int       `db:"user_id" json:"user_id"`
	CollegeID        int       `db:"college_id" json:"college_id"`
	KratosIdentityID string    `db:"kratos_identity_id" json:"kratos_identity_id"`
	EnrollmentYear   int       `db:"enrollment_year" json:"enrollment_year"`
	RollNo           string    `db:"roll_no" json:"roll_no"`     // Added this field
	IsActive         bool      `db:"is_active" json:"is_active"` // Added this field
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`

	// Relations - not stored in DB (add db:"-" tag)
	// College     *College      `db:"-" json:"college,omitempty"`
	// Enrollments []*Enrollment `db:"-" json:"enrollments,omitempty"`
	// QRCodes     []*QRCode     `db:"-" json:"qr_codes,omitempty"`
}

// UpdateStudentRequest provides fields for partial updates to Student via PATCH
type UpdateStudentRequest struct {
	UserID *int `json:"user_id" validate:"omitempty,gte=1"`
	CollegeID *int `json:"college_id" validate:"omitempty,gte=1"`
	EnrollmentYear *int `json:"enrollment_year" validate:"omitempty,gte=1947"`
	RollNo *string `json:"roll_no" validate:"omitempty,min=1,max=50"`
	IsActive *bool `json:"is_active" validate:"omitempty"`
}