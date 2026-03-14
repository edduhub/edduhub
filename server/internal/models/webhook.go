package models

import "time"

type Webhook struct {
	ID          int       `json:"id" db:"id"`
	CollegeID   int       `json:"college_id" db:"college_id"`
	Name        string    `json:"name,omitempty" db:"name"`
	Description string    `json:"description,omitempty" db:"description"`
	URL         string    `json:"url" db:"url"`
	Event       string    `json:"event,omitempty" db:"-"`
	EventTypes  []string  `json:"event_types,omitempty" db:"event_types"`
	Secret      string    `json:"secret,omitempty" db:"secret"`
	Active      bool      `json:"active" db:"is_active"`
	CreatedBy   *int      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
