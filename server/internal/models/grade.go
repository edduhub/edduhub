// /home/tgt/Desktop/edduhub/server/internal/models/grade.go (Example, create or adjust as needed)
package models

import "time"

type ExamType = string

const (
	Midterm1    ExamType = "midterm-1"
	Midterm2    ExamType = "midterm-2"
	Final       ExamType = "final"
	Assignments ExamType = "assignment"
)

type Grade struct {
	ID            int       `db:"id" json:"id"`
	StudentID     string    `db:"student_id" json:"student_id"` // Kratos ID or internal student identifier
	CourseID      int       `db:"course_id" json:"course_id"`
	CollegeID     int       `db:"college_id" json:"college_id"`
	MarksObtained float64   `db:"marks_obtained" json:"marks_obtained"`
	TotalMarks    float64   `db:"total_marks" json:"total_marks"`
	GradeLetter   *string   `db:"grade_letter" json:"grade_letter,omitempty"`
	Semester      int       `db:"semester" json:"semester"`
	AcademicYear  string    `db:"academic_year" json:"academic_year"` // e.g., "2023-2024"
	ExamType      ExamType  `db:"exam_type" json:"exam_type"`         // e.g., "Midterm", "Final", "Assignment"
	GradedAt      time.Time `db:"graded_at" json:"graded_at"`         // When the grade was officially given
	Comments      *string   `db:"comments" json:"comments,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// GradeFilter can be used for querying lists of grades with specific criteria
type GradeFilter struct {
	StudentID    *string  `json:"student_id,omitempty"`
	CourseID     *int     `json:"course_id,omitempty"`
	CollegeID    *int     `json:"college_id,omitempty"` // Essential for multi-tenancy
	Semester     *int     `json:"semester,omitempty"`
	AcademicYear *string  `json:"academic_year,omitempty"`
	ExamType     ExamType `json:"exam_type,omitempty"`
	// Add pagination fields if needed
	Limit  uint64 `json:"limit,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
}

// UpdateGradeRequest provides fields for partial updates to Grade via PATCH
type UpdateGradeRequest struct {
	StudentID      *string    `json:"student_id" validate:"omitempty"`
	CourseID       *int       `json:"course_id" validate:"omitempty,gte=1"`
	CollegeID      *int       `json:"college_id" validate:"omitempty,gte=1"`
	MarksObtained  *float64   `json:"marks_obtained" validate:"omitempty,gte=0"`
	TotalMarks     *float64   `json:"total_marks" validate:"omitempty,gte=0"`
	GradeLetter    *string    `json:"grade_letter" validate:"omitempty,min=1,max=5"`
	Semester       *int       `json:"semester" validate:"omitempty,gte=1,lte=8"`
	AcademicYear   *string    `json:"academic_year" validate:"omitempty,len=9"`
	ExamType       *ExamType  `json:"exam_type" validate:"omitempty"`
	GradedAt       *time.Time `json:"graded_at" validate:"omitempty"`
	Comments       *string    `json:"comments" validate:"omitempty,max=500"`
}