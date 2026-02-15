package models

import "time"

type Settings struct {
	UserID              string    `db:"user_id" json:"user_id"`
	EmailNotifications  bool      `db:"email_notifications" json:"email_notifications"`
	PushNotifications   bool      `db:"push_notifications" json:"push_notifications"`
	AssignmentReminders bool      `db:"assignment_reminders" json:"assignment_reminders"`
	GradeUpdates        bool      `db:"grade_updates" json:"grade_updates"`
	AnnouncementAlerts  bool      `db:"announcement_alerts" json:"announcement_alerts"`
	Theme               string    `db:"theme" json:"theme"`
	Language            string    `db:"language" json:"language"`
	Timezone            string    `db:"timezone" json:"timezone"`
	SessionTimeout      int       `db:"session_timeout" json:"session_timeout"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}

type SettingsUpdateRequest struct {
	EmailNotifications  *bool   `json:"email_notifications,omitempty"`
	PushNotifications   *bool   `json:"push_notifications,omitempty"`
	AssignmentReminders *bool   `json:"assignment_reminders,omitempty"`
	GradeUpdates        *bool   `json:"grade_updates,omitempty"`
	AnnouncementAlerts  *bool   `json:"announcement_alerts,omitempty"`
	Theme               *string `json:"theme,omitempty"`
	Language            *string `json:"language,omitempty"`
	Timezone            *string `json:"timezone,omitempty"`
	SessionTimeout      *int    `json:"session_timeout,omitempty"`
}
