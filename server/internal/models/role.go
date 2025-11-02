package models

import "time"

// Role represents a role in the system
type Role struct {
	ID           int       `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Description  *string   `db:"description" json:"description,omitempty"`
	CollegeID    *int      `db:"college_id" json:"college_id,omitempty"`
	IsSystemRole bool      `db:"is_system_role" json:"is_system_role"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`

	// Relations - not stored in DB
	Permissions []Permission `db:"-" json:"permissions,omitempty"`
}

// Permission represents a permission in the system
type Permission struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Resource    string    `db:"resource" json:"resource"`
	Action      string    `db:"action" json:"action"`
	Description *string   `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// RolePermission represents the junction table for roles and permissions
type RolePermission struct {
	ID           int       `db:"id" json:"id"`
	RoleID       int       `db:"role_id" json:"role_id"`
	PermissionID int       `db:"permission_id" json:"permission_id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// UserRoleAssignment represents the assignment of a role to a user
type UserRoleAssignment struct {
	ID         int        `db:"id" json:"id"`
	UserID     int        `db:"user_id" json:"user_id"`
	RoleID     int        `db:"role_id" json:"role_id"`
	AssignedBy *int       `db:"assigned_by" json:"assigned_by,omitempty"`
	AssignedAt time.Time  `db:"assigned_at" json:"assigned_at"`
	ExpiresAt  *time.Time `db:"expires_at" json:"expires_at,omitempty"`

	// Relations - not stored in DB
	Role *Role `db:"-" json:"role,omitempty"`
}

// CreateRoleRequest represents a request to create a new role
type CreateRoleRequest struct {
	Name        string   `json:"name" validate:"required,minlen=2,maxlen=100"`
	Description *string  `json:"description" validate:"omitempty,maxlen=500"`
	CollegeID   *int     `json:"college_id" validate:"omitempty"`
	Permissions []int    `json:"permissions" validate:"omitempty"` // Permission IDs
}

// UpdateRoleRequest represents a request to update a role
type UpdateRoleRequest struct {
	Name        *string  `json:"name" validate:"omitempty,minlen=2,maxlen=100"`
	Description *string  `json:"description" validate:"omitempty,maxlen=500"`
	Permissions *[]int   `json:"permissions" validate:"omitempty"` // Permission IDs
}

// AssignRoleRequest represents a request to assign a role to a user
type AssignRoleRequest struct {
	UserID    int        `json:"user_id" validate:"required"`
	RoleID    int        `json:"role_id" validate:"required"`
	ExpiresAt *time.Time `json:"expires_at" validate:"omitempty"`
}

// AssignPermissionsRequest represents a request to assign permissions to a role
type AssignPermissionsRequest struct {
	PermissionIDs []int `json:"permission_ids" validate:"required,minlen=1"`
}

// RoleFilter represents filters for querying roles
type RoleFilter struct {
	CollegeID    *int
	IsSystemRole *bool
	Name         *string
	Limit        int
	Offset       int
}

// PermissionFilter represents filters for querying permissions
type PermissionFilter struct {
	Resource *string
	Action   *string
	Limit    int
	Offset   int
}

// RoleWithPermissions represents a role with its associated permissions
type RoleWithPermissions struct {
	Role
	PermissionNames []string `json:"permission_names"`
}

// UserWithRoles represents a user with their assigned roles
type UserWithRoles struct {
	UserID int     `json:"user_id"`
	Name   string  `json:"name"`
	Email  string  `json:"email"`
	Role   string  `json:"role"`
	Roles  []Role  `json:"roles"`
}
