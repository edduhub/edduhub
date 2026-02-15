package handler

import (
	"math"
	"time"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/analytics"
	"eduhub/server/internal/services/announcement"
	"eduhub/server/internal/services/assignment"
	"eduhub/server/internal/services/attendance"
	"eduhub/server/internal/services/audit"
	"eduhub/server/internal/services/calendar"
	"eduhub/server/internal/services/course"
	"eduhub/server/internal/services/enrollment"
	"eduhub/server/internal/services/grades"
	"eduhub/server/internal/services/student"

	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	studentService      student.StudentService
	courseService       course.CourseService
	attendanceService   attendance.AttendanceService
	announcementService announcement.AnnouncementService
	calendarService     calendar.CalendarService
	analyticsService    analytics.AnalyticsService
	auditService        audit.AuditService
	assignmentService   assignment.AssignmentService
	enrollmentService   enrollment.EnrollmentService
	gradesService       grades.GradeServices
}

func NewDashboardHandler(
	studentService student.StudentService,
	courseService course.CourseService,
	attendanceService attendance.AttendanceService,
	announcementService announcement.AnnouncementService,
	calendarService calendar.CalendarService,
	analyticsService analytics.AnalyticsService,
	auditService audit.AuditService,
	assignmentService assignment.AssignmentService,
	enrollmentService enrollment.EnrollmentService,
	gradesService grades.GradeServices,
) *DashboardHandler {
	return &DashboardHandler{
		studentService:      studentService,
		courseService:       courseService,
		attendanceService:   attendanceService,
		announcementService: announcementService,
		calendarService:     calendarService,
		analyticsService:    analyticsService,
		auditService:        auditService,
		assignmentService:   assignmentService,
		enrollmentService:   enrollmentService,
		gradesService:       gradesService,
	}
}

// GetDashboard returns dashboard data based on user role
func (h *DashboardHandler) GetDashboard(c echo.Context) error {
	ctx := c.Request().Context()
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Get total students
	totalStudents := 0
	if students, err := h.studentService.ListStudents(ctx, collegeID, 1000, 0); err == nil {
		totalStudents = len(students)
	}

	// Get total courses
	totalCourses := 0
	if courses, err := h.courseService.FindAllCourses(ctx, collegeID, 1000, 0); err == nil {
		totalCourses = len(courses)
	}

	// Get actual attendance rate from analytics service
	attendanceRate := 0.0
	totalFaculty := 0
	if dashboard, err := h.analyticsService.GetCollegeDashboard(ctx, collegeID); err == nil {
		attendanceRate = dashboard.AverageAttendance
		totalFaculty = dashboard.TotalFaculty
	}

	// Get recent announcements
	announcements := []map[string]interface{}{}
	isPublished := true
	announcementFilter := models.AnnouncementFilter{
		CollegeID:   &collegeID,
		IsPublished: &isPublished,
		Limit:       5,
		Offset:      0,
	}
	if announcementList, err := h.announcementService.GetAnnouncements(ctx, announcementFilter); err == nil {
		for _, a := range announcementList {
			announcements = append(announcements, map[string]interface{}{
				"id":       a.ID,
				"title":    a.Title,
				"content":  a.Content,
				"priority": a.Priority,
			})
		}
	}

	// Get upcoming calendar events
	upcomingEvents := []map[string]interface{}{}
	now := time.Now()
	calendarFilter := models.CalendarBlockFilter{
		CollegeID: &collegeID,
		StartDate: &now,
		Limit:     10,
		Offset:    0,
	}
	if events, err := h.calendarService.GetEvents(ctx, calendarFilter); err == nil {
		for _, event := range events {
			upcomingEvents = append(upcomingEvents, map[string]interface{}{
				"id":     event.ID,
				"title":  event.Title,
				"start":  event.StartTime,
				"end":    event.EndTime,
				"course": event.Description,
			})
		}
	}

	// Get recent audit activity
	recentActivity := []map[string]interface{}{}
	if auditLogs, err := h.auditService.GetAuditLogs(ctx, collegeID, nil, "", "", 10, 0); err == nil {
		for _, log := range auditLogs {
			recentActivity = append(recentActivity, map[string]interface{}{
				"id":        log.ID,
				"entity":    log.EntityType,
				"message":   log.Action,
				"timestamp": log.Timestamp,
			})
		}
	}

	// Get pending submissions count
	pendingSubmissions := 0
	if count, err := h.assignmentService.CountPendingSubmissionsByCollege(ctx, collegeID); err == nil {
		pendingSubmissions = count
	}

	// Build response with real data
	response := map[string]interface{}{
		"metrics": map[string]interface{}{
			"totalStudents":      totalStudents,
			"totalCourses":       totalCourses,
			"totalFaculty":       totalFaculty,
			"attendanceRate":     attendanceRate,
			"pendingSubmissions": pendingSubmissions,
		},
		"announcements":  announcements,
		"upcomingEvents": upcomingEvents,
		"recentActivity": recentActivity,
	}

	return helpers.Success(c, response, 200)
}

// GetStudentDashboard returns comprehensive dashboard data for a specific student
// @Summary Get Student Dashboard
// @Description Retrieves comprehensive dashboard data including courses, grades, assignments, and attendance for the authenticated student
// @Tags Dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/student/dashboard [get]
func (h *DashboardHandler) GetStudentDashboard(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract college ID and user info
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return helpers.Error(c, "Unable to extract college ID", 400)
	}

	// Get the kratos ID from identity context to find student
	kratosID, err := helpers.GetKratosID(c)
	if err != nil {
		return helpers.Error(c, "Unable to extract user identity", 401)
	}

	// Get student record from Kratos ID
	student, err := h.studentService.FindByKratosID(ctx, kratosID)
	if err != nil {
		return helpers.Error(c, "Student not found", 404)
	}

	// Get enrolled courses
	enrollments, err := h.enrollmentService.FindEnrollmentsByStudent(ctx, collegeID, student.StudentID, 100, 0)
	if err != nil {
		enrollments = []*models.Enrollment{}
	}

	// Build course data with grades
	courseData := []map[string]interface{}{}
	totalCredits := 0.0
	weightedGradePoints := 0.0

	for _, enr := range enrollments {
		course, err := h.courseService.FindCourseByID(ctx, collegeID, enr.CourseID)
		if err != nil {
			continue
		}

		// Get grades for this course
		gradeFilter := models.GradeFilter{
			StudentID: &student.StudentID,
			CourseID:  &course.ID,
			CollegeID: &collegeID,
		}
		courseGrades, err := h.gradesService.GetGrades(ctx, gradeFilter)
		if err != nil {
			courseGrades = []*models.Grade{}
		}

		// Calculate average grade for the course
		var avgPercentage float64
		if len(courseGrades) > 0 {
			var total float64
			for _, g := range courseGrades {
				total += g.Percentage
			}
			avgPercentage = total / float64(len(courseGrades))
		}

		// Get attendance for this course
		attendanceRecords, err := h.attendanceService.GetAttendanceByStudentAndCourse(ctx, collegeID, student.StudentID, course.ID, 1000, 0)
		if err != nil {
			attendanceRecords = []*models.Attendance{}
		}

		// Calculate attendance percentage
		totalSessions := len(attendanceRecords)
		presentCount := 0
		for _, att := range attendanceRecords {
			if att.Status == "present" {
				presentCount++
			}
		}
		attendancePercentage := 0.0
		if totalSessions > 0 {
			attendancePercentage = math.Round((float64(presentCount) / float64(totalSessions)) * 100)
		}

		// Calculate GPA contribution (using 4.0 scale)
		gradePoint := calculateGradePoint(avgPercentage)
		credits := float64(course.Credits)
		if credits > 0 {
			totalCredits += credits
			weightedGradePoints += gradePoint * credits
		}

		courseData = append(courseData, map[string]interface{}{
			"id":               course.ID,
			"name":             course.Name,
			"description":      course.Description,
			"credits":          course.Credits,
			"instructorID":     course.InstructorID,
			"averageGrade":     math.Round(avgPercentage*100) / 100,
			"attendanceRate":   attendancePercentage,
			"totalSessions":    totalSessions,
			"presentSessions":  presentCount,
			"enrollmentStatus": enr.Status,
		})
	}

	// Calculate overall GPA
	gpa := 0.0
	if totalCredits > 0 {
		gpa = math.Round((weightedGradePoints/totalCredits)*100) / 100
	}

	// Get assignments for enrolled courses
	upcomingAssignments := []map[string]interface{}{}
	completedAssignments := []map[string]interface{}{}
	overdueAssignments := []map[string]interface{}{}
	now := time.Now()

	for _, enr := range enrollments {
		courseAssignments, err := h.assignmentService.GetAssignmentsByCourse(ctx, collegeID, enr.CourseID)
		if err != nil {
			continue
		}

		for _, assignment := range courseAssignments {
			// Get submission status
			submission, err := h.assignmentService.GetSubmissionByStudentAndAssignment(ctx, student.StudentID, assignment.ID)
			isSubmitted := err == nil && submission != nil

			assignmentData := map[string]interface{}{
				"id":          assignment.ID,
				"title":       assignment.Title,
				"courseID":    assignment.CourseID,
				"dueDate":     assignment.DueDate,
				"maxPoints":   assignment.MaxPoints,
				"isSubmitted": isSubmitted,
			}

			if isSubmitted {
				assignmentData["submittedAt"] = submission.SubmissionTime
				assignmentData["grade"] = submission.Grade
				assignmentData["feedback"] = submission.Feedback
				completedAssignments = append(completedAssignments, assignmentData)
			} else {
				if assignment.DueDate.Before(now) {
					overdueAssignments = append(overdueAssignments, assignmentData)
				} else {
					upcomingAssignments = append(upcomingAssignments, assignmentData)
				}
			}
		}
	}

	// Get recent grades (last 10)
	recentGradesFilter := models.GradeFilter{
		StudentID: &student.StudentID,
		CollegeID: &collegeID,
		Limit:     10,
	}
	recentGrades, err := h.gradesService.GetGrades(ctx, recentGradesFilter)
	if err != nil {
		recentGrades = []*models.Grade{}
	}

	recentGradesData := []map[string]interface{}{}
	for _, grade := range recentGrades {
		// Get course name
		course, err := h.courseService.FindCourseByID(ctx, collegeID, grade.CourseID)
		courseName := "Unknown Course"
		if err == nil {
			courseName = course.Name
		}

		recentGradesData = append(recentGradesData, map[string]interface{}{
			"id":             grade.ID,
			"courseName":     courseName,
			"assessmentName": grade.AssessmentName,
			"assessmentType": grade.AssessmentType,
			"obtainedMarks":  grade.ObtainedMarks,
			"totalMarks":     grade.TotalMarks,
			"percentage":     math.Round(grade.Percentage*100) / 100,
			"gradedAt":       grade.GradedAt,
		})
	}

	// Get overall attendance statistics
	totalAttendanceRecords := 0
	totalPresentSessions := 0
	for _, course := range courseData {
		if total, ok := course["totalSessions"].(int); ok {
			totalAttendanceRecords += total
		}
		if present, ok := course["presentSessions"].(int); ok {
			totalPresentSessions += present
		}
	}
	overallAttendanceRate := 0.0
	if totalAttendanceRecords > 0 {
		overallAttendanceRate = math.Round((float64(totalPresentSessions) / float64(totalAttendanceRecords)) * 100)
	}

	// Get upcoming calendar events
	upcomingEvents := []map[string]interface{}{}
	calendarFilter := models.CalendarBlockFilter{
		CollegeID: &collegeID,
		StartDate: &now,
		Limit:     10,
	}
	events, err := h.calendarService.GetEvents(ctx, calendarFilter)
	if err == nil {
		for _, event := range events {
			upcomingEvents = append(upcomingEvents, map[string]interface{}{
				"id":          event.ID,
				"title":       event.Title,
				"description": event.Description,
				"start":       event.StartTime,
				"end":         event.EndTime,
				"eventType":   event.EventType,
			})
		}
	}

	// Get recent announcements
	announcements := []map[string]interface{}{}
	isPublished := true
	announcementFilter := models.AnnouncementFilter{
		CollegeID:   &collegeID,
		IsPublished: &isPublished,
		Limit:       5,
	}
	announcementList, err := h.announcementService.GetAnnouncements(ctx, announcementFilter)
	if err == nil {
		for _, a := range announcementList {
			announcements = append(announcements, map[string]interface{}{
				"id":       a.ID,
				"title":    a.Title,
				"content":  a.Content,
				"priority": a.Priority,
			})
		}
	}

	// Build comprehensive response
	response := map[string]interface{}{
		"student": map[string]interface{}{
			"id":        student.StudentID,
			"rollNo":    student.RollNo,
			"collegeID": student.CollegeID,
		},
		"academicOverview": map[string]interface{}{
			"gpa":                     gpa,
			"totalCredits":            totalCredits,
			"enrolledCourses":         len(courseData),
			"attendanceRate":          overallAttendanceRate,
			"totalPresentSessions":    totalPresentSessions,
			"totalAttendanceSessions": totalAttendanceRecords,
		},
		"courses": courseData,
		"assignments": map[string]interface{}{
			"upcoming":  upcomingAssignments,
			"completed": completedAssignments,
			"overdue":   overdueAssignments,
			"summary": map[string]interface{}{
				"upcomingCount":  len(upcomingAssignments),
				"completedCount": len(completedAssignments),
				"overdueCount":   len(overdueAssignments),
			},
		},
		"recentGrades":   recentGradesData,
		"upcomingEvents": upcomingEvents,
		"announcements":  announcements,
	}

	return helpers.Success(c, response, 200)
}

// calculateGradePoint converts percentage to GPA point (4.0 scale)
// Uses the shared PercentageToGPA from analytics for consistency across the app.
func calculateGradePoint(percentage float64) float64 {
	return analytics.PercentageToGPA(percentage)
}
