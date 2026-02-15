package models

import "time"

// GradingRubric represents a rubric template/definition owned by a faculty user.
type GradingRubric struct {
	ID          int               `db:"id" json:"id"`
	FacultyID   int               `db:"faculty_id" json:"faculty_id"`
	CollegeID   int               `db:"college_id" json:"college_id"`
	Name        string            `db:"name" json:"name"`
	Description *string           `db:"description" json:"description,omitempty"`
	CourseID    *int              `db:"course_id" json:"course_id,omitempty"`
	IsTemplate  bool              `db:"is_template" json:"is_template"`
	IsActive    bool              `db:"is_active" json:"is_active"`
	MaxScore    int               `db:"max_score" json:"max_score"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at" json:"updated_at"`
	Criteria    []RubricCriterion `db:"-" json:"criteria,omitempty"`
}

// RubricCriterion is a criterion associated with a rubric.
type RubricCriterion struct {
	ID          int       `db:"id" json:"id"`
	RubricID    int       `db:"rubric_id" json:"rubric_id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description,omitempty"`
	Weight      float64   `db:"weight" json:"weight"`
	MaxScore    int       `db:"max_score" json:"max_score"`
	SortOrder   int       `db:"sort_order" json:"sort_order"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// OfficeHourSlot represents a faculty availability slot.
type OfficeHourSlot struct {
	ID         int       `db:"id" json:"id"`
	FacultyID  int       `db:"faculty_id" json:"faculty_id"`
	CollegeID  int       `db:"college_id" json:"college_id"`
	DayOfWeek  int       `db:"day_of_week" json:"day_of_week"`
	StartTime  string    `db:"start_time" json:"start_time"`
	EndTime    string    `db:"end_time" json:"end_time"`
	Location   *string   `db:"location" json:"location,omitempty"`
	IsVirtual  bool      `db:"is_virtual" json:"is_virtual"`
	VirtualLink *string  `db:"virtual_link" json:"virtual_link,omitempty"`
	MaxStudents int      `db:"max_students" json:"max_students"`
	IsActive    bool     `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	FacultyName *string   `db:"faculty_name" json:"faculty_name,omitempty"`
}

// OfficeHourBooking represents a booking made by a student.
type OfficeHourBooking struct {
	ID          int       `db:"id" json:"id"`
	OfficeHourID int      `db:"office_hour_id" json:"office_hour_id"`
	StudentID   int       `db:"student_id" json:"student_id"`
	CollegeID   int       `db:"college_id" json:"college_id"`
	BookingDate time.Time `db:"booking_date" json:"booking_date"`
	StartTime   string    `db:"start_time" json:"start_time"`
	EndTime     string    `db:"end_time" json:"end_time"`
	Purpose     *string   `db:"purpose" json:"purpose,omitempty"`
	Status      string    `db:"status" json:"status"`
	Notes       *string   `db:"notes" json:"notes,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	OfficeHour  *OfficeHourSlot `db:"-" json:"office_hour,omitempty"`
}
