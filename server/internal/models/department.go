package models

import "time"

type Department struct {
	ID           int       `db:"id" json:"id"`
	CollegeID    int       `db:"college_id" json:"college_id"`
	Name         string    `db:"name" json:"name"`
	Code         string    `db:"code" json:"code"`
	Description  *string   `db:"description" json:"description,omitempty"`
	HeadUserID   *int      `db:"head_user_id" json:"head_user_id,omitempty"`
	HOD          string    `db:"hod" json:"hod,omitempty"`
	HODName      string    `db:"hod_name" json:"hodName,omitempty"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	StudentCount int       `db:"student_count" json:"studentCount,omitempty"`
	FacultyCount int       `db:"faculty_count" json:"facultyCount,omitempty"`
	CoursesCount int       `db:"courses_count" json:"coursesCount,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type UpdateDepartmentRequest struct {
	ID          *int       `json:"id" validate:"omitempty,gte=1"`
	CollegeID   *int       `json:"college_id" validate:"omitempty,gte=1"`
	Name        *string    `json:"name" validate:"omitempty,min=2,max=200"`
	Code        *string    `json:"code" validate:"omitempty,min=2,max=20"`
	Description *string    `json:"description" validate:"omitempty,max=1000"`
	HeadUserID  *int       `json:"head_user_id" validate:"omitempty,gte=1"`
	HOD         *string    `json:"hod" validate:"omitempty,min=2,max=255"`
	IsActive    *bool      `json:"is_active"`
	CreatedAt   *time.Time `json:"-"`
	UpdatedAt   *time.Time `json:"-"`
}
