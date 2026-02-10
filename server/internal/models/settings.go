package models

import "time"

type Settings struct {
	UserID             string    `db:"user_id" json:"user_id"`
	EmailNotifications bool      `db:"email_notifications" json:"email_notifications"`
	PushNotifications  bool      `db:"push_notifications" json:"push_notifications"`
	Theme              string    `db:"theme" json:"theme"`
	Language           string    `db:"language" json:"language"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
}

type SettingsUpdateRequest struct {
	EmailNotifications *bool   `json:"email_notifications,omitempty"`
	PushNotifications  *bool   `json:"push_notifications,omitempty"`
	Theme              *string `json:"theme,omitempty"`
	Language           *string `json:"language,omitempty"`
}
