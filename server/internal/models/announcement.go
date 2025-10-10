package models

import "time"

type Announcement struct {
	ID          int       `json:"id" db:"id"`
	CollegeID   int       `json:"college_id" db:"college_id"`
	CourseID    *int      `json:"course_id,omitempty" db:"course_id"` // Optional, null if college-wide
	Title       string    `json:"title" db:"title"`
	Content     string    `json:"content" db:"content"`
	Priority    string    `json:"priority" db:"priority"` // low, normal, high, urgent
	IsPublished bool      `json:"is_published" db:"is_published"`
	PublishedAt *time.Time `json:"published_at,omitempty" db:"published_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedBy   *string   `json:"created_by,omitempty" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type AnnouncementFilter struct {
	CollegeID   *int    `json:"college_id,omitempty"`
	CourseID    *int    `json:"course_id,omitempty"`
	Priority    *string `json:"priority,omitempty"`
	IsPublished *bool   `json:"is_published,omitempty"`
	Limit       uint64  `json:"limit,omitempty"`
	Offset      uint64  `json:"offset,omitempty"`
}

type UpdateAnnouncementRequest struct {
	ID          *int       `json:"id,omitempty"`
	Title       *string    `json:"title,omitempty"`
	Content     *string    `json:"content,omitempty"`
	Priority    *string    `json:"priority,omitempty"`
	IsPublished *bool      `json:"is_published,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedBy   *string    `json:"created_by,omitempty"`
}
