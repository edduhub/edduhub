package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/course"
	"eduhub/server/internal/services/enrollment"
	"eduhub/server/internal/services/student"

	"github.com/labstack/echo/v4"
)

type CourseHandler struct {
	courseService     course.CourseService
	enrollmentService enrollment.EnrollmentService
	studentService    student.StudentService
}

func NewCourseHandler(
	courseService course.CourseService,
	enrollmentService enrollment.EnrollmentService,
	studentService student.StudentService,
) *CourseHandler {
	return &CourseHandler{
		courseService:     courseService,
		enrollmentService: enrollmentService,
		studentService:    studentService,
	}
}

func (h *CourseHandler) UpdateCourse(c echo.Context) error {
	// Extract URL parameters
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Bind the UpdateCourseRequest struct
	var req models.UpdateCourseRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	// Call service method with partial update request
	err = h.courseService.UpdateCoursePartial(c.Request().Context(), collegeID, courseID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Success", 204)
}

func (h *CourseHandler) ListCourses(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Parse pagination parameters
	limit := uint64(100) // default limit - increased for better UX
	offset := uint64(0)  // default offset

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if parsedLimit, err := strconv.ParseUint(limitParam, 10, 64); err == nil {
			limit = parsedLimit
		}
	}

	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.ParseUint(offsetParam, 10, 64); err == nil {
			offset = parsedOffset
		}
	}

	courses, err := h.courseService.FindAllCourses(c.Request().Context(), collegeID, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	// Enrich courses with enrollment count and instructor name
	enrichedCourses := make([]map[string]any, 0, len(courses))

	for _, course := range courses {
		// Get enrollment count
		enrollmentCount := 0
		enrollments, err := h.enrollmentService.FindEnrollmentsByCourse(c.Request().Context(), collegeID, course.ID, 1000, 0)
		if err == nil {
			enrollmentCount = len(enrollments)
		}

		// Get instructor name
		instructorName := "Unknown"
		if course.Instructor != nil {
			instructorName = course.Instructor.Name
		}

		enrichedCourses = append(enrichedCourses, map[string]any{
			"id":               course.ID,
			"code":             "COURSE-" + strconv.Itoa(course.ID),
			"name":             course.Name,
			"description":      course.Description,
			"credits":          course.Credits,
			"instructor":       instructorName,
			"instructorId":     course.InstructorID,
			"enrolledStudents": enrollmentCount,
			"maxStudents":      100, // Default max, could be made configurable
			"semester":         "Current",
			"department":       "General",
		})
	}

	return helpers.Success(c, enrichedCourses, 200)
}

func (h *CourseHandler) GetCourse(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	course, err := h.courseService.FindCourseByID(c.Request().Context(), collegeID, courseID)
	if err != nil {
		return helpers.Error(c, "course not found", 404)
	}

	return helpers.Success(c, course, 200)
}

func (h *CourseHandler) CreateCourse(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var course models.Course
	if err := c.Bind(&course); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	course.CollegeID = collegeID

	err = h.courseService.CreateCourse(c.Request().Context(), &course)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, course, 201)
}

func (h *CourseHandler) DeleteCourse(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.courseService.DeleteCourse(c.Request().Context(), collegeID, courseID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Course deleted successfully", 204)
}

func (h *CourseHandler) EnrollStudents(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req struct {
		StudentIDs []int `json:"student_ids" validate:"required,min=1"`
	}

	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	if len(req.StudentIDs) == 0 {
		return helpers.Error(c, "at least one student ID is required", 400)
	}

	_, err = h.courseService.FindCourseByID(c.Request().Context(), collegeID, courseID)
	if err != nil {
		return helpers.Error(c, "course not found", 404)
	}

	successCount := 0
	failedEnrollments := []map[string]any{}

	for _, studentID := range req.StudentIDs {
		enrolled, err := h.enrollmentService.IsStudentEnrolled(c.Request().Context(), collegeID, studentID, courseID)
		if err != nil {
			failedEnrollments = append(failedEnrollments, map[string]any{
				"student_id": studentID,
				"error":      "failed to check enrollment status",
			})
			continue
		}

		if enrolled {
			failedEnrollments = append(failedEnrollments, map[string]any{
				"student_id": studentID,
				"error":      "already enrolled",
			})
			continue
		}

		enrollment := &models.Enrollment{
			StudentID: studentID,
			CourseID:  courseID,
			CollegeID: collegeID,
			Status:    models.Active,
		}

		err = h.enrollmentService.CreateEnrollment(c.Request().Context(), enrollment)
		if err != nil {
			failedEnrollments = append(failedEnrollments, map[string]any{
				"student_id": studentID,
				"error":      err.Error(),
			})
			continue
		}
		successCount++
	}

	return helpers.Success(c, map[string]any{
		"message":       "enrollment processed",
		"success_count": successCount,
		"failed":        failedEnrollments,
	}, 200)
}

func (h *CourseHandler) RemoveStudent(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	studentIDStr := c.Param("studentID")
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	enrollments, err := h.enrollmentService.FindEnrollmentsByStudent(c.Request().Context(), collegeID, studentID, 100, 0)
	if err != nil {
		return helpers.Error(c, "failed to find enrollments", 500)
	}

	var enrollmentID int
	found := false
	for _, enrollment := range enrollments {
		if enrollment.CourseID == courseID {
			enrollmentID = enrollment.ID
			found = true
			break
		}
	}

	if !found {
		return helpers.Error(c, "student not enrolled in this course", 404)
	}

	err = h.enrollmentService.DeleteEnrollment(c.Request().Context(), collegeID, enrollmentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Student removed from course successfully", 204)
}

func (h *CourseHandler) ListEnrolledStudents(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	limit := uint64(50)
	offset := uint64(0)

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if parsedLimit, err := strconv.ParseUint(limitParam, 10, 64); err == nil {
			limit = parsedLimit
		}
	}

	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.ParseUint(offsetParam, 10, 64); err == nil {
			offset = parsedOffset
		}
	}

	enrollments, err := h.enrollmentService.FindEnrollmentsByStudent(c.Request().Context(), collegeID, courseID, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	studentIDs := make([]int, 0, len(enrollments))
	for _, enrollment := range enrollments {
		studentIDs = append(studentIDs, enrollment.StudentID)
	}

	students := []map[string]any{}
	for _, enrollment := range enrollments {
		profile, err := h.studentService.GetStudentDetailedProfile(c.Request().Context(), collegeID, enrollment.StudentID)
		if err == nil && profile != nil {
			students = append(students, map[string]any{
				"student_id":      profile.Student.StudentID,
				"roll_no":         profile.Student.RollNo,
				"user_id":         profile.Student.UserID,
				"enrollment_date": enrollment.EnrollmentDate,
				"status":          enrollment.Status,
			})
		}
	}

	return helpers.Success(c, students, 200)
}
