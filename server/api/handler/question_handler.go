package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/quiz"

	"github.com/labstack/echo/v4"
)

type QuestionHandler struct {
	questionService quiz.QuestionServiceSimple
}

func NewQuestionHandler(questionService quiz.QuestionServiceSimple) *QuestionHandler {
	return &QuestionHandler{
		questionService: questionService,
	}
}

// ListQuestions retrieves all questions for a quiz
func (h *QuestionHandler) ListQuestions(c echo.Context) error {
	quizIDStr := c.Param("quizID")
	quizID, err := strconv.Atoi(quizIDStr)
	if err != nil {
		return helpers.Error(c, "invalid quiz ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit := uint64(50)
	offset := uint64(0)

	if limitStr != "" {
		l, err := strconv.ParseUint(limitStr, 10, 64)
		if err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		o, err := strconv.ParseUint(offsetStr, 10, 64)
		if err == nil {
			offset = o
		}
	}

	questions, err := h.questionService.ListQuestionsByQuiz(c.Request().Context(), collegeID, quizID, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, questions, 200)
}

// CreateQuestion creates a new question for a quiz
func (h *QuestionHandler) CreateQuestion(c echo.Context) error {
	quizIDStr := c.Param("quizID")
	quizID, err := strconv.Atoi(quizIDStr)
	if err != nil {
		return helpers.Error(c, "invalid quiz ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.Question
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	req.QuizID = quizID

	err = h.questionService.CreateQuestion(c.Request().Context(), collegeID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, req, 201)
}

// GetQuestion retrieves a specific question
func (h *QuestionHandler) GetQuestion(c echo.Context) error {
	questionIDStr := c.Param("questionID")
	questionID, err := strconv.Atoi(questionIDStr)
	if err != nil {
		return helpers.Error(c, "invalid question ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	question, err := h.questionService.GetQuestion(c.Request().Context(), collegeID, questionID)
	if err != nil {
		return helpers.Error(c, "question not found", 404)
	}

	return helpers.Success(c, question, 200)
}

// UpdateQuestion updates a question
func (h *QuestionHandler) UpdateQuestion(c echo.Context) error {
	questionIDStr := c.Param("questionID")
	questionID, err := strconv.Atoi(questionIDStr)
	if err != nil {
		return helpers.Error(c, "invalid question ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.Question
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	req.ID = questionID

	err = h.questionService.UpdateQuestion(c.Request().Context(), collegeID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Question updated successfully", 200)
}

// DeleteQuestion deletes a question
func (h *QuestionHandler) DeleteQuestion(c echo.Context) error {
	questionIDStr := c.Param("questionID")
	questionID, err := strconv.Atoi(questionIDStr)
	if err != nil {
		return helpers.Error(c, "invalid question ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.questionService.DeleteQuestion(c.Request().Context(), collegeID, questionID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Question deleted successfully", 200)
}
