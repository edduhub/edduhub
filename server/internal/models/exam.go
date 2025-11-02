package models

import "time"

// Exam represents a formal examination in the system
type Exam struct {
	ID          int       `db:"id" json:"id"`
	CollegeID   int       `db:"college_id" json:"college_id"`
	CourseID    int       `db:"course_id" json:"course_id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	ExamType    string    `db:"exam_type" json:"exam_type"` // midterm, final, quiz, practical
	StartTime   time.Time `db:"start_time" json:"start_time"`
	EndTime     time.Time `db:"end_time" json:"end_time"`
	Duration    int       `db:"duration" json:"duration"` // Duration in minutes
	TotalMarks  float64   `db:"total_marks" json:"total_marks"`
	PassingMarks float64  `db:"passing_marks" json:"passing_marks"`
	RoomID      *int      `db:"room_id" json:"room_id,omitempty"`
	Status      string    `db:"status" json:"status"` // scheduled, ongoing, completed, cancelled
	CreatedBy   int       `db:"created_by" json:"created_by"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`

	// Metadata
	Instructions       string            `db:"instructions" json:"instructions,omitempty"`
	AllowedMaterials   string            `db:"allowed_materials" json:"allowed_materials,omitempty"`
	QuestionPaperSets  int               `db:"question_paper_sets" json:"question_paper_sets"` // Number of different question paper sets
}

// ExamEnrollment represents a student's enrollment in an exam
type ExamEnrollment struct {
	ID              int        `db:"id" json:"id"`
	ExamID          int        `db:"exam_id" json:"exam_id"`
	StudentID       int        `db:"student_id" json:"student_id"`
	CollegeID       int        `db:"college_id" json:"college_id"`
	EnrollmentDate  time.Time  `db:"enrollment_date" json:"enrollment_date"`
	SeatNumber      *string    `db:"seat_number" json:"seat_number,omitempty"`
	RoomNumber      *string    `db:"room_number" json:"room_number,omitempty"`
	QuestionPaperSet *int      `db:"question_paper_set" json:"question_paper_set,omitempty"`
	Status          string     `db:"status" json:"status"` // enrolled, appeared, absent, disqualified
	HallTicketGenerated bool   `db:"hall_ticket_generated" json:"hall_ticket_generated"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// ExamResult represents the result of a student's exam
type ExamResult struct {
	ID                int        `db:"id" json:"id"`
	ExamID            int        `db:"exam_id" json:"exam_id"`
	StudentID         int        `db:"student_id" json:"student_id"`
	CollegeID         int        `db:"college_id" json:"college_id"`
	MarksObtained     *float64   `db:"marks_obtained" json:"marks_obtained,omitempty"`
	Grade             *string    `db:"grade" json:"grade,omitempty"`
	Percentage        *float64   `db:"percentage" json:"percentage,omitempty"`
	Result            string     `db:"result" json:"result"` // pass, fail, absent, pending
	Remarks           string     `db:"remarks" json:"remarks,omitempty"`
	EvaluatedBy       *int       `db:"evaluated_by" json:"evaluated_by,omitempty"`
	EvaluatedAt       *time.Time `db:"evaluated_at" json:"evaluated_at,omitempty"`
	RevaluationStatus string     `db:"revaluation_status" json:"revaluation_status"` // none, requested, in_progress, completed
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

// RevaluationRequest represents a request for exam re-evaluation
type RevaluationRequest struct {
	ID              int        `db:"id" json:"id"`
	ExamResultID    int        `db:"exam_result_id" json:"exam_result_id"`
	StudentID       int        `db:"student_id" json:"student_id"`
	CollegeID       int        `db:"college_id" json:"college_id"`
	Reason          string     `db:"reason" json:"reason"`
	Status          string     `db:"status" json:"status"` // pending, approved, rejected, completed
	PreviousMarks   float64    `db:"previous_marks" json:"previous_marks"`
	RevisedMarks    *float64   `db:"revised_marks" json:"revised_marks,omitempty"`
	ReviewedBy      *int       `db:"reviewed_by" json:"reviewed_by,omitempty"`
	ReviewComments  string     `db:"review_comments" json:"review_comments,omitempty"`
	RequestedAt     time.Time  `db:"requested_at" json:"requested_at"`
	ReviewedAt      *time.Time `db:"reviewed_at" json:"reviewed_at,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// ExamRoom represents a physical room/hall for conducting exams
type ExamRoom struct {
	ID           int       `db:"id" json:"id"`
	CollegeID    int       `db:"college_id" json:"college_id"`
	RoomNumber   string    `db:"room_number" json:"room_number"`
	RoomName     string    `db:"room_name" json:"room_name"`
	Capacity     int       `db:"capacity" json:"capacity"`
	Location     string    `db:"location" json:"location"`
	Facilities   string    `db:"facilities" json:"facilities,omitempty"` // JSON string or comma-separated
	IsActive     bool      `db:"is_active" json:"is_active"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// DTO for creating/updating exams
type CreateExamRequest struct {
	CourseID           int       `json:"course_id" validate:"required"`
	Title              string    `json:"title" validate:"required"`
	Description        string    `json:"description"`
	ExamType           string    `json:"exam_type" validate:"required,oneof=midterm final quiz practical"`
	StartTime          time.Time `json:"start_time" validate:"required"`
	EndTime            time.Time `json:"end_time" validate:"required,gtfield=StartTime"`
	Duration           int       `json:"duration" validate:"required,min=1"`
	TotalMarks         float64   `json:"total_marks" validate:"required,min=0"`
	PassingMarks       float64   `json:"passing_marks" validate:"required,min=0"`
	Instructions       string    `json:"instructions"`
	AllowedMaterials   string    `json:"allowed_materials"`
	QuestionPaperSets  int       `json:"question_paper_sets" validate:"min=1"`
}

// DTO for exam result submission
type ExamResultRequest struct {
	StudentID      int      `json:"student_id" validate:"required"`
	MarksObtained  float64  `json:"marks_obtained" validate:"required,min=0"`
	Remarks        string   `json:"remarks"`
}

// DTO for hall ticket generation
type HallTicketResponse struct {
	ExamID           int       `json:"exam_id"`
	StudentID        int       `json:"student_id"`
	StudentName      string    `json:"student_name"`
	ExamTitle        string    `json:"exam_title"`
	ExamDate         time.Time `json:"exam_date"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	Duration         int       `json:"duration"`
	SeatNumber       string    `json:"seat_number"`
	RoomNumber       string    `json:"room_number"`
	QuestionPaperSet int       `json:"question_paper_set"`
	Instructions     string    `json:"instructions"`
}
