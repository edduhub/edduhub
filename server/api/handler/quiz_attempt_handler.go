package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/quiz"

	"github.com/labstack/echo/v4"
)

type QuizAttemptHandler struct {
	attemptService quiz.QuizAttemptServiceSimple
}

func NewQuizAttemptHandler(attemptService quiz.QuizAttemptServiceSimple) *QuizAttemptHandler {
	return &QuizAttemptHandler{
		attemptService: attemptService,
	}
}

// StartQuizAttempt initiates a new quiz attempt for a student
func (h *QuizAttemptHandler) StartQuizAttempt(c echo.Context) error {
	quizIDStr := c.Param("quizID")
	quizID, err := strconv.Atoi(quizIDStr)
	if err != nil {
		return helpers.Error(c, "invalid quiz ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	studentID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "student ID required", 401)
	}

	attempt, err := h.attemptService.StartAttempt(c.Request().Context(), collegeID, quizID, studentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, attempt, 201)
}

// SubmitQuizAttempt submits answers and completes a quiz attempt
func (h *QuizAttemptHandler) SubmitQuizAttempt(c echo.Context) error {
	attemptIDStr := c.Param("attemptID")
	attemptID, err := strconv.Atoi(attemptIDStr)
	if err != nil {
		return helpers.Error(c, "invalid attempt ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	studentID, err := helpers.ExtractUserID(c)
	if err != nil {
		return helpers.Error(c, "student ID required", 401)
	}

	var req struct {
		Answers []models.StudentAnswer `json:"answers"`
	}
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	result, err := h.attemptService.SubmitAttempt(c.Request().Context(), collegeID, attemptID, studentID, req.Answers)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, result, 200)
}

// GetQuizAttempt retrieves details of a quiz attempt
func (h *QuizAttemptHandler) GetQuizAttempt(c echo.Context) error {
	attemptIDStr := c.Param("attemptID")
	attemptID, err := strconv.Atoi(attemptIDStr)
	if err != nil {
		return helpers.Error(c, "invalid attempt ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	attempt, err := h.attemptService.GetAttempt(c.Request().Context(), collegeID, attemptID)
	if err != nil {
		return helpers.Error(c, "attempt not found", 404)
	}

	return helpers.Success(c, attempt, 200)
}

// ListStudentAttempts retrieves all attempts for a student
func (h *QuizAttemptHandler) ListStudentAttempts(c echo.Context) error {
	studentIDStr := c.Param("studentID")
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	attempts, err := h.attemptService.GetStudentAttempts(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, attempts, 200)
}

// ListQuizAttempts retrieves all attempts for a quiz (Faculty/Admin)
func (h *QuizAttemptHandler) ListQuizAttempts(c echo.Context) error {
	quizIDStr := c.Param("quizID")
	quizID, err := strconv.Atoi(quizIDStr)
	if err != nil {
		return helpers.Error(c, "invalid quiz ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	attempts, err := h.attemptService.GetQuizAttempts(c.Request().Context(), collegeID, quizID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, attempts, 200)
}
