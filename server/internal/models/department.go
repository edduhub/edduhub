package models

import "time"

type Department struct {
	ID        int       `db:"id" json:"id"`
	CollegeID int       `db:"college_id" json:"college_id"` // Foreign key to colleges table
	Name      string    `db:"name" json:"name"`
	HOD       string    `db:"hod" json:"hod"` // Head of Department (could be a user ID in the future)
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type UpdateDepartmentRequest struct {
	ID        *int    `db:"id" validate:"omitempty,gte=0"`
	CollegeID *int    `db:"college_id" validate:"omitempty,gte=0"`
	Name      *string `db:"name" validate:"omitempty,min=4,max=40"`
	HOD       *string  `db:"hod" validate:"omitempty,min=3,max=30"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}
