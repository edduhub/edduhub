package handler

import (
	"strconv"
	"time"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/assignment"
	"eduhub/server/internal/services/course"
	"eduhub/server/internal/services/enrollment"

	"github.com/labstack/echo/v4"
)

type AssignmentHandler struct {
	assignmentService  assignment.AssignmentService
	enrollmentService  enrollment.EnrollmentService
	courseService      course.CourseService
}

func NewAssignmentHandler(assignmentService assignment.AssignmentService, enrollmentService enrollment.EnrollmentService, courseService course.CourseService) *AssignmentHandler {
	return &AssignmentHandler{
		assignmentService:  assignmentService,
		enrollmentService:  enrollmentService,
		courseService:      courseService,
	}
}

func (h *AssignmentHandler) ListAssignments(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	assignments, err := h.assignmentService.GetAssignmentsByCourse(c.Request().Context(), collegeID, courseID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, assignments, 200)
}

func (h *AssignmentHandler) CreateAssignment(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var assignment models.Assignment
	if err := c.Bind(&assignment); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	assignment.CourseID = courseID
	assignment.CollegeID = collegeID

	err = h.assignmentService.CreateAssignment(c.Request().Context(), &assignment)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, assignment, 201)
}

func (h *AssignmentHandler) GetAssignment(c echo.Context) error {
	assignmentIDStr := c.Param("assignmentID")
	assignmentID, err := strconv.Atoi(assignmentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid assignment ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	assignment, err := h.assignmentService.GetAssignment(c.Request().Context(), collegeID, assignmentID)
	if err != nil {
		return helpers.Error(c, "assignment not found", 404)
	}

	return helpers.Success(c, assignment, 200)
}

func (h *AssignmentHandler) UpdateAssignment(c echo.Context) error {
	assignmentIDStr := c.Param("assignmentID")
	assignmentID, err := strconv.Atoi(assignmentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid assignment ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.UpdateAssignmentRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.assignmentService.UpdateAssignment(c.Request().Context(), collegeID, assignmentID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Assignment updated successfully", 204)
}

func (h *AssignmentHandler) DeleteAssignment(c echo.Context) error {
	assignmentIDStr := c.Param("assignmentID")
	assignmentID, err := strconv.Atoi(assignmentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid assignment ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.assignmentService.DeleteAssignment(c.Request().Context(), collegeID, assignmentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Assignment deleted successfully", 204)
}

func (h *AssignmentHandler) SubmitAssignment(c echo.Context) error {
	assignmentIDStr := c.Param("assignmentID")
	assignmentID, err := strconv.Atoi(assignmentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid assignment ID", 400)
	}

	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return helpers.Error(c, "student ID required", 400)
	}

	var submission models.AssignmentSubmission
	if err := c.Bind(&submission); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	submission.AssignmentID = assignmentID
	submission.StudentID = studentID

	err = h.assignmentService.SubmitAssignment(c.Request().Context(), &submission)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, submission, 201)
}

func (h *AssignmentHandler) GradeSubmission(c echo.Context) error {
	submissionIDStr := c.Param("submissionID")
	submissionID, err := strconv.Atoi(submissionIDStr)
	if err != nil {
		return helpers.Error(c, "invalid submission ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req struct {
		Grade    *int    `json:"grade"`
		Feedback *string `json:"feedback"`
	}
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.assignmentService.GradeSubmission(c.Request().Context(), collegeID, submissionID, req.Grade, req.Feedback)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Submission graded successfully", 200)
}

// GetMyAssignments returns all assignments across all enrolled courses for current student
func (h *AssignmentHandler) GetMyAssignments(c echo.Context) error {
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
		return helpers.Error(c, "failed to fetch enrollments", 500)
	}

	response := make([]map[string]interface{}, 0)

	for _, enrollmentRecord := range enrollments {
		assignments, err := h.assignmentService.GetAssignmentsByCourse(ctx, collegeID, enrollmentRecord.CourseID)
		if err != nil {
			return helpers.Error(c, "failed to fetch assignments", 500)
		}

		courseDetails, err := h.courseService.FindCourseByID(ctx, collegeID, enrollmentRecord.CourseID)
		courseName := ""
		if err == nil && courseDetails != nil {
			courseName = courseDetails.Name
		}

		for _, a := range assignments {
			// Check if student has submitted this assignment
			status := "pending"
			var score *int
			submission, err := h.assignmentService.GetSubmissionByStudentAndAssignment(ctx, studentID, a.ID)
			if err == nil && submission != nil {
				if submission.Grade != nil {
					status = "graded"
					score = submission.Grade
				} else {
					status = "submitted"
				}
			}

			assignmentData := map[string]interface{}{
				"id":          a.ID,
				"title":       a.Title,
				"description": a.Description,
				"courseId":    a.CourseID,
				"courseName":  courseName,
				"dueDate":     a.DueDate.Format(time.RFC3339),
				"maxScore":    a.MaxPoints,
				"status":      status,
			}
			
			if score != nil {
				assignmentData["score"] = *score
			}

			response = append(response, assignmentData)
		}
	}

	return helpers.Success(c, response, 200)
}
