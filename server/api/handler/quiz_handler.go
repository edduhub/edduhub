package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/course"
	"eduhub/server/internal/services/enrollment"
	"eduhub/server/internal/services/quiz"

	"github.com/labstack/echo/v4"
)

type QuizHandler struct {
	quizService       quiz.QuizService
	enrollmentService enrollment.EnrollmentService
	courseService     course.CourseService
}

func NewQuizHandler(quizService quiz.QuizService, enrollmentService enrollment.EnrollmentService, courseService course.CourseService) *QuizHandler {
	return &QuizHandler{
		quizService:       quizService,
		enrollmentService: enrollmentService,
		courseService:     courseService,
	}
}

// ListQuizzes retrieves all quizzes for a course
func (h *QuizHandler) ListQuizzes(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	limitStr := c.QueryParam("limit")
	offsetStr := c.QueryParam("offset")

	limit := uint64(20)
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

	quizzes, err := h.quizService.FindQuizzesByCourse(c.Request().Context(), collegeID, courseID, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, quizzes, 200)
}

// CreateQuiz creates a new quiz for a course
func (h *QuizHandler) CreateQuiz(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var quizReq models.Quiz
	if err := c.Bind(&quizReq); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	quizReq.CourseID = courseID
	quizReq.CollegeID = collegeID

	err = h.quizService.CreateQuiz(c.Request().Context(), &quizReq)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, quizReq, 201)
}

// GetQuiz retrieves a specific quiz
func (h *QuizHandler) GetQuiz(c echo.Context) error {
	quizIDStr := c.Param("quizID")
	quizID, err := strconv.Atoi(quizIDStr)
	if err != nil {
		return helpers.Error(c, "invalid quiz ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	quiz, err := h.quizService.GetQuizByID(c.Request().Context(), collegeID, quizID)
	if err != nil {
		return helpers.Error(c, "quiz not found", 404)
	}

	return helpers.Success(c, quiz, 200)
}

// UpdateQuiz updates a quiz
func (h *QuizHandler) UpdateQuiz(c echo.Context) error {
	quizIDStr := c.Param("quizID")
	quizID, err := strconv.Atoi(quizIDStr)
	if err != nil {
		return helpers.Error(c, "invalid quiz ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Get the existing quiz first
	quiz, err := h.quizService.GetQuizByID(c.Request().Context(), collegeID, quizID)
	if err != nil {
		return helpers.Error(c, "quiz not found", 404)
	}

	// Bind update request
	var req models.UpdateQuizRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	// Apply updates to quiz
	if req.Title != nil {
		quiz.Title = *req.Title
	}
	if req.Description != nil {
		quiz.Description = *req.Description
	}
	if req.TimeLimitMinutes != nil {
		quiz.TimeLimitMinutes = *req.TimeLimitMinutes
	}
	if req.DueDate != nil {
		quiz.DueDate = *req.DueDate
	}

	err = h.quizService.UpdateQuiz(c.Request().Context(), quiz)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Quiz updated successfully", 200)
}

// DeleteQuiz deletes a quiz
func (h *QuizHandler) DeleteQuiz(c echo.Context) error {
	quizIDStr := c.Param("quizID")
	quizID, err := strconv.Atoi(quizIDStr)
	if err != nil {
		return helpers.Error(c, "invalid quiz ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.quizService.DeleteQuiz(c.Request().Context(), collegeID, quizID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Quiz deleted successfully", 200)
}

// GetMyQuizzes returns all quizzes across all enrolled courses for current student
func (h *QuizHandler) GetMyQuizzes(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return helpers.Error(c, "student ID required", 400)
	}

	ctx := c.Request().Context()
	enrollments, err := h.enrollmentService.FindEnrollmentsByStudent(ctx, collegeID, studentID, 1000, 0)
	if err != nil {
		return helpers.Error(c, "failed to load enrollments", 500)
	}

	courseNames := make(map[int]string)
	result := make([]map[string]any, 0)

	for _, enrollmentRecord := range enrollments {
		quizzes, err := h.quizService.FindQuizzesByCourse(ctx, collegeID, enrollmentRecord.CourseID, 100, 0)
		if err != nil {
			return helpers.Error(c, "failed to load quizzes", 500)
		}

		courseName, cached := courseNames[enrollmentRecord.CourseID]
		if !cached {
			courseInfo, err := h.courseService.FindCourseByID(ctx, collegeID, enrollmentRecord.CourseID)
			if err == nil && courseInfo != nil {
				courseName = courseInfo.Name
				courseNames[enrollmentRecord.CourseID] = courseName
			} else {
				courseName = "Course " + strconv.Itoa(enrollmentRecord.CourseID)
			}
		}

		for _, quiz := range quizzes {
			result = append(result, map[string]any{
				"id":              quiz.ID,
				"title":           quiz.Title,
				"description":     quiz.Description,
				"courseId":        quiz.CourseID,
				"courseName":      courseName,
				"duration":        quiz.TimeLimitMinutes,
				"dueDate":         quiz.DueDate,
				"status":          "not_started",
				"attempts":        0,
				"maxAttempts":     1,
				"allowedAttempts": 1,
				"questions":       quiz.Questions,
			})
		}
	}

	return helpers.Success(c, result, 200)
}
