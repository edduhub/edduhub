package models

import "time"

type CalendarEventType string

const (
	EventExamType     CalendarEventType = "exam"
	EventTypeHoliday  CalendarEventType = "holiday"
	EventTypeEvent    CalendarEventType = "event"
	EventTypeDeadline CalendarEventType = "deadline"
	EventTypeOther    CalendarEventType = "other"
)

type CalendarBlock struct {
	ID          int               `db:"id" json:"id"`
	CollegeID   int               `db:"college_id" json:"college_id"`
	Title       string            `db:"title" json:"title"`
	Description string            `db:"description" json:"description"`
	EventType   CalendarEventType `db:"event_type" json:"event_type"`
	Date        time.Time         `db:"date" json:"date"`
	// Optional: For multi-day events or events with specific times
	// StartTime   *time.Time        `db:"start_time" json:"start_time,omitempty"`
	// EndTime     *time.Time        `db:"end_time" json:"end_time,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// CalendarBlockFilter can be used for querying lists of calendar blocks
type CalendarBlockFilter struct {
	CollegeID *int               `json:"college_id"` // Mandatory for most queries
	EventType *CalendarEventType `json:"event_type,omitempty"`
	StartDate *time.Time         `json:"start_date,omitempty"` // Inclusive
	EndDate   *time.Time         `json:"end_date,omitempty"`   // Inclusive
	Limit     uint64             `json:"limit,omitempty"`
	Offset    uint64             `json:"offset,omitempty"`
}

// UpdateCalendarRequest provides fields for partial updates to Calendar via PATCH
type UpdateCalendarRequest struct {
	Title *string `json:"title" validate:"omitempty,min=1,max=200"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	EventType *CalendarEventType `json:"event_type" validate:"omitempty,oneof=exam holiday event deadline other"`
	Date *time.Time `json:"date" validate:"omitempty"`
}