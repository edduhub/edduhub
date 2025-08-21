package models

// StudentPerformance represents a student's performance across quizzes
type StudentPerformance struct {
	StudentID        int     `json:"student_id"`
	TotalQuizzes     int     `json:"total_quizzes"`
	CompletedQuizzes int     `json:"completed_quizzes"`
	AverageScore     float64 `json:"average_score"`
	HighestScore     int     `json:"highest_score"`
	LowestScore      int     `json:"lowest_score"`
}

type QuestionWithCorrectAnswers struct {
	Question       *Question       `json:"question"`
	CorrectOptions []*AnswerOption `json:"correct_options"`
}
