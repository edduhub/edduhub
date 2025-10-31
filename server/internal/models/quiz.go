package models

import "time"

type QuizType string

const (
	MultipleChoice QuizType = "MultipleChoice"
	TrueFalse      QuizType = "TrueFalse"
	ShortAnswer    QuizType = "ShortAnswer"
)

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
	ID            int       `db:"id" json:"id"`
	QuizID        int       `db:"quiz_id" json:"quiz_id"`
	Text          string    `db:"text" json:"text"`
	Type          QuizType  `db:"type" json:"type"` // e.g., MultipleChoice, TrueFalse, ShortAnswer
	Points        int       `db:"points" json:"points"`
	CorrectAnswer *string   `db:"correct_answer" json:"correct_answer,omitempty"` // For ShortAnswer questions
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`

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

// QuizAttempt represents a student's single attempt at taking a quiz.
type QuizAttempt struct {
	ID        int               `db:"id" json:"id"`
	StudentID int               `db:"student_id" json:"student_id" validate:"required"`
	QuizID    int               `db:"quiz_id" json:"quiz_id" validate:"required"`
	CollegeID int               `db:"college_id" json:"college_id" validate:"required"`
	CourseID  int               `db:"course_id" json:"course_id" validate:"required"`
	StartTime time.Time         `db:"start_time" json:"start_time"`
	EndTime   time.Time         `db:"end_time" json:"end_time"`
	Score     *int              `db:"score" json:"score"`
	Status    QuizAttemptStatus `db:"status" json:"status"`
	CreatedAt time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt time.Time         `db:"updated_at" json:"updated_at"`

	// Relations - not stored in DB
	Student *Student `db:"-" json:"student,omitempty"`
	Quiz    *Quiz    `db:"-" json:"quiz,omitempty"`
	Answers []*StudentAnswer `db:"-" json:"answers,omitempty"`
}

// StudentAnswer represents a student's answer to a specific question in an attempt.
type StudentAnswer struct {
	ID               int       `db:"id" json:"id"`
	QuizAttemptID    int       `db:"quiz_attempt_id" json:"quiz_attempt_id"`
	QuestionID       int       `db:"question_id" json:"question_id"`
	SelectedOptionID *[]int      `db:"selected_option_id" json:"selected_option_id"` // Nullable, for MC/TF
	AnswerText       string    `db:"answer_text" json:"answer_text"`               // Nullable, for ShortAnswer
	IsCorrect        *bool     `db:"is_correct" json:"is_correct"`                 // Nullable until graded
	PointsAwarded    *int      `db:"points_awarded" json:"points_awarded"`         // Nullable until graded
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

type QuestionWithCorrectAnswer struct {
	Question       Question       `json:"question"`
	CorrectOptions []*AnswerOption `json:"correct_options"` // Changed from CorrectAnswers to CorrectOptions
}
type QuestionWithStudentAnswer struct {
	Question      *Question        `json:"question"`
	StudentAnswer []*StudentAnswer `json:"student_answers"`
}

// QuestionWithOptions represents a question along with its answer options.
type QuestionWithOptions struct {
	Question      *Question       `json:"question"`
	AnswerOptions []*AnswerOption `json:"answer_options"`
}

// QuizStatistics represents various statistics for a quiz.
type QuizStatistics struct {
	QuizID            int `json:"quiz_id"`
	TotalAttempts     int `json:"total_attempts"`
	CompletedAttempts int `json:"completed_attempts"`
	AverageScore      int `json:"average_score"` // Could be float64 for more precision
	HighestScore      int `json:"highest_score"`
	LowestScore       int `json:"lowest_score"`
}

// UpdateQuizRequest provides fields for partial updates to Quiz via PATCH
type UpdateQuizRequest struct {
	CollegeID        *int       `json:"college_id" validate:"omitempty,gte=1"`
	CourseID         *int       `json:"course_id" validate:"omitempty,gte=1"`
	Title            *string    `json:"title" validate:"omitempty,min=1,max=100"`
	Description      *string    `json:"description" validate:"omitempty,max=500"`
	TimeLimitMinutes *int       `json:"time_limit_minutes" validate:"omitempty,gte=0"`
	DueDate          *time.Time `json:"due_date" validate:"omitempty"`
}

// UpdateQuestionRequest provides fields for partial updates to Question via PATCH
type UpdateQuestionRequest struct {
	QuizID *int      `json:"quiz_id" validate:"omitempty,gte=1"`
	Text   *string   `json:"text" validate:"omitempty,min=1,max=1000"`
	Type   *QuizType `json:"type" validate:"omitempty,oneof=MultipleChoice TrueFalse ShortAnswer"`
	Points *int      `json:"points" validate:"omitempty,gte=0,lte=100"`
}

// UpdateAnswerOptionRequest provides fields for partial updates to AnswerOption via PATCH
type UpdateAnswerOptionRequest struct {
	QuestionID *int    `json:"question_id" validate:"omitempty,gte=1"`
	Text       *string `json:"text" validate:"omitempty,min=1,max=250"`
	IsCorrect  *bool   `json:"is_correct" validate:"omitempty"`
}