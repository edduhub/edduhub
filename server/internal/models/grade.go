package models

import "time"

// Grade represents an assessment record stored in the grades table.
type Grade struct {
	ID             int        `db:"id" json:"id"`
	StudentID      int        `db:"student_id" json:"student_id"`
	CourseID       int        `db:"course_id" json:"course_id"`
	CollegeID      int        `db:"college_id" json:"college_id"`
	AssessmentName string     `db:"assessment_name" json:"assessment_name"`
	AssessmentType string     `db:"assessment_type" json:"assessment_type"`
	TotalMarks     int        `db:"total_marks" json:"total_marks"`
	ObtainedMarks  int        `db:"obtained_marks" json:"obtained_marks"`
	Percentage     float64    `db:"percentage" json:"percentage"`
	Grade          *string    `db:"grade" json:"grade,omitempty"`
	Remarks        *string    `db:"remarks" json:"remarks,omitempty"`
	GradedBy       *string    `db:"graded_by" json:"graded_by,omitempty"`
	GradedAt       *time.Time `db:"graded_at" json:"graded_at,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

// GradeFilter can be used for querying lists of grades with specific criteria.
type GradeFilter struct {
	StudentID      *int    `json:"student_id,omitempty"`
	CourseID       *int    `json:"course_id,omitempty"`
	CollegeID      *int    `json:"college_id,omitempty"`
	AssessmentType *string `json:"assessment_type,omitempty"`
	Limit          uint64  `json:"limit,omitempty"`
	Offset         uint64  `json:"offset,omitempty"`
}

// UpdateGradeRequest provides fields for partial updates to Grade via PATCH.
type UpdateGradeRequest struct {
	StudentID      *int       `json:"student_id" validate:"omitempty,gte=1"`
	CourseID       *int       `json:"course_id" validate:"omitempty,gte=1"`
	CollegeID      *int       `json:"college_id" validate:"omitempty,gte=1"`
	AssessmentName *string    `json:"assessment_name" validate:"omitempty,min=1,max=200"`
	AssessmentType *string    `json:"assessment_type" validate:"omitempty,min=1,max=50"`
	TotalMarks     *int       `json:"total_marks" validate:"omitempty,gte=0"`
	ObtainedMarks  *int       `json:"obtained_marks" validate:"omitempty,gte=0"`
	Percentage     *float64   `json:"percentage" validate:"omitempty,gte=0,lte=100"`
	Grade          *string    `json:"grade" validate:"omitempty,min=1,max=5"`
	Remarks        *string    `json:"remarks" validate:"omitempty,max=500"`
	GradedBy       *string    `json:"graded_by" validate:"omitempty,max=255"`
	GradedAt       *time.Time `json:"graded_at" validate:"omitempty"`
}
