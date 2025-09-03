package models

import (
	"time"

	"github.com/jackc/pgtype" // For PostgreSQL TIME type
)

type TimeTableBlock struct {
	ID           int          `db:"id" json:"id"`
	CollegeID    int          `db:"college_id" json:"college_id"` // For multi-tenancy
	DepartmentID *int         `db:"department_id" json:"department_id,omitempty"`
	CourseID     int          `db:"course_id" json:"course_id"`
	ClassID      *int         `db:"class_id" json:"class_id,omitempty"` // e.g., "Section A", "Batch 1" - could be FK to a 'classes' table
	DayOfWeek    time.Weekday `db:"day_of_week" json:"day_of_week"`     // e.g., time.Monday, time.Tuesday
	StartTime    pgtype.Time  `db:"start_time" json:"start_time"`       // Represents HH:MM:SS
	EndTime      pgtype.Time  `db:"end_time" json:"end_time"`           // Represents HH:MM:SS
	RoomNumber   *string      `db:"room_number" json:"room_number,omitempty"`
	FacultyID    *string      `db:"faculty_id" json:"faculty_id,omitempty"` // Kratos ID or internal faculty ID
	// Optional: Add fields for effective date ranges if schedules change mid-semester
	// EffectiveStartDate time.Time `db:"effective_start_date" json:"effective_start_date"`
	// EffectiveEndDate   *time.Time `db:"effective_end_date" json:"effective_end_date,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// TimeTableBlockFilter can be used for querying lists of timetable blocks
type TimeTableBlockFilter struct {
	CollegeID    int           `json:"college_id"` // Mandatory
	DepartmentID *int          `json:"department_id,omitempty"`
	CourseID     *int          `json:"course_id,omitempty"`
	ClassID      *int          `json:"class_id,omitempty"`
	DayOfWeek    *time.Weekday `json:"day_of_week,omitempty"`
	InstructorID *string       `json:"instructor_id,omitempty"`
	StartTime    *pgtype.Time  `json:"start_time,omitempty"`
	EndTime      *pgtype.Time  `json:"end_time,omitempty"`
	Limit        uint64        `json:"limit,omitempty"`
	Offset       uint64        `json:"offset,omitempty"`
}
