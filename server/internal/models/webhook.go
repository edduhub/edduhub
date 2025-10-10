package models

import "time"

type Webhook struct {
	ID        int       `json:"id" db:"id"`
	CollegeID int       `json:"college_id" db:"college_id"`
	URL       string    `json:"url" db:"url"`
	Event     string    `json:"event" db:"event"` // student.created, grade.updated, etc.
	Secret    string    `json:"secret,omitempty" db:"secret"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
