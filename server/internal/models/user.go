// filepath: /home/tgt/Desktop/edduhub/server/internal/models/user.go
package models

import "time"

type User struct {
	ID               int    `db:"id" json:"id"`
	Name             string `db:"name" json:"name"`
	Role             string `db:"role" json:"role"`
	Email            string `db:"email" json:"email"`
	KratosIdentityID string `db:"kratos_identity_id" json:"kratos_identity_id"`
	IsActive         bool   `db:"is_active" json:"is_active"`
	// RollNo           string    `db:"roll_no" json:"roll_no"` // Removed from User model
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Relations - not stored in DB
	Student *Student `db:"-" json:"student,omitempty"`
}

// UpdateUserRequest provides fields for partial updates to User via PATCH
type UpdateUserRequest struct {
	Name             *string `json:"name" validate:"omitempty,min=1,max=100"`
	Role             *string `json:"role" validate:"omitempty,min=1,max=50"`
	Email            *string `json:"email" validate:"omitempty,email"`
	KratosIdentityID *string `json:"kratos_identity_id" validate:"omitempty,len=36"`
	IsActive         *bool   `json:"is_active" validate:"omitempty"`
}