package models

import "time"

// SelfServiceRequest represents a student request submitted through self-service workflows.
type SelfServiceRequest struct {
	ID             int        `db:"id" json:"id"`
	StudentID      int        `db:"student_id" json:"student_id"`
	CollegeID      int        `db:"college_id" json:"college_id"`
	Type           string     `db:"type" json:"type"`
	Title          string     `db:"title" json:"title"`
	Description    string     `db:"description" json:"description"`
	Status         string     `db:"status" json:"status"`
	DocumentType   *string    `db:"document_type" json:"document_type,omitempty"`
	DeliveryMethod *string    `db:"delivery_method" json:"delivery_method,omitempty"`
	AdminResponse  *string    `db:"admin_response" json:"response,omitempty"`
	RespondedBy    *int       `db:"responded_by" json:"responded_by,omitempty"`
	RespondedAt    *time.Time `db:"responded_at" json:"responded_at,omitempty"`
	SubmittedAt    time.Time  `db:"submitted_at" json:"submitted_at"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

// CreateSelfServiceRequestInput is the payload for creating student requests.
type CreateSelfServiceRequestInput struct {
	Type           string  `json:"type"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	DocumentType   *string `json:"document_type,omitempty"`
	DeliveryMethod *string `json:"delivery_method,omitempty"`
}

// UpdateSelfServiceRequestInput is the payload for admin-side status updates.
type UpdateSelfServiceRequestInput struct {
	Status   string `json:"status"`
	Response string `json:"response"`
}
