package models

import "time"

type College struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Address   string    `db:"address" json:"address"`
	City      string    `db:"city" json:"city"`
	State     string    `db:"state" json:"state"`
	Country   string    `db:"country" json:"country"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Relations - not stored in DB
	Students []*Student `db:"-" json:"students,omitempty"`
}

// UpdateCollegeRequest provides fields for partial updates to College via PATCH
type UpdateCollegeRequest struct {
	Name    *string `json:"name" validate:"omitempty,min=1,max=100"`
	Address *string `json:"address" validate:"omitempty"`
	City    *string `json:"city" validate:"omitempty,min=1,max=50"`
	State   *string `json:"state" validate:"omitempty,min=1,max=50"`
	Country *string `json:"country" validate:"omitempty,min=1,max=50"`
}

// CollegeStats represents aggregated statistics for a college
type CollegeStats struct {
	CollegeID        int     `json:"college_id"`
	TotalStudents    int     `json:"total_students"`
	ActiveStudents   int     `json:"active_students"`
	TotalCourses     int     `json:"total_courses"`
	TotalDepartments int     `json:"total_departments"`
	TotalEnrollments int     `json:"total_enrollments"`
	AttendanceRate   float64 `json:"attendance_rate"`
	AverageGrade     float64 `json:"average_grade"`
	TotalFaculties   int     `json:"total_faculties"`
	PendingFees      float64 `json:"pending_fees"`
	PlacementRate    float64 `json:"placement_rate"`
	CreatedAt        string  `json:"created_at"`
}
