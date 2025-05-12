package models

// QuizStatistics represents aggregated statistics for a quiz
type QuizStatistics struct {
	QuizID            int `json:"quiz_id"`
	TotalAttempts     int `json:"total_attempts"`
	CompletedAttempts int `json:"completed_attempts"`
	HighestScore      int `json:"highest_score"`
	LowestScore       int `json:"lowest_score"`
	AverageScore      int `json:"average_score"`
}

// QuestionWithOptions combines a question with its answer options
type QuestionWithOptions struct {
	Question      *Question       `json:"question"`
	AnswerOptions []*AnswerOption `json:"answer_options"`
}

// StudentPerformance represents a student's performance across quizzes
type StudentPerformance struct {
	StudentID        int     `json:"student_id"`
	TotalQuizzes     int     `json:"total_quizzes"`
	CompletedQuizzes int     `json:"completed_quizzes"`
	AverageScore     float64 `json:"average_score"`
	HighestScore     int     `json:"highest_score"`
	LowestScore      int     `json:"lowest_score"`
}
