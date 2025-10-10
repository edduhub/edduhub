package models

import "time"

type Notification struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	CollegeID int       `json:"college_id" db:"college_id"`
	Title     string    `json:"title" db:"title"`
	Message   string    `json:"message" db:"message"`
	Type      string    `json:"type" db:"type"` // info, warning, success, error
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
