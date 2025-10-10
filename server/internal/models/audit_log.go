package models

import "time"

type AuditLog struct {
	ID         int       `json:"id" db:"id"`
	CollegeID  int       `json:"college_id" db:"college_id"`
	UserID     int       `json:"user_id" db:"user_id"`
	Action     string    `json:"action" db:"action"` // CREATE, UPDATE, DELETE, READ
	EntityType string    `json:"entity_type" db:"entity_type"` // student, course, grade, etc.
	EntityID   int       `json:"entity_id" db:"entity_id"`
	Changes    JSONMap   `json:"changes,omitempty" db:"changes"` // JSON of what changed
	IPAddress  string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent  string    `json:"user_agent,omitempty" db:"user_agent"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
}
