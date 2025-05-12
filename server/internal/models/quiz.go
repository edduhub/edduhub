package models

import "time"

// Quiz represents a quiz associated with a course.
type Quiz struct {
	ID               int       `db:"id" json:"id"`
	CollegeID        int       `db:"college_id" json:"college_id"`
	CourseID         int       `db:"course_id" json:"course_id"`
	Title            string    `db:"title" json:"title"`
	Description      string    `db:"description" json:"description"`
	TimeLimitMinutes int       `db:"time_limit_minutes" json:"time_limit_minutes"` // 0 for no limit
	DueDate          time.Time `db:"due_date" json:"due_date"`                     // Optional due date
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`

	// Relations - not stored in DB
	Course    *Course     `db:"-" json:"course,omitempty"`
	Questions []*Question `db:"-" json:"questions,omitempty"`
}

// Question represents a single question within a quiz.
type Question struct {
	ID        int       `db:"id" json:"id"`
	QuizID    int       `db:"quiz_id" json:"quiz_id"`
	Text      string    `db:"text" json:"text"`
	Type      string    `db:"type" json:"type"` // e.g., MultipleChoice, TrueFalse, ShortAnswer
	Points    int       `db:"points" json:"points"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Relations - not stored in DB
	Options []*AnswerOption `db:"-" json:"options,omitempty"` // For MultipleChoice/TrueFalse
}

// QuizAttemptStatus defines the possible statuses for a quiz attempt.
type QuizAttemptStatus string

const (
	QuizAttemptStatusInProgress QuizAttemptStatus = "InProgress"
	QuizAttemptStatusCompleted  QuizAttemptStatus = "Completed"
	QuizAttemptStatusGraded     QuizAttemptStatus = "Graded"
)

// AnswerOption represents a possible answer for a multiple-choice or true/false question.
type AnswerOption struct {
	ID         int       `db:"id" json:"id"`
	QuestionID int       `db:"question_id" json:"question_id"`
	Text       string    `db:"text" json:"text"`
	IsCorrect  bool      `db:"is_correct" json:"is_correct"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// QuizAttempt represents a student's attempt at taking a quiz.
type QuizAttempt struct {
	ID        int               `db:"id" json:"id"`
	StudentID int               `db:"student_id" json:"student_id"`
	QuizID    int               `db:"quiz_id" json:"quiz_id"`
	CollegeID int               `db:"college_id" json:"college_id"`
	CourseID  int               `db:"course_id" json:"course_id"` // Denormalized for easier querying
	StartTime time.Time         `db:"start_time" json:"start_time"`
	EndTime   time.Time         `db:"end_time" json:"end_time"` // Nullable until finished
	Score     *int              `db:"score" json:"score"`       // Nullable until graded
	Status    QuizAttemptStatus `db:"status" json:"status"`     // e.g., InProgress, Completed, Graded
	CreatedAt time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt time.Time         `db:"updated_at" json:"updated_at"`

	// Relations - not stored in DB
	Student *Student `db:"-" json:"student,omitempty"`
	Quiz    *Quiz    `db:"-" json:"quiz,omitempty"`
	// StudentAnswers []*StudentAnswer `db:"-" json:"student_answers,omitempty"` // Can be loaded separately
}

// StudentAnswer represents a student's answer to a specific question in an attempt.
type StudentAnswer struct {
	ID               int       `db:"id" json:"id"`
	QuizAttemptID    int       `db:"quiz_attempt_id" json:"quiz_attempt_id"`
	QuestionID       int       `db:"question_id" json:"question_id"`
	SelectedOptionID *int      `db:"selected_option_id" json:"selected_option_id"` // Nullable, for MC/TF
	AnswerText       string    `db:"answer_text" json:"answer_text"`               // Nullable, for ShortAnswer
	IsCorrect        *bool     `db:"is_correct" json:"is_correct"`                 // Nullable until graded
	PointsAwarded    *int      `db:"points_awarded" json:"points_awarded"`         // Nullable until graded
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

type QuestionWithCorrectAnswer struct {
}
