package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Placement represents a job placement/recruitment drive record.
type Placement struct {
	ID                  int         `db:"id" json:"id"`
	CollegeID           int         `db:"college_id" json:"college_id"`
	CompanyName         string      `db:"company_name" json:"company_name"`
	CompanyLogo         *string     `db:"company_logo" json:"company_logo,omitempty"`
	JobTitle            string      `db:"job_title" json:"job_title"`
	JobDescription      *string     `db:"job_description" json:"job_description,omitempty"`
	JobType             *string     `db:"job_type" json:"job_type,omitempty"` // full_time, part_time, internship, contract
	Location            *string     `db:"location" json:"location,omitempty"`
	IsRemote            bool        `db:"is_remote" json:"is_remote"`
	SalaryRangeMin      *float64    `db:"salary_range_min" json:"salary_range_min,omitempty"`
	SalaryRangeMax      *float64    `db:"salary_range_max" json:"salary_range_max,omitempty"`
	SalaryCurrency      string      `db:"salary_currency" json:"salary_currency"`
	RequiredSkills      StringArray `db:"required_skills" json:"required_skills,omitempty"`
	EligibilityCriteria *string     `db:"eligibility_criteria" json:"eligibility_criteria,omitempty"`
	ApplicationDeadline *time.Time  `db:"application_deadline" json:"application_deadline,omitempty"`
	DriveDate           *time.Time  `db:"drive_date" json:"drive_date,omitempty"`
	InterviewMode       *string     `db:"interview_mode" json:"interview_mode,omitempty"` // on_campus, virtual, hybrid
	MaxApplications     *int        `db:"max_applications" json:"max_applications,omitempty"`
	Status              string      `db:"status" json:"status"` // open, closed, in_progress, completed, cancelled
	CreatedBy           *int        `db:"created_by" json:"created_by,omitempty"`
	CreatedAt           time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time   `db:"updated_at" json:"updated_at"`
}

// PlacementApplication represents a student's application to a placement.
type PlacementApplication struct {
	ID          int       `db:"id" json:"id"`
	PlacementID int       `db:"placement_id" json:"placement_id"`
	StudentID   int       `db:"student_id" json:"student_id"`
	Status      string    `db:"status" json:"status"` // applied, shortlisted, interview_scheduled, selected, rejected, withdrawn
	ResumeURL   *string   `db:"resume_url" json:"resume_url,omitempty"`
	CoverLetter *string   `db:"cover_letter" json:"cover_letter,omitempty"`
	AppliedAt   time.Time `db:"applied_at" json:"applied_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`

	// Relations
	Placement *Placement `db:"-" json:"placement,omitempty"`
	Student   *Student   `db:"-" json:"student,omitempty"`
}

// PlacementInterview represents interview details for a placement application.
type PlacementInterview struct {
	ID               int        `db:"id" json:"id"`
	ApplicationID    int        `db:"application_id" json:"application_id"`
	RoundNumber      int        `db:"round_number" json:"round_number"`
	RoundName        *string    `db:"round_name" json:"round_name,omitempty"`
	ScheduledAt      *time.Time `db:"scheduled_at" json:"scheduled_at,omitempty"`
	DurationMinutes  *int       `db:"duration_minutes" json:"duration_minutes,omitempty"`
	Mode             *string    `db:"mode" json:"mode,omitempty"` // virtual, in_person, phone
	MeetingLink      *string    `db:"meeting_link" json:"meeting_link,omitempty"`
	Location         *string    `db:"location" json:"location,omitempty"`
	InterviewerName  *string    `db:"interviewer_name" json:"interviewer_name,omitempty"`
	InterviewerEmail *string    `db:"interviewer_email" json:"interviewer_email,omitempty"`
	Feedback         *string    `db:"feedback" json:"feedback,omitempty"`
	Result           *string    `db:"result" json:"result,omitempty"` // pending, passed, failed, no_show
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`

	// Relations
	Application *PlacementApplication `db:"-" json:"application,omitempty"`
}

// StringArray is a custom type for PostgreSQL TEXT[] array
type StringArray []string

// Value implements driver.Valuer for StringArray
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan implements sql.Scanner for StringArray
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan StringArray: not a byte slice")
	}
	return json.Unmarshal(bytes, a)
}

// Placement status constants
const (
	PlacementStatusOpen       = "open"
	PlacementStatusClosed     = "closed"
	PlacementStatusInProgress = "in_progress"
	PlacementStatusCompleted  = "completed"
	PlacementStatusCancelled  = "cancelled"
)

// PlacementApplication status constants
const (
	ApplicationStatusApplied            = "applied"
	ApplicationStatusShortlisted        = "shortlisted"
	ApplicationStatusInterviewScheduled = "interview_scheduled"
	ApplicationStatusSelected           = "selected"
	ApplicationStatusRejected           = "rejected"
	ApplicationStatusWithdrawn          = "withdrawn"
)

// Job type constants
const (
	JobTypeFullTime   = "full_time"
	JobTypePartTime   = "part_time"
	JobTypeInternship = "internship"
	JobTypeContract   = "contract"
)

// Interview mode constants
const (
	InterviewModeVirtual  = "virtual"
	InterviewModeInPerson = "in_person"
	InterviewModePhone    = "phone"
	InterviewModeOnCampus = "on_campus"
	InterviewModeHybrid   = "hybrid"
)
